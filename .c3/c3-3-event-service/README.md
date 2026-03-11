---
id: c3-3
c3-version: 4
title: event-service
type: container
boundary: service
parent: c3-0
goal: Manage events and enforce ticket availability with atomic updates
summary: gRPC service providing event CRUD and row-level-locked availability decrements
---

# event-service

## Goal

Manage events and enforce ticket availability with atomic updates

## Responsibilities

- Event lifecycle CRUD operations
- Ticket tier management
- Atomic availability updates with SELECT FOR UPDATE row-level locking
- Version-based optimistic locking as alternative concurrency mode

## Complexity Assessment

**Level:** complex
**Why:** Row-level locking critical path, concurrent availability management, optimistic versioning

## Components

| ID | Name | Category | Status | Goal Contribution |
|----|------|----------|--------|-------------------|
| c3-301 | grpc-server | foundation | active | Implements EventService protobuf interface |
| c3-302 | event-repository | foundation | active | Persists events/tiers with SELECT FOR UPDATE locking |
| c3-310 | event-crud | feature | active | Manages event and tier lifecycle |
| c3-311 | ticket-availability | feature | active | Enforces availability invariant under concurrency |

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
