// Design Ref: §2.5 — Inline callout box for MDX. Three tones.
type Tone = "info" | "warn" | "tip";

const TONE_STYLES: Record<Tone, string> = {
  info: "border-accent/40 bg-accent/5 text-foreground",
  warn: "border-red-500/40 bg-red-500/5 text-foreground",
  tip: "border-yellow-400/40 bg-yellow-400/5 text-foreground",
};

const TONE_ICONS: Record<Tone, string> = {
  info: "ℹ",
  warn: "⚠",
  tip: "★",
};

export function Callout({
  type = "info",
  children,
}: {
  type?: Tone;
  children: React.ReactNode;
}) {
  return (
    <aside
      className={`my-6 rounded-lg border px-4 py-3 leading-relaxed ${TONE_STYLES[type]}`}
      role="note"
    >
      <span className="mr-2 font-bold text-accent">{TONE_ICONS[type]}</span>
      {children}
    </aside>
  );
}
