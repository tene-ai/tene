"use client";

import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api";
import { VaultTable } from "@/components/vault-table";
import { useAuthReady } from "@/hooks/use-auth-ready";

export default function VaultsPage() {
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
          <h1 className="text-2xl font-bold">Vaults</h1>
          <p className="text-muted text-sm mt-1">Your synced project vaults</p>
        </div>
      </div>

      {isLoading ? (
        <div className="rounded-xl border border-border bg-surface p-12 text-center text-muted">
          <span className="font-mono text-2xl block mb-2 animate-pulse">◈</span>
          Loading vaults...
        </div>
      ) : (
        <VaultTable vaults={vaults ?? []} />
      )}

      <div className="rounded-xl border border-dashed border-border p-4 text-center text-xs text-muted">
        <p>Secret values are <strong>never</strong> visible here.</p>
        <p>Use <code className="font-mono text-accent">tene get KEY</code> in your CLI to access values.</p>
      </div>
    </div>
  );
}
