"use client";

import { useState, useRef, useEffect } from "react";
import { useRouter } from "next/navigation";
import { CreditCard, LogOut, ChevronUp, ChevronDown, Copy, Check } from "lucide-react";
import { useAuthStore } from "@/lib/auth-store";
import { api } from "@/lib/api";

function UserAvatar({ user, className = "w-6 h-6" }: { user: { name: string; avatar_url: string }; className?: string }) {
  if (user.avatar_url) {
    // eslint-disable-next-line @next/next/no-img-element
    return <img src={user.avatar_url} alt="" className={`${className} rounded-full`} />;
  }
  return (
    <div className={`${className} rounded-full bg-accent/20 flex items-center justify-center text-accent text-xs font-bold`}>
      {user.name?.[0]?.toUpperCase() || "?"}
    </div>
  );
}

function CopyUserId({ userId }: { userId: string }) {
  const [copied, setCopied] = useState(false);
  const handleCopy = () => {
    navigator.clipboard.writeText(userId);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };
  return (
    <button
      onClick={handleCopy}
      className="flex items-center justify-between w-full px-3 py-1.5 hover:bg-surface-2 transition-colors rounded group"
      title="Copy User ID for team invites"
    >
      <div className="flex flex-col items-start">
        <span className="text-[10px] text-muted uppercase tracking-wider">User ID</span>
        <span className="font-mono text-xs text-muted group-hover:text-foreground">{userId.slice(0, 8)}...</span>
      </div>
      {copied ? (
        <span className="flex items-center gap-1 text-accent text-[10px]"><Check size={10} /> Copied</span>
      ) : (
        <Copy size={10} className="text-muted group-hover:text-foreground shrink-0" />
      )}
    </button>
  );
}

// Desktop: sidebar bottom user menu (Billing + Sign out only — Settings is in sidebar nav)
export function UserMenu() {
  const user = useAuthStore((s) => s.user);
  const logout = useAuthStore((s) => s.logout);
  const router = useRouter();
  const [open, setOpen] = useState(false);
  const ref = useRef<HTMLDivElement>(null);

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
      router.push("/billing");
    }
  };

  if (!user) return null;

  return (
    <div ref={ref} className="relative">
      <button
        onClick={() => setOpen(!open)}
        className="flex items-center gap-2 w-full px-2 py-1.5 rounded-lg text-sm text-muted hover:text-foreground hover:bg-surface-2 transition-colors"
      >
        <UserAvatar user={user} />
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
          <CopyUserId userId={user.id} />
          <div className="border-t border-border" />
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

// Mobile: top-right avatar dropdown (profile, billing, sign out)
export function MobileUserMenu() {
  const user = useAuthStore((s) => s.user);
  const logout = useAuthStore((s) => s.logout);
  const router = useRouter();
  const [open, setOpen] = useState(false);
  const ref = useRef<HTMLDivElement>(null);

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
      router.push("/billing");
    }
  };

  if (!user) return null;

  return (
    <div ref={ref} className="relative">
      <button
        onClick={() => setOpen(!open)}
        className="flex items-center gap-1.5"
      >
        <UserAvatar user={user} className="w-7 h-7" />
        <ChevronDown size={12} className="text-muted" />
      </button>

      {open && (
        <div className="absolute right-0 top-full mt-2 w-56 rounded-xl border border-border bg-surface shadow-lg z-50">
          <div className="px-3 py-2.5">
            <p className="text-sm font-medium truncate">{user.name || "User"}</p>
            <p className="text-xs text-muted truncate">{user.email}</p>
            <p className="text-xs mt-1">
              Plan: <span className="text-accent font-medium">{user.plan === "pro" ? "Pro" : "Free"}</span>
            </p>
          </div>
          <CopyUserId userId={user.id} />
          <div className="border-t border-border" />
          <button
            onClick={handleBilling}
            className="flex items-center gap-2 w-full px-3 py-2.5 text-sm text-muted hover:text-foreground hover:bg-surface-2 transition-colors"
          >
            <CreditCard size={14} /> Billing
          </button>
          <div className="border-t border-border" />
          <button
            onClick={handleSignout}
            className="flex items-center gap-2 w-full px-3 py-2.5 text-sm text-muted hover:text-foreground hover:bg-surface-2 transition-colors"
          >
            <LogOut size={14} /> Sign out
          </button>
        </div>
      )}
    </div>
  );
}
