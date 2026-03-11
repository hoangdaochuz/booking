---
id: c3-510
c3-version: 4
title: notification-processing
type: component
category: feature
parent: c3-5
goal: Transform Kafka booking events into stored notification records
summary: Deserializes booking events and persists as notifications
---

# notification-processing

## Goal

Transform Kafka booking events into stored notification records.

## Container Connection

The core processing logic that creates notifications from booking events. Without this, consumed Kafka messages are never turned into actionable notification records. Deserializes incoming booking event payloads and persists them via the notification repository.

## Dependencies

| Direction | What | From/To |
|-----------|------|---------|
| IN (uses) | Kafka consumer | c3-501 for raw booking event messages |
| IN (uses) | Notification repository | c3-502 for notification persistence |
| OUT (provides) | Stored notifications | Durable notification records in PostgreSQL |

## Code References

| File | Purpose |
|------|---------|
| `backend/services/notification/internal/service/notification_service.go` | Event processing and notification creation logic |

## Related Refs

| Ref | How It Serves Goal |
|-----|-------------------|
| ref-outbox-pattern | Processes events published via the outbox pattern |

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
