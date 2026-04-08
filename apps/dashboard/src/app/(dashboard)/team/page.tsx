"use client";

import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { useAuthReady } from "@/hooks/use-auth-ready";
import { InviteModal } from "@/components/invite-modal";
import { api } from "@/lib/api";

export default function TeamPage() {
  const authReady = useAuthReady();
  const [inviteOpen, setInviteOpen] = useState(false);
  const queryClient = useQueryClient();

  const { data: teams, isLoading: teamsLoading } = useQuery({
    queryKey: ["teams"],
    queryFn: () => api.listTeams(),
    enabled: authReady,
  });

  const team = teams?.[0] ?? null;

  const { data: members, isLoading: membersLoading } = useQuery({
    queryKey: ["team-members", team?.id],
    queryFn: () => api.listTeamMembers(team!.id),
    enabled: authReady && !!team,
  });

  const inviteMutation = useMutation({
    mutationFn: ({ email, role }: { email: string; role: string }) =>
      api.inviteTeamMember(team!.id, email, role),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["team-members", team?.id] });
    },
  });

  const removeMutation = useMutation({
    mutationFn: (userId: string) => api.removeTeamMember(team!.id, userId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["team-members", team?.id] });
    },
  });

  const handleInvite = (email: string, role: string) => {
    inviteMutation.mutate({ email, role });
  };

  const handleRemove = (userId: string) => {
    if (confirm("Remove this member? This will trigger key rotation for all remaining members.")) {
      removeMutation.mutate(userId);
    }
  };

  if (teamsLoading) {
    return (
      <div className="space-y-6">
        <div>
          <h1 className="text-2xl font-bold">Team</h1>
          <p className="text-muted text-sm mt-1">Manage shared vault access</p>
        </div>
        <div className="rounded-xl border border-border bg-surface p-12 text-center text-muted animate-pulse">
          Loading teams...
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">Team</h1>
          <p className="text-muted text-sm mt-1">Manage shared vault access</p>
        </div>
        {team ? (
          <button
            onClick={() => setInviteOpen(true)}
            disabled={inviteMutation.isPending}
            className="px-4 py-2 rounded-lg bg-accent text-background text-sm font-medium hover:bg-accent-dim transition-colors active:scale-[0.98] disabled:opacity-50"
            aria-label="Invite team member"
          >
            + Invite
          </button>
        ) : (
          <div className="text-xs text-muted">
            Create a team with <code className="font-mono text-accent">tene team create</code>
          </div>
        )}
      </div>

      {inviteMutation.isError && (
        <div className="rounded-lg border border-danger/30 bg-danger/10 p-3 text-sm text-danger">
          {inviteMutation.error instanceof Error ? inviteMutation.error.message : "Failed to invite member"}
        </div>
      )}

      {removeMutation.isError && (
        <div className="rounded-lg border border-danger/30 bg-danger/10 p-3 text-sm text-danger">
          {removeMutation.error instanceof Error ? removeMutation.error.message : "Failed to remove member"}
        </div>
      )}

      {!team ? (
        <div className="rounded-xl border border-dashed border-border p-12 text-center text-muted">
          <span className="font-mono text-3xl block mb-3">&#x2295;</span>
          <p className="text-sm font-medium">No team yet</p>
          <p className="text-xs mt-1">
            Create a team from CLI: <code className="font-mono text-accent">tene team create &quot;My Team&quot;</code>
          </p>
          <p className="text-xs mt-3">
            Requires <span className="badge badge-pro">Pro</span> plan. Each member needs their own Pro subscription.
          </p>
        </div>
      ) : (
        <>
          <div className="rounded-xl border border-border bg-surface p-4">
            <div className="flex items-center gap-3">
              <span className="font-mono text-xl">&#x2295;</span>
              <div>
                <p className="font-semibold text-sm">{team.name}</p>
                <p className="text-xs text-muted font-mono">{team.slug}</p>
              </div>
            </div>
          </div>

          <div className="rounded-xl border border-border bg-surface overflow-hidden">
            <table className="w-full text-sm" role="table" aria-label="Team members">
              <thead>
                <tr className="border-b border-border text-left text-xs text-muted uppercase tracking-wider">
                  <th className="px-4 py-3" scope="col">Member</th>
                  <th className="px-4 py-3" scope="col">Role</th>
                  <th className="px-4 py-3 hidden sm:table-cell" scope="col">Environments</th>
                  <th className="px-4 py-3 hidden md:table-cell" scope="col">Joined</th>
                  <th className="px-4 py-3" scope="col">Actions</th>
                </tr>
              </thead>
              <tbody>
                {membersLoading ? (
                  <tr>
                    <td colSpan={5} className="px-4 py-8 text-center text-muted animate-pulse">
                      Loading members...
                    </td>
                  </tr>
                ) : !members || members.length === 0 ? (
                  <tr>
                    <td colSpan={5} className="px-4 py-8 text-center text-muted">
                      No members yet. Click &quot;+ Invite&quot; to add.
                    </td>
                  </tr>
                ) : (
                  members.map((m) => (
                    <tr key={m.user_id} className="border-b border-border last:border-0 hover:bg-surface-2 transition-colors">
                      <td className="px-4 py-3 font-mono text-sm">{m.user_id}</td>
                      <td className="px-4 py-3">
                        <span className={`badge ${m.role === "admin" ? "badge-pro" : "badge-free"}`}>
                          {m.role}
                        </span>
                      </td>
                      <td className="px-4 py-3 hidden sm:table-cell">
                        <div className="flex gap-1 flex-wrap">
                          {m.env_permissions.map((env) => (
                            <span key={env} className={`badge ${env === "prod" ? "badge-danger" : "badge-active"}`}>
                              {env}
                            </span>
                          ))}
                        </div>
                      </td>
                      <td className="px-4 py-3 hidden md:table-cell text-xs text-muted">
                        {new Date(m.joined_at).toLocaleDateString()}
                      </td>
                      <td className="px-4 py-3">
                        <button
                          onClick={() => handleRemove(m.user_id)}
                          disabled={removeMutation.isPending}
                          className="text-xs text-danger hover:underline disabled:opacity-50"
                          aria-label={`Remove ${m.user_id}`}
                        >
                          Remove
                        </button>
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        </>
      )}

      <div className="rounded-xl border border-border bg-surface p-4 text-xs text-muted space-y-1">
        <p>Project keys are shared via X25519 ECDH — the server never sees plaintext keys.</p>
        <p>Removing a member triggers automatic key rotation for all remaining members.</p>
      </div>

      <InviteModal
        teamId={team?.id ?? ""}
        open={inviteOpen}
        onClose={() => setInviteOpen(false)}
        onInvite={handleInvite}
      />
    </div>
  );
}
