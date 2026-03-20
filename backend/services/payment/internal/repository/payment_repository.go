package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/ticketbox/payment/internal/domain"
)

type PaymentRepositoryInterface interface {
	CreatePayment(ctx context.Context, payment *domain.Payment) error
	GetPaymentByID(ctx context.Context, ID uuid.UUID) (*domain.Payment, error)
	GetListPaymentsByCondition(ctx context.Context, req *domain.GetPaymentsByCondition) ([]domain.Payment, error)
	UpdatePaymentStatus(ctx context.Context, ID uuid.UUID, status domain.PaymentStatus) error
	UpdatePayment(ctx context.Context, payment *domain.Payment) error
	SoftDeletePayment(ctx context.Context, ID uuid.UUID) error
	DeletePayment(ctx context.Context, ID uuid.UUID) error
}
