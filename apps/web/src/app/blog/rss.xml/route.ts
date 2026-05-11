// Design Ref: §2.3 T3-4 — RSS 2.0 feed. Force-static so it's a plain XML
// file served from CDN edge. Regenerated on every deploy via getAllPosts.
import { getAllPosts } from "@/lib/blog";

export const dynamic = "force-static";

function escape(s: string): string {
  return s.replace(/[<>&'"]/g, (c) =>
    ({ "<": "&lt;", ">": "&gt;", "&": "&amp;", "'": "&apos;", '"': "&quot;" })[c]!,
  );
}

export async function GET(): Promise<Response> {
  const posts = getAllPosts();
  const lastBuildDate = new Date().toUTCString();

  const items = posts
    .map(
      (p) => `
    <item>
      <title>${escape(p.title)}</title>
      <link>https://tene.sh/blog/${p.slug}</link>
      <guid isPermaLink="true">https://tene.sh/blog/${p.slug}</guid>
      <description>${escape(p.description)}</description>
      <pubDate>${new Date(p.publishedAt).toUTCString()}</pubDate>
      ${p.tags.map((t) => `<category>${escape(t)}</category>`).join("")}
    </item>`,
    )
    .join("");

  const body = `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0" xmlns:atom="http://www.w3.org/2005/Atom">
  <channel>
    <title>tene Tech Blog</title>
    <link>https://tene.sh/blog</link>
    <description>AI-safe secrets · Vibe coding · Developer security · Local-first infrastructure</description>
    <language>en-us</language>
    <lastBuildDate>${lastBuildDate}</lastBuildDate>
    <atom:link href="https://tene.sh/blog/rss.xml" rel="self" type="application/rss+xml"/>${items}
  </channel>
</rss>`;

  return new Response(body, {
    headers: {
      "Content-Type": "application/rss+xml; charset=utf-8",
      "Cache-Control": "public, max-age=3600, stale-while-revalidate=86400",
      // X-Robots-Tag: noindex — RSS is an XML feed, not an HTML document.
      // Without this header, GSC indexes the URL as "Crawled — currently
      // not indexed" because the bytes don't parse as a page. The header
      // re-classifies it as "Excluded by 'noindex' tag", which is the
      // correct bucket. See .claude/rules/blog-content.md §10.1 and
      // verify-blog-indexability.mjs route-noindex assertion.
      "X-Robots-Tag": "noindex",
    },
  });
}
