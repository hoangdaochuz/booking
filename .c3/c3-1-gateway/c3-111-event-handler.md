---
id: c3-111
c3-version: 4
title: event-handler
type: component
category: feature
parent: c3-1
goal: Expose event management endpoints via HTTP REST
summary: Handlers translating HTTP event requests to Event Service gRPC calls
---

# event-handler

## Goal

Expose event management endpoints via HTTP REST.

## Container Connection

Enables event browsing and management through the HTTP API. Without this handler, no client can discover or manage events. Translates HTTP request/response formats to and from the Event Service gRPC protocol, converting protobuf messages to JSON.

## Dependencies

| Direction | What | From/To |
|-----------|------|---------|
| IN (uses) | Event Service gRPC client | For event CRUD and availability queries |
| OUT (provides) | HTTP event responses | JSON event data to API consumers |

## Code References

| File | Purpose |
|------|---------|
| `backend/services/gateway/internal/handler/event_handler.go` | Event CRUD and availability handlers |

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
