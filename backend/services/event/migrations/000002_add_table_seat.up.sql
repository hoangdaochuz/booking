-- Create seat status enum type
CREATE TYPE seat_status AS ENUM ('available', 'reserved', 'booked');

-- Create seats table
CREATE TABLE IF NOT EXISTS seats (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    ticket_tier_id UUID NOT NULL REFERENCES ticket_tiers(id) ON DELETE CASCADE,
    status seat_status NOT NULL DEFAULT 'available',
    booking_id UUID REFERENCES bookings(id) ON DELETE SET NULL,
    order_id UUID NOT NULL DEFAULT uuid_generate_v4(),
    position JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- Create indexes with partial index optimization for soft deletes
CREATE INDEX idx_seats_event_id ON seats(event_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_seats_tier_id ON seats(ticket_tier_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_seats_status ON seats(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_seats_booking_id ON seats(booking_id) WHERE booking_id IS NOT NULL;
