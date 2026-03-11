---
id: ref-repository-pattern
c3-version: 4
title: repository-pattern
goal: Separate data access from business logic with interface-based repositories
scope: [c3-2, c3-3, c3-4, c3-5]
---

# repository-pattern

## Goal

Separate data access from business logic with interface-based repositories.

## Choice

Go interface in `repository/` defining operations + PostgreSQL implementation using PGX in `postgres_*.go`. Services depend on the interface, not the implementation.

## Why

Enables testing with mocks, isolates PostgreSQL-specific code, and makes it possible to swap storage backends without changing business logic.

## How

| Guideline | Example |
|-----------|---------|
| Interface file: `{entity}_repository.go` | Defines CRUD operations as Go interface |
| Implementation: `postgres_{entity}_repository.go` | PGX-based, receives `*pgxpool.Pool` |
| Service receives interface | `NewBookingService(repo BookingRepository)` |
| Transactions via pool | `pool.BeginTx(ctx, pgx.TxOptions{})` for multi-step ops |

## Not This

| Alternative | Rejected Because |
|-------------|------------------|
| ORM (GORM, Ent) | Overhead for simple queries, less control over locking |
| Direct SQL in service | Mixes concerns, harder to test |
| Generic repository | Go generics add complexity without clear benefit here |

## Scope

**Applies to:**
- All services with databases (c3-2, c3-3, c3-4, c3-5)

**Does NOT apply to:**
- Gateway (c3-1) which has no database

## Override

To override this ref:
1. Document justification in an ADR under "Pattern Overrides"
2. Cite this ref and explain why the override is necessary
3. Specify the scope of the override (which components deviate)

## Cited By

- c3-202, c3-302, c3-310, c3-402, c3-411, c3-502, c3-211
