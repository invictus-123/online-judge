package rabbitmq

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/rabbitmq/amqp091-go"
)

// Exchange and Routing Key constants, must match the Java backend configuration.
const (
	ResultExchange   = "oj.ex.results"
	ResultRoutingKey = "submission.result"
	StatusExchange   = "oj.ex.status"
	StatusRoutingKey = "submission.status"
)

type Client struct {
	conn *amqp091.Connection
	ch   *amqp091.Channel
}

func NewClient(url string) (*Client, error) {
	conn, err := amqp091.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	// Declare exchanges to ensure they exist.
	err = ch.ExchangeDeclare(
		ResultExchange,
		"direct", // kind
		true,     // durable
		false,    // autoDelete
		false,    // internal
		false,    // noWait
		nil,      // args
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare result exchange: %w", err)
	}

	err = ch.ExchangeDeclare(
		StatusExchange,
		"direct", // kind
		true,     // durable
		false,    // autoDelete
		false,    // internal
		false,    // noWait
		nil,      // args
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare status exchange: %w", err)
	}

	return &Client{conn: conn, ch: ch}, nil
}

func (c *Client) ConsumeSubmissions(queueName string) (<-chan amqp091.Delivery, error) {
	err := c.ch.Qos(
		1,     // prefetchCount: Each worker gets one message at a time.
		0,     // prefetchSize
		false, // global
	)
	if err != nil {
		return nil, fmt.Errorf("failed to set QoS: %w", err)
	}

	msgs, err := c.ch.Consume(
		queueName,
		"",    // consumer
		false, // auto-ack: messages will be manually acked in the worker
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return nil, fmt.Errorf("failed to register a consumer: %w", err)
	}
	return msgs, nil
}

func (c *Client) Publish(exchange, routingKey string, body interface{}) error {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal body to JSON: %w", err)
	}

	err = c.ch.Publish(
		exchange,
		routingKey,
		false, // mandatory
		false, // immediate
		amqp091.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp091.Persistent,
			Body:         jsonBody,
		})
	if err != nil {
		return fmt.Errorf("failed to publish a message: %w", err)
	}
	log.Printf("Published message to exchange '%s' with key '%s'", exchange, routingKey)
	return nil
}

func (c *Client) Close() {
	if c.ch != nil {
		c.ch.Close()
	}
	if c.conn != nil {
		c.conn.Close()
	}
}
