---
id: c3-401
c3-version: 4
title: grpc-server
type: component
category: foundation
parent: c3-4
goal: Implement the BookingService protobuf interface
summary: gRPC server translating protobuf requests to booking service layer calls
---

# grpc-server

## Goal

Implement the BookingService protobuf interface.

## Container Connection

Entry point for all booking operations via gRPC. Without this server, the gateway cannot create or manage bookings. Listens on port 50053 and delegates to the booking service layer. Supports configurable `BOOKING_MODE` (pessimistic or optimistic).

## Dependencies

| Direction | What | From/To |
|-----------|------|---------|
| IN (uses) | Booking service business logic | Booking creation (c3-410) and management (c3-411) |
| OUT (provides) | gRPC BookingService interface | To gateway handler (c3-112) |

## Code References

| File | Purpose |
|------|---------|
| `backend/services/booking/internal/grpc/server.go` | BookingService gRPC implementation |
| `backend/services/booking/cmd/main.go` | Server bootstrap and dependency wiring |

## Related Refs

| Ref | How It Serves Goal |
|-----|-------------------|
| ref-grpc-service-structure | Follows standard gRPC service layout |

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
