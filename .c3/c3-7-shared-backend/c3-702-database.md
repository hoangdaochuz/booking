---
id: c3-702
c3-version: 4
title: database
type: component
category: foundation
parent: c3-7
goal: Provide PostgreSQL connection pool creation for all services
summary: PGX pool helper with connection string parsing
---

# database

## Goal

Provide PostgreSQL connection pool creation for all services.

## Container Connection

All services with databases use this to create their connection pool. Without it, no service can connect to its PostgreSQL instance. Parses DATABASE_URL and returns a configured `*pgxpool.Pool` used by all repository implementations across user (port 5433), event (port 5434), booking (port 5435), and notification (port 5436) databases.

## Dependencies

| Direction | What | From/To |
|-----------|------|---------|
| IN (uses) | DATABASE_URL | From config (c3-701) |
| OUT (provides) | *pgxpool.Pool | To all repository components (c3-202, c3-302, c3-402, c3-502) |

## Code References

| File | Purpose |
|------|---------|
| `backend/pkg/database/postgres.go` | PGX pool creation helper |

## Related Refs

*None*

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
