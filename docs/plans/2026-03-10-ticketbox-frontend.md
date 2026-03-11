# TicketBox Frontend Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build a Next.js frontend for a ticket booking app (TicketBox) that demonstrates the double booking problem.

**Architecture:** Next.js App Router with TypeScript, Tailwind CSS for styling, React Context for state management (cart + purchased tickets). Mock data with static events array. No backend — frontend-only with in-memory state.

**Tech Stack:** Next.js 14+, TypeScript, Tailwind CSS, React Context API, Lucide React (icons)

**Design tokens from .pen file:**
- Primary orange: `#FF8400`
- Text primary: `#111111`
- Text secondary: `#666666`
- Background: `#F2F3F0`
- Card bg: `#FFFFFF`
- Border: `#CBCCC9`
- Success green: `#004D1A` / `#DFE6E1`
- Font: Geist (sans-serif)
- Border radius cards: default (use rounded-lg)
- Navbar height: 64px

---

### Task 1: Scaffold Next.js Project

**Step 1:** Create Next.js app with TypeScript + Tailwind

```bash
cd /Users/dev/work/booking
npx create-next-app@latest frontend --typescript --tailwind --app --eslint --no-src-dir --import-alias "@/*"
```

**Step 2:** Install dependencies

```bash
cd frontend && npm install lucide-react
```

**Step 3:** Update `tailwind.config.ts` with design tokens (colors, fonts)

**Step 4:** Update `app/globals.css` with base styles

**Step 5:** Commit

---

### Task 2: Mock Data & Types

**Files:**
- Create: `frontend/lib/types.ts`
- Create: `frontend/lib/mock-data.ts`

Types: `Event`, `TicketTier`, `PurchasedTicket`, `CartItem`

Mock events:
1. The Weeknd — After Hours World Tour (Concert, MSG NY, Jul 15 2026, $130+)
2. Billie Eilish — Hit Me Hard and Soft Tour (Concert, O2 Arena London, Apr 8 2026, $85+)
3. NBA Finals 2026 — Game 3 (Sports, Chase Center SF, Jun 12 2026, $120+)
4. UFC 310 — Championship (Sports, T-Mobile Arena LV, May 3 2026, $95+)

Each event has 3 ticket tiers with `available` count (key for double booking demo).

---

### Task 3: Booking Context (State Management)

**Files:**
- Create: `frontend/lib/booking-context.tsx`

State: `{ cart, purchasedTickets, addToCart, clearCart, purchaseTickets, events }`

Events stored in context with mutable `available` counts. `purchaseTickets` decrements availability with a simulated delay (to expose race condition).

---

### Task 4: Shared Components

**Files:**
- Create: `frontend/components/navbar.tsx` — Logo, nav links, search, avatar
- Create: `frontend/components/event-card.tsx` — Card with image, title, date/venue, price, category badge
- Create: `frontend/components/category-tabs.tsx` — All/Concerts/Sports/Films filter

---

### Task 5: Home Page

**Files:**
- Modify: `frontend/app/page.tsx`
- Create: `frontend/app/layout.tsx` (update with context provider + navbar)

Sections: Hero banner (featured event) + Trending Events grid with category filtering.

---

### Task 6: Event Detail Page

**Files:**
- Create: `frontend/app/events/[id]/page.tsx`

Layout: Banner image + overlay, left column (About, Date/Venue/Genre), right column (ticket tier selection, quantity, total, Buy Tickets button).

---

### Task 7: Checkout Page

**Files:**
- Create: `frontend/app/checkout/page.tsx`

Layout: Step indicator, payment form (name, email, card — no validation needed), order summary sidebar. "Pay Now" completes purchase with simulated delay.

---

### Task 8: My Tickets Page

**Files:**
- Create: `frontend/app/my-tickets/page.tsx`

Layout: Success alert (conditional), ticket list with event details, status badges, View Ticket buttons.

---
