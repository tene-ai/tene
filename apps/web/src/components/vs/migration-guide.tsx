import type { Comparison } from "@/data/comparisons/types";

type Props = {
  migration: NonNullable<Comparison["migration"]>;
};

export function MigrationGuide({ migration }: Props) {
  return (
    <section className="px-4 py-12 sm:px-6">
      <div className="mx-auto max-w-4xl">
        <h2 className="text-2xl font-bold sm:text-3xl">{migration.title}</h2>
        <p className="mt-3 text-muted">{migration.summary}</p>

        <ol className="mt-8 space-y-5">
          {migration.steps.map((step, idx) => (
            <li
              key={`${step.title}-${idx}`}
              className="flex gap-4 rounded-lg border border-border bg-surface p-5"
            >
              <span className="flex h-7 w-7 shrink-0 items-center justify-center rounded-full border border-accent/30 bg-accent/10 text-sm font-bold text-accent">
                {idx + 1}
              </span>
              <div className="min-w-0 flex-1">
                <div className="font-medium">{step.title}</div>
                {step.command && (
                  <pre className="mt-3 overflow-x-auto rounded bg-surface-2 p-3 text-xs">
                    <code className="font-mono text-accent">$ {step.command}</code>
                  </pre>
                )}
                {step.note && (
                  <p className="mt-2 text-sm text-muted">{step.note}</p>
                )}
              </div>
            </li>
          ))}
        </ol>

        {migration.postMigrationNote && (
          <p className="mt-6 rounded-lg border border-border bg-surface px-5 py-4 text-sm text-muted">
            <strong className="text-foreground">After migration:</strong>{" "}
            {migration.postMigrationNote}
          </p>
        )}
      </div>
    </section>
  );
}
