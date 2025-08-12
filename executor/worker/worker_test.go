package worker

import (
	"online-judge/executor/docker"
	"online-judge/executor/types"
	"testing"
)

func TestComputeTestCaseStatus(t *testing.T) {
	tests := []struct {
		name           string
		execResult     *docker.ExecutionResult
		expectedOutput string
		want           string
	}{
		{
			name: "passed case",
			execResult: &docker.ExecutionResult{
				Output: "hello world",
				Status: "ACCEPTED",
			},
			expectedOutput: "hello world",
			want:           "PASSED",
		},
		{
			name: "wrong answer",
			execResult: &docker.ExecutionResult{
				Output: "hello",
				Status: "ACCEPTED",
			},
			expectedOutput: "hello world",
			want:           "WRONG_ANSWER",
		},
		{
			name: "time limit exceeded",
			execResult: &docker.ExecutionResult{
				Output: "partial output",
				Status: "TIME_LIMIT_EXCEEDED",
			},
			expectedOutput: "expected output",
			want:           "TIME_LIMIT_EXCEEDED",
		},
		{
			name: "compilation error",
			execResult: &docker.ExecutionResult{
				Output: "compilation failed",
				Status: "COMPILATION_ERROR",
			},
			expectedOutput: "expected output",
			want:           "COMPILATION_ERROR",
		},
		{
			name: "runtime error",
			execResult: &docker.ExecutionResult{
				Output: "segmentation fault",
				Status: "RUNTIME_ERROR",
			},
			expectedOutput: "expected output",
			want:           "RUNTIME_ERROR",
		},
		{
			name: "whitespace handling",
			execResult: &docker.ExecutionResult{
				Output: "  hello world  \n",
				Status: "ACCEPTED",
			},
			expectedOutput: "\n  hello world  ",
			want:           "PASSED",
		},
		{
			name: "empty output match",
			execResult: &docker.ExecutionResult{
				Output: "",
				Status: "ACCEPTED",
			},
			expectedOutput: "",
			want:           "PASSED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := computeTestCaseStatus(tt.execResult, tt.expectedOutput)
			if got != tt.want {
				t.Errorf("computeTestCaseStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestComputeOverallStatus(t *testing.T) {
	tests := []struct {
		name        string
		results     []types.TestCaseResultMessage
		wantStatus  string
		wantTime    float64
		wantMemory  int64
	}{
		{
			name:        "empty results",
			results:     []types.TestCaseResultMessage{},
			wantStatus:  "COMPILATION_ERROR",
			wantTime:    0.0,
			wantMemory:  0,
		},
		{
			name: "all passed",
			results: []types.TestCaseResultMessage{
				{Status: "PASSED", TimeTaken: 1.0, MemoryUsed: 100},
				{Status: "PASSED", TimeTaken: 1.5, MemoryUsed: 150},
			},
			wantStatus: "PASSED",
			wantTime:   1.5,
			wantMemory: 150,
		},
		{
			name: "compilation error priority",
			results: []types.TestCaseResultMessage{
				{Status: "PASSED", TimeTaken: 1.0, MemoryUsed: 100},
				{Status: "COMPILATION_ERROR", TimeTaken: 0.0, MemoryUsed: 0},
				{Status: "WRONG_ANSWER", TimeTaken: 2.0, MemoryUsed: 200},
			},
			wantStatus: "COMPILATION_ERROR",
			wantTime:   2.0,
			wantMemory: 200,
		},
		{
			name: "runtime error priority",
			results: []types.TestCaseResultMessage{
				{Status: "PASSED", TimeTaken: 1.0, MemoryUsed: 100},
				{Status: "RUNTIME_ERROR", TimeTaken: 1.5, MemoryUsed: 150},
				{Status: "WRONG_ANSWER", TimeTaken: 2.0, MemoryUsed: 200},
			},
			wantStatus: "RUNTIME_ERROR",
			wantTime:   2.0,
			wantMemory: 200,
		},
		{
			name: "time limit exceeded priority",
			results: []types.TestCaseResultMessage{
				{Status: "PASSED", TimeTaken: 1.0, MemoryUsed: 100},
				{Status: "TIME_LIMIT_EXCEEDED", TimeTaken: 3.0, MemoryUsed: 150},
				{Status: "WRONG_ANSWER", TimeTaken: 2.0, MemoryUsed: 200},
			},
			wantStatus: "TIME_LIMIT_EXCEEDED",
			wantTime:   3.0,
			wantMemory: 200,
		},
		{
			name: "memory limit exceeded priority",
			results: []types.TestCaseResultMessage{
				{Status: "PASSED", TimeTaken: 1.0, MemoryUsed: 100},
				{Status: "MEMORY_LIMIT_EXCEEDED", TimeTaken: 1.5, MemoryUsed: 512},
				{Status: "WRONG_ANSWER", TimeTaken: 2.0, MemoryUsed: 200},
			},
			wantStatus: "MEMORY_LIMIT_EXCEEDED",
			wantTime:   2.0,
			wantMemory: 512,
		},
		{
			name: "wrong answer priority",
			results: []types.TestCaseResultMessage{
				{Status: "PASSED", TimeTaken: 1.0, MemoryUsed: 100},
				{Status: "WRONG_ANSWER", TimeTaken: 2.0, MemoryUsed: 200},
				{Status: "PASSED", TimeTaken: 1.5, MemoryUsed: 150},
			},
			wantStatus: "WRONG_ANSWER",
			wantTime:   2.0,
			wantMemory: 200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotStatus, gotTime, gotMemory := computeOverallStatus(tt.results)
			if gotStatus != tt.wantStatus {
				t.Errorf("computeOverallStatus() status = %v, want %v", gotStatus, tt.wantStatus)
			}
			if gotTime != tt.wantTime {
				t.Errorf("computeOverallStatus() time = %v, want %v", gotTime, tt.wantTime)
			}
			if gotMemory != tt.wantMemory {
				t.Errorf("computeOverallStatus() memory = %v, want %v", gotMemory, tt.wantMemory)
			}
		})
	}
}