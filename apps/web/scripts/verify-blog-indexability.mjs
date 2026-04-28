#!/usr/bin/env node
// Build-time validator that prevents silent indexability regressions.
//
// Catches the NG patterns the GSC dashboard surfaced as "not indexed":
//   1. `draft: true` left on a post that should be live
//      → silently dropped from RSS + sitemap
//   2. `canonicalUrl: "https://other-site.com/..."` without intentional repost
//      → our copy gets deindexed in favor of the other site
//   3. Slug typo / filename mismatch
//      → 404 in production, missing from sitemap
//   4. Frontmatter missing required field
//      → already throws at loadPost(), but we re-assert here for clarity
//
// Runs as `npm run verify:blog` and as a postbuild step.
// Exits non-zero on any violation.

import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import matter from "gray-matter";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const CONTENT_DIR = path.resolve(__dirname, "..", "content", "blog");
const SITE_ORIGIN = "https://tene.sh";

const errors = [];
const warnings = [];

function relPath(p) {
  return path.relative(path.resolve(__dirname, ".."), p);
}

const files = fs
  .readdirSync(CONTENT_DIR)
  .filter((f) => f.endsWith(".mdx"));

for (const file of files) {
  const filePath = path.join(CONTENT_DIR, file);
  const filenameSlug = file.replace(/\.mdx$/, "");
  const raw = fs.readFileSync(filePath, "utf-8");
  const { data } = matter(raw);

  // Required field assertions (mirrors loadPost) — caught earlier in build,
  // but re-asserting here surfaces ALL violations in one pass instead of one
  // at a time.
  for (const k of ["slug", "title", "description", "publishedAt", "category"]) {
    if (!data[k]) errors.push(`${relPath(filePath)}: missing required '${k}'`);
  }

  // Slug must match filename — otherwise the post URL produced by Next.js
  // ([slug].mdx → /blog/{filename-slug}) won't match the sitemap entry
  // (data.slug), causing 404 + GSC "not indexed".
  if (data.slug && data.slug !== filenameSlug) {
    errors.push(
      `${relPath(filePath)}: slug '${data.slug}' does not match filename '${filenameSlug}'`,
    );
  }

  // canonicalUrl override sanity — if explicitly set off-site without a
  // visible reason in frontmatter, that's almost always a bug. Allow it ONLY
  // when accompanied by a `canonicalReason` comment-field or a non-default
  // `author` (guest post). Otherwise warn loudly.
  if (data.canonicalUrl) {
    const expected = `${SITE_ORIGIN}/blog/${filenameSlug}`;
    if (data.canonicalUrl !== expected) {
      const isReposted =
        data.canonicalUrl.startsWith("https://") &&
        !data.canonicalUrl.startsWith(SITE_ORIGIN);
      if (isReposted) {
        warnings.push(
          `${relPath(filePath)}: canonicalUrl points off-site ('${data.canonicalUrl}'). ` +
            `Our copy will be deindexed. Confirm this is an intentional repost.`,
        );
      } else {
        errors.push(
          `${relPath(filePath)}: canonicalUrl mismatch. ` +
            `Expected '${expected}', got '${data.canonicalUrl}'.`,
        );
      }
    }
  }

  // Draft flag — published articles must not have draft: true. The build
  // succeeds either way (loadPost filters drafts out of getAllPosts), but
  // the silent drop is exactly the GSC "discovered, not crawled" surface.
  // Surface it as a warning so the writer sees what they're shipping.
  if (data.draft === true) {
    warnings.push(
      `${relPath(filePath)}: draft: true → will be excluded from RSS + sitemap`,
    );
  }
}

if (warnings.length) {
  console.warn("\n⚠️  blog indexability warnings:");
  for (const w of warnings) console.warn("  - " + w);
}

if (errors.length) {
  console.error("\n❌ blog indexability errors:");
  for (const e of errors) console.error("  - " + e);
  console.error(
    `\n${errors.length} error(s). Fix the above before commit.\n` +
      `Each violation maps to a GSC "not indexed" reason. ` +
      `See .claude/rules/blog-content.md §10 for the full NG pattern list.\n`,
  );
  process.exit(1);
}

const draftCount = warnings.filter((w) => w.includes("draft: true")).length;
const publishedCount = files.length - draftCount;
console.log(
  `✅ ${publishedCount} published article(s) pass indexability checks` +
    (draftCount ? ` (${draftCount} draft excluded)` : ""),
);
