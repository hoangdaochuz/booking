import { ApiEvent, ApiTicketTier, ApiBooking, ApiSeat, ApiSeatPosition } from "./types";
import { Event, TicketTier, PurchasedTicket, VenueLayout, SeatRow, VenueSection } from "../types";
import { generateVenueLayout } from "./generate-venue-layout";
import { VenueSection as ApiVenueSection } from "../types";

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

// Frontend-specific seat type with enhanced properties
export interface FrontendSeat {
  id: string;
  eventId: string;
  tierId: string;
  status: "available" | "reserved" | "booked";
  bookingId: string | null;
  orderId: string;
  position: {
    sectionId: string;
    row: string;
    seatNumber: number;
    x: number;
    y: number;
  };
  createdAt: string;
  updatedAt: string;
}

function transformSeatPosition(position: ApiSeatPosition) {
  return {
    sectionId: position.sectionId,
    row: position.row,
    seatNumber: position.seat,
    x: position.x,
    y: position.y,
  };
}

function mapSeatStatus(status: ApiSeat["status"]): FrontendSeat["status"] {
  return status;
}

function formatDateTime(isoDate: string): string {
  const d = new Date(isoDate);
  return d.toLocaleString("en-US", {
    month: "short",
    day: "numeric",
    year: "numeric",
    hour: "numeric",
    minute: "2-digit",
    hour12: true,
  });
}

export function transformSeat(apiSeat: ApiSeat): FrontendSeat {
  return {
    id: apiSeat.id,
    eventId: apiSeat.event_id,
    tierId: apiSeat.ticket_tier_id,
    status: mapSeatStatus(apiSeat.status),
    bookingId: apiSeat.booking_id,
    orderId: apiSeat.order_id,
    position: transformSeatPosition(apiSeat.position),
    createdAt: formatDateTime(apiSeat.created_at),
    updatedAt: formatDateTime(apiSeat.updated_at),
  };
}

export function transformSeats(apiSeats: ApiSeat[]): FrontendSeat[] {
  return apiSeats.map(transformSeat);
}

// Generate SVG path from seat coordinates
function generatePathFromSeats(seats: ApiSeat[]): string {
  if (seats.length === 0) return "";

  // Add padding around the section
  const padding = 15;

  // Find bounding box of all seats
  let minX = Infinity, minY = Infinity, maxX = -Infinity, maxY = -Infinity;
  for (const seat of seats) {
    minX = Math.min(minX, seat.position.x);
    minY = Math.min(minY, seat.position.y);
    maxX = Math.max(maxX, seat.position.x);
    maxY = Math.max(maxY, seat.position.y);
  }

  // Create a rounded rectangle path with padding
  const x = minX - padding;
  const y = minY - padding;
  const width = maxX - minX + padding * 2;
  const height = maxY - minY + padding * 2;
  const radius = Math.min(width, height) * 0.1;

  return `M ${x + radius},${y}
          L ${x + width - radius},${y}
          Q ${x + width},${y} ${x + width},${y + radius}
          L ${x + width},${y + height - radius}
          Q ${x + width},${y + height} ${x + width - radius},${y + height}
          L ${x + radius},${y + height}
          Q ${x},${y + height} ${x},${y + height - radius}
          L ${x},${y + radius}
          Q ${x},${y} ${x + radius},${y}
          Z`.replace(/\s+/g, " ");
}

// Generate venue layout from actual seat data
export function generateVenueLayoutFromSeats(
  apiSeats: ApiSeat[],
  tiers: ApiTicketTier[],
  category: string
): VenueLayout {
  // Sort tiers by price (most expensive first) to assign colors correctly
  const sortedTiers = [...tiers].sort((a, b) => b.price_cents - a.price_cents);
  const tierMap = new Map(tiers.map(t => [t.id, t]));

  // Group seats by sectionId (format: "TierName-Template" like "Floor-center", "VIP-left")
  const sectionSeats = new Map<string, ApiSeat[]>();
  for (const seat of apiSeats) {
    const sectionId = seat.position.sectionId;
    if (!sectionSeats.has(sectionId)) {
      sectionSeats.set(sectionId, []);
    }
    sectionSeats.get(sectionId)!.push(seat);
  }

  // Get the base layout for SVG templates and stage info
  const baseLayout = generateVenueLayout(tiers, category);

  // Create a map of template names to their SVG paths and label positions
  // Extract from the base layout (section IDs are like "tierId:template", using : as separator)
  const templateMap = new Map<string, { path: string; labelPosition: { x: number; y: number } }>();
  for (const section of baseLayout.sections) {
    // Extract template part from section ID (e.g., "center" from "uuid:center")
    const parts = section.id.split(":");
    const template = parts.length > 1 ? parts[parts.length - 1] : section.id;
    templateMap.set(template, {
      path: section.path,
      labelPosition: section.labelPosition,
    });
  }

  // Build sections from seat data
  const sections: VenueSection[] = [];
  const tierColors = new Map<string, string>();

  let colorIndex = 0;
  const colors = ["#E8567F", "#7C5CFC", "#38A3A5", "#F59E0B", "#6366F1"];

  console.log("[transformer] Template map keys:", Array.from(templateMap.keys()));
  console.log("[transformer] Section seats map keys:", Array.from(sectionSeats.keys()));

  for (const [sectionId, seats] of sectionSeats) {
    if (seats.length === 0) continue;

    const firstSeat = seats[0];
    const tierId = firstSeat.ticket_tier_id;
    const tier = tierMap.get(tierId);

    // Extract tier name and template from sectionId (format: "TierName-Template")
    // Note: Templates can be compound like "lower-east", so we need to rejoin
    const parts = sectionId.split("-");
    const tierName = tier?.name || parts[0];

    // For templates like "lower-east", we need to extract after the tier name
    // If sectionId is "Category 1-lower-east", template should be "lower-east"
    let template = sectionId;
    if (parts.length > 1 && tierName) {
      // Remove tier name prefix (with hyphen) to get template
      const tierPrefix = `${tierName}-`;
      if (sectionId.startsWith(tierPrefix)) {
        template = sectionId.substring(tierPrefix.length);
      } else {
        // Fallback: use last part if we can't match tier name
        template = parts[parts.length - 1];
      }
    }

    const price = tier ? tier.price_cents / 100 : 50;

    // Find the tier's rank among sorted tiers
    const tierRank = sortedTiers.findIndex(t => t.id === tierId);

    // Assign color based on tier rank
    if (!tierColors.has(tierName)) {
      tierColors.set(tierName, colors[tierRank >= 0 ? tierRank % colors.length : colorIndex % colors.length]);
      if (tierRank < 0) colorIndex++;
    }

    // Get SVG template info
    const hasTemplate = templateMap.has(template);
    if (!hasTemplate) {
      console.warn(`[transformer] Template "${template}" not found in map for section "${sectionId}"`);
    }
    const templateInfo = templateMap.get(template) || {
      path: "",
      labelPosition: { x: 400, y: 300 },
    };

    // Group seats by row
    const rowMap = new Map<string, ApiSeat[]>();
    for (const seat of seats) {
      const row = seat.position.row;
      if (!rowMap.has(row)) {
        rowMap.set(row, []);
      }
      rowMap.get(row)!.push(seat);
    }

    // Create seat rows
    const seatRows: SeatRow[] = [];
    for (const [row, rowSeats] of rowMap) {
      // Sort by seat number
      rowSeats.sort((a, b) => a.position.seat - b.position.seat);

      seatRows.push({
        id: `${sectionId}-${row}`,
        label: `Row ${row}`,
        seats: rowSeats.map(seat => ({
          id: seat.id,
          label: `${seat.position.seat}`,
          status: seat.status as "available" | "reserved" | "booked",
        })),
      });
    }

    // Sort rows alphabetically
    seatRows.sort((a, b) => a.label.localeCompare(b.label));

    // Reserved seats are not yet confirmed, so they still count as available
    const availableCount = seats.filter(s => s.status === "available" || s.status === "reserved").length;

    sections.push({
      id: sectionId,
      name: tierName,
      tier: tierName,
      price,
      color: tierColors.get(tierName)!,
      path: templateInfo.path,
      labelPosition: templateInfo.labelPosition,
      totalSeats: seats.length,
      availableSeats: availableCount,
      rows: seatRows,
    });
  }

  // If we have actual seat data, use it; otherwise fall back to generated layout
  console.log(`[transformer] Generated ${sections.length} sections from ${apiSeats.length} seats`);
  if (sections.length > 0) {
    console.log("[transformer] Using real seat data");
    return {
      ...baseLayout,
      sections,
    };
  }

  console.warn("[transformer] No sections generated, falling back to base layout (mock data)");
  return baseLayout;
}

// Async function to fetch seats and generate layout
export async function fetchEventWithSeats(
  apiEvent: ApiEvent,
  getSeatsFn: (eventId: string) => Promise<{ seats: ApiSeat[] }>
): Promise<Event> {
  const baseEvent: Event = {
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
    venueLayout: undefined,
  };

  try {
    // Fetch actual seats from API
    const seatsData = await getSeatsFn(apiEvent.id);
    const venueLayout = generateVenueLayoutFromSeats(
      seatsData.seats,
      apiEvent.tiers || [],
      apiEvent.category
    );
    return { ...baseEvent, venueLayout };
  } catch (error) {
    // Fallback to generated layout if seat fetch fails
    console.warn("Failed to fetch seats, using generated layout:", error);
    return { ...baseEvent, venueLayout: generateVenueLayout(apiEvent.tiers || [], apiEvent.category) };
  }
}
