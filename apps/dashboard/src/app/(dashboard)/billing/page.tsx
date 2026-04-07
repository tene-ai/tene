import { UsageBar } from "@/components/usage-bar";

export default function BillingPage() {
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold">Billing</h1>
        <p className="text-muted text-sm mt-1">Manage your subscription</p>
      </div>

      <div className="grid sm:grid-cols-2 gap-4">
        {/* Free Plan */}
        <div className="rounded-xl border border-border bg-surface p-6">
          <div className="flex items-center justify-between mb-4">
            <h3 className="font-semibold">Free</h3>
            <span className="badge badge-free">Current</span>
          </div>
          <p className="text-3xl font-bold font-mono mb-1">$0<span className="text-sm text-muted font-normal">/forever</span></p>
          <ul className="mt-4 space-y-2 text-sm text-muted">
            <li className="flex items-center gap-2"><span className="text-accent">✓</span> Local secret management</li>
            <li className="flex items-center gap-2"><span className="text-accent">✓</span> AI agent auto-detection</li>
            <li className="flex items-center gap-2"><span className="text-accent">✓</span> 12-word recovery key</li>
            <li className="flex items-center gap-2"><span className="text-accent">✓</span> XChaCha20-Poly1305 encryption</li>
          </ul>
        </div>

        {/* Pro Plan */}
        <div className="rounded-xl border border-accent/30 bg-surface p-6 relative overflow-hidden">
          <div className="absolute inset-0 bg-gradient-to-br from-accent/5 to-transparent pointer-events-none" />
          <div className="relative">
            <div className="flex items-center justify-between mb-4">
              <h3 className="font-semibold">Pro</h3>
              <span className="badge badge-pro">Upgrade</span>
            </div>
            <p className="text-3xl font-bold font-mono mb-1">$5<span className="text-sm text-muted font-normal">/month</span></p>
            <ul className="mt-4 space-y-2 text-sm text-muted">
              <li className="flex items-center gap-2"><span className="text-accent">✓</span> Everything in Free</li>
              <li className="flex items-center gap-2"><span className="text-accent">✓</span> Vault cloud sync (push/pull)</li>
              <li className="flex items-center gap-2"><span className="text-accent">✓</span> Encrypted cloud backup</li>
              <li className="flex items-center gap-2"><span className="text-accent">✓</span> Device management</li>
              <li className="flex items-center gap-2"><span className="text-accent">✓</span> Team secret sharing + RBAC</li>
              <li className="flex items-center gap-2"><span className="text-accent">✓</span> Full audit log</li>
            </ul>
            <button className="mt-6 w-full py-2.5 rounded-lg bg-accent text-background font-medium text-sm hover:bg-accent-dim transition-colors active:scale-[0.98]">
              Upgrade to Pro
            </button>
          </div>
        </div>
      </div>

      <div className="rounded-xl border border-border bg-surface p-5 space-y-4">
        <h3 className="text-sm font-semibold">Usage</h3>
        <UsageBar label="Vaults" current={0} max={50} />
        <UsageBar label="Team Members" current={0} max={10} />
        <UsageBar label="Storage" current={0} max={100} unit="MB" />
      </div>

      <div className="rounded-xl border border-border bg-surface p-4 text-xs text-muted text-center">
        Payments processed securely by LemonSqueezy. Cancel anytime.
      </div>
    </div>
  );
}
