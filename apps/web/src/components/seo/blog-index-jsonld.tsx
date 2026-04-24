// Design Ref: blog-seo-enhancements §2.2 — G2 CollectionPage/Blog schema for
// /blog, /blog/tag/[tag], and /blog/category/[category]. AI search engines use
// this to identify the section as a blog and enumerate articles for citation.
import type { BlogPostMeta } from "@/lib/blog";
import { getCategoryLabel, getTagLabel } from "@/lib/tags";

type Props = {
  posts: BlogPostMeta[];
  // undefined = main /blog index
  // tag set = /blog/tag/{tag} view
  // category set = /blog/category/{category} view
  tag?: string;
  category?: string;
};

export function BlogIndexJsonLd({ posts, tag, category }: Props) {
  const isTagPage = !!tag;
  const isCategoryPage = !!category;

  let canonical: string;
  let label: string;
  let breadcrumbLabel: string;
  if (isTagPage) {
    canonical = `https://tene.sh/blog/tag/${tag}`;
    label = getTagLabel(tag!);
    breadcrumbLabel = `#${label}`;
  } else if (isCategoryPage) {
    canonical = `https://tene.sh/blog/category/${category}`;
    label = getCategoryLabel(category!);
    breadcrumbLabel = label;
  } else {
    canonical = "https://tene.sh/blog";
    label = "tene Tech Blog";
    breadcrumbLabel = "Blog";
  }

  const isFilteredView = isTagPage || isCategoryPage;

  const ld = {
    "@context": "https://schema.org",
    "@graph": [
      {
        "@type": isFilteredView ? "CollectionPage" : "Blog",
        name: isTagPage
          ? `${label} articles`
          : isCategoryPage
            ? `${label} — tene Tech Blog`
            : "tene Tech Blog",
        description: isTagPage
          ? `Articles tagged ${label} from the tene Tech Blog.`
          : isCategoryPage
            ? `Articles in the ${label} category of the tene Tech Blog.`
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
          articleSection: getCategoryLabel(p.category),
          keywords: p.tags.join(", "),
          author: {
            "@type": "Person",
            name: p.author ?? "tomo-kay",
          },
        })),
      },
      // BreadcrumbList: Home > Blog (> filtered view)
      {
        "@type": "BreadcrumbList",
        itemListElement: isFilteredView
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
                name: breadcrumbLabel,
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
