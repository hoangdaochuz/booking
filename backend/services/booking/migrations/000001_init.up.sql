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
