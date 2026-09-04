CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS scheduler_configs(
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL UNIQUE,
    timeout INT DEFAULT 0,
    version INT DEFAULT 1,
    interval_expression VARCHAR(50) NOT NULL,
    is_enable BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
)

CREATE TYPE outbound_status AS ENUM('pending', 'published');

CREATE TABLE IF NOT EXISTS outbound_events(
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    topic VARCHAR(100) NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    status outbound_status,
    published_at TIMESTAMPTZ,
    payload JSONB
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
)