package master

import (
	"encoding/json"
	"log"
	"online-judge/executor/rabbitmq"
	"online-judge/executor/types"
	"online-judge/executor/worker"

	"github.com/rabbitmq/amqp091-go"
)

type Master struct {
	mqClient    rabbitmq.ClientInterface
	jobQueue    chan amqp091.Delivery
	workerCount int
	queueName   string
}

func NewMaster(mqClient rabbitmq.ClientInterface, workerCount int, queueName string) (*Master, error) {
	return &Master{
		mqClient:    mqClient,
		jobQueue:    make(chan amqp091.Delivery, workerCount),
		workerCount: workerCount,
		queueName:   queueName,
	}, nil
}

func (m *Master) Start() {
	for workerID := 1; workerID <= m.workerCount; workerID++ {
		worker := worker.NewWorker(workerID, m.jobQueue, m.mqClient)
		go worker.Start()
	}

	go m.consumeAndDispatch()
}

func (m *Master) consumeAndDispatch() {
	msgs, err := m.mqClient.ConsumeSubmissions(m.queueName)
	if err != nil {
		log.Fatalf("Failed to start consuming submissions: %v", err)
	}

	log.Printf("Master is waiting for submissions on queue '%s'. To exit press CTRL+C", m.queueName)

	for d := range msgs {
		var submission types.SubmissionMessage
		if err := json.Unmarshal(d.Body, &submission); err != nil {
			log.Printf("Error deserializing submission: %v. Rejecting message.", err)
			d.Ack(false) // Ack the malformed request
			continue
		}
		log.Printf("[Submission %d] Received submission. Dispatching to a worker.", submission.SubmissionID)
		m.jobQueue <- d
	}
}
