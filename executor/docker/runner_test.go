package docker

import (
	"strings"
	"testing"
	"time"
)

func TestLanguageConfigs(t *testing.T) {
	tests := []struct {
		language string
		exists   bool
	}{
		{"JAVA", true},
		{"PYTHON", true},
		{"CPP", true},
		{"JAVASCRIPT", false},
		{"GOLANG", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.language, func(t *testing.T) {
			_, exists := langConfigs[tt.language]
			if exists != tt.exists {
				t.Errorf("langConfigs[%s] exists = %v, want %v", tt.language, exists, tt.exists)
			}
		})
	}
}

func TestLanguageConfigsStructure(t *testing.T) {
	for lang, config := range langConfigs {
		t.Run(lang, func(t *testing.T) {
			if config.Image == "" {
				t.Error("Image should not be empty")
			}
			if config.SourceFile == "" {
				t.Error("SourceFile should not be empty")
			}
			if len(config.ExecuteCmd) == 0 {
				t.Error("ExecuteCmd should not be empty")
			}
		})
	}
}

func TestJavaConfig(t *testing.T) {
	config := langConfigs["JAVA"]
	
	if config.Image != "openjdk:11-jdk-slim" {
		t.Errorf("Java image = %s, want openjdk:11-jdk-slim", config.Image)
	}
	if config.SourceFile != "Main.java" {
		t.Errorf("Java source file = %s, want Main.java", config.SourceFile)
	}
	if len(config.CompileCmd) != 2 || config.CompileCmd[0] != "javac" || config.CompileCmd[1] != "Main.java" {
		t.Errorf("Java compile command = %v, want [javac Main.java]", config.CompileCmd)
	}
	if len(config.ExecuteCmd) != 4 || config.ExecuteCmd[0] != "java" || config.ExecuteCmd[1] != "-cp" || config.ExecuteCmd[2] != "." || config.ExecuteCmd[3] != "Main" {
		t.Errorf("Java execute command = %v, want [java -cp . Main]", config.ExecuteCmd)
	}
}

func TestPythonConfig(t *testing.T) {
	config := langConfigs["PYTHON"]
	
	if config.Image != "python:3.9-slim" {
		t.Errorf("Python image = %s, want python:3.9-slim", config.Image)
	}
	if config.SourceFile != "main.py" {
		t.Errorf("Python source file = %s, want main.py", config.SourceFile)
	}
	if config.CompileCmd != nil {
		t.Errorf("Python compile command = %v, want nil", config.CompileCmd)
	}
	if len(config.ExecuteCmd) != 2 || config.ExecuteCmd[0] != "python" || config.ExecuteCmd[1] != "main.py" {
		t.Errorf("Python execute command = %v, want [python main.py]", config.ExecuteCmd)
	}
}

func TestCppConfig(t *testing.T) {
	config := langConfigs["CPP"]
	
	if config.Image != "gcc:latest" {
		t.Errorf("CPP image = %s, want gcc:latest", config.Image)
	}
	if config.SourceFile != "main.cpp" {
		t.Errorf("CPP source file = %s, want main.cpp", config.SourceFile)
	}
	expectedCompile := []string{"g++", "main.cpp", "-o", "main"}
	if len(config.CompileCmd) != 4 {
		t.Errorf("CPP compile command length = %d, want 4", len(config.CompileCmd))
	} else {
		for i, cmd := range expectedCompile {
			if config.CompileCmd[i] != cmd {
				t.Errorf("CPP compile command[%d] = %s, want %s", i, config.CompileCmd[i], cmd)
			}
		}
	}
	if len(config.ExecuteCmd) != 1 || config.ExecuteCmd[0] != "./main" {
		t.Errorf("CPP execute command = %v, want [./main]", config.ExecuteCmd)
	}
}

func TestExecutionResult(t *testing.T) {
	result := &ExecutionResult{
		Output:     "Hello World",
		Status:     "ACCEPTED",
		TimeMillis: 1500,
		MemoryKB:   256,
	}

	if result.Output != "Hello World" {
		t.Errorf("Output = %s, want Hello World", result.Output)
	}
	if result.Status != "ACCEPTED" {
		t.Errorf("Status = %s, want ACCEPTED", result.Status)
	}
	if result.TimeMillis != 1500 {
		t.Errorf("TimeMillis = %d, want 1500", result.TimeMillis)
	}
	if result.MemoryKB != 256 {
		t.Errorf("MemoryKB = %d, want 256", result.MemoryKB)
	}
}

func TestUnsupportedLanguage(t *testing.T) {
	result, err := RunInContainer("UNSUPPORTED", "code", "input")
	
	if err == nil {
		t.Error("Expected error for unsupported language, got nil")
	}
	if result != nil {
		t.Error("Expected nil result for unsupported language")
	}
	if !strings.Contains(err.Error(), "unsupported language") {
		t.Errorf("Error message = %s, want to contain 'unsupported language'", err.Error())
	}
}

func TestTimeoutCalculation(t *testing.T) {
	tests := []struct {
		name        string
		timeLimit   float64
		expectedDur time.Duration
	}{
		{"1 second", 1.0, time.Second},
		{"2.5 seconds", 2.5, 2*time.Second + 500*time.Millisecond},
		{"0.1 seconds", 0.1, 100 * time.Millisecond},
		{"5 seconds", 5.0, 5 * time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			duration := time.Duration(tt.timeLimit * float64(time.Second))
			if duration != tt.expectedDur {
				t.Errorf("Duration = %v, want %v", duration, tt.expectedDur)
			}
		})
	}
}

func TestMemoryLimitConversion(t *testing.T) {
	tests := []struct {
		name        string
		limitBytes  int64
		limitMB     int64
	}{
		{"256MB", 256 * 1024 * 1024, 256},
		{"512MB", 512 * 1024 * 1024, 512},
		{"1GB", 1024 * 1024 * 1024, 1024},
		{"128MB", 128 * 1024 * 1024, 128},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bytesFromMB := tt.limitMB * 1024 * 1024
			if bytesFromMB != tt.limitBytes {
				t.Errorf("Conversion = %d bytes, want %d bytes", bytesFromMB, tt.limitBytes)
			}
		})
	}
}

func TestOutputTrimming(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"no whitespace", "hello", "hello"},
		{"trailing newline", "hello\n", "hello"},
		{"leading spaces", "  hello", "hello"},
		{"both ends", "\n  hello world  \n", "hello world"},
		{"multiple lines", "line1\nline2\n", "line1\nline2"},
		{"empty", "", ""},
		{"only whitespace", "   \n\t  ", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := strings.TrimSpace(tt.input)
			if result != tt.expected {
				t.Errorf("TrimSpace(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}