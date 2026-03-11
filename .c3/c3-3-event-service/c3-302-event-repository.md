---
id: c3-302
c3-version: 4
title: event-repository
type: component
category: foundation
parent: c3-3
goal: Persist event and ticket tier data with row-level locking support
summary: PGX-based event/tier CRUD with SELECT FOR UPDATE for availability
---

# event-repository

## Goal

Persist event and ticket tier data with row-level locking support.

## Container Connection

Provides the atomic data operations that prevent double bookings. Without this repository, events cannot be stored and ticket availability cannot be atomically checked and decremented. Uses `SELECT FOR UPDATE` to serialize concurrent access to ticket tier rows in the `ticketbox_event` database (port 5434).

## Dependencies

| Direction | What | From/To |
|-----------|------|---------|
| IN (uses) | PostgreSQL connection pool | From shared database helper (c3-702) |
| OUT (provides) | Event/tier CRUD with row-level locking | To event CRUD (c3-310) and ticket availability (c3-311) |

## Code References

| File | Purpose |
|------|---------|
| `backend/services/event/internal/repository/event_repository.go` | Repository interface definition |
| `backend/services/event/internal/repository/postgres_event_repository.go` | PGX implementation with SELECT FOR UPDATE |

## Related Refs

| Ref | How It Serves Goal |
|-----|-------------------|
| ref-repository-pattern | Interface + PostgreSQL implementation pattern |
| ref-concurrency-control | Implements row-level locking for availability |

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
