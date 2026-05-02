import { faqs } from "@/data/faq";

// Design Ref: §4.7 — FAQ with data import, .env risk questions first.
//
// Rendered as semantic <details>/<summary> instead of a useState-driven
// accordion. Two reasons:
//   1. SSR HTML must contain every answer text. Google's FAQ policy +
//      Schema.org ↔ visible-content matching require the answers in
//      <FAQPage> JSON-LD to also appear as visible text on the page. The
//      previous `{openIndex === i && (...)}` pattern hid all answer text
//      from SSR output, which the postbuild verifier
//      (verify-blog-indexability.mjs §FAQ-mirror) rejects.
//   2. <details> is keyboard- and screen-reader-accessible by default,
//      with no `aria-expanded` plumbing or focus management code.
//
// The chevron rotation is driven by the `details[open]` selector — no
// React state, no JavaScript needed for the interaction. The block stays
// a server component (no "use client" directive).
export function FAQ() {
  return (
    <section id="faq" className="px-4 py-24 sm:px-6">
      <div className="mx-auto max-w-2xl">
        <h2 className="text-center text-3xl font-bold sm:text-4xl">FAQ</h2>

        <div className="mt-12 divide-y divide-border">
          {faqs.map((faq, i) => (
            <details key={i} className="group">
              <summary className="flex cursor-pointer list-none items-center justify-between py-5 text-left">
                <span className="text-sm font-medium sm:text-base">
                  {faq.question}
                </span>
                <svg
                  className="h-4 w-4 shrink-0 text-muted transition-transform group-open:rotate-180"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                  strokeWidth="2"
                  aria-hidden="true"
                >
                  <path d="M19 9l-7 7-7-7" />
                </svg>
              </summary>
              <p className="pb-5 text-sm leading-relaxed text-muted">
                {faq.answer}
              </p>
            </details>
          ))}
        </div>
      </div>
    </section>
  );
}
