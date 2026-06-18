// Design Ref: §3.1 — Hero copy data separated from UI.
// The `sub` paragraph opens with "Tene is a..." so AI agents that quote the
// page get a single self-contained definition sentence (ai-discoverability
// A-5).

export const heroData = {
  badge: "Open source · Local-first · Free",
  h1: "Your .env is not a secret.",
  h1Accent: "AI can read it.",
  sub: "Tene is a local-first encrypted secret manager CLI. It encrypts your API keys with XChaCha20-Poly1305 and injects them at runtime, so Claude Code, Cursor, and other AI agents never see plaintext values.",
  cta: {
    install: "curl -sSfL https://tene.sh/install.sh | sh",
    primary: {
      label: "View on GitHub",
      href: "https://github.com/tene-ai/tene",
    },
    secondary: {
      label: "Quickstart",
      href: "https://github.com/tene-ai/tene#quick-start",
    },
  },
};
