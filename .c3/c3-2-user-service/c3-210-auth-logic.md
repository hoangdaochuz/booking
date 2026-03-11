---
id: c3-210
c3-version: 4
title: auth-logic
type: component
category: feature
parent: c3-2
goal: Handle user authentication business logic
summary: Register, login, JWT generation, refresh token rotation, logout
---

# auth-logic

## Goal

Handle user authentication business logic.

## Container Connection

Core authentication flows enabling secure access to the platform. Implements register, login, JWT generation, refresh token rotation, and logout. Without this logic, no user can obtain credentials to access protected resources.

## Dependencies

| Direction | What | From/To |
|-----------|------|---------|
| IN (uses) | User repository | c3-202 for user persistence and credential lookup |
| IN (uses) | Kafka producer | For publishing user registration events |
| OUT (provides) | JWT tokens and user records | To gRPC server (c3-201) |

## Code References

| File | Purpose |
|------|---------|
| `backend/services/user/internal/service/user_service.go` | Auth business logic (register, login, refresh, logout) |

## Related Refs

| Ref | How It Serves Goal |
|-----|-------------------|
| ref-jwt-auth-flow | Generates JWT access and refresh tokens |

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
