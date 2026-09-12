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
