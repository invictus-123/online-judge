package testutil

import (
	"encoding/base64"
	"online-judge/executor/types"
	"testing"
)

func TestCreateTestSubmission(t *testing.T) {
	testCases := []TestCase{
		{ID: "tc1", Input: "hello", ExpectedOutput: "HELLO"},
		{ID: "tc2", Input: "world", ExpectedOutput: "WORLD"},
	}

	submission := CreateTestSubmission(123, "PYTHON", "print(input().upper())", 2.0, 256, testCases)

	if submission.SubmissionID != 123 {
		t.Errorf("SubmissionID = %d, want 123", submission.SubmissionID)
	}
	if submission.Language != "PYTHON" {
		t.Errorf("Language = %s, want PYTHON", submission.Language)
	}
	if submission.TimeLimit != 2.0 {
		t.Errorf("TimeLimit = %f, want 2.0", submission.TimeLimit)
	}
	if submission.MemoryLimit != 256 {
		t.Errorf("MemoryLimit = %d, want 256", submission.MemoryLimit)
	}

	decodedCode, _ := base64.StdEncoding.DecodeString(submission.Code)
	if string(decodedCode) != "print(input().upper())" {
		t.Errorf("Decoded code = %s, want print(input().upper())", string(decodedCode))
	}

	if len(submission.TestCases) != 2 {
		t.Fatalf("TestCases length = %d, want 2", len(submission.TestCases))
	}

	tc1 := submission.TestCases[0]
	if tc1.TestCaseID != "tc1" {
		t.Errorf("TestCaseID = %s, want tc1", tc1.TestCaseID)
	}

	decodedInput, _ := base64.StdEncoding.DecodeString(tc1.Input)
	if string(decodedInput) != "hello" {
		t.Errorf("Decoded input = %s, want hello", string(decodedInput))
	}

	decodedOutput, _ := base64.StdEncoding.DecodeString(tc1.ExpectedOutput)
	if string(decodedOutput) != "HELLO" {
		t.Errorf("Decoded output = %s, want HELLO", string(decodedOutput))
	}
}

func TestCreateTestDelivery(t *testing.T) {
	submission := CreatePythonHelloWorldSubmission()
	delivery := CreateTestDelivery(submission)

	if len(delivery.Body) == 0 {
		t.Error("Delivery body should not be empty")
	}
}

func TestCreateSimpleTestCase(t *testing.T) {
	tc := CreateSimpleTestCase("test1", "input data", "output data")

	if tc.ID != "test1" {
		t.Errorf("ID = %s, want test1", tc.ID)
	}
	if tc.Input != "input data" {
		t.Errorf("Input = %s, want input data", tc.Input)
	}
	if tc.ExpectedOutput != "output data" {
		t.Errorf("ExpectedOutput = %s, want output data", tc.ExpectedOutput)
	}
}

func TestPredefinedSubmissions(t *testing.T) {
	tests := []struct {
		name       string
		submission func() types.SubmissionMessage
		language   string
	}{
		{"Python Hello World", CreatePythonHelloWorldSubmission, "PYTHON"},
		{"Java Hello World", CreateJavaHelloWorldSubmission, "JAVA"},
		{"C++ Hello World", CreateCppHelloWorldSubmission, "CPP"},
		{"Addition", CreateAdditionSubmission, "PYTHON"},
		{"Infinite Loop", CreateInfiniteLoopSubmission, "PYTHON"},
		{"Compilation Error", CreateCompilationErrorSubmission, "JAVA"},
		{"Runtime Error", CreateRuntimeErrorSubmission, "PYTHON"},
		{"Wrong Answer", CreateWrongAnswerSubmission, "PYTHON"},
		{"Multiple Test Cases", CreateMultipleTestCasesSubmission, "PYTHON"},
		{"Fibonacci", CreateFibonacciSubmission, "PYTHON"},
		{"Mixed Results", CreateMixedResultsSubmission, "PYTHON"},
		{"Large Input", CreateLargeInputSubmission, "PYTHON"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			submission := tt.submission()
			if submission.Language != tt.language {
				t.Errorf("Language = %s, want %s", submission.Language, tt.language)
			}
			if submission.SubmissionID == 0 {
				t.Error("SubmissionID should not be 0")
			}
			if submission.TimeLimit <= 0 {
				t.Error("TimeLimit should be positive")
			}
			if submission.MemoryLimit <= 0 {
				t.Error("MemoryLimit should be positive")
			}
			if len(submission.TestCases) == 0 {
				t.Error("Should have at least one test case")
			}
		})
	}
}

func TestAssertSubmissionResult(t *testing.T) {
	result := types.ResultNotificationMessage{
		SubmissionID: 1,
		Status:       "PASSED",
		TimeTaken:    1.5,
		MemoryUsed:   128,
		Results: []types.TestCaseResultMessage{
			{TestCaseID: "tc1", Status: "PASSED"},
			{TestCaseID: "tc2", Status: "PASSED"},
		},
	}

	expected := ExpectedResult{
		OverallStatus: "PASSED",
		TestCaseResults: map[string]string{
			"tc1": "PASSED",
			"tc2": "PASSED",
		},
		ShouldHaveTime:   true,
		ShouldHaveMemory: true,
	}

	if !AssertSubmissionResult(result, expected) {
		t.Error("AssertSubmissionResult should return true for matching result")
	}
}

func TestAssertSubmissionResultMismatch(t *testing.T) {
	result := types.ResultNotificationMessage{
		SubmissionID: 1,
		Status:       "WRONG_ANSWER",
		TimeTaken:    1.5,
		MemoryUsed:   128,
		Results: []types.TestCaseResultMessage{
			{TestCaseID: "tc1", Status: "PASSED"},
			{TestCaseID: "tc2", Status: "WRONG_ANSWER"},
		},
	}

	expected := ExpectedResult{
		OverallStatus: "PASSED",
		TestCaseResults: map[string]string{
			"tc1": "PASSED",
			"tc2": "PASSED",
		},
		ShouldHaveTime:   true,
		ShouldHaveMemory: true,
	}

	if AssertSubmissionResult(result, expected) {
		t.Error("AssertSubmissionResult should return false for mismatched result")
	}
}

func TestExpectedResultHelpers(t *testing.T) {
	testCaseIDs := []string{"tc1", "tc2", "tc3"}

	t.Run("ExpectAllPassed", func(t *testing.T) {
		expected := ExpectAllPassed(testCaseIDs)
		if expected.OverallStatus != "PASSED" {
			t.Errorf("OverallStatus = %s, want PASSED", expected.OverallStatus)
		}
		if len(expected.TestCaseResults) != 3 {
			t.Errorf("TestCaseResults length = %d, want 3", len(expected.TestCaseResults))
		}
		for _, id := range testCaseIDs {
			if expected.TestCaseResults[id] != "PASSED" {
				t.Errorf("TestCaseResults[%s] = %s, want PASSED", id, expected.TestCaseResults[id])
			}
		}
	})

	t.Run("ExpectWrongAnswer", func(t *testing.T) {
		expected := ExpectWrongAnswer(testCaseIDs, "tc2")
		if expected.OverallStatus != "WRONG_ANSWER" {
			t.Errorf("OverallStatus = %s, want WRONG_ANSWER", expected.OverallStatus)
		}
		if expected.TestCaseResults["tc1"] != "PASSED" {
			t.Errorf("tc1 status = %s, want PASSED", expected.TestCaseResults["tc1"])
		}
		if expected.TestCaseResults["tc2"] != "WRONG_ANSWER" {
			t.Errorf("tc2 status = %s, want WRONG_ANSWER", expected.TestCaseResults["tc2"])
		}
		if expected.TestCaseResults["tc3"] != "PASSED" {
			t.Errorf("tc3 status = %s, want PASSED", expected.TestCaseResults["tc3"])
		}
	})

	t.Run("ExpectTimeLimit", func(t *testing.T) {
		expected := ExpectTimeLimit(testCaseIDs)
		if expected.OverallStatus != "TIME_LIMIT_EXCEEDED" {
			t.Errorf("OverallStatus = %s, want TIME_LIMIT_EXCEEDED", expected.OverallStatus)
		}
		if !expected.ShouldHaveTime {
			t.Error("ShouldHaveTime should be true")
		}
		if expected.ShouldHaveMemory {
			t.Error("ShouldHaveMemory should be false for timeout")
		}
	})

	t.Run("ExpectCompilationError", func(t *testing.T) {
		expected := ExpectCompilationError(testCaseIDs)
		if expected.OverallStatus != "COMPILATION_ERROR" {
			t.Errorf("OverallStatus = %s, want COMPILATION_ERROR", expected.OverallStatus)
		}
		if expected.ShouldHaveTime {
			t.Error("ShouldHaveTime should be false for compilation error")
		}
		if expected.ShouldHaveMemory {
			t.Error("ShouldHaveMemory should be false for compilation error")
		}
	})

	t.Run("ExpectRuntimeError", func(t *testing.T) {
		expected := ExpectRuntimeError(testCaseIDs)
		if expected.OverallStatus != "RUNTIME_ERROR" {
			t.Errorf("OverallStatus = %s, want RUNTIME_ERROR", expected.OverallStatus)
		}
		if !expected.ShouldHaveTime {
			t.Error("ShouldHaveTime should be true for runtime error")
		}
		if !expected.ShouldHaveMemory {
			t.Error("ShouldHaveMemory should be true for runtime error")
		}
	})
}

func TestInvalidSubmissions(t *testing.T) {
	t.Run("Invalid Base64 Code", func(t *testing.T) {
		submission := CreateInvalidBase64Submission()
		if submission.Code == "" {
			t.Error("Code should not be empty")
		}
		if submission.SubmissionID != 999 {
			t.Errorf("SubmissionID = %d, want 999", submission.SubmissionID)
		}
	})

	t.Run("Invalid Base64 Input", func(t *testing.T) {
		submission := CreateInvalidInputSubmission()
		if len(submission.TestCases) == 0 {
			t.Error("Should have test cases")
		}
		if submission.TestCases[0].Input == "" {
			t.Error("Input should not be empty")
		}
	})

	t.Run("Invalid Base64 Output", func(t *testing.T) {
		submission := CreateInvalidOutputSubmission()
		if len(submission.TestCases) == 0 {
			t.Error("Should have test cases")
		}
		if submission.TestCases[0].ExpectedOutput == "" {
			t.Error("Expected output should not be empty")
		}
	})
}

func TestLargeInputSubmission(t *testing.T) {
	submission := CreateLargeInputSubmission()
	
	if len(submission.TestCases) != 1 {
		t.Fatalf("Expected 1 test case, got %d", len(submission.TestCases))
	}

	decodedInput, err := base64.StdEncoding.DecodeString(submission.TestCases[0].Input)
	if err != nil {
		t.Fatalf("Failed to decode input: %v", err)
	}

	inputLines := len([]byte(string(decodedInput)))
	if inputLines < 1000 {
		t.Errorf("Large input should be substantial, got %d bytes", inputLines)
	}

	decodedOutput, err := base64.StdEncoding.DecodeString(submission.TestCases[0].ExpectedOutput)
	if err != nil {
		t.Fatalf("Failed to decode expected output: %v", err)
	}

	if string(decodedOutput) != "1000" {
		t.Errorf("Expected output = %s, want 1000", string(decodedOutput))
	}
}

func TestFibonacciSubmissionCases(t *testing.T) {
	submission := CreateFibonacciSubmission()
	
	if len(submission.TestCases) != 4 {
		t.Fatalf("Expected 4 test cases, got %d", len(submission.TestCases))
	}

	expectedResults := map[string]string{
		"tc1": "0",
		"tc2": "1", 
		"tc3": "5",
		"tc4": "55",
	}

	for _, tc := range submission.TestCases {
		expected, exists := expectedResults[tc.TestCaseID]
		if !exists {
			t.Errorf("Unexpected test case ID: %s", tc.TestCaseID)
			continue
		}

		decodedOutput, err := base64.StdEncoding.DecodeString(tc.ExpectedOutput)
		if err != nil {
			t.Fatalf("Failed to decode expected output for %s: %v", tc.TestCaseID, err)
		}

		if string(decodedOutput) != expected {
			t.Errorf("Test case %s expected output = %s, want %s", tc.TestCaseID, string(decodedOutput), expected)
		}
	}
}