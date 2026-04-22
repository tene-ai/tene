// Design Ref: §2.6 — FAQPage JSON-LD for articles that define faqs[] in
// frontmatter. Unlocks FAQ rich results on Google SERP + AI citations.
type FAQ = { question: string; answer: string };

export function FaqJsonLd({ faqs }: { faqs: FAQ[] }) {
  if (!faqs || faqs.length === 0) return null;

  const ld = {
    "@context": "https://schema.org",
    "@type": "FAQPage",
    mainEntity: faqs.map((f) => ({
      "@type": "Question",
      name: f.question,
      acceptedAnswer: {
        "@type": "Answer",
        text: f.answer,
      },
    })),
  };

  return (
    <script
      type="application/ld+json"
      dangerouslySetInnerHTML={{ __html: JSON.stringify(ld) }}
    />
  );
}
