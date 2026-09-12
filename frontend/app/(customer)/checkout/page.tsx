"use client";

import { useRouter } from "next/navigation";
import Link from "next/link";
import { Check, Lock } from "lucide-react";
import { useBooking } from "@/lib/booking-context";
import { useAuth } from "@/lib/auth-context";
import { StripePaymentForm } from "@/components/stripe-payment-form";

export default function CheckoutPage() {
  const router = useRouter();
  const {
    cart,
    getEvent,
    confirmPayment,
    paymentClientSecret,
  } = useBooking();
  const { user } = useAuth();

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

  function handlePaymentSuccess(paymentIntentId: string) {
    confirmPayment(paymentIntentId).then((success) => {
      if (success) {
        router.push("/my-tickets");
      }
    });
  }

  function handlePaymentError(message: string) {
    console.error("Payment error:", message);
  }

  const steps = [
    { label: "Select", done: true },
    { label: "Checkout", active: true },
    { label: "Confirm", done: false },
  ];

  // Payment not initialized - redirect back to seats
  if (!paymentClientSecret) {
    return (
      <div className="flex items-center justify-center h-96">
        <div className="flex flex-col items-center gap-4">
          <p className="text-muted text-lg">Payment session expired or not initialized.</p>
          <Link
            href={`/events/${cart.eventId}/seats`}
            className="bg-primary hover:bg-primary-hover text-white font-medium rounded-full px-6 h-12 flex items-center transition-colors"
          >
            Back to Seat Selection
          </Link>
        </div>
      </div>
    );
  }

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
        {/* Booking Info & Payment */}
        <div className="flex-1">
          <div className="bg-card border border-border rounded-lg p-8 shadow-sm flex flex-col gap-6">
            <h2 className="text-lg font-bold">Payment Details</h2>

            {/* User Info */}
            <div className="flex flex-col gap-4 pb-6 border-b border-border">
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

            {/* Payment Form */}
            <div className="flex flex-col gap-4">
              <div className="flex items-center gap-2 text-sm text-muted">
                <Lock size={14} />
                <span>Secure payment powered by Stripe</span>
              </div>
              <StripePaymentForm
                clientSecret={paymentClientSecret}
                amount={total}
                onSuccess={handlePaymentSuccess}
                onError={handlePaymentError}
              />
            </div>
          </div>
        </div>

        {/* Order Summary */}
        <div className="w-100 shrink-0">
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
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
