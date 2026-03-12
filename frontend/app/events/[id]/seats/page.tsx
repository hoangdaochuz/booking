"use client";

import { useState, useMemo, useEffect } from "react";
import { useParams, useRouter } from "next/navigation";
import { Check, ChevronRight, X, Armchair, Loader2 } from "lucide-react";
import { useBooking } from "@/lib/booking-context";
import { VenueMap } from "@/components/venue-map";
import SectionPicker from "@/components/section-picker";
import { apiClient } from "@/lib/api/client";
import { transformSeats, generateVenueLayoutFromSeats } from "@/lib/api/transformers";
import type { Seat, SeatRow, VenueSection, SelectedSeat, VenueLayout, Event } from "@/lib/types";

export default function SeatsPage() {
  const params = useParams();
  const router = useRouter();
  const { getEvent, setCart, isLoading } = useBooking();

  const eventId = params.id as string;
  const event = getEvent(eventId);

  const [selectedSectionId, setSelectedSectionId] = useState<string | null>(null);
  const [selectedSeats, setSelectedSeats] = useState<SelectedSeat[]>([]);
  const [venueLayout, setVenueLayout] = useState<VenueLayout | null>(null);
  const [isLoadingSeats, setIsLoadingSeats] = useState(true);

  // Fetch seats for this event
  useEffect(() => {
    async function fetchSeats() {
      if (!event) return;

      try {
        setIsLoadingSeats(true);
        const seatsData = await apiClient.getSeats({ event_id: eventId });
        console.log("Fetched seats count:", seatsData.seats?.length || 0);

        if (!seatsData.seats || seatsData.seats.length === 0) {
          console.warn("No seats returned from API, falling back to mock layout");
          setVenueLayout(event.venueLayout || null);
          return;
        }

        const transformedSeats = transformSeats(seatsData.seats);

        // Generate venue layout from actual seat data
        const layout = generateVenueLayoutFromSeats(
          seatsData.seats,
          event.tiers.map(t => ({
            id: t.id,
            event_id: eventId,
            name: t.name,
            price_cents: Math.round(t.price * 100),
            total_quantity: t.available, // Use available as total for display
            available_quantity: t.available,
            version: 0,
          })),
          event.category.toLowerCase()
        );

        console.log("Generated layout with sections:", layout.sections.map(s => ({ id: s.id, name: s.name, seats: s.totalSeats })));
        setVenueLayout(layout);
      } catch (error) {
        console.error("Failed to fetch seats:", error);
        // Fallback to event's venue layout
        setVenueLayout(event.venueLayout || null);
      } finally {
        setIsLoadingSeats(false);
      }
    }

    fetchSeats();
  }, [event, eventId]);

  const selectedSection = venueLayout?.sections.find((s) => s.id === selectedSectionId);

  const totalPrice = useMemo(
    () => selectedSeats.reduce((sum, s) => sum + s.price, 0),
    [selectedSeats]
  );

  // Build legend from unique tier + price combos
  const legend = useMemo(() => {
    if (!venueLayout) return [];
    const seen = new Map<string, { tier: string; price: number; color: string }>();
    for (const section of venueLayout.sections) {
      const key = `${section.tier}-${section.price}`;
      if (!seen.has(key)) {
        seen.set(key, { tier: section.tier, price: section.price, color: section.color });
      }
    }
    return Array.from(seen.values());
  }, [venueLayout]);

  if (isLoading || isLoadingSeats) {
    return (
      <div className="flex items-center justify-center h-96">
        <Loader2 size={32} className="animate-spin text-primary" />
      </div>
    );
  }

  if (!event || !venueLayout) {
    return (
      <div className="flex items-center justify-center h-96">
        <p className="text-muted text-lg">Event not found.</p>
      </div>
    );
  }

  function handleSeatToggle(seat: Seat, row: SeatRow, section: VenueSection) {
    console.log("[seats page] Seat clicked:", {
      seatId: seat.id,
      seatLabel: seat.label,
      sectionId: section.id,
      sectionName: section.name,
      tierId: section.tier,
    });
    setSelectedSeats((prev) => {
      const exists = prev.find((s) => s.seatId === seat.id);
      if (exists) {
        return prev.filter((s) => s.seatId !== seat.id);
      }
      if (seat.status !== "available") return prev;

      // Find the tier ID by matching section tier name with event tiers
      const tier = event?.tiers.find(t => t.name === section.tier);
      const tierId = tier?.id || "";

      const newSeat = {
        sectionId: section.id,
        sectionName: section.name,
        rowLabel: row.label,
        seatLabel: seat.label,
        seatId: seat.id,
        price: section.price,
        tierId: tierId,
      };
      console.log("[seats page] Adding to cart:", newSeat);

      return [
        ...prev,
        newSeat,
      ];
    });
  }

  function handleContinue() {
    setCart({ eventId, seats: selectedSeats });
    router.push("/checkout");
  }

  const steps = [
    { label: "Select", active: true, done: false },
    { label: "Checkout", active: false, done: false },
    { label: "Confirm", active: false, done: false },
  ];

  return (
    <div className="flex flex-col min-h-screen">
      {/* Event Summary Bar */}
      <div className="flex items-center justify-between h-14 px-12 bg-card border-b border-border">
        <div className="flex items-center gap-4">
          <span className="text-sm font-semibold">{event.title}</span>
          <span className="text-sm text-muted">
            {event.date} &middot; {event.venue}
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

      {/* Main Content */}
      <div className="flex gap-8 px-12 py-8 flex-1">
        {/* Left: Map or Section Picker */}
        <div className="flex-1 flex flex-col gap-6">
          {selectedSectionId && selectedSection ? (
            <SectionPicker
              section={selectedSection}
              selectedSeats={selectedSeats}
              onSeatToggle={handleSeatToggle}
              onBack={() => setSelectedSectionId(null)}
            />
          ) : (
            <>
              <VenueMap
                layout={venueLayout}
                onSectionClick={(id) => setSelectedSectionId(id)}
                selectedSectionId={selectedSectionId ?? undefined}
              />

              {/* Legend */}
              <div className="flex items-center gap-6 flex-wrap">
                {legend.map((item) => (
                  <div key={`${item.tier}-${item.price}`} className="flex items-center gap-2">
                    <div
                      className="w-3 h-3 rounded-sm"
                      style={{ backgroundColor: item.color }}
                    />
                    <span className="text-xs text-muted">
                      {item.tier} &mdash; ${item.price.toFixed(2)}
                    </span>
                  </div>
                ))}
                <div className="flex items-center gap-2">
                  <div className="w-3 h-3 rounded-sm" style={{ backgroundColor: "#D1D5DB" }} />
                  <span className="text-xs text-muted">Unavailable</span>
                </div>
              </div>
            </>
          )}
        </div>

        {/* Right Sidebar */}
        <div className="w-[360px] shrink-0">
          <div className="bg-card border border-border rounded-lg shadow-sm flex flex-col sticky top-8">
            <div className="p-6 border-b border-border">
              <h3 className="font-bold text-lg">Your Seats</h3>
            </div>

            <div className="p-6 flex flex-col gap-4">
              {selectedSeats.length === 0 ? (
                <div className="flex flex-col items-center gap-3 py-8 text-center">
                  <Armchair size={32} className="text-muted opacity-40" />
                  <p className="text-sm text-muted">Select seats from the venue map</p>
                </div>
              ) : (
                <>
                  <div className="flex flex-col gap-2 max-h-72 overflow-y-auto">
                    {selectedSeats.map((seat) => (
                      <div
                        key={seat.seatId}
                        className="flex items-center justify-between bg-background rounded-lg px-3 py-2"
                      >
                        <div className="flex flex-col">
                          <span className="text-sm font-medium">
                            {seat.sectionName}
                          </span>
                          <span className="text-xs text-muted">
                            Row {seat.rowLabel}, Seat {seat.seatLabel}
                          </span>
                        </div>
                        <div className="flex items-center gap-3">
                          <span className="text-sm font-semibold text-primary">
                            ${seat.price.toFixed(2)}
                          </span>
                          <button
                            onClick={() =>
                              setSelectedSeats((prev) =>
                                prev.filter((s) => s.seatId !== seat.seatId)
                              )
                            }
                            className="text-muted hover:text-foreground transition-colors"
                          >
                            <X size={14} />
                          </button>
                        </div>
                      </div>
                    ))}
                  </div>

                  <div className="flex items-center justify-between pt-4 border-t border-border">
                    <span className="text-sm font-medium">Total</span>
                    <span className="text-xl font-bold text-primary">
                      ${totalPrice.toFixed(2)}
                    </span>
                  </div>
                </>
              )}

              <button
                onClick={handleContinue}
                disabled={selectedSeats.length === 0}
                className="flex items-center justify-center gap-2 bg-primary hover:bg-primary-hover disabled:opacity-40 disabled:cursor-not-allowed text-white font-medium rounded-full h-12 transition-colors"
              >
                Continue to Checkout
                <ChevronRight size={16} />
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
