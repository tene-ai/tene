import { Suspense } from "react";
import type { Metadata } from "next";
import { Nav } from "@/components/nav";
import { Footer } from "@/components/footer";
import { InteractiveGrid } from "@/components/interactive-grid";
import { PostCard } from "@/components/blog/post-card";
import { RssLink } from "@/components/blog/rss-link";
import { CategoryPills } from "@/components/blog/category-pills";
import { TagFilter } from "@/components/blog/tag-filter";
import { BlogIndexClient } from "@/components/blog/blog-index-client";
import { BlogIndexJsonLd } from "@/components/seo/blog-index-jsonld";
import { getAllCategories, getAllPosts, getAllTags } from "@/lib/blog";

export const metadata: Metadata = {
  title: "Tech Blog — Tene",
  description:
    "Articles on AI-safe secret management, vibe coding, developer security, local-first infrastructure, and CLI design — from the team building tene.",
  alternates: {
    canonical: "https://tene.sh/blog",
    // FR-35 — RSS feed auto-discovery (<link rel="alternate"> in head).
    types: {
      "application/rss+xml": [
        { url: "https://tene.sh/blog/rss.xml", title: "tene Tech Blog RSS" },
      ],
    },
  },
  openGraph: {
    title: "tene Tech Blog",
    description:
      "AI-safe secrets · Vibe coding · Developer security · Local-first infrastructure",
    url: "https://tene.sh/blog",
    siteName: "Tene",
    type: "website",
    images: [
      {
        url: "/og-image.png",
        width: 1200,
        height: 630,
        alt: "tene Tech Blog",
      },
    ],
  },
  twitter: {
    card: "summary_large_image",
    title: "tene Tech Blog",
    description:
      "AI-safe secrets · Vibe coding · Developer security · Local-first infrastructure",
    images: ["/og-image.png"],
  },
  robots: { index: true, follow: true },
};

export default function BlogIndex() {
  const posts = getAllPosts();
  const tags = getAllTags();
  const categories = getAllCategories();

  // Plain JSON payload for client component — BlogPostMeta minus any
  // non-serialisable fields (there are none today).
  const postsForClient = posts.map((p) => ({
    slug: p.slug,
    title: p.title,
    description: p.description,
    publishedAt: p.publishedAt,
    updatedAt: p.updatedAt,
    category: p.category,
    tags: p.tags,
    author: p.author,
    readingMinutes: p.readingMinutes,
    wordCount: p.wordCount,
  }));

  return (
    <>
      {/* blog-seo-enhancements G2 — Blog + BreadcrumbList schema */}
      <BlogIndexJsonLd posts={posts} />

      <InteractiveGrid />
      <div className="dot-grid-fixed sm:hidden" />

      <Nav />
      <main className="relative z-10">
        <section className="px-4 pt-28 pb-6 sm:px-6">
          <div className="mx-auto max-w-4xl text-center">
            <h1 className="text-3xl font-bold leading-tight tracking-tight sm:text-4xl md:text-5xl">
              tene Tech Blog
            </h1>
            <p className="mx-auto mt-6 max-w-2xl text-base text-muted leading-relaxed sm:text-lg">
              Tools · Engineering · Vibe Coding · Philosophy
            </p>

            <div className="mt-6 flex justify-center">
              <RssLink location="blog_header" />
            </div>
          </div>
        </section>

        {categories.length > 0 && (
          <section className="px-4 pb-6 sm:px-6">
            <CategoryPills categories={categories} />
          </section>
        )}

        {tags.length > 0 && (
          <section className="px-4 pb-8 sm:px-6">
            {/* useSearchParams suspense boundary (Next.js 15 requirement). */}
            <Suspense
              fallback={
                <div className="mx-auto h-12 max-w-4xl rounded-lg border border-border/50 bg-surface/30" />
              }
            >
              <TagFilter allTags={tags} />
            </Suspense>
          </section>
        )}

        <section className="px-4 pb-16 sm:px-6">
          {posts.length === 0 ? (
            <div className="mx-auto max-w-xl text-center text-muted">
              No articles published yet. Check back soon.
            </div>
          ) : (
            <Suspense
              fallback={
                <ul className="mx-auto grid max-w-4xl gap-4 sm:grid-cols-2">
                  {postsForClient.map((post) => (
                    <li key={post.slug}>
                      <PostCard post={post} />
                    </li>
                  ))}
                </ul>
              }
            >
              <BlogIndexClient posts={postsForClient} />
            </Suspense>
          )}
        </section>
      </main>
      <Footer />
    </>
  );
}
