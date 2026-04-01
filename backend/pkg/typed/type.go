package typed

import "github.com/stripe/stripe-go/v84"

type PaymentEvent struct {
	Id      string
	Data    *stripe.EventData
	Request *stripe.EventRequest
}
