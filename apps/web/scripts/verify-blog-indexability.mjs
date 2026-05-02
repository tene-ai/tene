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
//   5. Schema.org / OG datetime regression (added 2026-05-02 after Rich
//      Results Test rejected date-only `datePublished` / `dateModified`)
//      → catches the case where a future emission site forgets the
//        toIsoDateTime() helper and ships date-only strings
//   6. FAQPage schema-content mismatch (added 2026-05-03 after GSC flagged
//      "FAQPage 입력란이 중복" caused by JSON-LD Q&A drifting from the
//      visible <FAQ /> UI on `/`. Same risk class for /blog/{slug} where
//      frontmatter `faqs:` could ship without a body ## FAQ section.)
//      → ensures every Question/Answer in any FAQPage appears as visible
//        text on the same page. See .claude/rules/blog-content.md §10.4.
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

// ──────────────────────────────────────────────────────────────────────
// Datetime regression guard — runs only if Next.js has produced its build
// output (.next/server/app). Skipped during `npm run verify:blog` outside
// of postbuild (where there's nothing to grep yet).
// ──────────────────────────────────────────────────────────────────────
const buildDir = path.resolve(__dirname, "..", ".next", "server", "app");

// ISO 8601 datetime with timezone (Z or ±HH:MM). What Schema.org / OG want.
// Examples that PASS: "2026-05-02T00:00:00.000Z", "2026-05-02T00:00:00+09:00"
// Examples that FAIL: "2026-05-02", "2026-05-02T00:00:00" (no TZ)
const ISO_DT = /T\d{2}:\d{2}:\d{2}(?:\.\d+)?(?:Z|[+-]\d{2}:?\d{2})/;

function checkDatetimeFile(htmlPath) {
  const html = fs.readFileSync(htmlPath, "utf-8");
  const rel = path.relative(path.resolve(__dirname, ".."), htmlPath);
  // Schema.org JSON-LD fields
  for (const field of ["datePublished", "dateModified"]) {
    const matches = html.match(
      new RegExp(`"${field}":"([^"]+)"`, "g"),
    );
    if (!matches) continue;
    for (const m of matches) {
      const v = m.match(/"[^"]+":"([^"]+)"/)[1];
      if (!ISO_DT.test(v)) {
        errors.push(
          `${rel}: Schema.org ${field}='${v}' is not full ISO 8601 datetime ` +
            `with timezone. Use src/lib/iso-date.ts → toIsoDateTime().`,
        );
      }
    }
  }
  // OpenGraph article:*_time meta tags
  for (const prop of ["article:published_time", "article:modified_time"]) {
    const matches = html.match(
      new RegExp(`property="${prop}"[^>]*content="([^"]+)"`, "g"),
    );
    if (!matches) continue;
    for (const m of matches) {
      const v = m.match(/content="([^"]+)"/)[1];
      if (!ISO_DT.test(v)) {
        errors.push(
          `${rel}: OG ${prop}='${v}' is not full ISO 8601 datetime ` +
            `with timezone. Wrap with toIsoDateTime() in the page metadata.`,
        );
      }
    }
  }
}

function walkHtml(dir) {
  if (!fs.existsSync(dir)) return [];
  const out = [];
  const entries = fs.readdirSync(dir, { withFileTypes: true });
  for (const e of entries) {
    const p = path.join(dir, e.name);
    if (e.isDirectory()) out.push(...walkHtml(p));
    else if (e.isFile() && e.name.endsWith(".html")) out.push(p);
  }
  return out;
}

// ──────────────────────────────────────────────────────────────────────
// FAQPage schema-content mismatch guard. Catches the GSC "FAQPage 입력란이
// 중복" warning at root cause: every Question/Answer that appears in
// JSON-LD must also appear as visible text on the page. Otherwise either:
//   - Google flags "FAQ markup not matching visible content", surfacing as
//     duplicate-entry warning (multiple FAQPage interpretations)
//   - Manual penalty for "hidden FAQ markup" — site-wide rich-result block
// ──────────────────────────────────────────────────────────────────────

// Strip <script>, <style>, all HTML tags, and decode the few HTML entities
// our content actually uses, so substring matching against the JSON-LD
// answer text works on the visible body.
function htmlToText(html) {
  let t = html;
  t = t.replace(/<script[\s\S]*?<\/script>/gi, " ");
  t = t.replace(/<style[\s\S]*?<\/style>/gi, " ");
  t = t.replace(/<[^>]+>/g, " ");
  t = t
    .replace(/&amp;/g, "&")
    .replace(/&lt;/g, "<")
    .replace(/&gt;/g, ">")
    .replace(/&quot;/g, '"')
    .replace(/&#x27;/g, "'")
    .replace(/&#39;/g, "'")
    .replace(/&apos;/g, "'")
    .replace(/&nbsp;/g, " ")
    .replace(/&mdash;/g, "—")
    .replace(/&ndash;/g, "–")
    .replace(/&hellip;/g, "…")
    .replace(/&ldquo;/g, "“")
    .replace(/&rdquo;/g, "”")
    .replace(/&lsquo;/g, "‘")
    .replace(/&rsquo;/g, "’");
  // Collapse whitespace so JSON-LD literal "two   spaces" matches HTML
  // text run with arbitrary whitespace.
  return t.replace(/\s+/g, " ");
}

// Decode the `\u00xx` escapes Next.js emits inside JSON-LD strings so we
// can substring-match them against the visible text (which uses the real
// glyphs). Lightweight — only handles the cases we actually emit.
function decodeJsonString(s) {
  return s.replace(/\\u([0-9a-fA-F]{4})/g, (_, h) =>
    String.fromCharCode(parseInt(h, 16)),
  );
}

function extractFaqs(html) {
  const out = [];
  const blocks = html.match(
    /<script type="application\/ld\+json"[^>]*>([\s\S]*?)<\/script>/g,
  );
  if (!blocks) return out;
  for (const block of blocks) {
    const m = block.match(
      /<script type="application\/ld\+json"[^>]*>([\s\S]*?)<\/script>/,
    );
    if (!m) continue;
    let data;
    try {
      data = JSON.parse(m[1]);
    } catch {
      continue;
    }
    const graph = data["@graph"] ? data["@graph"] : [data];
    for (const node of graph) {
      if (node["@type"] !== "FAQPage") continue;
      const entries = Array.isArray(node.mainEntity) ? node.mainEntity : [];
      for (const q of entries) {
        out.push({
          question: decodeJsonString(q.name ?? ""),
          answer: decodeJsonString(q.acceptedAnswer?.text ?? ""),
        });
      }
    }
  }
  return out;
}

// Normalise text to a tolerant "search canonical" form. Inline markdown
// formatting (backticks → <code>, bold/italic → <strong>/<em>) is stripped
// out by htmlToText, so the rendered visible text loses the surrounding
// markers. Frontmatter answers keep them. We normalise both sides by:
//   1. Removing inline-formatting punctuation (` * _) so byte-shifts in
//      formatting don't produce a false negative.
//   2. Decoding the few entities that may survive ("&quot;", "&#x27;").
//   3. Collapsing whitespace.
// The Q&A semantic content is what Google evaluates for FAQPage matching,
// not the markdown markup used to render it.
function canon(s) {
  return (
    s
      .replace(/[`*_]/g, "")
      .replace(/&quot;/g, '"')
      .replace(/&#x27;|&#39;/g, "'")
      .replace(/\s+/g, " ")
      // htmlToText replaces `<em>x</em>,` with ` x ,` — collapse the space
      // it injects before punctuation so the substring match doesn't fail
      // on inline-formatted phrases.
      .replace(/\s+([,.!?;:])/g, "$1")
      .trim()
  );
}

function checkFaqMirror(htmlPath) {
  const html = fs.readFileSync(htmlPath, "utf-8");
  const rel = path.relative(path.resolve(__dirname, ".."), htmlPath);
  const faqs = extractFaqs(html);
  if (faqs.length === 0) return; // no FAQPage on this page → nothing to verify
  const visibleCanon = canon(htmlToText(html));

  for (const f of faqs) {
    const qCanon = canon(f.question ?? "");
    const aCanon = canon(f.answer ?? "");
    if (qCanon && !visibleCanon.includes(qCanon)) {
      errors.push(
        `${rel}: FAQPage question NOT in visible body: "${f.question.slice(0, 80)}…"\n` +
          `    Add it to the body (## FAQ section) with byte-identical text. ` +
          `See blog-content.md §10.4.`,
      );
    }
    if (aCanon && !visibleCanon.includes(aCanon)) {
      errors.push(
        `${rel}: FAQPage answer NOT in visible body for question "${f.question.slice(0, 60)}…"\n` +
          `    Add the answer to the body (## FAQ section) with byte-identical text. ` +
          `See blog-content.md §10.4.`,
      );
    }
  }
}

if (fs.existsSync(buildDir)) {
  const htmlFiles = walkHtml(buildDir);
  // Datetime check — only surfaces that emit datePublished/dateModified.
  const datetimeRelevant = htmlFiles.filter((p) => {
    const rel = path.relative(buildDir, p);
    return (
      rel.startsWith("blog") ||
      rel === "cli.html" ||
      rel.startsWith("vs/") ||
      rel.startsWith("vs.html")
    );
  });
  for (const p of datetimeRelevant) checkDatetimeFile(p);

  // FAQ-mirror check — every page that emits FAQPage. Includes index.html
  // (`/`) which is the home FAQ surface alongside blog/cli/vs/.
  for (const p of htmlFiles) checkFaqMirror(p);
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
