package kafka

import (
	"awesomeProject/models"
	"encoding/json"
	"log"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/google/uuid"
)

type Producer struct {
	P *kafka.Producer
}

func NewProducer(p *kafka.Producer) *Producer {
	return &Producer{P: p}
}

func (p *Producer) PublishOrderCreated(order *models.Orders) error {
	event := map[string]interface{}{
		"event_id":   uuid.New().String(),
		"event_type": "order.created",
		"timestamp":  time.Now().UTC(),
		"data":       order,
	}

	payload, err := json.Marshal(event)
	if err != nil {
		log.Fatal(err)
	}

	topic := "order_created"
	err = p.P.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Key:            []byte(order.ID),
		Value:          payload,
	},
		nil)
	return err

}
