# Task 2: Docker Compose Infrastructure

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Set up Docker Compose with PostgreSQL (4 instances), Kafka, Zookeeper, Redis, and Kafka UI.

**Files:**
- Create: `backend/docker-compose.yml`
- Create: `backend/.env.example`

---

### Step 1: Create docker-compose.yml

`backend/docker-compose.yml`:
```yaml
version: "3.9"

services:
  # ─── Databases ───
  postgres-user:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: ticketbox_user
      POSTGRES_USER: ticketbox
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD:-ticketbox_secret}
    ports:
      - "5433:5432"
    volumes:
      - pgdata-user:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ticketbox -d ticketbox_user"]
      interval: 5s
      timeout: 5s
      retries: 5
    networks:
      - ticketbox-net

  postgres-event:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: ticketbox_event
      POSTGRES_USER: ticketbox
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD:-ticketbox_secret}
    ports:
      - "5434:5432"
    volumes:
      - pgdata-event:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ticketbox -d ticketbox_event"]
      interval: 5s
      timeout: 5s
      retries: 5
    networks:
      - ticketbox-net

  postgres-booking:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: ticketbox_booking
      POSTGRES_USER: ticketbox
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD:-ticketbox_secret}
    ports:
      - "5435:5432"
    volumes:
      - pgdata-booking:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ticketbox -d ticketbox_booking"]
      interval: 5s
      timeout: 5s
      retries: 5
    networks:
      - ticketbox-net

  postgres-notification:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: ticketbox_notification
      POSTGRES_USER: ticketbox
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD:-ticketbox_secret}
    ports:
      - "5436:5432"
    volumes:
      - pgdata-notification:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ticketbox -d ticketbox_notification"]
      interval: 5s
      timeout: 5s
      retries: 5
    networks:
      - ticketbox-net

  # ─── Redis ───
  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis-data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      timeout: 5s
      retries: 5
    networks:
      - ticketbox-net

  # ─── Kafka ───
  zookeeper:
    image: confluentinc/cp-zookeeper:7.6.0
    environment:
      ZOOKEEPER_CLIENT_PORT: 2181
      ZOOKEEPER_TICK_TIME: 2000
    ports:
      - "2181:2181"
    networks:
      - ticketbox-net

  kafka:
    image: confluentinc/cp-kafka:7.6.0
    depends_on:
      zookeeper:
        condition: service_started
    ports:
      - "9092:9092"
    environment:
      KAFKA_BROKER_ID: 1
      KAFKA_ZOOKEEPER_CONNECT: zookeeper:2181
      KAFKA_ADVERTISED_LISTENERS: PLAINTEXT://kafka:29092,PLAINTEXT_HOST://localhost:9092
      KAFKA_LISTENER_SECURITY_PROTOCOL_MAP: PLAINTEXT:PLAINTEXT,PLAINTEXT_HOST:PLAINTEXT
      KAFKA_INTER_BROKER_LISTENER_NAME: PLAINTEXT
      KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR: 1
      KAFKA_AUTO_CREATE_TOPICS_ENABLE: "true"
    healthcheck:
      test: ["CMD", "kafka-broker-api-versions", "--bootstrap-server", "localhost:9092"]
      interval: 10s
      timeout: 10s
      retries: 10
    networks:
      - ticketbox-net

  kafka-ui:
    image: provectuslabs/kafka-ui:latest
    depends_on:
      - kafka
    ports:
      - "8080:8080"
    environment:
      KAFKA_CLUSTERS_0_NAME: ticketbox
      KAFKA_CLUSTERS_0_BOOTSTRAPSERVERS: kafka:29092
    networks:
      - ticketbox-net

volumes:
  pgdata-user:
  pgdata-event:
  pgdata-booking:
  pgdata-notification:
  redis-data:

networks:
  ticketbox-net:
    driver: bridge
```

### Step 2: Create .env.example

`backend/.env.example`:
```env
POSTGRES_PASSWORD=ticketbox_secret
JWT_SECRET=your-jwt-secret-change-in-production
REDIS_URL=redis://localhost:6379
KAFKA_BROKERS=localhost:9092
```

### Step 3: Verify infrastructure starts

```bash
cd /Users/dev/work/booking/backend
docker-compose up -d
docker-compose ps
```
Expected: All containers healthy.

### Step 4: Verify connectivity

```bash
docker-compose exec postgres-user pg_isready -U ticketbox
docker-compose exec redis redis-cli ping
docker-compose exec kafka kafka-topics --bootstrap-server localhost:9092 --list
```

### Step 5: Tear down

```bash
docker-compose down
```

### Step 6: Commit

```bash
git add backend/docker-compose.yml backend/.env.example
git commit -m "feat(backend): add Docker Compose with Postgres, Kafka, Redis"
```
