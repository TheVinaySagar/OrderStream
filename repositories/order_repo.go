package repositories

import (
	"awesomeProject/models"
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/writeconcern"
)

type OrderRepository struct {
	client           *mongo.Client
	OrderCollection  *mongo.Collection
	OutboxCollection *mongo.Collection
}

func NewOrderRepository(database *mongo.Database, client *mongo.Client) *OrderRepository {
	return &OrderRepository{
		client:           client,
		OrderCollection:  database.Collection("Orders"),
		OutboxCollection: database.Collection("outbox"),
	}
}

func (r *OrderRepository) Create(req *models.Orders) (*models.Orders, error) {
	wc := writeconcern.Majority()
	txnOptions := options.Transaction().SetWriteConcern(wc)

	session, err := r.client.StartSession()
	if err != nil {
		log.Panic(err)
	}

	defer session.EndSession(context.TODO())

	var createdOrder *models.Orders
	err = mongo.WithSession(
		context.TODO(),
		session,
		func(ctx context.Context) error {
			if err := session.StartTransaction(txnOptions); err != nil {
				return err
			}

			order := &models.Orders{
				ID:        uuid.New().String(),
				Amount:    req.Amount,
				Detail:    req.Detail,
				Status:    req.Status,
				CreatedAt: time.Now().UTC(),
			}

			_, err := r.OrderCollection.InsertOne(ctx, order)
			if err != nil {
				return err
			}

			// creating Event
			event := models.OrderCreatedEvent{
				OrderID: order.ID,
				Total:   order.Amount,
			}
			// constructing outbox msg
			msg, err := models.NewMessage("order", order.ID, "order.created", event)
			if err != nil {
				return err
			}
			//inserting in the database
			_, err = r.OutboxCollection.InsertOne(ctx, msg)

			if err != nil {
				return err
			}

			if err = session.CommitTransaction(context.TODO()); err != nil {
				return err
			}

			createdOrder = order
			return err
		})
	if err != nil {
		if err := session.AbortTransaction(context.TODO()); err != nil {
			return nil, err
		}
		return nil, err
	}
	return createdOrder, err
}

func (r *OrderRepository) Update(req *models.Orders) error {
	filter := bson.M{"_id": req.ID}
	_, err := r.OrderCollection.ReplaceOne(context.TODO(), filter, req)

	return err
}

func (r *OrderRepository) Delete(id string) error {
	_, err := r.OrderCollection.DeleteOne(context.TODO(), bson.M{"_id": id})

	return err
}

func (r *OrderRepository) GetbyID(id string) (*models.Orders, error) {
	var order models.Orders

	err := r.OrderCollection.FindOne(context.TODO(), bson.M{"_id": id}).Decode(&order)
	if err != nil {
		return nil, err
	}

	return &order, err
}
