// Home-only JSON-LD: FAQPage + HowTo. These are page-specific schemas that
// describe content rendered on `/`, so they must NOT live in the root
// layout (which emits on every page). When emitted globally they collide
// with article-level FAQPage on /blog/{slug} and comparison FAQPage on
// /vs/{slug} — GSC flags the duplicate as "FAQPage 입력란이 중복" and
// drops the rich result.
//
// Site-wide entities (Organization, WebSite, SoftwareApplication) stay in
// `layout.tsx` because they are `@id`-keyed and reused across pages.

const homeJsonLd = {
  "@context": "https://schema.org",
  "@graph": [
    {
      "@type": "FAQPage",
      mainEntity: [
        {
          "@type": "Question",
          name: "What is Tene?",
          acceptedAnswer: {
            "@type": "Answer",
            text: "Tene is a local-first, encrypted secret management CLI. It stores your API keys, tokens, and credentials in an encrypted SQLite vault on your device. No server, no signup, no cloud dependency.",
          },
        },
        {
          "@type": "Question",
          name: "How does Claude Code auto-detection work?",
          acceptedAnswer: {
            "@type": "Answer",
            text: "When you run tene init, it generates a CLAUDE.md file in your project root. Claude Code reads this file automatically and learns how to use tene to retrieve secrets.",
          },
        },
        {
          "@type": "Question",
          name: "Is Tene free?",
          acceptedAnswer: {
            "@type": "Answer",
            text: "Yes, Tene is 100% free and open source under the MIT license. All local features — encryption, runtime injection, multi-environment, AI editor rules — are free forever with no limits.",
          },
        },
        {
          "@type": "Question",
          name: "Will there be team features?",
          acceptedAnswer: {
            "@type": "Answer",
            text: "Team sync and collaboration features are being designed. The goal is encrypted team sync without a central server. Join the waitlist at tene.sh to get notified when it launches.",
          },
        },
        {
          "@type": "Question",
          name: "How are my secrets encrypted?",
          acceptedAnswer: {
            "@type": "Answer",
            text: "Tene uses XChaCha20-Poly1305 encryption with 256-bit keys derived from your master password via Argon2id. Each secret gets a unique 192-bit nonce.",
          },
        },
        {
          "@type": "Question",
          name: "Does Tene work offline?",
          acceptedAnswer: {
            "@type": "Answer",
            text: "Tene is 100% offline. It makes zero network calls. Your secrets are encrypted and stored locally in a SQLite database.",
          },
        },
      ],
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
