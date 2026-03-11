# Seat Selection Feature — Design & Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Replace the tier-based ticket selection with a visual, data-driven seat map. Users select a section on the venue map, then pick individual seats within that section. The layout is generic — different events can have completely different venue shapes (arena, octagon, theater, etc.) driven by provider data.

**Flow:** Event Detail → "Select Seats" → Seat Selection (`/events/[id]/seats`) → Checkout → My Tickets

**Design tokens (existing):**
- Primary orange: `#FF8400`
- Seat available: tier color (per section)
- Seat taken: `#D1D5DB` (gray-300)
- Seat selected: `#FF8400` (primary)
- Background: `#F2F3F0`
- Card bg: `#FFFFFF`

---

### Task 1: Add Venue & Seat Types

**File:** `frontend/lib/types.ts`

Add these types:

```typescript
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
  path: string;           // SVG path data
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
```

Update `CartItem` to support seat-based selection:

```typescript
export interface CartItem {
  eventId: string;
  seats: SelectedSeat[];
}

export interface SelectedSeat {
  sectionId: string;
  sectionName: string;
  rowLabel: string;
  seatLabel: string;
  seatId: string;
  price: number;
}
```

Update `Event` to include optional `venueLayout`:

```typescript
export interface Event {
  // ... existing fields
  tiers: TicketTier[];
  venueLayout?: VenueLayout;
}
```

Update `PurchasedTicket` to include seat info:

```typescript
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
```

---

### Task 2: Create Mock Venue Layouts

**File:** `frontend/lib/mock-data.ts`

Create 2 venue layout templates:

1. **Concert Arena** — used by Weeknd (MSG) and Billie Eilish (O2 Arena)
   - Rectangular stage at top
   - Curved sections wrapping around: VIP Floor (front), Premium sections (sides), GA sections (back/upper)
   - ~6 sections with SVG paths forming an arena shape

2. **Sports Arena** — used by NBA Finals (Chase Center) and UFC 310 (T-Mobile Arena)
   - Central court/octagon (rectangle or circle)
   - Sections surrounding it: Courtside/Cageside (inner ring), Lower Bowl (middle), Upper Deck (outer)
   - ~8 sections with SVG paths

Each section should have 2-4 rows with 8-12 seats each. Mark some seats as "taken" to make it realistic.

Wire `venueLayout` into each of the 4 existing events.

SVG viewBox should be `0 0 800 600` for all layouts. Keep paths simple — use basic shapes (arcs, rectangles) that clearly convey the venue structure.

---

### Task 3: Update Booking Context for Seat Selection

**File:** `frontend/lib/booking-context.tsx`

Changes:
- Update `CartItem` type usage — cart now holds `{ eventId, seats: SelectedSeat[] }`
- `setCart` accepts the new shape
- `purchaseTickets` should:
  - Look up each selected seat in the venue layout
  - Mark them as "taken" after purchase
  - Store seat info in `PurchasedTicket`
- Keep the existing double-booking race condition (read seats, delay, then mark taken)

---

### Task 4: Seat Selection Page — Section Overview (Step 1)

**File:** `frontend/app/events/[id]/seats/page.tsx`

Layout (matches the design file frame `uN4kX`):
- **Event Summary Bar** at top: event title, date/venue on left; step indicator (Select [active] → Checkout → Confirm) on right
- **Main area** split into:
  - **Left (flex-1):** SVG venue map
    - Render `stage` element (labeled rectangle/circle)
    - Render each section as an SVG `<path>` filled with `section.color`
    - Section label text at `labelPosition`
    - Hover: highlight section, show tooltip (section name, price, "X available")
    - Click: transition to Step 2 (section detail)
    - Unavailable sections (0 available) shown grayed out
  - **Right sidebar (w-[360px]):** Booking summary
    - "Your Seats" heading
    - Empty state: "Select seats from the venue map"
    - When seats selected: list of selected seats with section/row/seat and price, remove button
    - Total price
    - "Continue to Checkout" button (disabled until seats selected)
- **Legend** below the map: color swatches for each tier + price

**File:** `frontend/components/venue-map.tsx`

Props: `{ layout: VenueLayout; onSectionClick: (sectionId: string) => void; selectedSectionId?: string }`

Renders the full SVG venue map. Handles hover states and tooltips.

**File:** `frontend/components/seat-tooltip.tsx`

Floating tooltip component. Shows section info or seat info depending on context.

---

### Task 5: Seat Selection Page — Seat Picker (Step 2)

When a section is clicked, the main area transitions to show a zoomed-in seat grid for that section.

**File:** `frontend/components/section-picker.tsx`

Props: `{ section: VenueSection; selectedSeats: SelectedSeat[]; onSeatToggle: (seat, row, section) => void; onBack: () => void }`

Layout:
- "← Back to venue map" link + section name heading
- Seat grid:
  - Each row is a horizontal line of seats
  - Row label on the left ("Row A", "Row B", etc.)
  - Each seat is a rounded rectangle or circle (~28px)
  - Colors: available = section tier color (lighter), taken = gray, selected = primary orange
  - Hover available seat: tooltip with "Row A, Seat 5 — $185"
  - Click available seat: toggle selection
  - Click selected seat: deselect
- Legend below: Available / Taken / Selected color indicators

**File:** `frontend/components/seat-legend.tsx`

Simple legend component used in both views.

---

### Task 6: Update Event Detail Page

**File:** `frontend/app/events/[id]/page.tsx`

Changes:
- Keep the existing tier display as informational (shows price ranges per tier)
- Change "Buy Tickets" button to **"Select Seats"** → navigates to `/events/[id]/seats`
- Remove quantity selector (quantity is determined by number of seats selected)

---

### Task 7: Update Checkout Page

**File:** `frontend/app/checkout/page.tsx`

Changes:
- Update order summary to show individual seats instead of tier × quantity
- List each seat: "Section A, Row 2, Seat 5 — $185"
- Subtotal = sum of all seat prices
- Keep service fee (12%) and handling charge logic

---

### Task 8: Update My Tickets Page

**File:** `frontend/app/my-tickets/page.tsx`

Changes:
- Show seat details on purchased tickets: "Seats: Section A Row 2 Seat 5, Section A Row 2 Seat 6"
- Keep existing layout otherwise
