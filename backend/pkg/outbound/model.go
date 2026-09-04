package outbound

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type OutboundStatus string

const (
	OutboundStatusPending   = "pending"
	OutboundStatusPublished = "published"
)

type OutboundEvent struct {
	Id        uuid.UUID
	Topic     string
	EventType string
	Payload   json.RawMessage
	PublishAt time.Time
	Status    OutboundStatus
}
