package master

import (
	"testing"

	"github.com/rabbitmq/amqp091-go"
)

type mockClient struct{}

func (m *mockClient) ConsumeSubmissions(queueName string) (<-chan amqp091.Delivery, error) {
	ch := make(chan amqp091.Delivery)
	close(ch)
	return ch, nil
}

func (m *mockClient) Publish(exchange, routingKey string, body interface{}) error {
	return nil
}

func TestNewMaster(t *testing.T) {
	mqClient := &mockClient{}
	
	master, err := NewMaster(mqClient, 5, "test.queue")
	if err != nil {
		t.Fatalf("NewMaster failed: %v", err)
	}
	
	if master == nil {
		t.Fatal("Master should not be nil")
	}
	if master.workerCount != 5 {
		t.Errorf("Worker count = %d, want 5", master.workerCount)
	}
	if master.queueName != "test.queue" {
		t.Errorf("Queue name = %s, want test.queue", master.queueName)
	}
	if master.jobQueue == nil {
		t.Error("Job queue should not be nil")
	}
}

func TestNewMasterWithZeroWorkers(t *testing.T) {
	mqClient := &mockClient{}
	
	master, err := NewMaster(mqClient, 0, "test.queue")
	if err != nil {
		t.Fatalf("NewMaster with zero workers failed: %v", err)
	}
	
	if master == nil {
		t.Fatal("Master should not be nil")
	}
	if master.workerCount != 0 {
		t.Errorf("Worker count = %d, want 0", master.workerCount)
	}
}

func TestMasterJobQueueCapacity(t *testing.T) {
	workerCount := 10
	mqClient := &mockClient{}
	
	master, err := NewMaster(mqClient, workerCount, "test.queue")
	if err != nil {
		t.Fatalf("NewMaster failed: %v", err)
	}
	
	if cap(master.jobQueue) != workerCount {
		t.Errorf("Job queue capacity = %d, want %d", cap(master.jobQueue), workerCount)
	}
}