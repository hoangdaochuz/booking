---
id: c3-301
c3-version: 4
title: grpc-server
type: component
category: foundation
parent: c3-3
goal: Implement the EventService protobuf interface
summary: gRPC server translating protobuf requests to event service layer calls
---

# grpc-server

## Goal

Implement the EventService protobuf interface.

## Container Connection

Entry point for all event operations via gRPC. Without this server, the gateway cannot serve event data and the booking service cannot check ticket availability. Listens on port 50052 and delegates to the event service layer.

## Dependencies

| Direction | What | From/To |
|-----------|------|---------|
| IN (uses) | Event service business logic | Event CRUD (c3-310) and ticket availability (c3-311) |
| OUT (provides) | gRPC EventService interface | To gateway handler (c3-111) and booking service (c3-410) |

## Code References

| File | Purpose |
|------|---------|
| `backend/services/event/internal/grpc/server.go` | EventService gRPC implementation |
| `backend/services/event/cmd/main.go` | Server bootstrap and dependency wiring |

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
