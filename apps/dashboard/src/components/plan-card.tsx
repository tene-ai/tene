import { clsx } from "clsx";

interface PlanCardProps {
  name: string;
  price: string;
  period: string;
  features: string[];
  current?: boolean;
  highlighted?: boolean;
  onUpgrade?: () => void;
}

export function PlanCard({ name, price, period, features, current, highlighted, onUpgrade }: PlanCardProps) {
  return (
    <div
      className={clsx(
        "rounded-xl border bg-surface p-6 relative overflow-hidden",
        highlighted ? "border-accent/30" : "border-border"
      )}
      role="article"
      aria-label={`${name} plan`}
    >
      {highlighted && (
        <div className="absolute inset-0 bg-gradient-to-br from-accent/5 to-transparent pointer-events-none" />
      )}
      <div className="relative">
        <div className="flex items-center justify-between mb-4">
          <h3 className="font-semibold">{name}</h3>
          <span className={`badge ${current ? "badge-active" : highlighted ? "badge-pro" : "badge-free"}`}>
            {current ? "Current" : highlighted ? "Upgrade" : ""}
          </span>
        </div>
        <p className="text-3xl font-bold font-mono mb-1">
          {price}<span className="text-sm text-muted font-normal">/{period}</span>
        </p>
        <ul className="mt-4 space-y-2 text-sm text-muted" aria-label={`${name} features`}>
          {features.map((f) => (
            <li key={f} className="flex items-center gap-2">
              <span className="text-accent" aria-hidden>✓</span>
              {f}
            </li>
          ))}
        </ul>
        {onUpgrade && (
          <button
            onClick={onUpgrade}
            className="mt-6 w-full py-2.5 rounded-lg bg-accent text-background font-medium text-sm hover:bg-accent-dim transition-colors active:scale-[0.98]"
            aria-label={`Upgrade to ${name}`}
          >
            Upgrade to {name}
          </button>
        )}
      </div>
    </div>
  );
}
