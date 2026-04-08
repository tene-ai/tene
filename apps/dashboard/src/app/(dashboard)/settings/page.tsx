"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { useAuthReady } from "@/hooks/use-auth-ready";
import { api } from "@/lib/api";
import { Monitor, Trash2 } from "lucide-react";

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

export default function SettingsPage() {
  const authReady = useAuthReady();
  const queryClient = useQueryClient();

  const { data: devices, isLoading } = useQuery({
    queryKey: ["devices"],
    queryFn: () => api.listDevices(),
    enabled: authReady,
  });

  const revokeMutation = useMutation({
    mutationFn: (id: string) => api.deleteDevice(id),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["devices"] }),
  });

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold">Settings</h1>
        <p className="text-muted text-sm mt-1">Devices and account settings</p>
      </div>

      {revokeMutation.isError && (
        <div className="rounded-lg border border-danger/30 bg-danger/10 p-3 text-sm text-danger">
          {revokeMutation.error instanceof Error ? revokeMutation.error.message : "Failed to revoke device"}
        </div>
      )}

      {/* Devices Section */}
      <section>
        <h2 className="text-sm font-semibold mb-4">Registered Devices</h2>
        {isLoading ? (
          <div className="grid gap-3 sm:grid-cols-2">
            {[1, 2].map((i) => (
              <div key={i} className="rounded-xl border border-border bg-surface p-4 animate-pulse h-20" />
            ))}
          </div>
        ) : !devices?.length ? (
          <div className="rounded-xl border border-dashed border-border p-8 text-center">
            <Monitor size={24} className="mx-auto text-muted mb-2" />
            <p className="text-sm text-muted">No devices registered</p>
            <p className="text-xs text-muted mt-1">
              Devices are registered when you run <code className="font-mono text-accent">tene login</code>
            </p>
          </div>
        ) : (
          <div className="grid gap-3 sm:grid-cols-2">
            {devices.map((d) => {
              const isOnline = Date.now() - new Date(d.last_seen_at).getTime() < 5 * 60 * 1000;
              return (
                <div key={d.id} className="rounded-xl border border-border bg-surface p-4 flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    <Monitor size={16} className="text-muted" />
                    <div>
                      <p className="text-sm font-medium">{d.device_name}</p>
                      <p className="text-xs text-muted flex items-center gap-1.5">
                        <span className={`w-1.5 h-1.5 rounded-full ${isOnline ? "bg-accent" : "bg-muted"}`} />
                        {isOnline ? "Online" : `Last seen ${timeAgo(d.last_seen_at)}`}
                      </p>
                    </div>
                  </div>
                  <button
                    onClick={() => revokeMutation.mutate(d.id)}
                    disabled={revokeMutation.isPending}
                    className="text-muted hover:text-danger transition-colors disabled:opacity-50"
                    aria-label={`Revoke ${d.device_name}`}
                  >
                    <Trash2 size={14} />
                  </button>
                </div>
              );
            })}
          </div>
        )}
      </section>

    </div>
  );
}
