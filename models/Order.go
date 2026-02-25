package models

import (
	"time"
)

type Orders struct {
	ID        string    `bson:"_id" json:"id"`
	Amount    int       `json:"amount"`
	Detail    string    `json:"details"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type OrderCreatedEvent struct {
	EventID   string `json:"event_id"`
	Data      Orders `json:"data"`
	EventType string `json:"event_type"`
}
