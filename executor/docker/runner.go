package docker

import (
	"archive/tar"
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

	if submissionID > 0 {
		log.Printf("[Submission %d] Created source file at %s, bind mounting to /app", submissionID, sourceFilePath)
	} else {
		log.Printf("Created source file at %s, bind mounting to /app", sourceFilePath)
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

	// Copy source file into container using Docker CopyToContainer API
	if err := copyFileToContainer(cli, ctx, resp.ID, sourceFilePath, config.SourceFile, submissionID); err != nil {
		return nil, fmt.Errorf("failed to copy source file to container: %w", err)
	}

	// List files before compilation to check source file exists
	preCompileListConfig := types.ExecConfig{
		Cmd:          []string{"ls", "-la"},
		AttachStdout: true,
		AttachStderr: true,
	}
	preCompileListExecID, err := cli.ContainerExecCreate(ctx, resp.ID, preCompileListConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create pre-compile ls exec: %w", err)
	}

	preCompileListResp, err := cli.ContainerExecAttach(ctx, preCompileListExecID.ID, types.ExecStartCheck{})
	if err != nil {
		return nil, fmt.Errorf("failed to attach to pre-compile ls exec: %w", err)
	}
	defer preCompileListResp.Close()

	if err := cli.ContainerExecStart(ctx, preCompileListExecID.ID, types.ExecStartCheck{}); err != nil {
		return nil, fmt.Errorf("failed to start pre-compile ls exec: %w", err)
	}

	var preCompileListOutput bytes.Buffer
	io.Copy(&preCompileListOutput, preCompileListResp.Reader)
	if submissionID > 0 {
		log.Printf("[Submission %d] Files before compilation: %s", submissionID, preCompileListOutput.String())
	} else {
		log.Printf("Files before compilation: %s", preCompileListOutput.String())
	}

	// --- COMPILE STEP ---
	if config.CompileCmd != nil {
		if submissionID > 0 {
			log.Printf("[Submission %d] Compiling code in container %s with command: %v", submissionID, resp.ID, config.CompileCmd)
		} else {
			log.Printf("Compiling code in container %s with command: %v", resp.ID, config.CompileCmd)
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

		// Always read compilation output (even on success)
		var compileOutput bytes.Buffer
		io.Copy(&compileOutput, execResp.Reader)
		compileOutputStr := compileOutput.String()

		// Check for compilation failure - either non-zero exit code OR error messages in output
		compilationFailed := inspect.ExitCode != 0 || strings.Contains(compileOutputStr, "fatal error") ||
			strings.Contains(compileOutputStr, "No such file") || strings.Contains(compileOutputStr, "error:")

		if compilationFailed {
			if submissionID > 0 {
				log.Printf("[Submission %d] Compilation failed with exit code %d: %s", submissionID, inspect.ExitCode, compileOutputStr)
			} else {
				log.Printf("Compilation failed with exit code %d: %s", inspect.ExitCode, compileOutputStr)
			}
			return &ExecutionResult{
				Status:     "COMPILATION_ERROR",
				Output:     compileOutputStr,
				TimeMillis: 0,
				MemoryKB:   0,
			}, nil
		}

		if submissionID > 0 {
			log.Printf("[Submission %d] Compilation completed successfully. Output: %s", submissionID, compileOutputStr)
		} else {
			log.Printf("Compilation completed successfully. Output: %s", compileOutputStr)
		}

		// List files after compilation to debug
		listConfig := types.ExecConfig{
			Cmd:          []string{"ls", "-la"},
			AttachStdout: true,
			AttachStderr: true,
		}
		listExecID, err := cli.ContainerExecCreate(ctx, resp.ID, listConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create ls exec: %w", err)
		}

		listResp, err := cli.ContainerExecAttach(ctx, listExecID.ID, types.ExecStartCheck{})
		if err != nil {
			return nil, fmt.Errorf("failed to attach to ls exec: %w", err)
		}
		defer listResp.Close()

		if err := cli.ContainerExecStart(ctx, listExecID.ID, types.ExecStartCheck{}); err != nil {
			return nil, fmt.Errorf("failed to start ls exec: %w", err)
		}

		var listOutput bytes.Buffer
		io.Copy(&listOutput, listResp.Reader)
		if submissionID > 0 {
			log.Printf("[Submission %d] Files after compilation: %s", submissionID, listOutput.String())
		} else {
			log.Printf("Files after compilation: %s", listOutput.String())
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
	if submissionID > 0 {
		log.Printf("[Submission %d] Providing input to program: %q", submissionID, input)
	} else {
		log.Printf("Providing input to program: %q", input)
	}
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
		runtimeErrMsg := outputBuffer.String()
		if submissionID > 0 {
			log.Printf("[Submission %d] Runtime error with exit code %d. Output/Error: %s", submissionID, inspect.ExitCode, runtimeErrMsg)
		} else {
			log.Printf("Runtime error with exit code %d. Output/Error: %s", inspect.ExitCode, runtimeErrMsg)
		}
		return &ExecutionResult{
			Status:     "RUNTIME_ERROR",
			Output:     runtimeErrMsg,
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

	finalOutput := strings.TrimSpace(outputBuffer.String())
	if submissionID > 0 {
		log.Printf("[Submission %d] Program executed successfully. Output: %q", submissionID, finalOutput)
	} else {
		log.Printf("Program executed successfully. Output: %q", finalOutput)
	}

	return &ExecutionResult{
		Status:     "ACCEPTED",
		Output:     finalOutput,
		TimeMillis: execTime.Milliseconds(),
		MemoryKB:   memoryUsageKB,
	}, nil
}

// copyFileToContainer copies a file from the host to the container using Docker's CopyToContainer API
func copyFileToContainer(cli *client.Client, ctx context.Context, containerID, hostFilePath, containerFileName string, submissionID int64) error {
	// Read the source file content
	fileContent, err := ioutil.ReadFile(hostFilePath)
	if err != nil {
		return fmt.Errorf("failed to read source file: %w", err)
	}

	// Create a tar archive containing the file
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	header := &tar.Header{
		Name: containerFileName,
		Mode: 0644,
		Size: int64(len(fileContent)),
	}

	if err := tw.WriteHeader(header); err != nil {
		return fmt.Errorf("failed to write tar header: %w", err)
	}

	if _, err := tw.Write(fileContent); err != nil {
		return fmt.Errorf("failed to write file content to tar: %w", err)
	}

	if err := tw.Close(); err != nil {
		return fmt.Errorf("failed to close tar writer: %w", err)
	}

	// Copy the tar archive to the container
	if err := cli.CopyToContainer(ctx, containerID, "/app", &buf, types.CopyToContainerOptions{}); err != nil {
		return fmt.Errorf("failed to copy to container: %w", err)
	}

	if submissionID > 0 {
		log.Printf("[Submission %d] Successfully copied %s to container /app/%s", submissionID, hostFilePath, containerFileName)
	} else {
		log.Printf("Successfully copied %s to container /app/%s", hostFilePath, containerFileName)
	}

	return nil
}
