package main

import (
	"awesomeProject/config"
	"awesomeProject/services/order_created/internal"
	"context"

	"github.com/joho/godotenv"
)

func init() {
	godotenv.Load()
}

func main() {
	kafkaProducer := config.KafkaProducer()
	Client := config.Mongodb()
	Database := Client.Database("NewData")
	ctx := context.Background()
	producer := internal.NewOrderService(kafkaProducer, Client, Database, 10)
	producer.Start(ctx)

}
