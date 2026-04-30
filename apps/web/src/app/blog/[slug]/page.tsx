import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { MDXRemote } from "next-mdx-remote/rsc";
import remarkGfm from "remark-gfm";
import rehypeSlug from "rehype-slug";
import rehypeAutolinkHeadings from "rehype-autolink-headings";
import rehypeShiki from "@shikijs/rehype";
import { Nav } from "@/components/nav";
import { Footer } from "@/components/footer";
import { InteractiveGrid } from "@/components/interactive-grid";
import { PostHero } from "@/components/blog/post-hero";
import { PostTOC } from "@/components/blog/post-toc";
import { PostFooter } from "@/components/blog/post-footer";
import { RelatedPosts } from "@/components/blog/related-posts";
import { ReadProgressTracker } from "@/components/blog/read-progress-tracker";
import { BlogPostingJsonLd } from "@/components/seo/article-jsonld";
import { FaqJsonLd } from "@/components/seo/faq-jsonld";
import { Breadcrumb } from "@/components/breadcrumb";
import { useMDXComponents } from "../../../../mdx-components";
import {
  getAllPostSlugs,
  getPostBySlug,
  getRelatedPosts,
} from "@/lib/blog";

export const dynamic = "error";

type Params = { slug: string };

export function generateStaticParams(): Params[] {
  return getAllPostSlugs().map((slug) => ({ slug }));
}

export async function generateMetadata({
  params,
}: {
  params: Promise<Params>;
}): Promise<Metadata> {
  const { slug } = await params;
  const post = getPostBySlug(slug);
  if (!post) return {};
  const canonical = post.meta.canonicalUrl!;

  return {
    title: `${post.meta.title} — tene`,
    description: post.meta.description,
    keywords: post.meta.tags,
    alternates: {
      canonical,
      // RSS auto-discovery is emitted on every page from root layout.tsx.
      // Per-article override of `alternates.types` was removed to prevent
      // duplicate <link rel="alternate" type="application/rss+xml"> tags.
    },
    openGraph: {
      title: post.meta.title,
      description: post.meta.description,
      url: canonical,
      siteName: "Tene",
      type: "article",
      publishedTime: post.meta.publishedAt,
      modifiedTime: post.meta.updatedAt ?? post.meta.publishedAt,
      authors: [post.meta.author ?? "agent-kay"],
      tags: post.meta.tags,
      // blog-seo-enhancements G3 — Next.js auto-wires og:image from the
      // co-located app/blog/[slug]/opengraph-image.tsx file convention.
      // Explicit `images:` was removed so per-article dynamic image wins.
    },
    twitter: {
      card: "summary_large_image",
      title: post.meta.title,
      description: post.meta.description,
      // Same: twitter:image auto-wires to opengraph-image.
    },
    robots: { index: true, follow: true },
  };
}

export default async function BlogPostPage({
  params,
}: {
  params: Promise<Params>;
}) {
  const { slug } = await params;
  const post = getPostBySlug(slug);
  if (!post) notFound();

  const related = getRelatedPosts(slug, post.meta.tags, 3);
  const components = useMDXComponents({}, { slug });

  return (
    <>
      <BlogPostingJsonLd meta={post.meta} />
      {post.meta.faqs && post.meta.faqs.length > 0 && (
        <FaqJsonLd faqs={post.meta.faqs} />
      )}

      <InteractiveGrid />
      <div className="dot-grid-fixed sm:hidden" />

      <Nav />
      <main className="relative z-10">
        <Breadcrumb
          items={[
            { label: "Home", href: "/" },
            { label: "Blog", href: "/blog" },
            { label: post.meta.title },
          ]}
        />
        <PostHero meta={post.meta} />

        {/*
         * Layout grid:
         *  - Up to `lg`, the article fills the natural column (max-w-3xl matches
         *    the hero / footer / related-posts widths so the whole page reads as
         *    one uniform column).
         *  - At `lg`+, we widen to max-w-5xl and split into
         *    [720px article] + [220px TOC], with a sticky TOC on the right.
         *
         * Why `minmax(0, 720px)` (not `minmax(auto, 720px)`): `auto` takes the
         * column's min-content, which for long code blocks = the widest line in
         * the code. That would let the column (and therefore the page) expand
         * beyond 720px and cause horizontal scroll. `0` as the floor + the
         * `<article className="min-w-0">` below lets inner `<pre overflow-x-auto>`
         * handle overflow itself, preventing page-level scroll.
         */}
        <div className="mx-auto grid w-full max-w-3xl gap-8 px-4 pb-12 sm:px-6 lg:max-w-5xl lg:grid-cols-[minmax(0,720px)_220px]">
          <article className="min-w-0">
            <MDXRemote
              source={post.content}
              components={components}
              options={{
                parseFrontmatter: false,
                mdxOptions: {
                  remarkPlugins: [remarkGfm],
                  rehypePlugins: [
                    rehypeSlug,
                    [
                      rehypeAutolinkHeadings,
                      {
                        behavior: "append",
                        properties: {
                          className: ["heading-anchor"],
                          ariaLabel: "Link to section",
                        },
                      },
                    ],
                    [rehypeShiki, { theme: "github-dark" }],
                  ],
                },
              }}
            />
            <div id="blog-end-sentinel" aria-hidden="true" />
            <PostFooter meta={post.meta} />
          </article>

          <aside className="hidden lg:block">
            <div className="sticky top-24">
              <PostTOC />
            </div>
          </aside>
        </div>

        <RelatedPosts fromSlug={slug} posts={related} />
        <ReadProgressTracker
          slug={slug}
          readingMinutes={post.meta.readingMinutes}
        />
      </main>
      <Footer />
    </>
  );
}
