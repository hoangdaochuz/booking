---
id: ref-grpc-service-structure
c3-version: 4
title: grpc-service-structure
goal: Enforce consistent Go microservice layout across all backend services
scope: [c3-1, c3-2, c3-3, c3-4, c3-5]
---

# grpc-service-structure

## Goal

Enforce consistent Go microservice layout across all backend services.

## Choice

Standard directory layout — `cmd/main.go`, `internal/{domain,repository,service,grpc,kafka}`, `migrations/`, `Dockerfile`, `go.mod` with replace directive. Go workspace (`go.work`) links all modules.

## Why

Consistent structure reduces cognitive load when switching between services, enables shared tooling (Makefile targets), and ensures each service is independently deployable with its own migrations.

## How

| Guideline | Example |
|-----------|---------|
| Entry point at `cmd/main.go` | Bootstraps config, DB pool, gRPC server |
| Domain models in `internal/domain/` | `booking.go`, `event.go` — pure structs, no DB deps |
| Repository interface + impl in `internal/repository/` | `booking_repository.go` (interface) + `postgres_booking_repository.go` (PGX) |
| Business logic in `internal/service/` | `booking_service.go` — orchestrates domain + repo |
| gRPC handlers in `internal/grpc/` | `server.go` — implements protobuf service interface |
| Kafka integration in `internal/kafka/` | `producer.go` or `consumer.go` |
| Migrations at `migrations/` | golang-migrate format: `000001_init.up.sql` / `000001_init.down.sql` |
| Replace directive in go.mod | `replace github.com/ticketbox/pkg => ../../pkg` |

## Not This

| Alternative | Rejected Because |
|-------------|------------------|
| Monorepo single binary | Cannot scale services independently |
| Flat package structure | Mixes concerns, harder to test |
| External domain package | Domain is service-specific, not shared |

## Scope

**Applies to:**
- c3-2, c3-3, c3-4, c3-5 (all gRPC services: user, event, booking, notification)

**Does NOT apply to:**
- Gateway (c3-1) follows a variant with `internal/{handler,middleware,router}` instead of `internal/{domain,repository,service,grpc}`

## Override

To override this ref:
1. Document justification in an ADR under "Pattern Overrides"
2. Cite this ref and explain why the override is necessary
3. Specify the scope of the override (which components deviate)

## Cited By

- c3-201, c3-301, c3-401, c3-501, c3-101, c3-704, c3-705
