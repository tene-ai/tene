import { StatCard } from "@/components/stat-card";

export default function OverviewPage() {
  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-2xl font-bold">Overview</h1>
        <p className="text-muted text-sm mt-1">Your Tene Cloud at a glance</p>
      </div>

      <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
        <StatCard label="Vaults" value={0} sub="synced projects" accent />
        <StatCard label="Secrets" value={0} sub="total keys" />
        <StatCard label="Devices" value={0} sub="registered" />
        <StatCard label="Events" value={0} sub="this week" />
      </div>

      <div className="rounded-xl border border-border bg-surface p-6">
        <h2 className="text-sm font-semibold mb-4">Recent Activity</h2>
        <div className="flex flex-col items-center justify-center py-12 text-muted">
          <span className="font-mono text-3xl mb-3">◇</span>
          <p className="text-sm">No activity yet</p>
          <p className="text-xs mt-1">
            Run <code className="font-mono text-accent bg-surface-2 px-1.5 py-0.5 rounded">tene push</code> to sync your first vault
          </p>
        </div>
      </div>

      <div className="rounded-xl border border-border bg-surface p-6">
        <h2 className="text-sm font-semibold mb-2">Quick Start</h2>
        <div className="font-mono text-sm bg-background rounded-lg p-4 space-y-1 text-muted">
          <p><span className="text-accent">$</span> tene login</p>
          <p><span className="text-accent">$</span> tene push</p>
          <p><span className="text-accent">$</span> tene pull <span className="text-muted/50"># on another device</span></p>
        </div>
      </div>
    </div>
  );
}
