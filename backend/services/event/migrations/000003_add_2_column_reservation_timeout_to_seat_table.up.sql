ALTER TABLE seats
ADD COLUMN reservation_expired_at TIMESTAMPTZ DEFAULT NULL,
ADD COLUMN reserved_by_booking_id UUID NULL;