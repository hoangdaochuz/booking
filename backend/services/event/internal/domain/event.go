package domain

import (
	"time"

	"github.com/google/uuid"
)

type Event struct {
	ID          uuid.UUID
	Title       string
	Description string
	Category    string
	Venue       string
	Location    string
	Date        time.Time
	ImageURL    string
	Status      string
	Tiers       []TicketTier
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type TicketTier struct {
	ID                uuid.UUID
	EventID           uuid.UUID
	Name              string
	PriceCents        int64
	TotalQuantity     int32
	AvailableQuantity int32
	Version           int32
	CreatedAt         time.Time
}
