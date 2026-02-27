package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Orders struct {
	ID        string    `bson:"_id" json:"id"`
	Amount    int       `json:"amount"`
	Detail    string    `json:"detail"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type OrderCreatedEvent struct {
	OrderID string `json:"event_id"`
	Total   int    `json:"total"`
}

type Message struct {
	ID            string          `bson:"_id" json:"id"`
	AggregateType string          `bson:"aggregatetype" json:"aggregate_type"`
	AggregateID   string          `bson:"aggregateid" json:"aggregate_id"`
	EventType     string          `bson:"eventtype" json:"event_type"`
	Payload       json.RawMessage `bson:"payload" json:"payload"`
	CreatedAt     time.Time       `bson:"createdat" json:"created_at"`
	ProcessedAt   *time.Time      `bson:"processedat" json:"processed_at"`
	LockedBy      string          `bson:"locked_by,omitempty" json:"locked_by,omitempty"`
	LockedAt      *time.Time      `bson:"locked_at,omitempty" json:"locked_at,omitempty"`
}

func NewMessage(aggregateType, aggregateID, eventType string, payload any) (*Message, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return &Message{
		ID:            uuid.New().String(),
		AggregateType: aggregateType,
		AggregateID:   aggregateID,
		EventType:     eventType,
		Payload:       data,
		CreatedAt:     time.Now(),
	}, nil
}
