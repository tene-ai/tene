// Design Ref: §2.2 T2-2 — Fixed tag vocabulary to prevent fragmentation.
// New tags require editing this file (PR review checkpoint). Plan D6.
export const TAG_VOCABULARY = {
  security: "Security",
  ai: "AI",
  cli: "CLI",
  go: "Go",
  devsecops: "DevSecOps",
  cryptography: "Cryptography",
  tutorial: "Tutorial",
  comparison: "Comparison",
  architecture: "Architecture",
  "vibe-coding": "Vibe Coding",
} as const;

export type TagKey = keyof typeof TAG_VOCABULARY;

export function isValidTag(tag: string): tag is TagKey {
  return tag in TAG_VOCABULARY;
}

export function getTagLabel(tag: string): string {
  return (TAG_VOCABULARY as Record<string, string>)[tag] ?? tag;
}
