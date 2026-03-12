# Seat Selection System Design

**Date:** 2026-03-12
**Status:** Design Complete, Ready for Implementation

## Overview

Add seat-level tracking to the TicketBox booking system. Seats are auto-generated when events are created and stored with position data. Users select individual seats which are tracked in the database.

## Architecture

### Backend Changes (Event Service)

- Add `Seat` domain model with position JSONB field
- Auto-generate seats when creating events (based on tier quantities)
- New gRPC methods: `GetSeats`, `UpdateSeatStatus`
- New repository: CRUD operations for seats

### Frontend Changes

- New API client methods: `getSeats()`, `updateSeatStatus()`
- Seat transformer for API responses
- Integration with existing seat selection UI

## Data Model

### Seat Table Structure

| Column | Type | Notes |
|--------|------|-------|
| id | UUID | Primary key |
| event_id | UUID | FK→events(id) |
| ticket_tier_id | UUID | FK→ticket_tiers(id) |
| status | seat_status | ENUM: available/reserved/booked |
| booking_id | UUID | Optional, when booked |
| order_id | UUID | Order tracking |
| position | JSONB | `{"sectionId":"vip-1","row":"A","seat":1,"x":100,"y":200}` |
| created_at, updated_at | TIMESTAMPTZ | Timestamps |
| deleted_at | TIMESTAMPTZ | Soft delete |

## API Contract

### Proto Definitions

```protobuf
message Seat {
    string id = 1;
    string event_id = 2;
    string ticket_tier_id = 3;
    string status = 4;
    string booking_id = 5;
    string order_id = 6;
    string position = 7;
    google.protobuf.Timestamp created_at = 8;
    google.protobuf.Timestamp updated_at = 9;
}

message GetSeatsRequest {
    string event_id = 1;
    string ticket_tier_id = 2;  // optional filter
}

message GetSeatsResponse {
    repeated Seat seats = 1;
}

message UpdateSeatStatusRequest {
    string seat_id = 1;
    string status = 2;
    string booking_id = 3;  // optional
}

message UpdateSeatStatusResponse {
    Seat seat = 1;
}

service EventService {
    rpc GetSeats(GetSeatsRequest) returns (GetSeatsResponse);
    rpc UpdateSeatStatus(UpdateSeatStatusRequest) returns (UpdateSeatStatusResponse);
}
```

## Flows

### Admin: Create Event (Auto-Generate Seats)

1. Admin creates event with ticket tiers and quantities
2. EventService generates seats for each tier based on quantity
3. Position structure: `{"sectionId": tier.name, "row": "A", "seat": 1, "x": 0, "y": 0}`
4. Save: event + tiers + seats in transaction

### User: Book Seats

1. User views event → `GetSeats` API → fetch seats with status/position
2. User selects seat → `UpdateSeatStatus` API (updates status + booking_id)
3. Payment completes → status: "booked"

## Files to Create/Modify

### Backend

| File | Action |
|------|--------|
| `proto/event/v1/event.proto` | Add seat messages & RPC methods |
| `services/event/internal/domain/seat.go` | NEW: Seat domain model |
| `services/event/internal/repository/seat_repository.go` | NEW: Repository interface |
| `services/event/internal/repository/postgres_seat_repository.go` | NEW: Postgres implementation |
| `services/event/internal/service/event_service.go` | Add seat generation to CreateEvent |
| `services/event/internal/grpc/server.go` | Register seat handlers |
| `services/gateway/` | Add HTTP routes for seats |

### Frontend

| File | Action |
|------|--------|
| `lib/api/client.ts` | Add getSeats, updateSeatStatus |
| `lib/api/transformers.ts` | Add transformSeat |
| `lib/api/types.ts` | Add seat types |

## Future Enhancements (Out of Scope)

- Real-time seat availability with Redis
- Seat reservation timeout logic
- Optimistic locking for concurrent seat selection
- Advanced position generation based on venue layouts
