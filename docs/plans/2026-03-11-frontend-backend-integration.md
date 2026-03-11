# Frontend-Backend Integration Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Connect the Next.js frontend to the Go microservices backend via the HTTP gateway, replacing all mock data with real API calls and adding authentication.

**Architecture:** Create an API client layer (`lib/api/`) that talks to the gateway at `localhost:8000`. Add auth context for login/register/token management. Refactor `BookingProvider` to fetch from API instead of mock data. Add login/register pages. Proxy API requests through Next.js rewrites to avoid CORS in production.

**Tech Stack:** Next.js 16, React 19, TypeScript 5, Go gateway on port 8000 (Gin), JWT auth, PostgreSQL, Redis, Kafka

---

## Data Shape Mapping (Backend → Frontend)

The backend returns different shapes than the frontend currently expects. Key differences:

| Field | Backend API | Frontend Type | Transform Needed |
|-------|-------------|---------------|-----------------|
| Price | `price_cents: 15000` (int64) | `price: 150` (number) | Divide by 100 |
| Date | `date: "2026-06-15T20:00:00Z"` (RFC3339) | `date: "Jun 15"`, `time: "8:00 PM"` | Parse + format |
| Image | `image_url: string` | `image: string` | Rename field |
| Tier availability | `available_quantity: int32` | `available: number` | Rename field |
| Tier total | `total_quantity: int32` | Not used | Drop |
| Event status | `status: string` | Not in current type | Add optionally |
| Booking total | `total_amount_cents: int64` | `totalPrice: number` | Divide by 100 |
| Venue layout | Not in backend | `venueLayout?: VenueLayout` | Keep client-side for now |

---

## Task 1: API Client Foundation

**Files:**
- Create: `frontend/lib/api/client.ts`
- Create: `frontend/lib/api/types.ts`
- Modify: `frontend/next.config.ts`
- Create: `frontend/.env.local`

### Step 1: Create `.env.local` with API base URL

```env
NEXT_PUBLIC_API_URL=http://localhost:8000
```

### Step 2: Configure Next.js API rewrites (proxy to avoid CORS issues)

Modify `frontend/next.config.ts`:

```typescript
import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  async rewrites() {
    return [
      {
        source: "/api/:path*",
        destination: `${process.env.NEXT_PUBLIC_API_URL || "http://localhost:8000"}/api/:path*`,
      },
    ];
  },
};

export default nextConfig;
```

### Step 3: Create API response types matching backend JSON

Create `frontend/lib/api/types.ts`:

```typescript
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
}

export interface ApiBooking {
  id: string;
  user_id: string;
  event_id: string;
  status: string;
  total_amount_cents: number;
  items: ApiBookingItem[];
  created_at: string;
}

export interface ApiListBookingsResponse {
  bookings: ApiBooking[];
  total: number;
}

export interface ApiError {
  error: string;
}
```

### Step 4: Create the HTTP client with token handling

Create `frontend/lib/api/client.ts`:

```typescript
import { ApiAuthResponse, ApiEvent, ApiListEventsResponse, ApiBooking, ApiListBookingsResponse, ApiUser } from "./types";

class ApiClient {
  private baseUrl: string;

  constructor() {
    // Use relative URL so Next.js rewrites proxy to backend
    this.baseUrl = "/api";
  }

  private getToken(): string | null {
    if (typeof window === "undefined") return null;
    return localStorage.getItem("access_token");
  }

  private getRefreshToken(): string | null {
    if (typeof window === "undefined") return null;
    return localStorage.getItem("refresh_token");
  }

  private setTokens(access: string, refresh: string) {
    localStorage.setItem("access_token", access);
    localStorage.setItem("refresh_token", refresh);
  }

  private clearTokens() {
    localStorage.removeItem("access_token");
    localStorage.removeItem("refresh_token");
  }

  private async request<T>(path: string, options: RequestInit = {}): Promise<T> {
    const token = this.getToken();
    const headers: Record<string, string> = {
      "Content-Type": "application/json",
      ...((options.headers as Record<string, string>) || {}),
    };
    if (token) {
      headers["Authorization"] = `Bearer ${token}`;
    }

    const res = await fetch(`${this.baseUrl}${path}`, {
      ...options,
      headers,
    });

    if (res.status === 401 && token) {
      // Try refresh
      const refreshed = await this.tryRefresh();
      if (refreshed) {
        headers["Authorization"] = `Bearer ${this.getToken()}`;
        const retry = await fetch(`${this.baseUrl}${path}`, { ...options, headers });
        if (!retry.ok) {
          const err = await retry.json().catch(() => ({ error: "Request failed" }));
          throw new Error(err.error || "Request failed");
        }
        return retry.json();
      }
      this.clearTokens();
      throw new Error("Session expired");
    }

    if (!res.ok) {
      const err = await res.json().catch(() => ({ error: "Request failed" }));
      throw new Error(err.error || `Request failed with status ${res.status}`);
    }

    return res.json();
  }

  private async tryRefresh(): Promise<boolean> {
    const refreshToken = this.getRefreshToken();
    if (!refreshToken) return false;
    try {
      const res = await fetch(`${this.baseUrl}/auth/refresh`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ refresh_token: refreshToken }),
      });
      if (!res.ok) return false;
      const data: ApiAuthResponse = await res.json();
      this.setTokens(data.access_token, data.refresh_token);
      return true;
    } catch {
      return false;
    }
  }

  // ── Auth ──────────────────────────────────────────────
  async register(email: string, password: string, name: string): Promise<ApiAuthResponse> {
    const data = await this.request<ApiAuthResponse>("/auth/register", {
      method: "POST",
      body: JSON.stringify({ email, password, name }),
    });
    this.setTokens(data.access_token, data.refresh_token);
    return data;
  }

  async login(email: string, password: string): Promise<ApiAuthResponse> {
    const data = await this.request<ApiAuthResponse>("/auth/login", {
      method: "POST",
      body: JSON.stringify({ email, password }),
    });
    this.setTokens(data.access_token, data.refresh_token);
    return data;
  }

  async logout(): Promise<void> {
    try {
      await this.request("/auth/logout", { method: "POST" });
    } finally {
      this.clearTokens();
    }
  }

  async getProfile(): Promise<ApiUser> {
    return this.request<ApiUser>("/users/me");
  }

  // ── Events ────────────────────────────────────────────
  async listEvents(params?: { category?: string; search?: string; page?: number; page_size?: number }): Promise<ApiListEventsResponse> {
    const query = new URLSearchParams();
    if (params?.category) query.set("category", params.category);
    if (params?.search) query.set("search", params.search);
    if (params?.page) query.set("page", String(params.page));
    if (params?.page_size) query.set("page_size", String(params.page_size));
    const qs = query.toString();
    return this.request<ApiListEventsResponse>(`/events${qs ? `?${qs}` : ""}`);
  }

  async getEvent(id: string): Promise<ApiEvent> {
    return this.request<ApiEvent>(`/events/${id}`);
  }

  // ── Bookings ──────────────────────────────────────────
  async createBooking(eventId: string, items: { ticket_tier_id: string; quantity: number }[]): Promise<ApiBooking> {
    return this.request<ApiBooking>("/bookings", {
      method: "POST",
      body: JSON.stringify({ event_id: eventId, items }),
    });
  }

  async listBookings(page?: number, pageSize?: number): Promise<ApiListBookingsResponse> {
    const query = new URLSearchParams();
    if (page) query.set("page", String(page));
    if (pageSize) query.set("page_size", String(pageSize));
    const qs = query.toString();
    return this.request<ApiListBookingsResponse>(`/bookings${qs ? `?${qs}` : ""}`);
  }

  async getBooking(id: string): Promise<ApiBooking> {
    return this.request<ApiBooking>(`/bookings/${id}`);
  }

  async cancelBooking(id: string): Promise<ApiBooking> {
    return this.request<ApiBooking>(`/bookings/${id}/cancel`, { method: "POST" });
  }

  // ── Token state check ─────────────────────────────────
  isLoggedIn(): boolean {
    return !!this.getToken();
  }
}

export const apiClient = new ApiClient();
```

### Step 5: Verify the frontend dev server starts

Run: `cd /Users/dev/work/booking/frontend && npm run dev`
Expected: Dev server starts on port 3000, no TypeScript errors

### Step 6: Commit

```bash
git add frontend/lib/api/ frontend/next.config.ts frontend/.env.local
git commit -m "feat: add API client layer with auth, events, and bookings endpoints"
```

---

## Task 2: Data Transformers (Backend → Frontend Types)

**Files:**
- Create: `frontend/lib/api/transformers.ts`

### Step 1: Create transformer functions

Create `frontend/lib/api/transformers.ts`:

```typescript
import { ApiEvent, ApiTicketTier, ApiBooking } from "./types";
import { Event, TicketTier, PurchasedTicket } from "../types";

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
    // venueLayout is not provided by backend — kept for future seat-map integration
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
```

### Step 2: Commit

```bash
git add frontend/lib/api/transformers.ts
git commit -m "feat: add data transformers from backend API types to frontend types"
```

---

## Task 3: Auth Context & Login/Register Pages

**Files:**
- Create: `frontend/lib/auth-context.tsx`
- Create: `frontend/app/login/page.tsx`
- Create: `frontend/app/register/page.tsx`
- Modify: `frontend/app/layout.tsx`
- Modify: `frontend/components/navbar.tsx`

### Step 1: Create AuthContext

Create `frontend/lib/auth-context.tsx`:

```typescript
"use client";

import React, { createContext, useContext, useState, useEffect, useCallback } from "react";
import { apiClient } from "./api/client";

interface User {
  id: string;
  email: string;
  name: string;
  role: string;
}

interface AuthState {
  user: User | null;
  isLoading: boolean;
  login: (email: string, password: string) => Promise<void>;
  register: (email: string, password: string, name: string) => Promise<void>;
  logout: () => Promise<void>;
}

const AuthContext = createContext<AuthState | null>(null);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  // Check for existing session on mount
  useEffect(() => {
    if (apiClient.isLoggedIn()) {
      apiClient
        .getProfile()
        .then((profile) => {
          setUser({ id: profile.id, email: profile.email, name: profile.name, role: profile.role });
        })
        .catch(() => {
          // Token expired or invalid
        })
        .finally(() => setIsLoading(false));
    } else {
      setIsLoading(false);
    }
  }, []);

  const login = useCallback(async (email: string, password: string) => {
    const res = await apiClient.login(email, password);
    setUser({ id: res.user.id, email: res.user.email, name: res.user.name, role: res.user.role });
  }, []);

  const register = useCallback(async (email: string, password: string, name: string) => {
    const res = await apiClient.register(email, password, name);
    setUser({ id: res.user.id, email: res.user.email, name: res.user.name, role: res.user.role });
  }, []);

  const logout = useCallback(async () => {
    await apiClient.logout();
    setUser(null);
  }, []);

  return (
    <AuthContext.Provider value={{ user, isLoading, login, register, logout }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error("useAuth must be used within AuthProvider");
  return ctx;
}
```

### Step 2: Create Login page

Create `frontend/app/login/page.tsx`:

```typescript
"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { Loader2 } from "lucide-react";
import { useAuth } from "@/lib/auth-context";

export default function LoginPage() {
  const router = useRouter();
  const { login } = useAuth();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [isLoading, setIsLoading] = useState(false);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    setIsLoading(true);
    try {
      await login(email, password);
      router.push("/");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Login failed");
    } finally {
      setIsLoading(false);
    }
  }

  return (
    <div className="flex items-center justify-center min-h-[calc(100vh-4rem)]">
      <div className="w-full max-w-md bg-card border border-border rounded-lg p-8 shadow-sm">
        <h1 className="text-2xl font-bold mb-6">Sign In</h1>
        {error && (
          <div className="bg-red-50 text-red-600 text-sm p-3 rounded-lg mb-4">{error}</div>
        )}
        <form onSubmit={handleSubmit} className="flex flex-col gap-4">
          <div className="flex flex-col gap-1.5">
            <label className="text-xs font-medium text-muted">Email</label>
            <input
              type="email"
              required
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              className="border border-border rounded-lg px-3 py-2.5 text-sm bg-card focus:outline-none focus:border-primary transition-colors"
              placeholder="you@example.com"
            />
          </div>
          <div className="flex flex-col gap-1.5">
            <label className="text-xs font-medium text-muted">Password</label>
            <input
              type="password"
              required
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              className="border border-border rounded-lg px-3 py-2.5 text-sm bg-card focus:outline-none focus:border-primary transition-colors"
              placeholder="••••••••"
            />
          </div>
          <button
            type="submit"
            disabled={isLoading}
            className="flex items-center justify-center gap-2 bg-primary hover:bg-primary-hover disabled:opacity-70 text-white font-medium rounded-full h-12 transition-colors mt-2"
          >
            {isLoading ? <Loader2 size={16} className="animate-spin" /> : null}
            {isLoading ? "Signing in..." : "Sign In"}
          </button>
        </form>
        <p className="text-sm text-muted text-center mt-6">
          Don&apos;t have an account?{" "}
          <Link href="/register" className="text-primary font-medium hover:underline">
            Sign Up
          </Link>
        </p>
      </div>
    </div>
  );
}
```

### Step 3: Create Register page

Create `frontend/app/register/page.tsx`:

```typescript
"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { Loader2 } from "lucide-react";
import { useAuth } from "@/lib/auth-context";

export default function RegisterPage() {
  const router = useRouter();
  const { register } = useAuth();
  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [isLoading, setIsLoading] = useState(false);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    setIsLoading(true);
    try {
      await register(email, password, name);
      router.push("/");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Registration failed");
    } finally {
      setIsLoading(false);
    }
  }

  return (
    <div className="flex items-center justify-center min-h-[calc(100vh-4rem)]">
      <div className="w-full max-w-md bg-card border border-border rounded-lg p-8 shadow-sm">
        <h1 className="text-2xl font-bold mb-6">Create Account</h1>
        {error && (
          <div className="bg-red-50 text-red-600 text-sm p-3 rounded-lg mb-4">{error}</div>
        )}
        <form onSubmit={handleSubmit} className="flex flex-col gap-4">
          <div className="flex flex-col gap-1.5">
            <label className="text-xs font-medium text-muted">Full Name</label>
            <input
              type="text"
              required
              value={name}
              onChange={(e) => setName(e.target.value)}
              className="border border-border rounded-lg px-3 py-2.5 text-sm bg-card focus:outline-none focus:border-primary transition-colors"
              placeholder="John Doe"
            />
          </div>
          <div className="flex flex-col gap-1.5">
            <label className="text-xs font-medium text-muted">Email</label>
            <input
              type="email"
              required
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              className="border border-border rounded-lg px-3 py-2.5 text-sm bg-card focus:outline-none focus:border-primary transition-colors"
              placeholder="you@example.com"
            />
          </div>
          <div className="flex flex-col gap-1.5">
            <label className="text-xs font-medium text-muted">Password</label>
            <input
              type="password"
              required
              minLength={6}
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              className="border border-border rounded-lg px-3 py-2.5 text-sm bg-card focus:outline-none focus:border-primary transition-colors"
              placeholder="••••••••"
            />
          </div>
          <button
            type="submit"
            disabled={isLoading}
            className="flex items-center justify-center gap-2 bg-primary hover:bg-primary-hover disabled:opacity-70 text-white font-medium rounded-full h-12 transition-colors mt-2"
          >
            {isLoading ? <Loader2 size={16} className="animate-spin" /> : null}
            {isLoading ? "Creating account..." : "Create Account"}
          </button>
        </form>
        <p className="text-sm text-muted text-center mt-6">
          Already have an account?{" "}
          <Link href="/login" className="text-primary font-medium hover:underline">
            Sign In
          </Link>
        </p>
      </div>
    </div>
  );
}
```

### Step 4: Wrap layout with AuthProvider

Modify `frontend/app/layout.tsx` — add AuthProvider wrapping BookingProvider:

```typescript
import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import "./globals.css";
import { AuthProvider } from "@/lib/auth-context";
import { BookingProvider } from "@/lib/booking-context";
import { Navbar } from "@/components/navbar";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  title: "TicketBox — Book Event Tickets",
  description: "Your one-stop destination for concert, sports, and entertainment tickets",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body className={`${geistSans.variable} ${geistMono.variable} antialiased`}>
        <AuthProvider>
          <BookingProvider>
            <div className="min-h-screen flex flex-col">
              <Navbar />
              <main className="flex-1">{children}</main>
            </div>
          </BookingProvider>
        </AuthProvider>
      </body>
    </html>
  );
}
```

### Step 5: Update Navbar with auth state

Modify `frontend/components/navbar.tsx`:

```typescript
"use client";

import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { Search, User, LogOut } from "lucide-react";
import { useAuth } from "@/lib/auth-context";

const navLinks = [
  { href: "/", label: "Discover" },
  { href: "/?category=Concerts", label: "Concerts" },
  { href: "/?category=Sports", label: "Sports" },
  { href: "/?category=Films", label: "Films" },
];

export function Navbar() {
  const pathname = usePathname();
  const router = useRouter();
  const { user, logout } = useAuth();

  async function handleLogout() {
    await logout();
    router.push("/");
  }

  return (
    <nav className="flex items-center justify-between h-16 px-12 bg-card border-b border-border w-full shrink-0">
      <div className="flex items-center gap-8">
        <Link href="/" className="text-primary font-bold text-xl font-mono tracking-tight">
          TICKETBOX
        </Link>
        <div className="flex items-center gap-6">
          {navLinks.map((link) => (
            <Link
              key={link.label}
              href={link.href}
              className={`text-sm font-medium transition-colors hover:text-primary ${
                pathname === link.href ? "text-primary" : "text-muted"
              }`}
            >
              {link.label}
            </Link>
          ))}
        </div>
      </div>
      <div className="flex items-center gap-4">
        <div className="flex items-center gap-2 border border-border rounded-sm px-3 py-1.5 w-48">
          <Search size={14} className="text-muted" />
          <span className="text-sm text-muted">Search...</span>
        </div>
        {user ? (
          <div className="flex items-center gap-3">
            <Link
              href="/my-tickets"
              className="text-sm font-medium text-muted hover:text-primary transition-colors"
            >
              My Tickets
            </Link>
            <span className="text-sm font-medium">{user.name}</span>
            <button
              onClick={handleLogout}
              className="flex items-center justify-center w-10 h-10 rounded-full bg-tag-bg border border-border hover:bg-red-50 transition-colors"
              title="Logout"
            >
              <LogOut size={16} className="text-muted" />
            </button>
          </div>
        ) : (
          <Link
            href="/login"
            className="flex items-center justify-center gap-2 bg-primary hover:bg-primary-hover text-white font-medium rounded-full px-5 h-10 text-sm transition-colors"
          >
            <User size={14} />
            Sign In
          </Link>
        )}
      </div>
    </nav>
  );
}
```

### Step 6: Verify the frontend dev server starts with no TS errors

Run: `cd /Users/dev/work/booking/frontend && npx tsc --noEmit`
Expected: No errors

### Step 7: Commit

```bash
git add frontend/lib/auth-context.tsx frontend/app/login/ frontend/app/register/ frontend/app/layout.tsx frontend/components/navbar.tsx
git commit -m "feat: add auth context, login/register pages, and auth-aware navbar"
```

---

## Task 4: Refactor BookingProvider to Fetch Events from API

**Files:**
- Modify: `frontend/lib/booking-context.tsx`

### Step 1: Rewrite BookingProvider to use API for events

Replace `frontend/lib/booking-context.tsx` with:

```typescript
"use client";

import React, { createContext, useContext, useState, useCallback, useEffect } from "react";
import { Event, CartItem, PurchasedTicket } from "./types";
import { apiClient } from "./api/client";
import { transformEvent, transformBookingToTicket } from "./api/transformers";
import { events as mockEvents } from "./mock-data";

interface BookingState {
  events: Event[];
  cart: CartItem | null;
  purchasedTickets: PurchasedTicket[];
  lastPurchaseSuccess: boolean;
  isLoading: boolean;
  error: string | null;
  setCart: (item: CartItem | null) => void;
  purchaseTickets: (name: string, email: string) => Promise<boolean>;
  clearSuccessFlag: () => void;
  getEvent: (id: string) => Event | undefined;
  refreshEvents: () => Promise<void>;
}

const BookingContext = createContext<BookingState | null>(null);

export function BookingProvider({ children }: { children: React.ReactNode }) {
  const [events, setEvents] = useState<Event[]>([]);
  const [cart, setCart] = useState<CartItem | null>(null);
  const [purchasedTickets, setPurchasedTickets] = useState<PurchasedTicket[]>([]);
  const [lastPurchaseSuccess, setLastPurchaseSuccess] = useState(false);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchEvents = useCallback(async () => {
    setIsLoading(true);
    setError(null);
    try {
      const res = await apiClient.listEvents({ page_size: 50 });
      setEvents(res.events.map(transformEvent));
    } catch {
      // Fallback to mock data if backend is unavailable
      console.warn("Backend unavailable, using mock data");
      setEvents(JSON.parse(JSON.stringify(mockEvents)));
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchEvents();
  }, [fetchEvents]);

  const clearSuccessFlag = useCallback(() => setLastPurchaseSuccess(false), []);

  const getEvent = useCallback(
    (id: string) => events.find((e) => e.id === id),
    [events]
  );

  const purchaseTickets = useCallback(
    async (_name: string, _email: string): Promise<boolean> => {
      if (!cart || cart.seats.length === 0) return false;

      const event = events.find((e) => e.id === cart.eventId);
      if (!event) return false;

      try {
        // Group seats by tier/section to create booking items
        const itemMap = new Map<string, { ticket_tier_id: string; quantity: number }>();
        for (const seat of cart.seats) {
          // Find the tier matching this seat's section
          const tier = event.tiers.find((t) => t.name === seat.sectionName);
          if (tier) {
            const existing = itemMap.get(tier.id);
            if (existing) {
              existing.quantity += 1;
            } else {
              itemMap.set(tier.id, { ticket_tier_id: tier.id, quantity: 1 });
            }
          }
        }

        const items = Array.from(itemMap.values());
        if (items.length === 0) return false;

        const booking = await apiClient.createBooking(cart.eventId, items);
        const ticket = transformBookingToTicket(booking, event);

        // Add seat details from cart since backend doesn't track individual seats
        ticket.seats = cart.seats.map((s) => ({
          section: s.sectionName,
          row: s.rowLabel,
          seat: s.seatLabel,
        }));

        setPurchasedTickets((prev) => [ticket, ...prev]);
        setCart(null);
        setLastPurchaseSuccess(true);

        // Refresh events to get updated availability
        fetchEvents();
        return true;
      } catch (err) {
        console.error("Booking failed:", err);
        return false;
      }
    },
    [cart, events, fetchEvents]
  );

  return (
    <BookingContext.Provider
      value={{
        events,
        cart,
        purchasedTickets,
        lastPurchaseSuccess,
        isLoading,
        error,
        setCart,
        purchaseTickets,
        clearSuccessFlag,
        getEvent,
        refreshEvents: fetchEvents,
      }}
    >
      {children}
    </BookingContext.Provider>
  );
}

export function useBooking() {
  const ctx = useContext(BookingContext);
  if (!ctx) throw new Error("useBooking must be used within BookingProvider");
  return ctx;
}
```

**Key changes:**
- Fetches events from API on mount with fallback to mock data
- `purchaseTickets` calls `apiClient.createBooking()` instead of client-side simulation
- Refreshes event list after purchase to update availability
- Added `isLoading` and `error` state
- Mock data import kept as fallback when backend is offline

### Step 2: Verify TypeScript compiles

Run: `cd /Users/dev/work/booking/frontend && npx tsc --noEmit`
Expected: No errors

### Step 3: Commit

```bash
git add frontend/lib/booking-context.tsx
git commit -m "feat: refactor BookingProvider to fetch events from API with mock fallback"
```

---

## Task 5: Update Homepage to Handle Loading State

**Files:**
- Modify: `frontend/app/page.tsx`

### Step 1: Add loading state handling

Modify `frontend/app/page.tsx` — add loading skeleton while events load:

```typescript
"use client";

import { useState } from "react";
import Link from "next/link";
import { Ticket, Sparkles, Loader2 } from "lucide-react";
import { useBooking } from "@/lib/booking-context";
import { EventCard } from "@/components/event-card";

const categories = ["All", "Concerts", "Sports", "Films"] as const;

export default function HomePage() {
  const { events, isLoading } = useBooking();
  const [activeCategory, setActiveCategory] = useState<string>("All");

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-96">
        <Loader2 size={32} className="animate-spin text-primary" />
      </div>
    );
  }

  const featured = events.find((e) => e.featured) || events[0];
  const filteredEvents =
    activeCategory === "All"
      ? events
      : events.filter((e) => e.category === activeCategory);

  if (!featured) {
    return (
      <div className="flex items-center justify-center h-96">
        <p className="text-muted text-lg">No events available.</p>
      </div>
    );
  }

  return (
    <div className="flex flex-col">
      {/* Hero Section */}
      <div className="relative h-[360px] overflow-hidden">
        <img
          src={featured.image}
          alt={featured.title}
          className="w-full h-full object-cover"
        />
        <div className="absolute inset-0 bg-gradient-to-t from-black/80 to-transparent" />
        <div className="absolute inset-0 flex flex-col justify-end p-12 gap-4">
          <div className="flex items-center gap-1.5 bg-[#E9E3D8] rounded-full px-3 py-1.5 w-fit border border-border">
            <Sparkles size={12} />
            <span className="text-xs font-medium">Featured</span>
          </div>
          <h1 className="text-white text-4xl font-bold leading-tight max-w-xl">
            {featured.title}
          </h1>
          <div className="flex items-center gap-6 text-white/80 text-sm">
            <span>{featured.date}, {featured.time}</span>
            <span>{featured.venue}, {featured.location}</span>
          </div>
          <Link
            href={`/events/${featured.id}`}
            className="flex items-center gap-2 bg-primary hover:bg-primary-hover text-white font-medium rounded-full px-6 h-12 w-fit transition-colors"
          >
            <Ticket size={16} />
            Get Tickets
          </Link>
        </div>
      </div>

      {/* Trending Events */}
      <div className="flex flex-col gap-8 px-12 py-8">
        <div className="flex items-center justify-between">
          <h2 className="text-2xl font-bold">Trending Events</h2>
          <div className="flex items-center gap-1 bg-tag-bg rounded-full p-1">
            {categories.map((cat) => (
              <button
                key={cat}
                onClick={() => setActiveCategory(cat)}
                className={`px-4 py-1.5 rounded-full text-sm font-medium transition-colors ${
                  activeCategory === cat
                    ? "bg-card text-foreground shadow-sm"
                    : "text-muted hover:text-foreground"
                }`}
              >
                {cat}
              </button>
            ))}
          </div>
        </div>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {filteredEvents.map((event) => (
            <EventCard key={event.id} event={event} />
          ))}
        </div>
      </div>
    </div>
  );
}
```

### Step 2: Commit

```bash
git add frontend/app/page.tsx
git commit -m "feat: add loading state to homepage"
```

---

## Task 6: Update Checkout to Require Auth

**Files:**
- Modify: `frontend/app/checkout/page.tsx`

### Step 1: Add auth guard and simplify checkout

The checkout page currently collects payment details client-side. Since the backend handles bookings via `POST /api/bookings`, we need to:
1. Require the user to be logged in
2. Remove fake card fields (backend doesn't process payments yet)
3. Show login prompt if not authenticated

Modify `frontend/app/checkout/page.tsx`:

```typescript
"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { CreditCard, Check, Loader2 } from "lucide-react";
import { useBooking } from "@/lib/booking-context";
import { useAuth } from "@/lib/auth-context";

export default function CheckoutPage() {
  const router = useRouter();
  const { cart, getEvent, purchaseTickets } = useBooking();
  const { user } = useAuth();
  const [isProcessing, setIsProcessing] = useState(false);
  const [error, setError] = useState("");

  if (!cart) {
    return (
      <div className="flex items-center justify-center h-96">
        <p className="text-muted text-lg">No items in cart. Please select tickets first.</p>
      </div>
    );
  }

  if (!user) {
    return (
      <div className="flex items-center justify-center h-96">
        <div className="flex flex-col items-center gap-4">
          <p className="text-muted text-lg">Please sign in to complete your purchase.</p>
          <Link
            href="/login"
            className="bg-primary hover:bg-primary-hover text-white font-medium rounded-full px-6 h-12 flex items-center transition-colors"
          >
            Sign In
          </Link>
        </div>
      </div>
    );
  }

  const event = getEvent(cart.eventId);
  if (!event) return null;

  const subtotal = cart.seats.reduce((sum, s) => sum + s.price, 0);
  const serviceFee = Math.round(subtotal * 0.12);
  const handlingCharge = 3.5;
  const total = subtotal + serviceFee + handlingCharge;

  async function handlePay() {
    setIsProcessing(true);
    setError("");
    const success = await purchaseTickets(user!.name, user!.email);
    setIsProcessing(false);
    if (success) {
      router.push("/my-tickets");
    } else {
      setError("Purchase failed — tickets may have sold out. Please try again.");
    }
  }

  const steps = [
    { label: "Select", done: true },
    { label: "Checkout", active: true },
    { label: "Confirm", done: false },
  ];

  return (
    <div className="flex flex-col">
      {/* Step Bar */}
      <div className="flex items-center justify-between h-14 px-12 bg-card border-b border-border">
        <div className="flex items-center gap-4">
          <span className="text-sm font-semibold">{event.title}</span>
          <span className="text-sm text-muted">
            {event.date} · {event.venue}
          </span>
        </div>
        <div className="flex items-center gap-2">
          {steps.map((step, i) => (
            <div key={step.label} className="flex items-center gap-2">
              <div className="flex items-center gap-1.5">
                <div
                  className={`w-6 h-6 rounded-full flex items-center justify-center text-xs font-bold ${
                    step.done
                      ? "bg-primary text-white"
                      : step.active
                        ? "bg-primary text-white"
                        : "bg-tag-bg text-muted"
                  }`}
                >
                  {step.done && !step.active ? <Check size={12} /> : i + 1}
                </div>
                <span
                  className={`text-sm ${
                    step.active ? "font-semibold" : "text-muted"
                  }`}
                >
                  {step.label}
                </span>
              </div>
              {i < steps.length - 1 && (
                <div
                  className={`w-8 h-0.5 ${
                    step.done ? "bg-primary" : "bg-border"
                  }`}
                />
              )}
            </div>
          ))}
        </div>
      </div>

      {/* Content */}
      <div className="flex gap-8 px-12 py-8">
        {/* Booking Info */}
        <div className="flex-1">
          <div className="bg-card border border-border rounded-lg p-8 shadow-sm flex flex-col gap-6">
            <h2 className="text-lg font-bold">Booking Confirmation</h2>
            <div className="flex flex-col gap-4">
              <div className="flex flex-col gap-1.5">
                <span className="text-xs font-medium text-muted">Name</span>
                <span className="text-sm font-medium">{user.name}</span>
              </div>
              <div className="flex flex-col gap-1.5">
                <span className="text-xs font-medium text-muted">Email</span>
                <span className="text-sm font-medium">{user.email}</span>
              </div>
              <div className="flex flex-col gap-1.5">
                <span className="text-xs font-medium text-muted">Selected Seats</span>
                <div className="flex flex-col gap-1">
                  {cart.seats.map((seat) => (
                    <span key={seat.seatId} className="text-sm">
                      {seat.sectionName}, {seat.rowLabel}, Seat {seat.seatLabel}
                    </span>
                  ))}
                </div>
              </div>
            </div>
            {error && (
              <div className="bg-red-50 text-red-600 text-sm p-3 rounded-lg">{error}</div>
            )}
          </div>
        </div>

        {/* Order Summary */}
        <div className="w-[400px] shrink-0">
          <div className="bg-card border border-border rounded-lg shadow-sm flex flex-col">
            <div className="flex items-center gap-4 p-6 border-b border-border">
              <div className="w-16 h-16 rounded-lg overflow-hidden shrink-0">
                <img src={event.image} alt={event.title} className="w-full h-full object-cover" />
              </div>
              <div className="flex flex-col gap-0.5">
                <h3 className="font-semibold text-sm">{event.title}</h3>
                <p className="text-xs text-muted">
                  {event.date} · {event.venue}, {event.location}
                </p>
              </div>
            </div>
            <div className="flex flex-col gap-4 p-6">
              <h3 className="font-bold">Order Summary</h3>
              {cart.seats.map((seat) => (
                <div key={seat.seatId} className="flex items-center justify-between text-sm">
                  <span className="text-muted">
                    {seat.sectionName}, {seat.rowLabel}, Seat {seat.seatLabel}
                  </span>
                  <span className="font-medium">${seat.price.toFixed(2)}</span>
                </div>
              ))}
              <div className="flex items-center justify-between text-sm">
                <span className="text-muted">Service fee</span>
                <span className="font-medium">${serviceFee.toFixed(2)}</span>
              </div>
              <div className="flex items-center justify-between text-sm">
                <span className="text-muted">Handling charge</span>
                <span className="font-medium">${handlingCharge.toFixed(2)}</span>
              </div>
              <div className="flex items-center justify-between pt-4 border-t border-border">
                <span className="font-medium">Total</span>
                <span className="text-2xl font-bold text-primary">${total.toFixed(2)}</span>
              </div>
              <button
                onClick={handlePay}
                disabled={isProcessing}
                className="flex items-center justify-center gap-2 bg-primary hover:bg-primary-hover disabled:opacity-70 text-white font-medium rounded-full h-12 transition-colors"
              >
                {isProcessing ? (
                  <>
                    <Loader2 size={16} className="animate-spin" />
                    Processing...
                  </>
                ) : (
                  <>
                    <CreditCard size={16} />
                    Confirm Booking
                  </>
                )}
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
```

### Step 2: Commit

```bash
git add frontend/app/checkout/page.tsx
git commit -m "feat: update checkout to require auth and use real booking API"
```

---

## Task 7: Update My Tickets to Fetch from API

**Files:**
- Modify: `frontend/app/my-tickets/page.tsx`

### Step 1: Add API-based ticket fetching

The my-tickets page currently only shows tickets from the in-memory context state. When the user refreshes, those are lost. We need to also load bookings from the API.

Modify `frontend/app/my-tickets/page.tsx` — add `useEffect` to load bookings on mount:

```typescript
"use client";

import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import { CheckCircle, Eye, Calendar, MapPin, Loader2 } from "lucide-react";
import { useBooking } from "@/lib/booking-context";
import { useAuth } from "@/lib/auth-context";
import { apiClient } from "@/lib/api/client";
import { transformBookingToTicket } from "@/lib/api/transformers";
import { PurchasedTicket } from "@/lib/types";

export default function MyTicketsPage() {
  const router = useRouter();
  const { purchasedTickets: contextTickets, getEvent, lastPurchaseSuccess, clearSuccessFlag } = useBooking();
  const { user } = useAuth();
  const [showSuccess, setShowSuccess] = useState(false);
  const [activeTab, setActiveTab] = useState<"Upcoming" | "Past">("Upcoming");
  const [apiTickets, setApiTickets] = useState<PurchasedTicket[]>([]);
  const [isLoading, setIsLoading] = useState(false);

  useEffect(() => {
    if (lastPurchaseSuccess) {
      setShowSuccess(true);
      clearSuccessFlag();
    }
  }, [lastPurchaseSuccess, clearSuccessFlag]);

  // Fetch bookings from API
  useEffect(() => {
    if (!user) return;
    setIsLoading(true);
    apiClient
      .listBookings(1, 50)
      .then((res) => {
        setApiTickets(res.bookings.map((b) => transformBookingToTicket(b)));
      })
      .catch(() => {
        // API unavailable, rely on context tickets
      })
      .finally(() => setIsLoading(false));
  }, [user]);

  if (!user) {
    return (
      <div className="flex flex-col items-center justify-center h-96 gap-4">
        <p className="text-muted text-lg">Please sign in to view your tickets.</p>
        <button
          onClick={() => router.push("/login")}
          className="bg-primary hover:bg-primary-hover text-white font-medium rounded-full px-6 h-12 transition-colors"
        >
          Sign In
        </button>
      </div>
    );
  }

  // Merge: use API tickets, but prefer context tickets for just-purchased ones (have seat details)
  const contextIds = new Set(contextTickets.map((t) => t.id));
  const mergedTickets = [
    ...contextTickets,
    ...apiTickets.filter((t) => !contextIds.has(t.id)),
  ];

  const filtered = mergedTickets.filter((t) =>
    activeTab === "Upcoming" ? t.status === "Upcoming" : t.status === "Confirmed"
  );

  return (
    <div className="flex flex-col gap-6 px-12 py-8">
      {/* Success Banner */}
      {showSuccess && (
        <div className="flex items-start gap-3 bg-success-bg p-4 rounded-lg">
          <CheckCircle size={20} className="text-success shrink-0 mt-0.5" />
          <div className="flex flex-col gap-1">
            <span className="text-success font-medium">Payment Successful!</span>
            <span className="text-success text-sm">
              Your tickets have been confirmed and sent to your email.
            </span>
          </div>
        </div>
      )}

      {/* Header */}
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">My Tickets</h1>
        <div className="flex items-center gap-1 bg-tag-bg rounded-full p-1">
          {(["Upcoming", "Past"] as const).map((tab) => (
            <button
              key={tab}
              onClick={() => setActiveTab(tab)}
              className={`px-4 py-1.5 rounded-full text-sm font-medium transition-colors ${
                activeTab === tab
                  ? "bg-card text-foreground shadow-sm"
                  : "text-muted hover:text-foreground"
              }`}
            >
              {tab}
            </button>
          ))}
        </div>
      </div>

      {/* Loading */}
      {isLoading && (
        <div className="flex justify-center py-8">
          <Loader2 size={24} className="animate-spin text-primary" />
        </div>
      )}

      {/* Ticket List */}
      {!isLoading && filtered.length === 0 && mergedTickets.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-20 gap-4">
          <div className="w-16 h-16 rounded-full bg-tag-bg flex items-center justify-center">
            <Calendar size={24} className="text-muted" />
          </div>
          <p className="text-muted">No tickets yet. Browse events to get started!</p>
        </div>
      ) : !isLoading && filtered.length === 0 ? (
        <p className="text-muted py-8 text-center">No {activeTab.toLowerCase()} tickets.</p>
      ) : (
        <div className="flex flex-col gap-4">
          {filtered.map((ticket) => {
            const event = getEvent(ticket.eventId);
            return (
              <div
                key={ticket.id}
                className="flex items-center bg-card border border-border rounded-lg shadow-sm overflow-hidden"
              >
                {event && (
                  <div className="w-40 h-28 shrink-0">
                    <img
                      src={event.image}
                      alt={event.title}
                      className="w-full h-full object-cover"
                    />
                  </div>
                )}
                <div className="flex-1 flex items-center justify-between p-5">
                  <div className="flex flex-col gap-1.5">
                    <h3 className="font-semibold">{event?.title || "Event"}</h3>
                    {event && (
                      <div className="flex items-center gap-4 text-sm text-muted">
                        <span className="flex items-center gap-1">
                          <Calendar size={12} />
                          {event.date} · {event.time}
                        </span>
                        <span className="flex items-center gap-1">
                          <MapPin size={12} />
                          {event.venue}, {event.location}
                        </span>
                      </div>
                    )}
                    {ticket.seats && ticket.seats.length > 0 ? (
                      <p className="text-xs text-muted">
                        Seats: {ticket.seats.map(s => `${s.section} ${s.row} Seat ${s.seat}`).join(", ")}
                      </p>
                    ) : (
                      <p className="text-xs text-muted">
                        {ticket.tierName} · {ticket.quantity} Ticket{ticket.quantity > 1 ? "s" : ""}
                      </p>
                    )}
                    <span
                      className={`flex items-center gap-1 text-xs font-medium w-fit px-2 py-0.5 rounded-full ${
                        ticket.status === "Confirmed"
                          ? "text-success bg-success-bg"
                          : "text-primary bg-orange-50"
                      }`}
                    >
                      <CheckCircle size={10} />
                      {ticket.status}
                    </span>
                  </div>
                  <div className="flex flex-col items-end gap-3">
                    <span className="text-xl font-bold text-primary">
                      ${ticket.totalPrice.toFixed(2)}
                    </span>
                    <button className="flex items-center gap-1.5 border border-border rounded-lg px-4 py-2 text-sm font-medium hover:bg-tag-bg transition-colors">
                      <Eye size={14} />
                      View Ticket
                    </button>
                  </div>
                </div>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}
```

### Step 2: Commit

```bash
git add frontend/app/my-tickets/page.tsx
git commit -m "feat: load user bookings from API on my-tickets page"
```

---

## Task 8: Backend Startup & End-to-End Verification

**Files:** None (infrastructure only)

### Step 1: Start the backend services

```bash
cd /Users/dev/work/booking/backend
make up
```

Expected: Docker Compose starts all 13 services (4 Postgres, Redis, Zookeeper, Kafka, Kafka-UI, 5 microservices)

### Step 2: Run database migrations

```bash
cd /Users/dev/work/booking/backend
make migrate
```

Expected: Migrations applied to all 4 databases

### Step 3: Verify gateway is healthy

```bash
curl http://localhost:8000/api/events
```

Expected: JSON response with `{"events":[], "total":0, "page":1, "page_size":20}` (empty initially)

### Step 4: Create a test user via API

```bash
curl -X POST http://localhost:8000/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123","name":"Test User"}'
```

Expected: `201` with `access_token`, `refresh_token`, and `user` object

### Step 5: Create a test event via API (as admin — may need to manually set role in DB)

```bash
# First, update the test user to admin in the user database
PGPASSWORD=ticketbox_secret psql -h localhost -p 5433 -U ticketbox -d ticketbox_user -c \
  "UPDATE users SET role='admin' WHERE email='test@example.com';"

# Login again to get token with admin role
TOKEN=$(curl -s -X POST http://localhost:8000/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}' | jq -r '.access_token')

# Create an event
curl -X POST http://localhost:8000/api/events \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "title": "Summer Music Festival 2026",
    "description": "The biggest music event of the year featuring top artists",
    "category": "concerts",
    "venue": "Madison Square Garden",
    "location": "New York, NY",
    "date": "2026-07-15T19:00:00Z",
    "image_url": "https://images.unsplash.com/photo-1459749411175-04bf5292ceea",
    "tiers": [
      {"name": "VIP", "price_cents": 25000, "total_quantity": 100},
      {"name": "Premium", "price_cents": 15000, "total_quantity": 500},
      {"name": "General", "price_cents": 7500, "total_quantity": 2000}
    ]
  }'
```

Expected: `201` with created event details

### Step 6: Start the frontend

```bash
cd /Users/dev/work/booking/frontend
npm run dev
```

### Step 7: Manual verification checklist

Open `http://localhost:3000` in the browser:

1. **Homepage loads** — Should show the event created via API (or mock data if backend is down)
2. **Login** — Navigate to `/login`, sign in with `test@example.com` / `password123`
3. **Navbar updates** — Should show user name and logout button instead of "Sign In"
4. **Event detail** — Click an event, verify tiers display with prices (cents converted to dollars)
5. **Seat selection** — Select seats (still uses client-side venue layout)
6. **Checkout** — Shows user info from auth, "Confirm Booking" button
7. **Booking** — Click confirm, verify it calls backend API (check gateway logs)
8. **My Tickets** — Redirects after purchase, shows the booking
9. **Refresh** — Refresh `/my-tickets`, bookings should persist (loaded from API)
10. **Logout** — Click logout, verify redirects and tokens cleared

### Step 8: Commit any fixes

```bash
git add -A
git commit -m "fix: address integration issues found during e2e verification"
```

---

## Task 9: Seed Data Script (Optional Enhancement)

**Files:**
- Create: `backend/scripts/seed.sh`

### Step 1: Create a seed script for development

Create `backend/scripts/seed.sh`:

```bash
#!/usr/bin/env bash
set -euo pipefail

API_URL="${API_URL:-http://localhost:8000}"

echo "==> Registering admin user..."
REGISTER_RESP=$(curl -s -X POST "$API_URL/api/auth/register" \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@ticketbox.com","password":"admin123","name":"Admin User"}')

echo "$REGISTER_RESP" | jq .

# Promote to admin
echo "==> Promoting to admin..."
PGPASSWORD=ticketbox_secret psql -h localhost -p 5433 -U ticketbox -d ticketbox_user -c \
  "UPDATE users SET role='admin' WHERE email='admin@ticketbox.com';"

# Re-login to get admin token
echo "==> Logging in as admin..."
LOGIN_RESP=$(curl -s -X POST "$API_URL/api/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@ticketbox.com","password":"admin123"}')

TOKEN=$(echo "$LOGIN_RESP" | jq -r '.access_token')

echo "==> Creating events..."

curl -s -X POST "$API_URL/api/events" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "title": "Neon Lights World Tour — Live in NYC",
    "description": "Experience the electrifying Neon Lights World Tour at Madison Square Garden.",
    "category": "concerts",
    "venue": "Madison Square Garden",
    "location": "New York, NY",
    "date": "2026-06-15T20:00:00Z",
    "image_url": "https://images.unsplash.com/photo-1459749411175-04bf5292ceea",
    "tiers": [
      {"name": "Floor", "price_cents": 35000, "total_quantity": 200},
      {"name": "Lower Bowl", "price_cents": 18500, "total_quantity": 1000},
      {"name": "Upper Deck", "price_cents": 8500, "total_quantity": 3000}
    ]
  }' | jq .

curl -s -X POST "$API_URL/api/events" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "title": "Champions League Final 2026",
    "description": "The biggest match in European football. Two titans clash for glory.",
    "category": "sports",
    "venue": "Wembley Stadium",
    "location": "London, UK",
    "date": "2026-05-30T21:00:00Z",
    "image_url": "https://images.unsplash.com/photo-1489944440615-453fc2b6a9a9",
    "tiers": [
      {"name": "Hospitality", "price_cents": 50000, "total_quantity": 100},
      {"name": "Category 1", "price_cents": 25000, "total_quantity": 5000},
      {"name": "Category 2", "price_cents": 15000, "total_quantity": 10000},
      {"name": "Category 3", "price_cents": 7500, "total_quantity": 20000}
    ]
  }' | jq .

curl -s -X POST "$API_URL/api/events" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "title": "Premiere — Starfall: A Space Odyssey",
    "description": "Be among the first to witness the most anticipated sci-fi epic of the decade.",
    "category": "films",
    "venue": "TCL Chinese Theatre",
    "location": "Los Angeles, CA",
    "date": "2026-08-20T19:30:00Z",
    "image_url": "https://images.unsplash.com/photo-1478720568477-152d9b164e26",
    "tiers": [
      {"name": "Red Carpet", "price_cents": 50000, "total_quantity": 50},
      {"name": "Premium", "price_cents": 15000, "total_quantity": 200},
      {"name": "Standard", "price_cents": 5000, "total_quantity": 500}
    ]
  }' | jq .

echo "==> Registering test user..."
curl -s -X POST "$API_URL/api/auth/register" \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"user123","name":"Jane Smith"}' | jq .

echo "==> Done! Seed data created."
echo "    Admin: admin@ticketbox.com / admin123"
echo "    User:  user@example.com / user123"
```

### Step 2: Make executable

```bash
chmod +x backend/scripts/seed.sh
```

### Step 3: Commit

```bash
git add backend/scripts/seed.sh
git commit -m "feat: add seed data script for development"
```

---

## Summary of Changes

| Task | What | Files |
|------|------|-------|
| 1 | API client + Next.js proxy | `lib/api/client.ts`, `lib/api/types.ts`, `next.config.ts`, `.env.local` |
| 2 | Data transformers | `lib/api/transformers.ts` |
| 3 | Auth context + pages | `lib/auth-context.tsx`, `app/login/`, `app/register/`, `layout.tsx`, `navbar.tsx` |
| 4 | BookingProvider → API | `lib/booking-context.tsx` |
| 5 | Homepage loading state | `app/page.tsx` |
| 6 | Checkout auth guard | `app/checkout/page.tsx` |
| 7 | My Tickets from API | `app/my-tickets/page.tsx` |
| 8 | E2E verification | Infrastructure only |
| 9 | Seed data script | `backend/scripts/seed.sh` |

**Architecture decisions:**
- Mock data kept as fallback when backend is offline (graceful degradation)
- Next.js rewrites proxy `/api/*` to backend gateway (avoids CORS in production)
- JWT tokens stored in `localStorage` (simple; consider `httpOnly` cookies for production)
- Venue layout / seat maps remain client-side (not yet in backend schema)
- Auth state in separate context from booking state (separation of concerns)
