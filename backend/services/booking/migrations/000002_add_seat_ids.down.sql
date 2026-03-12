-- Remove seat_ids column from booking_items
ALTER TABLE booking_items DROP COLUMN IF EXISTS seat_ids;
