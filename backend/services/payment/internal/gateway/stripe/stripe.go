package stripe

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/stripe/stripe-go/v84"
	"github.com/stripe/stripe-go/v84/paymentintent"
	"github.com/stripe/stripe-go/v84/webhook"
	"github.com/ticketbox/payment/internal/domain"
	"github.com/ticketbox/pkg/utils"
	"go.uber.org/zap"
)

type StripePaymentGateway struct {
	logger        *zap.Logger
	webhookSecret string
}

func NewStripePaymentGateway(logger *zap.Logger, webhookSecret string) *StripePaymentGateway {
	return &StripePaymentGateway{
		logger:        logger,
		webhookSecret: webhookSecret,
	}
}

func (s *StripePaymentGateway) CreatePaymentIntent(ctx context.Context, req *domain.CreatePaymentIntentRequest) (*domain.CreatePaymentIntentResponse, error) {
	params := &stripe.PaymentIntentParams{
		Amount:   utils.NewPointer(int64(req.Amount)),
		Currency: utils.NewPointer(req.Currency),
		AutomaticPaymentMethods: &stripe.PaymentIntentAutomaticPaymentMethodsParams{
			Enabled:        &req.AutomaticPaymentMethods.Enabled,
			AllowRedirects: &req.AutomaticPaymentMethods.AllowRedirects,
		},
		// Customer: utils.NewPointer(req.Customer),
		// PaymentMethod: utils.NewPointer(req.PaymentMethod),
		ReceiptEmail: utils.NewPointer(req.ReceiptEmail),
	}
	pi, err := paymentintent.New(params)

	if err != nil {
		s.logger.Sugar().Errorf("[Payment-intent]: Create payment intent fail", zap.Error(err))
		return nil, fmt.Errorf("fail to create a payment intent: %w", err)
	}

	return &domain.CreatePaymentIntentResponse{
		Id:                 pi.ID,
		Object:             pi.Object,
		Amount:             int(pi.Amount),
		AmountReceived:     int(pi.AmountReceived),
		CanceledAt:         pi.CanceledAt,
		CancellationReason: string(pi.CancellationReason),
		ClientSecret:       pi.ClientSecret,
		Created:            pi.Created,
		Currency:           string(pi.Currency),
		// Customer:           pi.Customer.ID,
		ReceiptEmail: pi.ReceiptEmail,
		Status:       string(pi.Status),
	}, nil
}

func (s *StripePaymentGateway) HandleWebhookAfterPayment(ctx context.Context, req *http.Request, res http.ResponseWriter) (*domain.WebhookPaymentResponse, error) {
	const MaxBodyBytes = int64(65536)
	req.Body = http.MaxBytesReader(res, req.Body, MaxBodyBytes)
	payload, err := ioutil.ReadAll(req.Body)
	if err != nil {
		s.logger.Error("fail to read request payment webhook", zap.Error(err))
		return nil, err
	}
	event, err := webhook.ConstructEvent(payload, req.Header.Get("Stripe-Signature"), s.webhookSecret)
	if err != nil {
		s.logger.Error("the webhook event doesn't come from stripe", zap.Error(err))
		return nil, err
	}
	switch event.Type {
	case "payment_intent.succeeded":
		var pi stripe.PaymentIntent
		if err := json.Unmarshal(event.Data.Raw, &pi); err != nil {
			s.logger.Error("fail to read event data raw", zap.Error(err))
			return nil, err
		}
		// TODO

	case "payment_intent.payment_failed":
		var pi stripe.PaymentIntent
		if err := json.Unmarshal(event.Data.Raw, &pi); err != nil {
			s.logger.Error("fail to read event data raw", zap.Error(err))
			return nil, err
		}
		// TODO
	default:
		s.logger.Error("Webhook event unhandled: " + string(event.Type))
		return nil, fmt.Errorf("webhook event %s unhandled", event.Type)
	}
	return &domain.WebhookPaymentResponse{
		EventType: string(event.Type),
	}, nil
}
