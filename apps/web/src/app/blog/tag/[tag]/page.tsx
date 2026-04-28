import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { Nav } from "@/components/nav";
import { Footer } from "@/components/footer";
import { InteractiveGrid } from "@/components/interactive-grid";
import { PostMasonry } from "@/components/blog/post-masonry";
import { BlogIndexJsonLd } from "@/components/seo/blog-index-jsonld";
import { getAllTags, getPostsByTag } from "@/lib/blog";
import { getTagLabel, isValidTag } from "@/lib/tags";

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
  const canonical = `https://tene.sh/blog/tag/${tag}`;

  return {
    title: `${label} articles — tene Tech Blog`,
    description: `Articles tagged ${label} from the tene Tech Blog.`,
    alternates: { canonical },
    openGraph: {
      title: `${label} — tene Tech Blog`,
      description: `Articles tagged ${label}.`,
      url: canonical,
      siteName: "Tene",
      type: "website",
    },
    robots: { index: true, follow: true },
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
            <p className="mt-4 text-muted">
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
