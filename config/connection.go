package config

import (
	"context"
	"log"
	"os"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func KafkaProducer() *kafka.Producer {
	producer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": "localhost:9092",
		"client.id":         ",myproducer",
		"acks":              "all",
	})

	if err != nil {
		log.Fatal("Failed to create producer %s\n", err)
	}

	return producer
}

func KafkaConsumer() *kafka.Consumer {
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": "localhost:9092",
		"group.id":          "dev-payment-group-3",
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		log.Fatal("Failed to create consumer %s\n", err)
	}
	return consumer
}

func Mongodb() *mongo.Client {
	uri := os.Getenv("MONGODB_URI")

	if uri == "" {
		log.Fatal("Set your MONGODB_URI environment variable")
	}

	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal(err)
	}

	if err := client.Ping(context.TODO(), nil); err != nil {
		log.Fatal("MongoDB connection failed:", err)
	}

	log.Println("MongoDB connected")

	return client
}
