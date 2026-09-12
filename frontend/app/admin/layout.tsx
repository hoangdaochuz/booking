"use client";

/**
 * Admin shell — Ops sidebar + auth guard.
 *
 * Unauthenticated visitors are bounced to /login with a ?redirect= back
 * here; authenticated non-admins get a 403 panel. The customer Navbar
 * lives in the (customer) route group, so it never renders here.
 */

import { useEffect } from "react";
import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import {
  Flag,
  LayoutGrid,
  Loader2,
  Send,
  Timer,
} from "lucide-react";
import { useAuth } from "@/lib/auth-context";

const NAV_ITEMS = [
  { href: "/admin", label: "Overview", icon: LayoutGrid, soon: false },
  { href: "/admin/ops", label: "Cron Jobs", icon: Timer, soon: false },
  { href: "/admin/outbox", label: "Outbox Events", icon: Send, soon: true },
  { href: "/admin/flags", label: "Feature Flags", icon: Flag, soon: true },
];

export default function AdminLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const { user, isLoading } = useAuth();
  const router = useRouter();
  const pathname = usePathname();

  useEffect(() => {
    if (!isLoading && !user) {
      router.replace(`/login?redirect=${encodeURIComponent(pathname)}`);
    }
  }, [isLoading, user, router, pathname]);

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-screen">
        <Loader2 size={32} className="animate-spin text-primary" />
      </div>
    );
  }

  if (!user) {
    return (
      <div className="flex items-center justify-center h-screen">
        <Loader2 size={32} className="animate-spin text-primary" />
      </div>
    );
  }

  if (user.role !== "admin") {
    return (
      <div className="flex flex-col items-center justify-center h-screen gap-3 text-center px-6">
        <h1 className="text-2xl font-bold">403 — Admins only</h1>
        <p className="text-muted text-sm max-w-sm">
          You are signed in as <strong>{user.email}</strong> ({user.role}).
          This console requires an admin account.
        </p>
        <Link
          href="/"
          className="bg-primary hover:bg-primary-hover text-white font-medium rounded-full px-6 h-11 flex items-center text-sm transition-colors mt-2"
        >
          Back to TicketBox
        </Link>
      </div>
    );
  }

  return (
    <div className="flex min-h-screen">
      {/* Sidebar */}
      <aside className="w-60 bg-card border-r border-border flex flex-col shrink-0 sticky top-0 h-screen">
        <div className="flex items-center gap-2.5 px-5 pt-5 pb-4">
          <div className="w-7 h-7 bg-primary text-white rounded-lg grid place-items-center font-mono font-bold text-xs">
            TB
          </div>
          <div>
            <div className="font-mono font-bold tracking-tight text-[15px]">
              TICKETBOX
            </div>
            <div className="text-[11px] text-muted">Internal Ops Console</div>
          </div>
        </div>

        <nav className="px-3 py-2 flex-1">
          <div className="text-[11px] font-semibold text-muted uppercase tracking-wider px-2 pt-3 pb-1.5">
            Operations
          </div>
          {NAV_ITEMS.map((item) => {
            const active =
              item.href === "/admin"
                ? pathname === "/admin"
                : pathname.startsWith(item.href);
            return (
              <Link
                key={item.href}
                href={item.soon ? "/admin/ops" : item.href}
                aria-disabled={item.soon}
                className={`flex items-center gap-2.5 px-2.5 py-2 rounded-lg text-[13px] font-medium transition-colors ${
                  active && !item.soon
                    ? "bg-tag-bg text-foreground font-semibold"
                    : "text-muted hover:bg-tag-bg hover:text-foreground"
                } ${item.soon ? "opacity-60" : ""}`}
              >
                <item.icon size={16} className="shrink-0" />
                {item.label}
                {item.soon && (
                  <span className="ml-auto text-[10px] bg-tag-bg rounded-full px-1.5 py-0.5">
                    soon
                  </span>
                )}
              </Link>
            );
          })}
        </nav>

        <div className="px-5 py-4 border-t border-border text-xs text-muted flex items-center gap-2">
          <span className="w-2 h-2 rounded-full bg-success shrink-0" />
          scheduler-service · connected
        </div>
      </aside>

      {/* Content — admin pages render their own topbar */}
      <div className="flex-1 min-w-0 flex flex-col bg-background">
        {children}
      </div>
    </div>
  );
}
