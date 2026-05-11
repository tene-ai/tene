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
// Per-tag descriptions. Two purposes:
//   1. Unique <meta name="description"> + OG description on /blog/tag/{tag},
//      so Google sees distinct copy per tag page (otherwise every tag page
//      shipped the generic "Articles tagged X" string, which triggered the
//      GSC "Crawled — currently not indexed" signal for thin / duplicate
//      content. 2026-05-11 fix.)
//   2. Visible intro paragraph on the tag page itself — gives the index
//      surface something other than a card grid, which is the actual
//      thin-content fix (meta-only changes don't materially shift Google's
//      quality classifier — the visible body has to carry weight too).
// Each entry is ≤155 chars so it survives SERP truncation.
// ---------------------------------------------------------------------------
export const TAG_DESCRIPTIONS: Record<TagKey, string> = {
  tene: "tene is an open-source CLI secret manager for AI-augmented development. Articles on local-first vault design and AI-tool integration.",
  bkit: "bkit is a Claude Code plugin that verifies AI-generated code against its own design spec. Articles on context engineering and PDCA verification.",
  security: "AI-coding-era security: secrets sprawl, vibe-coded RLS misconfigurations, prompt injection, and OWASP gaps the agent age makes worse.",
  cryptography: "XChaCha20-Poly1305, AEAD, nonces — the crypto primitives behind tene's local-first vault and why they outperform classic AES-GCM.",
  architecture: "Software architecture in the AI agent era — clean architecture, ports and adapters, and design choices that survive AI rewrites.",
  devsecops: "Shift-left security for AI-augmented teams: secret scanning, CI gates, and the new attack surface that lives inside the prompt loop.",
  cli: "Designing terminal-first tools for developers and AI agents. UX, machine-readability, and the unix philosophy applied to LLM toolchains.",
  "harness-engineering":
    "The workflow around the model — skills, agents, hooks, state, and verification loops that raise the floor of org productivity.",
  workflow:
    "Process discipline in AI-augmented development: PDCA cycles, spec-first design, gap analysis, and patterns that keep agent outputs reproducible.",
  "claude-code":
    "Articles on Claude Code — Anthropic's agentic coding CLI. Skills, subagents, MCP integration, plan mode, and production-grade workflows.",
  cursor:
    "Cursor IDE deep dives — composer, .cursor/rules, MCP integration, and how Cursor compares to other agentic coding environments in 2026.",
  tutorial:
    "Step-by-step guides for AI-safe development. Concrete commands, real examples, and the smallest path from problem to working code.",
  comparison:
    "Head-to-head tool comparisons. Tradeoffs, when each option wins, and the questions to ask before picking a secret manager or coding agent.",
  "deep-dive":
    "Long-form technical articles that go below the surface — internals, design decisions, threat models, and the engineering behind headlines.",
  "founder-story":
    "Notes from a solo builder shipping in the AI era — successes, failures, and the day-to-day operator reality of running an open-core company.",
};

export function getTagDescription(tag: string): string {
  return (
    (TAG_DESCRIPTIONS as Record<string, string>)[tag] ??
    `Articles tagged ${getTagLabel(tag)} from the tene Tech Blog.`
  );
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
