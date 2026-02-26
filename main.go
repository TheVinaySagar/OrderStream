package main

import (
	"awesomeProject/config"
	"awesomeProject/handlers"
	"awesomeProject/repositories"
	"context"
	"log"
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

	session, err := client.StartSession()
	if err != nil {
		log.Panic(err)
	}

	defer session.EndSession(context.TODO())

	// init repos
	orderRepo := repositories.NewOrderRepository(db, client)

	//Cons := kafka.NewConsumer(c)

	//Health Check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
		})
	})

	// Order Routes
	r.POST("/orders", handlers.CreateOrder(orderRepo))
	r.GET("/orders/:id", handlers.GetOrder(orderRepo))
	r.PUT("/orders/:id", handlers.UpdateOrder(orderRepo))
	r.DELETE("/orders/:id", handlers.DeleteOrder(orderRepo))

	r.Run(":3000")
}
