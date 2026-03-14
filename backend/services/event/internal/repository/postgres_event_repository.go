package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ticketbox/event/internal/domain"
)

type PostgresEventRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresEventRepository(pool *pgxpool.Pool) *PostgresEventRepository {
	return &PostgresEventRepository{pool: pool}
}

func (r *PostgresEventRepository) Create(ctx context.Context, event *domain.Event) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `INSERT INTO events (id, title, description, category, venue, location, date, image_url, status, created_at, updated_at)
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`
	_, err = tx.Exec(ctx, query,
		event.ID, event.Title, event.Description, event.Category, event.Venue,
		event.Location, event.Date, event.ImageURL, event.Status, event.CreatedAt, event.UpdatedAt)
	if err != nil {
		return fmt.Errorf("insert event: %w", err)
	}

	for _, tier := range event.Tiers {
		tierQuery := `INSERT INTO ticket_tiers (id, event_id, name, price_cents, total_quantity, available_quantity, version, created_at)
                      VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
		_, err = tx.Exec(ctx, tierQuery,
			tier.ID, event.ID, tier.Name, tier.PriceCents, tier.TotalQuantity,
			tier.AvailableQuantity, 1, tier.CreatedAt)
		if err != nil {
			return fmt.Errorf("insert tier: %w", err)
		}
	}

	return tx.Commit(ctx)
}

func (r *PostgresEventRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Event, error) {
	query := `SELECT id, title, description, category, venue, location, date, image_url, status, created_at, updated_at
              FROM events WHERE id = $1`
	event := &domain.Event{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&event.ID, &event.Title, &event.Description, &event.Category, &event.Venue,
		&event.Location, &event.Date, &event.ImageURL, &event.Status, &event.CreatedAt, &event.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get event: %w", err)
	}

	tierQuery := `SELECT id, event_id, name, price_cents, total_quantity, available_quantity, version, created_at
                  FROM ticket_tiers WHERE event_id = $1`
	rows, err := r.pool.Query(ctx, tierQuery, id)
	if err != nil {
		return nil, fmt.Errorf("get tiers: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var tier domain.TicketTier
		if err := rows.Scan(&tier.ID, &tier.EventID, &tier.Name, &tier.PriceCents,
			&tier.TotalQuantity, &tier.AvailableQuantity, &tier.Version, &tier.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan tier: %w", err)
		}
		event.Tiers = append(event.Tiers, tier)
	}

	return event, nil
}

func (r *PostgresEventRepository) List(ctx context.Context, category, search string, page, pageSize int) ([]*domain.Event, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	baseWhere := " WHERE 1=1"
	args := []interface{}{}
	argIdx := 1

	if category != "" {
		baseWhere += fmt.Sprintf(" AND category = $%d", argIdx)
		args = append(args, category)
		argIdx++
	}
	if search != "" {
		baseWhere += fmt.Sprintf(" AND (title ILIKE $%d OR description ILIKE $%d)", argIdx, argIdx)
		args = append(args, "%"+search+"%")
		argIdx++
	}

	countQuery := "SELECT COUNT(*) FROM events" + baseWhere
	var total int
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count events: %w", err)
	}

	query := fmt.Sprintf("SELECT id, title, description, category, venue, location, date, image_url, status, created_at, updated_at FROM events%s ORDER BY date ASC LIMIT $%d OFFSET $%d", baseWhere, argIdx, argIdx+1)
	args = append(args, pageSize, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list events: %w", err)
	}
	defer rows.Close()

	var events []*domain.Event
	for rows.Next() {
		e := &domain.Event{}
		if err := rows.Scan(&e.ID, &e.Title, &e.Description, &e.Category, &e.Venue,
			&e.Location, &e.Date, &e.ImageURL, &e.Status, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan event: %w", err)
		}
		events = append(events, e)
	}

	// Load tiers for each event
	for _, e := range events {
		tierQuery := `SELECT id, event_id, name, price_cents, total_quantity, available_quantity, version, created_at
                      FROM ticket_tiers WHERE event_id = $1`
		tierRows, err := r.pool.Query(ctx, tierQuery, e.ID)
		if err != nil {
			return nil, 0, fmt.Errorf("get tiers: %w", err)
		}
		for tierRows.Next() {
			var tier domain.TicketTier
			if err := tierRows.Scan(&tier.ID, &tier.EventID, &tier.Name, &tier.PriceCents,
				&tier.TotalQuantity, &tier.AvailableQuantity, &tier.Version, &tier.CreatedAt); err != nil {
				tierRows.Close()
				return nil, 0, fmt.Errorf("scan tier: %w", err)
			}
			e.Tiers = append(e.Tiers, tier)
		}
		tierRows.Close()
	}

	return events, total, nil
}

func (r *PostgresEventRepository) Update(ctx context.Context, event *domain.Event) error {
	query := `UPDATE events SET title = $2, description = $3, category = $4, venue = $5, location = $6, date = $7, image_url = $8, updated_at = $9 WHERE id = $1`
	_, err := r.pool.Exec(ctx, query,
		event.ID, event.Title, event.Description, event.Category, event.Venue,
		event.Location, event.Date, event.ImageURL, time.Now())
	return err
}

func (r *PostgresEventRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM events WHERE id = $1`, id)
	return err
}

// --- Ticket Tier Repository ---

type PostgresTicketTierRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresTicketTierRepository(pool *pgxpool.Pool) *PostgresTicketTierRepository {
	return &PostgresTicketTierRepository{pool: pool}
}

func (r *PostgresTicketTierRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.TicketTier, error) {
	query := `SELECT id, event_id, name, price_cents, total_quantity, available_quantity, version, created_at
              FROM ticket_tiers WHERE id = $1`
	tier := &domain.TicketTier{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&tier.ID, &tier.EventID, &tier.Name, &tier.PriceCents,
		&tier.TotalQuantity, &tier.AvailableQuantity, &tier.Version, &tier.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get tier: %w", err)
	}
	return tier, nil
}

func (r *PostgresTicketTierRepository) GetByEventID(ctx context.Context, eventID uuid.UUID) ([]domain.TicketTier, error) {
	query := `SELECT id, event_id, name, price_cents, total_quantity, available_quantity, version, created_at
              FROM ticket_tiers WHERE event_id = $1`
	rows, err := r.pool.Query(ctx, query, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tiers []domain.TicketTier
	for rows.Next() {
		var t domain.TicketTier
		if err := rows.Scan(&t.ID, &t.EventID, &t.Name, &t.PriceCents,
			&t.TotalQuantity, &t.AvailableQuantity, &t.Version, &t.CreatedAt); err != nil {
			return nil, err
		}
		tiers = append(tiers, t)
	}
	return tiers, nil
}

// Optimistic locking -- returns ErrConflict if version mismatch
func (r *PostgresTicketTierRepository) UpdateAvailability(ctx context.Context, tierID uuid.UUID, delta int32, expectedVersion int32) (*domain.TicketTier, error) {
	query := `UPDATE ticket_tiers
              SET available_quantity = available_quantity + $2, version = version + 1
              WHERE id = $1 AND version = $3 AND available_quantity + $2 >= 0
              RETURNING id, event_id, name, price_cents, total_quantity, available_quantity, version, created_at`
	tier := &domain.TicketTier{}
	err := r.pool.QueryRow(ctx, query, tierID, delta, expectedVersion).Scan(
		&tier.ID, &tier.EventID, &tier.Name, &tier.PriceCents,
		&tier.TotalQuantity, &tier.AvailableQuantity, &tier.Version, &tier.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrConflict
	}
	return tier, err
}

// Pessimistic locking -- SELECT FOR UPDATE then UPDATE
func (r *PostgresTicketTierRepository) UpdateAvailabilityPessimistic(ctx context.Context, tierID uuid.UUID, delta int32) (*domain.TicketTier, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var available int32
	err = tx.QueryRow(ctx,
		`SELECT available_quantity FROM ticket_tiers WHERE id = $1 FOR UPDATE`, tierID).Scan(&available)
	if err != nil {
		return nil, err
	}

	if available+delta < 0 {
		return nil, ErrInsufficientTickets
	}

	tier := &domain.TicketTier{}
	err = tx.QueryRow(ctx,
		`UPDATE ticket_tiers SET available_quantity = available_quantity + $2, version = version + 1
         WHERE id = $1
         RETURNING id, event_id, name, price_cents, total_quantity, available_quantity, version, created_at`,
		tierID, delta).Scan(
		&tier.ID, &tier.EventID, &tier.Name, &tier.PriceCents,
		&tier.TotalQuantity, &tier.AvailableQuantity, &tier.Version, &tier.CreatedAt)
	if err != nil {
		return nil, err
	}

	return tier, tx.Commit(ctx)
}

// Naive -- no locking at all (WILL cause double booking under load!)
func (r *PostgresTicketTierRepository) UpdateAvailabilityNoLock(ctx context.Context, tierID uuid.UUID, delta int32) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE ticket_tiers SET available_quantity = available_quantity + $2 WHERE id = $1`,
		tierID, delta)
	return err
}

// --- Seat Repository ---

type PostgresSeatRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresSeatRepository(pool *pgxpool.Pool) *PostgresSeatRepository {
	return &PostgresSeatRepository{pool: pool}
}

func (r *PostgresSeatRepository) Create(ctx context.Context, seat *domain.Seat) error {
	positionJSON, err := json.Marshal(seat.Position)
	if err != nil {
		return fmt.Errorf("marshal position: %w", err)
	}

	query := `INSERT INTO seats (id, event_id, ticket_tier_id, status, booking_id, order_id, position, created_at, updated_at)
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	// Handle nullable UUID fields
	var bookingID, orderID uuid.UUID
	if seat.BookingID != nil {
		bookingID = *seat.BookingID
	}
	if seat.OrderID != nil {
		orderID = *seat.OrderID
	}

	_, err = r.pool.Exec(ctx, query,
		seat.ID, seat.EventID, seat.TicketTierID, seat.Status,
		bookingID, orderID, positionJSON, seat.CreatedAt, seat.UpdatedAt)
	return err
}

func (r *PostgresSeatRepository) CreateBatch(ctx context.Context, seats []*domain.Seat) error {
	if len(seats) == 0 {
		return nil
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `INSERT INTO seats (id, event_id, ticket_tier_id, status, booking_id, order_id, position, created_at, updated_at)
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	for _, seat := range seats {
		positionJSON, err := json.Marshal(seat.Position)
		if err != nil {
			return fmt.Errorf("marshal position: %w", err)
		}

		// Handle nullable UUID fields
		var bookingID, orderID uuid.UUID
		if seat.BookingID != nil {
			bookingID = *seat.BookingID
		}
		if seat.OrderID != nil {
			orderID = *seat.OrderID
		}

		_, err = tx.Exec(ctx, query,
			seat.ID, seat.EventID, seat.TicketTierID, seat.Status,
			bookingID, orderID, positionJSON, seat.CreatedAt, seat.UpdatedAt)
		if err != nil {
			return fmt.Errorf("insert seat: %w", err)
		}
	}

	return tx.Commit(ctx)
}

func (r *PostgresSeatRepository) GetByEventID(ctx context.Context, eventID uuid.UUID, tierID *uuid.UUID) ([]*domain.Seat, error) {
	var rows pgx.Rows
	var err error

	if tierID != nil {
		query := `SELECT id, event_id, ticket_tier_id, status, booking_id, order_id, position, created_at, updated_at
                  FROM seats WHERE event_id = $1 AND ticket_tier_id = $2 AND deleted_at IS NULL
                  ORDER BY position->>'sectionId', position->>'row', (position->>'seat')::int`
		rows, err = r.pool.Query(ctx, query, eventID, *tierID)
	} else {
		query := `SELECT id, event_id, ticket_tier_id, status, booking_id, order_id, position, created_at, updated_at
                  FROM seats WHERE event_id = $1 AND deleted_at IS NULL
                  ORDER BY position->>'sectionId', position->>'row', (position->>'seat')::int`
		rows, err = r.pool.Query(ctx, query, eventID)
	}

	if err != nil {
		return nil, fmt.Errorf("get seats: %w", err)
	}
	defer rows.Close()

	var seats []*domain.Seat
	for rows.Next() {
		seat, err := r.scanSeat(rows)
		if err != nil {
			return nil, fmt.Errorf("scan seat: %w", err)
		}
		seats = append(seats, seat)
	}

	return seats, nil
}

func (r *PostgresSeatRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Seat, error) {
	query := `SELECT id, event_id, ticket_tier_id, status, booking_id, order_id, position, created_at, updated_at
              FROM seats WHERE id = $1 AND deleted_at IS NULL`
	seat := &domain.Seat{}
	var positionJSON []byte
	var bookingID, orderID uuid.UUID

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&seat.ID, &seat.EventID, &seat.TicketTierID, &seat.Status,
		&bookingID, &orderID, &positionJSON, &seat.CreatedAt, &seat.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get seat: %w", err)
	}

	// Handle nullable UUID fields
	if bookingID != uuid.Nil {
		seat.BookingID = &bookingID
	}
	if orderID != uuid.Nil {
		seat.OrderID = &orderID
	}

	if err := json.Unmarshal(positionJSON, &seat.Position); err != nil {
		return nil, fmt.Errorf("unmarshal position: %w", err)
	}

	return seat, nil
}

func (r *PostgresSeatRepository) UpdateStatus(ctx context.Context, seatID uuid.UUID, status domain.SeatStatus, bookingID *uuid.UUID) (*domain.Seat, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// Lock the seat row and check current status
	var currentStatus domain.SeatStatus
	err = tx.QueryRow(ctx,
		`SELECT status FROM seats WHERE id = $1`, seatID).Scan(&currentStatus)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("lock seat error: %w", err)
	}

	// Prevent double booking: can only update from available to reserved/booked
	if currentStatus != domain.SeatStatusAvailable && status != domain.SeatStatusAvailable {
		return nil, ErrSeatNotAvailable
	}

	// Convert bookingID pointer to uuid.Nil for nil pointer
	var bookingIDValue uuid.UUID
	if bookingID != nil {
		bookingIDValue = *bookingID
	} else {
		bookingIDValue = uuid.Nil
	}

	// Update status
	var seat domain.Seat
	var positionJSON []byte
	var dbBookingID, dbOrderID uuid.UUID
	query := `UPDATE seats SET status = $2, booking_id = $3, updated_at = NOW()
              WHERE id = $1
              RETURNING id, event_id, ticket_tier_id, status, booking_id, order_id, position, created_at, updated_at`
	err = tx.QueryRow(ctx, query, seatID, status, bookingIDValue).Scan(
		&seat.ID, &seat.EventID, &seat.TicketTierID, &seat.Status,
		&dbBookingID, &dbOrderID, &positionJSON, &seat.CreatedAt, &seat.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("update seat: %w", err)
	}

	// Handle nullable UUID fields
	if dbBookingID != uuid.Nil {
		seat.BookingID = &dbBookingID
	}
	if dbOrderID != uuid.Nil {
		seat.OrderID = &dbOrderID
	}

	if err := json.Unmarshal(positionJSON, &seat.Position); err != nil {
		return nil, fmt.Errorf("unmarshal position: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	return &seat, nil
}

// scanSeat is a helper to scan a seat from a database row
func (r *PostgresSeatRepository) scanSeat(rows pgx.Rows) (*domain.Seat, error) {
	seat := &domain.Seat{}
	var positionJSON []byte
	var bookingID, orderID uuid.UUID

	err := rows.Scan(
		&seat.ID, &seat.EventID, &seat.TicketTierID, &seat.Status,
		&bookingID, &orderID, &positionJSON, &seat.CreatedAt, &seat.UpdatedAt)
	if err != nil {
		return nil, err
	}

	// Handle nullable UUID fields
	if bookingID != uuid.Nil {
		seat.BookingID = &bookingID
	}
	if orderID != uuid.Nil {
		seat.OrderID = &orderID
	}

	if err := json.Unmarshal(positionJSON, &seat.Position); err != nil {
		return nil, fmt.Errorf("unmarshal position: %w", err)
	}

	return seat, nil
}

func (r *PostgresSeatRepository) GetByTierID(ctx context.Context, tierID uuid.UUID) ([]*domain.Seat, error) {
	query := `SELECT id, event_id, ticket_tier_id, status, booking_id, order_id, position, created_at, updated_at
              FROM seats WHERE ticket_tier_id = $1 AND deleted_at IS NULL
              ORDER BY position->>'sectionId', position->>'row', (position->>'seat')::int`
	rows, err := r.pool.Query(ctx, query, tierID)
	if err != nil {
		return nil, fmt.Errorf("get seats by tier: %w", err)
	}
	defer rows.Close()

	var seats []*domain.Seat
	for rows.Next() {
		seat, err := r.scanSeat(rows)
		if err != nil {
			return nil, fmt.Errorf("scan seat: %w", err)
		}
		seats = append(seats, seat)
	}

	return seats, nil
}

func (r *PostgresSeatRepository) UpdateStatusBatch(ctx context.Context, seatIDs []uuid.UUID, status domain.SeatStatus, bookingID *uuid.UUID) error {
	if len(seatIDs) == 0 {
		return nil
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// Use pessimistic locking to prevent race conditions
	query := `SELECT id, status FROM seats WHERE id = ANY($1) AND deleted_at IS NULL`
	rows, err := tx.Query(ctx, query, seatIDs)
	if err != nil {
		return fmt.Errorf("lock seats: %w", err)
	}
	defer rows.Close()

	// Verify all seats exist and are available (if transitioning to reserved/booked)
	var lockedIDs []uuid.UUID
	var currentSeatsStatus []domain.SeatStatus
	for rows.Next() {
		var id uuid.UUID
		var status domain.SeatStatus
		if err := rows.Scan(&id, &status); err != nil {
			return fmt.Errorf("scan locked seats: %w", err)
		}
		lockedIDs = append(lockedIDs, id)
		if len(currentSeatsStatus) > 1 && currentSeatsStatus[len(currentSeatsStatus)-1] != status {
			return fmt.Errorf("Exist seat has been inconsistent status")
		}
		currentSeatsStatus = append(currentSeatsStatus, status)
	}

	if len(lockedIDs) != len(seatIDs) {
		return ErrNotFound
	}

	// Prevent double booking: can only update from available to reserved/booked
	if currentSeatsStatus[0] != domain.SeatStatusAvailable && status != domain.SeatStatusAvailable {
		return ErrSeatNotAvailable
	}

	// Convert bookingID pointer to uuid.Nil for nil pointer
	var bookingIDValue uuid.UUID
	if bookingID != nil {
		bookingIDValue = *bookingID
	} else {
		bookingIDValue = uuid.Nil
	}

	// Update all seats
	updateQuery := `UPDATE seats SET status = $2, booking_id = $3, updated_at = NOW()
                    WHERE id = ANY($1) AND deleted_at IS NULL`
	result, err := tx.Exec(ctx, updateQuery, seatIDs, status, bookingIDValue)
	if err != nil {
		return fmt.Errorf("update seats: %w", err)
	}

	if result.RowsAffected() != int64(len(seatIDs)) {
		return ErrNotFound
	}

	return tx.Commit(ctx)
}

func (r *PostgresSeatRepository) GetAvailableSeats(ctx context.Context, eventID uuid.UUID) ([]*domain.Seat, error) {
	query := `SELECT id, event_id, ticket_tier_id, status, booking_id, order_id, position, created_at, updated_at
              FROM seats WHERE event_id = $1 AND status = 'available' AND deleted_at IS NULL
              ORDER BY position->>'sectionId', position->>'row', (position->>'seat')::int`
	rows, err := r.pool.Query(ctx, query, eventID)
	if err != nil {
		return nil, fmt.Errorf("get available seats: %w", err)
	}
	defer rows.Close()

	var seats []*domain.Seat
	for rows.Next() {
		seat, err := r.scanSeat(rows)
		if err != nil {
			return nil, fmt.Errorf("scan seat: %w", err)
		}
		seats = append(seats, seat)
	}

	return seats, nil
}

func (r *PostgresSeatRepository) GetAvailableSeatsByTier(ctx context.Context, tierID uuid.UUID) ([]*domain.Seat, error) {
	query := `SELECT id, event_id, ticket_tier_id, status, booking_id, order_id, position, created_at, updated_at
              FROM seats WHERE ticket_tier_id = $1 AND status = 'available' AND deleted_at IS NULL
              ORDER BY position->>'sectionId', position->>'row', (position->>'seat')::int`
	rows, err := r.pool.Query(ctx, query, tierID)
	if err != nil {
		return nil, fmt.Errorf("get available seats by tier: %w", err)
	}
	defer rows.Close()

	var seats []*domain.Seat
	for rows.Next() {
		seat, err := r.scanSeat(rows)
		if err != nil {
			return nil, fmt.Errorf("scan seat: %w", err)
		}
		seats = append(seats, seat)
	}

	return seats, nil
}

// Delete soft deletes a seat by setting deleted_at
func (r *PostgresSeatRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE seats SET deleted_at = NOW() WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

// DeleteByEventID soft deletes all seats for an event
func (r *PostgresSeatRepository) DeleteByEventID(ctx context.Context, eventID uuid.UUID) error {
	query := `UPDATE seats SET deleted_at = NOW() WHERE event_id = $1 AND deleted_at IS NULL`
	_, err := r.pool.Exec(ctx, query, eventID)
	return err
}
