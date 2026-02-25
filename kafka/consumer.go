package kafka

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type Consumer struct {
	Consumer *kafka.Consumer
}

func NewConsumer(c *kafka.Consumer) *Consumer {
	return &Consumer{Consumer: c}
}

func (c *Consumer) ConsumeOrderCreated() {
	topics := "order_created"
	err := c.Consumer.SubscribeTopics([]string{topics}, nil)
	// Set up a channel for handling Ctrl-C, etc
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)
	if err != nil {
		log.Fatal(err)
	}

	run := true
	for run {
		select {
		case sig := <-sigchan:
			fmt.Printf("Caught signal %v: terminating\n", sig)
			run = false
		default:
			ev, err := c.Consumer.ReadMessage(100 * time.Millisecond)
			if err != nil {
				// Errors are informational and automatically handled by the consumer
				continue
			}
			fmt.Printf("Consumed event from topic %s: key = %-10s value = %s\n",
				*ev.TopicPartition.Topic, string(ev.Key), string(ev.Value))
		}
	}

	c.Consumer.Close()
}
