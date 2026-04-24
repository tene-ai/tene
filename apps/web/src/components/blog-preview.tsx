// Landing page blog preview — one latest post per category (max 4) so the
// home page always showcases the 2-layer taxonomy. Categories with zero
// posts are skipped rather than shown as "Coming soon" cards, keeping the
// landing density on actual content.
import Link from "next/link";
import { PostCard } from "@/components/blog/post-card";
import { getAllCategories, getPostsByCategory } from "@/lib/blog";

export function BlogPreview() {
  const categories = getAllCategories();

  // Pick the latest post per category (if any).
  const highlights = categories
    .map(({ category }) => {
      const posts = getPostsByCategory(category);
      return posts.length > 0 ? posts[0] : null;
    })
    .filter((p): p is NonNullable<typeof p> => p !== null);

  if (highlights.length === 0) return null;

  return (
    <section className="px-4 py-20 sm:px-6">
      <div className="mx-auto max-w-5xl">
        <div className="text-center">
          <h2 className="text-3xl font-bold tracking-tight sm:text-4xl">
            From the Blog
          </h2>
          <p className="mt-3 text-muted">
            Tools · Engineering · Vibe Coding · Philosophy
          </p>
        </div>

        <ul className="mt-10 grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          {highlights.map((post) => (
            <li key={post.slug}>
              <PostCard post={post} />
            </li>
          ))}
        </ul>

        <div className="mt-8 flex justify-center">
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
      </div>
    </section>
  );
}
