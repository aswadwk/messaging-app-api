package services

import (
	"fmt"
	"log"
	"time"

	"github.com/streadway/amqp"
)

type RabbitMQ struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	url     string
}

// NewRabbitMQ creates and connects to RabbitMQ
func NewRabbitMQ(amqpURL string) *RabbitMQ {
	rmq := &RabbitMQ{
		url: amqpURL,
	}

	if err := rmq.connect(); err != nil {
		log.Fatalf("[RabbitMQ] Failed to connect: %v", err)
		return nil
	}

	return rmq
}

// connect handles the actual connection and channel creation
func (r *RabbitMQ) connect() error {
	var err error

	r.conn, err = amqp.Dial(r.url)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	r.channel, err = r.conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open a channel: %w", err)
	}

	log.Println("[RabbitMQ] Connected successfully.")
	return nil
}

// DeclareQueue declares a queue with given name
func (r *RabbitMQ) DeclareQueue(queueName string) (amqp.Queue, error) {
	q, err := r.channel.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // auto-delete
		false,     // exclusive
		false,     // no-wait
		nil,       // args
	)
	if err != nil {
		return q, fmt.Errorf("failed to declare queue %s: %w", queueName, err)
	}
	return q, nil
}

// PublishMessage publishes message to the given queue
func (r *RabbitMQ) PublishMessage(queueName string, body []byte) error {
	err := r.channel.Publish(
		"",        // exchange
		queueName, // routing key
		false,     // mandatory
		false,     // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}
	return nil
}

// ConsumeMessages consumes messages from the given queue with a specific consumer tag
func (r *RabbitMQ) ConsumeMessages(queueName, consumerTag string) (<-chan amqp.Delivery, error) {
	msgs, err := r.channel.Consume(
		queueName,
		consumerTag,
		true,  // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return nil, fmt.Errorf("failed to consume messages: %w", err)
	}
	return msgs, nil
}

// Close shuts down channel and connection gracefully
func (r *RabbitMQ) Close() error {
	if err := r.channel.Close(); err != nil {
		log.Printf("[RabbitMQ] Channel close error: %v", err)
	}
	if err := r.conn.Close(); err != nil {
		log.Printf("[RabbitMQ] Connection close error: %v", err)
	}
	log.Println("[RabbitMQ] Closed connection and channel.")
	return nil
}

// Reconnect tries to reconnect if connection lost
func (r *RabbitMQ) Reconnect() {
	for {
		time.Sleep(5 * time.Second)
		log.Println("[RabbitMQ] Attempting to reconnect...")
		if err := r.connect(); err != nil {
			log.Printf("[RabbitMQ] Reconnect failed: %v", err)
			continue
		}
		log.Println("[RabbitMQ] Reconnected successfully.")
		break
	}
}

func (r *RabbitMQ) Channel() *amqp.Channel {
	if r.channel == nil {
		log.Println("[RabbitMQ] Channel is nil, reconnecting...")
		r.Reconnect()
	}
	return r.channel
}

// DeleteQueue deletes a queue with the given name
func (r *RabbitMQ) DeleteQueue(queueName string) error {
	_, err := r.channel.QueueDelete(queueName, false, false, false)
	if err != nil {
		return fmt.Errorf("failed to delete queue %s: %w", queueName, err)
	}
	log.Printf("[RabbitMQ] Queue %s deleted successfully.", queueName)
	return nil
}
