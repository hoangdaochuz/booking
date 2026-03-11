# TicketBox Backend Design

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build a Go microservice backend for TicketBox — a ticket booking system that demonstrates the double booking problem under high concurrency (1000-10000 simultaneous requests).

**Architecture:** Pragmatic CQRS — PostgreSQL as source of truth, Kafka for async event-driven communication and read model projection. gRPC for inter-service communication, REST API gateway for the frontend.

**Tech Stack:**
- Go 1.22+
- PostgreSQL (database-per-service)
- Apache Kafka (event bus)
- Redis (caching, JWT blacklist)
- gRPC (inter-service), REST/JSON (frontend API)
- Docker Compose (local orchestration)

---

## Service Architecture

5 microservices + 1 API gateway:

```
Frontend (Next.js)
    │ REST
API Gateway (:8000)
    │ gRPC
    ├── User Service (:50051)      → postgres-user
    ├── Event Service (:50052)     → postgres-event
    ├── Booking Service (:50053)   → postgres-booking
    └── Notification Service       → postgres-notif (Kafka consumer only)

Kafka (event bus) ← all services publish/consume
Redis ← gateway + booking + event services
```

**Kafka Topics:**
- `booking.events` — BookingCreated, BookingConfirmed, BookingFailed, BookingCancelled
- `notification.events` — SendEmail, SendSMS
- `user.events` — UserRegistered, UserUpdated
- `event.events` — EventCreated, TicketAvailabilityUpdated

---

## Booking Service (Core)

### Booking Flow (Command Side)

1. User selects tickets → POST /api/bookings
2. API Gateway validates JWT → forwards via gRPC to Booking Service
3. Booking Service:
   a. Checks ticket availability (Redis cache → fallback to DB)
   b. Creates booking record with status=PENDING
   c. Acquires lock based on configured mode
   d. Decrements available count
   e. Updates booking status=CONFIRMED
   f. Publishes BookingConfirmed to Kafka
4. Notification Service consumes event → sends confirmation email

### Three Booking Modes (toggled via config/header)

| Mode | How it works | Result under load |
|------|-------------|-------------------|
| `naive` | No locking, direct UPDATE | Double bookings — overselling |
| `pessimistic` | `SELECT FOR UPDATE` row lock | Safe but slower — requests queue |
| `optimistic` | Version column + `WHERE version = ?` retry | Safe, better throughput, retries |

### Database Schema (booking DB)

```sql
bookings: id, user_id, event_id, status, total_amount, version, created_at
booking_items: id, booking_id, ticket_tier_id, quantity, unit_price
ticket_locks: id, tier_id, lock_mode, locked_until
```

### Read Side (CQRS Query Path)

- Kafka consumer listens to `booking.events`
- Projects into denormalized `booking_read_model` table
- API Gateway reads from this model for GET endpoints

---

## User Service & Authentication

### Auth Flow (JWT-based, self-managed)

- **Register:** POST /api/auth/register → bcrypt hash → store → publish UserRegistered → return JWT
- **Login:** POST /api/auth/login → verify → return access + refresh tokens
- **Refresh:** POST /api/auth/refresh → validate refresh → return new access token
- **Logout:** POST /api/auth/logout → add token to Redis blacklist

### Token Strategy

- Access token: 15 min TTL, contains user_id, email, role
- Refresh token: 7 days TTL, stored in DB, rotated on use
- Redis blacklist: checked by API Gateway on every request

### Database Schema (user DB)

```sql
users: id, email, password_hash, name, role, created_at, updated_at
refresh_tokens: id, user_id, token_hash, expires_at, revoked_at
```

### Roles

- `user` (default) — can browse events and book tickets
- `admin` — can manage events

---

## Event Service

### Database Schema (event DB)

```sql
events: id, title, description, category, venue, location, date, image_url, status, created_at, updated_at
ticket_tiers: id, event_id, name, price, total_quantity, available_quantity, version, created_at
```

### Key Operations

- CRUD for events (admin only)
- Query events with filtering (category, date, search)
- Ticket availability cached in Redis with 5s TTL for burst reads
- Publishes TicketAvailabilityUpdated to Kafka on changes
- Kafka consumer projects into denormalized read table

---

## Notification Service

Purely event-driven — no REST/gRPC API.

### Kafka Consumers

- `booking.events` → BookingConfirmed → confirmation email
- `booking.events` → BookingFailed → failure email
- `booking.events` → BookingCancelled → cancellation email
- `user.events` → UserRegistered → welcome email

### Implementation

- MVP: log to stdout (simulated delivery)
- Pluggable interface for real providers (SMTP, SendGrid, Twilio)
- Outbox pattern to prevent duplicate sends on reprocessing

### Database Schema (notification DB)

```sql
notifications: id, type, recipient, channel, payload, status, sent_at, created_at
```

---

## Project Structure

```
backend/
├── docker-compose.yml
├── Makefile
├── proto/                          # Shared protobuf definitions
│   ├── user/v1/user.proto
│   ├── event/v1/event.proto
│   └── booking/v1/booking.proto
├── pkg/                            # Shared libraries
│   ├── kafka/                      # Producer/consumer helpers
│   ├── middleware/                  # Common middleware
│   ├── database/                   # DB connection, migration helpers
│   └── config/                     # Config loading (env/yaml)
├── services/
│   ├── gateway/                    # API Gateway
│   │   ├── cmd/main.go
│   │   ├── internal/
│   │   │   ├── handler/            # REST handlers
│   │   │   ├── middleware/         # Auth, rate limiting
│   │   │   └── router/
│   │   └── Dockerfile
│   ├── user/                       # User Service
│   │   ├── cmd/main.go
│   │   ├── internal/
│   │   │   ├── domain/             # Entities, value objects
│   │   │   ├── repository/         # DB access
│   │   │   ├── service/            # Business logic
│   │   │   ├── grpc/               # gRPC server handlers
│   │   │   └── kafka/              # Event producers/consumers
│   │   ├── migrations/
│   │   └── Dockerfile
│   ├── event/                      # Same structure as user
│   ├── booking/                    # Same structure as user
│   └── notification/               # Same structure (no grpc/)
└── scripts/
    ├── migrate.sh
    └── load-test.sh                # k6/vegeta for double booking demo
```

## Key Dependencies

| Purpose | Library |
|---------|---------|
| HTTP router | `github.com/gin-gonic/gin` |
| gRPC | `google.golang.org/grpc` |
| PostgreSQL | `github.com/jackc/pgx/v5` |
| Kafka | `github.com/segmentio/kafka-go` |
| Redis | `github.com/redis/go-redis/v9` |
| Config | `github.com/spf13/viper` |
| Logging | `go.uber.org/zap` |
| Migrations | `github.com/golang-migrate/migrate/v4` |
| JWT | `github.com/golang-jwt/jwt/v5` |
| Protobuf | `protoc` + `protoc-gen-go-grpc` |
| Load testing | `k6` or `github.com/tsenart/vegeta` |

## Design Patterns

- **Clean Architecture** — domain → service → repository layers
- **Repository pattern** — interface-based DB access
- **CQRS** — separate command/query paths with Kafka projections
- **Outbox pattern** — reliable event publishing
- **Circuit breaker** — inter-service gRPC resilience

---

## Docker Compose Infrastructure

```
# Databases (one per service)
postgres-user:     :5433
postgres-event:    :5434
postgres-booking:  :5435
postgres-notif:    :5436

# Infrastructure
redis:             :6379
zookeeper:         :2181
kafka:             :9092
kafka-ui:          :8080 (optional)

# Application
gateway:           :8000 (REST)
user-service:      :50051 (gRPC)
event-service:     :50052 (gRPC)
booking-service:   :50053 (gRPC)
notification-service: (no port, Kafka consumer)
```

**Makefile targets:** proto, build, up, down, migrate, test, test-load, logs

---

## API Endpoints (Gateway)

### Auth
- POST /api/auth/register
- POST /api/auth/login
- POST /api/auth/refresh
- POST /api/auth/logout

### Events
- GET /api/events (list, filter by category/date/search)
- GET /api/events/:id
- POST /api/events (admin)
- PUT /api/events/:id (admin)
- DELETE /api/events/:id (admin)

### Bookings
- POST /api/bookings (create booking)
- GET /api/bookings (user's bookings — from read model)
- GET /api/bookings/:id
- POST /api/bookings/:id/cancel

### Users
- GET /api/users/me (profile)
- PUT /api/users/me (update profile)
