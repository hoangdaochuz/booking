---
id: ref-data-shape-mapping
c3-version: 4
title: data-shape-mapping
goal: Define the transformation rules between API response types and frontend display types
scope: [c3-1, c3-6]
---

# data-shape-mapping

## Goal

Define the transformation rules between API response types and frontend display types.

## Choice

Backend returns snake_case JSON with prices in cents. Frontend transforms to camelCase with prices in dollars. Transformation happens in a dedicated transformer layer (`lib/api/transformers.ts`).

## Why

Backend follows Go/protobuf conventions (snake_case). Frontend follows TypeScript conventions (camelCase). Cents avoid floating-point issues in financial calculations. Dedicated transformer layer keeps conversion logic centralized and testable.

## How

| Guideline | Example |
|-----------|---------|
| Prices: cents to dollars | `price_cents: 5000` becomes `price: 50.00` (divide by 100) |
| Fields: snake_case to camelCase | `image_url` becomes `image`, `available_quantity` becomes `available` |
| Dates: RFC3339 to formatted | `2026-03-15T19:00:00Z` becomes `"March 15, 2026, 7:00 PM"` |
| Raw types in `types.ts` | `ApiEvent`, `ApiBooking`, `ApiTicketTier` (match backend JSON) |
| Transformed types in `types.ts` (lib) | `Event`, `PurchasedTicket`, `TicketTier` (frontend display) |
| Transformer functions | `transformEvent()`, `transformBookingToTicket()` |

## Not This

| Alternative | Rejected Because |
|-------------|------------------|
| Backend returns frontend-friendly format | Couples backend to frontend, violates service independence |
| Transform inline in components | Duplicates logic, easy to miss conversions |
| Shared schema generation | Adds build-time complexity for limited benefit |

## Scope

**Applies to:**
- c3-6 (transformer layer) and c3-1 (defines the API shape)

**Does NOT apply to:**
- Backend inter-service communication (protobuf handles that)

## Override

To override this ref:
1. Document justification in an ADR under "Pattern Overrides"
2. Cite this ref and explain why the override is necessary
3. Specify the scope of the override (which components deviate)

## Cited By

- c3-604, c3-111, c3-112
