"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { Search, User } from "lucide-react";

const navLinks = [
  { href: "/", label: "Discover" },
  { href: "/?category=Concerts", label: "Concerts" },
  { href: "/?category=Sports", label: "Sports" },
  { href: "/?category=Films", label: "Films" },
];

export function Navbar() {
  const pathname = usePathname();

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
        <Link
          href="/my-tickets"
          className="flex items-center justify-center w-10 h-10 rounded-full bg-tag-bg border border-border"
        >
          <User size={18} className="text-muted" />
        </Link>
      </div>
    </nav>
  );
}
