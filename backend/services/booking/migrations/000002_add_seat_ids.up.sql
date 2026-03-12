-- Add seat_ids column to booking_items
ALTER TABLE booking_items ADD COLUMN IF NOT EXISTS seat_ids UUID[] DEFAULT '{}';
