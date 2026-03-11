---
id: ref-concurrency-control
c3-version: 4
title: concurrency-control
goal: Prevent double bookings under 100K+ concurrent users
scope: [c3-3, c3-4]
---

# concurrency-control

## Goal

Prevent double bookings under 100K+ concurrent users.

## Choice

Two modes configurable via `BOOKING_MODE` env var. **Pessimistic (default):** Event Service uses `SELECT ... FOR UPDATE` row-level lock on ticket_tiers, serializing concurrent availability checks. **Optimistic:** Uses `version` column, failing on version mismatch.

## Why

Pessimistic locking is simpler and guarantees correctness under high contention. Optimistic mode available for lower-contention scenarios where throughput matters more. Both prevent the invariant violation: `available_quantity` must never go negative.

## How

| Guideline | Example |
|-----------|---------|
| Pessimistic: SELECT FOR UPDATE | `SELECT ... FROM ticket_tiers WHERE id = $1 FOR UPDATE` |
| Check before decrement | `IF available_quantity >= requested THEN decrement` |
| Optimistic: version check | `UPDATE ... SET version = version + 1 WHERE version = $expected` |
| BOOKING_MODE env var | `pessimistic` (default) or `optimistic` |
| Booking Service calls Event Service | gRPC `UpdateTicketAvailability` for atomic check-and-decrement |

## Not This

| Alternative | Rejected Because |
|-------------|------------------|
| Application-level mutex | Single point of failure, doesn't scale across instances |
| Redis distributed lock | Adds infrastructure dependency to critical path |
| Queue-based serialization | Higher latency for the booking flow |

## Scope

**Applies to:**
- c3-3 (row-level locking in event service repository) and c3-4 (booking creation orchestration)

**Does NOT apply to:**
- Read operations or non-availability updates

## Override

To override this ref:
1. Document justification in an ADR under "Pattern Overrides"
2. Cite this ref and explain why the override is necessary
3. Specify the scope of the override (which components deviate)

## Cited By

- c3-302, c3-311, c3-410
