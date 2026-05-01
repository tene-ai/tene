import fs from "node:fs";
import path from "node:path";
import type { MetadataRoute } from "next";
import { comparisons } from "@/data/comparisons";
import {
  type BlogPostMeta,
  getAllCategories,
  getAllPosts,
  getIndexableTags,
  getPostsByCategory,
  getPostsByTag,
} from "@/lib/blog";

// Next.js generates /sitemap.xml from this file at build time (force-static
// route). Includes the home page, the /vs index, every /vs/{slug} page, and
// the /blog + /blog/{slug} + /blog/tag/{tag} + /blog/category/{cat}.
//
// /blog/rss.xml is intentionally NOT in this sitemap — it's an RSS feed
// (XML), not an HTML page. Google flagged it as "크롤링됨 - 색인 안 됨"
// (crawled but not indexed) when it was here. RSS discovery is handled
// separately via the `Sitemap:` directive in robots.txt and the
// <link rel="alternate" type="application/rss+xml"> in the root layout.
// Submit /sitemap.xml to Google Search Console via "Add sitemap".
//
// `lastModified` is computed from real content dates (not `new Date()`).
// Google de-weights sitemaps where every URL bumps lastmod every deploy
// (treats it as a spam signal); accurate timestamps prioritise the crawl
// queue. Per Google's "Build and submit a sitemap" documentation.
export default function sitemap(): MetadataRoute.Sitemap {
  const base = "https://tene.sh";

  // Convenience: newest publishedAt/updatedAt across an array of posts.
  // Falls back to today's date if the array is empty (an empty category
  // or tag bucket — shouldn't happen, but render a safe value).
  const newestPostDate = (
    posts: ReadonlyArray<BlogPostMeta>,
  ): string | Date => {
    if (posts.length === 0) return new Date();
    const epoch = posts.reduce((max, p) => {
      const d = new Date(p.updatedAt ?? p.publishedAt).getTime();
      return d > max ? d : max;
    }, 0);
    return new Date(epoch).toISOString();
  };

  const posts = getAllPosts();
  const newestBlogDate = newestPostDate(posts);
  const newestComparisonDate = comparisons.reduce<string>(
    (max, c) => (c.updatedAt > max ? c.updatedAt : max),
    "",
  );

  // /cli renders docs/cli-reference.md at build time. Use the file mtime
  // for sitemap lastmod so Google sees a real "the doc changed" signal.
  const cliLastmod: string | Date = (() => {
    const candidates = [
      path.join(process.cwd(), "..", "..", "docs", "cli-reference.md"),
      path.join(process.cwd(), "docs", "cli-reference.md"),
    ];
    for (const p of candidates) {
      try {
        return fs.statSync(p).mtime.toISOString();
      } catch {
        // try next
      }
    }
    return newestBlogDate;
  })();

  const comparisonUrls: MetadataRoute.Sitemap = comparisons.map((c) => ({
    url: `${base}/vs/${c.slug}`,
    lastModified: c.updatedAt,
    changeFrequency: "monthly",
    priority: 0.8,
  }));

  const blogPostUrls: MetadataRoute.Sitemap = posts.map((p) => ({
    url: `${base}/blog/${p.slug}`,
    lastModified: p.updatedAt ?? p.publishedAt,
    changeFrequency: "monthly",
    priority: 0.7,
  }));

  // Only emit tag pages that have ≥ INDEXABLE_TAG_THRESHOLD articles. Thin
  // tag pages (1–2 articles) are noindex'd at the page level AND excluded
  // here so they never enter Google's crawl queue in the first place. See
  // lib/blog.ts and .claude/rules/blog-content.md §10.1.
  const blogTagUrls: MetadataRoute.Sitemap = getIndexableTags().map(
    ({ tag }) => ({
      url: `${base}/blog/tag/${tag}`,
      lastModified: newestPostDate(getPostsByTag(tag)),
      changeFrequency: "monthly",
      priority: 0.5,
    }),
  );

  // 4 category hub pages — always present, even with 0 posts (empty
  // categories render a "Coming soon" placeholder for taxonomy discovery).
  const blogCategoryUrls: MetadataRoute.Sitemap = getAllCategories().map(
    ({ category }) => ({
      url: `${base}/blog/category/${category}`,
      lastModified: newestPostDate(getPostsByCategory(category)),
      changeFrequency: "monthly",
      priority: 0.6,
    }),
  );

  return [
    {
      // Home — bumps when newest blog post or comparison ships
      url: base,
      lastModified:
        new Date(newestBlogDate).getTime() >
        new Date(newestComparisonDate).getTime()
          ? newestBlogDate
          : newestComparisonDate,
      changeFrequency: "weekly",
      priority: 1.0,
    },
    {
      url: `${base}/vs`,
      lastModified: newestComparisonDate,
      changeFrequency: "monthly",
      priority: 0.9,
    },
    {
      url: `${base}/cli`,
      lastModified: cliLastmod,
      changeFrequency: "weekly",
      priority: 0.8,
    },
    ...comparisonUrls,
    {
      url: `${base}/blog`,
      lastModified: newestBlogDate,
      changeFrequency: "weekly",
      priority: 0.9,
    },
    ...blogPostUrls,
    ...blogCategoryUrls,
    ...blogTagUrls,
  ];
}
