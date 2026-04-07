import { clsx } from "clsx";

interface UsageBarProps {
  label: string;
  current: number;
  max: number;
  unit?: string;
}

export function UsageBar({ label, current, max, unit = "" }: UsageBarProps) {
  const pct = max > 0 ? Math.min((current / max) * 100, 100) : 0;
  const isWarning = pct >= 80;

  return (
    <div className="space-y-1.5">
      <div className="flex items-center justify-between text-xs">
        <span className="text-muted">{label}</span>
        <span className={clsx("font-mono", isWarning ? "text-warning" : "text-foreground")}>
          {current}{unit} / {max}{unit}
        </span>
      </div>
      <div className="h-1.5 rounded-full bg-surface-2 overflow-hidden" role="progressbar" aria-valuenow={current} aria-valuemin={0} aria-valuemax={max}>
        <div
          className={clsx(
            "h-full rounded-full transition-all duration-500",
            isWarning ? "bg-warning" : "bg-accent"
          )}
          style={{ width: `${pct}%` }}
        />
      </div>
    </div>
  );
}
