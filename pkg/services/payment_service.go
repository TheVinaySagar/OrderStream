package order_created

import (
	"awesomeProject/models"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
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

		err = c.Handler.HandleOrderPayment(ev.Value)
		if err != nil {
			log.Fatal(err)
		}

		var order models.Message

		json.Unmarshal(ev.Value, &order)

		OrderPaymentEvent := &models.OrderCreatedEvent{
			OrderID:     order.AggregateID,
			TotalAmount: 1000,
		}

		payload, err := models.NewMessage(order.AggregateType, order.AggregateID, "payment.successed", OrderPaymentEvent)
		data, err := json.Marshal(payload)
		Topic := order.AggregateType + ".payment.successed"

		msg := &kafka.Message{
			TopicPartition: kafka.TopicPartition{
				Topic:     &Topic,
				Partition: kafka.PartitionAny,
			},
			Key:   []byte(order.AggregateID),
			Value: data,
		}
		c.Producer.Produce(msg, nil)
	}

	c.Consumer.Close()
}
