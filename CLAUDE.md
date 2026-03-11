# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

TicketBox is a ticket booking platform solving the **double booking problem** under high concurrency (100K+ simultaneous users). Go microservices backend with gRPC, Next.js frontend, PostgreSQL per service, Kafka for async events.

## Commands

### Backend (from `backend/`)

```bash
make up              # Start all 13 Docker containers
make down            # Stop everything
make migrate         # Run golang-migrate migrations for all 4 databases
make test            # Run tests for all services
make test-load       # Load test: 100K concurrent bookings
make proto           # Regenerate protobuf Go code from proto/ definitions
make build           # Build all service binaries to bin/
make logs            # Tail Docker Compose logs
```

Run a single service's tests:
```bash
cd backend/services/booking && go test ./... -v
cd backend/services/event && go test ./... -v -run TestSpecificFunction
```

### Frontend (from `frontend/`)

```bash
npm run dev          # Dev server on :3000
npm run build        # Production build
npm run lint         # ESLint (next/core-web-vitals + typescript)
npx tsc --noEmit     # Type check only
```

## Architecture

### Backend: Go Microservices with gRPC

Five services communicate via gRPC, exposed through an HTTP gateway:

- **Gateway** (:8000, Gin) — HTTP REST API, translates to gRPC calls. Handles JWT validation via middleware.
- **User Service** (:50051) — Auth (register/login/refresh/logout), profile management. DB: `ticketbox_user` on port 5433.
- **Event Service** (:50052) — CRUD events, ticket tiers, **atomic availability updates** with `SELECT FOR UPDATE`. DB: `ticketbox_event` on port 5434.
- **Booking Service** (:50053) — Creates bookings transactionally, calls event service for availability. DB: `ticketbox_booking` on port 5435. Configurable `BOOKING_MODE`: pessimistic (default) or optimistic.
- **Notification Service** — Kafka consumer, no gRPC port. DB: `ticketbox_notification` on port 5436.

**Go workspace** (`backend/go.work`) links all modules. Each service has `replace github.com/ticketbox/pkg => ../../pkg` in its go.mod.

**Shared code** lives in `backend/pkg/`: config, database helpers, kafka producer/consumer, gRPC middleware, generated proto code (`pkg/proto/{service}/v1/`).

**Proto definitions** at `backend/proto/{user,event,booking}/v1/*.proto`. Run `make proto` after changes.

**Migrations** per service at `backend/services/{name}/migrations/` using golang-migrate format (`000001_init.up.sql` / `000001_init.down.sql`).

### Frontend: Next.js 16 App Router

Path alias: `@/*` maps to project root (e.g., `@/lib/api/client`, `@/components/navbar`).

Key layers:
- **`lib/api/client.ts`** — Singleton `apiClient` with JWT token management (auto-refresh on 401), talks to gateway via `/api/*` (proxied by Next.js rewrites in `next.config.ts`).
- **`lib/api/types.ts`** — Raw API response types matching backend JSON (snake_case, cents for prices).
- **`lib/api/transformers.ts`** — Converts API types to frontend types (cents→dollars, RFC3339→formatted strings, field renames).
- **`lib/api/generate-venue-layout.ts`** — Generates SVG venue seat maps from API tier data (concert arena or sports stadium layouts).
- **`lib/auth-context.tsx`** — AuthProvider wrapping the app. Login/register/logout + session restore from localStorage.
- **`lib/booking-context.tsx`** — BookingProvider fetches events from API with **mock data fallback** when backend is offline. Manages cart and purchase flow.
- **`lib/mock-data.ts`** — Hardcoded events with full venue layouts for offline development.

### Double Booking Prevention

The critical path: Booking Service → Event Service `UpdateTicketAvailability` → `SELECT ticket_tiers FOR UPDATE WHERE id = $1` → check `available_quantity >= requested` → decrement → commit. Row-level lock serializes concurrent attempts. The `version` column enables optimistic mode as an alternative.

### Data Shape Mapping

Backend returns `price_cents` (int64), `image_url`, `available_quantity`, `total_amount_cents`. Frontend expects `price` (number, dollars), `image`, `available`, `totalPrice`. Transformers in `lib/api/transformers.ts` handle all conversions.

## Architecture Docs (C3)

This project uses C3 docs in `.c3/`.
For architecture questions, changes, audits, file context → `/c3`.
Operations: query, audit, change, ref, sweep.
File lookup: `c3x lookup <file-or-glob>` maps files/directories to components + refs.
Code-map coverage: 100% of source files mapped to components.
