interface Vault {
  id: string;
  project_name: string;
  secret_count: number;
  vault_version: number;
  updated_at: string;
  size: number;
}

interface VaultTableProps {
  vaults: Vault[];
}

export function VaultTable({ vaults }: VaultTableProps) {
  if (vaults.length === 0) {
    return (
      <div className="rounded-xl border border-border bg-surface p-12 text-center text-muted">
        <span className="font-mono text-2xl block mb-2">◈</span>
        No vaults synced yet.
        <span className="block text-xs mt-1">
          Run <code className="font-mono text-accent bg-surface-2 px-1 py-0.5 rounded">tene push</code> from your project directory.
        </span>
      </div>
    );
  }

  return (
    <div className="rounded-xl border border-border bg-surface overflow-hidden">
      <table className="w-full text-sm" role="table" aria-label="Vault list">
        <thead>
          <tr className="border-b border-border text-left text-xs text-muted uppercase tracking-wider">
            <th className="px-4 py-3" scope="col">Project</th>
            <th className="px-4 py-3" scope="col">Secrets</th>
            <th className="px-4 py-3" scope="col">Version</th>
            <th className="px-4 py-3 hidden sm:table-cell" scope="col">Last Sync</th>
            <th className="px-4 py-3 hidden md:table-cell" scope="col">Size</th>
          </tr>
        </thead>
        <tbody>
          {vaults.map((v) => (
            <tr key={v.id} className="border-b border-border last:border-0 hover:bg-surface-2 transition-colors">
              <td className="px-4 py-3 font-mono text-accent">{v.project_name}</td>
              <td className="px-4 py-3">{v.secret_count} keys</td>
              <td className="px-4 py-3 font-mono">v{v.vault_version}</td>
              <td className="px-4 py-3 hidden sm:table-cell text-muted">{new Date(v.updated_at).toLocaleDateString()}</td>
              <td className="px-4 py-3 hidden md:table-cell text-muted">{formatBytes(v.size)}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

function formatBytes(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}
