# Task 5: Database Migrations

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Create SQL migrations for all 4 service databases and a migration runner script.

**Files:**
- Create: `backend/services/user/migrations/000001_init.up.sql`
- Create: `backend/services/user/migrations/000001_init.down.sql`
- Create: `backend/services/event/migrations/000001_init.up.sql`
- Create: `backend/services/event/migrations/000001_init.down.sql`
- Create: `backend/services/booking/migrations/000001_init.up.sql`
- Create: `backend/services/booking/migrations/000001_init.down.sql`
- Create: `backend/services/notification/migrations/000001_init.up.sql`
- Create: `backend/services/notification/migrations/000001_init.down.sql`
- Create: `backend/scripts/migrate.sh`

---

### Step 1: User service migration

`backend/services/user/migrations/000001_init.up.sql`:
```sql
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'user',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);

CREATE TABLE refresh_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_token_hash ON refresh_tokens(token_hash);
```

`backend/services/user/migrations/000001_init.down.sql`:
```sql
DROP TABLE IF EXISTS refresh_tokens;
DROP TABLE IF EXISTS users;
```

### Step 2: Event service migration

`backend/services/event/migrations/000001_init.up.sql`:
```sql
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title VARCHAR(500) NOT NULL,
    description TEXT,
    category VARCHAR(100) NOT NULL,
    venue VARCHAR(500) NOT NULL,
    location VARCHAR(500) NOT NULL,
    date TIMESTAMPTZ NOT NULL,
    image_url TEXT,
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_events_category ON events(category);
CREATE INDEX idx_events_date ON events(date);
CREATE INDEX idx_events_status ON events(status);

CREATE TABLE ticket_tiers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    price_cents BIGINT NOT NULL,
    total_quantity INT NOT NULL,
    available_quantity INT NOT NULL,
    version INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ticket_tiers_event_id ON ticket_tiers(event_id);

CREATE TABLE events_read_model (
    id UUID PRIMARY KEY,
    title VARCHAR(500) NOT NULL,
    description TEXT,
    category VARCHAR(100) NOT NULL,
    venue VARCHAR(500) NOT NULL,
    location VARCHAR(500) NOT NULL,
    date TIMESTAMPTZ NOT NULL,
    image_url TEXT,
    status VARCHAR(50) NOT NULL,
    min_price_cents BIGINT,
    total_available INT,
    tiers_json JSONB,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

`backend/services/event/migrations/000001_init.down.sql`:
```sql
DROP TABLE IF EXISTS events_read_model;
DROP TABLE IF EXISTS ticket_tiers;
DROP TABLE IF EXISTS events;
```

### Step 3: Booking service migration

`backend/services/booking/migrations/000001_init.up.sql`:
```sql
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE bookings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL,
    event_id UUID NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'PENDING',
    total_amount_cents BIGINT NOT NULL DEFAULT 0,
    version INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_bookings_user_id ON bookings(user_id);
CREATE INDEX idx_bookings_event_id ON bookings(event_id);
CREATE INDEX idx_bookings_status ON bookings(status);

CREATE TABLE booking_items (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    booking_id UUID NOT NULL REFERENCES bookings(id) ON DELETE CASCADE,
    ticket_tier_id UUID NOT NULL,
    quantity INT NOT NULL,
    unit_price_cents BIGINT NOT NULL
);

CREATE INDEX idx_booking_items_booking_id ON booking_items(booking_id);

CREATE TABLE bookings_read_model (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    event_id UUID NOT NULL,
    event_title VARCHAR(500),
    event_date TIMESTAMPTZ,
    event_venue VARCHAR(500),
    status VARCHAR(50) NOT NULL,
    total_amount_cents BIGINT NOT NULL,
    items_json JSONB,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_bookings_read_user_id ON bookings_read_model(user_id);

CREATE TABLE outbox (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    event_type VARCHAR(255) NOT NULL,
    event_key VARCHAR(255) NOT NULL,
    payload JSONB NOT NULL,
    published BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_outbox_unpublished ON outbox(published) WHERE published = FALSE;
```

`backend/services/booking/migrations/000001_init.down.sql`:
```sql
DROP TABLE IF EXISTS outbox;
DROP TABLE IF EXISTS bookings_read_model;
DROP TABLE IF EXISTS booking_items;
DROP TABLE IF EXISTS bookings;
```

### Step 4: Notification service migration

`backend/services/notification/migrations/000001_init.up.sql`:
```sql
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE notifications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    type VARCHAR(100) NOT NULL,
    recipient VARCHAR(255) NOT NULL,
    channel VARCHAR(50) NOT NULL DEFAULT 'email',
    payload JSONB NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'PENDING',
    sent_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_notifications_status ON notifications(status);
CREATE INDEX idx_notifications_recipient ON notifications(recipient);
```

`backend/services/notification/migrations/000001_init.down.sql`:
```sql
DROP TABLE IF EXISTS notifications;
```

### Step 5: Create migration runner script

`backend/scripts/migrate.sh`:
```bash
#!/bin/bash
set -e

SERVICES=("user:5433:ticketbox_user" "event:5434:ticketbox_event" "booking:5435:ticketbox_booking" "notification:5436:ticketbox_notification")

for entry in "${SERVICES[@]}"; do
    IFS=: read -r svc port db <<< "$entry"
    echo "Running migrations for $svc service..."
    migrate -path "services/$svc/migrations" \
        -database "postgres://ticketbox:${POSTGRES_PASSWORD:-ticketbox_secret}@localhost:$port/$db?sslmode=disable" \
        up
    echo "$svc migrations complete."
done
```

### Step 6: Make executable and test

```bash
chmod +x /Users/dev/work/booking/backend/scripts/migrate.sh
cd /Users/dev/work/booking/backend
docker-compose up -d postgres-user postgres-event postgres-booking postgres-notification
sleep 5
bash scripts/migrate.sh
```
Expected: All migrations succeed.

### Step 7: Commit

```bash
git add backend/services/*/migrations/ backend/scripts/migrate.sh
git commit -m "feat(backend): add database migrations for all services"
```
