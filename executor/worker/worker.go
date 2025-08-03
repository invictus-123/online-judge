package worker

import (
	"encoding/base64"
	"encoding/json"
	"log"
	"online-judge/executor/docker"
	"online-judge/executor/rabbitmq"
	"online-judge/executor/types"

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

		execResult, err := docker.RunInContainer(submission.Language, string(decodedCode), string(decodedInput))
		if err != nil {
			log.Printf("[Worker %d] Execution failed for test case %s: %v", w.id, testCase.TestCaseID, err)
			results = append(results, types.TestCaseResultMessage{
				TestCaseID: testCase.TestCaseID,
				Status:     "COMPILATION_ERROR",
				Output:     base64.StdEncoding.EncodeToString([]byte(err.Error())),
			})
			continue
		}

		results = append(results, types.TestCaseResultMessage{
			TestCaseID: testCase.TestCaseID,
			Output:     base64.StdEncoding.EncodeToString([]byte(execResult.Output)),
			// Status:     execResult.Status,
			Status:     "PASSED",
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
	resultNotification := types.ResultNotificationMessage{
		SubmissionID: submissionID,
		Status:       "PASSED",
		TimeTaken:    1.5,
		MemoryUsed:   102,
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
