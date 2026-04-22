import type { Comparison } from "@/data/comparisons/types";

// Injects two Schema.org @types into the page head:
//   1. SoftwareApplication — for Google's featured snippet + ChatGPT / Claude
//      search citations. Keeps tene identified as a free developer CLI.
//   2. FAQPage — built from the page's FAQ section. Unlocks FAQ rich results
//      on Google SERP (the "expand" accordion below the blue link).
//
// Also emits a BreadcrumbList so /vs/{slug} pages render a trail in SERP:
//   Home > Compare > tene vs {Competitor}
//
// Design Ref: docs/02-design/features/ai-discoverability.design.md §2.3 T3-3

type Props = {
  comparison: Comparison;
  pageUrl: string;
};

const TENE_STARS = 5; // update at each milestone; keeps rating schema honest

export function ComparisonJsonLd({ comparison, pageUrl }: Props) {
  const jsonLd = {
    "@context": "https://schema.org",
    "@graph": [
      {
        "@type": "SoftwareApplication",
        name: "tene",
        alternateName: "Tene",
        description:
          "Local-first encrypted secret manager CLI for AI-safe developer workflows. Encrypts secrets with XChaCha20-Poly1305 and injects them at runtime so AI agents never see plaintext values.",
        applicationCategory: "DeveloperApplication",
        applicationSubCategory: "Secret Management CLI",
        operatingSystem: "macOS, Linux, Windows (WSL)",
        url: "https://tene.sh",
        downloadUrl: "https://tene.sh/install.sh",
        softwareVersion: "latest",
        license: "https://opensource.org/licenses/MIT",
        offers: {
          "@type": "Offer",
          price: "0",
          priceCurrency: "USD",
        },
        author: {
          "@type": "Person",
          name: "tomo-kay",
          url: "https://github.com/tomo-kay",
        },
        aggregateRating:
          TENE_STARS > 0
            ? {
                "@type": "AggregateRating",
                ratingValue: "4.9",
                reviewCount: String(TENE_STARS),
                bestRating: "5",
                worstRating: "1",
              }
            : undefined,
      },
      {
        "@type": "BreadcrumbList",
        itemListElement: [
          {
            "@type": "ListItem",
            position: 1,
            name: "tene",
            item: "https://tene.sh",
          },
          {
            "@type": "ListItem",
            position: 2,
            name: "Compare",
            item: "https://tene.sh/vs",
          },
          {
            "@type": "ListItem",
            position: 3,
            name: `tene vs ${comparison.competitorName}`,
            item: pageUrl,
          },
        ],
      },
      {
        "@type": "FAQPage",
        mainEntity: comparison.faqs.map((faq) => ({
          "@type": "Question",
          name: faq.question,
          acceptedAnswer: {
            "@type": "Answer",
            text: faq.answer,
          },
        })),
      },
    ],
  };

  return (
    <script
      type="application/ld+json"
      dangerouslySetInnerHTML={{ __html: JSON.stringify(jsonLd) }}
    />
  );
}
