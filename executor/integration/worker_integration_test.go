package integration

import (
	"encoding/base64"
	"online-judge/executor/rabbitmq"
	"online-judge/executor/testutil"
	"online-judge/executor/types"
	"online-judge/executor/worker"
	"sync"
	"testing"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

type integrationMQClient struct {
	publishedMessages []integrationMessage
	mu                sync.Mutex
}

type integrationMessage struct {
	exchange   string
	routingKey string
	body       interface{}
	timestamp  time.Time
}

func (m *integrationMQClient) Publish(exchange, routingKey string, body interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.publishedMessages = append(m.publishedMessages, integrationMessage{
		exchange:   exchange,
		routingKey: routingKey,
		body:       body,
		timestamp:  time.Now(),
	})
	return nil
}

func (m *integrationMQClient) ConsumeSubmissions(queueName string) (<-chan amqp091.Delivery, error) {
	ch := make(chan amqp091.Delivery)
	close(ch)
	return ch, nil
}

func (m *integrationMQClient) getMessages() []integrationMessage {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]integrationMessage{}, m.publishedMessages...)
}

func (m *integrationMQClient) reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.publishedMessages = nil
}

func TestWorkerIntegration_PythonHelloWorld(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	mqClient := &integrationMQClient{}
	jobQueue := make(chan amqp091.Delivery, 1)
	w := worker.NewWorker(1, jobQueue, mqClient)

	submission := testutil.CreatePythonHelloWorldSubmission()
	delivery := testutil.CreateTestDelivery(submission)

	jobQueue <- delivery
	close(jobQueue)

	w.Start()

	time.Sleep(3 * time.Second)

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

	result, ok := resultMsg.body.(types.ResultNotificationMessage)
	if !ok {
		t.Fatal("Result message is not ResultNotificationMessage")
	}

	if result.Status != "PASSED" {
		t.Errorf("Overall status = %s, want PASSED", result.Status)
	}
	if len(result.Results) != 1 {
		t.Errorf("Test case count = %d, want 1", len(result.Results))
	}
	if result.Results[0].Status != "PASSED" {
		t.Errorf("Test case status = %s, want PASSED", result.Results[0].Status)
	}
}

func TestWorkerIntegration_JavaCompilationSuccess(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	mqClient := &integrationMQClient{}
	jobQueue := make(chan amqp091.Delivery, 1)
	w := worker.NewWorker(1, jobQueue, mqClient)

	submission := testutil.CreateJavaHelloWorldSubmission()
	delivery := testutil.CreateTestDelivery(submission)

	jobQueue <- delivery
	close(jobQueue)

	w.Start()

	time.Sleep(10 * time.Second)

	messages := mqClient.getMessages()
	if len(messages) != 2 {
		t.Fatalf("Expected 2 messages, got %d", len(messages))
	}

	result := messages[1].body.(types.ResultNotificationMessage)
	if result.Status != "PASSED" {
		t.Errorf("Java execution status = %s, want PASSED", result.Status)
	}
}

func TestWorkerIntegration_CppCompilationSuccess(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	mqClient := &integrationMQClient{}
	jobQueue := make(chan amqp091.Delivery, 1)
	w := worker.NewWorker(1, jobQueue, mqClient)

	submission := testutil.CreateCppHelloWorldSubmission()
	delivery := testutil.CreateTestDelivery(submission)

	jobQueue <- delivery
	close(jobQueue)

	w.Start()

	time.Sleep(10 * time.Second)

	messages := mqClient.getMessages()
	if len(messages) != 2 {
		t.Fatalf("Expected 2 messages, got %d", len(messages))
	}

	result := messages[1].body.(types.ResultNotificationMessage)
	if result.Status != "PASSED" {
		t.Errorf("C++ execution status = %s, want PASSED", result.Status)
	}
}

func TestWorkerIntegration_MultipleTestCases(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	mqClient := &integrationMQClient{}
	jobQueue := make(chan amqp091.Delivery, 1)
	w := worker.NewWorker(1, jobQueue, mqClient)

	submission := testutil.CreateAdditionSubmission()
	delivery := testutil.CreateTestDelivery(submission)

	jobQueue <- delivery
	close(jobQueue)

	w.Start()

	time.Sleep(5 * time.Second)

	messages := mqClient.getMessages()
	if len(messages) != 2 {
		t.Fatalf("Expected 2 messages, got %d", len(messages))
	}

	result := messages[1].body.(types.ResultNotificationMessage)
	if result.Status != "PASSED" {
		t.Errorf("Addition submission status = %s, want PASSED", result.Status)
	}
	if len(result.Results) != 3 {
		t.Errorf("Test case count = %d, want 3", len(result.Results))
	}

	for i, testResult := range result.Results {
		if testResult.Status != "PASSED" {
			t.Errorf("Test case %d status = %s, want PASSED", i+1, testResult.Status)
		}
	}
}

func TestWorkerIntegration_WrongAnswer(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	mqClient := &integrationMQClient{}
	jobQueue := make(chan amqp091.Delivery, 1)
	w := worker.NewWorker(1, jobQueue, mqClient)

	submission := testutil.CreateWrongAnswerSubmission()
	delivery := testutil.CreateTestDelivery(submission)

	jobQueue <- delivery
	close(jobQueue)

	w.Start()

	time.Sleep(3 * time.Second)

	messages := mqClient.getMessages()
	if len(messages) != 2 {
		t.Fatalf("Expected 2 messages, got %d", len(messages))
	}

	result := messages[1].body.(types.ResultNotificationMessage)
	if result.Status != "WRONG_ANSWER" {
		t.Errorf("Wrong answer status = %s, want WRONG_ANSWER", result.Status)
	}
	if result.Results[0].Status != "WRONG_ANSWER" {
		t.Errorf("Test case status = %s, want WRONG_ANSWER", result.Results[0].Status)
	}
}

func TestWorkerIntegration_CompilationError(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	mqClient := &integrationMQClient{}
	jobQueue := make(chan amqp091.Delivery, 1)
	w := worker.NewWorker(1, jobQueue, mqClient)

	submission := testutil.CreateCompilationErrorSubmission()
	delivery := testutil.CreateTestDelivery(submission)

	jobQueue <- delivery
	close(jobQueue)

	w.Start()

	time.Sleep(5 * time.Second)

	messages := mqClient.getMessages()
	if len(messages) != 2 {
		t.Fatalf("Expected 2 messages, got %d", len(messages))
	}

	result := messages[1].body.(types.ResultNotificationMessage)
	if result.Status != "COMPILATION_ERROR" {
		t.Errorf("Compilation error status = %s, want COMPILATION_ERROR", result.Status)
	}
	if result.Results[0].Status != "COMPILATION_ERROR" {
		t.Errorf("Test case status = %s, want COMPILATION_ERROR", result.Results[0].Status)
	}
}

func TestWorkerIntegration_RuntimeError(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	mqClient := &integrationMQClient{}
	jobQueue := make(chan amqp091.Delivery, 1)
	w := worker.NewWorker(1, jobQueue, mqClient)

	submission := testutil.CreateRuntimeErrorSubmission()
	delivery := testutil.CreateTestDelivery(submission)

	jobQueue <- delivery
	close(jobQueue)

	w.Start()

	time.Sleep(3 * time.Second)

	messages := mqClient.getMessages()
	if len(messages) != 2 {
		t.Fatalf("Expected 2 messages, got %d", len(messages))
	}

	result := messages[1].body.(types.ResultNotificationMessage)
	if result.Status != "RUNTIME_ERROR" {
		t.Errorf("Runtime error status = %s, want RUNTIME_ERROR", result.Status)
	}
	if result.Results[0].Status != "RUNTIME_ERROR" {
		t.Errorf("Test case status = %s, want RUNTIME_ERROR", result.Results[0].Status)
	}
}

func TestWorkerIntegration_TimeLimit(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	mqClient := &integrationMQClient{}
	jobQueue := make(chan amqp091.Delivery, 1)
	w := worker.NewWorker(1, jobQueue, mqClient)

	submission := testutil.CreateInfiniteLoopSubmission()
	delivery := testutil.CreateTestDelivery(submission)

	jobQueue <- delivery
	close(jobQueue)

	w.Start()

	time.Sleep(3 * time.Second)

	messages := mqClient.getMessages()
	if len(messages) != 2 {
		t.Fatalf("Expected 2 messages, got %d", len(messages))
	}

	result := messages[1].body.(types.ResultNotificationMessage)
	if result.Status != "TIME_LIMIT_EXCEEDED" {
		t.Errorf("Time limit status = %s, want TIME_LIMIT_EXCEEDED", result.Status)
	}
	if result.Results[0].Status != "TIME_LIMIT_EXCEEDED" {
		t.Errorf("Test case status = %s, want TIME_LIMIT_EXCEEDED", result.Results[0].Status)
	}
}

func TestWorkerIntegration_LargeInput(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	mqClient := &integrationMQClient{}
	jobQueue := make(chan amqp091.Delivery, 1)
	w := worker.NewWorker(1, jobQueue, mqClient)

	submission := testutil.CreateLargeInputSubmission()
	delivery := testutil.CreateTestDelivery(submission)

	jobQueue <- delivery
	close(jobQueue)

	w.Start()

	time.Sleep(10 * time.Second)

	messages := mqClient.getMessages()
	if len(messages) != 2 {
		t.Fatalf("Expected 2 messages, got %d", len(messages))
	}

	result := messages[1].body.(types.ResultNotificationMessage)
	if result.Status != "PASSED" {
		t.Errorf("Large input status = %s, want PASSED", result.Status)
	}
}

func TestWorkerIntegration_InvalidBase64Code(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	mqClient := &integrationMQClient{}
	jobQueue := make(chan amqp091.Delivery, 1)
	w := worker.NewWorker(1, jobQueue, mqClient)

	submission := testutil.CreateInvalidBase64Submission()
	delivery := testutil.CreateTestDelivery(submission)

	jobQueue <- delivery
	close(jobQueue)

	w.Start()

	time.Sleep(2 * time.Second)

	messages := mqClient.getMessages()
	if len(messages) != 1 {
		t.Fatalf("Expected 1 message (status only), got %d", len(messages))
	}

	statusMsg := messages[0]
	status := statusMsg.body.(types.StatusUpdateMessage)
	if status.Status != "RUNNING" {
		t.Errorf("Status = %s, want RUNNING", status.Status)
	}
}

func TestWorkerIntegration_FibonacciAlgorithm(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	mqClient := &integrationMQClient{}
	jobQueue := make(chan amqp091.Delivery, 1)
	w := worker.NewWorker(1, jobQueue, mqClient)

	submission := testutil.CreateFibonacciSubmission()
	delivery := testutil.CreateTestDelivery(submission)

	jobQueue <- delivery
	close(jobQueue)

	w.Start()

	time.Sleep(5 * time.Second)

	messages := mqClient.getMessages()
	if len(messages) != 2 {
		t.Fatalf("Expected 2 messages, got %d", len(messages))
	}

	result := messages[1].body.(types.ResultNotificationMessage)
	if result.Status != "PASSED" {
		t.Errorf("Fibonacci status = %s, want PASSED", result.Status)
	}
	if len(result.Results) != 4 {
		t.Errorf("Fibonacci test case count = %d, want 4", len(result.Results))
	}

	for i, testResult := range result.Results {
		if testResult.Status != "PASSED" {
			t.Errorf("Fibonacci test case %d status = %s, want PASSED", i+1, testResult.Status)
		}
	}
}

func TestWorkerIntegration_MixedResults(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	mqClient := &integrationMQClient{}
	jobQueue := make(chan amqp091.Delivery, 1)
	w := worker.NewWorker(1, jobQueue, mqClient)

	submission := testutil.CreateMixedResultsSubmission()
	delivery := testutil.CreateTestDelivery(submission)

	jobQueue <- delivery
	close(jobQueue)

	w.Start()

	time.Sleep(5 * time.Second)

	messages := mqClient.getMessages()
	if len(messages) != 2 {
		t.Fatalf("Expected 2 messages, got %d", len(messages))
	}

	result := messages[1].body.(types.ResultNotificationMessage)
	if result.Status != "WRONG_ANSWER" {
		t.Errorf("Mixed results status = %s, want WRONG_ANSWER", result.Status)
	}
	if len(result.Results) != 4 {
		t.Errorf("Mixed results test case count = %d, want 4", len(result.Results))
	}

	expectedStatuses := []string{"PASSED", "PASSED", "WRONG_ANSWER", "PASSED"}
	for i, testResult := range result.Results {
		if testResult.Status != expectedStatuses[i] {
			t.Errorf("Test case %d status = %s, want %s", i+1, testResult.Status, expectedStatuses[i])
		}
	}
}

func TestWorkerIntegration_ResourceUsageTracking(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	mqClient := &integrationMQClient{}
	jobQueue := make(chan amqp091.Delivery, 1)
	w := worker.NewWorker(1, jobQueue, mqClient)

	submission := types.SubmissionMessage{
		SubmissionID: 999,
		Language:     "PYTHON",
		Code:         base64.StdEncoding.EncodeToString([]byte("import time; time.sleep(0.1); print('done')")),
		TimeLimit:    2.0,
		MemoryLimit:  256,
		TestCases: []types.TestCaseMessage{
			{
				TestCaseID:     "tc1",
				Input:          base64.StdEncoding.EncodeToString([]byte("")),
				ExpectedOutput: base64.StdEncoding.EncodeToString([]byte("done")),
			},
		},
	}

	delivery := testutil.CreateTestDelivery(submission)
	jobQueue <- delivery
	close(jobQueue)

	w.Start()

	time.Sleep(3 * time.Second)

	messages := mqClient.getMessages()
	if len(messages) != 2 {
		t.Fatalf("Expected 2 messages, got %d", len(messages))
	}

	result := messages[1].body.(types.ResultNotificationMessage)
	if result.TimeTaken <= 0 {
		t.Error("Expected positive time taken")
	}
	if result.MemoryUsed <= 0 {
		t.Error("Expected positive memory usage")
	}
	if result.Results[0].TimeTaken <= 0 {
		t.Error("Expected positive test case time")
	}
	if result.Results[0].MemoryUsed <= 0 {
		t.Error("Expected positive test case memory usage")
	}
}

func TestWorkerIntegration_StatusUpdateTiming(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	mqClient := &integrationMQClient{}
	jobQueue := make(chan amqp091.Delivery, 1)
	w := worker.NewWorker(1, jobQueue, mqClient)

	submission := testutil.CreatePythonHelloWorldSubmission()
	delivery := testutil.CreateTestDelivery(submission)

	start := time.Now()
	jobQueue <- delivery
	close(jobQueue)

	w.Start()

	time.Sleep(3 * time.Second)

	messages := mqClient.getMessages()
	if len(messages) != 2 {
		t.Fatalf("Expected 2 messages, got %d", len(messages))
	}

	statusTime := messages[0].timestamp
	resultTime := messages[1].timestamp

	if statusTime.Before(start) {
		t.Error("Status message timestamp should be after start time")
	}
	if resultTime.Before(statusTime) {
		t.Error("Result message should come after status message")
	}

	timeDiff := resultTime.Sub(statusTime)
	if timeDiff < 0 || timeDiff > 10*time.Second {
		t.Errorf("Time between status and result should be reasonable, got %v", timeDiff)
	}
}