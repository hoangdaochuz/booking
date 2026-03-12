export interface TicketTier {
  id: string;
  name: string;
  description: string;
  price: number;
  available: number;
}

export interface VenueLayout {
  id: string;
  name: string;
  sections: VenueSection[];
  stage?: {
    label: string;
    shape: "rectangle" | "circle" | "polygon";
    position: { x: number; y: number; width: number; height: number };
  };
}

export interface VenueSection {
  id: string;
  name: string;
  tier: string;
  price: number;
  color: string;
  path: string;
  labelPosition: { x: number; y: number };
  totalSeats: number;
  availableSeats: number;
  rows: SeatRow[];
}

export interface SeatRow {
  id: string;
  label: string;
  seats: Seat[];
}

export interface Seat {
  id: string;
  label: string;
  status: "available" | "taken" | "selected";
}

export interface SelectedSeat {
  sectionId: string;
  sectionName: string;
  rowLabel: string;
  seatLabel: string;
  seatId: string;
  price: number;
  tierId: string;
}

export interface Event {
  id: string;
  title: string;
  date: string;
  time: string;
  venue: string;
  location: string;
  category: "Concerts" | "Sports" | "Films";
  genre: string;
  description: string;
  image: string;
  featured?: boolean;
  tiers: TicketTier[];
  venueLayout?: VenueLayout;
}

export interface CartItem {
  eventId: string;
  seats: SelectedSeat[];
}

export interface PurchasedTicket {
  id: string;
  eventId: string;
  tierName: string;
  quantity: number;
  totalPrice: number;
  purchasedAt: string;
  status: "Confirmed" | "Upcoming";
  seats?: { section: string; row: string; seat: string }[];
}
