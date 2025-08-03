package main

import (
	"log"
	"online-judge/executor/master"
	"online-judge/executor/rabbitmq"
	"os"
	"os/signal"
	"syscall"
)

const (
	rabbitMqQUrl    = "amqp://guest:guest@localhost:5672/"
	workerCount     = 20
	submissionQueue = "oj.q.submissions"
)

func main() {
	log.Println("Starting Go Code Executor...")

	mqClient, err := rabbitmq.NewClient(rabbitMqQUrl)
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

func waitForShutdown() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
}
