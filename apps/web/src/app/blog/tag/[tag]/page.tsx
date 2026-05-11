import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { Nav } from "@/components/nav";
import { Footer } from "@/components/footer";
import { InteractiveGrid } from "@/components/interactive-grid";
import { PostMasonry } from "@/components/blog/post-masonry";
import { BlogIndexJsonLd } from "@/components/seo/blog-index-jsonld";
import { getAllTags, getPostsByTag, isIndexableTag } from "@/lib/blog";
import { getTagDescription, getTagLabel, isValidTag } from "@/lib/tags";

export const dynamic = "error";

type Params = { tag: string };

export function generateStaticParams(): Params[] {
  return getAllTags().map(({ tag }) => ({ tag }));
}

export async function generateMetadata({
  params,
}: {
  params: Promise<Params>;
}): Promise<Metadata> {
  const { tag } = await params;
  if (!isValidTag(tag)) return {};
  const label = getTagLabel(tag);
  const description = getTagDescription(tag);
  const canonical = `https://tene.sh/blog/tag/${tag}`;
  // Thin-tag protection: tag pages with < INDEXABLE_TAG_THRESHOLD articles
  // render normally for UX but emit noindex so Google's helpful-content
  // classifier doesn't penalise the whole site for thin archives. They
  // become indexable automatically once the article count crosses the
  // threshold (no manual flip required). See lib/blog.ts.
  const indexable = isIndexableTag(tag);

  return {
    title: `${label} articles — tene Tech Blog`,
    description,
    alternates: { canonical },
    openGraph: {
      title: `${label} — tene Tech Blog`,
      description,
      url: canonical,
      siteName: "Tene",
      type: "website",
    },
    // follow:true even when noindex — Google should still walk the article
    // links from this page (preserves internal link graph for the articles
    // themselves; only the archive page itself is hidden from the index).
    robots: indexable
      ? { index: true, follow: true }
      : { index: false, follow: true },
  };
}

export default async function TagPage({
  params,
}: {
  params: Promise<Params>;
}) {
  const { tag } = await params;
  if (!isValidTag(tag)) notFound();
  const posts = getPostsByTag(tag);
  if (posts.length === 0) notFound();

  const label = getTagLabel(tag);
  const description = getTagDescription(tag);

  return (
    <>
      {/* blog-seo-enhancements G2 — CollectionPage + BreadcrumbList schema */}
      <BlogIndexJsonLd posts={posts} tag={tag} />

      <InteractiveGrid />
      <div className="dot-grid-fixed sm:hidden" />

      <Nav />
      <main className="relative z-10">
        <section className="px-4 pt-28 pb-8 sm:px-6">
          <div className="mx-auto max-w-4xl text-center">
            <div className="text-sm text-muted">
              <a
                href="/blog"
                className="underline underline-offset-4 hover:text-foreground"
              >
                Blog
              </a>
            </div>
            <h1 className="mt-3 text-3xl font-bold sm:text-4xl md:text-5xl">
              <span className="text-accent">#</span>
              {label}
            </h1>
            {/* Tag intro — gives the page unique visible content per tag, the
                visible-body half of the 2026-05-11 fix that reduces Google's
                "thin content" signal for indexable tag archives. */}
            <p className="mt-5 mx-auto max-w-2xl text-base leading-relaxed text-muted">
              {description}
            </p>
            <p className="mt-3 text-sm text-muted">
              {posts.length} article{posts.length > 1 ? "s" : ""}
            </p>
          </div>
        </section>

        <section className="px-4 pb-16 sm:px-6">
          <PostMasonry posts={posts} />
        </section>
      </main>
      <Footer />
    </>
  );
}
