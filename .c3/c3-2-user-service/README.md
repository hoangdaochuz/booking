---
id: c3-2
c3-version: 4
title: user-service
type: container
boundary: service
parent: c3-0
goal: Manage user identity, authentication, and profile data
summary: gRPC service handling registration, login, JWT token lifecycle, and profile CRUD
---

# user-service

## Goal

Manage user identity, authentication, and profile data

## Responsibilities

- User registration and login with password hashing
- JWT access and refresh token generation and rotation
- Profile management (get/update)
- Kafka event publishing for user lifecycle events

## Complexity Assessment

**Level:** moderate
**Why:** JWT generation, password hashing, refresh token rotation

## Components

| ID | Name | Category | Status | Goal Contribution |
|----|------|----------|--------|-------------------|
| c3-201 | grpc-server | foundation | active | Implements UserService protobuf interface |
| c3-202 | user-repository | foundation | active | Persists user data with password hashing in PostgreSQL |
| c3-210 | auth-logic | feature | active | Handles register/login/refresh/logout business logic |
| c3-211 | profile-management | feature | active | Handles profile get/update operations |

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
