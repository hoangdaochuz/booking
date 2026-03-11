---
id: c3-613
c3-version: 4
title: my-tickets
type: component
category: feature
parent: c3-6
goal: Display user's purchased ticket history
summary: Ticket listing page showing booking status and details
---

# my-tickets

## Goal

Display user's purchased ticket history.

## Container Connection

Enables users to review their past and upcoming bookings. Without this, users have no way to see their purchase history or check booking status. Retrieves purchased tickets from the booking state context and renders them with status and event details.

## Dependencies

| Direction | What | From/To |
|-----------|------|---------|
| IN (uses) | Booking state | c3-603 for purchased tickets via useBooking() |
| OUT (provides) | Ticket history UI | Purchased ticket listing for end users |

## Code References

| File | Purpose |
|------|---------|
| `frontend/app/my-tickets/page.tsx` | My tickets page |

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
