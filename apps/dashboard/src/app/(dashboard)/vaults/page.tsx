export default function VaultsPage() {
  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">Vaults</h1>
          <p className="text-muted text-sm mt-1">Your synced project vaults</p>
        </div>
      </div>

      <div className="rounded-xl border border-border bg-surface overflow-hidden">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-border text-left text-xs text-muted uppercase tracking-wider">
              <th className="px-4 py-3">Project</th>
              <th className="px-4 py-3">Secrets</th>
              <th className="px-4 py-3">Version</th>
              <th className="px-4 py-3 hidden sm:table-cell">Last Sync</th>
              <th className="px-4 py-3 hidden md:table-cell">Size</th>
            </tr>
          </thead>
          <tbody>
            <tr>
              <td colSpan={5} className="px-4 py-12 text-center text-muted">
                <span className="font-mono text-2xl block mb-2">◈</span>
                No vaults synced yet.
                <span className="block text-xs mt-1">
                  Run <code className="font-mono text-accent bg-surface-2 px-1 py-0.5 rounded">tene push</code> from your project directory.
                </span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <div className="rounded-xl border border-dashed border-border p-4 text-center text-xs text-muted">
        <p>Secret values are <strong>never</strong> visible here.</p>
        <p>Use <code className="font-mono text-accent">tene get KEY</code> in your CLI to access values.</p>
      </div>
    </div>
  );
}
