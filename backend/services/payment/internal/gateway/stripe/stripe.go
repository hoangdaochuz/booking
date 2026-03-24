package stripe

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v84"
	"github.com/stripe/stripe-go/v84/paymentintent"
	"github.com/stripe/stripe-go/v84/webhook"
	"github.com/ticketbox/payment/internal/domain"
	"github.com/ticketbox/payment/internal/kafka"
	pkgkafka "github.com/ticketbox/pkg/kafka"
	"github.com/ticketbox/pkg/utils"
	"go.uber.org/zap"
)

type StripePaymentGateway struct {
	logger          *zap.Logger
	webhookSecret   string
	paymentProducer *kafka.PaymentProducer
}

func NewStripePaymentGateway(logger *zap.Logger, webhookSecret string, paymentProducer *kafka.PaymentProducer) *StripePaymentGateway {
	return &StripePaymentGateway{
		logger:          logger,
		webhookSecret:   webhookSecret,
		paymentProducer: paymentProducer,
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

type PaymentEvent struct {
	Id      string
	Data    *stripe.EventData
	Request *stripe.EventRequest
}

func (s *StripePaymentGateway) HandleWebhookAfterPayment(ctx context.Context, req *http.Request, res http.ResponseWriter) (*domain.WebhookPaymentResponse, error) {
	const MaxBodyBytes = int64(65536)
	req.Body = http.MaxBytesReader(res, req.Body, MaxBodyBytes)
	payload, err := ioutil.ReadAll(req.Body)
	if err != nil {
		s.logger.Error("fail to read request payment webhook", zap.Error(err))
		return nil, err
	}
	event, err := webhook.ConstructEventWithOptions(payload, req.Header.Get("Stripe-Signature"), s.webhookSecret, webhook.ConstructEventOptions{
		IgnoreAPIVersionMismatch: true,
	})
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
		s.logger.Info("Payment intent create successfully")
		// TODO
		msgKey := uuid.New().String()
		payload := PaymentEvent{
			Id:      event.ID,
			Data:    event.Data,
			Request: event.Request,
		}
		payloadData, err := json.Marshal(payload)
		if err != nil {
			s.logger.Sugar().Errorf("Fail to marshal payment event for %s", event.Type, zap.Error(err))
			return nil, err
		}
		kafkaEvent := pkgkafka.Event{
			Type:      "PaymentSucceed",
			Timestamp: time.Now(),
			Data:      payloadData,
		}
		err = s.paymentProducer.PublishEvent(ctx, "payment.events", msgKey, kafkaEvent)
		if err != nil {
			return nil, fmt.Errorf("Fail to publish event PaymentSucceed to topic payment.events")
		}

	case "payment_intent.payment_failed":
		var pi stripe.PaymentIntent
		if err := json.Unmarshal(event.Data.Raw, &pi); err != nil {
			s.logger.Error("fail to read event data raw", zap.Error(err))
			return nil, err
		}
		s.logger.Info("Payment intent create fail")
		// TODO
		msgKey := uuid.New().String()
		payload := PaymentEvent{
			Id:      event.ID,
			Data:    event.Data,
			Request: event.Request,
		}
		payloadData, err := json.Marshal(payload)
		if err != nil {
			s.logger.Sugar().Errorf("Fail to marshal payment event for %s", event.Type, zap.Error(err))
			return nil, err
		}
		kafkaEvent := pkgkafka.Event{
			Type:      "PaymentFail",
			Timestamp: time.Now(),
			Data:      payloadData,
		}
		err = s.paymentProducer.PublishEvent(ctx, "payment.events", msgKey, kafkaEvent)
		if err != nil {
			return nil, fmt.Errorf("Fail to publish event PaymentFail to topic payment.events")
		}
	default:
		s.logger.Error("Webhook event unhandled: " + string(event.Type))
		return nil, fmt.Errorf("webhook event %s unhandled", event.Type)
	}
	return &domain.WebhookPaymentResponse{
		EventType: string(event.Type),
	}, nil
}
