"use client";

import { use, useState, useEffect } from "react";
import Link from "next/link";
import { useQuery } from "@tanstack/react-query";
import { useAuthReady } from "@/hooks/use-auth-ready";
import { api } from "@/lib/api";
import { clsx } from "clsx";

interface ProjectDetailPageProps {
  params: Promise<{ id: string }>;
}

function formatBytes(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

function timeAgo(date: string): string {
  const seconds = Math.floor((Date.now() - new Date(date).getTime()) / 1000);
  if (seconds < 60) return "just now";
  const minutes = Math.floor(seconds / 60);
  if (minutes < 60) return `${minutes}m ago`;
  const hours = Math.floor(minutes / 60);
  if (hours < 24) return `${hours}h ago`;
  const days = Math.floor(hours / 24);
  return `${days}d ago`;
}

export default function ProjectDetailPage({ params }: ProjectDetailPageProps) {
  const { id } = use(params);
  const authReady = useAuthReady();
  const [activeEnv, setActiveEnv] = useState<string>();

  const { data: vault, isLoading: vaultLoading } = useQuery({
    queryKey: ["vaults", id],
    queryFn: () => api.getVault(id),
    enabled: authReady,
  });

  const { data: keysData, isLoading: keysLoading } = useQuery({
    queryKey: ["vault-keys", id, activeEnv],
    queryFn: () => api.getVaultKeys(id, activeEnv),
    enabled: authReady && !!vault,
  });

  // Set first environment as active on initial load
  useEffect(() => {
    if (keysData?.environments?.length && !activeEnv) {
      setActiveEnv(keysData.environments[0]);
    }
  }, [keysData, activeEnv]);

  if (vaultLoading) {
    return (
      <div className="space-y-6">
        <nav className="flex items-center gap-2 text-sm text-muted">
          <Link href="/" className="hover:text-foreground">Projects</Link>
          <span>/</span>
          <span className="text-foreground animate-pulse">Loading...</span>
        </nav>
      </div>
    );
  }

  if (!vault) {
    return (
      <div className="space-y-6">
        <nav className="flex items-center gap-2 text-sm text-muted">
          <Link href="/" className="hover:text-foreground">Projects</Link>
          <span>/</span>
          <span className="text-danger">Not found</span>
        </nav>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Breadcrumb */}
      <nav className="flex items-center gap-2 text-sm text-muted">
        <Link href="/" className="hover:text-foreground">Projects</Link>
        <span>/</span>
        <span className="text-foreground">{vault.project_name}</span>
      </nav>

      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold font-mono">{vault.project_name}</h1>
          <p className="text-muted text-sm mt-1">
            v{vault.vault_version} · {vault.secret_count} secrets ·{" "}
            {formatBytes(vault.size)} ·{" "}
            synced {timeAgo(vault.updated_at)}
          </p>
        </div>
      </div>

      {/* Environment Tabs */}
      {keysData?.environments && keysData.environments.length > 0 && (
        <div className="flex gap-1 border-b border-border">
          {keysData.environments.map((env) => (
            <button
              key={env}
              onClick={() => setActiveEnv(env)}
              className={clsx(
                "px-4 py-2 text-sm font-mono border-b-2 transition-colors -mb-px",
                env === activeEnv
                  ? "border-accent text-accent"
                  : "border-transparent text-muted hover:text-foreground"
              )}
            >
              {env}
            </button>
          ))}
        </div>
      )}

      {/* Secret Key Table */}
      <div className="rounded-xl border border-border bg-surface overflow-hidden">
        <table className="w-full text-sm" role="table" aria-label="Secret keys">
          <thead>
            <tr className="border-b border-border text-left text-xs text-muted uppercase tracking-wider">
              <th className="px-4 py-3" scope="col">Key</th>
              <th className="px-4 py-3 hidden sm:table-cell" scope="col">Updated</th>
              <th className="px-4 py-3 hidden md:table-cell" scope="col">Version</th>
            </tr>
          </thead>
          <tbody>
            {keysLoading ? (
              <tr>
                <td colSpan={3} className="px-4 py-12 text-center text-muted animate-pulse">
                  Loading keys...
                </td>
              </tr>
            ) : !keysData?.keys?.length ? (
              <tr>
                <td colSpan={3} className="px-4 py-12 text-center text-muted">
                  <p className="text-sm">No keys in this environment</p>
                  <p className="text-xs mt-1">
                    Push secrets with <code className="font-mono text-accent">tene push</code>
                  </p>
                </td>
              </tr>
            ) : (
              keysData.keys.map((k) => (
                <tr key={k.name} className="border-b border-border last:border-0 hover:bg-surface-2 transition-colors">
                  <td className="px-4 py-3 font-mono text-sm">{k.name}</td>
                  <td className="px-4 py-3 hidden sm:table-cell text-xs text-muted">
                    {timeAgo(k.updated_at)}
                  </td>
                  <td className="px-4 py-3 hidden md:table-cell">
                    <span className="badge badge-active">v{k.version}</span>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {/* ZK Notice */}
      <div className="rounded-xl border border-dashed border-border p-4 text-center text-xs text-muted">
        <p>Secret values are <strong>never</strong> visible here.</p>
        <p>Use <code className="font-mono text-accent">tene get KEY</code> in your CLI.</p>
        <p className="mt-1 text-muted/60">Zero-Knowledge: XChaCha20-Poly1305 encryption</p>
      </div>
    </div>
  );
}
