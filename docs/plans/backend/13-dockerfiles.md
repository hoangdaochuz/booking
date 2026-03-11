# Task 13: Dockerfiles for All Services

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Create multi-stage Dockerfiles for each service and add application containers to docker-compose.yml.

**Files:**
- Create: `backend/services/gateway/Dockerfile`
- Create: `backend/services/user/Dockerfile`
- Create: `backend/services/event/Dockerfile`
- Create: `backend/services/booking/Dockerfile`
- Create: `backend/services/notification/Dockerfile`
- Modify: `backend/docker-compose.yml` (add application services)

---

### Step 1: Create shared Dockerfile pattern

All services use the same multi-stage build. Example for user service:

`backend/services/user/Dockerfile`:
```dockerfile
FROM golang:1.22-alpine AS builder
WORKDIR /app

COPY pkg/ ./pkg/
COPY services/user/ ./services/user/
COPY go.work go.work

RUN cd services/user && go mod download && \
    CGO_ENABLED=0 GOOS=linux go build -o /bin/user-service ./cmd

FROM alpine:3.19
RUN apk --no-cache add ca-certificates
COPY --from=builder /bin/user-service /bin/user-service
COPY services/user/migrations /migrations

ENTRYPOINT ["/bin/user-service"]
```

Create the same pattern for:
- `gateway` → `/bin/gateway` (no migrations)
- `event` → `/bin/event-service`
- `booking` → `/bin/booking-service`
- `notification` → `/bin/notification-service`

### Step 2: Add application services to docker-compose.yml

Append to `backend/docker-compose.yml` services section:

```yaml
  user-service:
    build:
      context: .
      dockerfile: services/user/Dockerfile
    environment:
      SERVICE_NAME: user-service
      GRPC_PORT: "50051"
      DATABASE_URL: postgres://ticketbox:${POSTGRES_PASSWORD:-ticketbox_secret}@postgres-user:5432/ticketbox_user?sslmode=disable
      REDIS_URL: redis://redis:6379
      KAFKA_BROKERS: kafka:29092
      JWT_SECRET: ${JWT_SECRET:-your-jwt-secret-change-in-production}
    depends_on:
      postgres-user:
        condition: service_healthy
      kafka:
        condition: service_healthy
      redis:
        condition: service_healthy
    networks:
      - ticketbox-net

  event-service:
    build:
      context: .
      dockerfile: services/event/Dockerfile
    environment:
      SERVICE_NAME: event-service
      GRPC_PORT: "50052"
      DATABASE_URL: postgres://ticketbox:${POSTGRES_PASSWORD:-ticketbox_secret}@postgres-event:5432/ticketbox_event?sslmode=disable
      REDIS_URL: redis://redis:6379
      KAFKA_BROKERS: kafka:29092
    depends_on:
      postgres-event:
        condition: service_healthy
      kafka:
        condition: service_healthy
      redis:
        condition: service_healthy
    networks:
      - ticketbox-net

  booking-service:
    build:
      context: .
      dockerfile: services/booking/Dockerfile
    environment:
      SERVICE_NAME: booking-service
      GRPC_PORT: "50053"
      DATABASE_URL: postgres://ticketbox:${POSTGRES_PASSWORD:-ticketbox_secret}@postgres-booking:5432/ticketbox_booking?sslmode=disable
      KAFKA_BROKERS: kafka:29092
      EVENT_SERVICE_ADDR: event-service:50052
      BOOKING_MODE: ${BOOKING_MODE:-pessimistic}
    depends_on:
      postgres-booking:
        condition: service_healthy
      kafka:
        condition: service_healthy
      event-service:
        condition: service_started
    networks:
      - ticketbox-net

  notification-service:
    build:
      context: .
      dockerfile: services/notification/Dockerfile
    environment:
      SERVICE_NAME: notification-service
      DATABASE_URL: postgres://ticketbox:${POSTGRES_PASSWORD:-ticketbox_secret}@postgres-notification:5432/ticketbox_notification?sslmode=disable
      KAFKA_BROKERS: kafka:29092
    depends_on:
      postgres-notification:
        condition: service_healthy
      kafka:
        condition: service_healthy
    networks:
      - ticketbox-net

  gateway:
    build:
      context: .
      dockerfile: services/gateway/Dockerfile
    ports:
      - "8000:8000"
    environment:
      SERVICE_NAME: gateway
      HTTP_PORT: "8000"
      USER_SERVICE_ADDR: user-service:50051
      EVENT_SERVICE_ADDR: event-service:50052
      BOOKING_SERVICE_ADDR: booking-service:50053
      REDIS_URL: redis://redis:6379
      JWT_SECRET: ${JWT_SECRET:-your-jwt-secret-change-in-production}
    depends_on:
      - user-service
      - event-service
      - booking-service
      - redis
    networks:
      - ticketbox-net
```

### Step 3: Build and test

```bash
cd /Users/dev/work/booking/backend
docker-compose build
docker-compose up -d
docker-compose ps
docker-compose logs -f gateway
```
Expected: All services start and connect.

### Step 4: Commit

```bash
git add backend/services/*/Dockerfile backend/docker-compose.yml
git commit -m "feat(backend): add Dockerfiles and compose app services"
```
