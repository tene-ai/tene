// Design Ref: §2.2 T2-1 — Blog post registry. Reads from content/blog/*.mdx at
// build time. No DB, no CMS — git is the source of truth. Plan D2, C-01.
import fs from "node:fs";
import path from "node:path";
import matter from "gray-matter";
import readingTimeLib from "reading-time";

const CONTENT_DIR = path.join(process.cwd(), "content", "blog");

export type BlogPostFrontmatter = {
  slug: string;
  title: string;
  description: string;
  publishedAt: string; // ISO 8601 (YYYY-MM-DD OK)
  updatedAt?: string;
  tags: string[];
  author?: string; // default: "tomo-kay"
  cover?: string;
  canonicalUrl?: string; // default: https://tene.sh/blog/{slug}
  draft?: boolean;
  faqs?: Array<{ question: string; answer: string }>;
};

export type BlogPostMeta = BlogPostFrontmatter & {
  readingMinutes: number;
  wordCount: number;
};

function contentDirExists(): boolean {
  try {
    return fs.existsSync(CONTENT_DIR);
  } catch {
    return false;
  }
}

export function getAllPostSlugs(): string[] {
  if (!contentDirExists()) return [];
  return fs
    .readdirSync(CONTENT_DIR)
    .filter((f) => f.endsWith(".mdx"))
    .map((f) => f.replace(/\.mdx$/, ""));
}

type LoadedPost = { meta: BlogPostMeta; content: string };

function loadPost(slug: string): LoadedPost | null {
  const filePath = path.join(CONTENT_DIR, `${slug}.mdx`);
  if (!fs.existsSync(filePath)) return null;
  const raw = fs.readFileSync(filePath, "utf-8");
  const { data, content } = matter(raw);
  const rt = readingTimeLib(content);

  const meta: BlogPostMeta = {
    slug: (data.slug as string) ?? slug,
    title: data.title as string,
    description: data.description as string,
    publishedAt: data.publishedAt as string,
    updatedAt: data.updatedAt as string | undefined,
    tags: (data.tags as string[]) ?? [],
    author: (data.author as string) ?? "tomo-kay",
    cover: data.cover as string | undefined,
    canonicalUrl:
      (data.canonicalUrl as string) ??
      `https://tene.sh/blog/${(data.slug as string) ?? slug}`,
    draft: (data.draft as boolean) ?? false,
    faqs: data.faqs as Array<{ question: string; answer: string }> | undefined,
    readingMinutes: Math.max(1, Math.ceil(rt.minutes)),
    wordCount: rt.words,
  };
  return { meta, content };
}

export function getAllPosts(
  { includeDrafts = false }: { includeDrafts?: boolean } = {},
): BlogPostMeta[] {
  return getAllPostSlugs()
    .map((slug) => loadPost(slug))
    .filter((p): p is LoadedPost => p !== null)
    .filter(({ meta }) => includeDrafts || !meta.draft)
    .map(({ meta }) => meta)
    .sort((a, b) => b.publishedAt.localeCompare(a.publishedAt));
}

export function getPostBySlug(slug: string): LoadedPost | null {
  const post = loadPost(slug);
  if (!post || post.meta.draft) return null;
  return post;
}

export function getRelatedPosts(
  slug: string,
  tags: string[],
  limit = 3,
): BlogPostMeta[] {
  const all = getAllPosts();
  const scored = all
    .filter((p) => p.slug !== slug)
    .map((p) => ({
      post: p,
      overlap: p.tags.filter((t) => tags.includes(t)).length,
    }))
    .filter(({ overlap }) => overlap > 0)
    .sort(
      (a, b) =>
        b.overlap - a.overlap ||
        b.post.publishedAt.localeCompare(a.post.publishedAt),
    );

  // Fill with latest posts if not enough tag overlap found.
  if (scored.length < limit) {
    const fillers = all.filter(
      (p) => p.slug !== slug && !scored.some((s) => s.post.slug === p.slug),
    );
    return [...scored.map((s) => s.post), ...fillers].slice(0, limit);
  }
  return scored.slice(0, limit).map(({ post }) => post);
}

export function getAllTags(): Array<{ tag: string; count: number }> {
  const counts = new Map<string, number>();
  for (const post of getAllPosts()) {
    for (const tag of post.tags) {
      counts.set(tag, (counts.get(tag) ?? 0) + 1);
    }
  }
  return [...counts.entries()]
    .map(([tag, count]) => ({ tag, count }))
    .sort((a, b) => b.count - a.count);
}

export function getPostsByTag(tag: string): BlogPostMeta[] {
  return getAllPosts().filter((p) => p.tags.includes(tag));
}
