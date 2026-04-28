// Landing page blog preview — latest 4 posts in the same row-major masonry
// grid the /blog index uses. Identical width/breakpoints/card variant logic
// so the section reads as a continuation of the listing experience rather
// than a separate teaser style. Container width matches PostMasonry:
// max-w-4xl (1col) → sm:max-w-4xl (2col) → lg:max-w-6xl (3col) →
// xl:max-w-[1500px] (4col) → 2xl:max-w-[1800px].
import Link from "next/link";
import { PostMasonry } from "@/components/blog/post-masonry";
import { getAllPosts } from "@/lib/blog";

const HOME_PREVIEW_LIMIT = 4;

export function BlogPreview() {
  const latest = getAllPosts().slice(0, HOME_PREVIEW_LIMIT);

  if (latest.length === 0) return null;

  return (
    <section className="px-4 py-20 sm:px-6">
      <div className="text-center">
        <h2 className="text-3xl font-bold tracking-tight sm:text-4xl">
          From the Blog
        </h2>
        <p className="mt-3 text-muted">
          Tools · Engineering · Vibe Coding · Philosophy
        </p>
      </div>

      <div className="mt-10">
        <PostMasonry posts={latest} />
      </div>

      <div className="mt-10 flex justify-center">
        <Link
          href="/blog"
          className="inline-flex items-center gap-2 rounded-md border border-border bg-surface/60 px-4 py-2 text-sm text-muted backdrop-blur-sm transition-colors hover:border-accent/40 hover:text-foreground"
        >
          Browse all posts
          <svg
            xmlns="http://www.w3.org/2000/svg"
            width="14"
            height="14"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
            aria-hidden
          >
            <line x1="5" y1="12" x2="19" y2="12" />
            <polyline points="12 5 19 12 12 19" />
          </svg>
        </Link>
      </div>
    </section>
  );
}
