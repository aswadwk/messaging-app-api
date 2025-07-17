package services

import (
	"encoding/json"

	"github.com/streadway/amqp"
)

type Message struct {
	TenantID string `json:"tenant_id"`
	Payload  any    `json:"payload"`
}

type PublisherService struct {
	rabbit *RabbitMQ
}

func NewPublisherService(rabbit *RabbitMQ) *PublisherService {
	return &PublisherService{
		rabbit: rabbit,
	}
}

func (s *PublisherService) Publish(queueName string, msg Message) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return s.rabbit.Channel().Publish(
		"",        // exchange
		queueName, // routing key
		false,     // mandatory
		false,     // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
}
