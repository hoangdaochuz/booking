package payment_gateway

import (
	"context"
	"net/http"

	"github.com/ticketbox/payment/internal/domain"
)

type PaymentGatewayInterface interface {
	CreatePaymentIntent(ctx context.Context, req *domain.CreatePaymentIntentRequest) (*domain.CreatePaymentIntentResponse, error)
	HandleWebhookAfterPayment(ctx context.Context, req *http.Request, res http.ResponseWriter) (*domain.WebhookPaymentResponse, error)
}
