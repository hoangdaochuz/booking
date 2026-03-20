package domain

import (
	"time"

	"github.com/google/uuid"
)

type PaymentStatus string

var (
	PaymentPending PaymentStatus = "pending"
	PaymentSuccess PaymentStatus = "success"
	PaymentFail    PaymentStatus = "fail"
	PaymentCancel  PaymentStatus = "cancel"
	PaymentTimeout PaymentStatus = "timeout"
)

type Payment struct {
	ID             uuid.UUID
	UserId         uuid.UUID
	BookingId      *uuid.UUID
	OrderId        *uuid.UUID
	Status         PaymentStatus
	Price          int32
	Currency       string
	Transaction_id *uuid.UUID
	PaymentMethod  string
	CreatedAt      time.Time
	UpdateAt       time.Time
}

type GetPaymentsByCondition struct {
	Status        string
	UserId        *uuid.UUID
	BookingId     *uuid.UUID
	Price         int32
	Currency      string
	PaymentMethod string
}
