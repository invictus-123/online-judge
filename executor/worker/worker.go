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
	mqClient rabbitmq.ClientInterface
}

func NewWorker(id int, jobQueue <-chan amqp091.Delivery, mqClient rabbitmq.ClientInterface) *Worker {
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
	log.Printf("[Submission %d] [Worker %d] Processing submission.", submission.SubmissionID, w.id)

	if err := updateStatus(submission.SubmissionID, "RUNNING", w); err != nil {
		log.Printf("[Submission %d] [Worker %d] Failed to publish running status: %v", submission.SubmissionID, w.id, err)
		// We will continue processing but NACK at the end if results also fail to publish.
	}

	// 2. Execute all test cases
	decodedCode, err := base64.StdEncoding.DecodeString(submission.Code)
	if err != nil {
		log.Printf("[Submission %d] [Worker %d] Failed to decode submission code: %v. Rejecting message.", submission.SubmissionID, w.id, err)
		job.Ack(false) // Ack the message as there is no point executing further with a malformed code
		return
	}
	var results []types.TestCaseResultMessage
	for _, testCase := range submission.TestCases {
		log.Printf("[Submission %d] [Worker %d] Running test case %s.", submission.SubmissionID, w.id, testCase.TestCaseID)

		decodedInput, err := base64.StdEncoding.DecodeString(testCase.Input)
		if err != nil {
			log.Printf("[Submission %d] [Worker %d] Failed to decode test case input %s: %v. Failing this test case.", submission.SubmissionID, w.id, testCase.TestCaseID, err)
			results = append(results, types.TestCaseResultMessage{
				TestCaseID: testCase.TestCaseID,
				Status:     "COMPILATION_ERROR",
				Output:     base64.StdEncoding.EncodeToString([]byte("Invalid Base64 for test case input.")),
			})
			continue
		}

		memoryLimitBytes := submission.MemoryLimit * 1024 * 1024 // Convert MB to bytes
		execResult, err := docker.RunInContainerWithLimits(submission.SubmissionID, submission.Language, string(decodedCode), string(decodedInput), submission.TimeLimit, memoryLimitBytes)
		if err != nil {
			log.Printf("[Submission %d] [Worker %d] Execution failed for test case %s: %v", submission.SubmissionID, w.id, testCase.TestCaseID, err)
			results = append(results, types.TestCaseResultMessage{
				TestCaseID: testCase.TestCaseID,
				Status:     "COMPILATION_ERROR",
				Output:     base64.StdEncoding.EncodeToString([]byte(err.Error())),
			})
			continue
		}

		decodedExpectedOutput, err := base64.StdEncoding.DecodeString(testCase.ExpectedOutput)
		if err != nil {
			log.Printf("[Submission %d] [Worker %d] Failed to decode expected output for test case %s: %v", submission.SubmissionID, w.id, testCase.TestCaseID, err)
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
			MemoryUsed: execResult.MemoryKB,
		})
	}

	if err := sendResults(submission.SubmissionID, results, w); err != nil {
		log.Printf("[Submission %d] [Worker %d] Failed to publish results: %v. NACKing message.", submission.SubmissionID, w.id, err)
		job.Nack(false, true) // Nack and requeue, as results failed to send
		return
	}

	// 4. Acknowledge the message from the submission queue as processing is complete.
	job.Ack(false)
	log.Printf("[Submission %d] [Worker %d] Finished processing submission.", submission.SubmissionID, w.id)
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
	overallStatus := "PASSED"
	
	for _, result := range results {
		if result.TimeTaken > maxTime {
			maxTime = result.TimeTaken
		}
		if result.MemoryUsed > maxMemory {
			maxMemory = result.MemoryUsed
		}
		
		if result.Status == "COMPILATION_ERROR" {
			overallStatus = "COMPILATION_ERROR"
		} else if result.Status == "RUNTIME_ERROR" && overallStatus == "PASSED" {
			overallStatus = "RUNTIME_ERROR"
		} else if result.Status == "TIME_LIMIT_EXCEEDED" && (overallStatus == "PASSED" || overallStatus == "WRONG_ANSWER") {
			overallStatus = "TIME_LIMIT_EXCEEDED"
		} else if result.Status == "MEMORY_LIMIT_EXCEEDED" && (overallStatus == "PASSED" || overallStatus == "WRONG_ANSWER") {
			overallStatus = "MEMORY_LIMIT_EXCEEDED"
		} else if result.Status == "WRONG_ANSWER" && overallStatus == "PASSED" {
			overallStatus = "WRONG_ANSWER"
		}
	}
	
	return overallStatus, maxTime, maxMemory
}
