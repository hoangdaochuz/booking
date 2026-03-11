---
id: c3-113
c3-version: 4
title: user-handler
type: component
category: feature
parent: c3-1
goal: Expose user profile endpoints via HTTP REST
summary: Handlers translating HTTP profile requests to User Service gRPC calls
---

# user-handler

## Goal

Expose user profile endpoints via HTTP REST.

## Container Connection

Enables profile management through the HTTP API. Without this handler, authenticated users cannot view or update their profile information. Requires authenticated user context from auth-middleware (c3-102).

## Dependencies

| Direction | What | From/To |
|-----------|------|---------|
| IN (uses) | User Service gRPC client | For get/update profile operations |
| OUT (provides) | HTTP user profile responses | JSON profile data to API consumers |

## Code References

| File | Purpose |
|------|---------|
| `backend/services/gateway/internal/handler/user_handler.go` | Get/update profile handlers |

## Related Refs

| Ref | How It Serves Goal |
|-----|-------------------|
| ref-jwt-auth-flow | Requires authenticated user context |

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
