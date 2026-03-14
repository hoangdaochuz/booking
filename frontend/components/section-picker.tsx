"use client";

import { ArrowLeft } from "lucide-react";
import SeatLegend from "@/components/seat-legend";
import type { VenueSection, SeatRow, Seat, SelectedSeat } from "@/lib/types";

interface SectionPickerProps {
  section: VenueSection;
  selectedSeats: SelectedSeat[];
  onSeatToggle: (seat: Seat, row: SeatRow, section: VenueSection) => void;
  onBack: () => void;
}

export default function SectionPicker({
  section,
  selectedSeats,
  onSeatToggle,
  onBack,
}: SectionPickerProps) {
  const isSeatSelected = (seat: Seat) =>
    selectedSeats.some((s) => s.seatId === seat.id);

  return (
    <div className="flex flex-col gap-6 p-6">
      {/* Header */}
      <div className="flex flex-col gap-3">
        <button
          onClick={onBack}
          className="flex items-center gap-1.5 text-sm text-muted hover:text-foreground transition-colors cursor-pointer self-start"
        >
          <ArrowLeft className="w-4 h-4" />
          Back to venue map
        </button>
        <div className="flex items-center gap-3">
          <h2 className="text-xl font-semibold">{section.name}</h2>
          <span
            className="rounded-full px-3 py-0.5 text-sm font-medium"
            style={{
              backgroundColor: `${section.color}30`,
              color: section.color,
            }}
          >
            ${section.price}
          </span>
        </div>
      </div>

      {/* Seat Grid */}
      <div className="flex flex-col items-center gap-2">
        {section.rows.map((row) => (
          <div key={row.id} className="flex items-center gap-3">
            <span className="w-14 text-right text-xs text-muted font-medium shrink-0">
              Row {row.label}
            </span>
            <div className="flex items-center gap-1.5">
              {row.seats.map((seat) => {
                const selected = isSeatSelected(seat);
                const isBooked = seat.status === "booked";
                const isReserved = seat.status === "reserved";
                const isDisabled = isBooked;

                let backgroundColor: string;
                let color: string;
                let borderColor: string | undefined;
                let cursor: string;

                if (selected) {
                  backgroundColor = "#FF8400";
                  color = "#FFFFFF";
                  cursor = "pointer";
                } else if (isBooked) {
                  backgroundColor = "#D1D5DB";
                  color = "#9CA3AF";
                  cursor = "not-allowed";
                } else if (isReserved) {
                  backgroundColor = "#FEF3C7";
                  color = "#D97706";
                  borderColor = "#F59E0B";
                  cursor = "not-allowed";
                } else {
                  backgroundColor = `${section.color}4D`; // 30% opacity
                  color = section.color;
                  borderColor = section.color;
                  cursor = "pointer";
                }

                return (
                  <button
                    key={seat.id}
                    disabled={isDisabled || isReserved}
                    onClick={() => {
                      if (!isDisabled && !isReserved) {
                        onSeatToggle(seat, row, section);
                      }
                    }}
                    className="flex items-center justify-center rounded-lg text-xs font-medium transition-all hover:brightness-90 disabled:hover:brightness-100"
                    style={{
                      width: 32,
                      height: 32,
                      backgroundColor,
                      color,
                      border: borderColor
                        ? `2px solid ${borderColor}`
                        : "2px solid transparent",
                      cursor,
                    }}
                  >
                    {seat.label}
                  </button>
                );
              })}
            </div>
          </div>
        ))}
      </div>

      {/* Legend */}
      <div className="flex justify-center pt-2">
        <SeatLegend
          items={[
            { label: "Available", color: `${section.color}4D`, border: section.color },
            { label: "Reserved", color: "#FEF3C7", border: "#F59E0B" },
            { label: "Booked", color: "#D1D5DB" },
            { label: "Selected", color: "#FF8400" },
          ]}
        />
      </div>
    </div>
  );
}
