#!/usr/bin/env node
// One-time backfill: ensure every blog post with frontmatter `faqs:` has a
// matching `## FAQ` section in the body that mirrors Q+A byte-for-byte.
// Idempotent — replaces any existing `## FAQ` block in place, so repeated
// runs converge on the same output.
//
// Why this exists: the postbuild verifier (verify-blog-indexability.mjs
// §FAQ-mirror) requires every Question/Answer in FAQPage JSON-LD to
// appear as visible text in the same HTML. Frontmatter is the source of
// truth; this script keeps body content in sync after schema changes.
//
// MDX gotcha: bare `<command>` / `<key>` etc. in prose are parsed as JSX
// elements and crash the build. Frontmatter answers stay raw (matching
// JSON-LD). When we render them into the body, we escape `<` and `>` to
// HTML entities — MDX treats those as literal characters, and the
// post-render visible text decodes back to `<command>`, so the
// FAQ-mirror substring match still works.
//
// Run: `node scripts/backfill-faq.mjs` (from apps/web).

import fs from "node:fs";
import path from "node:path";
import matter from "gray-matter";

const CONTENT_DIR = path.resolve("content/blog");
const files = fs.readdirSync(CONTENT_DIR).filter((f) => f.endsWith(".mdx"));

// Escape characters MDX interprets as JSX / expression syntax in plain
// prose. Use backslash escapes (the MDX-spec way) — entity escapes don't
// work because MDX 3 decodes entities before the JSX check, so
// `&lt;command&gt;` still triggers the JSX parser. Backslash escapes are
// stripped before rendering, so the visible HTML text matches the
// frontmatter answer byte-for-byte.
function escapeForMdxProse(s) {
  return s
    .replace(/</g, "\\<")
    .replace(/>/g, "\\>")
    .replace(/\{/g, "\\{")
    .replace(/\}/g, "\\}");
}

// Strip an existing `## FAQ` block (everything until the next h2 or the
// `Related reading:` line). Lets repeated runs converge on the same
// output regardless of prior state.
function stripFaqSection(content) {
  // /g flag — strip ALL occurrences (earlier broken runs may have stacked
  // multiple ## FAQ sections; we converge to a single block).
  return content.replace(
    /\n##\s+FAQ\s*\n[\s\S]*?(?=\n##\s|\nRelated reading:|$)/g,
    "\n",
  );
}

let modified = 0,
  skipped = 0;
for (const file of files) {
  const filePath = path.join(CONTENT_DIR, file);
  const raw = fs.readFileSync(filePath, "utf-8");
  const { data, content } = matter(raw);
  if (!Array.isArray(data.faqs) || data.faqs.length === 0) {
    skipped++;
    continue;
  }
  const stripped = stripFaqSection(content);
  const block = [
    "",
    "## FAQ",
    "",
    ...data.faqs.flatMap((f) => [
      `**${escapeForMdxProse(f.question)}**`,
      "",
      escapeForMdxProse(f.answer),
      "",
    ]),
  ].join("\n");
  let updated;
  const relatedMatch = stripped.match(/\n(Related reading[\s\S]*)$/);
  if (relatedMatch) {
    const before = stripped.slice(
      0,
      stripped.length - relatedMatch[1].length - 1,
    );
    updated = `${before.trimEnd()}\n${block}\n\n${relatedMatch[1]}`;
  } else {
    updated = `${stripped.trimEnd()}\n${block}\n`;
  }
  const newRaw = matter.stringify(updated, data);
  fs.writeFileSync(filePath, newRaw, "utf-8");
  modified++;
  console.log(`  ✓ ${file} (synced ${data.faqs.length} Q&A)`);
}
console.log(
  `\n${modified} synced, ${skipped} skipped (no faqs[] in frontmatter)`,
);
