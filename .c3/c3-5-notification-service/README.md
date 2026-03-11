---
id: c3-5
c3-version: 4
title: notification-service
type: container
boundary: worker
parent: c3-0
goal: Consume booking events asynchronously and persist notifications
summary: Kafka consumer processing booking events into stored notifications
---

# notification-service

## Goal

Consume booking events asynchronously and persist notifications

## Responsibilities

- Kafka topic consumption for booking lifecycle events
- Notification persistence in dedicated PostgreSQL database
- Event deserialization and validation

## Complexity Assessment

**Level:** simple
**Why:** Straightforward consume-and-store pattern

## Components

| ID | Name | Category | Status | Goal Contribution |
|----|------|----------|--------|-------------------|
| c3-501 | kafka-consumer | foundation | active | Subscribes to Kafka booking topics |
| c3-502 | notification-repository | foundation | active | Persists notification records in PostgreSQL |
| c3-510 | notification-processing | feature | active | Transforms booking events into notification records |

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
