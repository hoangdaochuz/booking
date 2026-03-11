---
id: c3-604
c3-version: 4
title: data-transformers
type: component
category: foundation
parent: c3-6
goal: Convert API response types to frontend display types
summary: Transforms snake_case/cents to camelCase/dollars, formats dates
---

# data-transformers

## Goal

Convert API response types to frontend display types.

## Container Connection

Without transformers, the frontend would display raw API format (price_cents, snake_case field names, RFC3339 timestamps). This component bridges the data shape gap between the Go backend JSON responses and the React frontend expectations, converting cents to dollars, renaming fields (e.g., `image_url` to `image`, `available_quantity` to `available`), and formatting dates.

## Dependencies

| Direction | What | From/To |
|-----------|------|---------|
| IN (uses) | Raw API types | From api-client (c3-601) responses |
| OUT (provides) | Frontend-ready types | Dollars, formatted dates, camelCase to booking-state (c3-603) |

## Code References

| File | Purpose |
|------|---------|
| `frontend/lib/api/transformers.ts` | transformEvent, transformBookingToTicket functions |
| `frontend/lib/api/types.ts` | Raw API type definitions matching backend JSON |

## Related Refs

| Ref | How It Serves Goal |
|-----|-------------------|
| ref-data-shape-mapping | Implements the API-to-Frontend type conversion |

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
