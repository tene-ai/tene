"use client";

import { useState, useMemo } from "react";
import { useQuery } from "@tanstack/react-query";
import { useAuthReady } from "@/hooks/use-auth-ready";
import { api, type Vault, type VaultKeysResponse, type TeamMember } from "@/lib/api";
import { OnboardingChecklist } from "@/components/onboarding-checklist";
import Link from "next/link";
import { FolderKey, Users, Key, Monitor, Search, ArrowUpDown, ChevronUp, ChevronDown } from "lucide-react";

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

export default function OverviewPage() {
  const authReady = useAuthReady();

  const { data: vaults } = useQuery({
    queryKey: ["vaults"],
    queryFn: () => api.listVaults(),
    enabled: authReady,
  });

  const { data: teams } = useQuery({
    queryKey: ["teams"],
    queryFn: () => api.listTeams(),
    enabled: authReady,
  });

  const { data: devices } = useQuery({
    queryKey: ["devices"],
    queryFn: () => api.listDevices(),
    enabled: authReady,
  });

  const { data: onboarding } = useQuery({
    queryKey: ["onboarding"],
    queryFn: () => api.getOnboardingStatus(),
    enabled: authReady,
  });

  // Fetch keys for all vaults
  const vaultIds = vaults?.map((v) => v.id) ?? [];
  const keysQueries = useQuery({
    queryKey: ["all-vault-keys", vaultIds],
    queryFn: async () => {
      if (!vaults?.length) return [];
      const results: { vault: Vault; keys: VaultKeysResponse }[] = [];
      for (const v of vaults) {
        try {
          const keys = await api.getVaultKeys(v.id);
          results.push({ vault: v, keys });
        } catch {
          // skip failed fetches
        }
      }
      return results;
    },
    enabled: authReady && !!vaults?.length,
  });

  // Fetch members for first team (for avatar display)
  const team = teams?.[0];
  const { data: members } = useQuery({
    queryKey: ["team-members", team?.id],
    queryFn: () => api.listTeamMembers(team!.id),
    enabled: authReady && !!team,
  });

  // Aggregate stats
  const projectCount = vaults?.length ?? 0;
  const totalSecrets = vaults?.reduce((sum, v) => sum + (v.secret_count || 0), 0) ?? 0;
  const memberCount = members?.length ?? 0;
  const deviceCount = devices?.length ?? 0;
  const totalStorage = vaults?.reduce((sum, v) => sum + (v.size || 0), 0) ?? 0;

  // Flatten all keys across all vaults for the table
  const allKeys: { projectName: string; vaultId: string; env: string; keyName: string; version: number; updatedAt: string }[] = [];
  if (keysQueries.data) {
    for (const { vault, keys } of keysQueries.data) {
      if (keys.keys) {
        for (const k of keys.keys) {
          allKeys.push({
            projectName: vault.project_name,
            vaultId: vault.id,
            env: keys.environment,
            keyName: k.name,
            version: k.version,
            updatedAt: k.updated_at,
          });
        }
      }
    }
  }

  return (
    <div className="space-y-6">
      {/* Onboarding */}
      {onboarding && !onboarding.completed && !onboarding.dismissed && (
        <OnboardingChecklist status={onboarding} />
      )}

      <div>
        <h1 className="text-2xl font-bold">Overview</h1>
        <p className="text-muted text-sm mt-1">Your Tene Cloud at a glance</p>
      </div>

      {/* Stats Grid */}
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
        <Link href="/projects" className="rounded-xl border border-border bg-surface p-5 hover:border-accent/30 transition-colors group">
          <div className="flex items-center gap-3">
            <div className="w-9 h-9 rounded-lg bg-accent/10 flex items-center justify-center">
              <FolderKey size={16} className="text-accent" />
            </div>
            <div>
              <p className="text-2xl font-bold font-mono">{projectCount}</p>
              <p className="text-xs text-muted">Projects</p>
            </div>
          </div>
        </Link>

        <div className="rounded-xl border border-border bg-surface p-5">
          <div className="flex items-center gap-3">
            <div className="w-9 h-9 rounded-lg bg-accent/10 flex items-center justify-center">
              <Key size={16} className="text-accent" />
            </div>
            <div>
              <p className="text-2xl font-bold font-mono">{totalSecrets}</p>
              <p className="text-xs text-muted">Secrets</p>
            </div>
          </div>
        </div>

        <Link href="/team" className="rounded-xl border border-border bg-surface p-5 hover:border-accent/30 transition-colors group">
          <div className="flex items-center gap-3">
            <div className="w-9 h-9 rounded-lg bg-accent/10 flex items-center justify-center">
              <Users size={16} className="text-accent" />
            </div>
            <div>
              <p className="text-2xl font-bold font-mono">{memberCount}</p>
              <p className="text-xs text-muted">Members</p>
            </div>
          </div>
        </Link>

        <Link href="/settings" className="rounded-xl border border-border bg-surface p-5 hover:border-accent/30 transition-colors group">
          <div className="flex items-center gap-3">
            <div className="w-9 h-9 rounded-lg bg-accent/10 flex items-center justify-center">
              <Monitor size={16} className="text-accent" />
            </div>
            <div>
              <p className="text-2xl font-bold font-mono">{deviceCount}</p>
              <p className="text-xs text-muted">Devices</p>
            </div>
          </div>
        </Link>
      </div>

      {/* Storage bar */}
      <div className="rounded-xl border border-border bg-surface p-4 flex items-center justify-between">
        <span className="text-sm text-muted">Cloud Storage</span>
        <span className="font-mono text-sm font-medium">{formatBytes(totalStorage)}</span>
      </div>

      {/* All Keys Table */}
      <AllSecretsTable allKeys={allKeys} members={members ?? []} />

      {/* ZK Notice */}
      <div className="rounded-xl border border-dashed border-border p-4 text-center text-xs text-muted">
        <p>Secret values are <strong>never</strong> visible here. Zero-Knowledge: XChaCha20-Poly1305 encryption.</p>
      </div>
    </div>
  );
}

type KeyRow = { projectName: string; vaultId: string; env: string; keyName: string; version: number; updatedAt: string };
type SortField = "keyName" | "projectName" | "env" | "version" | "updatedAt";
type SortDir = "asc" | "desc";

function AllSecretsTable({ allKeys, members }: { allKeys: KeyRow[]; members: TeamMember[] }) {
  const [search, setSearch] = useState("");
  const [sortField, setSortField] = useState<SortField>("keyName");
  const [sortDir, setSortDir] = useState<SortDir>("asc");

  const toggleSort = (field: SortField) => {
    if (sortField === field) {
      setSortDir((d) => (d === "asc" ? "desc" : "asc"));
    } else {
      setSortField(field);
      setSortDir("asc");
    }
  };

  const filtered = useMemo(() => {
    const q = search.toLowerCase();
    let result = allKeys;
    if (q) {
      result = allKeys.filter(
        (k) =>
          k.keyName.toLowerCase().includes(q) ||
          k.projectName.toLowerCase().includes(q) ||
          k.env.toLowerCase().includes(q)
      );
    }
    return [...result].sort((a, b) => {
      let cmp = 0;
      switch (sortField) {
        case "keyName": cmp = a.keyName.localeCompare(b.keyName); break;
        case "projectName": cmp = a.projectName.localeCompare(b.projectName); break;
        case "env": cmp = a.env.localeCompare(b.env); break;
        case "version": cmp = a.version - b.version; break;
        case "updatedAt": cmp = new Date(a.updatedAt).getTime() - new Date(b.updatedAt).getTime(); break;
      }
      return sortDir === "asc" ? cmp : -cmp;
    });
  }, [allKeys, search, sortField, sortDir]);

  const SortIcon = ({ field }: { field: SortField }) => {
    if (sortField !== field) return <ArrowUpDown size={10} className="text-muted/40" />;
    return sortDir === "asc" ? <ChevronUp size={10} className="text-accent" /> : <ChevronDown size={10} className="text-accent" />;
  };

  return (
    <div>
      <div className="flex items-center justify-between mb-3">
        <h2 className="text-sm font-semibold">All Secrets</h2>
        <div className="relative">
          <Search size={14} className="absolute left-2.5 top-1/2 -translate-y-1/2 text-muted" />
          <input
            type="text"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Search keys..."
            className="pl-8 pr-3 py-1.5 rounded-lg bg-background border border-border text-xs w-48 focus:outline-none focus:ring-1 focus:ring-accent/50 placeholder:text-muted"
          />
        </div>
      </div>
      <div className="rounded-xl border border-border bg-surface overflow-hidden">
        <table className="w-full text-sm" role="table" aria-label="All secrets overview">
          <thead>
            <tr className="border-b border-border text-left text-xs text-muted uppercase tracking-wider">
              <th className="px-4 py-3 cursor-pointer select-none" scope="col" onClick={() => toggleSort("keyName")}>
                <span className="flex items-center gap-1">Key <SortIcon field="keyName" /></span>
              </th>
              <th className="px-4 py-3 cursor-pointer select-none" scope="col" onClick={() => toggleSort("projectName")}>
                <span className="flex items-center gap-1">Project <SortIcon field="projectName" /></span>
              </th>
              <th className="px-4 py-3 hidden sm:table-cell cursor-pointer select-none" scope="col" onClick={() => toggleSort("env")}>
                <span className="flex items-center gap-1">Env <SortIcon field="env" /></span>
              </th>
              <th className="px-4 py-3 hidden md:table-cell cursor-pointer select-none" scope="col" onClick={() => toggleSort("version")}>
                <span className="flex items-center gap-1">Version <SortIcon field="version" /></span>
              </th>
              <th className="px-4 py-3 hidden md:table-cell cursor-pointer select-none" scope="col" onClick={() => toggleSort("updatedAt")}>
                <span className="flex items-center gap-1">Updated <SortIcon field="updatedAt" /></span>
              </th>
              <th className="px-4 py-3" scope="col">Shared with</th>
            </tr>
          </thead>
          <tbody>
            {!filtered.length ? (
              <tr>
                <td colSpan={6} className="px-4 py-12 text-center text-muted">
                  {search ? (
                    <p className="text-sm">No secrets matching &quot;{search}&quot;</p>
                  ) : (
                    <>
                      <Key size={24} className="mx-auto mb-2 text-muted" />
                      <p className="text-sm">No secrets synced yet</p>
                      <p className="text-xs mt-1">Run <code className="font-mono text-accent">tene push</code> to sync</p>
                    </>
                  )}
                </td>
              </tr>
            ) : (
              filtered.map((k, i) => (
                <tr key={`${k.vaultId}-${k.env}-${k.keyName}-${i}`} className="border-b border-border last:border-0 hover:bg-surface-2 transition-colors">
                  <td className="px-4 py-3 font-mono text-sm">{k.keyName}</td>
                  <td className="px-4 py-3">
                    <Link href={`/projects/${k.vaultId}`} className="text-sm text-muted hover:text-accent transition-colors font-mono">
                      {k.projectName}
                    </Link>
                  </td>
                  <td className="px-4 py-3 hidden sm:table-cell">
                    <span className="badge badge-active">{k.env}</span>
                  </td>
                  <td className="px-4 py-3 hidden md:table-cell">
                    <span className="badge badge-active">v{k.version}</span>
                  </td>
                  <td className="px-4 py-3 hidden md:table-cell text-xs text-muted">
                    {timeAgo(k.updatedAt)}
                  </td>
                  <td className="px-4 py-3">
                    <MemberAvatars members={members} />
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
      {search && filtered.length > 0 && (
        <p className="text-xs text-muted mt-2">{filtered.length} of {allKeys.length} secrets</p>
      )}
    </div>
  );
}

function MemberAvatars({ members }: { members: TeamMember[] }) {
  if (!members.length) {
    return <span className="text-xs text-muted">Only you</span>;
  }

  const shown = members.slice(0, 3);
  const extra = members.length - 3;

  return (
    <div className="flex -space-x-1.5">
      {shown.map((m) => (
        m.avatar_url ? (
          // eslint-disable-next-line @next/next/no-img-element
          <img
            key={m.user_id}
            src={m.avatar_url}
            alt={m.name || ""}
            title={m.name || m.user_id.slice(0, 8)}
            className="w-5 h-5 rounded-full border border-background"
          />
        ) : (
          <div
            key={m.user_id}
            title={m.name || m.user_id.slice(0, 8)}
            className="w-5 h-5 rounded-full border border-background bg-accent/20 flex items-center justify-center text-accent text-[8px] font-bold"
          >
            {(m.name || "?")[0]?.toUpperCase()}
          </div>
        )
      ))}
      {extra > 0 && (
        <div className="w-5 h-5 rounded-full border border-background bg-surface-2 flex items-center justify-center text-[8px] text-muted font-medium">
          +{extra}
        </div>
      )}
    </div>
  );
}
