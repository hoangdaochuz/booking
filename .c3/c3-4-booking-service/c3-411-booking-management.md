---
id: c3-411
c3-version: 4
title: booking-management
type: component
category: feature
parent: c3-4
goal: Handle booking queries and cancellation
summary: Get, list, and cancel bookings via repository
---

# booking-management

## Goal

Handle booking queries and cancellation.

## Container Connection

Enables users to view and manage their booking history. Without this, users have no way to check booking status or cancel unwanted reservations. Provides get-by-ID, list-by-user, and cancel operations through the repository layer.

## Dependencies

| Direction | What | From/To |
|-----------|------|---------|
| IN (uses) | Booking repository | c3-402 for booking data access |
| OUT (provides) | Booking query/cancel results | To gRPC server (c3-401) |

## Code References

| File | Purpose |
|------|---------|
| `backend/services/booking/internal/service/booking_service.go` | Get/list/cancel logic (shared service file with creation) |

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
