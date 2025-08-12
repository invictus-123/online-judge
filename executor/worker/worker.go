package worker

import (
	"encoding/base64"
	"encoding/json"
	"log"
	"online-judge/executor/docker"
	"online-judge/executor/rabbitmq"
	"online-judge/executor/types"
	"strings"

	"github.com/rabbitmq/amqp091-go"
)

type Worker struct {
	id       int
	jobQueue <-chan amqp091.Delivery
	mqClient *rabbitmq.Client
}

func NewWorker(id int, jobQueue <-chan amqp091.Delivery, mqClient *rabbitmq.Client) *Worker {
	return &Worker{
		id:       id,
		jobQueue: jobQueue,
		mqClient: mqClient,
	}
}

func (w *Worker) Start() {
	for job := range w.jobQueue {
		w.process(job)
	}
}

func (w *Worker) process(job amqp091.Delivery) {
	var submission types.SubmissionMessage
	if err := json.Unmarshal(job.Body, &submission); err != nil {
		log.Printf("[Worker %d] Error deserializing submission: %v. Rejecting message.", w.id, err)
		job.Nack(false, false) // Nack and send to DLQ
		return
	}
	log.Printf("[Worker %d] Processing submission %d.", w.id, submission.SubmissionID)

	if err := updateStatus(submission.SubmissionID, "RUNNING", w); err != nil {
		log.Printf("[Worker %d] Failed to publish running status for submission %d: %v", w.id, submission.SubmissionID, err)
		// We will continue processing but NACK at the end if results also fail to publish.
	}

	// 2. Execute all test cases
	decodedCode, err := base64.StdEncoding.DecodeString(submission.Code)
	if err != nil {
		log.Printf("[Worker %d] Failed to decode submission code for submission %d: %v. Rejecting message.", w.id, submission.SubmissionID, err)
		job.Ack(false) // Ack the message as there is no point executing further with a malformed code
		return
	}
	var results []types.TestCaseResultMessage
	for _, testCase := range submission.TestCases {
		log.Printf("[Worker %d] Running test case %s for submission %d.", w.id, testCase.TestCaseID, submission.SubmissionID)

		decodedInput, err := base64.StdEncoding.DecodeString(testCase.Input)
		if err != nil {
			log.Printf("[Worker %d] Failed to decode test case input %s for submission %d: %v. Failing this test case.", w.id, testCase.TestCaseID, submission.SubmissionID, err)
			results = append(results, types.TestCaseResultMessage{
				TestCaseID: testCase.TestCaseID,
				Status:     "COMPILATION_ERROR",
				Output:     base64.StdEncoding.EncodeToString([]byte("Invalid Base64 for test case input.")),
			})
			continue
		}

		memoryLimitBytes := submission.MemoryLimit * 1024 * 1024 // Convert MB to bytes
		execResult, err := docker.RunInContainerWithLimits(submission.Language, string(decodedCode), string(decodedInput), submission.TimeLimit, memoryLimitBytes)
		if err != nil {
			log.Printf("[Worker %d] Execution failed for test case %s: %v", w.id, testCase.TestCaseID, err)
			results = append(results, types.TestCaseResultMessage{
				TestCaseID: testCase.TestCaseID,
				Status:     "COMPILATION_ERROR",
				Output:     base64.StdEncoding.EncodeToString([]byte(err.Error())),
			})
			continue
		}

		decodedExpectedOutput, err := base64.StdEncoding.DecodeString(testCase.ExpectedOutput)
		if err != nil {
			log.Printf("[Worker %d] Failed to decode expected output for test case %s: %v", w.id, testCase.TestCaseID, err)
			results = append(results, types.TestCaseResultMessage{
				TestCaseID: testCase.TestCaseID,
				Status:     "COMPILATION_ERROR",
				Output:     base64.StdEncoding.EncodeToString([]byte("Invalid Base64 for expected output.")),
			})
			continue
		}

		status := computeTestCaseStatus(execResult, string(decodedExpectedOutput))
		
		results = append(results, types.TestCaseResultMessage{
			TestCaseID: testCase.TestCaseID,
			Output:     base64.StdEncoding.EncodeToString([]byte(execResult.Output)),
			Status:     status,
			TimeTaken:  float64(execResult.TimeMillis) / 1000,
			MemoryUsed: execResult.MemoryKB / 1024,
		})
	}

	if err := sendResults(submission.SubmissionID, results, w); err != nil {
		log.Printf("[Worker %d] Failed to publish results for submission %d: %v. NACKing message.", w.id, submission.SubmissionID, err)
		job.Nack(false, true) // Nack and requeue, as results failed to send
		return
	}

	// 4. Acknowledge the message from the submission queue as processing is complete.
	job.Ack(false)
	log.Printf("[Worker %d] Finished processing submission %d.", w.id, submission.SubmissionID)
}

func sendResults(submissionID int64, results []types.TestCaseResultMessage, w *Worker) error {
	overallStatus, maxTime, maxMemory := computeOverallStatus(results)
	
	resultNotification := types.ResultNotificationMessage{
		SubmissionID: submissionID,
		Status:       overallStatus,
		TimeTaken:    maxTime,
		MemoryUsed:   maxMemory,
		Results:      results,
	}
	return w.mqClient.Publish(rabbitmq.ResultExchange, rabbitmq.ResultRoutingKey, resultNotification)
}

func updateStatus(submissionID int64, status string, w *Worker) error {
	statusUpdate := types.StatusUpdateMessage{
		SubmissionID: submissionID,
		Status:       status,
	}
	return w.mqClient.Publish(rabbitmq.StatusExchange, rabbitmq.StatusRoutingKey, statusUpdate)
}

func computeTestCaseStatus(execResult *docker.ExecutionResult, expectedOutput string) string {
	if execResult.Status == "TIME_LIMIT_EXCEEDED" {
		return "TIME_LIMIT_EXCEEDED"
	}
	if execResult.Status == "COMPILATION_ERROR" {
		return "COMPILATION_ERROR"
	}
	if execResult.Status == "RUNTIME_ERROR" {
		return "RUNTIME_ERROR"
	}
	
	actualOutput := strings.TrimSpace(execResult.Output)
	expectedOutput = strings.TrimSpace(expectedOutput)
	
	if actualOutput == expectedOutput {
		return "PASSED"
	}
	return "WRONG_ANSWER"
}

func computeOverallStatus(results []types.TestCaseResultMessage) (string, float64, int64) {
	if len(results) == 0 {
		return "COMPILATION_ERROR", 0.0, 0
	}
	
	var maxTime float64
	var maxMemory int64
	
	for _, result := range results {
		if result.TimeTaken > maxTime {
			maxTime = result.TimeTaken
		}
		if result.MemoryUsed > maxMemory {
			maxMemory = result.MemoryUsed
		}
		
		if result.Status == "COMPILATION_ERROR" {
			return "COMPILATION_ERROR", maxTime, maxMemory
		}
		if result.Status == "RUNTIME_ERROR" {
			return "RUNTIME_ERROR", maxTime, maxMemory
		}
		if result.Status == "TIME_LIMIT_EXCEEDED" {
			return "TIME_LIMIT_EXCEEDED", maxTime, maxMemory
		}
		if result.Status == "MEMORY_LIMIT_EXCEEDED" {
			return "MEMORY_LIMIT_EXCEEDED", maxTime, maxMemory
		}
		if result.Status == "WRONG_ANSWER" {
			return "WRONG_ANSWER", maxTime, maxMemory
		}
	}
	
	return "PASSED", maxTime, maxMemory
}
