---
id: c3-601
c3-version: 4
title: api-client
type: component
category: foundation
parent: c3-6
goal: Provide HTTP client with automatic JWT token management
summary: Singleton apiClient with auto-refresh on 401, proxied via Next.js rewrites
---

# api-client

## Goal

Provide HTTP client with automatic JWT token management.

## Container Connection

All backend communication flows through this client. Without it, no frontend component can reach the gateway API. Manages access and refresh tokens in localStorage, automatically refreshes on 401 responses, and routes requests through Next.js rewrites that proxy `/api/*` to the gateway on port 8000.

## Dependencies

| Direction | What | From/To |
|-----------|------|---------|
| IN (uses) | localStorage | For JWT access and refresh token persistence |
| IN (uses) | Next.js rewrites | Proxy configuration in `next.config.ts` |
| OUT (provides) | Typed API methods | To auth-state (c3-602), booking-state (c3-603), and all page components |

## Code References

| File | Purpose |
|------|---------|
| `frontend/lib/api/client.ts` | Singleton HTTP client with token refresh logic |

## Related Refs

| Ref | How It Serves Goal |
|-----|-------------------|
| ref-jwt-auth-flow | Manages access/refresh tokens client-side |
| ref-data-shape-mapping | Returns raw API types (snake_case, cents) before transformation |

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
