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
