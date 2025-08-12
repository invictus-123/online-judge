package integration

import (
	"context"
	"fmt"
	"log"
	"online-judge/executor/testutil"
	"online-judge/executor/types"
	"os"
	"testing"
	"time"

	dockerTypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

const (
	DockerTimeoutSeconds = 30
	TestTimeoutSeconds   = 60
	MessageWaitSeconds   = 5
)

type IntegrationTestHelper struct {
	dockerClient *client.Client
	testContext  context.Context
}

func NewIntegrationTestHelper() (*IntegrationTestHelper, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %w", err)
	}

	ctx := context.Background()
	
	return &IntegrationTestHelper{
		dockerClient: cli,
		testContext:  ctx,
	}, nil
}

func (h *IntegrationTestHelper) CheckDockerAvailability() error {
	ctx, cancel := context.WithTimeout(h.testContext, DockerTimeoutSeconds*time.Second)
	defer cancel()

	_, err := h.dockerClient.Ping(ctx)
	if err != nil {
		return fmt.Errorf("docker is not available: %w", err)
	}

	return nil
}

func (h *IntegrationTestHelper) PullRequiredImages() error {
	requiredImages := []string{
		"python:3.9-slim",
		"openjdk:11-jdk-slim", 
		"gcc:latest",
	}

	for _, image := range requiredImages {
		log.Printf("Checking image: %s", image)
		ctx, cancel := context.WithTimeout(h.testContext, 300*time.Second)
		
		reader, err := h.dockerClient.ImagePull(ctx, image, dockerTypes.ImagePullOptions{})
		if err != nil {
			cancel()
			return fmt.Errorf("failed to pull image %s: %w", image, err)
		}
		
		_, _ = reader.Read(make([]byte, 1024))
		reader.Close()
		cancel()
		
		log.Printf("Image %s is available", image)
	}

	return nil
}

func (h *IntegrationTestHelper) CleanupContainers() error {
	containers, err := h.dockerClient.ContainerList(h.testContext, dockerTypes.ContainerListOptions{All: true})
	if err != nil {
		return fmt.Errorf("failed to list containers: %w", err)
	}

	for _, container := range containers {
		for _, name := range container.Names {
			if len(name) > 0 && name[0] == '/' && len(name) > 4 && name[1:4] == "oj-" {
				log.Printf("Removing test container: %s", container.ID[:12])
				err := h.dockerClient.ContainerRemove(h.testContext, container.ID, dockerTypes.ContainerRemoveOptions{Force: true})
				if err != nil {
					log.Printf("Warning: failed to remove container %s: %v", container.ID[:12], err)
				}
			}
		}
	}

	return nil
}

func (h *IntegrationTestHelper) WaitForStabilization() {
	time.Sleep(MessageWaitSeconds * time.Second)
}

func (h *IntegrationTestHelper) Close() {
	if h.dockerClient != nil {
		h.dockerClient.Close()
	}
}

func SkipIfDockerUnavailable(t *testing.T) {
	helper, err := NewIntegrationTestHelper()
	if err != nil {
		t.Skipf("Docker client unavailable: %v", err)
	}
	defer helper.Close()

	if err := helper.CheckDockerAvailability(); err != nil {
		t.Skipf("Docker daemon unavailable: %v", err)
	}
}

func SetupIntegrationTest(t *testing.T) *IntegrationTestHelper {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	helper, err := NewIntegrationTestHelper()
	if err != nil {
		t.Fatalf("Failed to create integration test helper: %v", err)
	}

	if err := helper.CheckDockerAvailability(); err != nil {
		helper.Close()
		t.Skipf("Docker daemon unavailable: %v", err)
	}

	if os.Getenv("PULL_IMAGES") == "true" {
		if err := helper.PullRequiredImages(); err != nil {
			t.Logf("Warning: failed to pull some images: %v", err)
		}
	}

	t.Cleanup(func() {
		helper.CleanupContainers()
		helper.Close()
	})

	return helper
}

type TestSubmissionBuilder struct {
	submission types.SubmissionMessage
}

func NewTestSubmissionBuilder(baseSubmission types.SubmissionMessage) *TestSubmissionBuilder {
	return &TestSubmissionBuilder{submission: baseSubmission}
}

func (b *TestSubmissionBuilder) WithID(id int64) *TestSubmissionBuilder {
	b.submission.SubmissionID = id
	return b
}

func (b *TestSubmissionBuilder) WithLanguage(lang string) *TestSubmissionBuilder {
	b.submission.Language = lang
	return b
}

func (b *TestSubmissionBuilder) WithTimeLimit(limit float64) *TestSubmissionBuilder {
	b.submission.TimeLimit = limit
	return b
}

func (b *TestSubmissionBuilder) WithMemoryLimit(limit int64) *TestSubmissionBuilder {
	b.submission.MemoryLimit = limit
	return b
}

func (b *TestSubmissionBuilder) WithTestCases(testCases []testutil.TestCase) *TestSubmissionBuilder {
	tcMessages := make([]types.TestCaseMessage, len(testCases))
	for i, tc := range testCases {
		tcMessages[i] = types.TestCaseMessage{
			TestCaseID:     tc.ID,
			Input:          tc.Input,
			ExpectedOutput: tc.ExpectedOutput,
		}
	}
	b.submission.TestCases = tcMessages
	return b
}

func (b *TestSubmissionBuilder) Build() types.SubmissionMessage {
	return b.submission
}

type TestMetrics struct {
	TotalSubmissions    int
	SuccessfulResults   int
	FailedResults       int
	ProcessingStartTime time.Time
	ProcessingEndTime   time.Time
	StatusMessageCount  int
	ResultMessageCount  int
	UniqueSubmissionIDs map[int64]bool
}

func NewTestMetrics() *TestMetrics {
	return &TestMetrics{
		UniqueSubmissionIDs: make(map[int64]bool),
		ProcessingStartTime: time.Now(),
	}
}

func (m *TestMetrics) RecordSubmission() {
	m.TotalSubmissions++
}

func (m *TestMetrics) RecordStatusMessage() {
	m.StatusMessageCount++
}

func (m *TestMetrics) RecordResultMessage(submissionID int64, status string) {
	m.ResultMessageCount++
	m.UniqueSubmissionIDs[submissionID] = true
	
	if status == "PASSED" {
		m.SuccessfulResults++
	} else {
		m.FailedResults++
	}
}

func (m *TestMetrics) FinishProcessing() {
	m.ProcessingEndTime = time.Now()
}

func (m *TestMetrics) GetProcessingDuration() time.Duration {
	if m.ProcessingEndTime.IsZero() {
		return time.Since(m.ProcessingStartTime)
	}
	return m.ProcessingEndTime.Sub(m.ProcessingStartTime)
}

func (m *TestMetrics) GetSuccessRate() float64 {
	if m.TotalSubmissions == 0 {
		return 0.0
	}
	return float64(m.SuccessfulResults) / float64(m.TotalSubmissions)
}

func (m *TestMetrics) GetUniqueResultCount() int {
	return len(m.UniqueSubmissionIDs)
}

func (m *TestMetrics) Validate(t *testing.T) {
	if m.ResultMessageCount != m.TotalSubmissions {
		t.Errorf("Result message count (%d) != total submissions (%d)", m.ResultMessageCount, m.TotalSubmissions)
	}
	
	if m.StatusMessageCount != m.TotalSubmissions {
		t.Errorf("Status message count (%d) != total submissions (%d)", m.StatusMessageCount, m.TotalSubmissions)
	}
	
	if len(m.UniqueSubmissionIDs) != m.TotalSubmissions {
		t.Errorf("Unique submission IDs (%d) != total submissions (%d)", len(m.UniqueSubmissionIDs), m.TotalSubmissions)
	}
	
	if m.SuccessfulResults+m.FailedResults != m.TotalSubmissions {
		t.Errorf("Success + Failed (%d + %d) != total submissions (%d)", 
			m.SuccessfulResults, m.FailedResults, m.TotalSubmissions)
	}
}

type PerformanceBenchmark struct {
	WorkerCount       int
	SubmissionCount   int
	ExpectedDuration  time.Duration
	ActualDuration    time.Duration
	ThroughputPerSec  float64
	AverageLatency    time.Duration
}

func NewPerformanceBenchmark(workerCount, submissionCount int, expectedDuration time.Duration) *PerformanceBenchmark {
	return &PerformanceBenchmark{
		WorkerCount:      workerCount,
		SubmissionCount:  submissionCount,
		ExpectedDuration: expectedDuration,
	}
}

func (b *PerformanceBenchmark) RecordCompletion(actualDuration time.Duration) {
	b.ActualDuration = actualDuration
	b.ThroughputPerSec = float64(b.SubmissionCount) / actualDuration.Seconds()
	b.AverageLatency = actualDuration / time.Duration(b.SubmissionCount)
}

func (b *PerformanceBenchmark) IsWithinExpectedTime() bool {
	return b.ActualDuration <= b.ExpectedDuration
}

func (b *PerformanceBenchmark) GetEfficiency() float64 {
	if b.ExpectedDuration == 0 {
		return 1.0
	}
	return b.ExpectedDuration.Seconds() / b.ActualDuration.Seconds()
}

func (b *PerformanceBenchmark) LogResults(t *testing.T) {
	t.Logf("Performance Benchmark Results:")
	t.Logf("  Workers: %d", b.WorkerCount)
	t.Logf("  Submissions: %d", b.SubmissionCount)
	t.Logf("  Expected Duration: %v", b.ExpectedDuration)
	t.Logf("  Actual Duration: %v", b.ActualDuration)
	t.Logf("  Throughput: %.2f submissions/sec", b.ThroughputPerSec)
	t.Logf("  Average Latency: %v", b.AverageLatency)
	t.Logf("  Efficiency: %.2f%%", b.GetEfficiency()*100)
	
	if !b.IsWithinExpectedTime() {
		t.Logf("  WARNING: Execution exceeded expected time by %v", b.ActualDuration-b.ExpectedDuration)
	}
}

type ResourceUsageTracker struct {
	StartTime      time.Time
	EndTime        time.Time
	PeakContainers int
	TotalContainers int
}

func NewResourceUsageTracker() *ResourceUsageTracker {
	return &ResourceUsageTracker{
		StartTime: time.Now(),
	}
}

func (r *ResourceUsageTracker) UpdateContainerCount(current int) {
	r.TotalContainers++
	if current > r.PeakContainers {
		r.PeakContainers = current
	}
}

func (r *ResourceUsageTracker) Finish() {
	r.EndTime = time.Now()
}

func (r *ResourceUsageTracker) GetDuration() time.Duration {
	if r.EndTime.IsZero() {
		return time.Since(r.StartTime)
	}
	return r.EndTime.Sub(r.StartTime)
}

func (r *ResourceUsageTracker) LogResults(t *testing.T) {
	t.Logf("Resource Usage:")
	t.Logf("  Duration: %v", r.GetDuration())
	t.Logf("  Peak Containers: %d", r.PeakContainers)
	t.Logf("  Total Container Creates: %d", r.TotalContainers)
}

func CreateBatchSubmissions(baseSubmissions []types.SubmissionMessage, startID int64, count int) []types.SubmissionMessage {
	if count <= 0 {
		return []types.SubmissionMessage{}
	}
	
	submissions := make([]types.SubmissionMessage, count)
	baseCount := len(baseSubmissions)
	
	for i := 0; i < count; i++ {
		baseIndex := i % baseCount
		submissions[i] = baseSubmissions[baseIndex]
		submissions[i].SubmissionID = startID + int64(i)
	}
	
	return submissions
}

func ValidateSubmissionResults(t *testing.T, expected map[int64]string, actual map[int64]string) {
	for id, expectedStatus := range expected {
		if actualStatus, exists := actual[id]; !exists {
			t.Errorf("Missing result for submission %d", id)
		} else if actualStatus != expectedStatus {
			t.Errorf("Submission %d status = %s, want %s", id, actualStatus, expectedStatus)
		}
	}
	
	for id := range actual {
		if _, exists := expected[id]; !exists {
			t.Errorf("Unexpected result for submission %d", id)
		}
	}
}

func AssertProcessingTime(t *testing.T, startTime time.Time, maxDuration time.Duration, description string) {
	elapsed := time.Since(startTime)
	if elapsed > maxDuration {
		t.Errorf("%s took %v, expected less than %v", description, elapsed, maxDuration)
	}
}

func AssertNoDataRaces(t *testing.T, submissionIDs map[int64]bool, expectedCount int) {
	if len(submissionIDs) != expectedCount {
		t.Errorf("Expected %d unique submissions, got %d (possible race condition)", expectedCount, len(submissionIDs))
	}
}

func LogTestEnvironment(t *testing.T) {
	t.Logf("Test Environment:")
	t.Logf("  Docker Timeout: %d seconds", DockerTimeoutSeconds)
	t.Logf("  Test Timeout: %d seconds", TestTimeoutSeconds)
	t.Logf("  Message Wait: %d seconds", MessageWaitSeconds)
	t.Logf("  Pull Images: %s", os.Getenv("PULL_IMAGES"))
	t.Logf("  Testing Mode: %s", func() string {
		if testing.Short() {
			return "short"
		}
		return "full"
	}())
}