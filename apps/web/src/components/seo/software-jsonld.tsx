import type { Comparison } from "@/data/comparisons/types";

// Injects three Schema.org @types into the page head:
//   1. SoftwareApplication — for Google's featured snippet + ChatGPT / Claude
//      search citations. Keeps tene identified as a free developer CLI.
//   2. BreadcrumbList — so /vs/{slug} pages render a trail in SERP:
//      Home > Compare > tene vs {Competitor}
//   3. FAQPage — built from the page's FAQ section. Unlocks FAQ rich results
//      on Google SERP (the "expand" accordion below the blue link).
//
// Note: aggregateRating intentionally omitted — GitHub stars are not the
// same as product reviews (Schema.org reviewCount = reviews, not stars).
// Misusing this field risks Google rich-result penalties. We will reintroduce
// aggregateRating only when we have genuine user-written reviews.
//
// Design Ref: docs/02-design/features/ai-discoverability.design.md §A-3

type Props = {
  comparison: Comparison;
  pageUrl: string;
};

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
          name: "agent-kay",
          url: "https://agentkay.it",
        },
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
