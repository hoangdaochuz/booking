-- Drop indexes
DROP INDEX IF EXISTS idx_seats_booking_id;
DROP INDEX IF EXISTS idx_seats_status;
DROP INDEX IF EXISTS idx_seats_tier_id;
DROP INDEX IF EXISTS idx_seats_event_id;

-- Drop seats table
DROP TABLE IF EXISTS seats;

-- Drop seat status enum type
DROP TYPE IF EXISTS seat_status;
