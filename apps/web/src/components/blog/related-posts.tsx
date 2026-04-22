"use client";

// Design Ref: §2.4 FR-16 — Tag-based related posts. Fires blog_related_click
// on click (FR-34). Same GlowCard styling as /vs RelatedComparisons.
import Link from "next/link";
import { GlowCard } from "@/components/glow-card";
import { track } from "@/lib/track";
import type { BlogPostMeta } from "@/lib/blog";

type Props = {
  fromSlug: string;
  posts: BlogPostMeta[];
};

export function RelatedPosts({ fromSlug, posts }: Props) {
  if (posts.length === 0) return null;

  return (
    <section className="px-4 py-12 sm:px-6">
      <div className="mx-auto max-w-4xl">
        <h2 className="text-2xl font-bold sm:text-3xl">Related articles</h2>
        <p className="mt-3 text-muted">
          More from the tene Tech Blog on related topics.
        </p>

        <ul className="mt-8 grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {posts.map((p) => (
            <li key={p.slug}>
              <GlowCard className="h-full rounded-lg border border-border bg-surface/80 backdrop-blur-sm transition-colors hover:border-accent/40">
                <Link
                  href={`/blog/${p.slug}`}
                  onClick={() =>
                    track("blog_related_click", {
                      fromSlug,
                      toSlug: p.slug,
                    })
                  }
                  className="block h-full p-5"
                >
                  <div className="text-xs text-muted">
                    {p.readingMinutes} min read
                  </div>
                  <div className="mt-2 font-medium">{p.title}</div>
                  <p className="mt-2 text-sm text-muted line-clamp-3">
                    {p.description}
                  </p>
                </Link>
              </GlowCard>
            </li>
          ))}
        </ul>
      </div>
    </section>
  );
}
