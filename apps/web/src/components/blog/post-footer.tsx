import { CopyCommand } from "@/components/copy-command";
import { TrackedGithubLink } from "@/components/tracked-github-link";
import { TagChip } from "@/components/blog/tag-chip";
import type { BlogPostMeta } from "@/lib/blog";

type Props = {
  meta: BlogPostMeta;
};

export function PostFooter({ meta }: Props) {
  const githubEditUrl = `https://github.com/tene-ai/tene/edit/main/apps/web/content/blog/${meta.slug}.mdx`;

  return (
    <footer className="mt-16 border-t border-border/50 pt-8">
      <div className="flex flex-wrap gap-2">
        {meta.tags.map((tag) => (
          <TagChip key={tag} tag={tag} from="post_header" size="md" />
        ))}
      </div>

      <div className="mt-8 rounded-lg border border-border bg-surface/80 p-6 backdrop-blur-sm">
        <h3 className="text-lg font-semibold">Like this article?</h3>
        <p className="mt-2 text-sm text-muted">
          tene is a local-first encrypted secret manager CLI. Install with one line
          and keep secrets out of every AI agent&apos;s context window.
        </p>
        <div className="mt-4 flex">
          <CopyCommand
            command="curl -sSfL https://tene.sh/install.sh | sh"
            className="w-full justify-start text-xs sm:text-sm"
            source="blog_post"
          />
        </div>
      </div>

      <div className="mt-8 text-sm text-muted">
        <TrackedGithubLink
          href={githubEditUrl}
          location="blog_post"
          className="underline underline-offset-4 hover:text-foreground"
        >
          Edit this article on GitHub
        </TrackedGithubLink>
      </div>
    </footer>
  );
}
