---
id: c3-602
c3-version: 4
title: auth-state
type: component
category: foundation
parent: c3-6
goal: Manage global authentication state across the application
summary: AuthProvider React context with login/register/logout and session restore
---

# auth-state

## Goal

Manage global authentication state across the application.

## Container Connection

Without auth state, no component can determine the logged-in user. Provides the AuthProvider React context that wraps the entire app, exposing the `useAuth()` hook with user state, login, register, logout, and automatic session restore from localStorage on mount.

## Dependencies

| Direction | What | From/To |
|-----------|------|---------|
| IN (uses) | API client | c3-601 for auth API calls (login, register, refresh, logout) |
| OUT (provides) | useAuth() hook | User state, login, register, logout to all components |

## Code References

| File | Purpose |
|------|---------|
| `frontend/lib/auth-context.tsx` | AuthProvider and useAuth hook |

## Related Refs

| Ref | How It Serves Goal |
|-----|-------------------|
| ref-jwt-auth-flow | Orchestrates the client-side auth lifecycle |

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
