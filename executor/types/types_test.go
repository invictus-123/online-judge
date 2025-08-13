package types

import (
	"encoding/json"
	"testing"
)

func TestSubmissionMessage_JSON(t *testing.T) {
	tests := []struct {
		name string
		msg  SubmissionMessage
	}{
		{
			name: "complete submission message",
			msg: SubmissionMessage{
				SubmissionID: 123,
				Language:     "JAVA",
				Code:         "cHVibGljIGNsYXNzIE1haW4ge30=",
				TimeLimit:    2.5,
				MemoryLimit:  512,
				TestCases: []TestCaseMessage{
					{
						TestCaseID:     "test1",
						Input:          "aW5wdXQ=",
						ExpectedOutput: "b3V0cHV0",
					},
				},
			},
		},
		{
			name: "minimal submission message",
			msg: SubmissionMessage{
				SubmissionID: 1,
				Language:     "PYTHON",
				Code:         "cHJpbnQoImhlbGxvIik=",
				TimeLimit:    1.0,
				MemoryLimit:  128,
				TestCases:    []TestCaseMessage{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.msg)
			if err != nil {
				t.Fatalf("failed to marshal: %v", err)
			}

			var unmarshaled SubmissionMessage
			if err := json.Unmarshal(data, &unmarshaled); err != nil {
				t.Fatalf("failed to unmarshal: %v", err)
			}

			if unmarshaled.SubmissionID != tt.msg.SubmissionID {
				t.Errorf("SubmissionID = %d, want %d", unmarshaled.SubmissionID, tt.msg.SubmissionID)
			}
			if unmarshaled.Language != tt.msg.Language {
				t.Errorf("Language = %s, want %s", unmarshaled.Language, tt.msg.Language)
			}
			if unmarshaled.Code != tt.msg.Code {
				t.Errorf("Code = %s, want %s", unmarshaled.Code, tt.msg.Code)
			}
			if unmarshaled.TimeLimit != tt.msg.TimeLimit {
				t.Errorf("TimeLimit = %f, want %f", unmarshaled.TimeLimit, tt.msg.TimeLimit)
			}
			if unmarshaled.MemoryLimit != tt.msg.MemoryLimit {
				t.Errorf("MemoryLimit = %d, want %d", unmarshaled.MemoryLimit, tt.msg.MemoryLimit)
			}
			if len(unmarshaled.TestCases) != len(tt.msg.TestCases) {
				t.Errorf("TestCases length = %d, want %d", len(unmarshaled.TestCases), len(tt.msg.TestCases))
			}
		})
	}
}

func TestTestCaseMessage_JSON(t *testing.T) {
	testCase := TestCaseMessage{
		TestCaseID:     "tc123",
		Input:          "aW5wdXQgZGF0YQ==",
		ExpectedOutput: "ZXhwZWN0ZWQgb3V0cHV0",
	}

	data, err := json.Marshal(testCase)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var unmarshaled TestCaseMessage
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if unmarshaled.TestCaseID != testCase.TestCaseID {
		t.Errorf("TestCaseID = %s, want %s", unmarshaled.TestCaseID, testCase.TestCaseID)
	}
	if unmarshaled.Input != testCase.Input {
		t.Errorf("Input = %s, want %s", unmarshaled.Input, testCase.Input)
	}
	if unmarshaled.ExpectedOutput != testCase.ExpectedOutput {
		t.Errorf("ExpectedOutput = %s, want %s", unmarshaled.ExpectedOutput, testCase.ExpectedOutput)
	}
}

func TestStatusUpdateMessage_JSON(t *testing.T) {
	msg := StatusUpdateMessage{
		SubmissionID: 456,
		Status:       "RUNNING",
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var unmarshaled StatusUpdateMessage
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if unmarshaled.SubmissionID != msg.SubmissionID {
		t.Errorf("SubmissionID = %d, want %d", unmarshaled.SubmissionID, msg.SubmissionID)
	}
	if unmarshaled.Status != msg.Status {
		t.Errorf("Status = %s, want %s", unmarshaled.Status, msg.Status)
	}
}

func TestResultNotificationMessage_JSON(t *testing.T) {
	msg := ResultNotificationMessage{
		SubmissionID: 789,
		Status:       "PASSED",
		TimeTaken:    1.25,
		MemoryUsed:   128,
		Results: []TestCaseResultMessage{
			{
				TestCaseID: "tc1",
				Output:     "b3V0cHV0",
				Status:     "PASSED",
				TimeTaken:  0.5,
				MemoryUsed: 64,
			},
		},
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var unmarshaled ResultNotificationMessage
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if unmarshaled.SubmissionID != msg.SubmissionID {
		t.Errorf("SubmissionID = %d, want %d", unmarshaled.SubmissionID, msg.SubmissionID)
	}
	if unmarshaled.Status != msg.Status {
		t.Errorf("Status = %s, want %s", unmarshaled.Status, msg.Status)
	}
	if unmarshaled.TimeTaken != msg.TimeTaken {
		t.Errorf("TimeTaken = %f, want %f", unmarshaled.TimeTaken, msg.TimeTaken)
	}
	if unmarshaled.MemoryUsed != msg.MemoryUsed {
		t.Errorf("MemoryUsed = %d, want %d", unmarshaled.MemoryUsed, msg.MemoryUsed)
	}
	if len(unmarshaled.Results) != len(msg.Results) {
		t.Errorf("Results length = %d, want %d", len(unmarshaled.Results), len(msg.Results))
	}
}

func TestTestCaseResultMessage_JSON(t *testing.T) {
	result := TestCaseResultMessage{
		TestCaseID: "tc456",
		Output:     "cmVzdWx0",
		Status:     "WRONG_ANSWER",
		TimeTaken:  2.0,
		MemoryUsed: 256,
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var unmarshaled TestCaseResultMessage
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if unmarshaled.TestCaseID != result.TestCaseID {
		t.Errorf("TestCaseID = %s, want %s", unmarshaled.TestCaseID, result.TestCaseID)
	}
	if unmarshaled.Output != result.Output {
		t.Errorf("Output = %s, want %s", unmarshaled.Output, result.Output)
	}
	if unmarshaled.Status != result.Status {
		t.Errorf("Status = %s, want %s", unmarshaled.Status, result.Status)
	}
	if unmarshaled.TimeTaken != result.TimeTaken {
		t.Errorf("TimeTaken = %f, want %f", unmarshaled.TimeTaken, result.TimeTaken)
	}
	if unmarshaled.MemoryUsed != result.MemoryUsed {
		t.Errorf("MemoryUsed = %d, want %d", unmarshaled.MemoryUsed, result.MemoryUsed)
	}
}

func TestEmptyTestCases(t *testing.T) {
	msg := SubmissionMessage{
		SubmissionID: 1,
		Language:     "JAVA",
		Code:         "Y29kZQ==",
		TimeLimit:    1.0,
		MemoryLimit:  128,
		TestCases:    []TestCaseMessage{},
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var unmarshaled SubmissionMessage
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if len(unmarshaled.TestCases) != 0 {
		t.Errorf("TestCases length = %d, want 0", len(unmarshaled.TestCases))
	}
}
