"use client";

import { useState } from "react";
import { InviteModal } from "@/components/invite-modal";

interface Member {
  user_id: string;
  role: string;
  env_permissions: string[];
  joined_at: string;
}

export default function TeamPage() {
  const [inviteOpen, setInviteOpen] = useState(false);

  // Placeholder data — will be fetched via TanStack Query
  const team = null;
  const members: Member[] = [];

  const handleInvite = (email: string, role: string) => {
    // TODO: call api.inviteTeamMember()
    console.log("Invite:", email, role);
  };

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
            className="px-4 py-2 rounded-lg bg-accent text-background text-sm font-medium hover:bg-accent-dim transition-colors active:scale-[0.98]"
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

      {!team ? (
        <div className="rounded-xl border border-dashed border-border p-12 text-center text-muted">
          <span className="font-mono text-3xl block mb-3">⊕</span>
          <p className="text-sm font-medium">No team yet</p>
          <p className="text-xs mt-1">
            Create a team from CLI: <code className="font-mono text-accent">tene team create &quot;My Team&quot;</code>
          </p>
          <p className="text-xs mt-3">
            Requires <span className="badge badge-pro">Pro</span> plan. Each member needs their own Pro subscription.
          </p>
        </div>
      ) : (
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
              {members.length === 0 ? (
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
                      <button className="text-xs text-danger hover:underline" aria-label={`Remove ${m.user_id}`}>
                        Remove
                      </button>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      )}

      <div className="rounded-xl border border-border bg-surface p-4 text-xs text-muted space-y-1">
        <p>Project keys are shared via X25519 ECDH — the server never sees plaintext keys.</p>
        <p>Removing a member triggers automatic key rotation for all remaining members.</p>
      </div>

      <InviteModal
        teamId=""
        open={inviteOpen}
        onClose={() => setInviteOpen(false)}
        onInvite={handleInvite}
      />
    </div>
  );
}
