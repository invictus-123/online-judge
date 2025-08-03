package main

import (
	"fmt"
	"log"
	"online-judge/executor/master"
	"online-judge/executor/rabbitmq"
	"os"
	"os/signal"
	"syscall"
)

const (
	workerCount     = 20
	submissionQueue = "oj.q.submissions"
)

func main() {
	log.Println("Starting Go Code Executor...")

	rabbitMQHost := getEnv("RABBITMQ_HOST", "localhost")
	rabbitMQPort := getEnv("RABBITMQ_PORT", "5672")
	rabbitMQUser := getEnv("RABBITMQ_USER", "guest")
	rabbitMQPass := getEnv("RABBITMQ_PASS", "guest")

	rabbitMqURL := fmt.Sprintf("amqp://%s:%s@%s:%s/", rabbitMQUser, rabbitMQPass, rabbitMQHost, rabbitMQPort)

	log.Printf("Connecting to RabbitMQ at %s:%s", rabbitMQHost, rabbitMQPort)

	mqClient, err := rabbitmq.NewClient(rabbitMqURL)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer mqClient.Close()

	log.Println("RabbitMQ client initialized.")

	master, err := master.NewMaster(mqClient, workerCount, submissionQueue)
	if err != nil {
		log.Fatalf("Failed to create master node: %v", err)
	}

	master.Start()
	log.Printf("Master started with %d workers.", workerCount)

	waitForShutdown()
	log.Println("Shutting down executor...")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func waitForShutdown() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
}
