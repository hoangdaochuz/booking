"use client";

import Link from "next/link";
import { Calendar, Plus } from "lucide-react";
import { Event } from "@/lib/types";

function getCategoryColor(category: string) {
  switch (category) {
    case "Concerts":
      return "bg-primary text-white";
    case "Sports":
      return "bg-green-600 text-white";
    case "Films":
      return "bg-violet-600 text-white";
    default:
      return "bg-tag-bg text-foreground";
  }
}

export function EventCard({ event }: { event: Event }) {
  const lowestPrice = Math.min(...event.tiers.map((t) => t.price));

  return (
    <Link href={`/events/${event.id}`} className="group flex-1 min-w-0">
      <div className="bg-background border border-border rounded-lg overflow-hidden shadow-sm hover:shadow-md transition-shadow">
        <div className="relative h-48 overflow-hidden">
          <img
            src={event.image}
            alt={event.title}
            className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300"
          />
        </div>
        <div className="p-4 flex flex-col gap-3">
          <h3 className="font-semibold text-sm leading-tight line-clamp-2">
            {event.title}
          </h3>
          <div className="flex items-center gap-1 text-muted text-xs">
            <Calendar size={12} />
            <span>
              {event.date} · {event.venue}, {event.location}
            </span>
          </div>
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-1">
              <Plus size={12} className="text-primary" />
              <span className="text-sm font-semibold">From ${lowestPrice}</span>
            </div>
            <span
              className={`text-xs px-2 py-0.5 rounded-full font-medium ${getCategoryColor(event.category)}`}
            >
              {event.category}
            </span>
          </div>
        </div>
      </div>
    </Link>
  );
}
