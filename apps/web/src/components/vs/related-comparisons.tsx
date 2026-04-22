"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { GlowCard } from "@/components/glow-card";
import { track } from "@/lib/track";
import type { Comparison } from "@/data/comparisons/types";

type Props = {
  items: Comparison[];
};

export function RelatedComparisons({ items }: Props) {
  const pathname = usePathname();
  const fromSlug = pathname?.replace(/^\/vs\//, "") ?? "";
  if (items.length === 0) return null;

  return (
    <section className="px-4 py-12 sm:px-6">
      <div className="mx-auto max-w-4xl">
        <h2 className="text-2xl font-bold sm:text-3xl">Other comparisons</h2>
        <p className="mt-3 text-muted">
          See how tene compares against every other major secret manager.
        </p>

        <ul className="mt-8 grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {items.map((c) => (
            <li key={c.slug}>
              <GlowCard className="h-full rounded-lg border border-border bg-surface/80 backdrop-blur-sm transition-colors hover:border-accent/40">
                <Link
                  href={`/vs/${c.slug}`}
                  onClick={() =>
                    track("related_vs_click", {
                      from: fromSlug,
                      to: c.slug,
                    })
                  }
                  className="block h-full p-5"
                >
                  <div className="font-medium">
                    tene vs {c.competitorName}
                  </div>
                  <p className="mt-2 text-sm text-muted line-clamp-3">
                    {c.subheadline}
                  </p>
                </Link>
              </GlowCard>
            </li>
          ))}
        </ul>
      </div>
    </section>
  );
}
