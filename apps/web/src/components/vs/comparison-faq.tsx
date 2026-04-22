import type { ComparisonFAQ } from "@/data/comparisons/types";

type Props = {
  faqs: ComparisonFAQ[];
};

export function ComparisonFaqList({ faqs }: Props) {
  if (faqs.length === 0) return null;

  return (
    <section className="px-4 py-12 sm:px-6">
      <div className="mx-auto max-w-3xl">
        <h2 className="text-2xl font-bold sm:text-3xl">FAQ</h2>

        <dl className="mt-8 divide-y divide-border">
          {faqs.map((faq) => (
            <details
              key={faq.question}
              className="group py-4 [&_summary::-webkit-details-marker]:hidden"
            >
              <summary className="flex cursor-pointer items-start justify-between gap-4 font-medium">
                <dt className="flex-1">{faq.question}</dt>
                <span className="mt-1 shrink-0 text-accent transition-transform group-open:rotate-45">
                  +
                </span>
              </summary>
              <dd className="mt-3 text-muted leading-relaxed">{faq.answer}</dd>
            </details>
          ))}
        </dl>
      </div>
    </section>
  );
}
