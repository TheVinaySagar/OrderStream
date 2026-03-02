package order_created

import (
	"awesomeProject/models"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Inventory interface {
	Start(ctx context.Context)
	ProcessMessage(ctx context.Context) error
}
type InventoryService struct {
	DbClient        *mongo.Database
	OrderCollection *mongo.Collection
	Consumer        *kafka.Consumer
}

func NewInventoryService(client *mongo.Database, c *kafka.Consumer) *InventoryService {
	return &InventoryService{
		DbClient:        client,
		OrderCollection: client.Collection("Orders"),
		Consumer:        c,
	}
}

func (inventoryService *InventoryService) Start(ctx context.Context) {

	err := inventoryService.Consumer.Subscribe("dd", nil)
	fmt.Println("Consumer started, waiting for messages ...")
	if err != nil {
		log.Fatal(err)
	}

	//for {
	//
	//}
}

func (inventoryService *InventoryService) ProcessMessage(ctx context.Context) error {

	ev, err := inventoryService.Consumer.ReadMessage(100 * time.Millisecond)
	if err != nil {
		// later will handle error
		return err
	}

	var msg *models.Message
	json.Unmarshal(ev.Value, &msg)

	return nil

}
