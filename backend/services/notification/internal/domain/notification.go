package domain

import (
	"time"

	"github.com/google/uuid"
)

type Notification struct {
	ID        uuid.UUID
	Type      string
	Recipient string
	Channel   string
	Payload   map[string]interface{}
	Status    string
	SentAt    *time.Time
	CreatedAt time.Time
}
