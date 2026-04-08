"use client";

import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { UsageBar } from "@/components/usage-bar";
import { useAuthStore } from "@/lib/auth-store";
import { useAuthReady } from "@/hooks/use-auth-ready";
import { api } from "@/lib/api";

export default function BillingPage() {
  const user = useAuthStore((s) => s.user);
  const authReady = useAuthReady();
  const [upgrading, setUpgrading] = useState(false);

  const { data: vaults } = useQuery({
    queryKey: ["vaults"],
    queryFn: () => api.listVaults(),
    enabled: authReady,
  });

  const isPro = user?.plan === "pro";
  const vaultCount = vaults?.length ?? 0;
  const totalSecrets = vaults?.reduce((sum, v) => sum + (v.secret_count || 0), 0) ?? 0;
  const totalStorageMB = vaults?.reduce((sum, v) => sum + (v.size || 0), 0) ?? 0;

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold">Billing</h1>
        <p className="text-muted text-sm mt-1">Manage your subscription</p>
      </div>

      <div className="grid sm:grid-cols-2 gap-4">
        {/* Free Plan */}
        <div className={`rounded-xl border bg-surface p-6 ${!isPro ? "border-accent/30" : "border-border"}`}>
          <div className="flex items-center justify-between mb-4">
            <h3 className="font-semibold">Free</h3>
            {!isPro && <span className="badge badge-free">Current</span>}
          </div>
          <p className="text-3xl font-bold font-mono mb-1">$0<span className="text-sm text-muted font-normal">/forever</span></p>
          <ul className="mt-4 space-y-2 text-sm text-muted">
            <li className="flex items-center gap-2"><span className="text-accent">✓</span> Local secret management</li>
            <li className="flex items-center gap-2"><span className="text-accent">✓</span> AI agent auto-detection</li>
            <li className="flex items-center gap-2"><span className="text-accent">✓</span> 12-word recovery key</li>
            <li className="flex items-center gap-2"><span className="text-accent">✓</span> XChaCha20-Poly1305 encryption</li>
          </ul>
        </div>

        {/* Pro Plan */}
        <div className={`rounded-xl border bg-surface p-6 relative overflow-hidden ${isPro ? "border-accent/30" : "border-accent/30"}`}>
          <div className="absolute inset-0 bg-gradient-to-br from-accent/5 to-transparent pointer-events-none" />
          <div className="relative">
            <div className="flex items-center justify-between mb-4">
              <h3 className="font-semibold">Pro</h3>
              {isPro ? (
                <span className="badge badge-pro">Current</span>
              ) : (
                <span className="badge badge-pro">Upgrade</span>
              )}
            </div>
            <p className="text-3xl font-bold font-mono mb-1">$5<span className="text-sm text-muted font-normal">/month</span></p>
            <ul className="mt-4 space-y-2 text-sm text-muted">
              <li className="flex items-center gap-2"><span className="text-accent">✓</span> Everything in Free</li>
              <li className="flex items-center gap-2"><span className="text-accent">✓</span> Vault cloud sync (push/pull)</li>
              <li className="flex items-center gap-2"><span className="text-accent">✓</span> Encrypted cloud backup</li>
              <li className="flex items-center gap-2"><span className="text-accent">✓</span> Device management</li>
              <li className="flex items-center gap-2"><span className="text-accent">✓</span> Team secret sharing + RBAC</li>
              <li className="flex items-center gap-2"><span className="text-accent">✓</span> Full audit log</li>
            </ul>
            {!isPro && (
              <button
                disabled={upgrading}
                onClick={async () => {
                  setUpgrading(true);
                  try {
                    const { checkout_url } = await api.createCheckout(user?.email || "");
                    window.location.href = checkout_url;
                  } catch {
                    setUpgrading(false);
                  }
                }}
                className="mt-6 w-full py-2.5 rounded-lg bg-accent text-background font-medium text-sm hover:bg-accent-dim transition-colors active:scale-[0.98] disabled:opacity-50"
              >
                {upgrading ? "Redirecting..." : "Upgrade to Pro"}
              </button>
            )}
          </div>
        </div>
      </div>

      {isPro && (
        <div className="rounded-xl border border-border bg-surface p-5 space-y-4">
          <h3 className="text-sm font-semibold">Usage</h3>
          <div className="space-y-3">
            <div className="flex items-center justify-between text-sm">
              <span className="text-muted">Synced Vaults</span>
              <span className="font-mono">{vaultCount}</span>
            </div>
            <div className="flex items-center justify-between text-sm">
              <span className="text-muted">Total Secrets</span>
              <span className="font-mono">{totalSecrets}</span>
            </div>
            <div className="flex items-center justify-between text-sm">
              <span className="text-muted">Cloud Storage</span>
              <span className="font-mono">{formatBytes(totalStorageMB)}</span>
            </div>
          </div>
        </div>
      )}

      <div className="rounded-xl border border-border bg-surface p-4 text-xs text-muted text-center">
        Payments processed securely by LemonSqueezy. Cancel anytime.
      </div>
    </div>
  );
}

function formatBytes(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}
