// Design Ref: blog-seo-enhancements §2.1 — BlogPosting + BreadcrumbList in a
// single @graph. Adds articleSection (G4) and author.sameAs (G5) alongside
// the breadcrumb trail (G1).
import { getTagLabel } from "@/lib/tags";
import type { BlogPostMeta } from "@/lib/blog";

type Props = {
  meta: BlogPostMeta;
};

export function BlogPostingJsonLd({ meta }: Props) {
  const canonical = meta.canonicalUrl!;
  const primaryTag = meta.tags[0];

  const ld = {
    "@context": "https://schema.org",
    "@graph": [
      {
        "@type": "BlogPosting",
        headline: meta.title,
        description: meta.description,
        datePublished: meta.publishedAt,
        dateModified: meta.updatedAt ?? meta.publishedAt,
        // G4 — articleSection uses the first tag's display label
        ...(primaryTag ? { articleSection: getTagLabel(primaryTag) } : {}),
        author: {
          "@type": "Person",
          name: meta.author ?? "tomo-kay",
          url: "https://github.com/tomo-kay",
          // G5 — sameAs authorship graph
          sameAs: ["https://github.com/tomo-kay"],
        },
        publisher: {
          "@type": "Organization",
          name: "Tene",
          url: "https://tene.sh",
          logo: {
            "@type": "ImageObject",
            url: "https://tene.sh/logo.svg",
          },
        },
        mainEntityOfPage: {
          "@type": "WebPage",
          "@id": canonical,
        },
        image: meta.cover ?? "https://tene.sh/og-image.png",
        keywords: meta.tags.join(", "),
        wordCount: meta.wordCount,
        timeRequired: `PT${meta.readingMinutes}M`,
        inLanguage: "en-US",
      },
      // G1 — BreadcrumbList: Home > Blog > Article
      {
        "@type": "BreadcrumbList",
        itemListElement: [
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
            name: meta.title,
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
