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
		if err := cli.ContainerRemove(ctx, resp.ID, types.ContainerRemoveOptions{Force: true}); err != nil {
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

	// --- COMPILE STEP ---
	if config.CompileCmd != nil {
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
			return &ExecutionResult{
				Status:     "COMPILATION_ERROR",
				Output:     compileOutputStr,
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

	// Create execution command that redirects stdout/stderr to files
	execConfig := types.ExecConfig{
		Cmd:         []string{"sh", "-c", strings.Join(config.ExecuteCmd, " ") + " > /app/stdout.txt 2> /app/stderr.txt"},
		AttachStdin: true,
	}
	execID, err := cli.ContainerExecCreate(ctx, resp.ID, execConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create execution exec: %w", err)
	}

	// Add timeout for Docker exec operations to prevent hanging
	dockerCtx, dockerCancel := context.WithTimeout(ctx, 30.0)
	defer dockerCancel()

	execResp, err := cli.ContainerExecAttach(dockerCtx, execID.ID, types.ExecStartCheck{})
	if err != nil {
		return nil, fmt.Errorf("failed to attach to execution exec: %w", err)
	}
	defer execResp.Close()

	// Start execution
	if err := cli.ContainerExecStart(dockerCtx, execID.ID, types.ExecStartCheck{}); err != nil {
		return nil, fmt.Errorf("failed to start execution exec: %w", err)
	}

	// Write input to stdin
	log.Printf("[Submission %d] Writing input to container stdin", submissionID)
	_, err = execResp.Conn.Write([]byte(input))
	if err != nil {
		return nil, fmt.Errorf("failed to write to stdin: %w", err)
	}
	execResp.CloseWrite() // Close stdin to signal end of input
	log.Printf("[Submission %d] Starting execution monitoring", submissionID)

	startTime := time.Now()
	var memoryUsageKB int64

	// Start memory monitoring
	memoryDone := make(chan int64, 1)
	memoryCtx, memoryCancel := context.WithTimeout(ctx, time.Duration(timeLimitSeconds*1.5)*time.Second)
	go func() {
		defer close(memoryDone)
		defer memoryCancel()

		var maxMemory uint64 = 1024 * 1024 // Default 1MB in bytes
		ticker := time.NewTicker(10 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-memoryCtx.Done():
				memoryDone <- int64(maxMemory / 1024)
				return
			case <-ticker.C:
				stats, err := cli.ContainerStats(memoryCtx, resp.ID, false)
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

	// Wait for execution completion with timeout
	done := make(chan error)
	execCtx, execCancel := context.WithTimeout(ctx, time.Duration(timeLimitSeconds*float64(time.Second)))
	defer execCancel()

	go func() {
		defer close(done)
		// Use context-aware copy to prevent hanging
		_, err := io.Copy(ioutil.Discard, execResp.Reader)
		select {
		case done <- err:
		case <-execCtx.Done():
			// Context cancelled, don't block
		}
	}()

	timeLimit := time.Duration(timeLimitSeconds * float64(time.Second))
	var timedOut bool
	select {
	case <-time.After(timeLimit):
		execCancel() // Cancel the copy operation
		cli.ContainerKill(ctx, resp.ID, "SIGKILL")
		timedOut = true
		// Give a brief moment for cleanup
		time.Sleep(100 * time.Millisecond)
	case copyErr := <-done:
		if copyErr != nil && execCtx.Err() != nil {
			// Copy was cancelled due to timeout
			timedOut = true
		}
		// Execution completed, check exit code
	}

	// Wait for memory monitoring to complete with timeout
	select {
	case memoryUsageKB = <-memoryDone:
		if memoryUsageKB <= 0 {
			memoryUsageKB = 1024 // Default to 1MB if we can't measure
		}
	case <-time.After(1 * time.Second):
		memoryCancel()       // Ensure memory monitoring stops
		memoryUsageKB = 1024 // Default value
	}

	execTime := time.Since(startTime)

	if timedOut {
		log.Printf("[Submission %d] Code execution timed out after %.3fs", submissionID, execTime.Seconds())
		return &ExecutionResult{
			Status:     "TIME_LIMIT_EXCEEDED",
			Output:     "Time limit exceeded",
			TimeMillis: execTime.Milliseconds(),
			MemoryKB:   memoryUsageKB,
		}, nil
	}

	// Check execution result
	inspect, err := cli.ContainerExecInspect(ctx, execID.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect execution exec: %w", err)
	}

	// Read output files from container
	stdout, stderr, err := readOutputFiles(cli, ctx, resp.ID, submissionID)
	if err != nil {
		return nil, fmt.Errorf("failed to read output files: %w", err)
	}

	if inspect.ExitCode != 0 {

		// Return stderr for runtime errors, stdout for output if stderr is empty
		errorOutput := stderr
		if strings.TrimSpace(errorOutput) == "" {
			errorOutput = stdout
		}

		return &ExecutionResult{
			Status:     "RUNTIME_ERROR",
			Output:     strings.TrimSpace(errorOutput),
			TimeMillis: execTime.Milliseconds(),
			MemoryKB:   memoryUsageKB,
		}, nil
	}

	// Check memory limit
	if memoryUsageKB*1024 > memoryLimitBytes {
		return &ExecutionResult{
			Status:     "MEMORY_LIMIT_EXCEEDED",
			Output:     strings.TrimSpace(stdout),
			TimeMillis: execTime.Milliseconds(),
			MemoryKB:   memoryUsageKB,
		}, nil
	}

	return &ExecutionResult{
		Status:     "ACCEPTED",
		Output:     strings.TrimSpace(stdout),
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

	return nil
}

// readOutputFiles reads stdout and stderr files from the container using Docker's CopyFromContainer API
func readOutputFiles(cli *client.Client, ctx context.Context, containerID string, submissionID int64) (stdout, stderr string, err error) {
	// Read stdout file
	stdoutContent, err := readFileFromContainer(cli, ctx, containerID, "/app/stdout.txt")
	if err != nil {
		stdoutContent = "" // Not an error, file might not exist if no output
	}

	// Read stderr file
	stderrContent, err := readFileFromContainer(cli, ctx, containerID, "/app/stderr.txt")
	if err != nil {
		stderrContent = "" // Not an error, file might not exist if no errors
	}

	return stdoutContent, stderrContent, nil
}

// readFileFromContainer reads a single file from container using Docker's CopyFromContainer API
func readFileFromContainer(cli *client.Client, ctx context.Context, containerID, filePath string) (string, error) {
	// Copy file from container
	reader, stat, err := cli.CopyFromContainer(ctx, containerID, filePath)
	if err != nil {
		return "", fmt.Errorf("failed to copy file from container: %w", err)
	}
	defer reader.Close()

	// Create tar reader
	tr := tar.NewReader(reader)

	// Read the first (and should be only) file from the tar
	header, err := tr.Next()
	if err != nil {
		return "", fmt.Errorf("failed to read tar header: %w", err)
	}

	if header.Typeflag != tar.TypeReg {
		return "", fmt.Errorf("expected regular file, got type %c", header.Typeflag)
	}

	// Read file content
	var content bytes.Buffer
	_, err = io.Copy(&content, tr)
	if err != nil {
		return "", fmt.Errorf("failed to read file content: %w", err)
	}

	// Log file size for debugging
	_ = stat // Use stat to avoid unused variable error

	return content.String(), nil
}
