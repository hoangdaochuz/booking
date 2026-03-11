---
id: c3-410
c3-version: 4
title: booking-creation
type: component
category: feature
parent: c3-4
goal: Orchestrate transactional booking creation with cross-service availability check
summary: Creates booking by calling event-service for availability then persisting with outbox
---

# booking-creation

## Goal

Orchestrate transactional booking creation with cross-service availability check.

## Container Connection

The core booking flow -- coordinates availability check and booking persistence. Calls the Event Service gRPC client to atomically decrement ticket availability, then persists the booking record alongside an outbox event in a single transaction. Supports configurable pessimistic/optimistic concurrency mode via `BOOKING_MODE`.

## Dependencies

| Direction | What | From/To |
|-----------|------|---------|
| IN (uses) | Event Service gRPC client | For UpdateTicketAvailability (c3-311) |
| IN (uses) | Booking repository | c3-402 for transactional booking + outbox persistence |
| OUT (provides) | Created booking with outbox event | To gRPC server (c3-401) |

## Code References

| File | Purpose |
|------|---------|
| `backend/services/booking/internal/service/booking_service.go` | CreateBooking orchestration |

## Related Refs

| Ref | How It Serves Goal |
|-----|-------------------|
| ref-concurrency-control | Configurable pessimistic/optimistic booking mode |
| ref-outbox-pattern | Publishes booking events via transactional outbox |

## Layer Constraints

This component operates within these boundaries:

**MUST:**
- Focus on single responsibility within its domain
- Cite refs for patterns instead of re-implementing
- Hand off cross-component concerns to container

**MUST NOT:**
- Import directly from other containers (use container linkages)
- Define system-wide configuration (context responsibility)
- Orchestrate multiple peer components (container responsibility)
- Redefine patterns that exist in refs
