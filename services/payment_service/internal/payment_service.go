package internal

import (
	"awesomeProject/models"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/google/uuid"
)

type PaymentHandler interface {
	HandleOrderPayment(msg []byte) error
}

type paymentHandler struct{}

func (h *paymentHandler) HandleOrderPayment(msg []byte) error {
	fmt.Println("Processing order:", string(msg))
	return nil
}

func NewPaymentHandler() PaymentHandler {
	return &paymentHandler{}
}

type PaymentService struct {
	Consumer *kafka.Consumer
	Handler  PaymentHandler
	Producer *kafka.Producer
}

func NewPaymentService(c *kafka.Consumer, h PaymentHandler, p *kafka.Producer) *PaymentService {
	return &PaymentService{
		Consumer: c,
		Handler:  h,
		Producer: p,
	}
}

func (c *PaymentService) Start(topics string) {
	err := c.Consumer.SubscribeTopics([]string{topics}, nil)
	fmt.Println("Consumer started, waiting for messages...")
	if err != nil {
		log.Fatal(err)
	}

	for {
		ev, err := c.Consumer.ReadMessage(100 * time.Millisecond)
		if err != nil {
			// Errors are informational and automatically handled by the consumer
			continue
		}
		//fmt.Println("Received:", string(ev.Value))
		err = c.Handler.HandleOrderPayment(ev.Value)
		if err != nil {
			log.Fatal(err)
		}

		var order models.OrderCreatedEvent

		json.Unmarshal(ev.Value, &order)

		paymentEvent := map[string]interface{}{
			"event_id":   uuid.New().String(),
			"event_type": "payment.completed",
			"timestamp":  time.Now(),
			"data": map[string]interface{}{
				"order_id": order.Data.ID,
				"status":   "paid",
			},
		}
		payload, err := json.Marshal(paymentEvent)

		PaymentTopic := "payment.completed"
		msg := &kafka.Message{
			TopicPartition: kafka.TopicPartition{
				Topic:     &PaymentTopic,
				Partition: kafka.PartitionAny,
			},
			Key:   []byte(order.OrderID),
			Value: payload,
		}
		c.Producer.Produce(msg, nil)
	}

	c.Consumer.Close()
}
