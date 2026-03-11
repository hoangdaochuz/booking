---
id: c3-611
c3-version: 4
title: event-detail
type: component
category: feature
parent: c3-6
goal: Show event details with interactive seat selection
summary: Event detail page with SVG venue map for seat picking
---

# event-detail

## Goal

Show event details with interactive seat selection.

## Container Connection

Where users select specific seats before checkout. Without this, users cannot see event details or choose seats. Renders event information, an interactive SVG venue map generated from ticket tier data, and provides seat selection that feeds into the cart.

## Dependencies

| Direction | What | From/To |
|-----------|------|---------|
| IN (uses) | Booking state | c3-603 for event data and cart management via useBooking() |
| IN (uses) | Venue layout generator | `frontend/lib/api/generate-venue-layout.ts` for SVG generation |
| OUT (provides) | Selected seats to cart | Seat selections added to booking state |

## Code References

| File | Purpose |
|------|---------|
| `frontend/app/events/[id]/page.tsx` | Event detail view |
| `frontend/app/events/[id]/seats/page.tsx` | Seat selection page |
| `frontend/components/venue-map.tsx` | Interactive SVG venue component |
| `frontend/lib/api/generate-venue-layout.ts` | SVG venue generation from tier data |

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
