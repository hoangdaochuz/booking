---
id: c3-110
c3-version: 4
title: auth-handler
type: component
category: feature
parent: c3-1
goal: Expose authentication endpoints via HTTP REST
summary: Handlers translating HTTP auth requests to User Service gRPC calls
---

# auth-handler

## Goal

Expose authentication endpoints via HTTP REST.

## Container Connection

Enables user registration and login through the HTTP API. Without this handler, clients have no way to authenticate with the platform. Translates HTTP request/response formats to and from the User Service gRPC protocol.

## Dependencies

| Direction | What | From/To |
|-----------|------|---------|
| IN (uses) | User Service gRPC client | For register, login, refresh, logout operations |
| OUT (provides) | HTTP auth responses | JWT tokens and user data to API consumers |

## Code References

| File | Purpose |
|------|---------|
| `backend/services/gateway/internal/handler/auth_handler.go` | Register/login/refresh/logout handlers |

## Related Refs

| Ref | How It Serves Goal |
|-----|-------------------|
| ref-jwt-auth-flow | Returns JWT tokens from auth operations |

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
