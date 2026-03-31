package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ticketbox/payment/internal/domain"
)

var (
	RecordNotFound = errors.New("payment record not found")
)

type PostgresPaymentRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresPaymentRepository(pool *pgxpool.Pool) *PostgresPaymentRepository {
	return &PostgresPaymentRepository{
		pool: pool,
	}
}

func (p *PostgresPaymentRepository) CreatePayment(ctx context.Context, payment *domain.Payment) error {
	query := `INSERT INTO payments (id, user_id, booking_id, order_id, status, price, currency, transaction_id, payment_method, created_at)
			  VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);`

	bookingId := uuid.Nil
	orderId := uuid.Nil
	transactionId := uuid.Nil
	if payment.BookingId != nil {
		bookingId = *payment.BookingId
	}
	if payment.OrderId != nil {
		orderId = *payment.OrderId
	}
	if payment.Transaction_id != nil {
		transactionId = *payment.Transaction_id
	}

	_, err := p.pool.Exec(ctx, query, payment.ID, payment.UserId, bookingId, orderId, string(payment.Status), payment.Price, payment.Currency, transactionId, payment.PaymentMethod, time.Now())
	if err != nil {
		return fmt.Errorf("fail to create payment record: %w", err)
	}
	return nil
}

func (p *PostgresPaymentRepository) GetPaymentByID(ctx context.Context, ID uuid.UUID) (*domain.Payment, error) {
	query := `SELECT p.id, p.user_id, p.booking_id, p.order_id, p.status, p.price, p.currency, p.transaction_id, p.payment_intent_id, p.payment_method, p.created_at
		FROM payments as p
		WHERE p.id = $1;
	`
	result := domain.Payment{}
	var orderId uuid.UUID
	var status string
	var transactionId uuid.UUID
	var bookingId uuid.UUID
	err := p.pool.QueryRow(ctx, query, ID).Scan(&result.ID,
		&result.UserId,
		&bookingId,
		&orderId,
		&status,
		&result.Price,
		&result.Currency,
		&result.PaymentIntentId,
		&transactionId,
		&result.PaymentMethod,
		&result.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, RecordNotFound
		}
		return nil, fmt.Errorf("fail to get payment by id: %w", err)
	}
	result.BookingId = &bookingId
	result.OrderId = &orderId
	result.Transaction_id = &transactionId
	result.Status = domain.PaymentStatus(status)

	return &result, nil
}

func (p *PostgresPaymentRepository) GetListPaymentsByCondition(ctx context.Context, req *domain.GetPaymentsByCondition) ([]domain.Payment, error) {
	return []domain.Payment{}, nil
}

func (p *PostgresPaymentRepository) UpdatePaymentStatus(ctx context.Context, ID uuid.UUID, status domain.PaymentStatus) error {
	query := `UPDATE payments
			  SET status = $1
			  WHERE id = $2;`

	_, err := p.pool.Exec(ctx, query, string(status), ID)
	return err
}

func (p *PostgresPaymentRepository) UpdatePayment(ctx context.Context, payment *domain.Payment) error {

	bookingId := uuid.UUID{}
	orderId := uuid.UUID{}
	transactionId := uuid.UUID{}
	if payment.BookingId != nil {
		bookingId = *payment.BookingId
	}
	if payment.OrderId != nil {
		orderId = *payment.OrderId
	}
	if payment.Transaction_id != nil {
		transactionId = *payment.Transaction_id
	}

	query := `UPDATE payments
			  SET booking_id = $2, order_id = $3, status = $4, price = $5, currency = $6, transaction_id = $7, payment_method = $8, updated_at = $9, payment_intent_id = $10
			  WHERE id = $1`
	_, err := p.pool.Exec(ctx, query, payment.ID, bookingId, orderId, string(payment.Status), payment.Price, payment.Currency, transactionId, payment.PaymentMethod, time.Now(), payment.PaymentIntentId)
	return err
}

func (p *PostgresPaymentRepository) SoftDeletePayment(ctx context.Context, ID uuid.UUID) error {
	_, err := p.pool.Exec(ctx, "UPDATE payments SET deleted_at = $2 WHERE id = $1", ID, time.Now())
	return err
}

func (p *PostgresPaymentRepository) DeletePayment(ctx context.Context, ID uuid.UUID) error {
	_, err := p.pool.Exec(ctx, "DELETE FROM payments WHERE id = $1", ID)
	return err
}
