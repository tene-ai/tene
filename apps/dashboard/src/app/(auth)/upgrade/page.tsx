"use client";

import { useState } from "react";
import { useAuthStore } from "@/lib/auth-store";
import { api } from "@/lib/api";
import { Shield, Cloud, Users, ScrollText, Monitor, LayoutDashboard } from "lucide-react";

const features = [
  { icon: Cloud, label: "Vault cloud sync (push/pull)" },
  { icon: Shield, label: "Zero-Knowledge encrypted backup" },
  { icon: Users, label: "Team secret sharing + RBAC" },
  { icon: ScrollText, label: "Full audit log" },
  { icon: Monitor, label: "Device management" },
  { icon: LayoutDashboard, label: "Dashboard access" },
];

export default function UpgradePage() {
  const user = useAuthStore((s) => s.user);
  const [upgrading, setUpgrading] = useState(false);

  const handleUpgrade = async () => {
    setUpgrading(true);
    try {
      const { checkout_url } = await api.createCheckout(user?.email || "");
      window.location.href = checkout_url;
    } catch {
      setUpgrading(false);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center p-4">
      <div className="w-full max-w-md space-y-8">
        <div className="text-center">
          {/* eslint-disable-next-line @next/next/no-img-element */}
          <img src="/logo.svg" alt="Tene" width={48} height={48} className="rounded-xl mx-auto mb-3" />
          <h1 className="font-mono font-bold text-2xl mb-2">Unlock Tene Cloud</h1>
          <p className="text-muted text-sm">
            Manage vaults, team, and audit logs with zero-knowledge encryption.
          </p>
        </div>

        <div className="rounded-xl border border-accent/30 bg-surface p-6 relative overflow-hidden">
          <div className="absolute inset-0 bg-gradient-to-br from-accent/5 to-transparent pointer-events-none" />
          <div className="relative">
            <div className="flex items-center justify-between mb-4">
              <h3 className="font-semibold">Pro</h3>
              <span className="text-accent font-mono font-bold text-2xl">$5<span className="text-sm text-muted font-normal">/month</span></span>
            </div>

            <ul className="space-y-3 mb-6">
              {features.map(({ icon: Icon, label }) => (
                <li key={label} className="flex items-center gap-3 text-sm">
                  <Icon size={16} className="text-accent shrink-0" />
                  <span>{label}</span>
                </li>
              ))}
            </ul>

            <button
              disabled={upgrading}
              onClick={handleUpgrade}
              className="w-full py-3 rounded-lg bg-accent text-background font-medium text-sm hover:bg-accent-dim transition-colors active:scale-[0.98] disabled:opacity-50"
            >
              {upgrading ? "Redirecting to checkout..." : "Upgrade to Pro — $5/month"}
            </button>
          </div>
        </div>

        <div className="text-center text-xs text-muted space-y-1">
          <p>Already using the CLI? That&apos;s free forever.</p>
          <a href="https://tene.sh" className="text-accent hover:underline">
            ← Back to tene.sh
          </a>
        </div>
      </div>
    </div>
  );
}
