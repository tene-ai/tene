"use client";

import { useState, useEffect, useCallback } from "react";

interface InviteModalProps {
  teamId: string;
  open: boolean;
  onClose: () => void;
  onInvite: (email: string, role: string) => void;
}

export function InviteModal({ teamId, open, onClose, onInvite }: InviteModalProps) {
  const [email, setEmail] = useState("");
  const [role, setRole] = useState("member");
  const [error, setError] = useState("");

  const handleEscape = useCallback((e: KeyboardEvent) => {
    if (e.key === "Escape") onClose();
  }, [onClose]);

  useEffect(() => {
    if (open) {
      document.addEventListener("keydown", handleEscape);
      return () => document.removeEventListener("keydown", handleEscape);
    }
  }, [open, handleEscape]);

  if (!open) return null;

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    setError("");

    if (!email || !email.includes("@")) {
      setError("Valid email required");
      return;
    }

    onInvite(email, role);
    setEmail("");
    setRole("member");
    onClose();
  };

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm"
      onClick={onClose}
      role="dialog"
      aria-modal="true"
      aria-label="Invite team member"
    >
      <div
        className="w-full max-w-md rounded-xl border border-border bg-surface p-6 shadow-2xl shadow-black/40"
        onClick={(e) => e.stopPropagation()}
      >
        <h2 className="text-lg font-semibold mb-4">Invite Member</h2>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label htmlFor="invite-email" className="block text-xs text-muted mb-1">Email</label>
            <input
              id="invite-email"
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder="alice@example.com"
              className="w-full px-3 py-2 rounded-lg border border-border bg-background text-sm focus:border-accent focus:outline-none"
              autoFocus
            />
          </div>

          <div>
            <label htmlFor="invite-role" className="block text-xs text-muted mb-1">Role</label>
            <select
              id="invite-role"
              value={role}
              onChange={(e) => setRole(e.target.value)}
              className="w-full px-3 py-2 rounded-lg border border-border bg-background text-sm focus:border-accent focus:outline-none"
            >
              <option value="member">Member — read/write dev</option>
              <option value="admin">Admin — full access all envs</option>
            </select>
          </div>

          {error && <p className="text-xs text-danger">{error}</p>}

          <div className="flex gap-3 pt-2">
            <button
              type="button"
              onClick={onClose}
              className="flex-1 py-2 rounded-lg border border-border text-sm text-muted hover:bg-surface-2 transition-colors"
            >
              Cancel
            </button>
            <button
              type="submit"
              className="flex-1 py-2 rounded-lg bg-accent text-background text-sm font-medium hover:bg-accent-dim transition-colors active:scale-[0.98]"
            >
              Send Invite
            </button>
          </div>

          <p className="text-[10px] text-muted text-center">
            Project key will be wrapped with their X25519 public key (zero-knowledge).
          </p>
        </form>
      </div>
    </div>
  );
}
