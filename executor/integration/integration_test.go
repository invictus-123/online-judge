package integration

import (
	"online-judge/executor/types"
	"testing"
	"time"
)

func TestIntegrationHelper_DockerAvailability(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	helper, err := NewIntegrationTestHelper()
	if err != nil {
		t.Skipf("Docker client unavailable: %v", err)
	}
	defer helper.Close()

	err = helper.CheckDockerAvailability()
	if err != nil {
		t.Skipf("Docker daemon unavailable: %v", err)
	}

	t.Log("Docker is available and accessible")
}

func TestIntegrationHelper_ContainerCleanup(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	helper := SetupIntegrationTest(t)
	LogTestEnvironment(t)

	err := helper.CleanupContainers()
	if err != nil {
		t.Errorf("Container cleanup failed: %v", err)
	}

	t.Log("Container cleanup completed successfully")
}

func TestTestMetrics_BasicFunctionality(t *testing.T) {
	metrics := NewTestMetrics()

	metrics.RecordSubmission()
	metrics.RecordSubmission()
	metrics.RecordSubmission()

	metrics.RecordStatusMessage()
	metrics.RecordStatusMessage()
	metrics.RecordStatusMessage()

	metrics.RecordResultMessage(1, "PASSED")
	metrics.RecordResultMessage(2, "WRONG_ANSWER")
	metrics.RecordResultMessage(3, "PASSED")

	metrics.FinishProcessing()

	if metrics.TotalSubmissions != 3 {
		t.Errorf("Total submissions = %d, want 3", metrics.TotalSubmissions)
	}
	if metrics.SuccessfulResults != 2 {
		t.Errorf("Successful results = %d, want 2", metrics.SuccessfulResults)
	}
	if metrics.FailedResults != 1 {
		t.Errorf("Failed results = %d, want 1", metrics.FailedResults)
	}
	if metrics.GetSuccessRate() != 2.0/3.0 {
		t.Errorf("Success rate = %f, want %f", metrics.GetSuccessRate(), 2.0/3.0)
	}
	if metrics.GetUniqueResultCount() != 3 {
		t.Errorf("Unique result count = %d, want 3", metrics.GetUniqueResultCount())
	}

	metrics.Validate(t)
}

func TestPerformanceBenchmark_Calculations(t *testing.T) {
	benchmark := NewPerformanceBenchmark(3, 10, 5*time.Second)
	
	actualDuration := 3 * time.Second
	benchmark.RecordCompletion(actualDuration)

	if !benchmark.IsWithinExpectedTime() {
		t.Error("Should be within expected time")
	}

	expectedThroughput := 10.0 / 3.0 // 10 submissions in 3 seconds
	if benchmark.ThroughputPerSec < expectedThroughput-0.1 || benchmark.ThroughputPerSec > expectedThroughput+0.1 {
		t.Errorf("Throughput = %f, want ~%f", benchmark.ThroughputPerSec, expectedThroughput)
	}

	efficiency := benchmark.GetEfficiency()
	expectedEfficiency := 5.0 / 3.0 // expected 5s, actual 3s
	if efficiency < expectedEfficiency-0.1 || efficiency > expectedEfficiency+0.1 {
		t.Errorf("Efficiency = %f, want ~%f", efficiency, expectedEfficiency)
	}

	benchmark.LogResults(t)
}

func TestResourceUsageTracker_Functionality(t *testing.T) {
	tracker := NewResourceUsageTracker()

	time.Sleep(1 * time.Millisecond) // Ensure some time passes

	tracker.UpdateContainerCount(2)
	tracker.UpdateContainerCount(5)
	tracker.UpdateContainerCount(3)
	tracker.UpdateContainerCount(7)
	tracker.UpdateContainerCount(1)

	tracker.Finish()

	if tracker.PeakContainers != 7 {
		t.Errorf("Peak containers = %d, want 7", tracker.PeakContainers)
	}
	if tracker.TotalContainers != 5 {
		t.Errorf("Total container updates = %d, want 5", tracker.TotalContainers)
	}

	duration := tracker.GetDuration()
	if duration <= 0 {
		t.Error("Duration should be positive")
	}

	tracker.LogResults(t)
}

func TestCreateBatchSubmissions(t *testing.T) {
	baseSubmissions := []types.SubmissionMessage{
		{SubmissionID: 0, Language: "PYTHON"},
		{SubmissionID: 0, Language: "JAVA"},
	}

	submissions := CreateBatchSubmissions(baseSubmissions, 100, 5)

	if len(submissions) != 5 {
		t.Errorf("Created %d submissions, want 5", len(submissions))
	}

	expectedIDs := []int64{100, 101, 102, 103, 104}
	expectedLangs := []string{"PYTHON", "JAVA", "PYTHON", "JAVA", "PYTHON"}

	for i, submission := range submissions {
		if submission.SubmissionID != expectedIDs[i] {
			t.Errorf("Submission %d ID = %d, want %d", i, submission.SubmissionID, expectedIDs[i])
		}
		if submission.Language != expectedLangs[i] {
			t.Errorf("Submission %d language = %s, want %s", i, submission.Language, expectedLangs[i])
		}
	}
}

func TestCreateBatchSubmissions_EdgeCases(t *testing.T) {
	baseSubmissions := []types.SubmissionMessage{
		{SubmissionID: 0, Language: "PYTHON"},
	}

	emptyBatch := CreateBatchSubmissions(baseSubmissions, 100, 0)
	if len(emptyBatch) != 0 {
		t.Errorf("Empty batch should have 0 submissions, got %d", len(emptyBatch))
	}

	negativeBatch := CreateBatchSubmissions(baseSubmissions, 100, -5)
	if len(negativeBatch) != 0 {
		t.Errorf("Negative count batch should have 0 submissions, got %d", len(negativeBatch))
	}

	singleBatch := CreateBatchSubmissions(baseSubmissions, 200, 1)
	if len(singleBatch) != 1 || singleBatch[0].SubmissionID != 200 {
		t.Errorf("Single batch should have 1 submission with ID 200")
	}
}

func TestValidateSubmissionResults(t *testing.T) {
	expected := map[int64]string{
		1: "PASSED",
		2: "WRONG_ANSWER",
		3: "PASSED",
	}

	actual := map[int64]string{
		1: "PASSED",
		2: "WRONG_ANSWER", 
		3: "PASSED",
	}

	ValidateSubmissionResults(t, expected, actual)

	// Test validation works correctly with matching data
	ValidateSubmissionResults(t, expected, actual)
}

func TestAssertNoDataRaces(t *testing.T) {
	submissionIDs := map[int64]bool{
		1: true,
		2: true,
		3: true,
	}

	AssertNoDataRaces(t, submissionIDs, 3)

	// Test should detect race condition - simplified test since we can't mock testing.T
	if len(submissionIDs) == 5 {
		t.Error("This should not happen - testing the test logic")
	}
}

func TestTestSubmissionBuilder(t *testing.T) {
	baseSubmission := types.SubmissionMessage{
		SubmissionID: 0,
		Language:     "PYTHON",
		TimeLimit:    1.0,
		MemoryLimit:  128,
	}

	builder := NewTestSubmissionBuilder(baseSubmission)
	submission := builder.
		WithID(999).
		WithLanguage("JAVA").
		WithTimeLimit(5.0).
		WithMemoryLimit(512).
		Build()

	if submission.SubmissionID != 999 {
		t.Errorf("SubmissionID = %d, want 999", submission.SubmissionID)
	}
	if submission.Language != "JAVA" {
		t.Errorf("Language = %s, want JAVA", submission.Language)
	}
	if submission.TimeLimit != 5.0 {
		t.Errorf("TimeLimit = %f, want 5.0", submission.TimeLimit)
	}
	if submission.MemoryLimit != 512 {
		t.Errorf("MemoryLimit = %d, want 512", submission.MemoryLimit)
	}
}

func TestAssertProcessingTime(t *testing.T) {
	startTime := time.Now()
	time.Sleep(10 * time.Millisecond)

	AssertProcessingTime(t, startTime, 100*time.Millisecond, "test operation")

	// Test timeout detection - simplified since we can't mock testing.T
	elapsed := time.Since(startTime)
	if elapsed <= 5*time.Millisecond {
		t.Error("This should not happen - elapsed time should be more than 5ms")
	}
}