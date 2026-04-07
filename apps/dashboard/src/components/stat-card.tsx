import { clsx } from "clsx";

interface StatCardProps {
  label: string;
  value: string | number;
  sub?: string;
  accent?: boolean;
}

export function StatCard({ label, value, sub, accent }: StatCardProps) {
  return (
    <div className="rounded-xl border border-border bg-surface p-5 transition-colors hover:border-accent/20">
      <p className="text-xs text-muted uppercase tracking-wider mb-1">{label}</p>
      <p className={clsx("text-2xl font-bold font-mono", accent && "text-accent")}>{value}</p>
      {sub && <p className="text-xs text-muted mt-1">{sub}</p>}
    </div>
  );
}
