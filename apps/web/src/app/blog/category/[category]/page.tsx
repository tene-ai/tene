import type { Metadata } from "next";
import Link from "next/link";
import { notFound } from "next/navigation";
import { Nav } from "@/components/nav";
import { Footer } from "@/components/footer";
import { InteractiveGrid } from "@/components/interactive-grid";
import { PostMasonry } from "@/components/blog/post-masonry";
import { BlogIndexJsonLd } from "@/components/seo/blog-index-jsonld";
import { getAllCategories, getPostsByCategory } from "@/lib/blog";
import {
  CATEGORY_DESCRIPTIONS,
  getCategoryLabel,
  isValidCategory,
  type CategoryKey,
} from "@/lib/tags";

export const dynamic = "error";

type Params = { category: string };

export function generateStaticParams(): Params[] {
  // All 4 categories are generated even with count=0 — empty categories show
  // a "Coming soon" placeholder rather than 404, so the taxonomy is always
  // discoverable in navigation.
  return getAllCategories().map(({ category }) => ({ category }));
}

export async function generateMetadata({
  params,
}: {
  params: Promise<Params>;
}): Promise<Metadata> {
  const { category } = await params;
  if (!isValidCategory(category)) return {};
  const label = getCategoryLabel(category);
  const description = CATEGORY_DESCRIPTIONS[category as CategoryKey];
  const canonical = `https://tene.sh/blog/category/${category}`;

  return {
    title: `${label} — tene Tech Blog`,
    description,
    alternates: { canonical },
    openGraph: {
      title: `${label} — tene Tech Blog`,
      description,
      url: canonical,
      siteName: "Tene",
      type: "website",
    },
    twitter: {
      card: "summary_large_image",
      title: `${label} — tene Tech Blog`,
      description,
    },
    robots: { index: true, follow: true },
  };
}

export default async function CategoryPage({
  params,
}: {
  params: Promise<Params>;
}) {
  const { category } = await params;
  if (!isValidCategory(category)) notFound();

  const posts = getPostsByCategory(category);
  const label = getCategoryLabel(category);
  const description = CATEGORY_DESCRIPTIONS[category as CategoryKey];

  return (
    <>
      {/* CollectionPage + BreadcrumbList schema (reuses blog index json-ld) */}
      <BlogIndexJsonLd posts={posts} category={category} />

      <InteractiveGrid />
      <div className="dot-grid-fixed sm:hidden" />

      <Nav />
      <main className="relative z-10">
        <section className="px-4 pt-28 pb-8 sm:px-6">
          <div className="mx-auto max-w-4xl text-center">
            <nav aria-label="Breadcrumb" className="text-sm text-muted">
              <Link
                href="/blog"
                className="underline underline-offset-4 hover:text-foreground"
              >
                Blog
              </Link>
              <span className="mx-2 opacity-60">/</span>
              <span className="text-foreground">{label}</span>
            </nav>
            <h1 className="mt-3 text-3xl font-bold sm:text-4xl md:text-5xl">
              {label}
            </h1>
            <p className="mx-auto mt-4 max-w-2xl text-muted">{description}</p>
            <p className="mt-4 text-sm text-muted">
              {posts.length} article{posts.length === 1 ? "" : "s"}
            </p>
          </div>
        </section>

        <section className="px-4 pb-16 sm:px-6">
          {posts.length === 0 ? (
            <div className="mx-auto max-w-xl rounded-lg border border-border bg-surface/50 p-8 text-center text-muted">
              <p className="font-medium text-foreground">Coming soon.</p>
              <p className="mt-2 text-sm">
                No articles in <span className="text-accent">{label}</span> yet.
                Check back soon or{" "}
                <Link
                  href="/blog"
                  className="underline underline-offset-4 hover:text-foreground"
                >
                  browse other categories
                </Link>
                .
              </p>
            </div>
          ) : (
            <PostMasonry posts={posts} />
          )}
        </section>
      </main>
      <Footer />
    </>
  );
}
