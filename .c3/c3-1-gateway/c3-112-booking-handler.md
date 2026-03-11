---
id: c3-112
c3-version: 4
title: booking-handler
type: component
category: feature
parent: c3-1
goal: Expose booking endpoints via HTTP REST
summary: Handlers translating HTTP booking requests to Booking Service gRPC calls
---

# booking-handler

## Goal

Expose booking endpoints via HTTP REST.

## Container Connection

Enables ticket purchasing through the HTTP API. Without this handler, users cannot create, view, or cancel bookings. Translates HTTP request/response formats to and from the Booking Service gRPC protocol.

## Dependencies

| Direction | What | From/To |
|-----------|------|---------|
| IN (uses) | Booking Service gRPC client | For create, get, list, cancel booking operations |
| OUT (provides) | HTTP booking responses | JSON booking data to API consumers |

## Code References

| File | Purpose |
|------|---------|
| `backend/services/gateway/internal/handler/booking_handler.go` | Create/get/list/cancel booking handlers |

## Related Refs

| Ref | How It Serves Goal |
|-----|-------------------|
| ref-data-shape-mapping | Translates protobuf to JSON response format |

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
