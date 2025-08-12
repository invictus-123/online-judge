package docker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/google/uuid"
)

// ExecutionResult holds the outcome of running code in a container.
type ExecutionResult struct {
	Output     string
	Status     string // e.g., "ACCEPTED", "WRONG_ANSWER", "TIME_LIMIT_EXCEEDED"
	TimeMillis int64
	MemoryKB   int64
}

// LanguageConfig defines the Docker image and commands for a language.
type LanguageConfig struct {
	Image      string
	SourceFile string
	CompileCmd []string
	ExecuteCmd []string
}

// A map of supported languages to their Docker configurations.
var langConfigs = map[string]LanguageConfig{
	"JAVA": {
		Image:      "openjdk:11-jdk-slim",
		SourceFile: "Main.java",
		CompileCmd: []string{"javac", "Main.java"},
		ExecuteCmd: []string{"java", "-cp", ".", "Main"},
	},
	"PYTHON": {
		Image:      "python:3.9-slim",
		SourceFile: "main.py",
		CompileCmd: nil, // Interpreted language
		ExecuteCmd: []string{"python", "main.py"},
	},
	"CPP": {
		Image:      "gcc:latest",
		SourceFile: "main.cpp",
		CompileCmd: []string{"g++", "main.cpp", "-o", "main"},
		ExecuteCmd: []string{"./main"},
	},
	// Add other languages here
}

// RunInContainer creates a Docker container, executes the code, and returns the result.
func RunInContainer(language, code, input string) (*ExecutionResult, error) {
	return RunInContainerWithLimits(0, language, code, input, 2.0, 256*1024*1024) // 2 seconds, 256MB
}

// RunInContainerWithLimits creates a Docker container with custom limits, executes the code, and returns the result.
func RunInContainerWithLimits(submissionID int64, language, code, input string, timeLimitSeconds float64, memoryLimitBytes int64) (*ExecutionResult, error) {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %w", err)
	}

	config, ok := langConfigs[language]
	if !ok {
		return nil, fmt.Errorf("unsupported language: %s", language)
	}

	// Create a temporary directory to store the source code
	tempDir, err := ioutil.TempDir("", "online-judge-")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Write the source code to the file
	sourceFilePath := fmt.Sprintf("%s/%s", tempDir, config.SourceFile)
	if err := ioutil.WriteFile(sourceFilePath, []byte(code), 0644); err != nil {
		return nil, fmt.Errorf("failed to write source code: %w", err)
	}

	// Pull the Docker image if it doesn't exist
	reader, err := cli.ImagePull(ctx, config.Image, types.ImagePullOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to pull image %s: %w", config.Image, err)
	}
	io.Copy(ioutil.Discard, reader) // Wait for pull to complete

	// Create the container with a long-running command so we can exec into it
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image:        config.Image,
		Cmd:          []string{"sleep", "300"}, // Keep container alive for 5 minutes
		WorkingDir:   "/app",
		Tty:          false,
		OpenStdin:    true,
		AttachStdout: true,
		AttachStderr: true,
	}, &container.HostConfig{
		Binds: []string{fmt.Sprintf("%s:/app", tempDir)},
		Resources: container.Resources{
			Memory: memoryLimitBytes,
		},
	}, nil, nil, "oj-"+uuid.New().String())
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}
	defer func() {
		// Ensure container is removed
		if submissionID > 0 {
			log.Printf("[Submission %d] Removing container %s", submissionID, resp.ID)
		} else {
			log.Printf("Removing container %s", resp.ID)
		}
		if err := cli.ContainerRemove(ctx, resp.ID, types.ContainerRemoveOptions{Force: true}); err != nil {
			if submissionID > 0 {
				log.Printf("[Submission %d] Failed to remove container %s: %v", submissionID, resp.ID, err)
			} else {
				log.Printf("Failed to remove container %s: %v", resp.ID, err)
			}
		}
	}()

	// Start the container so we can execute commands in it
	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return nil, fmt.Errorf("failed to start container: %w", err)
	}

	// --- COMPILE STEP ---
	if config.CompileCmd != nil {
		if submissionID > 0 {
			log.Printf("[Submission %d] Compiling code in container %s", submissionID, resp.ID)
		} else {
			log.Printf("Compiling code in container %s", resp.ID)
		}
		execConfig := types.ExecConfig{
			Cmd:          config.CompileCmd,
			AttachStdout: true,
			AttachStderr: true,
		}
		execID, err := cli.ContainerExecCreate(ctx, resp.ID, execConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create compile exec: %w", err)
		}

		execResp, err := cli.ContainerExecAttach(ctx, execID.ID, types.ExecStartCheck{})
		if err != nil {
			return nil, fmt.Errorf("failed to attach to compile exec: %w", err)
		}
		defer execResp.Close()

		if err := cli.ContainerExecStart(ctx, execID.ID, types.ExecStartCheck{}); err != nil {
			return nil, fmt.Errorf("failed to start compile exec: %w", err)
		}

		// Check compilation result
		inspect, err := cli.ContainerExecInspect(ctx, execID.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to inspect compile exec: %w", err)
		}

		if inspect.ExitCode != 0 {
			var compileErr bytes.Buffer
			io.Copy(&compileErr, execResp.Reader)
			return &ExecutionResult{
				Status:     "COMPILATION_ERROR",
				Output:     compileErr.String(),
				TimeMillis: 0,
				MemoryKB:   0,
			}, nil
		}

		// For C++, make the executable file executable
		if language == "CPP" {
			chmodConfig := types.ExecConfig{
				Cmd:          []string{"chmod", "+x", "main"},
				AttachStdout: false,
				AttachStderr: false,
			}
			chmodExecID, err := cli.ContainerExecCreate(ctx, resp.ID, chmodConfig)
			if err != nil {
				return nil, fmt.Errorf("failed to create chmod exec: %w", err)
			}
			
			if err := cli.ContainerExecStart(ctx, chmodExecID.ID, types.ExecStartCheck{}); err != nil {
				return nil, fmt.Errorf("failed to start chmod exec: %w", err)
			}
		}
	}

	// --- EXECUTION STEP ---
	// Execute the program using docker exec
	if submissionID > 0 {
		log.Printf("[Submission %d] Executing code in container %s", submissionID, resp.ID)
	} else {
		log.Printf("Executing code in container %s", resp.ID)
	}
	execConfig := types.ExecConfig{
		Cmd:          config.ExecuteCmd,
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
	}
	execID, err := cli.ContainerExecCreate(ctx, resp.ID, execConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create execution exec: %w", err)
	}

	execResp, err := cli.ContainerExecAttach(ctx, execID.ID, types.ExecStartCheck{})
	if err != nil {
		return nil, fmt.Errorf("failed to attach to execution exec: %w", err)
	}
	defer execResp.Close()

	// Start execution
	if err := cli.ContainerExecStart(ctx, execID.ID, types.ExecStartCheck{}); err != nil {
		return nil, fmt.Errorf("failed to start execution exec: %w", err)
	}

	// Write input to stdin
	_, err = execResp.Conn.Write([]byte(input))
	if err != nil {
		return nil, fmt.Errorf("failed to write to stdin: %w", err)
	}
	execResp.CloseWrite() // Close stdin to signal end of input

	startTime := time.Now()
	var outputBuffer bytes.Buffer
	var memoryUsageKB int64

	// Start memory monitoring
	memoryDone := make(chan int64)
	go func() {
		var maxMemory uint64 = 1024 * 1024 // Default 1MB in bytes
		
		// Monitor for a maximum duration to prevent hanging
		timeout := time.After(time.Duration(timeLimitSeconds*1.5) * time.Second)
		ticker := time.NewTicker(10 * time.Millisecond)
		defer ticker.Stop()
		
		for {
			select {
			case <-timeout:
				memoryDone <- int64(maxMemory / 1024)
				return
			case <-ticker.C:
				stats, err := cli.ContainerStats(ctx, resp.ID, false)
				if err != nil {
					continue // Continue monitoring on error
				}
				var statsData types.StatsJSON
				if err := json.NewDecoder(stats.Body).Decode(&statsData); err != nil {
					stats.Body.Close()
					continue // Continue monitoring on decode error
				}
				stats.Body.Close()
				
				if statsData.MemoryStats.Usage > 0 && statsData.MemoryStats.Usage > maxMemory {
					maxMemory = statsData.MemoryStats.Usage
				}
			}
		}
	}()

	// Read output with timeout
	done := make(chan bool)
	go func() {
		io.Copy(&outputBuffer, execResp.Reader)
		done <- true
	}()

	timeLimit := time.Duration(timeLimitSeconds * float64(time.Second))
	var timedOut bool
	select {
	case <-time.After(timeLimit):
		cli.ContainerKill(ctx, resp.ID, "SIGKILL")
		timedOut = true
	case <-done:
		// Execution completed, check exit code
	}

	// Wait for memory monitoring to complete
	memoryUsageKB = <-memoryDone
	if memoryUsageKB <= 0 {
		memoryUsageKB = 1024 // Default to 1MB if we can't measure
	}

	execTime := time.Since(startTime)

	if timedOut {
		return &ExecutionResult{
			Status:     "TIME_LIMIT_EXCEEDED",
			Output:     outputBuffer.String(),
			TimeMillis: execTime.Milliseconds(),
			MemoryKB:   memoryUsageKB,
		}, nil
	}

	// Check execution result
	inspect, err := cli.ContainerExecInspect(ctx, execID.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect execution exec: %w", err)
	}

	if inspect.ExitCode != 0 {
		return &ExecutionResult{
			Status:     "RUNTIME_ERROR",
			Output:     outputBuffer.String(),
			TimeMillis: execTime.Milliseconds(),
			MemoryKB:   memoryUsageKB,
		}, nil
	}
	
	// Check memory limit
	if memoryUsageKB*1024 > memoryLimitBytes {
		return &ExecutionResult{
			Status:     "MEMORY_LIMIT_EXCEEDED",
			Output:     strings.TrimSpace(outputBuffer.String()),
			TimeMillis: execTime.Milliseconds(),
			MemoryKB:   memoryUsageKB,
		}, nil
	}

	return &ExecutionResult{
		Status:     "ACCEPTED",
		Output:     strings.TrimSpace(outputBuffer.String()),
		TimeMillis: execTime.Milliseconds(),
		MemoryKB:   memoryUsageKB,
	}, nil
}
