"use client";

import { useParams, useRouter } from "next/navigation";
import Link from "next/link";
import { Calendar, MapPin, Music, Ticket, ChevronRight } from "lucide-react";
import { useBooking } from "@/lib/booking-context";

export default function EventDetailPage() {
  const { id } = useParams<{ id: string }>();
  const router = useRouter();
  const { getEvent } = useBooking();
  const event = getEvent(id);

  if (!event) {
    return (
      <div className="flex items-center justify-center h-96">
        <p className="text-muted text-lg">Event not found.</p>
      </div>
    );
  }

  return (
    <div className="flex flex-col">
      {/* Banner */}
      <div className="relative h-[280px] overflow-hidden">
        <img src={event.image} alt={event.title} className="w-full h-full object-cover" />
        <div className="absolute inset-0 bg-gradient-to-t from-black/75 to-transparent" />
        <div className="absolute inset-0 flex flex-col justify-end p-12 gap-2">
          <div className="flex items-center gap-2 text-white/70 text-sm">
            <Link href="/" className="hover:text-white">Home</Link>
            <ChevronRight size={14} />
            <Link href={`/?category=${event.category}`} className="hover:text-white">{event.category}</Link>
            <ChevronRight size={14} />
            <span className="text-white">{event.title.split("—")[0].trim()}</span>
          </div>
          <h1 className="text-white text-3xl font-bold leading-tight">{event.title}</h1>
          <div className="flex items-center gap-4 text-white/80 text-sm">
            <span>{event.date} · {event.time}</span>
            <span>{event.venue}, {event.location}</span>
          </div>
        </div>
      </div>

      {/* Content */}
      <div className="flex gap-8 px-12 py-8">
        {/* Left — About */}
        <div className="flex-1 flex flex-col gap-6">
          <div className="bg-card border border-border rounded-lg p-8 shadow-sm flex flex-col gap-6">
            <h2 className="text-lg font-bold">About This Event</h2>
            <p className="text-muted text-sm leading-relaxed">{event.description}</p>
            <div className="grid grid-cols-3 gap-4 pt-4 border-t border-border">
              <div className="flex flex-col gap-1">
                <div className="flex items-center gap-1.5 text-muted text-xs font-medium">
                  <Calendar size={12} /> Date & Time
                </div>
                <p className="text-sm font-medium">{event.date} · {event.time}</p>
              </div>
              <div className="flex flex-col gap-1">
                <div className="flex items-center gap-1.5 text-muted text-xs font-medium">
                  <MapPin size={12} /> Venue
                </div>
                <p className="text-sm font-medium">{event.venue}, {event.location}</p>
              </div>
              <div className="flex flex-col gap-1">
                <div className="flex items-center gap-1.5 text-muted text-xs font-medium">
                  <Music size={12} /> Genre
                </div>
                <p className="text-sm font-medium">{event.genre}</p>
              </div>
            </div>
          </div>
        </div>

        {/* Right — Ticket Selection */}
        <div className="w-[400px] shrink-0">
          <div className="bg-card border border-border rounded-lg shadow-sm flex flex-col">
            <div className="flex items-center justify-between p-6 border-b border-border">
              <h2 className="text-lg font-bold">Ticket Tiers</h2>
              <span className="text-xs px-2 py-1 rounded-full bg-green-100 text-green-700 font-medium">
                Available
              </span>
            </div>

            <div className="flex flex-col p-6 gap-4">
              {event.tiers.map((tier) => (
                <div
                  key={tier.id}
                  className="flex items-center justify-between p-4 rounded-lg border border-border text-left"
                >
                  <div className="flex flex-col gap-0.5">
                    <span className="font-semibold text-sm">{tier.name}</span>
                    <span className="text-xs text-muted">{tier.description}</span>
                    <span className="text-xs text-muted mt-1">
                      {tier.available} tickets left
                    </span>
                  </div>
                  <span className="text-lg font-bold text-primary">${tier.price}</span>
                </div>
              ))}

              <button
                onClick={() => router.push(`/events/${event.id}/seats`)}
                className="flex items-center justify-center gap-2 bg-primary hover:bg-primary-hover text-white font-medium rounded-full h-12 transition-colors mt-2"
              >
                <Ticket size={16} />
                Select Seats
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
