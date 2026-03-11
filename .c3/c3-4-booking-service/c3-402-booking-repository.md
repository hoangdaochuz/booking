---
id: c3-402
c3-version: 4
title: booking-repository
type: component
category: foundation
parent: c3-4
goal: Persist booking data and outbox entries in PostgreSQL
summary: PGX-based booking CRUD with transactional outbox writes
---

# booking-repository

## Goal

Persist booking data and outbox entries in PostgreSQL.

## Container Connection

Stores booking records and ensures reliable event publishing via the outbox pattern. Without this repository, bookings cannot be persisted and downstream notification events would be lost. Writes booking and outbox records within the same database transaction in the `ticketbox_booking` database (port 5435).

## Dependencies

| Direction | What | From/To |
|-----------|------|---------|
| IN (uses) | PostgreSQL connection pool | From shared database helper (c3-702) |
| OUT (provides) | Booking CRUD + outbox operations | To booking creation (c3-410) and management (c3-411) |

## Code References

| File | Purpose |
|------|---------|
| `backend/services/booking/internal/repository/booking_repository.go` | Repository interface definition |
| `backend/services/booking/internal/repository/postgres_booking_repository.go` | PGX implementation with transactional outbox |

## Related Refs

| Ref | How It Serves Goal |
|-----|-------------------|
| ref-repository-pattern | Interface + PostgreSQL implementation pattern |
| ref-outbox-pattern | Transactional outbox writes alongside booking records |

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
