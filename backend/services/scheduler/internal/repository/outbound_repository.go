package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ticketbox/pkg/outbound"
)

type OutBoundEventRepo struct {
	pool *pgxpool.Pool
	tx   *pgx.Tx
}

func NewOutboundEventRepository(pool *pgxpool.Pool, tx *pgx.Tx) *OutBoundEventRepo {
	return &OutBoundEventRepo{
		pool: pool,
		tx:   tx,
	}
}

func (o *OutBoundEventRepo) Create(ctx context.Context, outboundEvent outbound.OutboundEvent) error {
	query := `INSERT INTO outbound_events(topic, event_type, status, payload, created_at) values($1,$2,$3,$4,now())`
	_, err := o.pool.Exec(ctx, query, outboundEvent.Topic, outboundEvent.EventType, outboundEvent.Status, outboundEvent.Payload)
	if err != nil {
		return err
	}
	return nil
}

func (o *OutBoundEventRepo) GetListOutboundPending(ctx context.Context, limit, offset int) ([]outbound.OutboundEvent, error) {
	query := `SELECT id, topic, event_type, status, published_at, payload, created_at FROM outbound_events
			WHERE status = $1 LIMIT $2 OFFSET $3`

	rows, err := o.pool.Query(ctx, query, string(outbound.OutboundStatusPending), limit, offset)
	if err != nil {
		return nil, err
	}
	outboundEvents := []outbound.OutboundEvent{}
	var outboundEvent outbound.OutboundEvent
	var status string
	var createdAt time.Time
	for rows.Next() {
		err := rows.Scan(&outboundEvent.Id, &outboundEvent.Topic, &outboundEvent.EventType, &status, &outboundEvent.PublishAt, &outboundEvent.Payload, &createdAt)
		if err != nil {
			fmt.Println("Fail to scan outbound event: %w", err)
			continue
		}
		outboundEvent.Status = outbound.OutboundStatus(status)
		outboundEvents = append(outboundEvents, outboundEvent)
	}
	return outboundEvents, nil
}

func (o *OutBoundEventRepo) MarkEventAsPublished(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE outbound_events SET published_at = now(), status = $2 WHERE id = $1`
	_, err := o.pool.Exec(ctx, query, id, string(outbound.OutboundStatusPublished))
	return err
}

func (o *OutBoundEventRepo) MarkEventsAsPublished(ctx context.Context, ids []uuid.UUID) error {
	query := `UPDATE outbound_events SET published_at = now(), status = $2 WHERE id = ANY($1)`
	_, err := o.pool.Exec(ctx, query, ids, string(outbound.OutboundStatusPublished))
	return err
}
