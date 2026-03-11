package domain

import (
	"time"

	"github.com/google/uuid"
)

type BookingStatus string

const (
	StatusPending   BookingStatus = "PENDING"
	StatusConfirmed BookingStatus = "CONFIRMED"
	StatusFailed    BookingStatus = "FAILED"
	StatusCancelled BookingStatus = "CANCELLED"
)

type Booking struct {
	ID               uuid.UUID
	UserID           uuid.UUID
	EventID          uuid.UUID
	Status           BookingStatus
	TotalAmountCents int64
	Version          int32
	Items            []BookingItem
	CreatedAt        time.Time
}

type BookingItem struct {
	ID             uuid.UUID
	BookingID      uuid.UUID
	TicketTierID   uuid.UUID
	TierName       string
	Quantity       int32
	UnitPriceCents int64
}
