---
id: c3-310
c3-version: 4
title: event-crud
type: component
category: feature
parent: c3-3
goal: Manage event and ticket tier lifecycle
summary: Create, read, update, delete events with their associated ticket tiers
---

# event-crud

## Goal

Manage event and ticket tier lifecycle.

## Container Connection

Provides the event catalog that users browse and book from. Without event CRUD, no events exist in the system for users to discover or purchase tickets for. Manages the full lifecycle of events and their associated ticket tiers.

## Dependencies

| Direction | What | From/To |
|-----------|------|---------|
| IN (uses) | Event repository | c3-302 for event and tier persistence |
| OUT (provides) | Event lifecycle operations | To gRPC server (c3-301) |

## Code References

| File | Purpose |
|------|---------|
| `backend/services/event/internal/service/event_service.go` | Event CRUD business logic |

## Related Refs

| Ref | How It Serves Goal |
|-----|-------------------|
| ref-repository-pattern | Uses repository for data access |

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
