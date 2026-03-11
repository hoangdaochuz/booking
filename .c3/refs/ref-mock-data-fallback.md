---
id: ref-mock-data-fallback
c3-version: 4
title: mock-data-fallback
goal: Enable frontend development and demo without a running backend
scope: [c3-6]
---

# mock-data-fallback

## Goal

Enable frontend development and demo without a running backend.

## Choice

Frontend BookingProvider tries to fetch events from API, falls back to hardcoded mock data (`lib/mock-data.ts`) on failure. Mock data includes full venue layouts for realistic rendering.

## Why

Frontend development often proceeds independently of backend. Mock data fallback eliminates the need to run 13 Docker containers during UI work. Also useful for demos and testing.

## How

| Guideline | Example |
|-----------|---------|
| Fetch first, fallback second | BookingProvider tries API, catches error, uses mock |
| Mock data matches frontend types | Mock events have same shape as transformed API responses |
| Full venue layouts included | Concert arena and sports stadium with sections, rows, seats |
| No mock auth | Auth pages require backend — mock only covers event browsing |

## Not This

| Alternative | Rejected Because |
|-------------|------------------|
| MSW (Mock Service Worker) | More setup, overkill for simple data fallback |
| Docker-compose for frontend devs | Heavy dependency, slow startup |
| Static JSON files | Less flexible than TypeScript objects with types |

## Scope

**Applies to:**
- c3-6 (BookingProvider only)

**Does NOT apply to:**
- Auth flow or backend services

## Override

To override this ref:
1. Document justification in an ADR under "Pattern Overrides"
2. Cite this ref and explain why the override is necessary
3. Specify the scope of the override (which components deviate)

## Cited By

- c3-603
