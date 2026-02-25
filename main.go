package main

import (
	"awesomeProject/config"
	"awesomeProject/handlers"
	"awesomeProject/kafka"
	"awesomeProject/repositories"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	r := gin.Default() // server

	// MongoDB Client
	client := config.Mongodb()
	db := client.Database("NewData")

	// Kafka Client
	p := config.KafkaProducer()
	//c := config.KafkaConsumer()
	//topic := "orders"

	// init repos
	orderRepo := repositories.NewOrderRepository(db)
	Prod := kafka.NewProducer(p)
	//Cons := kafka.NewConsumer(c)

	//Health Check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
		})
	})

	// Order Routes
	r.POST("/orders", handlers.CreateOrder(orderRepo, Prod))
	r.GET("/orders/:id", handlers.GetOrder(orderRepo))
	r.PUT("/orders/:id", handlers.UpdateOrder(orderRepo))
	r.DELETE("/orders/:id", handlers.DeleteOrder(orderRepo))

	r.Run(":3000")
}
