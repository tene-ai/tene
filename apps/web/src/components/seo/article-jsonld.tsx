// Design Ref: blog-seo-enhancements §2.1 — BlogPosting + BreadcrumbList in a
// single @graph. Adds articleSection (G4) and author.sameAs (G5) alongside
// the breadcrumb trail (G1).
import { getCategoryLabel } from "@/lib/tags";
import type { BlogPostMeta } from "@/lib/blog";

type Props = {
  meta: BlogPostMeta;
};

export function BlogPostingJsonLd({ meta }: Props) {
  const canonical = meta.canonicalUrl!;

  const ld = {
    "@context": "https://schema.org",
    "@graph": [
      {
        "@type": "BlogPosting",
        headline: meta.title,
        description: meta.description,
        datePublished: meta.publishedAt,
        dateModified: meta.updatedAt ?? meta.publishedAt,
        // G4 — articleSection maps to the post's category (taxonomy change
        // 2026-04-24: previously used first tag; category is now the
        // primary navigation axis and a better Schema.org match).
        articleSection: getCategoryLabel(meta.category),
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
        image: meta.cover ?? "https://tene.sh/og-image.webp",
        keywords: meta.tags.join(", "),
        wordCount: meta.wordCount,
        timeRequired: `PT${meta.readingMinutes}M`,
        inLanguage: "en-US",
        // Speakable schema — tells Google Assistant and other voice
        // engines which parts of the article are appropriate to read aloud
        // for "Hey Google, read me this article" style queries. Targets:
        //   - h1: the title
        //   - article h2: section headings (structure cues)
        //   - article h2 + p: the first paragraph after each h2 (which is
        //     the 50-word answer block in our Q-heading pattern)
        speakable: {
          "@type": "SpeakableSpecification",
          cssSelector: ["h1", "article h2", "article h2 + p"],
        },
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
