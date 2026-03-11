"use client";

import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { Search, User, LogOut } from "lucide-react";
import { useAuth } from "@/lib/auth-context";

const navLinks = [
  { href: "/", label: "Discover" },
  { href: "/?category=Concerts", label: "Concerts" },
  { href: "/?category=Sports", label: "Sports" },
  { href: "/?category=Films", label: "Films" },
];

export function Navbar() {
  const pathname = usePathname();
  const router = useRouter();
  const { user, isLoading, logout } = useAuth();

  async function handleLogout() {
    await logout();
    router.push("/");
  }

  return (
    <nav className="flex items-center justify-between h-16 px-12 bg-card border-b border-border w-full shrink-0">
      <div className="flex items-center gap-8">
        <Link href="/" className="text-primary font-bold text-xl font-mono tracking-tight">
          TICKETBOX
        </Link>
        <div className="flex items-center gap-6">
          {navLinks.map((link) => (
            <Link
              key={link.label}
              href={link.href}
              className={`text-sm font-medium transition-colors hover:text-primary ${
                pathname === link.href ? "text-primary" : "text-muted"
              }`}
            >
              {link.label}
            </Link>
          ))}
        </div>
      </div>
      <div className="flex items-center gap-4">
        <div className="flex items-center gap-2 border border-border rounded-sm px-3 py-1.5 w-48">
          <Search size={14} className="text-muted" />
          <span className="text-sm text-muted">Search...</span>
        </div>
        {isLoading ? null : user ? (
          <div className="flex items-center gap-3">
            <Link
              href="/my-tickets"
              className="text-sm font-medium text-muted hover:text-primary transition-colors"
            >
              My Tickets
            </Link>
            <span className="text-sm font-medium">{user.name}</span>
            <button
              onClick={handleLogout}
              className="flex items-center justify-center w-10 h-10 rounded-full bg-tag-bg border border-border hover:bg-red-50 transition-colors"
              title="Logout"
            >
              <LogOut size={16} className="text-muted" />
            </button>
          </div>
        ) : (
          <Link
            href="/login"
            className="flex items-center justify-center gap-2 bg-primary hover:bg-primary-hover text-white font-medium rounded-full px-5 h-10 text-sm transition-colors"
          >
            <User size={14} />
            Sign In
          </Link>
        )}
      </div>
    </nav>
  );
}
