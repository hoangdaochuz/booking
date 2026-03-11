---
id: c3-202
c3-version: 4
title: user-repository
type: component
category: foundation
parent: c3-2
goal: Persist user data in PostgreSQL with secure password storage
summary: PGX-based user CRUD with bcrypt password hashing
---

# user-repository

## Goal

Persist user data in PostgreSQL with secure password storage.

## Container Connection

Without data persistence, no users can register or authenticate. This component provides the data access layer for user records in the `ticketbox_user` database (port 5433), using bcrypt for password hashing and PGX for PostgreSQL connectivity.

## Dependencies

| Direction | What | From/To |
|-----------|------|---------|
| IN (uses) | PostgreSQL connection pool | From shared database helper (c3-702) |
| OUT (provides) | User CRUD operations | To auth logic (c3-210) and profile management (c3-211) |

## Code References

| File | Purpose |
|------|---------|
| `backend/services/user/internal/repository/user_repository.go` | Repository interface definition |
| `backend/services/user/internal/repository/postgres_user_repository.go` | PGX implementation with bcrypt |

## Related Refs

| Ref | How It Serves Goal |
|-----|-------------------|
| ref-repository-pattern | Interface + PostgreSQL implementation pattern |

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
