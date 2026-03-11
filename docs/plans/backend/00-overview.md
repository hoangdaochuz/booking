# TicketBox Backend Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build a Go microservice backend for TicketBox — a ticket booking system demonstrating the double booking problem under high concurrency (1000-10000 simultaneous requests).

**Architecture:** Pragmatic CQRS with PostgreSQL as source of truth, Kafka for async event-driven communication and read model projection. gRPC for inter-service communication, REST API gateway for the frontend. Clean Architecture per service.

**Tech Stack:** Go 1.22+, PostgreSQL, Kafka, Redis, gRPC, Gin, Docker Compose

**Reference:** See `docs/plans/2026-03-10-ticketbox-backend-design.md` for full design.

---

## Task Files

### Phase 1: Foundation
- `01-scaffold.md` — Project scaffold & Go workspace
- `02-docker-compose.md` — Docker Compose infrastructure
- `03-shared-packages.md` — Config, database, Kafka shared libraries
- `04-protobuf.md` — Protobuf definitions
- `05-migrations.md` — Database migrations for all services

### Phase 2: User Service
- `06-user-domain-repo.md` — Domain entities & Postgres repository
- `07-user-auth-service.md` — Auth service with JWT & bcrypt
- `08-user-grpc-main.md` — gRPC server, Kafka producer, main.go

### Phase 3: Event & Notification Services
- `09-event-service.md` — Full event service with ticket availability
- `10-notification-service.md` — Kafka consumer notification service

### Phase 4: Booking Service (Core)
- `11-booking-service.md` — Booking service with 3 locking modes (naive/pessimistic/optimistic)

### Phase 5: API Gateway
- `12-gateway.md` — REST API gateway with auth middleware

### Phase 6: Deployment & Testing
- `13-dockerfiles.md` — Dockerfiles for all services
- `14-seed-data.md` — Seed script with mock events
- `15-load-test.md` — Load test for double booking demo
- `16-integration-test.md` — End-to-end smoke test

---

## Dependency Graph

```
Phase 1 (sequential): 01 → 02 → 03 → 04 → 05
Phase 2 (sequential): 06 → 07 → 08
Phase 3 (parallel):   09 + 10 (can run alongside Phase 2)
Phase 4:              11 (depends on 09)
Phase 5:              12 (depends on 08, 09, 10, 11)
Phase 6 (sequential): 13 → 14 → 15 → 16
```

## Summary

| # | Component | Description |
|---|-----------|-------------|
| 01 | Scaffold | Go workspace, modules, Makefile |
| 02 | Docker | Docker Compose with Postgres, Kafka, Redis |
| 03 | Shared | Config, database, Kafka shared packages |
| 04 | Proto | Protobuf definitions for all services |
| 05 | Migrations | Database schemas for all 4 services |
| 06 | User | Domain entities & Postgres repository |
| 07 | User | Auth service with JWT & bcrypt |
| 08 | User | gRPC server, Kafka producer, main.go |
| 09 | Event | Full event service with ticket availability |
| 10 | Notification | Kafka consumer notification service |
| 11 | Booking | Core booking service with 3 locking modes |
| 12 | Gateway | REST API gateway with auth middleware |
| 13 | Docker | Dockerfiles for all services |
| 14 | Data | Seed script with mock events |
| 15 | Testing | Load test for double booking demo |
| 16 | Integration | End-to-end smoke test |
