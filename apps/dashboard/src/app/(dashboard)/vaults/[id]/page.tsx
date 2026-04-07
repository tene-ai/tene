import { SecretKeyRow } from "@/components/secret-key-row";

interface VaultDetailPageProps {
  params: Promise<{ id: string }>;
}

export default async function VaultDetailPage({ params }: VaultDetailPageProps) {
  const { id } = await params;

  // Placeholder — will be fetched via TanStack Query
  const vault = {
    id,
    project_name: "my-project",
    vault_version: 3,
    secret_count: 0,
    environment: "default",
  };

  const secrets: { name: string; env: string; updated_at: string }[] = [];

  return (
    <div className="space-y-6">
      <div>
        <div className="flex items-center gap-2">
          <a href="/vaults" className="text-muted hover:text-foreground text-sm">&larr; Vaults</a>
          <span className="text-muted">/</span>
          <h1 className="text-2xl font-bold font-mono text-accent">{vault.project_name}</h1>
        </div>
        <div className="flex items-center gap-3 mt-2">
          <span className="badge badge-active">v{vault.vault_version}</span>
          <span className="text-xs text-muted">{vault.secret_count} secrets</span>
          <span className="text-xs text-muted">Env: {vault.environment}</span>
        </div>
      </div>

      <div className="rounded-xl border border-border bg-surface overflow-hidden">
        <table className="w-full text-sm" role="table" aria-label="Secret keys">
          <thead>
            <tr className="border-b border-border text-left text-xs text-muted uppercase tracking-wider">
              <th className="px-4 py-3" scope="col">Key</th>
              <th className="px-4 py-3" scope="col">Value</th>
              <th className="px-4 py-3 hidden sm:table-cell" scope="col">Env</th>
              <th className="px-4 py-3 hidden md:table-cell" scope="col">Updated</th>
              <th className="px-4 py-3" scope="col"></th>
            </tr>
          </thead>
          <tbody>
            {secrets.length === 0 ? (
              <tr>
                <td colSpan={5} className="px-4 py-12 text-center text-muted">
                  <span className="font-mono text-2xl block mb-2">◇</span>
                  No secrets in this vault yet.
                  <span className="block text-xs mt-1">
                    Run <code className="font-mono text-accent">tene set KEY VALUE</code> then <code className="font-mono text-accent">tene push</code>
                  </span>
                </td>
              </tr>
            ) : (
              secrets.map((s) => (
                <SecretKeyRow key={s.name + s.env} name={s.name} env={s.env} updatedAt={s.updated_at} />
              ))
            )}
          </tbody>
        </table>
      </div>

      <div className="rounded-xl border border-dashed border-border p-4 text-center text-xs text-muted space-y-1">
        <p>Secret values are <strong>never</strong> sent to the server or displayed in the dashboard.</p>
        <p>
          Values are encrypted with XChaCha20-Poly1305 on your device.
          Use <code className="font-mono text-accent">tene get KEY</code> in your CLI to access values.
        </p>
      </div>
    </div>
  );
}
