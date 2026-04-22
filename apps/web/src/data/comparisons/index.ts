import type { Comparison } from "./types";
import { dotenvComparison } from "./dotenv";
import { dopplerComparison } from "./doppler";
import { dotenvVaultComparison } from "./dotenv-vault";
import { infisicalComparison } from "./infisical";
import { vaultComparison } from "./vault";

// Master registry. Order defines the sitemap + related-comparison order.
const rawComparisons: Comparison[] = [
  dotenvComparison,
  dopplerComparison,
  dotenvVaultComparison,
  infisicalComparison,
  vaultComparison,
];

// Auto-populate `related` with the other four comparisons so each page links
// to every peer. Keeps internal linking symmetric + boosts PageRank flow.
const comparisonsBySlug = new Map<string, Comparison>();
for (const c of rawComparisons) {
  const related = rawComparisons
    .filter((other) => other.slug !== c.slug)
    .map((other) => other.slug);
  comparisonsBySlug.set(c.slug, { ...c, related });
}

export const comparisons: Comparison[] = Array.from(comparisonsBySlug.values());

export function getAllComparisonSlugs(): string[] {
  return comparisons.map((c) => c.slug);
}

export function getComparison(slug: string): Comparison | undefined {
  return comparisonsBySlug.get(slug);
}

export function getRelatedComparisons(slug: string, limit = 3): Comparison[] {
  const current = comparisonsBySlug.get(slug);
  if (!current) return [];
  return (current.related ?? [])
    .map((s) => comparisonsBySlug.get(s))
    .filter((c): c is Comparison => c !== undefined)
    .slice(0, limit);
}
