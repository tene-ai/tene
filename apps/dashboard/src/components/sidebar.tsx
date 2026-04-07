"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { clsx } from "clsx";

const nav = [
  { href: "/", label: "Overview", icon: "⌂" },
  { href: "/vaults", label: "Vaults", icon: "◈" },
  { href: "/devices", label: "Devices", icon: "⊞" },
  { href: "/team", label: "Team", icon: "⊕" },
  { href: "/audit", label: "Audit Log", icon: "☰" },
  { href: "/billing", label: "Billing", icon: "◇" },
];

export function Sidebar() {
  const pathname = usePathname();

  return (
    <aside className="hidden lg:flex flex-col w-56 border-r border-border bg-surface h-screen sticky top-0">
      <div className="p-4 border-b border-border">
        <Link href="/" className="flex items-center gap-2">
          <span className="text-accent font-mono font-bold text-lg">tene</span>
          <span className="badge badge-pro text-[10px]">cloud</span>
        </Link>
      </div>
      <nav className="flex-1 p-3 space-y-1">
        {nav.map((item) => {
          const active = pathname === item.href || (item.href !== "/" && pathname.startsWith(item.href));
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
              <span className="font-mono text-xs w-4 text-center">{item.icon}</span>
              {item.label}
            </Link>
          );
        })}
      </nav>
      <div className="p-4 border-t border-border">
        <div className="flex items-center gap-2 text-sm text-muted">
          <span className="w-2 h-2 rounded-full bg-accent" />
          Connected
        </div>
      </div>
    </aside>
  );
}

export function MobileNav() {
  const pathname = usePathname();

  return (
    <nav className="lg:hidden fixed bottom-0 left-0 right-0 z-50 bg-surface border-t border-border flex justify-around py-2">
      {nav.slice(0, 5).map((item) => {
        const active = pathname === item.href || (item.href !== "/" && pathname.startsWith(item.href));
        return (
          <Link
            key={item.href}
            href={item.href}
            className={clsx(
              "flex flex-col items-center gap-0.5 px-3 py-1 text-xs",
              active ? "text-accent" : "text-muted"
            )}
          >
            <span className="font-mono">{item.icon}</span>
            <span>{item.label}</span>
          </Link>
        );
      })}
    </nav>
  );
}
