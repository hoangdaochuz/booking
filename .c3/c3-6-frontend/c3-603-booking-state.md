---
id: c3-603
c3-version: 4
title: booking-state
type: component
category: foundation
parent: c3-6
goal: Manage global event listing, cart, and purchase state
summary: BookingProvider React context with events, cart, purchase flow, and mock data fallback
---

# booking-state

## Goal

Manage global event listing, cart, and purchase state.

## Container Connection

Central state for the booking flow -- events, cart selection, and purchase. Without this, no page can access the event catalog or manage cart items. Provides the BookingProvider React context with `useBooking()` hook, fetching events from the API and falling back to mock data when the backend is unavailable.

## Dependencies

| Direction | What | From/To |
|-----------|------|---------|
| IN (uses) | API client | c3-601 for event and booking API calls |
| IN (uses) | Data transformers | c3-604 for converting API responses to frontend types |
| IN (uses) | Mock data | `frontend/lib/mock-data.ts` for offline fallback |
| OUT (provides) | useBooking() hook | Events, cart, purchase flow to all page components |

## Code References

| File | Purpose |
|------|---------|
| `frontend/lib/booking-context.tsx` | BookingProvider and useBooking hook |

## Related Refs

| Ref | How It Serves Goal |
|-----|-------------------|
| ref-mock-data-fallback | Falls back to mock data when backend is unavailable |

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
