import { TagChip } from "@/components/blog/tag-chip";
import { CategoryBadge } from "@/components/blog/category-badge";
import { TTSButton } from "@/components/blog/tts-button";
import type { BlogPostMeta } from "@/lib/blog";

type Props = {
  meta: BlogPostMeta;
};

// en-US fixed format. See post-card.tsx for the rationale (site is
// English-only; UTC anchor keeps the calendar day stable across viewers).
function formatDate(iso: string): string {
  try {
    const anchor = /T\d/.test(iso) ? iso : `${iso}T12:00:00.000Z`;
    return new Date(anchor).toLocaleDateString("en-US", {
      year: "numeric",
      month: "long",
      day: "numeric",
      timeZone: "UTC",
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
              <span className="inline-flex items-center gap-1.5">
                by{" "}
                <a
                  href="https://agentkay.it"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="inline-flex items-center gap-1.5 hover:text-accent transition-colors"
                >
                  <img
                    src="/agent-kay.png"
                    alt=""
                    width={20}
                    height={20}
                    className="h-5 w-5 rounded-full ring-1 ring-border"
                  />
                  {meta.author}
                </a>
              </span>
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
