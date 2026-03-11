# Task 11: Booking Service — Full Implementation (Core)

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build the core booking service with three locking modes (naive/pessimistic/optimistic) that demonstrates the double booking problem.

**Depends on:** Task 9 (Event Service) — Booking Service calls Event Service via gRPC.

**Files:**
- Create: `backend/services/booking/internal/domain/booking.go`
- Create: `backend/services/booking/internal/repository/booking_repository.go`
- Create: `backend/services/booking/internal/repository/postgres_booking_repository.go`
- Create: `backend/services/booking/internal/service/booking_service.go`
- Create: `backend/services/booking/internal/grpc/server.go`
- Create: `backend/services/booking/internal/kafka/producer.go`
- Create: `backend/services/booking/internal/kafka/consumer.go`
- Modify: `backend/services/booking/cmd/main.go`

---

### Step 1: Create domain entities

`backend/services/booking/internal/domain/booking.go`:
```go
package domain

import (
    "time"

    "github.com/google/uuid"
)

type BookingStatus string

const (
    StatusPending   BookingStatus = "PENDING"
    StatusConfirmed BookingStatus = "CONFIRMED"
    StatusFailed    BookingStatus = "FAILED"
    StatusCancelled BookingStatus = "CANCELLED"
)

type Booking struct {
    ID               uuid.UUID
    UserID           uuid.UUID
    EventID          uuid.UUID
    Status           BookingStatus
    TotalAmountCents int64
    Version          int32
    Items            []BookingItem
    CreatedAt        time.Time
}

type BookingItem struct {
    ID             uuid.UUID
    BookingID      uuid.UUID
    TicketTierID   uuid.UUID
    Quantity       int32
    UnitPriceCents int64
}
```

### Step 2: Create repository interface

`backend/services/booking/internal/repository/booking_repository.go`:
```go
package repository

import (
    "context"
    "errors"

    "github.com/google/uuid"
    "github.com/ticketbox/booking/internal/domain"
)

var ErrNotFound = errors.New("not found")

type BookingRepository interface {
    Create(ctx context.Context, booking *domain.Booking) error
    GetByID(ctx context.Context, id uuid.UUID) (*domain.Booking, error)
    ListByUserID(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]*domain.Booking, int, error)
    UpdateStatus(ctx context.Context, id uuid.UUID, status domain.BookingStatus) error
}
```

### Step 3: Implement Postgres booking repository

Standard CRUD following User Service patterns. `Create` inserts both `bookings` and `booking_items` in a transaction.

### Step 4: Create booking service — THE CORE LOGIC

`backend/services/booking/internal/service/booking_service.go`:
```go
package service

import (
    "context"
    "fmt"

    "github.com/google/uuid"
    "go.uber.org/zap"

    eventv1 "github.com/ticketbox/pkg/proto/event/v1"
    "github.com/ticketbox/booking/internal/domain"
    "github.com/ticketbox/booking/internal/repository"
)

type BookingService struct {
    bookingRepo repository.BookingRepository
    eventClient eventv1.EventServiceClient
    logger      *zap.Logger
}

func NewBookingService(
    bookingRepo repository.BookingRepository,
    eventClient eventv1.EventServiceClient,
    logger *zap.Logger,
) *BookingService {
    return &BookingService{
        bookingRepo: bookingRepo,
        eventClient: eventClient,
        logger:      logger,
    }
}

func (s *BookingService) CreateBooking(ctx context.Context, userID, eventID uuid.UUID, items []domain.BookingItem, mode string) (*domain.Booking, error) {
    booking := &domain.Booking{
        ID:      uuid.New(),
        UserID:  userID,
        EventID: eventID,
        Status:  domain.StatusPending,
        Items:   items,
    }

    var totalCents int64
    for i, item := range items {
        // Get current tier info
        tier, err := s.eventClient.GetTicketAvailability(ctx, &eventv1.GetTicketAvailabilityRequest{
            TierId: item.TicketTierID.String(),
        })
        if err != nil {
            return nil, fmt.Errorf("get tier availability: %w", err)
        }

        items[i].UnitPriceCents = tier.PriceCents
        totalCents += tier.PriceCents * int64(item.Quantity)

        // Decrement availability based on booking mode
        switch mode {
        case "naive":
            // NO LOCKING — race condition will occur under load!
            _, err = s.eventClient.UpdateTicketAvailability(ctx, &eventv1.UpdateTicketAvailabilityRequest{
                TierId:        item.TicketTierID.String(),
                QuantityDelta: -item.Quantity,
            })

        case "pessimistic":
            // SELECT FOR UPDATE — safe but slower
            _, err = s.eventClient.UpdateTicketAvailability(ctx, &eventv1.UpdateTicketAvailabilityRequest{
                TierId:          item.TicketTierID.String(),
                QuantityDelta:   -item.Quantity,
                ExpectedVersion: tier.Version, // signals pessimistic mode to Event Service
            })

        case "optimistic":
            // Version check with retry — safe, better throughput
            maxRetries := 3
            for attempt := 0; attempt < maxRetries; attempt++ {
                tier, err = s.eventClient.GetTicketAvailability(ctx, &eventv1.GetTicketAvailabilityRequest{
                    TierId: item.TicketTierID.String(),
                })
                if err != nil {
                    break
                }
                _, err = s.eventClient.UpdateTicketAvailability(ctx, &eventv1.UpdateTicketAvailabilityRequest{
                    TierId:          item.TicketTierID.String(),
                    QuantityDelta:   -item.Quantity,
                    ExpectedVersion: tier.Version,
                })
                if err == nil {
                    break
                }
                s.logger.Warn("Optimistic lock retry", zap.Int("attempt", attempt+1))
            }

        default:
            return nil, fmt.Errorf("unknown booking mode: %s", mode)
        }

        if err != nil {
            booking.Status = domain.StatusFailed
            s.bookingRepo.Create(ctx, booking)
            return booking, fmt.Errorf("reserve tickets failed: %w", err)
        }
    }

    booking.TotalAmountCents = totalCents
    booking.Status = domain.StatusConfirmed
    booking.Items = items

    if err := s.bookingRepo.Create(ctx, booking); err != nil {
        return nil, fmt.Errorf("save booking: %w", err)
    }

    return booking, nil
}

func (s *BookingService) CancelBooking(ctx context.Context, bookingID, userID uuid.UUID) (*domain.Booking, error) {
    booking, err := s.bookingRepo.GetByID(ctx, bookingID)
    if err != nil {
        return nil, err
    }

    if booking.UserID != userID {
        return nil, fmt.Errorf("unauthorized")
    }

    // Restore ticket availability
    for _, item := range booking.Items {
        _, err := s.eventClient.UpdateTicketAvailability(ctx, &eventv1.UpdateTicketAvailabilityRequest{
            TierId:        item.TicketTierID.String(),
            QuantityDelta: item.Quantity,
        })
        if err != nil {
            s.logger.Error("Failed to restore tickets", zap.Error(err))
        }
    }

    booking.Status = domain.StatusCancelled
    if err := s.bookingRepo.UpdateStatus(ctx, bookingID, domain.StatusCancelled); err != nil {
        return nil, err
    }

    return booking, nil
}

func (s *BookingService) GetBooking(ctx context.Context, bookingID uuid.UUID) (*domain.Booking, error) {
    return s.bookingRepo.GetByID(ctx, bookingID)
}

func (s *BookingService) ListUserBookings(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]*domain.Booking, int, error) {
    return s.bookingRepo.ListByUserID(ctx, userID, page, pageSize)
}
```

### Step 5: Create gRPC server

Implement all `BookingService` RPCs. Maps proto requests to service calls. The `booking_mode` field from the request is passed through to `CreateBooking`.

### Step 6: Create Kafka producer

Publishes events to `booking.events` topic:
- `BookingConfirmed` — after successful booking
- `BookingFailed` — after failed booking
- `BookingCancelled` — after cancellation

### Step 7: Create Kafka consumer (CQRS read model)

Consumes `booking.events` and projects into `bookings_read_model` table for fast query reads.

### Step 8: Wire up main.go

- Connect to Postgres
- Connect to Event Service via gRPC client
- Create BookingService
- Start gRPC server on port 50053
- Start Kafka producer + consumer
- Handle graceful shutdown

### Step 9: Install dependencies and verify

```bash
cd /Users/dev/work/booking/backend/services/booking
go get github.com/google/uuid
go get github.com/jackc/pgx/v5
go get go.uber.org/zap
go get google.golang.org/grpc
go get google.golang.org/protobuf
go mod tidy
go build ./...
```
Expected: No errors.

### Step 10: Commit

```bash
git add backend/services/booking/
git commit -m "feat(booking): add booking service with naive/pessimistic/optimistic modes"
```
