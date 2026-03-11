---
id: c3-1
c3-version: 4
title: gateway
type: container
boundary: service
parent: c3-0
goal: Route external HTTP requests to internal gRPC services with authentication enforcement
summary: Gin HTTP gateway translating REST to gRPC with JWT validation and Redis token blacklist
---

# gateway

## Goal

Route external HTTP requests to internal gRPC services with authentication enforcement

## Responsibilities

- HTTP→gRPC translation for all backend service endpoints
- JWT authentication enforcement on protected routes
- CORS policy management for frontend access
- Request/response mapping between HTTP and gRPC formats

## Complexity Assessment

**Level:** moderate
**Why:** Multiple middleware layers, gRPC client management, error translation

## Components

| ID | Name | Category | Status | Goal Contribution |
|----|------|----------|--------|-------------------|
| c3-101 | http-router | foundation | active | Defines route structure and middleware chain for all API endpoints |
| c3-102 | auth-middleware | foundation | active | Enforces JWT auth and Redis blacklist on protected routes |
| c3-110 | auth-handler | feature | active | Exposes register/login/refresh/logout via HTTP |
| c3-111 | event-handler | feature | active | Exposes event CRUD and availability via HTTP |
| c3-112 | booking-handler | feature | active | Exposes booking create/get/list/cancel via HTTP |
| c3-113 | user-handler | feature | active | Exposes profile get/update via HTTP |

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
