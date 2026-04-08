"use client";

import { useQuery } from "@tanstack/react-query";
import { useAuthReady } from "@/hooks/use-auth-ready";
import { api } from "@/lib/api";
import Link from "next/link";
import { FolderKey, ArrowUpRight } from "lucide-react";

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

function formatBytes(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

export default function ProjectsPage() {
  const authReady = useAuthReady();

  const { data: vaults, isLoading } = useQuery({
    queryKey: ["vaults"],
    queryFn: () => api.listVaults(),
    enabled: authReady,
  });

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">Projects</h1>
          <p className="text-muted text-sm mt-1">
            {vaults?.length ?? 0} synced projects
          </p>
        </div>
      </div>

      {isLoading ? (
        <div className="grid gap-4">
          {[1, 2, 3].map((i) => (
            <div key={i} className="rounded-xl border border-border bg-surface p-5 animate-pulse">
              <div className="h-5 w-40 bg-surface-2 rounded mb-3" />
              <div className="h-3 w-60 bg-surface-2 rounded" />
            </div>
          ))}
        </div>
      ) : !vaults?.length ? (
        <div className="rounded-xl border border-dashed border-border p-12 text-center">
          <FolderKey size={32} className="mx-auto text-muted mb-3" />
          <p className="text-sm font-medium">No projects yet</p>
          <p className="text-xs text-muted mt-2">
            Push your first vault to see it here:
          </p>
          <div className="font-mono text-sm bg-background rounded-lg p-4 mt-4 inline-block text-left space-y-1 text-muted">
            <p><span className="text-accent">$</span> cd your-project</p>
            <p><span className="text-accent">$</span> tene init</p>
            <p><span className="text-accent">$</span> tene push</p>
          </div>
        </div>
      ) : (
        <div className="grid gap-4">
          {vaults.map((v) => (
            <Link
              key={v.id}
              href={`/projects/${v.id}`}
              className="group rounded-xl border border-border bg-surface p-5 hover:border-accent/30 hover:bg-surface-2 transition-all"
            >
              <div className="flex items-start justify-between">
                <div className="flex items-center gap-3">
                  <FolderKey size={18} className="text-accent shrink-0" />
                  <div>
                    <h2 className="font-mono font-semibold text-sm group-hover:text-accent transition-colors">
                      {v.project_name}
                    </h2>
                    <p className="text-xs text-muted mt-1">
                      {v.secret_count} secrets · v{v.vault_version} · {formatBytes(v.size)} ·
                      synced {timeAgo(v.updated_at)}
                    </p>
                  </div>
                </div>
                <ArrowUpRight size={14} className="text-muted group-hover:text-accent transition-colors shrink-0 mt-1" />
              </div>
            </Link>
          ))}
        </div>
      )}
    </div>
  );
}
