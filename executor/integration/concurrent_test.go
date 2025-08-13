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

type concurrentMQClient struct {
	publishedMessages []concurrentMessage
	consumeQueue      chan amqp091.Delivery
	mu                sync.Mutex
	closed            bool
}

type concurrentMessage struct {
	exchange   string
	routingKey string
	body       interface{}
	timestamp  time.Time
	workerID   int
}

func (m *concurrentMQClient) Publish(exchange, routingKey string, body interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.publishedMessages = append(m.publishedMessages, concurrentMessage{
		exchange:   exchange,
		routingKey: routingKey,
		body:       body,
		timestamp:  time.Now(),
	})
	return nil
}

func (m *concurrentMQClient) ConsumeSubmissions(queueName string) (<-chan amqp091.Delivery, error) {
	return m.consumeQueue, nil
}

func (m *concurrentMQClient) getMessages() []concurrentMessage {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]concurrentMessage{}, m.publishedMessages...)
}

func (m *concurrentMQClient) reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.publishedMessages = nil
}

func (m *concurrentMQClient) close() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if !m.closed {
		close(m.consumeQueue)
		m.closed = true
	}
}

func (m *concurrentMQClient) addSubmissions(submissions []types.SubmissionMessage) {
	for _, submission := range submissions {
		data, _ := json.Marshal(submission)
		delivery := amqp091.Delivery{Body: data}
		
		// Provide proper ack/nack functionality
		delivery.Acknowledger = &mockAcknowledger{}
		
		m.consumeQueue <- delivery
	}
}

type mockAcknowledger struct{}

func (a *mockAcknowledger) Ack(tag uint64, multiple bool) error {
	return nil
}

func (a *mockAcknowledger) Nack(tag uint64, multiple bool, requeue bool) error {
	return nil
}

func (a *mockAcknowledger) Reject(tag uint64, requeue bool) error {
	return nil
}

func TestConcurrent_MultipleWorkersParallelProcessing(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	mqClient := &concurrentMQClient{
		consumeQueue: make(chan amqp091.Delivery, 100),
	}
	defer mqClient.close()

	workerCount := 5
	m, err := master.NewMaster(mqClient, workerCount, "test.queue")
	if err != nil {
		t.Fatalf("Failed to create master: %v", err)
	}

	m.Start()
	time.Sleep(500 * time.Millisecond) // Give more time for workers to start

	numSubmissions := 20
	submissions := make([]types.SubmissionMessage, numSubmissions)
	for i := 0; i < numSubmissions; i++ {
		submissions[i] = testutil.CreatePythonHelloWorldSubmission()
		submissions[i].SubmissionID = int64(i + 1000)
	}

	startTime := time.Now()
	mqClient.addSubmissions(submissions)

	// Wait for all submissions to be processed with timeout
	timeout := time.After(30 * time.Second)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			t.Fatalf("Timeout waiting for submissions to complete")
		case <-ticker.C:
			messages := mqClient.getMessages()
			resultCount := 0
			for _, msg := range messages {
				if msg.exchange == rabbitmq.ResultExchange {
					resultCount++
				}
			}
			
			if resultCount >= numSubmissions {
				goto checkResults // Break out of the select/for loop
			}
		}
	}

checkResults:
	messages := mqClient.getMessages()
	
	resultMessages := []concurrentMessage{}
	statusMessages := []concurrentMessage{}
	for _, msg := range messages {
		if msg.exchange == rabbitmq.ResultExchange {
			resultMessages = append(resultMessages, msg)
		} else if msg.exchange == rabbitmq.StatusExchange {
			statusMessages = append(statusMessages, msg)
		}
	}

	if len(resultMessages) != numSubmissions {
		t.Errorf("Expected %d result messages, got %d", numSubmissions, len(resultMessages))
	}
	if len(statusMessages) != numSubmissions {
		t.Errorf("Expected %d status messages, got %d", numSubmissions, len(statusMessages))
	}

	submissionResults := make(map[int64]string)
	completionTimes := make(map[int64]time.Time)
	for _, msg := range resultMessages {
		result := msg.body.(types.ResultNotificationMessage)
		submissionResults[result.SubmissionID] = result.Status
		completionTimes[result.SubmissionID] = msg.timestamp
	}

	for i := 0; i < numSubmissions; i++ {
		submissionID := int64(i + 1000)
		if status, exists := submissionResults[submissionID]; !exists {
			t.Errorf("Missing result for submission %d", submissionID)
		} else if status != "PASSED" {
			t.Errorf("Submission %d status = %s, want PASSED", submissionID, status)
		}
	}

	totalTime := time.Since(startTime)
	expectedMaxTime := time.Duration(numSubmissions/workerCount+2) * time.Second
	if totalTime > expectedMaxTime {
		t.Errorf("Processing took %v, expected less than %v with %d workers", totalTime, expectedMaxTime, workerCount)
	}
}

func TestConcurrent_RaceConditionHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	mqClient := &concurrentMQClient{
		consumeQueue: make(chan amqp091.Delivery, 50),
	}

	m, err := master.NewMaster(mqClient, 3, "test.queue")
	if err != nil {
		t.Fatalf("Failed to create master: %v", err)
	}

	m.Start()
	time.Sleep(100 * time.Millisecond)

	numSubmissions := 15
	submissions := make([]types.SubmissionMessage, numSubmissions)
	for i := 0; i < numSubmissions; i++ {
		submissions[i] = testutil.CreateAdditionSubmission()
		submissions[i].SubmissionID = int64(i + 2000)
	}

	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(offset int) {
			defer wg.Done()
			for j := 0; j < 5; j++ {
				submission := submissions[offset*5+j]
				data, _ := json.Marshal(submission)
				delivery := amqp091.Delivery{Body: data}
				mqClient.consumeQueue <- delivery
				time.Sleep(10 * time.Millisecond)
			}
		}(i)
	}

	wg.Wait()
	time.Sleep(15 * time.Second)

	messages := mqClient.getMessages()
	
	resultMessages := []concurrentMessage{}
	for _, msg := range messages {
		if msg.exchange == rabbitmq.ResultExchange {
			resultMessages = append(resultMessages, msg)
		}
	}

	if len(resultMessages) != numSubmissions {
		t.Errorf("Expected %d result messages, got %d", numSubmissions, len(resultMessages))
	}

	submissionIDs := make(map[int64]bool)
	for _, msg := range resultMessages {
		result := msg.body.(types.ResultNotificationMessage)
		if submissionIDs[result.SubmissionID] {
			t.Errorf("Duplicate result for submission %d", result.SubmissionID)
		}
		submissionIDs[result.SubmissionID] = true
		
		if result.Status != "PASSED" {
			t.Errorf("Submission %d status = %s, want PASSED", result.SubmissionID, result.Status)
		}
	}
}

func TestConcurrent_MixedWorkloadDistribution(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	mqClient := &concurrentMQClient{
		consumeQueue: make(chan amqp091.Delivery, 50),
	}

	m, err := master.NewMaster(mqClient, 4, "test.queue")
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
		testutil.CreateWrongAnswerSubmission(),
		testutil.CreateCompilationErrorSubmission(),
		testutil.CreateRuntimeErrorSubmission(),
		testutil.CreateMultipleTestCasesSubmission(),
		testutil.CreateLargeInputSubmission(),
	}

	for i, submission := range submissions {
		submission.SubmissionID = int64(i + 3000)
		submissions[i] = submission
	}

	mqClient.addSubmissions(submissions)

	time.Sleep(20 * time.Second)

	messages := mqClient.getMessages()
	
	resultMessages := []concurrentMessage{}
	for _, msg := range messages {
		if msg.exchange == rabbitmq.ResultExchange {
			resultMessages = append(resultMessages, msg)
		}
	}

	if len(resultMessages) != len(submissions) {
		t.Errorf("Expected %d result messages, got %d", len(submissions), len(resultMessages))
	}

	expectedResults := map[int64]string{
		3000: "PASSED",             // Python Hello World
		3001: "PASSED",             // Java Hello World
		3002: "PASSED",             // C++ Hello World
		3003: "PASSED",             // Addition
		3004: "PASSED",             // Fibonacci
		3005: "WRONG_ANSWER",       // Wrong Answer
		3006: "COMPILATION_ERROR",  // Compilation Error
		3007: "RUNTIME_ERROR",      // Runtime Error
		3008: "PASSED",             // Multiple Test Cases
		3009: "PASSED",             // Large Input
	}

	submissionResults := make(map[int64]string)
	for _, msg := range resultMessages {
		result := msg.body.(types.ResultNotificationMessage)
		submissionResults[result.SubmissionID] = result.Status
	}

	for id, expectedStatus := range expectedResults {
		if actualStatus, exists := submissionResults[id]; !exists {
			t.Errorf("Missing result for submission %d", id)
		} else if actualStatus != expectedStatus {
			t.Errorf("Submission %d status = %s, want %s", id, actualStatus, expectedStatus)
		}
	}
}

func TestConcurrent_HighVolumeStressTest(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	mqClient := &concurrentMQClient{
		consumeQueue: make(chan amqp091.Delivery, 200),
	}

	workerCount := 5
	m, err := master.NewMaster(mqClient, workerCount, "test.queue")
	if err != nil {
		t.Fatalf("Failed to create master: %v", err)
	}

	m.Start()
	time.Sleep(200 * time.Millisecond)

	numSubmissions := 50
	submissions := make([]types.SubmissionMessage, numSubmissions)
	for i := 0; i < numSubmissions; i++ {
		switch i % 4 {
		case 0:
			submissions[i] = testutil.CreatePythonHelloWorldSubmission()
		case 1:
			submissions[i] = testutil.CreateAdditionSubmission()
		case 2:
			submissions[i] = testutil.CreateMultipleTestCasesSubmission()
		case 3:
			submissions[i] = testutil.CreateFibonacciSubmission()
		}
		submissions[i].SubmissionID = int64(i + 4000)
	}

	startTime := time.Now()
	
	batchSize := 10
	for i := 0; i < numSubmissions; i += batchSize {
		end := i + batchSize
		if end > numSubmissions {
			end = numSubmissions
		}
		mqClient.addSubmissions(submissions[i:end])
		time.Sleep(100 * time.Millisecond)
	}

	time.Sleep(45 * time.Second)

	messages := mqClient.getMessages()
	
	resultMessages := []concurrentMessage{}
	statusMessages := []concurrentMessage{}
	for _, msg := range messages {
		if msg.exchange == rabbitmq.ResultExchange {
			resultMessages = append(resultMessages, msg)
		} else if msg.exchange == rabbitmq.StatusExchange {
			statusMessages = append(statusMessages, msg)
		}
	}

	if len(resultMessages) != numSubmissions {
		t.Errorf("Expected %d result messages, got %d", numSubmissions, len(resultMessages))
	}

	successCount := 0
	submissionIDs := make(map[int64]bool)
	for _, msg := range resultMessages {
		result := msg.body.(types.ResultNotificationMessage)
		
		if submissionIDs[result.SubmissionID] {
			t.Errorf("Duplicate result for submission %d", result.SubmissionID)
		}
		submissionIDs[result.SubmissionID] = true
		
		if result.Status == "PASSED" {
			successCount++
		}
	}

	if successCount != numSubmissions {
		t.Errorf("Expected all %d submissions to pass, got %d successes", numSubmissions, successCount)
	}

	processingTime := time.Since(startTime)
	t.Logf("Processed %d submissions in %v with %d workers", numSubmissions, processingTime, workerCount)
}

func TestConcurrent_WorkerIsolation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	mqClient := &concurrentMQClient{
		consumeQueue: make(chan amqp091.Delivery, 20),
	}

	m, err := master.NewMaster(mqClient, 3, "test.queue")
	if err != nil {
		t.Fatalf("Failed to create master: %v", err)
	}

	m.Start()
	time.Sleep(100 * time.Millisecond)

	problematicSubmissions := []types.SubmissionMessage{
		testutil.CreateInfiniteLoopSubmission(),
		testutil.CreateCompilationErrorSubmission(),
		testutil.CreateRuntimeErrorSubmission(),
	}

	normalSubmissions := []types.SubmissionMessage{
		testutil.CreatePythonHelloWorldSubmission(),
		testutil.CreateAdditionSubmission(),
		testutil.CreateFibonacciSubmission(),
	}

	for i, submission := range problematicSubmissions {
		submission.SubmissionID = int64(i + 5000)
		problematicSubmissions[i] = submission
	}

	for i, submission := range normalSubmissions {
		submission.SubmissionID = int64(i + 5100)
		normalSubmissions[i] = submission
	}

	mqClient.addSubmissions(problematicSubmissions)
	time.Sleep(500 * time.Millisecond)
	mqClient.addSubmissions(normalSubmissions)

	time.Sleep(15 * time.Second)

	messages := mqClient.getMessages()
	
	resultMessages := []concurrentMessage{}
	for _, msg := range messages {
		if msg.exchange == rabbitmq.ResultExchange {
			resultMessages = append(resultMessages, msg)
		}
	}

	if len(resultMessages) != 6 {
		t.Fatalf("Expected 6 result messages, got %d", len(resultMessages))
	}

	problematicResults := make(map[int64]string)
	normalResults := make(map[int64]string)

	for _, msg := range resultMessages {
		result := msg.body.(types.ResultNotificationMessage)
		if result.SubmissionID >= 5000 && result.SubmissionID < 5100 {
			problematicResults[result.SubmissionID] = result.Status
		} else if result.SubmissionID >= 5100 {
			normalResults[result.SubmissionID] = result.Status
		}
	}

	expectedProblematic := map[int64]string{
		5000: "TIME_LIMIT_EXCEEDED",
		5001: "COMPILATION_ERROR",
		5002: "RUNTIME_ERROR",
	}

	for id, expectedStatus := range expectedProblematic {
		if actualStatus := problematicResults[id]; actualStatus != expectedStatus {
			t.Errorf("Problematic submission %d status = %s, want %s", id, actualStatus, expectedStatus)
		}
	}

	for id, status := range normalResults {
		if status != "PASSED" {
			t.Errorf("Normal submission %d should pass despite problematic submissions, got %s", id, status)
		}
	}
}

func TestConcurrent_OrderIndependence(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	mqClient := &concurrentMQClient{
		consumeQueue: make(chan amqp091.Delivery, 30),
	}

	m, err := master.NewMaster(mqClient, 3, "test.queue")
	if err != nil {
		t.Fatalf("Failed to create master: %v", err)
	}

	m.Start()
	time.Sleep(100 * time.Millisecond)

	submissions := []types.SubmissionMessage{
		testutil.CreatePythonHelloWorldSubmission(),
		testutil.CreateAdditionSubmission(),
		testutil.CreateFibonacciSubmission(),
		testutil.CreateMultipleTestCasesSubmission(),
		testutil.CreateWrongAnswerSubmission(),
		testutil.CreatePythonHelloWorldSubmission(),
	}

	for i, submission := range submissions {
		submission.SubmissionID = int64(i + 6000)
		submissions[i] = submission
	}

	for i := 0; i < 3; i++ {
		mqClient.reset()
		
		shuffledSubmissions := make([]types.SubmissionMessage, len(submissions))
		copy(shuffledSubmissions, submissions)
		
		switch i {
		case 1:
			for j := 0; j < len(shuffledSubmissions)/2; j++ {
				k := len(shuffledSubmissions) - 1 - j
				shuffledSubmissions[j], shuffledSubmissions[k] = shuffledSubmissions[k], shuffledSubmissions[j]
			}
		case 2:
			for j := 1; j < len(shuffledSubmissions); j += 2 {
				if j-1 >= 0 {
					shuffledSubmissions[j], shuffledSubmissions[j-1] = shuffledSubmissions[j-1], shuffledSubmissions[j]
				}
			}
		}

		mqClient.addSubmissions(shuffledSubmissions)
		time.Sleep(15 * time.Second)

		messages := mqClient.getMessages()
		
		resultMessages := []concurrentMessage{}
		for _, msg := range messages {
			if msg.exchange == rabbitmq.ResultExchange {
				resultMessages = append(resultMessages, msg)
			}
		}

		if len(resultMessages) != len(submissions) {
			t.Errorf("Run %d: Expected %d result messages, got %d", i+1, len(submissions), len(resultMessages))
		}

		expectedResults := map[int64]string{
			6000: "PASSED",       // Python Hello World
			6001: "PASSED",       // Addition
			6002: "PASSED",       // Fibonacci
			6003: "PASSED",       // Multiple Test Cases
			6004: "WRONG_ANSWER", // Wrong Answer
			6005: "PASSED",       // Python Hello World (duplicate)
		}

		submissionResults := make(map[int64]string)
		for _, msg := range resultMessages {
			result := msg.body.(types.ResultNotificationMessage)
			submissionResults[result.SubmissionID] = result.Status
		}

		for id, expectedStatus := range expectedResults {
			if actualStatus := submissionResults[id]; actualStatus != expectedStatus {
				t.Errorf("Run %d: Submission %d status = %s, want %s", i+1, id, actualStatus, expectedStatus)
			}
		}
	}
}

func TestConcurrent_ResourceContentionHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	mqClient := &concurrentMQClient{
		consumeQueue: make(chan amqp091.Delivery, 20),
	}

	m, err := master.NewMaster(mqClient, 2, "test.queue")
	if err != nil {
		t.Fatalf("Failed to create master: %v", err)
	}

	m.Start()
	time.Sleep(100 * time.Millisecond)

	submissions := make([]types.SubmissionMessage, 10)
	for i := 0; i < 10; i++ {
		submissions[i] = testutil.CreateLargeInputSubmission()
		submissions[i].SubmissionID = int64(i + 7000)
		submissions[i].TimeLimit = 10.0
		submissions[i].MemoryLimit = 512
	}

	startTime := time.Now()
	mqClient.addSubmissions(submissions)

	time.Sleep(60 * time.Second)

	messages := mqClient.getMessages()
	
	resultMessages := []concurrentMessage{}
	for _, msg := range messages {
		if msg.exchange == rabbitmq.ResultExchange {
			resultMessages = append(resultMessages, msg)
		}
	}

	if len(resultMessages) != 10 {
		t.Errorf("Expected 10 result messages, got %d", len(resultMessages))
	}

	successCount := 0
	for _, msg := range resultMessages {
		result := msg.body.(types.ResultNotificationMessage)
		if result.Status == "PASSED" {
			successCount++
		}
		
		if result.TimeTaken <= 0 {
			t.Errorf("Submission %d should have positive execution time", result.SubmissionID)
		}
		if result.MemoryUsed <= 0 {
			t.Errorf("Submission %d should have positive memory usage", result.SubmissionID)
		}
	}

	if successCount != 10 {
		t.Errorf("Expected all 10 large input submissions to pass, got %d", successCount)
	}

	totalTime := time.Since(startTime)
	t.Logf("Processed 10 resource-intensive submissions in %v", totalTime)
}