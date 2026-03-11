---
id: c3-201
c3-version: 4
title: grpc-server
type: component
category: foundation
parent: c3-2
goal: Implement the UserService protobuf interface
summary: gRPC server translating protobuf requests to service layer calls
---

# grpc-server

## Goal

Implement the UserService protobuf interface.

## Container Connection

Entry point for all user service operations via gRPC. Without this server, the gateway has no way to reach user authentication or profile functionality. Listens on port 50051 and delegates to the service layer for business logic.

## Dependencies

| Direction | What | From/To |
|-----------|------|---------|
| IN (uses) | User service business logic | Auth logic (c3-210) and profile management (c3-211) |
| OUT (provides) | gRPC UserService interface | To gateway handlers (c3-110, c3-113) |

## Code References

| File | Purpose |
|------|---------|
| `backend/services/user/internal/grpc/server.go` | UserService gRPC implementation |
| `backend/services/user/cmd/main.go` | Server bootstrap and dependency wiring |

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
