import Link from "next/link";
import { GlowCard } from "@/components/glow-card";
import { TagChip } from "@/components/blog/tag-chip";
import type { BlogPostMeta } from "@/lib/blog";

type Props = {
  post: BlogPostMeta;
};

function formatDate(iso: string): string {
  try {
    return new Date(iso).toLocaleDateString("en-US", {
      year: "numeric",
      month: "short",
      day: "numeric",
    });
  } catch {
    return iso;
  }
}

export function PostCard({ post }: Props) {
  return (
    <GlowCard className="h-full rounded-lg border border-border bg-surface/80 backdrop-blur-sm transition-colors hover:border-accent/40">
      <Link href={`/blog/${post.slug}`} className="block h-full p-6">
        <time
          className="text-xs text-muted"
          dateTime={post.publishedAt}
        >
          {formatDate(post.publishedAt)} · {post.readingMinutes} min read
        </time>
        <h3 className="mt-3 text-lg font-semibold leading-tight">
          {post.title}
        </h3>
        <p className="mt-2 text-sm text-muted line-clamp-3">
          {post.description}
        </p>
        {post.tags.length > 0 && (
          <div className="mt-4 flex flex-wrap gap-1.5">
            {post.tags.slice(0, 3).map((tag) => (
              <span
                key={tag}
                className="inline-flex items-center gap-0.5 rounded-full border border-border/60 px-2 py-0.5 text-xs text-muted"
              >
                <span className="text-accent">#</span>
                {tag}
              </span>
            ))}
          </div>
        )}
      </Link>
    </GlowCard>
  );
}
