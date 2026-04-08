"use client";

import { use } from "react";
import Link from "next/link";
import { useQuery } from "@tanstack/react-query";
import { useAuthReady } from "@/hooks/use-auth-ready";
import { api } from "@/lib/api";

interface VaultDetailPageProps {
  params: Promise<{ id: string }>;
}

function formatBytes(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

export default function VaultDetailPage({ params }: VaultDetailPageProps) {
  const { id } = use(params);
  const authReady = useAuthReady();

  const { data: vault, isLoading } = useQuery({
    queryKey: ["vaults", id],
    queryFn: () => api.getVault(id),
    enabled: authReady,
  });

  if (isLoading) {
    return (
      <div className="space-y-6">
        <div className="flex items-center gap-2">
          <Link href="/vaults" className="text-muted hover:text-foreground text-sm">&larr; Vaults</Link>
          <span className="text-muted">/</span>
          <span className="text-2xl font-bold font-mono text-muted animate-pulse">Loading...</span>
        </div>
      </div>
    );
  }

  if (!vault) {
    return (
      <div className="space-y-6">
        <div className="flex items-center gap-2">
          <Link href="/vaults" className="text-muted hover:text-foreground text-sm">&larr; Vaults</Link>
          <span className="text-muted">/</span>
          <span className="text-2xl font-bold text-danger">Vault not found</span>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div>
        <div className="flex items-center gap-2">
          <Link href="/vaults" className="text-muted hover:text-foreground text-sm">&larr; Vaults</Link>
          <span className="text-muted">/</span>
          <h1 className="text-2xl font-bold font-mono text-accent">{vault.project_name}</h1>
        </div>
        <div className="flex items-center gap-3 mt-2">
          <span className="badge badge-active">v{vault.vault_version}</span>
          <span className="text-xs text-muted">{vault.secret_count} secrets</span>
          <span className="text-xs text-muted">{formatBytes(vault.size)}</span>
          <span className="text-xs text-muted">Synced {new Date(vault.updated_at).toLocaleDateString()}</span>
        </div>
      </div>

      <div className="grid sm:grid-cols-3 gap-4">
        <div className="rounded-xl border border-border bg-surface p-4">
          <p className="text-xs text-muted uppercase tracking-wider">Version</p>
          <p className="text-xl font-mono mt-1">v{vault.vault_version}</p>
        </div>
        <div className="rounded-xl border border-border bg-surface p-4">
          <p className="text-xs text-muted uppercase tracking-wider">Secrets</p>
          <p className="text-xl font-mono mt-1">{vault.secret_count}</p>
        </div>
        <div className="rounded-xl border border-border bg-surface p-4">
          <p className="text-xs text-muted uppercase tracking-wider">Size</p>
          <p className="text-xl font-mono mt-1">{formatBytes(vault.size)}</p>
        </div>
      </div>

      <div className="rounded-xl border border-border bg-surface p-6">
        <h2 className="text-sm font-semibold mb-3">Vault Details</h2>
        <dl className="space-y-2 text-sm">
          <div className="flex justify-between">
            <dt className="text-muted">Vault ID</dt>
            <dd className="font-mono text-xs">{vault.id}</dd>
          </div>
          <div className="flex justify-between">
            <dt className="text-muted">Created</dt>
            <dd>{new Date(vault.created_at).toLocaleString()}</dd>
          </div>
          <div className="flex justify-between">
            <dt className="text-muted">Last Sync</dt>
            <dd>{new Date(vault.updated_at).toLocaleString()}</dd>
          </div>
          {vault.team_id && (
            <div className="flex justify-between">
              <dt className="text-muted">Team</dt>
              <dd className="font-mono text-xs">{vault.team_id}</dd>
            </div>
          )}
        </dl>
      </div>

      <div className="rounded-xl border border-dashed border-border p-4 text-center text-xs text-muted space-y-1">
        <p>Secret values are <strong>never</strong> sent to the server or displayed in the dashboard.</p>
        <p>
          Values are encrypted with XChaCha20-Poly1305 on your device.
          Use <code className="font-mono text-accent">tene get KEY</code> in your CLI to access values.
        </p>
      </div>
    </div>
  );
}
