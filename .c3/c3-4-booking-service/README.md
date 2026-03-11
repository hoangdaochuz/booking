---
id: c3-4
c3-version: 4
title: booking-service
type: container
boundary: service
parent: c3-0
goal: Create and manage bookings transactionally with concurrency control
summary: gRPC service orchestrating cross-service booking transactions with outbox pattern
---

# booking-service

## Goal

Create and manage bookings transactionally with concurrency control

## Responsibilities

- Transactional booking creation with atomicity guarantees
- Cross-service availability checks via event service gRPC calls
- Outbox-based event publishing to Kafka for downstream consumers
- Booking lifecycle management (get/list/cancel)

## Complexity Assessment

**Level:** complex
**Why:** Cross-service transaction coordination, outbox pattern, configurable locking modes

## Components

| ID | Name | Category | Status | Goal Contribution |
|----|------|----------|--------|-------------------|
| c3-401 | grpc-server | foundation | active | Implements BookingService protobuf interface |
| c3-402 | booking-repository | foundation | active | Persists bookings and outbox entries in PostgreSQL |
| c3-410 | booking-creation | feature | active | Orchestrates booking creation with event-service availability check |
| c3-411 | booking-management | feature | active | Handles get/list/cancel booking operations |

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
