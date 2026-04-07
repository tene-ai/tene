export default function AuditPage() {
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold">Audit Log</h1>
        <p className="text-muted text-sm mt-1">Security activity across your account</p>
      </div>

      <div className="flex gap-2 flex-wrap">
        {["All", "Push", "Pull", "Login", "Delete"].map((filter) => (
          <button
            key={filter}
            className={`px-3 py-1.5 rounded-lg text-xs border transition-colors ${
              filter === "All"
                ? "border-accent/30 bg-accent/10 text-accent"
                : "border-border bg-surface text-muted hover:text-foreground"
            }`}
          >
            {filter}
          </button>
        ))}
      </div>

      <div className="rounded-xl border border-border bg-surface overflow-hidden">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-border text-left text-xs text-muted uppercase tracking-wider">
              <th className="px-4 py-3">Time</th>
              <th className="px-4 py-3">Action</th>
              <th className="px-4 py-3 hidden sm:table-cell">Target</th>
              <th className="px-4 py-3 hidden md:table-cell">IP</th>
            </tr>
          </thead>
          <tbody>
            <tr>
              <td colSpan={4} className="px-4 py-12 text-center text-muted">
                <span className="font-mono text-2xl block mb-2">☰</span>
                No audit events yet
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  );
}
