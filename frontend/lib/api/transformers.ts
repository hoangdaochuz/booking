import { ApiEvent, ApiTicketTier, ApiBooking } from "./types";
import { Event, TicketTier, PurchasedTicket } from "../types";
import { generateVenueLayout } from "./generate-venue-layout";

function formatDate(isoDate: string): string {
  const d = new Date(isoDate);
  return d.toLocaleDateString("en-US", { month: "short", day: "numeric", year: "numeric" });
}

function formatTime(isoDate: string): string {
  const d = new Date(isoDate);
  return d.toLocaleTimeString("en-US", { hour: "numeric", minute: "2-digit", hour12: true });
}

function mapCategory(category: string): Event["category"] {
  const map: Record<string, Event["category"]> = {
    concerts: "Concerts",
    sports: "Sports",
    films: "Films",
  };
  return map[category.toLowerCase()] || "Concerts";
}

export function transformTier(tier: ApiTicketTier): TicketTier {
  return {
    id: tier.id,
    name: tier.name,
    description: `${tier.available_quantity} of ${tier.total_quantity} remaining`,
    price: tier.price_cents / 100,
    available: tier.available_quantity,
  };
}

export function transformEvent(apiEvent: ApiEvent): Event {
  return {
    id: apiEvent.id,
    title: apiEvent.title,
    date: formatDate(apiEvent.date),
    time: formatTime(apiEvent.date),
    venue: apiEvent.venue,
    location: apiEvent.location,
    category: mapCategory(apiEvent.category),
    genre: apiEvent.category,
    description: apiEvent.description,
    image: apiEvent.image_url,
    tiers: (apiEvent.tiers || []).map(transformTier),
    venueLayout: generateVenueLayout(apiEvent.tiers || [], apiEvent.category),
  };
}

export function transformBookingToTicket(booking: ApiBooking, event?: Event): PurchasedTicket {
  const tierName = booking.items?.[0]?.tier_name || "General";
  const totalQuantity = booking.items?.reduce((sum, item) => sum + item.quantity, 0) || 0;

  return {
    id: booking.id,
    eventId: booking.event_id,
    tierName,
    quantity: totalQuantity,
    totalPrice: booking.total_amount_cents / 100,
    purchasedAt: booking.created_at,
    status: booking.status === "confirmed" ? "Confirmed" : "Upcoming",
  };
}
