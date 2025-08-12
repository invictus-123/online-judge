package integration

import (
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

type endToEndMQClient struct {
	publishedMessages []e2eMessage
	consumeQueue      chan amqp091.Delivery
	mu                sync.Mutex
}

type e2eMessage struct {
	exchange   string
	routingKey string
	body       interface{}
	timestamp  time.Time
}

func (m *endToEndMQClient) Publish(exchange, routingKey string, body interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.publishedMessages = append(m.publishedMessages, e2eMessage{
		exchange:   exchange,
		routingKey: routingKey,
		body:       body,
		timestamp:  time.Now(),
	})
	return nil
}

func (m *endToEndMQClient) ConsumeSubmissions(queueName string) (<-chan amqp091.Delivery, error) {
	return m.consumeQueue, nil
}

func (m *endToEndMQClient) getMessages() []e2eMessage {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]e2eMessage{}, m.publishedMessages...)
}

func (m *endToEndMQClient) reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.publishedMessages = nil
}

func (m *endToEndMQClient) addSubmission(submission types.SubmissionMessage) {
	data, _ := json.Marshal(submission)
	delivery := amqp091.Delivery{Body: data}
	m.consumeQueue <- delivery
}

func TestEndToEnd_SingleSubmissionFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	mqClient := &endToEndMQClient{
		consumeQueue: make(chan amqp091.Delivery, 10),
	}

	m, err := master.NewMaster(mqClient, 1, "test.queue")
	if err != nil {
		t.Fatalf("Failed to create master: %v", err)
	}

	m.Start()
	time.Sleep(100 * time.Millisecond)

	submission := testutil.CreatePythonHelloWorldSubmission()
	mqClient.addSubmission(submission)

	time.Sleep(5 * time.Second)

	messages := mqClient.getMessages()
	if len(messages) != 2 {
		t.Fatalf("Expected 2 messages (status + result), got %d", len(messages))
	}

	statusMsg := messages[0]
	if statusMsg.exchange != rabbitmq.StatusExchange {
		t.Errorf("Status exchange = %s, want %s", statusMsg.exchange, rabbitmq.StatusExchange)
	}

	resultMsg := messages[1]
	if resultMsg.exchange != rabbitmq.ResultExchange {
		t.Errorf("Result exchange = %s, want %s", resultMsg.exchange, rabbitmq.ResultExchange)
	}

	result := resultMsg.body.(types.ResultNotificationMessage)
	if result.SubmissionID != submission.SubmissionID {
		t.Errorf("SubmissionID = %d, want %d", result.SubmissionID, submission.SubmissionID)
	}
	if result.Status != "PASSED" {
		t.Errorf("Status = %s, want PASSED", result.Status)
	}
}

func TestEndToEnd_MultipleSubmissionsSequential(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	mqClient := &endToEndMQClient{
		consumeQueue: make(chan amqp091.Delivery, 10),
	}

	m, err := master.NewMaster(mqClient, 1, "test.queue")
	if err != nil {
		t.Fatalf("Failed to create master: %v", err)
	}

	m.Start()
	time.Sleep(100 * time.Millisecond)

	submissions := []types.SubmissionMessage{
		testutil.CreatePythonHelloWorldSubmission(),
		testutil.CreateAdditionSubmission(),
		testutil.CreateWrongAnswerSubmission(),
	}

	for _, submission := range submissions {
		mqClient.addSubmission(submission)
	}

	time.Sleep(10 * time.Second)

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

	if statusCount != 3 {
		t.Errorf("Expected 3 status messages, got %d", statusCount)
	}
	if resultCount != 3 {
		t.Errorf("Expected 3 result messages, got %d", resultCount)
	}

	submissionIDs := make(map[int64]bool)
	for _, msg := range messages {
		if msg.exchange == rabbitmq.ResultExchange {
			result := msg.body.(types.ResultNotificationMessage)
			submissionIDs[result.SubmissionID] = true
		}
	}

	if len(submissionIDs) != 3 {
		t.Errorf("Expected 3 unique submission IDs, got %d", len(submissionIDs))
	}
}

func TestEndToEnd_MultipleWorkersParallel(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	mqClient := &endToEndMQClient{
		consumeQueue: make(chan amqp091.Delivery, 10),
	}

	m, err := master.NewMaster(mqClient, 3, "test.queue")
	if err != nil {
		t.Fatalf("Failed to create master: %v", err)
	}

	m.Start()
	time.Sleep(100 * time.Millisecond)

	submissions := []types.SubmissionMessage{
		testutil.CreatePythonHelloWorldSubmission(),
		testutil.CreateJavaHelloWorldSubmission(),
		testutil.CreateCppHelloWorldSubmission(),
		testutil.CreateAdditionSubmission(),
		testutil.CreateFibonacciSubmission(),
	}

	for i, submission := range submissions {
		submission.SubmissionID = int64(i + 100)
		mqClient.addSubmission(submission)
	}

	time.Sleep(15 * time.Second)

	messages := mqClient.getMessages()
	
	resultMessages := []e2eMessage{}
	for _, msg := range messages {
		if msg.exchange == rabbitmq.ResultExchange {
			resultMessages = append(resultMessages, msg)
		}
	}

	if len(resultMessages) != 5 {
		t.Fatalf("Expected 5 result messages, got %d", len(resultMessages))
	}

	submissionResults := make(map[int64]string)
	for _, msg := range resultMessages {
		result := msg.body.(types.ResultNotificationMessage)
		submissionResults[result.SubmissionID] = result.Status
	}

	expectedResults := map[int64]string{
		100: "PASSED", // Python Hello World
		101: "PASSED", // Java Hello World  
		102: "PASSED", // C++ Hello World
		103: "PASSED", // Addition
		104: "PASSED", // Fibonacci
	}

	for id, expectedStatus := range expectedResults {
		if actualStatus, exists := submissionResults[id]; !exists {
			t.Errorf("Missing result for submission %d", id)
		} else if actualStatus != expectedStatus {
			t.Errorf("Submission %d status = %s, want %s", id, actualStatus, expectedStatus)
		}
	}
}

func TestEndToEnd_MixedSuccessFailure(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	mqClient := &endToEndMQClient{
		consumeQueue: make(chan amqp091.Delivery, 10),
	}

	m, err := master.NewMaster(mqClient, 2, "test.queue")
	if err != nil {
		t.Fatalf("Failed to create master: %v", err)
	}

	m.Start()
	time.Sleep(100 * time.Millisecond)

	submissions := []types.SubmissionMessage{
		testutil.CreatePythonHelloWorldSubmission(),
		testutil.CreateWrongAnswerSubmission(),
		testutil.CreateCompilationErrorSubmission(),
		testutil.CreateRuntimeErrorSubmission(),
		testutil.CreateInfiniteLoopSubmission(),
	}

	for i, submission := range submissions {
		submission.SubmissionID = int64(i + 200)
		mqClient.addSubmission(submission)
	}

	time.Sleep(15 * time.Second)

	messages := mqClient.getMessages()
	
	resultMessages := []e2eMessage{}
	for _, msg := range messages {
		if msg.exchange == rabbitmq.ResultExchange {
			resultMessages = append(resultMessages, msg)
		}
	}

	if len(resultMessages) != 5 {
		t.Fatalf("Expected 5 result messages, got %d", len(resultMessages))
	}

	submissionResults := make(map[int64]string)
	for _, msg := range resultMessages {
		result := msg.body.(types.ResultNotificationMessage)
		submissionResults[result.SubmissionID] = result.Status
	}

	expectedResults := map[int64]string{
		200: "PASSED",             // Python Hello World
		201: "WRONG_ANSWER",       // Wrong Answer
		202: "COMPILATION_ERROR",  // Compilation Error
		203: "RUNTIME_ERROR",      // Runtime Error
		204: "TIME_LIMIT_EXCEEDED", // Infinite Loop
	}

	for id, expectedStatus := range expectedResults {
		if actualStatus, exists := submissionResults[id]; !exists {
			t.Errorf("Missing result for submission %d", id)
		} else if actualStatus != expectedStatus {
			t.Errorf("Submission %d status = %s, want %s", id, actualStatus, expectedStatus)
		}
	}
}

func TestEndToEnd_LargeVolumeProcessing(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	mqClient := &endToEndMQClient{
		consumeQueue: make(chan amqp091.Delivery, 50),
	}

	m, err := master.NewMaster(mqClient, 3, "test.queue")
	if err != nil {
		t.Fatalf("Failed to create master: %v", err)
	}

	m.Start()
	time.Sleep(100 * time.Millisecond)

	const numSubmissions = 20
	submissionIDs := make(map[int64]bool)

	for i := 0; i < numSubmissions; i++ {
		var submission types.SubmissionMessage
		switch i % 3 {
		case 0:
			submission = testutil.CreatePythonHelloWorldSubmission()
		case 1:
			submission = testutil.CreateAdditionSubmission()
		case 2:
			submission = testutil.CreateMultipleTestCasesSubmission()
		}
		submission.SubmissionID = int64(i + 300)
		submissionIDs[submission.SubmissionID] = true
		mqClient.addSubmission(submission)
	}

	time.Sleep(30 * time.Second)

	messages := mqClient.getMessages()
	
	resultMessages := []e2eMessage{}
	for _, msg := range messages {
		if msg.exchange == rabbitmq.ResultExchange {
			resultMessages = append(resultMessages, msg)
		}
	}

	if len(resultMessages) != numSubmissions {
		t.Fatalf("Expected %d result messages, got %d", numSubmissions, len(resultMessages))
	}

	processedIDs := make(map[int64]bool)
	allPassed := true
	for _, msg := range resultMessages {
		result := msg.body.(types.ResultNotificationMessage)
		processedIDs[result.SubmissionID] = true
		if result.Status != "PASSED" {
			allPassed = false
			t.Logf("Submission %d failed with status %s", result.SubmissionID, result.Status)
		}
	}

	if len(processedIDs) != numSubmissions {
		t.Errorf("Expected %d unique processed submissions, got %d", numSubmissions, len(processedIDs))
	}

	for id := range submissionIDs {
		if !processedIDs[id] {
			t.Errorf("Submission %d was not processed", id)
		}
	}

	if !allPassed {
		t.Error("Not all submissions passed as expected")
	}
}

func TestEndToEnd_ResourceLimitEnforcement(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	mqClient := &endToEndMQClient{
		consumeQueue: make(chan amqp091.Delivery, 10),
	}

	m, err := master.NewMaster(mqClient, 1, "test.queue")
	if err != nil {
		t.Fatalf("Failed to create master: %v", err)
	}

	m.Start()
	time.Sleep(100 * time.Millisecond)

	timeoutSubmission := testutil.CreateInfiniteLoopSubmission()
	timeoutSubmission.SubmissionID = 400
	timeoutSubmission.TimeLimit = 0.5

	memorySubmission := testutil.CreatePythonHelloWorldSubmission()
	memorySubmission.SubmissionID = 401
	memorySubmission.MemoryLimit = 1

	mqClient.addSubmission(timeoutSubmission)
	mqClient.addSubmission(memorySubmission)

	time.Sleep(10 * time.Second)

	messages := mqClient.getMessages()
	
	resultMessages := []e2eMessage{}
	for _, msg := range messages {
		if msg.exchange == rabbitmq.ResultExchange {
			resultMessages = append(resultMessages, msg)
		}
	}

	if len(resultMessages) != 2 {
		t.Fatalf("Expected 2 result messages, got %d", len(resultMessages))
	}

	submissionResults := make(map[int64]string)
	for _, msg := range resultMessages {
		result := msg.body.(types.ResultNotificationMessage)
		submissionResults[result.SubmissionID] = result.Status
	}

	if submissionResults[400] != "TIME_LIMIT_EXCEEDED" {
		t.Errorf("Timeout submission status = %s, want TIME_LIMIT_EXCEEDED", submissionResults[400])
	}

	if submissionResults[401] != "PASSED" {
		t.Errorf("Low memory submission status = %s, want PASSED", submissionResults[401])
	}
}

func TestEndToEnd_MessageTiming(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	mqClient := &endToEndMQClient{
		consumeQueue: make(chan amqp091.Delivery, 10),
	}

	m, err := master.NewMaster(mqClient, 1, "test.queue")
	if err != nil {
		t.Fatalf("Failed to create master: %v", err)
	}

	m.Start()
	time.Sleep(100 * time.Millisecond)

	submission := testutil.CreatePythonHelloWorldSubmission()
	submission.SubmissionID = 500

	startTime := time.Now()
	mqClient.addSubmission(submission)

	time.Sleep(5 * time.Second)

	messages := mqClient.getMessages()
	
	var statusTime, resultTime time.Time
	for _, msg := range messages {
		if msg.exchange == rabbitmq.StatusExchange {
			statusTime = msg.timestamp
		} else if msg.exchange == rabbitmq.ResultExchange {
			resultTime = msg.timestamp
		}
	}

	if statusTime.IsZero() {
		t.Error("Status message not found")
	}
	if resultTime.IsZero() {
		t.Error("Result message not found")
	}

	if statusTime.Before(startTime) {
		t.Error("Status message should come after submission start")
	}
	if resultTime.Before(statusTime) {
		t.Error("Result message should come after status message")
	}

	processingTime := resultTime.Sub(statusTime)
	if processingTime > 10*time.Second {
		t.Errorf("Processing time %v seems too long", processingTime)
	}
}

func TestEndToEnd_QueueManagement(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	mqClient := &endToEndMQClient{
		consumeQueue: make(chan amqp091.Delivery, 5),
	}

	m, err := master.NewMaster(mqClient, 1, "test.queue")
	if err != nil {
		t.Fatalf("Failed to create master: %v", err)
	}

	m.Start()
	time.Sleep(100 * time.Millisecond)

	submissions := make([]types.SubmissionMessage, 5)
	for i := 0; i < 5; i++ {
		submissions[i] = testutil.CreatePythonHelloWorldSubmission()
		submissions[i].SubmissionID = int64(i + 600)
		mqClient.addSubmission(submissions[i])
	}

	time.Sleep(15 * time.Second)

	messages := mqClient.getMessages()
	
	resultCount := 0
	for _, msg := range messages {
		if msg.exchange == rabbitmq.ResultExchange {
			resultCount++
		}
	}

	if resultCount != 5 {
		t.Errorf("Expected 5 results, got %d", resultCount)
	}
}