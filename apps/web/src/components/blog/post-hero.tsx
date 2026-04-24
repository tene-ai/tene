import { TagChip } from "@/components/blog/tag-chip";
import { CategoryBadge } from "@/components/blog/category-badge";
import { TTSButton } from "@/components/blog/tts-button";
import type { BlogPostMeta } from "@/lib/blog";

type Props = {
  meta: BlogPostMeta;
};

function formatDate(iso: string): string {
  try {
    return new Date(iso).toLocaleDateString("en-US", {
      year: "numeric",
      month: "long",
      day: "numeric",
    });
  } catch {
    return iso;
  }
}

export function PostHero({ meta }: Props) {
  return (
    // Top padding is `pt-4` (not `pt-28`) because `<Breadcrumb />` renders
    // just above this header and already carries the fixed-nav clearance.
    // Keeping pt-28 here would double-space the gap between breadcrumb and
    // article title.
    <header className="px-4 pt-4 pb-8 sm:px-6">
      <div className="mx-auto max-w-3xl">
        <div className="flex flex-wrap items-center gap-2 text-sm text-muted">
          <CategoryBadge category={meta.category} asLink />
          <time dateTime={meta.publishedAt}>{formatDate(meta.publishedAt)}</time>
          <span aria-hidden="true">·</span>
          <span>{meta.readingMinutes} min read</span>
          {meta.author && (
            <>
              <span aria-hidden="true">·</span>
              <span>by {meta.author}</span>
            </>
          )}
        </div>

        <h1 className="mt-4 text-3xl font-bold leading-tight tracking-tight sm:text-4xl md:text-5xl">
          {meta.title}
        </h1>

        <p className="mt-6 text-base text-muted leading-relaxed sm:text-lg">
          {meta.description}
        </p>

        <TTSButton slug={meta.slug} readingMinutes={meta.readingMinutes} />

        {meta.tags.length > 0 && (
          <div className="mt-6 flex flex-wrap gap-2">
            {meta.tags.map((tag) => (
              <TagChip key={tag} tag={tag} from="post_header" size="sm" />
            ))}
          </div>
        )}
      </div>
    </header>
  );
}
