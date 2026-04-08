"use client";

import { useState, useRef, useEffect } from "react";
import { useRouter } from "next/navigation";
import { Monitor, CreditCard, LogOut, ChevronUp } from "lucide-react";
import { useAuthStore } from "@/lib/auth-store";
import { api } from "@/lib/api";

export function UserMenu() {
  const user = useAuthStore((s) => s.user);
  const logout = useAuthStore((s) => s.logout);
  const router = useRouter();
  const [open, setOpen] = useState(false);
  const ref = useRef<HTMLDivElement>(null);

  // Close on outside click
  useEffect(() => {
    const handler = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) setOpen(false);
    };
    document.addEventListener("mousedown", handler);
    return () => document.removeEventListener("mousedown", handler);
  }, []);

  const handleSignout = async () => {
    try {
      const refreshToken = useAuthStore.getState().refreshToken;
      await api.signout(refreshToken || undefined);
    } catch {
      // Non-fatal
    }
    logout();
    document.cookie = "tene_access_token=; max-age=0; path=/";
    router.replace("/login");
  };

  const handleBilling = async () => {
    setOpen(false);
    try {
      const { portal_url } = await api.getPortal();
      window.open(portal_url, "_blank");
    } catch {
      // Fallback: navigate to LemonSqueezy
      router.push("/settings");
    }
  };

  if (!user) return null;

  return (
    <div ref={ref} className="relative">
      <button
        onClick={() => setOpen(!open)}
        className="flex items-center gap-2 w-full px-2 py-1.5 rounded-lg text-sm text-muted hover:text-foreground hover:bg-surface-2 transition-colors"
      >
        {user.avatar_url ? (
          // eslint-disable-next-line @next/next/no-img-element
          <img src={user.avatar_url} alt="" className="w-6 h-6 rounded-full" />
        ) : (
          <div className="w-6 h-6 rounded-full bg-accent/20 flex items-center justify-center text-accent text-xs font-bold">
            {user.name?.[0]?.toUpperCase() || "?"}
          </div>
        )}
        <span className="truncate flex-1 text-left">{user.name || "User"}</span>
        <ChevronUp size={14} className={open ? "rotate-180 transition-transform" : "transition-transform"} />
      </button>

      {open && (
        <div className="absolute bottom-full left-0 right-0 mb-1 rounded-xl border border-border bg-surface shadow-lg z-50">
          <div className="px-3 py-2">
            <p className="text-xs text-muted truncate">{user.email}</p>
            <p className="text-xs mt-0.5">
              Plan: <span className="text-accent font-medium">{user.plan === "pro" ? "Pro" : "Free"}</span>
            </p>
          </div>
          <div className="border-t border-border" />
          <button
            onClick={() => { setOpen(false); router.push("/settings"); }}
            className="flex items-center gap-2 w-full px-3 py-2 text-sm text-muted hover:text-foreground hover:bg-surface-2 transition-colors"
          >
            <Monitor size={14} /> Devices
          </button>
          <button
            onClick={handleBilling}
            className="flex items-center gap-2 w-full px-3 py-2 text-sm text-muted hover:text-foreground hover:bg-surface-2 transition-colors"
          >
            <CreditCard size={14} /> Billing
          </button>
          <div className="border-t border-border" />
          <button
            onClick={handleSignout}
            className="flex items-center gap-2 w-full px-3 py-2 text-sm text-muted hover:text-foreground hover:bg-surface-2 transition-colors"
          >
            <LogOut size={14} /> Sign out
          </button>
        </div>
      )}
    </div>
  );
}
