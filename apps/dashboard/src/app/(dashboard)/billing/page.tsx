"use client";

import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { useAuthStore } from "@/lib/auth-store";
import { useAuthReady } from "@/hooks/use-auth-ready";
import { api } from "@/lib/api";
import { ExternalLink } from "lucide-react";

export default function BillingPage() {
  const user = useAuthStore((s) => s.user);
  const authReady = useAuthReady();
  const [upgrading, setUpgrading] = useState(false);
  const [openingPortal, setOpeningPortal] = useState(false);
  const [portalError, setPortalError] = useState("");

  const { data: vaults } = useQuery({
    queryKey: ["vaults"],
    queryFn: () => api.listVaults(),
    enabled: authReady,
  });

  const { data: devices } = useQuery({
    queryKey: ["devices"],
    queryFn: () => api.listDevices(),
    enabled: authReady,
  });

  const isPro = user?.plan === "pro";
  const vaultCount = vaults?.length ?? 0;
  const totalSecrets = vaults?.reduce((sum, v) => sum + (v.secret_count || 0), 0) ?? 0;
  const totalStorage = vaults?.reduce((sum, v) => sum + (v.size || 0), 0) ?? 0;
  const deviceCount = devices?.length ?? 0;

  const handleManageSubscription = async () => {
    setOpeningPortal(true);
    setPortalError("");
    try {
      const { portal_url } = await api.getPortal();
      window.open(portal_url, "_blank");
    } catch (err) {
      setPortalError(
        err instanceof Error && err.message.includes("no active subscription")
          ? "No subscription found. Run 'tene billing upgrade' in CLI to subscribe."
          : "Failed to open billing portal. Try again later."
      );
    } finally {
      setOpeningPortal(false);
    }
  };

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold">Billing</h1>
        <p className="text-muted text-sm mt-1">Manage your subscription and usage</p>
      </div>

      {/* Current Plan */}
      <div className="rounded-xl border border-accent/30 bg-surface p-6 relative overflow-hidden">
        <div className="absolute inset-0 bg-gradient-to-br from-accent/5 to-transparent pointer-events-none" />
        <div className="relative flex items-start justify-between">
          <div>
            <p className="text-xs text-muted uppercase tracking-wider">Current Plan</p>
            <p className="text-2xl font-bold mt-1">{isPro ? "Pro" : "Free"}</p>
            <p className="text-muted text-sm mt-1">
              {isPro ? "$5/month · All features included" : "Local-only · Free forever"}
            </p>
          </div>
          {isPro ? (
            <button
              onClick={handleManageSubscription}
              disabled={openingPortal}
              className="flex items-center gap-1.5 px-4 py-2 rounded-lg border border-border text-sm text-muted hover:text-foreground transition-colors disabled:opacity-50"
            >
              Manage <ExternalLink size={12} />
            </button>
          ) : (
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
              className="px-4 py-2 rounded-lg bg-accent text-background text-sm font-medium hover:bg-accent-dim transition-colors active:scale-[0.98] disabled:opacity-50"
            >
              {upgrading ? "Redirecting..." : "Upgrade to Pro"}
            </button>
          )}
        </div>
      </div>

      {portalError && (
        <div className="rounded-lg border border-warning/30 bg-warning/10 p-3 text-sm text-warning">
          {portalError}
        </div>
      )}

      {/* Usage Stats */}
      <div className="rounded-xl border border-border bg-surface p-5">
        <h3 className="text-sm font-semibold mb-4">Usage</h3>
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
          <div>
            <p className="text-2xl font-bold font-mono">{vaultCount}</p>
            <p className="text-xs text-muted mt-0.5">Synced Vaults</p>
          </div>
          <div>
            <p className="text-2xl font-bold font-mono">{totalSecrets}</p>
            <p className="text-xs text-muted mt-0.5">Total Secrets</p>
          </div>
          <div>
            <p className="text-2xl font-bold font-mono">{deviceCount}</p>
            <p className="text-xs text-muted mt-0.5">Devices</p>
          </div>
          <div>
            <p className="text-2xl font-bold font-mono">{formatBytes(totalStorage)}</p>
            <p className="text-xs text-muted mt-0.5">Cloud Storage</p>
          </div>
        </div>
      </div>

      {/* Plan Comparison */}
      <div className="grid sm:grid-cols-2 gap-4">
        <div className={`rounded-xl border bg-surface p-6 ${!isPro ? "border-accent/20" : "border-border"}`}>
          <div className="flex items-center justify-between mb-3">
            <h3 className="font-semibold">Free</h3>
            {!isPro && <span className="badge badge-free">Current</span>}
          </div>
          <p className="text-2xl font-bold font-mono">$0<span className="text-sm text-muted font-normal">/forever</span></p>
          <ul className="mt-4 space-y-2 text-sm text-muted">
            <li className="flex items-center gap-2"><span className="text-accent">✓</span> Local secret management</li>
            <li className="flex items-center gap-2"><span className="text-accent">✓</span> AI agent auto-detection</li>
            <li className="flex items-center gap-2"><span className="text-accent">✓</span> 12-word recovery key</li>
            <li className="flex items-center gap-2"><span className="text-accent">✓</span> XChaCha20-Poly1305 encryption</li>
          </ul>
        </div>

        <div className={`rounded-xl border bg-surface p-6 ${isPro ? "border-accent/20" : "border-border"}`}>
          <div className="flex items-center justify-between mb-3">
            <h3 className="font-semibold">Pro</h3>
            {isPro && <span className="badge badge-pro">Current</span>}
          </div>
          <p className="text-2xl font-bold font-mono">$5<span className="text-sm text-muted font-normal">/month</span></p>
          <ul className="mt-4 space-y-2 text-sm text-muted">
            <li className="flex items-center gap-2"><span className="text-accent">✓</span> Everything in Free</li>
            <li className="flex items-center gap-2"><span className="text-accent">✓</span> Vault cloud sync (push/pull)</li>
            <li className="flex items-center gap-2"><span className="text-accent">✓</span> Team secret sharing + RBAC</li>
            <li className="flex items-center gap-2"><span className="text-accent">✓</span> Device management</li>
            <li className="flex items-center gap-2"><span className="text-accent">✓</span> Audit log &amp; Dashboard</li>
          </ul>
        </div>
      </div>

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
