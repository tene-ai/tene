"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { clsx } from "clsx";
import { LayoutDashboard, FolderKey, Users, ScrollText, Settings } from "lucide-react";
import { UserMenu, MobileUserMenu } from "./user-menu";

const nav = [
  { href: "/", label: "Overview", icon: LayoutDashboard },
  { href: "/projects", label: "Projects", icon: FolderKey },
  { href: "/team", label: "Team", icon: Users },
  { href: "/activity", label: "Activity", icon: ScrollText },
  { href: "/settings", label: "Settings", icon: Settings },
];

export function Sidebar() {
  const pathname = usePathname();

  return (
    <aside aria-label="Main navigation" className="hidden lg:flex flex-col w-56 border-r border-border/60 glass h-screen sticky top-0 z-20">
      <div className="p-4 border-b border-border">
        <Link href="/" className="flex items-center gap-2">
          {/* eslint-disable-next-line @next/next/no-img-element */}
          <img src="/logo.svg" alt="Tene" width={28} height={28} className="rounded-md" />
          <span className="font-mono font-bold text-lg">tene</span>
        </Link>
      </div>
      <nav className="flex-1 p-3 space-y-1">
        {nav.map((item) => {
          const active = item.href === "/"
            ? pathname === "/"
            : pathname.startsWith(item.href);
          const Icon = item.icon;
          return (
            <Link
              key={item.href}
              href={item.href}
              className={clsx(
                "flex items-center gap-3 px-3 py-2 rounded-lg text-sm transition-colors",
                active
                  ? "bg-accent/10 text-accent border border-accent/20"
                  : "text-muted hover:text-foreground hover:bg-surface-2"
              )}
            >
              <Icon size={16} />
              {item.label}
            </Link>
          );
        })}
      </nav>
      <div className="p-3 border-t border-border">
        <UserMenu />
      </div>
    </aside>
  );
}

// Mobile: top header bar with logo + user avatar (설계서 §2.2)
export function MobileHeader() {
  return (
    <header className="lg:hidden flex items-center justify-between px-4 py-3 border-b border-border/60 glass sticky top-0 z-40">
      <Link href="/" className="flex items-center gap-2">
        {/* eslint-disable-next-line @next/next/no-img-element */}
        <img src="/logo.svg" alt="Tene" width={24} height={24} className="rounded-md" />
        <span className="font-mono font-bold">tene</span>
      </Link>
      <MobileUserMenu />
    </header>
  );
}

// Mobile: bottom navigation bar (4 tabs)
export function MobileNav() {
  const pathname = usePathname();

  return (
    <nav className="lg:hidden fixed bottom-0 left-0 right-0 z-50 glass border-t border-border/60 flex justify-around py-2">
      {nav.map((item) => {
        const active = item.href === "/"
          ? pathname === "/" || pathname.startsWith("/projects")
          : pathname.startsWith(item.href);
        const Icon = item.icon;
        return (
          <Link
            key={item.href}
            href={item.href}
            className={clsx(
              "flex flex-col items-center gap-0.5 px-3 py-1 text-xs",
              active ? "text-accent" : "text-muted"
            )}
          >
            <Icon size={18} />
            <span>{item.label}</span>
          </Link>
        );
      })}
    </nav>
  );
}
