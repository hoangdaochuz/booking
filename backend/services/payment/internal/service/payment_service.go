package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/ticketbox/payment/internal/domain"
	payment_gateway "github.com/ticketbox/payment/internal/gateway"
	"github.com/ticketbox/payment/internal/repository"
	"go.uber.org/zap"
)

type PaymentService struct {
	repo    repository.PaymentRepositoryInterface
	gateway payment_gateway.PaymentGatewayInterface
	logger  *zap.Logger
}

func NewPaymentService(repo repository.PaymentRepositoryInterface, logger *zap.Logger, gateway payment_gateway.PaymentGatewayInterface) *PaymentService {
	return &PaymentService{
		repo:    repo,
		logger:  logger,
		gateway: gateway,
	}
}
func (p *PaymentService) CreatePayment(ctx context.Context, req *domain.CreatePaymentRequest) (*domain.CreatePaymentResponse, error) {
	err := p.repo.CreatePayment(ctx, &req.Payment)
	if err != nil {
		p.logger.Sugar().Errorf("Fail to create payment: %w", err)
		return nil, err
	}
	res, err := p.gateway.CreatePaymentIntent(ctx, &domain.CreatePaymentIntentRequest{
		Amount:   int(req.Price),
		Currency: req.Currency,
		AutomaticPaymentMethods: domain.AutomaticPaymentMethods{
			Enabled:        true,
			AllowRedirects: "always",
		},
		Customer:      req.UserId.String(),
		PaymentMethod: req.PaymentMethod,
		ReceiptEmail:  req.UserEmail,
	})
	if err != nil {
		return nil, err
	}
	// Update payment intent
	updatePayment := req.Payment
	updatePayment.PaymentIntentId = res.Id
	err = p.repo.UpdatePayment(ctx, &updatePayment)
	if err != nil {
		return nil, fmt.Errorf("fail to update payment after create payment intent: %w", err)
	}

	return &domain.CreatePaymentResponse{
		PaymentItentId:            res.Id,
		PaymentIntentClientSecret: res.ClientSecret,
		Created:                   res.Created,
		Currency:                  res.Currency,
		CustomerID:                res.Customer,
		Status:                    res.Status,
		ReceiptEmail:              res.ReceiptEmail,
	}, nil
}

func (p *PaymentService) GetPaymentById(ctx context.Context, id uuid.UUID) (*domain.Payment, error) {
	return p.repo.GetPaymentByID(ctx, id)
}

func (p *PaymentService) GetPaymentsByCondition(ctx context.Context, req *domain.GetPaymentsByCondition) ([]domain.Payment, error) {
	return []domain.Payment{}, nil
}

func (p *PaymentService) UpdatePaymentStatus(ctx context.Context, id uuid.UUID, status domain.PaymentStatus) error {
	return p.repo.UpdatePaymentStatus(ctx, id, status)
}

func (p *PaymentService) UpdatePayment(ctx context.Context, payment *domain.Payment) error {
	return p.repo.UpdatePayment(ctx, payment)
}

func (p *PaymentService) SoftDeletePayment(ctx context.Context, ID uuid.UUID) error {
	return p.repo.SoftDeletePayment(ctx, ID)
}

func (p *PaymentService) DeletePayment(ctx context.Context, id uuid.UUID) error {
	return p.repo.DeletePayment(ctx, id)
}
