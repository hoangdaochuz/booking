---
id: c3-311
c3-version: 4
title: ticket-availability
type: component
category: feature
parent: c3-3
goal: Enforce ticket availability invariant under high concurrency
summary: Atomic availability decrements via row-level locking or optimistic versioning
---

# ticket-availability

## Goal

Enforce ticket availability invariant under high concurrency.

## Container Connection

THE critical component -- without it, double bookings occur. When 100K+ users attempt simultaneous purchases, this component ensures `available_quantity` is never decremented below zero. Uses `SELECT FOR UPDATE` on ticket tier rows to serialize concurrent attempts, with an optional optimistic mode via the `version` column.

## Dependencies

| Direction | What | From/To |
|-----------|------|---------|
| IN (uses) | Event repository with locking | c3-302 for `SELECT FOR UPDATE` on ticket_tiers |
| OUT (provides) | Availability check/decrement result | To gRPC server (c3-301), consumed by booking service (c3-410) |

## Code References

| File | Purpose |
|------|---------|
| `backend/services/event/internal/service/event_service.go` | UpdateTicketAvailability logic |

## Related Refs

| Ref | How It Serves Goal |
|-----|-------------------|
| ref-concurrency-control | Implements the core locking strategy (pessimistic and optimistic) |

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
