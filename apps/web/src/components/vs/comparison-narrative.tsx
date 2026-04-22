import type { Comparison } from "@/data/comparisons/types";

type Props = {
  intro: string;
  sections: Comparison["sections"];
};

function paragraphs(text: string) {
  return text.split(/\n\n+/).map((para, i) => (
    <p key={i} className="mt-4 text-muted leading-relaxed first:mt-0">
      {para}
    </p>
  ));
}

export function ComparisonNarrative({ intro, sections }: Props) {
  return (
    <section className="px-4 py-12 sm:px-6">
      <div className="mx-auto max-w-3xl">
        <div className="text-base sm:text-lg">{paragraphs(intro)}</div>

        <div className="mt-12 space-y-12">
          {sections.map((section) => (
            <article key={section.heading}>
              <h2 className="text-2xl font-bold sm:text-3xl">
                {section.heading}
              </h2>
              <div className="mt-4 text-base">{paragraphs(section.body)}</div>
            </article>
          ))}
        </div>
      </div>
    </section>
  );
}
