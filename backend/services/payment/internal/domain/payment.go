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
	ID              uuid.UUID
	UserId          uuid.UUID
	BookingId       *uuid.UUID
	OrderId         *uuid.UUID
	Status          PaymentStatus
	Price           int32
	Currency        string
	Transaction_id  *uuid.UUID
	PaymentIntentId string
	PaymentMethod   string
	CreatedAt       time.Time
	UpdateAt        time.Time
}

type CreatePaymentRequest struct {
	Payment
	UserEmail string
}

type CreatePaymentResponse struct {
	PaymentItentId            string
	PaymentIntentClientSecret string
	Created                   int64
	Currency                  string
	CustomerID                string
	Status                    string
	ReceiptEmail              string
}

type GetPaymentsByCondition struct {
	Status        string
	UserId        *uuid.UUID
	BookingId     *uuid.UUID
	Price         int32
	Currency      string
	PaymentMethod string
}

type AllowRedirect string

var (
	ALWAYS AllowRedirect = "always"
	NEVER  AllowRedirect = "never"
)

type AutomaticPaymentMethods struct {
	Enabled        bool   `json:"enabled" binding:"required"`
	AllowRedirects string `json:"allow_redirects"`
}

type CreatePaymentIntentRequest struct {
	Amount                  int                     `json:"amount" binding:"required"`
	Currency                string                  `json:"currency" binding:"required"`
	AutomaticPaymentMethods AutomaticPaymentMethods `json:"automatic_payment_methods"`
	Customer                string                  `json:"customer"`
	PaymentMethod           string                  `json:"payment_method"`
	ReceiptEmail            string                  `json:"receipt_email"`
}

type CreatePaymentIntentResponse struct {
	Id                 string
	Object             string
	Amount             int
	AmountReceived     int
	CanceledAt         int64
	CancellationReason string
	ClientSecret       string
	Created            int64
	Currency           string
	Customer           string
	ReceiptEmail       string
	Status             string
}

type WebhookPaymentResponse struct {
	EventType string
}
