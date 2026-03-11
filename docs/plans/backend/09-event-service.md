# Task 9: Event Service — Full Implementation

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build the Event Service with CRUD, ticket tier management, and the critical availability update methods (naive/pessimistic/optimistic) that power the double booking demo.

**Files:**
- Create: `backend/services/event/internal/domain/event.go`
- Create: `backend/services/event/internal/repository/event_repository.go`
- Create: `backend/services/event/internal/repository/postgres_event_repository.go`
- Create: `backend/services/event/internal/service/event_service.go`
- Create: `backend/services/event/internal/grpc/server.go`
- Create: `backend/services/event/internal/kafka/producer.go`
- Modify: `backend/services/event/cmd/main.go`

**Follows same Clean Architecture as User Service.**

---

### Step 1: Create domain entities

`backend/services/event/internal/domain/event.go`:
```go
package domain

import (
    "time"

    "github.com/google/uuid"
)

type Event struct {
    ID          uuid.UUID
    Title       string
    Description string
    Category    string
    Venue       string
    Location    string
    Date        time.Time
    ImageURL    string
    Status      string
    Tiers       []TicketTier
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

type TicketTier struct {
    ID                uuid.UUID
    EventID           uuid.UUID
    Name              string
    PriceCents        int64
    TotalQuantity     int32
    AvailableQuantity int32
    Version           int32
    CreatedAt         time.Time
}
```

### Step 2: Create repository interface

`backend/services/event/internal/repository/event_repository.go`:
```go
package repository

import (
    "context"
    "errors"

    "github.com/google/uuid"
    "github.com/ticketbox/event/internal/domain"
)

var (
    ErrNotFound            = errors.New("not found")
    ErrConflict            = errors.New("version conflict")
    ErrInsufficientTickets = errors.New("insufficient tickets")
)

type EventRepository interface {
    Create(ctx context.Context, event *domain.Event) error
    GetByID(ctx context.Context, id uuid.UUID) (*domain.Event, error)
    List(ctx context.Context, category, search string, page, pageSize int) ([]*domain.Event, int, error)
    Update(ctx context.Context, event *domain.Event) error
    Delete(ctx context.Context, id uuid.UUID) error
}

type TicketTierRepository interface {
    GetByID(ctx context.Context, id uuid.UUID) (*domain.TicketTier, error)
    GetByEventID(ctx context.Context, eventID uuid.UUID) ([]domain.TicketTier, error)
    UpdateAvailability(ctx context.Context, tierID uuid.UUID, delta int32, expectedVersion int32) (*domain.TicketTier, error)
    UpdateAvailabilityNoLock(ctx context.Context, tierID uuid.UUID, delta int32) error
    UpdateAvailabilityPessimistic(ctx context.Context, tierID uuid.UUID, delta int32) (*domain.TicketTier, error)
}
```

### Step 3: Implement Postgres event repository

`backend/services/event/internal/repository/postgres_event_repository.go`:

Implement full CRUD for events following the User Service pattern. Key methods for the `TicketTierRepository`:

```go
// Optimistic locking — returns ErrConflict if version mismatch
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

// Pessimistic locking — SELECT FOR UPDATE then UPDATE
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

// Naive — no locking at all (WILL cause double booking under load!)
func (r *PostgresTicketTierRepository) UpdateAvailabilityNoLock(ctx context.Context, tierID uuid.UUID, delta int32) error {
    _, err := r.pool.Exec(ctx,
        `UPDATE ticket_tiers SET available_quantity = available_quantity + $2 WHERE id = $1`,
        tierID, delta)
    return err
}
```

### Step 4: Create event service

`backend/services/event/internal/service/event_service.go`:

Business logic layer — delegates to repository. Includes Redis caching for tier availability:

```go
func (s *EventService) GetTicketAvailability(ctx context.Context, tierID uuid.UUID) (*domain.TicketTier, error) {
    // Try Redis cache first (5s TTL)
    // Fallback to DB
    return s.tierRepo.GetByID(ctx, tierID)
}
```

### Step 5: Create gRPC server

`backend/services/event/internal/grpc/server.go`:

Implement all `EventService` RPCs. The `UpdateTicketAvailability` RPC delegates to the appropriate repository method based on a mode field or request metadata.

### Step 6: Create Kafka producer

`backend/services/event/internal/kafka/producer.go`:

Publishes `EventCreated`, `TicketAvailabilityUpdated` events to `event.events` topic.

### Step 7: Wire up main.go

Same pattern as User Service. Listens on port `50052` (from `GRPC_PORT` env var).

### Step 8: Install dependencies and verify

```bash
cd /Users/dev/work/booking/backend/services/event
go get github.com/google/uuid
go get github.com/jackc/pgx/v5
go get github.com/redis/go-redis/v9
go get go.uber.org/zap
go get google.golang.org/grpc
go get google.golang.org/protobuf
go mod tidy
go build ./...
```
Expected: No errors.

### Step 9: Commit

```bash
git add backend/services/event/
git commit -m "feat(event): add event service with ticket tier availability management"
```
