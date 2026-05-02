// Home-only JSON-LD: FAQPage + HowTo. These are page-specific schemas that
// describe content rendered on `/`, so they must NOT live in the root
// layout (which emits on every page). When emitted globally they collide
// with article-level FAQPage on /blog/{slug} and comparison FAQPage on
// /vs/{slug} — GSC flags the duplicate as "FAQPage 입력란이 중복" and
// drops the rich result.
//
// Site-wide entities (Organization, WebSite, SoftwareApplication) stay in
// `layout.tsx` because they are `@id`-keyed and reused across pages.
//
// Source-of-truth note (2026-05-03): the FAQ Q&A list is imported from
// `@/data/faq`, which is the SAME data used by the visible <FAQ /> UI
// component on the home page. Hardcoding a separate set in this file used
// to drift — UI showed 8 Q&A while JSON-LD emitted a different 6 Q&A,
// triggering GSC's "FAQPage 입력란이 중복" / Schema-vs-content mismatch.
// Per Google policy ("All FAQ markup must match visible content"), the
// schema must reflect exactly what users see. Importing one array enforces
// this at build time.
import { faqs } from "@/data/faq";

const homeJsonLd = {
  "@context": "https://schema.org",
  "@graph": [
    {
      "@type": "FAQPage",
      mainEntity: faqs.map((f) => ({
        "@type": "Question",
        name: f.question,
        acceptedAnswer: {
          "@type": "Answer",
          text: f.answer,
        },
      })),
    },
    {
      "@type": "HowTo",
      name: "How to use Tene for secret management",
      step: [
        {
          "@type": "HowToStep",
          name: "Install",
          text: "Run: curl -sSfL https://tene.sh/install.sh | sh — or download from GitHub Releases (github.com/tomo-kay/tene/releases)",
        },
        {
          "@type": "HowToStep",
          name: "Initialize",
          text: "Run tene init to create an encrypted vault and generate context files for Claude, Cursor, Windsurf, Gemini, and Codex.",
        },
        {
          "@type": "HowToStep",
          name: "Store secrets",
          text: "Run tene set KEY value to encrypt and store secrets locally.",
        },
        {
          "@type": "HowToStep",
          name: "Develop with secrets",
          text: "Run tene run -- your-command to inject secrets as environment variables.",
        },
      ],
    },
  ],
};

export function HomeJsonLd() {
  return (
    <script
      type="application/ld+json"
      dangerouslySetInnerHTML={{ __html: JSON.stringify(homeJsonLd) }}
    />
  );
}
