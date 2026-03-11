---
id: c3-703
c3-version: 4
title: kafka-pkg
type: component
category: foundation
parent: c3-7
goal: Provide Kafka producer and consumer utilities for all services
summary: Shared Kafka setup for event publishing and consumption
---

# kafka-pkg

## Goal

Provide Kafka producer and consumer utilities for all services.

## Container Connection

Services use this for reliable async event communication. Without it, the booking service cannot publish outbox events and the notification service cannot consume them. Provides reusable producer and consumer setup that standardizes Kafka connectivity across the platform.

## Dependencies

| Direction | What | From/To |
|-----------|------|---------|
| IN (uses) | Kafka broker address | From config (c3-701) KAFKA_BROKERS |
| OUT (provides) | Producer and consumer instances | Producer to booking service (c3-410), consumer to notification service (c3-501) |

## Code References

| File | Purpose |
|------|---------|
| `backend/pkg/kafka/producer.go` | Kafka event producer |
| `backend/pkg/kafka/consumer.go` | Kafka event consumer |

## Related Refs

| Ref | How It Serves Goal |
|-----|-------------------|
| ref-outbox-pattern | Producer used for outbox event publishing |

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
