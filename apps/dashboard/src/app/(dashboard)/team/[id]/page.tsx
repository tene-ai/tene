"use client";

import { use, useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { useAuthReady } from "@/hooks/use-auth-ready";
import { api } from "@/lib/api";
import Link from "next/link";
import { UserPlus, Shield, Trash2 } from "lucide-react";

function formatDate(date: string): string {
  if (!date) return "-";
  const d = new Date(date);
  if (isNaN(d.getTime())) return "-";
  return d.toLocaleDateString();
}

interface TeamDetailPageProps {
  params: Promise<{ id: string }>;
}

export default function TeamDetailPage({ params }: TeamDetailPageProps) {
  const { id } = use(params);
  const authReady = useAuthReady();
  const queryClient = useQueryClient();
  const [inviteOpen, setInviteOpen] = useState(false);

  const { data: teams } = useQuery({
    queryKey: ["teams"],
    queryFn: () => api.listTeams(),
    enabled: authReady,
  });

  const team = teams?.find((t) => t.id === id);

  const { data: members, isLoading: membersLoading } = useQuery({
    queryKey: ["team-members", id],
    queryFn: () => api.listTeamMembers(id),
    enabled: authReady && !!team,
  });

  const removeMutation = useMutation({
    mutationFn: (userId: string) => api.removeTeamMember(id, userId),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["team-members", id] }),
  });

  if (!team && !membersLoading) {
    return (
      <div className="space-y-6">
        <nav className="flex items-center gap-2 text-sm text-muted">
          <Link href="/team" className="hover:text-foreground">Team</Link>
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
        <Link href="/team" className="hover:text-foreground">Team</Link>
        <span>/</span>
        <span className="text-foreground">{team?.name}</span>
      </nav>

      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">{team?.name}</h1>
          <p className="text-muted text-sm mt-1">
            <span className="font-mono">{team?.slug}</span> · {members?.length ?? 0} members
          </p>
        </div>
        <button
          onClick={() => setInviteOpen(true)}
          className="flex items-center gap-2 px-4 py-2 rounded-lg bg-accent text-background text-sm font-medium hover:bg-accent-dim transition-colors active:scale-[0.98]"
        >
          <UserPlus size={14} />
          Invite
        </button>
      </div>

      {removeMutation.isError && (
        <div className="rounded-lg border border-danger/30 bg-danger/10 p-3 text-sm text-danger">
          {removeMutation.error instanceof Error ? removeMutation.error.message : "Failed to remove member"}
        </div>
      )}

      {/* Members Table */}
      <div className="rounded-xl border border-border bg-surface overflow-hidden">
        <table className="w-full text-sm" role="table" aria-label="Team members">
          <thead>
            <tr className="border-b border-border text-left text-xs text-muted uppercase tracking-wider">
              <th className="px-4 py-3" scope="col">Member</th>
              <th className="px-4 py-3" scope="col">Role</th>
              <th className="px-4 py-3 hidden sm:table-cell" scope="col">Environments</th>
              <th className="px-4 py-3 hidden md:table-cell" scope="col">Joined</th>
              <th className="px-4 py-3 w-16" scope="col"></th>
            </tr>
          </thead>
          <tbody>
            {membersLoading ? (
              <tr>
                <td colSpan={5} className="px-4 py-8 text-center text-muted animate-pulse">Loading...</td>
              </tr>
            ) : !members?.length ? (
              <tr>
                <td colSpan={5} className="px-4 py-8 text-center text-muted">
                  No members yet. Click &quot;Invite&quot; to add.
                </td>
              </tr>
            ) : (
              members.map((m) => (
                <tr key={m.user_id} className="border-b border-border last:border-0 hover:bg-surface-2 transition-colors">
                  <td className="px-4 py-3">
                    <div className="flex items-center gap-2.5">
                      {m.avatar_url ? (
                        // eslint-disable-next-line @next/next/no-img-element
                        <img src={m.avatar_url} alt="" className="w-6 h-6 rounded-full" />
                      ) : (
                        <div className="w-6 h-6 rounded-full bg-accent/20 flex items-center justify-center text-accent text-xs font-bold">
                          {(m.name || "?")[0]?.toUpperCase()}
                        </div>
                      )}
                      <div>
                        <p className="text-sm font-medium">{m.name || m.user_id.slice(0, 8) + "..."}</p>
                        {m.name && <p className="text-[10px] text-muted font-mono">{m.user_id.slice(0, 8)}...</p>}
                      </div>
                    </div>
                  </td>
                  <td className="px-4 py-3">
                    <span className={`badge ${m.role === "admin" ? "badge-pro" : "badge-free"}`}>
                      {m.role === "admin" && <Shield size={10} className="inline mr-1" />}
                      {m.role}
                    </span>
                  </td>
                  <td className="px-4 py-3 hidden sm:table-cell">
                    <div className="flex gap-1 flex-wrap">
                      {(m.env_permissions?.length ? m.env_permissions : ["all"]).map((env) => (
                        <span key={env} className={`badge ${env === "prod" ? "badge-danger" : env === "*" || env === "all" ? "badge-pro" : "badge-active"}`}>
                          {env === "*" ? "all" : env}
                        </span>
                      ))}
                    </div>
                  </td>
                  <td className="px-4 py-3 hidden md:table-cell text-xs text-muted">
                    {formatDate(m.joined_at)}
                  </td>
                  <td className="px-4 py-3">
                    {m.role !== "admin" && (
                      <button
                        onClick={() => removeMutation.mutate(m.user_id)}
                        disabled={removeMutation.isPending}
                        className="text-muted hover:text-danger transition-colors disabled:opacity-50"
                        aria-label="Remove member"
                      >
                        <Trash2 size={14} />
                      </button>
                    )}
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {/* ZK Notice */}
      <div className="rounded-xl border border-dashed border-border p-4 text-xs text-muted text-center space-y-1">
        <p>Project keys shared via X25519 ECDH — the server never sees plaintext keys.</p>
        <p>Removing a member triggers automatic key rotation.</p>
        <p className="text-muted/60">Environments control which secret environments (dev/staging/prod) a member can access.</p>
      </div>

      {inviteOpen && (
        <InviteModal teamId={id} onClose={() => setInviteOpen(false)} />
      )}
    </div>
  );
}

function InviteModal({ teamId, onClose }: { teamId: string; onClose: () => void }) {
  const queryClient = useQueryClient();
  const authReady = useAuthReady();
  const [userId, setUserId] = useState("");
  const [role, setRole] = useState("member");
  const [envs, setEnvs] = useState<Set<string>>(new Set());
  const [envsInitialized, setEnvsInitialized] = useState(false);

  // Fetch all vaults to get available environments dynamically
  const { data: vaults } = useQuery({
    queryKey: ["vaults"],
    queryFn: () => api.listVaults(),
    enabled: authReady,
  });

  // Fetch keys for each vault to collect all environments
  const { data: allEnvs } = useQuery({
    queryKey: ["all-envs", vaults?.map((v) => v.id)],
    queryFn: async () => {
      if (!vaults?.length) return [];
      const envSet = new Set<string>();
      for (const v of vaults) {
        try {
          const keys = await api.getVaultKeys(v.id);
          keys.environments?.forEach((e) => envSet.add(e));
        } catch { /* skip */ }
      }
      return Array.from(envSet).sort();
    },
    enabled: authReady && !!vaults?.length,
  });

  // Auto-select non-prod environments on first load
  if (allEnvs?.length && !envsInitialized) {
    setEnvs(new Set(allEnvs.filter((e) => e !== "prod")));
    setEnvsInitialized(true);
  }

  const toggleEnv = (env: string) => {
    setEnvs((prev) => {
      const next = new Set(prev);
      if (next.has(env)) next.delete(env);
      else next.add(env);
      return next;
    });
  };

  const envPermissions = role === "admin" ? ["*"] : Array.from(envs);

  const inviteMutation = useMutation({
    mutationFn: () => api.inviteTeamMember(teamId, userId.trim(), role, envPermissions),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["team-members", teamId] });
      onClose();
    },
  });

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4" onClick={onClose}>
      <div className="absolute inset-0 bg-black/60 backdrop-blur-sm" />
      <div className="relative rounded-xl border border-border bg-surface p-6 w-full max-w-sm shadow-xl" onClick={(e) => e.stopPropagation()}>
        <h2 className="font-semibold mb-1">Invite Team Member</h2>
        <p className="text-xs text-muted mb-4">
          The member must have a Tene Cloud account.
          They can copy their User ID from the profile menu in Dashboard.
        </p>

        {inviteMutation.isError && (
          <div className="rounded-lg border border-danger/30 bg-danger/10 p-3 text-sm text-danger mb-4">
            {inviteMutation.error instanceof Error ? inviteMutation.error.message : "Failed to invite"}
          </div>
        )}

        <form onSubmit={(e) => { e.preventDefault(); if (userId.trim()) inviteMutation.mutate(); }} className="space-y-4">
          <div>
            <label className="text-xs text-muted block mb-1">User ID</label>
            <input
              type="text" value={userId} onChange={(e) => setUserId(e.target.value)}
              placeholder="e.g. c1b1037f-..."
              className="w-full px-3 py-2 rounded-lg bg-background border border-border text-sm font-mono focus:outline-none focus:ring-2 focus:ring-accent/50"
              required
            />
          </div>
          <div>
            <label className="text-xs text-muted block mb-1.5">Role</label>
            <div className="flex gap-2">
              {(["member", "admin"] as const).map((r) => (
                <button
                  key={r} type="button" onClick={() => setRole(r)}
                  className={`flex-1 py-2 rounded-lg text-sm font-medium border transition-colors ${
                    role === r ? "border-accent/40 bg-accent/10 text-accent" : "border-border bg-background text-muted hover:text-foreground"
                  }`}
                >
                  {r === "admin" ? "Admin" : "Member"}
                </button>
              ))}
            </div>
          </div>
          <div>
            <label className="text-xs text-muted block mb-1.5">
              Environments
              {role === "admin" && <span className="text-accent ml-1">(admin has all access)</span>}
            </label>
            {!allEnvs?.length ? (
              <p className="text-xs text-muted">No environments found. Push a vault first.</p>
            ) : (
              <div className="flex gap-2 flex-wrap">
                {allEnvs.map((env) => (
                  <button
                    key={env} type="button"
                    disabled={role === "admin"}
                    onClick={() => toggleEnv(env)}
                    className={`px-3 py-2 rounded-lg text-sm font-mono border transition-colors ${
                      role === "admin" || envs.has(env)
                        ? env === "prod" || env === "production"
                          ? "border-danger/40 bg-danger/10 text-danger"
                          : "border-accent/40 bg-accent/10 text-accent"
                        : "border-border bg-background text-muted hover:text-foreground"
                    } disabled:opacity-60`}
                  >
                    {env}
                  </button>
                ))}
              </div>
            )}
          </div>
          <div className="flex gap-2">
            <button type="button" onClick={onClose} className="flex-1 py-2 rounded-lg border border-border text-sm text-muted hover:text-foreground transition-colors">Cancel</button>
            <button
              type="submit"
              disabled={inviteMutation.isPending || !userId.trim() || (role !== "admin" && envs.size === 0)}
              className="flex-1 py-2 rounded-lg bg-accent text-background text-sm font-medium hover:bg-accent-dim transition-colors disabled:opacity-50"
            >
              {inviteMutation.isPending ? "Inviting..." : "Invite Member"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
