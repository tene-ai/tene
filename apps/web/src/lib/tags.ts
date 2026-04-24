// Design Ref: §2.1 T2-2 — Fixed taxonomy (Category + Tag) to prevent
// fragmentation. Both vocabularies are closed sets; new entries require a PR.
// See: docs/02-design/features/blog-categories-and-tooling.design.md

// ---------------------------------------------------------------------------
// Categories — 1 required per post. Primary navigation axis on /blog.
// ---------------------------------------------------------------------------
export const CATEGORY_VOCABULARY = {
  tools: "Tools",
  engineering: "Engineering",
  "vibe-coding": "Vibe Coding",
  philosophy: "Philosophy",
} as const;

export type CategoryKey = keyof typeof CATEGORY_VOCABULARY;

export function isValidCategory(cat: string): cat is CategoryKey {
  return cat in CATEGORY_VOCABULARY;
}

export function getCategoryLabel(cat: string): string {
  return (CATEGORY_VOCABULARY as Record<string, string>)[cat] ?? cat;
}

export const CATEGORY_DESCRIPTIONS: Record<CategoryKey, string> = {
  tools:
    "Posts about tene, bkit, and the wider ADK family of developer tools.",
  engineering:
    "Engineering concepts, system design, cryptography, CLI architecture.",
  "vibe-coding":
    "AI-assisted development in practice — editor setups, agent workflows.",
  philosophy:
    "Essays on the AI era, solo-founder journey, and the craft of software.",
};

// ---------------------------------------------------------------------------
// Tags — 2–4 per post. Cross-cutting sub-topics. Organised into four axes
// mentally, but flat at the schema level.
// ---------------------------------------------------------------------------
export const TAG_VOCABULARY = {
  // Products
  tene: "tene",
  bkit: "bkit",
  // Technical domain
  security: "Security",
  cryptography: "Cryptography",
  architecture: "Architecture",
  devsecops: "DevSecOps",
  cli: "CLI",
  // AI era concepts
  "harness-engineering": "Harness Engineering",
  workflow: "Workflow",
  "claude-code": "Claude Code",
  cursor: "Cursor",
  // Format / character
  tutorial: "Tutorial",
  comparison: "Comparison",
  "deep-dive": "Deep Dive",
  "founder-story": "Founder Story",
} as const;

export type TagKey = keyof typeof TAG_VOCABULARY;

export function isValidTag(tag: string): tag is TagKey {
  return tag in TAG_VOCABULARY;
}

export function getTagLabel(tag: string): string {
  return (TAG_VOCABULARY as Record<string, string>)[tag] ?? tag;
}

// ---------------------------------------------------------------------------
// Retired vocabulary — kept here so route generation can emit 301 redirects
// for any external links that still point at the old slugs.
// ---------------------------------------------------------------------------
export const RETIRED_TAG_REDIRECTS: Record<string, string> = {
  ai: "vibe-coding", // redirect target is a *category*, handled by middleware
  go: "architecture",
  "vibe-coding": "__category__", // the tag was promoted to a category
};
