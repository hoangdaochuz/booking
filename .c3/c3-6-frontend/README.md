---
id: c3-6
c3-version: 4
title: frontend
type: container
boundary: app
parent: c3-0
goal: Provide the user-facing ticket booking experience with offline development support
summary: Next.js 16 App Router with auth/booking contexts, API proxy, and mock data fallback
---

# frontend

## Goal

Provide the user-facing ticket booking experience with offline development support

## Responsibilities

- Event browsing and search interface
- Seat selection and booking flow with interactive venue maps
- Auth state management (login/register/logout with session persistence)
- API type transformation (snake_case/cents to camelCase/dollars)

## Complexity Assessment

**Level:** moderate
**Why:** Multiple state contexts, SVG venue generation, type transformation layer

## Components

| ID | Name | Category | Status | Goal Contribution |
|----|------|----------|--------|-------------------|
| c3-601 | api-client | foundation | active | Singleton HTTP client with JWT auto-refresh on 401 |
| c3-602 | auth-state | foundation | active | AuthProvider context for login/register/logout state |
| c3-603 | booking-state | foundation | active | BookingProvider context for events, cart, and purchase flow |
| c3-604 | data-transformers | foundation | active | Converts API snake_case/cents to frontend camelCase/dollars |
| c3-610 | event-browsing | feature | active | Home page event listing with search and filtering |
| c3-611 | event-detail | feature | active | Event detail page with interactive SVG venue map |
| c3-612 | checkout | feature | active | Cart and payment checkout flow |
| c3-613 | my-tickets | feature | active | Purchased ticket history display |
| c3-614 | auth-pages | feature | active | Login and registration form pages |

## Layer Constraints

This container operates within these boundaries:

**MUST:**
- Coordinate components within its boundary
- Define how context linkages are fulfilled internally
- Own its technology stack decisions

**MUST NOT:**
- Define system-wide policies (context responsibility)
- Implement business logic directly (component responsibility)
- Bypass refs for cross-cutting concerns
- Orchestrate other containers (context responsibility)
