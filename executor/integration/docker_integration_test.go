package integration

import (
	"context"
	"online-judge/executor/docker"
	"strings"
	"testing"
	"time"

	dockerTypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

func TestDockerIntegration_PythonExecution(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	code := "print('Hello, World!')"
	input := ""
	
	result, err := docker.RunInContainer("PYTHON", code, input)
	if err != nil {
		t.Fatalf("Python execution failed: %v", err)
	}
	
	if result.Status != "ACCEPTED" {
		t.Errorf("Expected ACCEPTED, got %s", result.Status)
	}
	if result.Output != "Hello, World!" {
		t.Errorf("Expected 'Hello, World!', got '%s'", result.Output)
	}
	if result.TimeMillis <= 0 {
		t.Error("Expected positive execution time")
	}
	if result.MemoryKB <= 0 {
		t.Error("Expected positive memory usage")
	}
}

func TestDockerIntegration_JavaExecution(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	code := `public class Main {
		public static void main(String[] args) {
			System.out.println("Hello Java");
		}
	}`
	input := ""
	
	result, err := docker.RunInContainer("JAVA", code, input)
	if err != nil {
		t.Fatalf("Java execution failed: %v", err)
	}
	
	if result.Status != "ACCEPTED" {
		t.Errorf("Expected ACCEPTED, got %s", result.Status)
	}
	if result.Output != "Hello Java" {
		t.Errorf("Expected 'Hello Java', got '%s'", result.Output)
	}
}

func TestDockerIntegration_CppExecution(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	code := `#include <iostream>
	int main() {
		std::cout << "Hello C++";
		return 0;
	}`
	input := ""
	
	result, err := docker.RunInContainer("CPP", code, input)
	if err != nil {
		t.Fatalf("C++ execution failed: %v", err)
	}
	
	if result.Status != "ACCEPTED" {
		t.Errorf("Expected ACCEPTED, got %s", result.Status)
	}
	if result.Output != "Hello C++" {
		t.Errorf("Expected 'Hello C++', got '%s'", result.Output)
	}
}

func TestDockerIntegration_PythonWithInput(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	code := "print(int(input()) + int(input()))"
	input := "5\n7"
	
	result, err := docker.RunInContainer("PYTHON", code, input)
	if err != nil {
		t.Fatalf("Python with input execution failed: %v", err)
	}
	
	if result.Status != "ACCEPTED" {
		t.Errorf("Expected ACCEPTED, got %s", result.Status)
	}
	if result.Output != "12" {
		t.Errorf("Expected '12', got '%s'", result.Output)
	}
}

func TestDockerIntegration_CompilationError(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	code := `public class Main {
		invalid syntax here
	}`
	input := ""
	
	result, err := docker.RunInContainer("JAVA", code, input)
	if err != nil {
		t.Fatalf("Java compilation error test failed: %v", err)
	}
	
	if result.Status != "COMPILATION_ERROR" {
		t.Errorf("Expected COMPILATION_ERROR, got %s", result.Status)
	}
	if result.Output == "" {
		t.Error("Expected compilation error output")
	}
}

func TestDockerIntegration_RuntimeError(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	code := `arr = [1, 2, 3]
print(arr[10])`
	input := ""
	
	result, err := docker.RunInContainer("PYTHON", code, input)
	if err != nil {
		t.Fatalf("Python runtime error test failed: %v", err)
	}
	
	if result.Status != "RUNTIME_ERROR" {
		t.Errorf("Expected RUNTIME_ERROR, got %s", result.Status)
	}
}

func TestDockerIntegration_TimeLimit(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	code := `import time
while True:
	time.sleep(0.1)`
	input := ""
	
	result, err := docker.RunInContainerWithLimits(123, "PYTHON", code, input, 0.5, 128*1024*1024)
	if err != nil {
		t.Fatalf("Python timeout test failed: %v", err)
	}
	
	if result.Status != "TIME_LIMIT_EXCEEDED" {
		t.Errorf("Expected TIME_LIMIT_EXCEEDED, got %s", result.Status)
	}
	if result.TimeMillis < 400 || result.TimeMillis > 600 {
		t.Errorf("Expected time around 500ms, got %dms", result.TimeMillis)
	}
}

func TestDockerIntegration_CustomLimits(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	code := "print('test')"
	input := ""
	timeLimit := 3.0
	memoryLimit := int64(512 * 1024 * 1024)
	
	result, err := docker.RunInContainerWithLimits(456, "PYTHON", code, input, timeLimit, memoryLimit)
	if err != nil {
		t.Fatalf("Custom limits test failed: %v", err)
	}
	
	if result.Status != "ACCEPTED" {
		t.Errorf("Expected ACCEPTED, got %s", result.Status)
	}
	if result.Output != "test" {
		t.Errorf("Expected 'test', got '%s'", result.Output)
	}
}

func TestDockerIntegration_MultiLineOutput(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	code := `print("Line 1")
print("Line 2")
print("Line 3")`
	input := ""
	
	result, err := docker.RunInContainer("PYTHON", code, input)
	if err != nil {
		t.Fatalf("Multi-line output test failed: %v", err)
	}
	
	expected := "Line 1\nLine 2\nLine 3"
	if result.Output != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result.Output)
	}
}

func TestDockerIntegration_LargeOutput(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	code := `for i in range(100):
	print(f"Line {i}")`
	input := ""
	
	result, err := docker.RunInContainer("PYTHON", code, input)
	if err != nil {
		t.Fatalf("Large output test failed: %v", err)
	}
	
	if result.Status != "ACCEPTED" {
		t.Errorf("Expected ACCEPTED, got %s", result.Status)
	}
	
	lines := strings.Split(result.Output, "\n")
	if len(lines) != 100 {
		t.Errorf("Expected 100 lines, got %d", len(lines))
	}
	
	if !strings.Contains(result.Output, "Line 0") {
		t.Error("Expected output to contain 'Line 0'")
	}
	if !strings.Contains(result.Output, "Line 99") {
		t.Error("Expected output to contain 'Line 99'")
	}
}

func TestDockerIntegration_ContainerCleanup(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		t.Skipf("Docker client not available: %v", err)
	}

	initialContainers, err := cli.ContainerList(context.Background(), dockerTypes.ContainerListOptions{All: true})
	if err != nil {
		t.Fatalf("Failed to list containers: %v", err)
	}

	code := "print('cleanup test')"
	input := ""
	
	_, err = docker.RunInContainer("PYTHON", code, input)
	if err != nil {
		t.Fatalf("Container cleanup test execution failed: %v", err)
	}

	time.Sleep(1 * time.Second)

	finalContainers, err := cli.ContainerList(context.Background(), dockerTypes.ContainerListOptions{All: true})
	if err != nil {
		t.Fatalf("Failed to list containers after execution: %v", err)
	}

	if len(finalContainers) > len(initialContainers) {
		t.Errorf("Container not cleaned up properly. Initial: %d, Final: %d", 
			len(initialContainers), len(finalContainers))
	}
}

func TestDockerIntegration_ConcurrentExecution(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	const numGoroutines = 5
	code := "print('concurrent test')"
	input := ""
	
	results := make(chan *docker.ExecutionResult, numGoroutines)
	errors := make(chan error, numGoroutines)
	
	for i := 0; i < numGoroutines; i++ {
		go func() {
			result, err := docker.RunInContainer("PYTHON", code, input)
			if err != nil {
				errors <- err
				return
			}
			results <- result
		}()
	}
	
	successCount := 0
	for i := 0; i < numGoroutines; i++ {
		select {
		case result := <-results:
			if result.Status == "ACCEPTED" && result.Output == "concurrent test" {
				successCount++
			}
		case err := <-errors:
			t.Errorf("Concurrent execution failed: %v", err)
		case <-time.After(30 * time.Second):
			t.Fatal("Concurrent execution timed out")
		}
	}
	
	if successCount != numGoroutines {
		t.Errorf("Expected %d successful executions, got %d", numGoroutines, successCount)
	}
}

func TestDockerIntegration_MemoryIntensive(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	code := `data = []
for i in range(100000):
	data.append(str(i))
print(len(data))`
	input := ""
	
	result, err := docker.RunInContainerWithLimits(789, "PYTHON", code, input, 5.0, 256*1024*1024)
	if err != nil {
		t.Fatalf("Memory intensive test failed: %v", err)
	}
	
	if result.Status != "ACCEPTED" {
		t.Errorf("Expected ACCEPTED, got %s", result.Status)
	}
	if result.Output != "100000" {
		t.Errorf("Expected '100000', got '%s'", result.Output)
	}
	if result.MemoryKB <= 0 {
		t.Error("Expected positive memory usage")
	}
}

func TestDockerIntegration_ErrorOutput(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	code := `import sys
print("stdout message")
print("stderr message", file=sys.stderr)`
	input := ""
	
	result, err := docker.RunInContainer("PYTHON", code, input)
	if err != nil {
		t.Fatalf("Error output test failed: %v", err)
	}
	
	if result.Status != "ACCEPTED" {
		t.Errorf("Expected ACCEPTED, got %s", result.Status)
	}
	
	if !strings.Contains(result.Output, "stdout message") {
		t.Error("Expected output to contain stdout message")
	}
}