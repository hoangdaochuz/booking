---
id: c3-7
c3-version: 4
title: shared-backend
type: container
boundary: library
parent: c3-0
goal: Provide common infrastructure enabling consistent behavior across all backend services
summary: Go packages for config, database, Kafka, gRPC middleware, and generated protobuf code
---

# shared-backend

## Goal

Provide common infrastructure enabling consistent behavior across all backend services

## Responsibilities

- Environment-based configuration loading for all services
- PostgreSQL connection pooling and management
- Kafka producer/consumer setup utilities
- gRPC logging middleware for observability

## Complexity Assessment

**Level:** simple
**Why:** Utility packages with minimal business logic

## Components

| ID | Name | Category | Status | Goal Contribution |
|----|------|----------|--------|-------------------|
| c3-701 | config | foundation | active | Viper-based environment config loading |
| c3-702 | database | foundation | active | PGX connection pool creation and management |
| c3-703 | kafka-pkg | foundation | active | Kafka producer and consumer setup utilities |
| c3-704 | grpc-middleware | foundation | active | gRPC unary logging interceptor |
| c3-705 | proto-generated | foundation | active | Generated protobuf Go code for all service contracts |

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
