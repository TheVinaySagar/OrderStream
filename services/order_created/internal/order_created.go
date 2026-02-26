package internal

import (
	"awesomeProject/models"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type OrderService struct {
	Producer   *kafka.Producer
	client     *mongo.Client
	collection *mongo.Collection
}

func NewOrderService(p *kafka.Producer, client *mongo.Client, database *mongo.Database) *OrderService {
	return &OrderService{
		Producer:   p,
		client:     client,
		collection: database.Collection("outbox"),
	}
}

func (orderService *OrderService) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			if err := orderService.ProcessMessages(ctx); err != nil {
				log.Println("Outbox Error: %v", err)
			}
			time.Sleep(1 * time.Second)
		}
	}
}

func (orderService *OrderService) ProcessMessages(ctx context.Context) error {
	filter := bson.M{
		"$or": []bson.M{
			{"processedat": bson.M{"$exists": false}},
			{"processedat": nil},
		},
	}

	//update := bson.M{
	//	"$set": bson.M{
	//		"locked":    true,
	//		"locked_at": time.Now(),
	//	},
	//}

	opts := options.Find().
		SetSort(bson.M{"createdat": 1})

	cursor, err := orderService.collection.Find(ctx, filter, opts)

	if err != nil {
		return err
	}

	var messages []models.Message

	for cursor.Next(ctx) {
		var m models.Message
		if err := cursor.Decode(&m); err != nil {
			return err
		}
		messages = append(messages, m)
	}

	for _, payload := range messages {
		topic := payload.AggregateType + ".published"

		consumerPayload, err := json.Marshal(&payload)
		fmt.Println(consumerPayload)
		if err != nil {
			return err
		}
		msg := &kafka.Message{
			TopicPartition: kafka.TopicPartition{
				Topic:     &topic,
				Partition: kafka.PartitionAny,
			},
			Key:   []byte(payload.AggregateID),
			Value: consumerPayload,
		}

		if err := orderService.Producer.Produce(msg, nil); err != nil {
			log.Printf("failed to publish message %d: %v", payload.ID, err)
			continue
		}

		_, err = orderService.collection.UpdateOne(
			ctx,
			bson.M{"_id": payload.ID},
			bson.M{
				"$set": bson.M{
					"processedat": time.Now(),
				},
			},
		)

		if err != nil {
			log.Printf("failed to update processed flag: %v", err)
		}
	}
	orderService.Producer.Flush(5000)
	return nil

}
