"use client";

import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import { useQuery } from "@tanstack/react-query";
import { Command } from "cmdk";
import { useAuthReady } from "@/hooks/use-auth-ready";
import { api } from "@/lib/api";
import { FolderKey, Users, ScrollText, Settings, Copy, LogOut } from "lucide-react";
import { useAuthStore } from "@/lib/auth-store";

export function CommandPalette() {
  const [open, setOpen] = useState(false);
  const router = useRouter();
  const authReady = useAuthReady();
  const logout = useAuthStore((s) => s.logout);

  const { data: vaults } = useQuery({
    queryKey: ["vaults"],
    queryFn: () => api.listVaults(),
    enabled: authReady && open,
  });

  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if (e.key === "k" && (e.metaKey || e.ctrlKey)) {
        e.preventDefault();
        setOpen((o) => !o);
      }
    };
    window.addEventListener("keydown", handler);
    return () => window.removeEventListener("keydown", handler);
  }, []);

  const go = (path: string) => {
    router.push(path);
    setOpen(false);
  };

  const copy = (text: string) => {
    navigator.clipboard.writeText(text);
    setOpen(false);
  };

  if (!open) return (
    <div className="fixed bottom-24 right-5 lg:bottom-8 lg:right-8 z-40 group">
      <button
        onClick={() => setOpen(true)}
        className="w-11 h-11 rounded-full bg-accent shadow-lg shadow-accent/20 flex items-center justify-center hover:bg-accent-dim transition-all active:scale-95"
        aria-label="Open command palette"
      >
        <svg width="20" height="20" viewBox="0 0 32 32" fill="none">
          <path d="M6 22L14 14L6 6" stroke="#0a0a0a" strokeWidth="3.5" strokeLinecap="round" strokeLinejoin="round"/>
          <line x1="16" y1="24" x2="26" y2="24" stroke="#0a0a0a" strokeWidth="3.5" strokeLinecap="round"/>
        </svg>
      </button>
      <div className="absolute bottom-full right-0 mb-2 px-2.5 py-1 rounded-lg bg-surface border border-border text-xs text-muted whitespace-nowrap opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none">
        <kbd className="font-mono">⌘K</kbd>
      </div>
    </div>
  );

  return (
    <div className="fixed inset-0 z-50">
      <div className="absolute inset-0 bg-black/60 backdrop-blur-sm" onClick={() => setOpen(false)} />
      <div className="absolute inset-x-0 top-[20vh] flex justify-center px-4">
        <Command
          className="w-full max-w-lg rounded-xl border border-border bg-surface shadow-2xl overflow-hidden"
          label="Command palette"
        >
          <Command.Input
            placeholder="Search projects, commands..."
            className="w-full px-4 py-3 bg-transparent border-b border-border text-sm outline-none placeholder:text-muted"
          />
          <Command.List className="max-h-[300px] overflow-y-auto p-2">
            <Command.Empty className="px-4 py-8 text-center text-sm text-muted">
              No results found.
            </Command.Empty>

            {vaults && vaults.length > 0 && (
              <Command.Group heading="Projects" className="px-2 py-1 text-xs text-muted uppercase tracking-wider">
                {vaults.map((v) => (
                  <Command.Item
                    key={v.id}
                    value={v.project_name}
                    onSelect={() => go(`/projects/${v.id}`)}
                    className="flex items-center gap-3 px-3 py-2 rounded-lg text-sm cursor-pointer data-[selected=true]:bg-accent/10 data-[selected=true]:text-accent"
                  >
                    <FolderKey size={14} />
                    <span className="font-mono">{v.project_name}</span>
                  </Command.Item>
                ))}
              </Command.Group>
            )}

            <Command.Group heading="Navigation" className="px-2 py-1 text-xs text-muted uppercase tracking-wider">
              <Command.Item
                onSelect={() => go("/projects")}
                className="flex items-center gap-3 px-3 py-2 rounded-lg text-sm cursor-pointer data-[selected=true]:bg-accent/10 data-[selected=true]:text-accent"
              >
                <FolderKey size={14} /> Go to Projects
              </Command.Item>
              <Command.Item
                onSelect={() => go("/team")}
                className="flex items-center gap-3 px-3 py-2 rounded-lg text-sm cursor-pointer data-[selected=true]:bg-accent/10 data-[selected=true]:text-accent"
              >
                <Users size={14} /> Go to Team
              </Command.Item>
              <Command.Item
                onSelect={() => go("/activity")}
                className="flex items-center gap-3 px-3 py-2 rounded-lg text-sm cursor-pointer data-[selected=true]:bg-accent/10 data-[selected=true]:text-accent"
              >
                <ScrollText size={14} /> Go to Activity
              </Command.Item>
              <Command.Item
                onSelect={() => go("/settings")}
                className="flex items-center gap-3 px-3 py-2 rounded-lg text-sm cursor-pointer data-[selected=true]:bg-accent/10 data-[selected=true]:text-accent"
              >
                <Settings size={14} /> Go to Settings
              </Command.Item>
            </Command.Group>

            <Command.Group heading="CLI Commands" className="px-2 py-1 text-xs text-muted uppercase tracking-wider">
              <Command.Item
                onSelect={() => copy("tene push")}
                className="flex items-center gap-3 px-3 py-2 rounded-lg text-sm cursor-pointer data-[selected=true]:bg-accent/10 data-[selected=true]:text-accent"
              >
                <Copy size={14} /> Copy &apos;tene push&apos;
              </Command.Item>
              <Command.Item
                onSelect={() => copy("tene pull")}
                className="flex items-center gap-3 px-3 py-2 rounded-lg text-sm cursor-pointer data-[selected=true]:bg-accent/10 data-[selected=true]:text-accent"
              >
                <Copy size={14} /> Copy &apos;tene pull&apos;
              </Command.Item>
            </Command.Group>

            <Command.Group heading="Actions" className="px-2 py-1 text-xs text-muted uppercase tracking-wider">
              <Command.Item
                onSelect={() => { logout(); setOpen(false); }}
                className="flex items-center gap-3 px-3 py-2 rounded-lg text-sm cursor-pointer data-[selected=true]:bg-accent/10 data-[selected=true]:text-danger"
              >
                <LogOut size={14} /> Sign out
              </Command.Item>
            </Command.Group>
          </Command.List>

          <div className="border-t border-border px-4 py-2 text-xs text-muted flex justify-between">
            <span>Navigate with <kbd className="px-1 py-0.5 rounded bg-surface-2 font-mono text-[10px]">↑↓</kbd></span>
            <span><kbd className="px-1 py-0.5 rounded bg-surface-2 font-mono text-[10px]">esc</kbd> to close</span>
          </div>
        </Command>
      </div>
    </div>
  );
}
