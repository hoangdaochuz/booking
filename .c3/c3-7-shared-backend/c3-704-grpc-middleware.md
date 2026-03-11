---
id: c3-704
c3-version: 4
title: grpc-middleware
type: component
category: foundation
parent: c3-7
goal: Provide gRPC logging interceptor for all services
summary: Unary interceptor with structured zap logging
---

# grpc-middleware

## Goal

Provide gRPC logging interceptor for all services.

## Container Connection

Consistent request logging across all gRPC services. Without this, gRPC calls across user, event, and booking services would lack structured observability. Implements a unary server interceptor that logs method, duration, and error status using zap structured logging.

## Dependencies

| Direction | What | From/To |
|-----------|------|---------|
| IN (uses) | zap logger | Structured logging library |
| OUT (provides) | gRPC UnaryServerInterceptor | To all gRPC server bootstraps (c3-201, c3-301, c3-401) |

## Code References

| File | Purpose |
|------|---------|
| `backend/pkg/middleware/logging.go` | gRPC logging interceptor |

## Related Refs

| Ref | How It Serves Goal |
|-----|-------------------|
| ref-grpc-service-structure | Part of the standard gRPC service setup |

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
