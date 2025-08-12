package integration

import (
	"encoding/base64"
	"encoding/json"
	"online-judge/executor/master"
	"online-judge/executor/rabbitmq"
	"online-judge/executor/testutil"
	"online-judge/executor/types"
	"sync"
	"testing"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

type errorScenarioMQClient struct {
	publishedMessages []errorMessage
	consumeQueue      chan amqp091.Delivery
	publishError      error
	consumeError      error
	mu                sync.Mutex
}

type errorMessage struct {
	exchange   string
	routingKey string
	body       interface{}
	timestamp  time.Time
}

func (m *errorScenarioMQClient) Publish(exchange, routingKey string, body interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.publishError != nil {
		return m.publishError
	}
	m.publishedMessages = append(m.publishedMessages, errorMessage{
		exchange:   exchange,
		routingKey: routingKey,
		body:       body,
		timestamp:  time.Now(),
	})
	return nil
}

func (m *errorScenarioMQClient) ConsumeSubmissions(queueName string) (<-chan amqp091.Delivery, error) {
	if m.consumeError != nil {
		return nil, m.consumeError
	}
	return m.consumeQueue, nil
}

func (m *errorScenarioMQClient) getMessages() []errorMessage {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]errorMessage{}, m.publishedMessages...)
}

func (m *errorScenarioMQClient) reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.publishedMessages = nil
}

func (m *errorScenarioMQClient) addSubmission(submission types.SubmissionMessage) {
	data, _ := json.Marshal(submission)
	delivery := amqp091.Delivery{Body: data}
	m.consumeQueue <- delivery
}

func (m *errorScenarioMQClient) addInvalidJSON(invalidJSON string) {
	delivery := amqp091.Delivery{Body: []byte(invalidJSON)}
	m.consumeQueue <- delivery
}

func TestErrorScenarios_InvalidJSONSubmission(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	mqClient := &errorScenarioMQClient{
		consumeQueue: make(chan amqp091.Delivery, 10),
	}

	m, err := master.NewMaster(mqClient, 1, "test.queue")
	if err != nil {
		t.Fatalf("Failed to create master: %v", err)
	}

	m.Start()
	time.Sleep(100 * time.Millisecond)

	mqClient.addInvalidJSON("{invalid json syntax")
	mqClient.addInvalidJSON("not json at all")
	mqClient.addInvalidJSON("")

	validSubmission := testutil.CreatePythonHelloWorldSubmission()
	mqClient.addSubmission(validSubmission)

	time.Sleep(5 * time.Second)

	messages := mqClient.getMessages()
	
	statusCount := 0
	resultCount := 0
	for _, msg := range messages {
		if msg.exchange == rabbitmq.StatusExchange {
			statusCount++
		} else if msg.exchange == rabbitmq.ResultExchange {
			resultCount++
		}
	}

	if statusCount != 1 {
		t.Errorf("Expected 1 status message (valid submission only), got %d", statusCount)
	}
	if resultCount != 1 {
		t.Errorf("Expected 1 result message (valid submission only), got %d", resultCount)
	}
}

func TestErrorScenarios_MalformedSubmissionData(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	mqClient := &errorScenarioMQClient{
		consumeQueue: make(chan amqp091.Delivery, 10),
	}

	m, err := master.NewMaster(mqClient, 1, "test.queue")
	if err != nil {
		t.Fatalf("Failed to create master: %v", err)
	}

	m.Start()
	time.Sleep(100 * time.Millisecond)

	malformedSubmissions := []types.SubmissionMessage{
		{
			SubmissionID: 1,
			Language:     "UNSUPPORTED_LANG",
			Code:         base64.StdEncoding.EncodeToString([]byte("print('test')")),
			TimeLimit:    1.0,
			MemoryLimit:  128,
			TestCases:    []types.TestCaseMessage{},
		},
		{
			SubmissionID: 2,
			Language:     "",
			Code:         base64.StdEncoding.EncodeToString([]byte("print('test')")),
			TimeLimit:    1.0,
			MemoryLimit:  128,
			TestCases:    []types.TestCaseMessage{},
		},
		{
			SubmissionID: 3,
			Language:     "PYTHON",
			Code:         "invalid-base64",
			TimeLimit:    1.0,
			MemoryLimit:  128,
			TestCases: []types.TestCaseMessage{
				{
					TestCaseID:     "tc1",
					Input:          base64.StdEncoding.EncodeToString([]byte("")),
					ExpectedOutput: base64.StdEncoding.EncodeToString([]byte("test")),
				},
			},
		},
	}

	for _, submission := range malformedSubmissions {
		mqClient.addSubmission(submission)
	}

	time.Sleep(10 * time.Second)

	messages := mqClient.getMessages()
	
	statusMessages := []errorMessage{}
	for _, msg := range messages {
		if msg.exchange == rabbitmq.StatusExchange {
			statusMessages = append(statusMessages, msg)
		}
	}

	if len(statusMessages) < 1 {
		t.Error("Expected at least some status messages for malformed submissions")
	}
}

func TestErrorScenarios_InvalidBase64TestCases(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	mqClient := &errorScenarioMQClient{
		consumeQueue: make(chan amqp091.Delivery, 10),
	}

	m, err := master.NewMaster(mqClient, 1, "test.queue")
	if err != nil {
		t.Fatalf("Failed to create master: %v", err)
	}

	m.Start()
	time.Sleep(100 * time.Millisecond)

	submissionWithInvalidInput := types.SubmissionMessage{
		SubmissionID: 100,
		Language:     "PYTHON",
		Code:         base64.StdEncoding.EncodeToString([]byte("print('test')")),
		TimeLimit:    1.0,
		MemoryLimit:  128,
		TestCases: []types.TestCaseMessage{
			{
				TestCaseID:     "tc1",
				Input:          "invalid-base64-input",
				ExpectedOutput: base64.StdEncoding.EncodeToString([]byte("test")),
			},
		},
	}

	submissionWithInvalidOutput := types.SubmissionMessage{
		SubmissionID: 101,
		Language:     "PYTHON",
		Code:         base64.StdEncoding.EncodeToString([]byte("print('test')")),
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

	mqClient.addSubmission(submissionWithInvalidInput)
	mqClient.addSubmission(submissionWithInvalidOutput)

	time.Sleep(10 * time.Second)

	messages := mqClient.getMessages()
	
	resultMessages := []errorMessage{}
	for _, msg := range messages {
		if msg.exchange == rabbitmq.ResultExchange {
			resultMessages = append(resultMessages, msg)
		}
	}

	if len(resultMessages) != 2 {
		t.Fatalf("Expected 2 result messages, got %d", len(resultMessages))
	}

	for _, msg := range resultMessages {
		result := msg.body.(types.ResultNotificationMessage)
		if result.Status != "COMPILATION_ERROR" {
			t.Errorf("Expected COMPILATION_ERROR for invalid base64, got %s", result.Status)
		}
		if len(result.Results) < 1 {
			t.Error("Expected at least one test case result")
		}
		if result.Results[0].Status != "COMPILATION_ERROR" {
			t.Errorf("Expected test case COMPILATION_ERROR, got %s", result.Results[0].Status)
		}
	}
}

func TestErrorScenarios_ExtremeResourceLimits(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	mqClient := &errorScenarioMQClient{
		consumeQueue: make(chan amqp091.Delivery, 10),
	}

	m, err := master.NewMaster(mqClient, 1, "test.queue")
	if err != nil {
		t.Fatalf("Failed to create master: %v", err)
	}

	m.Start()
	time.Sleep(100 * time.Millisecond)

	veryLowTimeLimit := testutil.CreatePythonHelloWorldSubmission()
	veryLowTimeLimit.SubmissionID = 200
	veryLowTimeLimit.TimeLimit = 0.001

	veryLowMemoryLimit := testutil.CreatePythonHelloWorldSubmission()
	veryLowMemoryLimit.SubmissionID = 201
	veryLowMemoryLimit.MemoryLimit = 1

	veryHighTimeLimit := testutil.CreatePythonHelloWorldSubmission()
	veryHighTimeLimit.SubmissionID = 202
	veryHighTimeLimit.TimeLimit = 3600.0

	veryHighMemoryLimit := testutil.CreatePythonHelloWorldSubmission()
	veryHighMemoryLimit.SubmissionID = 203
	veryHighMemoryLimit.MemoryLimit = 8192

	mqClient.addSubmission(veryLowTimeLimit)
	mqClient.addSubmission(veryLowMemoryLimit)
	mqClient.addSubmission(veryHighTimeLimit)
	mqClient.addSubmission(veryHighMemoryLimit)

	time.Sleep(15 * time.Second)

	messages := mqClient.getMessages()
	
	resultMessages := []errorMessage{}
	for _, msg := range messages {
		if msg.exchange == rabbitmq.ResultExchange {
			resultMessages = append(resultMessages, msg)
		}
	}

	if len(resultMessages) != 4 {
		t.Fatalf("Expected 4 result messages, got %d", len(resultMessages))
	}

	submissionResults := make(map[int64]string)
	for _, msg := range resultMessages {
		result := msg.body.(types.ResultNotificationMessage)
		submissionResults[result.SubmissionID] = result.Status
	}

	if submissionResults[200] != "TIME_LIMIT_EXCEEDED" {
		t.Errorf("Very low time limit result = %s, want TIME_LIMIT_EXCEEDED", submissionResults[200])
	}

	if submissionResults[201] != "PASSED" {
		t.Errorf("Very low memory limit result = %s, want PASSED", submissionResults[201])
	}

	if submissionResults[202] != "PASSED" {
		t.Errorf("Very high time limit result = %s, want PASSED", submissionResults[202])
	}

	if submissionResults[203] != "PASSED" {
		t.Errorf("Very high memory limit result = %s, want PASSED", submissionResults[203])
	}
}

func TestErrorScenarios_EmptyAndSpecialInputs(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	mqClient := &errorScenarioMQClient{
		consumeQueue: make(chan amqp091.Delivery, 10),
	}

	m, err := master.NewMaster(mqClient, 1, "test.queue")
	if err != nil {
		t.Fatalf("Failed to create master: %v", err)
	}

	m.Start()
	time.Sleep(100 * time.Millisecond)

	emptyCodeSubmission := types.SubmissionMessage{
		SubmissionID: 300,
		Language:     "PYTHON",
		Code:         base64.StdEncoding.EncodeToString([]byte("")),
		TimeLimit:    1.0,
		MemoryLimit:  128,
		TestCases: []types.TestCaseMessage{
			{
				TestCaseID:     "tc1",
				Input:          base64.StdEncoding.EncodeToString([]byte("")),
				ExpectedOutput: base64.StdEncoding.EncodeToString([]byte("")),
			},
		},
	}

	specialCharsSubmission := types.SubmissionMessage{
		SubmissionID: 301,
		Language:     "PYTHON",
		Code:         base64.StdEncoding.EncodeToString([]byte("print('Hello\\nWorld\\t!')")),
		TimeLimit:    1.0,
		MemoryLimit:  128,
		TestCases: []types.TestCaseMessage{
			{
				TestCaseID:     "tc1",
				Input:          base64.StdEncoding.EncodeToString([]byte("")),
				ExpectedOutput: base64.StdEncoding.EncodeToString([]byte("Hello\nWorld\t!")),
			},
		},
	}

	unicodeSubmission := types.SubmissionMessage{
		SubmissionID: 302,
		Language:     "PYTHON",
		Code:         base64.StdEncoding.EncodeToString([]byte("print('Hello 世界! 🌍')")),
		TimeLimit:    1.0,
		MemoryLimit:  128,
		TestCases: []types.TestCaseMessage{
			{
				TestCaseID:     "tc1",
				Input:          base64.StdEncoding.EncodeToString([]byte("")),
				ExpectedOutput: base64.StdEncoding.EncodeToString([]byte("Hello 世界! 🌍")),
			},
		},
	}

	mqClient.addSubmission(emptyCodeSubmission)
	mqClient.addSubmission(specialCharsSubmission)
	mqClient.addSubmission(unicodeSubmission)

	time.Sleep(10 * time.Second)

	messages := mqClient.getMessages()
	
	resultMessages := []errorMessage{}
	for _, msg := range messages {
		if msg.exchange == rabbitmq.ResultExchange {
			resultMessages = append(resultMessages, msg)
		}
	}

	if len(resultMessages) != 3 {
		t.Fatalf("Expected 3 result messages, got %d", len(resultMessages))
	}

	submissionResults := make(map[int64]string)
	for _, msg := range resultMessages {
		result := msg.body.(types.ResultNotificationMessage)
		submissionResults[result.SubmissionID] = result.Status
	}

	if submissionResults[300] == "PASSED" {
		t.Error("Empty code should not pass")
	}

	if submissionResults[301] != "PASSED" {
		t.Errorf("Special characters result = %s, want PASSED", submissionResults[301])
	}

	if submissionResults[302] != "PASSED" {
		t.Errorf("Unicode result = %s, want PASSED", submissionResults[302])
	}
}

func TestErrorScenarios_LargeCodeSubmission(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	mqClient := &errorScenarioMQClient{
		consumeQueue: make(chan amqp091.Delivery, 10),
	}

	m, err := master.NewMaster(mqClient, 1, "test.queue")
	if err != nil {
		t.Fatalf("Failed to create master: %v", err)
	}

	m.Start()
	time.Sleep(100 * time.Millisecond)

	largeCode := ""
	for i := 0; i < 1000; i++ {
		largeCode += "# This is a comment line " + string(rune(i)) + "\n"
	}
	largeCode += "print('Hello from large code')"

	largeCodeSubmission := types.SubmissionMessage{
		SubmissionID: 400,
		Language:     "PYTHON",
		Code:         base64.StdEncoding.EncodeToString([]byte(largeCode)),
		TimeLimit:    5.0,
		MemoryLimit:  512,
		TestCases: []types.TestCaseMessage{
			{
				TestCaseID:     "tc1",
				Input:          base64.StdEncoding.EncodeToString([]byte("")),
				ExpectedOutput: base64.StdEncoding.EncodeToString([]byte("Hello from large code")),
			},
		},
	}

	mqClient.addSubmission(largeCodeSubmission)

	time.Sleep(10 * time.Second)

	messages := mqClient.getMessages()
	
	resultMessages := []errorMessage{}
	for _, msg := range messages {
		if msg.exchange == rabbitmq.ResultExchange {
			resultMessages = append(resultMessages, msg)
		}
	}

	if len(resultMessages) != 1 {
		t.Fatalf("Expected 1 result message, got %d", len(resultMessages))
	}

	result := resultMessages[0].body.(types.ResultNotificationMessage)
	if result.Status != "PASSED" {
		t.Errorf("Large code result = %s, want PASSED", result.Status)
	}
}

func TestErrorScenarios_CompilerSpecificErrors(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	mqClient := &errorScenarioMQClient{
		consumeQueue: make(chan amqp091.Delivery, 10),
	}

	m, err := master.NewMaster(mqClient, 1, "test.queue")
	if err != nil {
		t.Fatalf("Failed to create master: %v", err)
	}

	m.Start()
	time.Sleep(100 * time.Millisecond)

	javaCompileError := types.SubmissionMessage{
		SubmissionID: 500,
		Language:     "JAVA",
		Code:         base64.StdEncoding.EncodeToString([]byte("public class Main { invalid syntax }")),
		TimeLimit:    2.0,
		MemoryLimit:  256,
		TestCases: []types.TestCaseMessage{
			{
				TestCaseID:     "tc1",
				Input:          base64.StdEncoding.EncodeToString([]byte("")),
				ExpectedOutput: base64.StdEncoding.EncodeToString([]byte("hello")),
			},
		},
	}

	cppCompileError := types.SubmissionMessage{
		SubmissionID: 501,
		Language:     "CPP",
		Code:         base64.StdEncoding.EncodeToString([]byte("#include <iostream>\nint main() { undefined_function(); }")),
		TimeLimit:    2.0,
		MemoryLimit:  256,
		TestCases: []types.TestCaseMessage{
			{
				TestCaseID:     "tc1",
				Input:          base64.StdEncoding.EncodeToString([]byte("")),
				ExpectedOutput: base64.StdEncoding.EncodeToString([]byte("hello")),
			},
		},
	}

	pythonSyntaxError := types.SubmissionMessage{
		SubmissionID: 502,
		Language:     "PYTHON",
		Code:         base64.StdEncoding.EncodeToString([]byte("print('hello'")),
		TimeLimit:    2.0,
		MemoryLimit:  256,
		TestCases: []types.TestCaseMessage{
			{
				TestCaseID:     "tc1",
				Input:          base64.StdEncoding.EncodeToString([]byte("")),
				ExpectedOutput: base64.StdEncoding.EncodeToString([]byte("hello")),
			},
		},
	}

	mqClient.addSubmission(javaCompileError)
	mqClient.addSubmission(cppCompileError)
	mqClient.addSubmission(pythonSyntaxError)

	time.Sleep(15 * time.Second)

	messages := mqClient.getMessages()
	
	resultMessages := []errorMessage{}
	for _, msg := range messages {
		if msg.exchange == rabbitmq.ResultExchange {
			resultMessages = append(resultMessages, msg)
		}
	}

	if len(resultMessages) != 3 {
		t.Fatalf("Expected 3 result messages, got %d", len(resultMessages))
	}

	submissionResults := make(map[int64]string)
	for _, msg := range resultMessages {
		result := msg.body.(types.ResultNotificationMessage)
		submissionResults[result.SubmissionID] = result.Status
	}

	expectedErrors := []int64{500, 501, 502}
	for _, id := range expectedErrors {
		status := submissionResults[id]
		if status != "COMPILATION_ERROR" && status != "RUNTIME_ERROR" {
			t.Errorf("Submission %d status = %s, want COMPILATION_ERROR or RUNTIME_ERROR", id, status)
		}
	}
}

func TestErrorScenarios_ZeroTimeLimit(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	mqClient := &errorScenarioMQClient{
		consumeQueue: make(chan amqp091.Delivery, 10),
	}

	m, err := master.NewMaster(mqClient, 1, "test.queue")
	if err != nil {
		t.Fatalf("Failed to create master: %v", err)
	}

	m.Start()
	time.Sleep(100 * time.Millisecond)

	zeroTimeLimitSubmission := testutil.CreatePythonHelloWorldSubmission()
	zeroTimeLimitSubmission.SubmissionID = 600
	zeroTimeLimitSubmission.TimeLimit = 0.0

	mqClient.addSubmission(zeroTimeLimitSubmission)

	time.Sleep(5 * time.Second)

	messages := mqClient.getMessages()
	
	resultMessages := []errorMessage{}
	for _, msg := range messages {
		if msg.exchange == rabbitmq.ResultExchange {
			resultMessages = append(resultMessages, msg)
		}
	}

	if len(resultMessages) != 1 {
		t.Fatalf("Expected 1 result message, got %d", len(resultMessages))
	}

	result := resultMessages[0].body.(types.ResultNotificationMessage)
	if result.Status != "TIME_LIMIT_EXCEEDED" {
		t.Errorf("Zero time limit result = %s, want TIME_LIMIT_EXCEEDED", result.Status)
	}
}

func TestErrorScenarios_NoTestCases(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	mqClient := &errorScenarioMQClient{
		consumeQueue: make(chan amqp091.Delivery, 10),
	}

	m, err := master.NewMaster(mqClient, 1, "test.queue")
	if err != nil {
		t.Fatalf("Failed to create master: %v", err)
	}

	m.Start()
	time.Sleep(100 * time.Millisecond)

	noTestCasesSubmission := types.SubmissionMessage{
		SubmissionID: 700,
		Language:     "PYTHON",
		Code:         base64.StdEncoding.EncodeToString([]byte("print('hello')")),
		TimeLimit:    1.0,
		MemoryLimit:  128,
		TestCases:    []types.TestCaseMessage{},
	}

	mqClient.addSubmission(noTestCasesSubmission)

	time.Sleep(5 * time.Second)

	messages := mqClient.getMessages()
	
	resultMessages := []errorMessage{}
	for _, msg := range messages {
		if msg.exchange == rabbitmq.ResultExchange {
			resultMessages = append(resultMessages, msg)
		}
	}

	if len(resultMessages) != 1 {
		t.Fatalf("Expected 1 result message, got %d", len(resultMessages))
	}

	result := resultMessages[0].body.(types.ResultNotificationMessage)
	if result.Status != "COMPILATION_ERROR" {
		t.Errorf("No test cases result = %s, want COMPILATION_ERROR", result.Status)
	}
	if len(result.Results) != 0 {
		t.Errorf("Expected 0 test case results, got %d", len(result.Results))
	}
}