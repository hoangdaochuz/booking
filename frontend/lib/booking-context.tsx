"use client";

import React, { createContext, useContext, useState, useCallback } from "react";
import { events as initialEvents } from "./mock-data";
import { Event, CartItem, SelectedSeat, PurchasedTicket } from "./types";

interface BookingState {
  events: Event[];
  cart: CartItem | null;
  purchasedTickets: PurchasedTicket[];
  lastPurchaseSuccess: boolean;
  setCart: (item: CartItem | null) => void;
  purchaseTickets: (name: string, email: string) => Promise<boolean>;
  clearSuccessFlag: () => void;
  getEvent: (id: string) => Event | undefined;
}

const BookingContext = createContext<BookingState | null>(null);

export function BookingProvider({ children }: { children: React.ReactNode }) {
  const [events, setEvents] = useState<Event[]>(
    () => JSON.parse(JSON.stringify(initialEvents)) // deep clone so mutations don't affect source
  );
  const [cart, setCart] = useState<CartItem | null>(null);
  const [purchasedTickets, setPurchasedTickets] = useState<PurchasedTicket[]>([]);
  const [lastPurchaseSuccess, setLastPurchaseSuccess] = useState(false);

  const clearSuccessFlag = useCallback(() => setLastPurchaseSuccess(false), []);

  const getEvent = useCallback(
    (id: string) => events.find((e) => e.id === id),
    [events]
  );

  const purchaseTickets = useCallback(
    async (_name: string, _email: string): Promise<boolean> => {
      if (!cart || cart.seats.length === 0) return false;

      const event = events.find((e) => e.id === cart.eventId);
      if (!event || !event.venueLayout) return false;

      // *** DOUBLE BOOKING VULNERABILITY ***
      // We read the seat statuses now, then simulate a network delay,
      // then mark them taken. Two concurrent requests can both read the same
      // statuses before either marks them — classic race condition.
      const seatStatuses = new Map<string, string>();
      for (const selectedSeat of cart.seats) {
        const section = event.venueLayout.sections.find(
          (s) => s.id === selectedSeat.sectionId
        );
        if (section) {
          for (const row of section.rows) {
            const seat = row.seats.find((s) => s.id === selectedSeat.seatId);
            if (seat) {
              seatStatuses.set(selectedSeat.seatId, seat.status);
            }
          }
        }
      }

      // Simulate network/processing delay (makes race condition easy to trigger)
      await new Promise((resolve) => setTimeout(resolve, 2000));

      // Check if any seats are taken (using stale data — this is the bug)
      for (const selectedSeat of cart.seats) {
        const status = seatStatuses.get(selectedSeat.seatId);
        if (status === "taken") {
          return false; // seat already taken
        }
      }

      // Mark each selected seat as "taken" (using stale read — this is the bug)
      setEvents((prev) =>
        prev.map((e) =>
          e.id === cart.eventId && e.venueLayout
            ? {
                ...e,
                venueLayout: {
                  ...e.venueLayout,
                  sections: e.venueLayout.sections.map((section) => ({
                    ...section,
                    rows: section.rows.map((row) => ({
                      ...row,
                      seats: row.seats.map((seat) => {
                        const isSelected = cart.seats.some(
                          (s) => s.seatId === seat.id
                        );
                        return isSelected
                          ? { ...seat, status: "taken" as const }
                          : seat;
                      }),
                    })),
                    availableSeats:
                      section.availableSeats -
                      cart.seats.filter((s) => s.sectionId === section.id)
                        .length,
                  })),
                },
              }
            : e
        )
      );

      const firstSeat = cart.seats[0];
      const totalPrice = cart.seats.reduce((sum, s) => sum + s.price, 0);

      const newTicket: PurchasedTicket = {
        id: `ticket-${Date.now()}-${Math.random().toString(36).slice(2)}`,
        eventId: cart.eventId,
        tierName: firstSeat.sectionName,
        quantity: cart.seats.length,
        totalPrice,
        purchasedAt: new Date().toISOString(),
        status: new Date(event.date) > new Date() ? "Upcoming" : "Confirmed",
        seats: cart.seats.map((s) => ({
          section: s.sectionName,
          row: s.rowLabel,
          seat: s.seatLabel,
        })),
      };

      setPurchasedTickets((prev) => [newTicket, ...prev]);
      setCart(null);
      setLastPurchaseSuccess(true);
      return true;
    },
    [cart, events]
  );

  return (
    <BookingContext.Provider
      value={{ events, cart, purchasedTickets, lastPurchaseSuccess, setCart, purchaseTickets, clearSuccessFlag, getEvent }}
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
