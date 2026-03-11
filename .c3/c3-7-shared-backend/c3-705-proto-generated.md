---
id: c3-705
c3-version: 4
title: proto-generated
type: component
category: foundation
parent: c3-7
goal: Provide generated protobuf Go code for all service contracts
summary: Generated from proto/ definitions, used by all services for gRPC communication
---

# proto-generated

## Goal

Provide generated protobuf Go code for all service contracts.

## Container Connection

Type-safe service contracts enabling gRPC communication. Without generated stubs, no service can implement or call gRPC interfaces. Generated from proto definitions at `backend/proto/{user,event,booking}/v1/*.proto` via `make proto`, producing Go stubs that all services import for request/response types and client/server interfaces.

## Dependencies

| Direction | What | From/To |
|-----------|------|---------|
| IN (uses) | Proto definitions | `backend/proto/` source .proto files |
| OUT (provides) | Go stubs for user/event/booking services | To all gRPC servers and clients across the platform |

## Code References

| File | Purpose |
|------|---------|
| `backend/pkg/proto/user/v1/user.pb.go` | User service generated stubs |
| `backend/pkg/proto/event/v1/event.pb.go` | Event service generated stubs |
| `backend/pkg/proto/booking/v1/booking.pb.go` | Booking service generated stubs |

## Related Refs

| Ref | How It Serves Goal |
|-----|-------------------|
| ref-grpc-service-structure | Generated code follows the protobuf contract pattern |

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
