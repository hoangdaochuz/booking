// Raw API response types (match backend JSON exactly)

export interface ApiUser {
  id: string;
  email: string;
  name: string;
  role: string;
  created_at: string;
}

export interface ApiAuthResponse {
  access_token: string;
  refresh_token: string;
  user: ApiUser;
}

export interface ApiTicketTier {
  id: string;
  event_id: string;
  name: string;
  price_cents: number;
  total_quantity: number;
  available_quantity: number;
  version: number;
}

export interface ApiEvent {
  id: string;
  title: string;
  description: string;
  category: string;
  venue: string;
  location: string;
  date: string; // RFC3339
  image_url: string;
  status: string;
  tiers: ApiTicketTier[];
  created_at: string;
}

export interface ApiListEventsResponse {
  events: ApiEvent[];
  total: number;
  page: number;
  page_size: number;
}

export interface ApiBookingItem {
  id: string;
  ticket_tier_id: string;
  tier_name: string;
  quantity: number;
  unit_price_cents: number;
  seat_ids: string[];
}

export interface ApiBooking {
  id: string;
  user_id: string;
  event_id: string;
  status: string;
  total_amount_cents: number;
  items: ApiBookingItem[];
  created_at: string;
  paymentIntentClientSecret?: string;
}

export interface ApiListBookingsResponse {
  bookings: ApiBooking[];
  total: number;
}

export interface ApiError {
  error: string;
}

// ── Seats ────────────────────────────────────────────────
export interface ApiSeatPosition {
  sectionId: string;
  row: string;
  seat: number;
  x: number;
  y: number;
}

export interface ApiSeat {
  id: string;
  event_id: string;
  ticket_tier_id: string;
  status: 'available' | 'reserved' | 'booked';
  booking_id: string | null;
  order_id: string;
  position: ApiSeatPosition;
  created_at: string;
  updated_at: string;
}

export interface ApiGetSeatsParams {
  event_id: string;
  ticket_tier_id?: string;
}

export interface ApiUpdateSeatStatusParams {
  seat_id: string;
  status: 'available' | 'reserved' | 'booked';
  booking_id?: string;
}
