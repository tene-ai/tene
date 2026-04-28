import type { MetadataRoute } from "next";
import { comparisons } from "@/data/comparisons";
import { getAllCategories, getAllPosts, getAllTags } from "@/lib/blog";

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
export default function sitemap(): MetadataRoute.Sitemap {
  const base = "https://tene.sh";
  const lastModified = new Date().toISOString();

  const comparisonUrls: MetadataRoute.Sitemap = comparisons.map((c) => ({
    url: `${base}/vs/${c.slug}`,
    lastModified: c.updatedAt,
    changeFrequency: "monthly",
    priority: 0.8,
  }));

  const posts = getAllPosts();
  const blogPostUrls: MetadataRoute.Sitemap = posts.map((p) => ({
    url: `${base}/blog/${p.slug}`,
    lastModified: p.updatedAt ?? p.publishedAt,
    changeFrequency: "monthly",
    priority: 0.7,
  }));

  const blogTagUrls: MetadataRoute.Sitemap = getAllTags().map(({ tag }) => ({
    url: `${base}/blog/tag/${tag}`,
    lastModified,
    changeFrequency: "monthly",
    priority: 0.5,
  }));

  // 4 category hub pages — always present, even with 0 posts (empty
  // categories render a "Coming soon" placeholder for taxonomy discovery).
  const blogCategoryUrls: MetadataRoute.Sitemap = getAllCategories().map(
    ({ category }) => ({
      url: `${base}/blog/category/${category}`,
      lastModified,
      changeFrequency: "monthly",
      priority: 0.6,
    }),
  );

  return [
    {
      url: base,
      lastModified,
      changeFrequency: "weekly",
      priority: 1.0,
    },
    {
      url: `${base}/vs`,
      lastModified,
      changeFrequency: "monthly",
      priority: 0.9,
    },
    {
      url: `${base}/cli`,
      lastModified,
      changeFrequency: "weekly",
      priority: 0.8,
    },
    ...comparisonUrls,
    {
      url: `${base}/blog`,
      lastModified,
      changeFrequency: "weekly",
      priority: 0.9,
    },
    ...blogPostUrls,
    ...blogCategoryUrls,
    ...blogTagUrls,
  ];
}
