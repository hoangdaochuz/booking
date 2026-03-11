---
id: c3-101
c3-version: 4
title: http-router
type: component
category: foundation
parent: c3-1
goal: Define HTTP route structure and middleware chain for all API endpoints
summary: Gin router with CORS, auth, and admin middleware wiring
---

# http-router

## Goal

Define HTTP route structure and middleware chain for all API endpoints.

## Container Connection

Without routing, no HTTP requests reach any handler -- the gateway cannot function. This component maps every REST endpoint to its handler and applies the correct middleware (CORS, auth, admin) per route group, fulfilling the gateway's role as the single HTTP entry point for all backend services.

## Dependencies

| Direction | What | From/To |
|-----------|------|---------|
| IN (uses) | auth-middleware | c3-102 for JWT enforcement on protected routes |
| OUT (provides) | Structured routes | To all handler components (c3-110 through c3-113) |

## Code References

| File | Purpose |
|------|---------|
| `backend/services/gateway/internal/router/router.go` | Route definitions and middleware setup |

## Related Refs

| Ref | How It Serves Goal |
|-----|-------------------|
| ref-grpc-service-structure | Follows the gateway variant of the service pattern |

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
