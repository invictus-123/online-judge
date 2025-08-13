package testutil

import (
	"encoding/base64"
	"encoding/json"
	"online-judge/executor/types"

	"github.com/rabbitmq/amqp091-go"
)

func CreateTestSubmission(submissionID int64, language, code string, timeLimit float64, memoryLimit int64, testCases []TestCase) types.SubmissionMessage {
	var testCaseMessages []types.TestCaseMessage
	for _, tc := range testCases {
		testCaseMessages = append(testCaseMessages, types.TestCaseMessage{
			TestCaseID:     tc.ID,
			Input:          base64.StdEncoding.EncodeToString([]byte(tc.Input)),
			ExpectedOutput: base64.StdEncoding.EncodeToString([]byte(tc.ExpectedOutput)),
		})
	}

	return types.SubmissionMessage{
		SubmissionID: submissionID,
		Language:     language,
		Code:         base64.StdEncoding.EncodeToString([]byte(code)),
		TimeLimit:    timeLimit,
		MemoryLimit:  memoryLimit,
		TestCases:    testCaseMessages,
	}
}

type TestCase struct {
	ID             string
	Input          string
	ExpectedOutput string
}

func CreateTestDelivery(submission types.SubmissionMessage) amqp091.Delivery {
	data, _ := json.Marshal(submission)
	return amqp091.Delivery{Body: data}
}

func CreateSimpleTestCase(id, input, expectedOutput string) TestCase {
	return TestCase{
		ID:             id,
		Input:          input,
		ExpectedOutput: expectedOutput,
	}
}

func CreatePythonHelloWorldSubmission() types.SubmissionMessage {
	return CreateTestSubmission(
		1,
		"PYTHON",
		"print('Hello, World!')",
		2.0,
		128,
		[]TestCase{
			CreateSimpleTestCase("tc1", "", "Hello, World!"),
		},
	)
}

func CreateJavaHelloWorldSubmission() types.SubmissionMessage {
	javaCode := `public class Main {
    public static void main(String[] args) {
        System.out.println("Hello, World!");
    }
}`
	return CreateTestSubmission(
		2,
		"JAVA",
		javaCode,
		3.0,
		256,
		[]TestCase{
			CreateSimpleTestCase("tc1", "", "Hello, World!"),
		},
	)
}

func CreateCppHelloWorldSubmission() types.SubmissionMessage {
	cppCode := `#include <iostream>
int main() {
    std::cout << "Hello, World!" << std::endl;
    return 0;
}`
	return CreateTestSubmission(
		3,
		"CPP",
		cppCode,
		2.5,
		512,
		[]TestCase{
			CreateSimpleTestCase("tc1", "", "Hello, World!"),
		},
	)
}

func CreateAdditionSubmission() types.SubmissionMessage {
	return CreateTestSubmission(
		4,
		"PYTHON",
		"print(int(input()) + int(input()))",
		1.0,
		64,
		[]TestCase{
			CreateSimpleTestCase("tc1", "1\n2", "3"),
			CreateSimpleTestCase("tc2", "5\n7", "12"),
			CreateSimpleTestCase("tc3", "0\n0", "0"),
		},
	)
}

func CreateInfiniteLoopSubmission() types.SubmissionMessage {
	return CreateTestSubmission(
		5,
		"PYTHON",
		"while True: pass",
		0.5,
		128,
		[]TestCase{
			CreateSimpleTestCase("tc1", "", "should timeout"),
		},
	)
}

func CreateCompilationErrorSubmission() types.SubmissionMessage {
	return CreateTestSubmission(
		6,
		"JAVA",
		"invalid java code without main method",
		2.0,
		256,
		[]TestCase{
			CreateSimpleTestCase("tc1", "", "should not compile"),
		},
	)
}

func CreateRuntimeErrorSubmission() types.SubmissionMessage {
	pythonCode := `
arr = [1, 2, 3]
print(arr[10])  # Index out of bounds
`
	return CreateTestSubmission(
		7,
		"PYTHON",
		pythonCode,
		2.0,
		128,
		[]TestCase{
			CreateSimpleTestCase("tc1", "", "should crash"),
		},
	)
}

func CreateWrongAnswerSubmission() types.SubmissionMessage {
	return CreateTestSubmission(
		8,
		"PYTHON",
		"print('Wrong Answer')",
		1.0,
		64,
		[]TestCase{
			CreateSimpleTestCase("tc1", "", "Correct Answer"),
		},
	)
}

func CreateMultipleTestCasesSubmission() types.SubmissionMessage {
	return CreateTestSubmission(
		9,
		"PYTHON",
		"print(input().upper())",
		1.5,
		128,
		[]TestCase{
			CreateSimpleTestCase("tc1", "hello", "HELLO"),
			CreateSimpleTestCase("tc2", "world", "WORLD"),
			CreateSimpleTestCase("tc3", "test", "TEST"),
			CreateSimpleTestCase("tc4", "python", "PYTHON"),
		},
	)
}

func CreateFibonacciSubmission() types.SubmissionMessage {
	pythonCode := `
n = int(input())
if n <= 1:
    print(n)
else:
    a, b = 0, 1
    for _ in range(2, n + 1):
        a, b = b, a + b
    print(b)
`
	return CreateTestSubmission(
		10,
		"PYTHON",
		pythonCode,
		3.0,
		256,
		[]TestCase{
			CreateSimpleTestCase("tc1", "0", "0"),
			CreateSimpleTestCase("tc2", "1", "1"),
			CreateSimpleTestCase("tc3", "5", "5"),
			CreateSimpleTestCase("tc4", "10", "55"),
		},
	)
}

func CreateMixedResultsSubmission() types.SubmissionMessage {
	pythonCode := `
n = int(input())
if n == 1:
    print("one")
elif n == 2:
    print("two")
elif n == 3:
    print("wrong")
else:
    print("unknown")
`
	return CreateTestSubmission(
		11,
		"PYTHON",
		pythonCode,
		2.0,
		128,
		[]TestCase{
			CreateSimpleTestCase("tc1", "1", "one"),
			CreateSimpleTestCase("tc2", "2", "two"),
			CreateSimpleTestCase("tc3", "3", "three"),
			CreateSimpleTestCase("tc4", "4", "unknown"),
		},
	)
}

func CreateLargeInputSubmission() types.SubmissionMessage {
	pythonCode := `
import sys
total = 0
for line in sys.stdin:
    total += int(line.strip())
print(total)
`
	largeInput := ""
	for i := 1; i <= 1000; i++ {
		largeInput += "1\n"
	}

	return CreateTestSubmission(
		12,
		"PYTHON",
		pythonCode,
		5.0,
		512,
		[]TestCase{
			CreateSimpleTestCase("tc1", largeInput, "1000"),
		},
	)
}

type ExpectedResult struct {
	OverallStatus    string
	TestCaseResults  map[string]string
	ShouldHaveTime   bool
	ShouldHaveMemory bool
}

func AssertSubmissionResult(result types.ResultNotificationMessage, expected ExpectedResult) bool {
	if result.Status != expected.OverallStatus {
		return false
	}

	if len(result.Results) != len(expected.TestCaseResults) {
		return false
	}

	for _, tcResult := range result.Results {
		expectedStatus, exists := expected.TestCaseResults[tcResult.TestCaseID]
		if !exists || tcResult.Status != expectedStatus {
			return false
		}
	}

	if expected.ShouldHaveTime && result.TimeTaken <= 0 {
		return false
	}

	if expected.ShouldHaveMemory && result.MemoryUsed <= 0 {
		return false
	}

	return true
}

func ExpectAllPassed(testCaseIDs []string) ExpectedResult {
	results := make(map[string]string)
	for _, id := range testCaseIDs {
		results[id] = "PASSED"
	}
	return ExpectedResult{
		OverallStatus:    "PASSED",
		TestCaseResults:  results,
		ShouldHaveTime:   true,
		ShouldHaveMemory: true,
	}
}

func ExpectWrongAnswer(testCaseIDs []string, wrongCaseID string) ExpectedResult {
	results := make(map[string]string)
	for _, id := range testCaseIDs {
		if id == wrongCaseID {
			results[id] = "WRONG_ANSWER"
		} else {
			results[id] = "PASSED"
		}
	}
	return ExpectedResult{
		OverallStatus:    "WRONG_ANSWER",
		TestCaseResults:  results,
		ShouldHaveTime:   true,
		ShouldHaveMemory: true,
	}
}

func ExpectTimeLimit(testCaseIDs []string) ExpectedResult {
	results := make(map[string]string)
	for _, id := range testCaseIDs {
		results[id] = "TIME_LIMIT_EXCEEDED"
	}
	return ExpectedResult{
		OverallStatus:    "TIME_LIMIT_EXCEEDED",
		TestCaseResults:  results,
		ShouldHaveTime:   true,
		ShouldHaveMemory: false,
	}
}

func ExpectCompilationError(testCaseIDs []string) ExpectedResult {
	results := make(map[string]string)
	for _, id := range testCaseIDs {
		results[id] = "COMPILATION_ERROR"
	}
	return ExpectedResult{
		OverallStatus:    "COMPILATION_ERROR",
		TestCaseResults:  results,
		ShouldHaveTime:   false,
		ShouldHaveMemory: false,
	}
}

func ExpectRuntimeError(testCaseIDs []string) ExpectedResult {
	results := make(map[string]string)
	for _, id := range testCaseIDs {
		results[id] = "RUNTIME_ERROR"
	}
	return ExpectedResult{
		OverallStatus:    "RUNTIME_ERROR",
		TestCaseResults:  results,
		ShouldHaveTime:   true,
		ShouldHaveMemory: true,
	}
}

func CreateInvalidBase64Submission() types.SubmissionMessage {
	return types.SubmissionMessage{
		SubmissionID: 999,
		Language:     "PYTHON",
		Code:         "invalid-base64-code",
		TimeLimit:    1.0,
		MemoryLimit:  128,
		TestCases: []types.TestCaseMessage{
			{
				TestCaseID:     "tc1",
				Input:          base64.StdEncoding.EncodeToString([]byte("")),
				ExpectedOutput: base64.StdEncoding.EncodeToString([]byte("output")),
			},
		},
	}
}

func CreateInvalidInputSubmission() types.SubmissionMessage {
	return types.SubmissionMessage{
		SubmissionID: 998,
		Language:     "PYTHON",
		Code:         base64.StdEncoding.EncodeToString([]byte("print('hello')")),
		TimeLimit:    1.0,
		MemoryLimit:  128,
		TestCases: []types.TestCaseMessage{
			{
				TestCaseID:     "tc1",
				Input:          "invalid-base64-input",
				ExpectedOutput: base64.StdEncoding.EncodeToString([]byte("hello")),
			},
		},
	}
}

func CreateInvalidOutputSubmission() types.SubmissionMessage {
	return types.SubmissionMessage{
		SubmissionID: 997,
		Language:     "PYTHON",
		Code:         base64.StdEncoding.EncodeToString([]byte("print('hello')")),
		TimeLimit:    1.0,
		MemoryLimit:  128,
		TestCases: []types.TestCaseMessage{
			{
				TestCaseID:     "tc1",
				Input:          base64.StdEncoding.EncodeToString([]byte("")),
				ExpectedOutput: "invalid-base64-output",
			},
		},
	}
}
