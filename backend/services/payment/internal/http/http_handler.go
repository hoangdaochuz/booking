package payment_http

import (
	"context"
	"net/http"
	"time"

	payment_gateway "github.com/ticketbox/payment/internal/gateway"
	"github.com/ticketbox/payment/internal/service"
	"go.uber.org/zap"
)

type PaymentHttpHandler struct {
	service        *service.PaymentService
	logger         *zap.Logger
	webhookHandler payment_gateway.PaymentGatewayInterface
}

func NewPaymentHttpHandler(service *service.PaymentService, logger *zap.Logger, webhookHandler payment_gateway.PaymentGatewayInterface) *PaymentHttpHandler {
	return &PaymentHttpHandler{
		service:        service,
		logger:         logger,
		webhookHandler: webhookHandler,
	}
}

func (p *PaymentHttpHandler) WebhookHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	p.logger.Info("Webhook handler.....")
	_, err := p.webhookHandler.HandleWebhookAfterPayment(ctx, r, w)
	if err != nil {
		p.logger.Error("Fail to handle webhook payment", zap.Error(err))
	}
	p.logger.Info("Handle webhook payment successfully")
}
