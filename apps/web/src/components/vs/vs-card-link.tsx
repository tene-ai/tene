"use client";

// Client wrapper for /vs index cards that fires `vs_card_click` on click.
// Answers the product question "어떤 제품과 비교를 많이 하는지".
import Link from "next/link";
import { GlowCard } from "@/components/glow-card";
import { track } from "@/lib/track";

type Props = {
  slug: string;
  competitorName: string;
  subheadline: string;
};

export function VsCardLink({ slug, competitorName, subheadline }: Props) {
  return (
    <GlowCard className="h-full rounded-lg border border-border bg-surface/80 backdrop-blur-sm transition-colors hover:border-accent/40">
      <Link
        href={`/vs/${slug}`}
        onClick={() => track("vs_card_click", { competitor: slug })}
        className="block h-full p-6"
      >
        <div className="text-lg font-semibold">tene vs {competitorName}</div>
        <p className="mt-3 text-sm text-muted">{subheadline}</p>
        <span className="mt-4 inline-flex items-center gap-1 text-sm text-accent">
          Read comparison →
        </span>
      </Link>
    </GlowCard>
  );
}
