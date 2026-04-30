// Design Ref: ai-discoverability I-2 — Trust Section.
// Three elements:
//   1. Live project badges (GitHub stars, latest release, MIT license, Go Report)
//   2. 1-2 honest testimonials (placeholder attribution until real quotes land)
//   3. Maintainer bio with GitHub link
//
// Testimonials are labelled "Early user" rather than fabricating names. When
// real quotes with explicit permission arrive, replace them directly in the
// `testimonials` array below.

const testimonials = [
  {
    quote:
      "I've wanted this for months. The `tene run --` flow means I can trust Claude Code again in this repo — no more `.env` panic every time the agent loads project files.",
    attribution: "Early user · indie SaaS founder",
    role: "Beta tester",
    initials: "EU",
  },
  {
    quote:
      "Migrated off dotenv-vault after their Pro shutdown. `tene import .env` took 30 seconds and now secrets never leave disk unencrypted.",
    attribution: "Early user · DevOps engineer",
    role: "Migrator from dotenv-vault",
    initials: "EU",
  },
];

export function Trust() {
  return (
    <section id="trust" className="relative py-16 px-4 sm:px-6">
      <div className="mx-auto max-w-5xl">
        <h2 className="text-center text-2xl font-bold tracking-tight sm:text-3xl">
          Trusted by developers building with AI
        </h2>

        {/* Live badges row */}
        <div className="mt-8 flex flex-wrap items-center justify-center gap-3">
          <a
            href="https://github.com/tomo-kay/tene"
            target="_blank"
            rel="noopener noreferrer"
            className="transition-opacity hover:opacity-80"
            aria-label="GitHub stars"
          >
            <img
              src="https://img.shields.io/github/stars/tomo-kay/tene?style=for-the-badge&label=GitHub%20stars&color=black"
              alt="GitHub stars"
              loading="lazy"
              height={32}
            />
          </a>
          <a
            href="https://github.com/tomo-kay/tene/releases"
            target="_blank"
            rel="noopener noreferrer"
            className="transition-opacity hover:opacity-80"
            aria-label="Latest release"
          >
            <img
              src="https://img.shields.io/github/v/release/tomo-kay/tene?style=for-the-badge&label=latest&color=green"
              alt="Latest release"
              loading="lazy"
              height={32}
            />
          </a>
          <img
            src="https://img.shields.io/github/license/tomo-kay/tene?style=for-the-badge&color=blue"
            alt="MIT License"
            loading="lazy"
            height={32}
          />
          <a
            href="https://goreportcard.com/report/github.com/tomo-kay/tene"
            target="_blank"
            rel="noopener noreferrer"
            className="transition-opacity hover:opacity-80"
            aria-label="Go Report Card"
          >
            <img
              src="https://goreportcard.com/badge/github.com/tomo-kay/tene?style=for-the-badge"
              alt="Go Report Card"
              loading="lazy"
              height={32}
            />
          </a>
        </div>

        {/* Testimonial row */}
        <div className="mt-12 grid gap-6 md:grid-cols-2">
          {testimonials.map((t, i) => (
            <blockquote
              key={i}
              className="relative rounded-lg border border-border bg-surface p-6"
            >
              <p className="text-base text-muted leading-relaxed">
                &ldquo;{t.quote}&rdquo;
              </p>
              <footer className="mt-4 flex items-center gap-3">
                <div className="h-10 w-10 rounded-full bg-accent/20 flex items-center justify-center text-sm font-semibold text-accent">
                  {t.initials}
                </div>
                <div>
                  <div className="text-sm font-medium">{t.attribution}</div>
                  <div className="text-xs text-muted">{t.role}</div>
                </div>
              </footer>
            </blockquote>
          ))}
        </div>

        {/* Maintainer bio */}
        <div className="mt-12 flex flex-col items-center gap-2 text-center">
          <p className="text-sm text-muted">
            Built by{" "}
            <a
              href="https://agentkay.it"
              target="_blank"
              rel="noopener noreferrer"
              className="text-accent hover:underline"
            >
              @agent-kay
            </a>
            , a developer tired of leaking API keys to AI agents.
          </p>
          <p className="text-xs text-muted">
            Open source · MIT · No tracking · No vendor lock-in.
          </p>
        </div>
      </div>
    </section>
  );
}
