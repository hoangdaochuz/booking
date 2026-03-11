---
id: c3-610
c3-version: 4
title: event-browsing
type: component
category: feature
parent: c3-6
goal: Display event listings with search and filtering
summary: Home page with hero section and event card grid
---

# event-browsing

## Goal

Display event listings with search and filtering.

## Container Connection

The entry point for users to discover available events. Without this, users have no way to see what events are available for booking. Renders the home page with a hero section and a grid of event cards sourced from the booking state context.

## Dependencies

| Direction | What | From/To |
|-----------|------|---------|
| IN (uses) | Booking state | c3-603 for event listing data via useBooking() |
| OUT (provides) | Event listing UI | Home page rendering for end users |

## Code References

| File | Purpose |
|------|---------|
| `frontend/app/page.tsx` | Home page with event listing |
| `frontend/components/event-card.tsx` | Event card component |

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
