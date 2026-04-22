// Design Ref: blog-seo-enhancements §2.2 — G2 CollectionPage/Blog schema for
// /blog and /blog/tag/[tag]. AI search engines use this to identify the
// section as a blog and enumerate articles for potential citation.
import type { BlogPostMeta } from "@/lib/blog";
import { getTagLabel } from "@/lib/tags";

type Props = {
  posts: BlogPostMeta[];
  // undefined = main /blog index; string = /blog/tag/{tag} view
  tag?: string;
};

export function BlogIndexJsonLd({ posts, tag }: Props) {
  const isTagPage = !!tag;
  const canonical = isTagPage
    ? `https://tene.sh/blog/tag/${tag}`
    : "https://tene.sh/blog";
  const label = isTagPage ? getTagLabel(tag!) : "tene Tech Blog";

  const ld = {
    "@context": "https://schema.org",
    "@graph": [
      {
        "@type": isTagPage ? "CollectionPage" : "Blog",
        name: isTagPage ? `${label} articles` : "tene Tech Blog",
        description: isTagPage
          ? `Articles tagged ${label} from the tene Tech Blog.`
          : "AI-safe secrets · Vibe coding · Developer security · Local-first infrastructure",
        url: canonical,
        inLanguage: "en-US",
        publisher: {
          "@type": "Organization",
          name: "Tene",
          url: "https://tene.sh",
          logo: {
            "@type": "ImageObject",
            url: "https://tene.sh/logo.svg",
          },
        },
        hasPart: posts.map((p) => ({
          "@type": "BlogPosting",
          headline: p.title,
          description: p.description,
          url: `https://tene.sh/blog/${p.slug}`,
          datePublished: p.publishedAt,
          dateModified: p.updatedAt ?? p.publishedAt,
          keywords: p.tags.join(", "),
          author: {
            "@type": "Person",
            name: p.author ?? "tomo-kay",
          },
        })),
      },
      // BreadcrumbList: Home > Blog (> #tag)
      {
        "@type": "BreadcrumbList",
        itemListElement: isTagPage
          ? [
              {
                "@type": "ListItem",
                position: 1,
                name: "Home",
                item: "https://tene.sh",
              },
              {
                "@type": "ListItem",
                position: 2,
                name: "Blog",
                item: "https://tene.sh/blog",
              },
              {
                "@type": "ListItem",
                position: 3,
                name: `#${label}`,
                item: canonical,
              },
            ]
          : [
              {
                "@type": "ListItem",
                position: 1,
                name: "Home",
                item: "https://tene.sh",
              },
              {
                "@type": "ListItem",
                position: 2,
                name: "Blog",
                item: canonical,
              },
            ],
      },
    ],
  };

  return (
    <script
      type="application/ld+json"
      dangerouslySetInnerHTML={{ __html: JSON.stringify(ld) }}
    />
  );
}
