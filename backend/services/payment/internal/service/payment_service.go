package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/ticketbox/payment/internal/domain"
	"github.com/ticketbox/payment/internal/repository"
	"go.uber.org/zap"
)

type PaymentService struct {
	repo   repository.PaymentRepositoryInterface
	logger *zap.Logger
}

func NewPaymentService(repo repository.PaymentRepositoryInterface) *PaymentService {
	return &PaymentService{
		repo: repo,
	}
}
func (p *PaymentService) CreatePayment(ctx context.Context, payment *domain.Payment) (uuid.UUID, error) {
	err := p.repo.CreatePayment(ctx, payment)
	if err != nil {
		p.logger.Sugar().Errorf("Fail to create payment: %w", err)
		return uuid.Nil, err
	}
	return payment.ID, nil
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
