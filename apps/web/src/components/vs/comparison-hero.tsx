import { CopyCommand } from "@/components/copy-command";
import { TrackedGithubLink } from "@/components/tracked-github-link";
import type { Comparison } from "@/data/comparisons/types";

type Props = {
  comparison: Comparison;
};

export function ComparisonHero({ comparison }: Props) {
  return (
    <section className="relative px-4 pt-28 pb-16 sm:px-6">
      <div className="absolute inset-0 bg-[radial-gradient(ellipse_at_top,_var(--tw-gradient-stops))] from-accent/5 via-transparent to-transparent" />

      <div className="relative z-10 mx-auto max-w-4xl text-center">
        <div className="mb-6 inline-flex items-center gap-2 rounded-full border border-border bg-surface px-4 py-1.5 text-sm text-muted">
          <span className="h-2 w-2 rounded-full bg-accent animate-pulse" />
          Comparison
        </div>

        <h1 className="text-3xl font-bold leading-tight tracking-tight sm:text-4xl md:text-5xl">
          {comparison.headline}
        </h1>

        <p className="mx-auto mt-6 max-w-2xl text-base text-muted leading-relaxed sm:text-lg">
          {comparison.subheadline}
        </p>

        <div className="mt-10 flex flex-col items-center gap-4">
          <CopyCommand
            command="curl -sSfL https://tene.sh/install.sh | sh"
            className="w-full justify-center sm:w-auto"
            source="vs_page"
          />
          <TrackedGithubLink
            href="https://github.com/tomo-kay/tene"
            location="vs_page"
            className="inline-flex items-center gap-2 text-sm text-muted underline underline-offset-4 transition-colors hover:text-foreground"
          >
            Star on GitHub
          </TrackedGithubLink>
        </div>
      </div>
    </section>
  );
}
