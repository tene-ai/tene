// JSON-LD for /cli (CLI Reference). The /cli page is a high-intent
// reference surface — when an AI assistant answers "how do I install tene"
// or "what does tene run do", it pulls the citation from this page. Without
// page-level structured data, the page only carries the site-wide
// Organization + WebSite + SoftwareApplication graph from layout.tsx and
// is invisible as a reference-citation source.
//
// @graph contents:
//   - TechArticle    — canonical reference for the CLI
//   - HowTo + Steps  — mirrors the 4-step install/init/set/run flow from
//                      components/how-it-works.tsx (kept in sync below)
//   - BreadcrumbList — Home > CLI Reference
//   - FAQPage        — 5 common CLI questions; +3.2x lift in AI Overviews
//                      per Google's published data on FAQ markup
type Props = {
  dateModified: string;
};

const HOWTO_STEPS = [
  {
    name: "Install",
    text: "curl -sSfL https://tene.sh/install.sh | sh",
  },
  {
    name: "Initialize the project",
    text: "tene init",
  },
  {
    name: "Store a secret",
    text: "tene set STRIPE_KEY sk_test_xxx",
  },
  {
    name: "Run a command with secrets injected",
    text: "tene run -- claude",
  },
];

const FAQS = [
  {
    question: "What is tene?",
    answer:
      "tene is a local-first encrypted secret manager CLI. It encrypts API keys with XChaCha20-Poly1305 and injects them at runtime via tene run, so AI coding agents never see plaintext values.",
  },
  {
    question: "How do I install tene?",
    answer:
      "Run curl -sSfL https://tene.sh/install.sh | sh on macOS or Linux. The installer auto-detects your platform, downloads the latest release binary, and places it on your PATH. No Go toolchain or account is required.",
  },
  {
    question: "How do I run a command with secrets injected?",
    answer:
      "tene run -- <command> launches your command with every secret in the active environment exposed as a process-scoped environment variable. The values never appear on stdout or in shell history. Example: tene run -- npm start.",
  },
  {
    question: "Can my AI assistant read tene secrets?",
    answer:
      "No. tene get refuses to print secrets to a non-TTY stdout (exit code 2, STDOUT_SECRET_BLOCKED) so an AI agent or log pipeline cannot pipe the value out. Use tene run -- <command> instead, which keeps secrets in the child process environment only.",
  },
  {
    question: "Where are secrets stored?",
    answer:
      "Encrypted in a local SQLite vault at .tene/vault.db, scoped per project directory. The master key is held in your OS keychain (Keychain on macOS, libsecret on Linux, Credential Manager on Windows). A 12-word BIP-39 recovery key is issued on tene init for offline backup.",
  },
];

export function CliJsonLd({ dateModified }: Props) {
  const canonical = "https://tene.sh/cli";

  const ld = {
    "@context": "https://schema.org",
    "@graph": [
      {
        "@type": "TechArticle",
        headline: "tene CLI Reference",
        description:
          "Canonical reference for every tene command, flag, exit code, and JSON schema.",
        // Full ISO 8601 datetime with timezone (UTC) — Google Rich Results
        // Test rejects date-only values. `dateModified` is already
        // `mtime.toISOString()` which is full ISO; this literal mirrors
        // that format for `datePublished`.
        datePublished: "2026-04-22T00:00:00.000Z",
        dateModified,
        author: {
          "@type": "Person",
          name: "agent-kay",
          url: "https://agentkay.it",
          sameAs: ["https://agentkay.it"],
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
        inLanguage: "en-US",
        proficiencyLevel: "Beginner",
        dependencies: "macOS, Linux, Windows (WSL)",
        keywords:
          "tene, CLI, secret manager, environment variables, XChaCha20, Claude Code, Cursor",
      },
      {
        "@type": "HowTo",
        name: "Install and use tene",
        description:
          "Install tene, create an encrypted vault for your project, store a secret, and run a command with the secret injected as an environment variable.",
        totalTime: "PT1M",
        step: HOWTO_STEPS.map((s, i) => ({
          "@type": "HowToStep",
          position: i + 1,
          name: s.name,
          text: s.text,
        })),
      },
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
            name: "CLI Reference",
            item: canonical,
          },
        ],
      },
      {
        "@type": "FAQPage",
        mainEntity: FAQS.map((f) => ({
          "@type": "Question",
          name: f.question,
          acceptedAnswer: {
            "@type": "Answer",
            text: f.answer,
          },
        })),
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
