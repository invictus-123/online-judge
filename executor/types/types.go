package types

// SubmissionMessage corresponds to the message received from the submission queue.
type SubmissionMessage struct {
	SubmissionID int64             `json:"submissionId"`
	Language     string            `json:"language"`
	Code         string            `json:"code"`
	TimeLimit    float64           `json:"timeLimit"`
	MemoryLimit  int64             `json:"memoryLimit"`
	TestCases    []TestCaseMessage `json:"testCases"`
}

// TestCaseMessage represents a single test case for a problem.
type TestCaseMessage struct {
	TestCaseID     string `json:"testCaseId"`
	Input          string `json:"input"`
	ExpectedOutput string `json:"output"`
}

// StatusUpdateMessage is sent to the status queue.
type StatusUpdateMessage struct {
	SubmissionID int64  `json:"submissionId"`
	Status       string `json:"status"`
}

// ResultNotificationMessage is sent to the result queue.
type ResultNotificationMessage struct {
	SubmissionID int64                   `json:"submissionId"`
	Status       string                  `json:"status"`
	TimeTaken    float64                 `json:"timeTaken"`
	MemoryUsed   int64                   `json:"memoryUsed"`
	Results      []TestCaseResultMessage `json:"testCaseResults"`
}

// TestCaseResultMessage contains the outcome of a single test case execution.
type TestCaseResultMessage struct {
	TestCaseID string  `json:"testCaseId"`
	Output     string  `json:"output"`
	Status     string  `json:"status"`
	TimeTaken  float64 `json:"timeTaken"`
	MemoryUsed int64   `json:"memoryUsed"`
}
