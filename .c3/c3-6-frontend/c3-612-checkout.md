---
id: c3-612
c3-version: 4
title: checkout
type: component
category: feature
parent: c3-6
goal: Handle cart review and booking purchase
summary: Checkout page with cart summary and purchase confirmation
---

# checkout

## Goal

Handle cart review and booking purchase.

## Container Connection

Converts selected seats into a confirmed booking. Without this, users can select seats but never complete a purchase. Displays the cart summary with pricing and triggers the purchase flow through the booking state context, which calls the backend booking API.

## Dependencies

| Direction | What | From/To |
|-----------|------|---------|
| IN (uses) | Booking state | c3-603 for cart data and purchase flow via useBooking() |
| OUT (provides) | Booking confirmation | Purchase result displayed to user |

## Code References

| File | Purpose |
|------|---------|
| `frontend/app/checkout/page.tsx` | Checkout page with cart summary and purchase |

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
