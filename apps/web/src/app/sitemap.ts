import type { MetadataRoute } from "next";
import { comparisons } from "@/data/comparisons";
import { getAllPosts, getAllTags } from "@/lib/blog";

// Next.js generates /sitemap.xml from this file at build time (force-static
// route). Includes the home page, the /vs index, every /vs/{slug} page, and
// the /blog + /blog/{slug} + /blog/tag/{tag} + /blog/rss.xml.
// Submit to Google Search Console via "Add sitemap".
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
    ...comparisonUrls,
    {
      url: `${base}/blog`,
      lastModified,
      changeFrequency: "weekly",
      priority: 0.9,
    },
    {
      url: `${base}/blog/rss.xml`,
      lastModified,
      changeFrequency: "weekly",
      priority: 0.6,
    },
    ...blogPostUrls,
    ...blogTagUrls,
  ];
}
