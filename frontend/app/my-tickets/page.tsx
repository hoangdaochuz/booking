"use client";

import { useState, useEffect } from "react";
import { CheckCircle, Eye, Calendar, MapPin } from "lucide-react";
import { useBooking } from "@/lib/booking-context";

export default function MyTicketsPage() {
  const { purchasedTickets, getEvent, lastPurchaseSuccess, clearSuccessFlag } = useBooking();
  const [showSuccess, setShowSuccess] = useState(false);
  const [activeTab, setActiveTab] = useState<"Upcoming" | "Past">("Upcoming");

  useEffect(() => {
    if (lastPurchaseSuccess) {
      setShowSuccess(true);
      clearSuccessFlag();
    }
  }, [lastPurchaseSuccess, clearSuccessFlag]);

  const filtered = purchasedTickets.filter((t) =>
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

      {/* Ticket List */}
      {filtered.length === 0 && purchasedTickets.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-20 gap-4">
          <div className="w-16 h-16 rounded-full bg-tag-bg flex items-center justify-center">
            <Calendar size={24} className="text-muted" />
          </div>
          <p className="text-muted">No tickets yet. Browse events to get started!</p>
        </div>
      ) : filtered.length === 0 ? (
        <p className="text-muted py-8 text-center">No {activeTab.toLowerCase()} tickets.</p>
      ) : (
        <div className="flex flex-col gap-4">
          {filtered.map((ticket) => {
            const event = getEvent(ticket.eventId);
            if (!event) return null;
            return (
              <div
                key={ticket.id}
                className="flex items-center bg-card border border-border rounded-lg shadow-sm overflow-hidden"
              >
                <div className="w-40 h-28 shrink-0">
                  <img
                    src={event.image}
                    alt={event.title}
                    className="w-full h-full object-cover"
                  />
                </div>
                <div className="flex-1 flex items-center justify-between p-5">
                  <div className="flex flex-col gap-1.5">
                    <h3 className="font-semibold">{event.title}</h3>
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
