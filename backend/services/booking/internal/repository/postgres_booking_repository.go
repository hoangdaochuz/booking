package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ticketbox/booking/internal/domain"
)

type PostgresBookingRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresBookingRepository(pool *pgxpool.Pool) *PostgresBookingRepository {
	return &PostgresBookingRepository{pool: pool}
}

func (r *PostgresBookingRepository) Create(ctx context.Context, booking *domain.Booking) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `INSERT INTO bookings (id, user_id, event_id, status, total_amount_cents, version, created_at)
              VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err = tx.Exec(ctx, query, booking.ID, booking.UserID, booking.EventID,
		string(booking.Status), booking.TotalAmountCents, booking.Version, booking.CreatedAt)
	if err != nil {
		return fmt.Errorf("insert booking: %w", err)
	}

	for _, item := range booking.Items {
		itemQuery := `INSERT INTO booking_items (id, booking_id, ticket_tier_id, quantity, unit_price_cents, seat_ids)
                      VALUES ($1, $2, $3, $4, $5, $6)`
		_, err = tx.Exec(ctx, itemQuery, item.ID, booking.ID, item.TicketTierID, item.Quantity, item.UnitPriceCents, item.SeatIDs)
		if err != nil {
			return fmt.Errorf("insert booking item: %w", err)
		}
	}

	return tx.Commit(ctx)
}

func (r *PostgresBookingRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Booking, error) {
	query := `SELECT id, user_id, event_id, status, total_amount_cents, version, created_at FROM bookings WHERE id = $1`
	b := &domain.Booking{}
	var status string
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&b.ID, &b.UserID, &b.EventID, &status, &b.TotalAmountCents, &b.Version, &b.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get booking: %w", err)
	}
	b.Status = domain.BookingStatus(status)

	itemQuery := `SELECT id, booking_id, ticket_tier_id, quantity, unit_price_cents, seat_ids FROM booking_items WHERE booking_id = $1`
	rows, err := r.pool.Query(ctx, itemQuery, id)
	if err != nil {
		return nil, fmt.Errorf("get items: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var item domain.BookingItem
		if err := rows.Scan(&item.ID, &item.BookingID, &item.TicketTierID, &item.Quantity, &item.UnitPriceCents, &item.SeatIDs); err != nil {
			return nil, fmt.Errorf("scan item: %w", err)
		}
		b.Items = append(b.Items, item)
	}

	return b, nil
}

func (r *PostgresBookingRepository) ListByUserID(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]*domain.Booking, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	var total int
	err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM bookings WHERE user_id = $1", userID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := `SELECT id, user_id, event_id, status, total_amount_cents, version, created_at
              FROM bookings WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	rows, err := r.pool.Query(ctx, query, userID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var bookings []*domain.Booking
	for rows.Next() {
		b := &domain.Booking{}
		var status string
		if err := rows.Scan(&b.ID, &b.UserID, &b.EventID, &status, &b.TotalAmountCents, &b.Version, &b.CreatedAt); err != nil {
			return nil, 0, err
		}
		b.Status = domain.BookingStatus(status)

		itemQuery := `SELECT id, booking_id, ticket_tier_id, quantity, unit_price_cents, seat_ids FROM booking_items WHERE booking_id = $1`
		itemRows, err := r.pool.Query(ctx, itemQuery, b.ID)
		if err != nil {
			return nil, 0, err
		}
		for itemRows.Next() {
			var item domain.BookingItem
			if err := itemRows.Scan(&item.ID, &item.BookingID, &item.TicketTierID, &item.Quantity, &item.UnitPriceCents, &item.SeatIDs); err != nil {
				itemRows.Close()
				return nil, 0, err
			}
			b.Items = append(b.Items, item)
		}
		itemRows.Close()

		bookings = append(bookings, b)
	}

	return bookings, total, nil
}

func (r *PostgresBookingRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.BookingStatus) error {
	_, err := r.pool.Exec(ctx, "UPDATE bookings SET status = $2 WHERE id = $1", id, string(status))
	return err
}
