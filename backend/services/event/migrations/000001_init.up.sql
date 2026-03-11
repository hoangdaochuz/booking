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
