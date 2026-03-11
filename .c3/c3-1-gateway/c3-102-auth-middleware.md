---
id: c3-102
c3-version: 4
title: auth-middleware
type: component
category: foundation
parent: c3-1
goal: Validate JWT tokens and enforce authentication on protected routes
summary: JWT parsing middleware with Redis token blacklist integration
---

# auth-middleware

## Goal

Validate JWT tokens and enforce authentication on protected routes.

## Container Connection

Without auth middleware, protected endpoints would be publicly accessible. This component intercepts requests before they reach handlers, verifying JWT validity and checking the Redis blacklist for revoked tokens. It injects authenticated user context that downstream handlers depend on.

## Dependencies

| Direction | What | From/To |
|-----------|------|---------|
| IN (uses) | Redis connection | Token blacklist store for revoked tokens |
| IN (uses) | JWT secret | From config (c3-701) for token verification |
| OUT (provides) | Authenticated user context | To all protected route handlers |

## Code References

| File | Purpose |
|------|---------|
| `backend/services/gateway/internal/middleware/auth.go` | JWT validation and blacklist check |
| `backend/services/gateway/internal/middleware/cors.go` | CORS policy configuration |

## Related Refs

| Ref | How It Serves Goal |
|-----|-------------------|
| ref-jwt-auth-flow | Implements the gateway side of JWT validation |

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
