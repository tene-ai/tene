#!/usr/bin/env node
// Postbuild assertion: thin tag pages (<3 articles) must be (a) excluded
// from sitemap.xml AND (b) carry a `noindex` meta in their built HTML.
// Indexable tag pages (≥3 articles) must be in sitemap AND must NOT have
// noindex.
//
// Catches three regression classes in one pass:
//   1. lib/blog.ts INDEXABLE_TAG_THRESHOLD changed accidentally.
//   2. sitemap.ts switched back to getAllTags() instead of getIndexableTags().
//   3. tag/[tag]/page.tsx generateMetadata() forgot to wire isIndexableTag().
//
// Runs as `npm run verify:thin` and as a postbuild step (chained after
// verify-blog-indexability.mjs). Exits non-zero on any violation.

import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import matter from "gray-matter";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const ROOT = path.resolve(__dirname, "..");
const CONTENT_DIR = path.join(ROOT, "content", "blog");
const NEXT_OUT = path.join(ROOT, ".next", "server", "app");
const THRESHOLD = 3; // mirrors INDEXABLE_TAG_THRESHOLD in lib/blog.ts

const errors = [];
const warnings = [];

// ---- 1. Compute tag → article count from MDX frontmatter ----
const tagCounts = new Map();
for (const file of fs.readdirSync(CONTENT_DIR).filter((f) => f.endsWith(".mdx"))) {
  const raw = fs.readFileSync(path.join(CONTENT_DIR, file), "utf-8");
  const { data } = matter(raw);
  if (data.draft === true) continue;
  for (const t of Array.isArray(data.tags) ? data.tags : []) {
    tagCounts.set(t, (tagCounts.get(t) ?? 0) + 1);
  }
}

const thinTags = [...tagCounts.entries()]
  .filter(([, c]) => c < THRESHOLD)
  .map(([t]) => t);
const indexableTags = [...tagCounts.entries()]
  .filter(([, c]) => c >= THRESHOLD)
  .map(([t]) => t);

// ---- 2. Locate sitemap.xml build artifact ----
// Next.js 16 emits sitemap.xml via app/sitemap.ts as a Route. The build
// artifact path is `.next/server/app/sitemap.xml.body` for force-static
// routes — fall back to a few other known locations across versions.
const sitemapCandidates = [
  path.join(NEXT_OUT, "sitemap.xml.body"),
  path.join(NEXT_OUT, "sitemap.xml", "route.body.txt"),
  path.join(NEXT_OUT, "sitemap.xml", "route.js"),
];
let sitemap = "";
for (const p of sitemapCandidates) {
  if (fs.existsSync(p)) {
    sitemap = fs.readFileSync(p, "utf-8");
    break;
  }
}
if (!sitemap) {
  warnings.push(
    "sitemap.xml build artifact not found. Tried: " +
      sitemapCandidates.map((p) => path.relative(ROOT, p)).join(", "),
  );
}

// ---- 3. Sitemap assertions ----
if (sitemap) {
  for (const tag of thinTags) {
    const url = `https://tene.sh/blog/tag/${tag}`;
    if (sitemap.includes(url)) {
      errors.push(
        `thin tag '${tag}' (${tagCounts.get(tag)} article${
          tagCounts.get(tag) === 1 ? "" : "s"
        } < ${THRESHOLD}) appears in sitemap.xml. ` +
          `Check apps/web/src/app/sitemap.ts uses getIndexableTags().`,
      );
    }
  }
  for (const tag of indexableTags) {
    const url = `https://tene.sh/blog/tag/${tag}`;
    if (!sitemap.includes(url)) {
      errors.push(
        `indexable tag '${tag}' (${tagCounts.get(tag)} articles >= ${THRESHOLD}) ` +
          `is missing from sitemap.xml.`,
      );
    }
  }
}

// ---- 4. HTML noindex assertions for thin tag pages ----
for (const tag of thinTags) {
  // Static-rendered tag pages emit `.next/server/app/blog/tag/{tag}.html`.
  const htmlPath = path.join(NEXT_OUT, "blog", "tag", `${tag}.html`);
  if (!fs.existsSync(htmlPath)) {
    warnings.push(
      `tag HTML not found: ${path.relative(ROOT, htmlPath)} (skipping noindex check for '${tag}')`,
    );
    continue;
  }
  const html = fs.readFileSync(htmlPath, "utf-8");
  // Next.js metadata API normalises to `<meta name="robots" content="noindex, follow">`
  // (or "noindex,follow" without space) — accept either.
  const hasNoindex =
    /<meta[^>]+name=["']robots["'][^>]+content=["'][^"']*noindex/i.test(html);
  if (!hasNoindex) {
    errors.push(
      `thin tag '${tag}' HTML missing noindex meta. ` +
        `Check apps/web/src/app/blog/tag/[tag]/page.tsx generateMetadata() ` +
        `wires isIndexableTag(tag) into robots.index.`,
    );
  }
}

// ---- 5. Indexable tag pages MUST NOT have noindex (regression guard) ----
for (const tag of indexableTags) {
  const htmlPath = path.join(NEXT_OUT, "blog", "tag", `${tag}.html`);
  if (!fs.existsSync(htmlPath)) continue;
  const html = fs.readFileSync(htmlPath, "utf-8");
  const hasNoindex =
    /<meta[^>]+name=["']robots["'][^>]+content=["'][^"']*noindex/i.test(html);
  if (hasNoindex) {
    errors.push(
      `indexable tag '${tag}' (${tagCounts.get(tag)} articles) HAS noindex ` +
        `meta — should be index. Check generateMetadata() boolean inversion.`,
    );
  }
}

// ---- 6. Report ----
console.log(
  `\nThin-tag scan: ${thinTags.length} thin (excluded), ` +
    `${indexableTags.length} indexable.`,
);
if (thinTags.length) {
  console.log(
    `  Thin: ${thinTags
      .map((t) => `${t}(${tagCounts.get(t)})`)
      .join(", ")}`,
  );
}

if (warnings.length) {
  console.warn("\n⚠️  thin-page warnings:");
  for (const w of warnings) console.warn("  - " + w);
}

if (errors.length) {
  console.error("\n❌ thin-page errors:");
  for (const e of errors) console.error("  - " + e);
  console.error(
    `\n${errors.length} error(s). See .claude/rules/blog-content.md §10.1.\n`,
  );
  process.exit(1);
}

console.log(`\n✅ thin-page assertions passed.\n`);
