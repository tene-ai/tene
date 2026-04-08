"use client";

import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { useAuthReady } from "@/hooks/use-auth-ready";
import { api } from "@/lib/api";
import Link from "next/link";
import { Users, ArrowUpRight, Plus } from "lucide-react";

export default function TeamListPage() {
  const authReady = useAuthReady();
  const queryClient = useQueryClient();
  const [createOpen, setCreateOpen] = useState(false);

  const { data: teams, isLoading } = useQuery({
    queryKey: ["teams"],
    queryFn: () => api.listTeams(),
    enabled: authReady,
  });

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">Team</h1>
          <p className="text-muted text-sm mt-1">{teams?.length ?? 0} teams</p>
        </div>
        <button
          onClick={() => setCreateOpen(true)}
          className="flex items-center gap-2 px-4 py-2 rounded-lg bg-accent text-background text-sm font-medium hover:bg-accent-dim transition-colors active:scale-[0.98]"
        >
          <Plus size={14} />
          Create Team
        </button>
      </div>

      {isLoading ? (
        <div className="grid gap-4">
          {[1, 2].map((i) => (
            <div key={i} className="rounded-xl border border-border bg-surface p-5 animate-pulse">
              <div className="h-5 w-40 bg-surface-2 rounded mb-3" />
              <div className="h-3 w-60 bg-surface-2 rounded" />
            </div>
          ))}
        </div>
      ) : !teams?.length ? (
        <div className="rounded-xl border border-dashed border-border p-12 text-center">
          <Users size={32} className="mx-auto text-muted mb-3" />
          <p className="text-sm font-medium">No teams yet</p>
          <p className="text-xs text-muted mt-2">
            Create a team to share secrets with X25519 ECDH key wrapping.
          </p>
          <button
            onClick={() => setCreateOpen(true)}
            className="mt-4 px-4 py-2 rounded-lg bg-accent text-background text-sm font-medium hover:bg-accent-dim transition-colors"
          >
            Create Team
          </button>
        </div>
      ) : (
        <div className="grid gap-4">
          {teams.map((t) => (
            <Link
              key={t.id}
              href={`/team/${t.id}`}
              className="group rounded-xl border border-border bg-surface p-5 hover:border-accent/30 hover:bg-surface-2 transition-all"
            >
              <div className="flex items-start justify-between">
                <div className="flex items-center gap-3">
                  <Users size={18} className="text-accent shrink-0" />
                  <div>
                    <h2 className="font-semibold text-sm group-hover:text-accent transition-colors">
                      {t.name}
                    </h2>
                    <p className="text-xs text-muted mt-0.5 font-mono">{t.slug}</p>
                  </div>
                </div>
                <ArrowUpRight size={14} className="text-muted group-hover:text-accent transition-colors shrink-0 mt-1" />
              </div>
            </Link>
          ))}
        </div>
      )}

      {createOpen && (
        <CreateTeamModal
          onClose={() => setCreateOpen(false)}
          onCreated={() => {
            queryClient.invalidateQueries({ queryKey: ["teams"] });
            setCreateOpen(false);
          }}
        />
      )}
    </div>
  );
}

function CreateTeamModal({ onClose, onCreated }: { onClose: () => void; onCreated: () => void }) {
  const [name, setName] = useState("");
  const [slug, setSlug] = useState("");
  const [slugManuallyEdited, setSlugManuallyEdited] = useState(false);

  const createMutation = useMutation({
    mutationFn: () => api.createTeam(name, slug || name.toLowerCase().replace(/[^a-z0-9]+/g, "-")),
    onSuccess: () => onCreated(),
  });

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4" onClick={onClose}>
      <div className="absolute inset-0 bg-black/60 backdrop-blur-sm" />
      <div className="relative rounded-xl border border-border bg-surface p-6 w-full max-w-sm shadow-xl" onClick={(e) => e.stopPropagation()}>
        <h2 className="font-semibold mb-4">Create Team</h2>

        {createMutation.isError && (
          <div className="rounded-lg border border-danger/30 bg-danger/10 p-3 text-sm text-danger mb-4">
            {createMutation.error instanceof Error ? createMutation.error.message : "Failed to create team"}
          </div>
        )}

        <form onSubmit={(e) => { e.preventDefault(); if (name.trim()) createMutation.mutate(); }} className="space-y-4">
          <div>
            <label className="text-xs text-muted block mb-1">Team Name</label>
            <input
              type="text" value={name}
              onChange={(e) => {
                setName(e.target.value);
                if (!slugManuallyEdited) setSlug(e.target.value.toLowerCase().replace(/[^a-z0-9]+/g, "-").replace(/-+$/, ""));
              }}
              placeholder="My Startup"
              className="w-full px-3 py-2 rounded-lg bg-background border border-border text-sm focus:outline-none focus:ring-2 focus:ring-accent/50"
              required
            />
          </div>
          <div>
            <label className="text-xs text-muted block mb-1">Team Slug</label>
            <input
              type="text" value={slug}
              onChange={(e) => { setSlug(e.target.value); setSlugManuallyEdited(true); }}
              placeholder="my-startup"
              className="w-full px-3 py-2 rounded-lg bg-background border border-border text-sm font-mono focus:outline-none focus:ring-2 focus:ring-accent/50"
            />
          </div>
          <div className="flex gap-2">
            <button type="button" onClick={onClose} className="flex-1 py-2 rounded-lg border border-border text-sm text-muted hover:text-foreground transition-colors">Cancel</button>
            <button type="submit" disabled={createMutation.isPending || !name.trim()} className="flex-1 py-2 rounded-lg bg-accent text-background text-sm font-medium hover:bg-accent-dim transition-colors disabled:opacity-50">
              {createMutation.isPending ? "Creating..." : "Create"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
