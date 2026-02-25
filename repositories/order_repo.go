package repositories

import (
	"awesomeProject/models"
	"context"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type OrderRepository struct {
	DB *mongo.Collection
}

func NewOrderRepository(db *mongo.Database) *OrderRepository {
	return &OrderRepository{
		DB: db.Collection("Orders"),
	}
}

func (r *OrderRepository) Create(req *models.Orders) (*models.Orders, error) {
	order := models.Orders{
		ID:        uuid.New().String(),
		Amount:    req.Amount,
		Detail:    req.Detail,
		Status:    req.Status,
		CreatedAt: time.Now().UTC(),
	}

	_, err := r.DB.InsertOne(context.TODO(), order)
	return &order, err
}

func (r *OrderRepository) Update(req *models.Orders) error {
	filter := bson.M{"_id": req.ID}
	_, err := r.DB.ReplaceOne(context.TODO(), filter, req)

	return err
}

func (r *OrderRepository) Delete(id string) error {
	_, err := r.DB.DeleteOne(context.TODO(), bson.M{"_id": id})

	return err
}

func (r *OrderRepository) GetbyID(id string) (*models.Orders, error) {
	var order models.Orders

	err := r.DB.FindOne(context.TODO(), bson.M{"_id": id}).Decode(&order)
	if err != nil {
		return nil, err
	}

	return &order, err
}
