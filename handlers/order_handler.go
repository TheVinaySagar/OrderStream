package handlers

import (
	"awesomeProject/kafka"
	"awesomeProject/models"
	"awesomeProject/repositories"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CreateOrder(repo *repositories.OrderRepository, p *kafka.Producer) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.Orders

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		order, err := repo.Create(&req)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		err = p.PublishOrderCreated(order)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusCreated, order)
	}
}

func UpdateOrder(repo *repositories.OrderRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var updateReq models.Orders

		if err := c.ShouldBindJSON(&updateReq); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		order, err := repo.GetbyID(id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": err.Error(),
			})
			return
		}

		order.Detail = updateReq.Detail
		order.Amount = updateReq.Amount
		order.Status = updateReq.Status

		if err := repo.Update(order); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "Order Updated",
		})

	}
}

func DeleteOrder(repo *repositories.OrderRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		if err := repo.Delete(id); err != nil {
			c.JSON(http.StatusOK, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "Order Deleted",
		})
	}
}

func GetOrder(repo *repositories.OrderRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		order, err := repo.GetbyID(id)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error1": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Success",
			"data":    order,
		})
	}
}
