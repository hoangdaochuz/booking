"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { CreditCard, Check, Loader2 } from "lucide-react";
import { useBooking } from "@/lib/booking-context";

export default function CheckoutPage() {
  const router = useRouter();
  const { cart, getEvent, purchaseTickets } = useBooking();
  const [isProcessing, setIsProcessing] = useState(false);
  const [form, setForm] = useState({
    firstName: "",
    lastName: "",
    email: "",
    cardNumber: "",
    expiry: "",
    cvv: "",
  });

  if (!cart) {
    return (
      <div className="flex items-center justify-center h-96">
        <p className="text-muted text-lg">No items in cart. Please select tickets first.</p>
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
    const success = await purchaseTickets(form.firstName, form.email);
    setIsProcessing(false);
    if (success) {
      router.push("/my-tickets");
    } else {
      alert("Purchase failed — tickets may have sold out. This is the double booking problem!");
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
        {/* Payment Form */}
        <div className="flex-1">
          <div className="bg-card border border-border rounded-lg p-8 shadow-sm flex flex-col gap-6">
            <h2 className="text-lg font-bold">Payment Details</h2>

            <div className="grid grid-cols-2 gap-4">
              <div className="flex flex-col gap-1.5">
                <label className="text-xs font-medium text-muted">First Name</label>
                <input
                  type="text"
                  placeholder="John"
                  value={form.firstName}
                  onChange={(e) => setForm({ ...form, firstName: e.target.value })}
                  className="border border-border rounded-lg px-3 py-2.5 text-sm bg-card focus:outline-none focus:border-primary transition-colors"
                />
              </div>
              <div className="flex flex-col gap-1.5">
                <label className="text-xs font-medium text-muted">Last Name</label>
                <input
                  type="text"
                  placeholder="Doe"
                  value={form.lastName}
                  onChange={(e) => setForm({ ...form, lastName: e.target.value })}
                  className="border border-border rounded-lg px-3 py-2.5 text-sm bg-card focus:outline-none focus:border-primary transition-colors"
                />
              </div>
            </div>

            <div className="flex flex-col gap-1.5">
              <label className="text-xs font-medium text-muted">Email Address</label>
              <input
                type="email"
                placeholder="john.doe@email.com"
                value={form.email}
                onChange={(e) => setForm({ ...form, email: e.target.value })}
                className="border border-border rounded-lg px-3 py-2.5 text-sm bg-card focus:outline-none focus:border-primary transition-colors"
              />
            </div>

            <div className="flex flex-col gap-1.5">
              <label className="text-xs font-medium text-muted">Card Number</label>
              <input
                type="text"
                placeholder="4242 4242 4242 4242"
                value={form.cardNumber}
                onChange={(e) => setForm({ ...form, cardNumber: e.target.value })}
                className="border border-border rounded-lg px-3 py-2.5 text-sm bg-card focus:outline-none focus:border-primary transition-colors"
              />
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div className="flex flex-col gap-1.5">
                <label className="text-xs font-medium text-muted">Expiry Date</label>
                <input
                  type="text"
                  placeholder="12/26"
                  value={form.expiry}
                  onChange={(e) => setForm({ ...form, expiry: e.target.value })}
                  className="border border-border rounded-lg px-3 py-2.5 text-sm bg-card focus:outline-none focus:border-primary transition-colors"
                />
              </div>
              <div className="flex flex-col gap-1.5">
                <label className="text-xs font-medium text-muted">CVV</label>
                <input
                  type="text"
                  placeholder="•••"
                  value={form.cvv}
                  onChange={(e) => setForm({ ...form, cvv: e.target.value })}
                  className="border border-border rounded-lg px-3 py-2.5 text-sm bg-card focus:outline-none focus:border-primary transition-colors"
                />
              </div>
            </div>
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
                    Pay Now
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
