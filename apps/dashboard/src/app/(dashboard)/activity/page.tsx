"use client";

import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { useAuthReady } from "@/hooks/use-auth-ready";
import { api, type AuditLog, type AuditFilter } from "@/lib/api";

const ACTION_FILTERS = [
  { value: "", label: "All" },
  { value: "vault.push", label: "Push" },
  { value: "vault.pull", label: "Pull" },
  { value: "vault.create", label: "Create" },
  { value: "vault.delete", label: "Delete" },
  { value: "auth.login", label: "Login" },
  { value: "auth.logout", label: "Logout" },
  { value: "team.invite", label: "Invite" },
  { value: "team.remove", label: "Remove" },
] as const;

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

function actionBadgeClass(action: string): string {
  if (action.includes("push")) return "badge-active";
  if (action.includes("pull")) return "badge-free";
  if (action.includes("delete") || action.includes("remove")) return "badge-danger";
  if (action.includes("login") || action.includes("logout")) return "badge-pro";
  return "badge-free";
}

export default function ActivityPage() {
  const authReady = useAuthReady();
  const [filters, setFilters] = useState<AuditFilter>({ limit: 50, offset: 0 });
  const [allLogs, setAllLogs] = useState<AuditLog[]>([]);

  const { data: logs, isLoading } = useQuery({
    queryKey: ["audit-logs", filters],
    queryFn: () => api.listAuditLogs(filters),
    enabled: authReady,
  });

  // Accumulate logs for "load more" pagination
  const displayLogs = filters.offset === 0 ? (logs ?? []) : [...allLogs, ...(logs ?? [])];

  const handleFilterChange = (action: string) => {
    setAllLogs([]);
    setFilters({ action: action || undefined, limit: 50, offset: 0 });
  };

  const loadMore = () => {
    setAllLogs(displayLogs);
    setFilters((f) => ({ ...f, offset: (f.offset ?? 0) + 50 }));
  };

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold">Activity</h1>
        <p className="text-muted text-sm mt-1">Security activity across your account</p>
      </div>

      {/* Filters */}
      <div className="flex gap-2 flex-wrap">
        {ACTION_FILTERS.map((f) => (
          <button
            key={f.value}
            onClick={() => handleFilterChange(f.value)}
            className={`px-3 py-1.5 rounded-lg text-xs border transition-colors ${
              (filters.action ?? "") === f.value
                ? "border-accent/30 bg-accent/10 text-accent"
                : "border-border bg-surface text-muted hover:text-foreground"
            }`}
          >
            {f.label}
          </button>
        ))}
      </div>

      {/* Table */}
      <div className="rounded-xl border border-border bg-surface overflow-hidden">
        <table className="w-full text-sm" role="table" aria-label="Activity log">
          <thead>
            <tr className="border-b border-border text-left text-xs text-muted uppercase tracking-wider">
              <th className="px-4 py-3" scope="col">Time</th>
              <th className="px-4 py-3" scope="col">Action</th>
              <th className="px-4 py-3 hidden sm:table-cell" scope="col">Project</th>
              <th className="px-4 py-3 hidden md:table-cell" scope="col">IP</th>
            </tr>
          </thead>
          <tbody>
            {isLoading && filters.offset === 0 ? (
              <tr>
                <td colSpan={4} className="px-4 py-12 text-center text-muted animate-pulse">
                  Loading activity...
                </td>
              </tr>
            ) : displayLogs.length === 0 ? (
              <tr>
                <td colSpan={4} className="px-4 py-12 text-center text-muted">
                  <span className="font-mono text-2xl block mb-2">&#x2630;</span>
                  No activity{filters.action ? ` for "${filters.action}"` : ""} yet
                </td>
              </tr>
            ) : (
              displayLogs.map((log) => (
                <tr key={log.id} className="border-b border-border last:border-0 hover:bg-surface-2 transition-colors">
                  <td className="px-4 py-3 text-xs text-muted whitespace-nowrap">
                    {timeAgo(log.created_at)}
                  </td>
                  <td className="px-4 py-3">
                    <span className={`badge ${actionBadgeClass(log.action)}`}>
                      {log.action}
                    </span>
                  </td>
                  <td className="px-4 py-3 hidden sm:table-cell text-xs text-muted font-mono truncate max-w-[200px]">
                    {log.detail || "-"}
                  </td>
                  <td className="px-4 py-3 hidden md:table-cell font-mono text-xs text-muted">
                    {log.ip_address || "-"}
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {/* Load More */}
      {(logs?.length ?? 0) >= 50 && (
        <button
          onClick={loadMore}
          disabled={isLoading}
          className="w-full py-2 text-sm text-muted hover:text-foreground border border-border rounded-lg hover:bg-surface transition-colors disabled:opacity-50"
        >
          {isLoading ? "Loading..." : "Load more..."}
        </button>
      )}
    </div>
  );
}
