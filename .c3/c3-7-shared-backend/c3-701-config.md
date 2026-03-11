---
id: c3-701
c3-version: 4
title: config
type: component
category: foundation
parent: c3-7
goal: Provide environment-based configuration loading for all services
summary: Viper config loading SERVICE_NAME, ports, DATABASE_URL, JWT_SECRET
---

# config

## Goal

Provide environment-based configuration loading for all services.

## Container Connection

Every service needs configuration -- this provides the common loader. Without it, services cannot read DATABASE_URL, JWT_SECRET, service ports, or BOOKING_MODE from the environment. Uses Viper for environment variable binding and returns a typed config struct.

## Dependencies

| Direction | What | From/To |
|-----------|------|---------|
| IN (uses) | Environment variables | SERVICE_NAME, DATABASE_URL, JWT_SECRET, ports, KAFKA_BROKERS |
| OUT (provides) | Typed config struct | To all service bootstrap code (cmd/main.go files) |

## Code References

| File | Purpose |
|------|---------|
| `backend/pkg/config/config.go` | Viper-based config loading |

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
