import Link from "next/link";
import { GlowCard } from "@/components/glow-card";
import { CategoryBadge } from "@/components/blog/category-badge";
import type { BlogPostMeta } from "@/lib/blog";

type Props = {
  post: BlogPostMeta;
  // When true, render the larger featured layout (2-col span + 21:9 hero +
  // larger title). Used for the latest post when it has a thumbnail.
  featured?: boolean;
};

// en-US fixed format ("May 2, 2026"). The site is English-only
// (`inLanguage: 'en-US'`) so the visible date stays English regardless of
// browser locale. `timeZone: "UTC"` pins the rendered calendar day to the
// publish day; otherwise UTC+13 (Auckland) viewers would see a date one day
// ahead of the author's intent.
function formatDate(iso: string): string {
  try {
    const anchor = /T\d/.test(iso) ? iso : `${iso}T12:00:00.000Z`;
    return new Date(anchor).toLocaleDateString("en-US", {
      year: "numeric",
      month: "short",
      day: "numeric",
      timeZone: "UTC",
    });
  } catch {
    return iso;
  }
}

export function PostCard({ post, featured = false }: Props) {
  const visibleTags = post.tags.slice(0, 2);
  const hiddenTagCount = post.tags.length - visibleTags.length;
  const hasThumb = !!post.thumbnail;
  const aspectClass = featured ? "aspect-[21/9]" : "aspect-[16/9]";
  const titleClass = featured
    ? "text-xl font-semibold leading-tight sm:text-2xl"
    : "text-lg font-semibold leading-tight";

  return (
    <GlowCard className="rounded-lg border border-border bg-surface/80 backdrop-blur-sm transition-colors hover:border-accent/40">
      <Link href={`/blog/${post.slug}`} className="block overflow-hidden rounded-lg">
        {hasThumb && (
          <div className={`${aspectClass} w-full overflow-hidden bg-surface-2`}>
            <img
              src={post.thumbnail}
              alt=""
              loading="lazy"
              decoding="async"
              className="h-full w-full object-cover"
            />
          </div>
        )}
        <div className="p-6">
          <div className="flex flex-wrap items-center gap-2">
            <CategoryBadge category={post.category} />
            <time className="text-xs text-muted" dateTime={post.publishedAt}>
              {formatDate(post.publishedAt)} · {post.readingMinutes} min read
            </time>
          </div>
          <h3 className={`mt-3 ${titleClass}`}>{post.title}</h3>
          <p className="mt-2 text-sm text-muted line-clamp-3">
            {post.description}
          </p>
          {post.tags.length > 0 && (
            <div className="mt-4 flex flex-wrap items-center gap-1.5">
              {visibleTags.map((tag) => (
                <span
                  key={tag}
                  className="inline-flex items-center gap-0.5 rounded-full border border-border/60 px-2 py-0.5 text-xs text-muted"
                >
                  <span className="text-accent">#</span>
                  {tag}
                </span>
              ))}
              {hiddenTagCount > 0 && (
                <span className="text-xs text-muted">+{hiddenTagCount} more</span>
              )}
            </div>
          )}
        </div>
      </Link>
    </GlowCard>
  );
}
