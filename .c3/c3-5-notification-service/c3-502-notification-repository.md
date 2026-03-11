---
id: c3-502
c3-version: 4
title: notification-repository
type: component
category: foundation
parent: c3-5
goal: Persist notification records in PostgreSQL
summary: PGX-based notification storage
---

# notification-repository

## Goal

Persist notification records in PostgreSQL.

## Container Connection

Without persistence, notifications are lost after processing. This component stores notification records in the `ticketbox_notification` database (port 5436), providing a durable record of all notifications generated from booking events.

## Dependencies

| Direction | What | From/To |
|-----------|------|---------|
| IN (uses) | PostgreSQL connection pool | From shared database helper (c3-702) |
| OUT (provides) | Notification CRUD operations | To notification processing (c3-510) |

## Code References

| File | Purpose |
|------|---------|
| `backend/services/notification/internal/repository/notification_repository.go` | Repository interface definition |
| `backend/services/notification/internal/repository/postgres_notification_repository.go` | PGX implementation |

## Related Refs

| Ref | How It Serves Goal |
|-----|-------------------|
| ref-repository-pattern | Interface + PostgreSQL implementation pattern |

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
