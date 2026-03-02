package order_created

import (
	"awesomeProject/models"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const lockTimeout = 5 * time.Minute // stale lock threshold

type OrderServiceInterface interface {
	ProcessMessages(ctx context.Context) error
	Start(ctx context.Context)
}

type OrderService struct {
	Producer   *kafka.Producer
	Client     *mongo.Client
	Collection *mongo.Collection
	Batch      int
	InstanceID string // unique ID for this process instance
}

func NewOrderService(p *kafka.Producer, client *mongo.Client, database *mongo.Database, batch int) *OrderService {
	return &OrderService{
		Producer:   p,
		Client:     client,
		Collection: database.Collection("outbox"),
		Batch:      batch,
		InstanceID: uuid.New().String(),
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
	now := time.Now()
	lockID := orderService.InstanceID

	// Step 1: Claim a batch of unlocked/unprocessed documents (or reclaim stale locks)
	filter := bson.M{
		"$or": []bson.M{
			// Not yet processed and not locked
			{
				"processedat": nil,
				"$or": []bson.M{
					{"locked_by": bson.M{"$exists": false}},
					{"locked_by": ""},
				},
			},
			// Not yet processed but lock is stale (crashed process)
			{
				"processedat": nil,
				"locked_at":   bson.M{"$lt": now.Add(-lockTimeout)},
			},
		},
	}

	claimUpdate := bson.M{
		"$set": bson.M{
			"locked_by": lockID,
			"locked_at": now,
		},
	}

	opts := options.Find().
		SetSort(bson.M{"createdat": 1}).
		SetLimit(int64(orderService.Batch))

	// Find the IDs we want to claim
	cursor, err := orderService.Collection.Find(ctx, filter, opts)
	if err != nil {
		return err
	}

	var candidates []models.Message
	for cursor.Next(ctx) {
		var m models.Message
		if err := cursor.Decode(&m); err != nil {
			return err
		}
		candidates = append(candidates, m)
	}

	if len(candidates) == 0 {
		return nil
	}

	// Collect candidate IDs
	ids := make([]string, len(candidates))
	for i, c := range candidates {
		ids[i] = c.ID
	}

	// Atomically claim them with UpdateMany
	claimFilter := bson.M{
		"_id": bson.M{"$in": ids},
		"$or": []bson.M{
			{"locked_by": bson.M{"$exists": false}},
			{"locked_by": ""},
			{"locked_at": bson.M{"$lt": now.Add(-lockTimeout)}},
		},
	}

	_, err = orderService.Collection.UpdateMany(ctx, claimFilter, claimUpdate)
	if err != nil {
		return fmt.Errorf("failed to claim documents: %w", err)
	}

	// Step 2: Re-fetch only the documents we actually locked
	lockedFilter := bson.M{
		"_id":       bson.M{"$in": ids},
		"locked_by": lockID,
	}

	cursor, err = orderService.Collection.Find(ctx, lockedFilter)
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

	// Step 3: Process the locked messages
	for _, payload := range messages {
		topic := payload.AggregateType + ".published"

		consumerPayload, err := json.Marshal(&payload)
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
			log.Printf("failed to publish message %s: %v", payload.ID, err)
			continue
		}

		// Mark as processed and release the lock
		_, err = orderService.Collection.UpdateOne(
			ctx,
			bson.M{"_id": payload.ID},
			bson.M{
				"$set": bson.M{
					"processedat": time.Now(),
				},
				"$unset": bson.M{
					"locked_by": "",
					"locked_at": "",
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
