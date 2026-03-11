---
id: ref-outbox-pattern
c3-version: 4
title: outbox-pattern
goal: Ensure reliable async event publishing without dual-write problems
scope: [c3-4, c3-5]
---

# outbox-pattern

## Goal

Ensure reliable async event publishing without dual-write problems.

## Choice

Booking Service writes booking record AND outbox event in the same database transaction. A separate process publishes outbox events to Kafka. Notification Service consumes from Kafka.

## Why

Writing to the database and Kafka in the same operation risks inconsistency (one succeeds, other fails). The outbox pattern ensures atomicity: if the booking is committed, the event is guaranteed to be published eventually.

## How

| Guideline | Example |
|-----------|---------|
| Outbox table in booking DB | `outbox(id, event_type, payload, published_at)` |
| Same transaction | `BEGIN; INSERT booking; INSERT outbox; COMMIT;` |
| Publisher reads unpublished | Polls outbox for `published_at IS NULL`, publishes to Kafka |
| Consumer processes idempotently | Notification Service handles duplicate deliveries |

## Not This

| Alternative | Rejected Because |
|-------------|------------------|
| Direct Kafka publish in transaction | Dual-write problem — DB commits but Kafka fails |
| Saga pattern | Overkill for notification side-effect |
| CDC (Debezium) | Infrastructure complexity for simple use case |

## Scope

**Applies to:**
- c3-4 (outbox writes in booking service) and c3-5 (Kafka consumption in notification service)

**Does NOT apply to:**
- User or event services which don't use the outbox

## Override

To override this ref:
1. Document justification in an ADR under "Pattern Overrides"
2. Cite this ref and explain why the override is necessary
3. Specify the scope of the override (which components deviate)

## Cited By

- c3-402, c3-410, c3-501, c3-510, c3-703
