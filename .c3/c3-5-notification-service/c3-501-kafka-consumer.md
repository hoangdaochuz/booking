---
id: c3-501
c3-version: 4
title: kafka-consumer
type: component
category: foundation
parent: c3-5
goal: Subscribe to Kafka booking event topics
summary: Kafka consumer group setup for booking events
---

# kafka-consumer

## Goal

Subscribe to Kafka booking event topics.

## Container Connection

Without the consumer, no booking events are received for notification processing. This component establishes the Kafka consumer group connection and receives raw booking event messages published via the outbox pattern from the booking service.

## Dependencies

| Direction | What | From/To |
|-----------|------|---------|
| IN (uses) | Kafka broker | Message source for booking events |
| IN (uses) | Shared kafka-pkg | c3-703 for consumer setup utilities |
| OUT (provides) | Raw booking event messages | To notification processing (c3-510) |

## Code References

| File | Purpose |
|------|---------|
| `backend/services/notification/internal/kafka/consumer.go` | Kafka consumer setup |
| `backend/services/notification/cmd/main.go` | Consumer bootstrap |

## Related Refs

| Ref | How It Serves Goal |
|-----|-------------------|
| ref-outbox-pattern | Consumes events published via the outbox pattern |

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
