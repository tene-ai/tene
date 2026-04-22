import { TrackedGithubLink } from "./tracked-github-link";

export function Footer() {
  return (
    <footer className="border-t border-border px-4 py-8 sm:px-6">
      <div className="mx-auto flex max-w-5xl flex-col items-center justify-between gap-4 sm:flex-row">
        <div className="flex items-center gap-2 font-mono text-sm text-muted">
          <svg className="h-5 w-5" viewBox="0 0 32 32" fill="none">
            <rect width="32" height="32" rx="6" fill="#1e1e1e"/>
            <path d="M8 20L14 14L8 8" stroke="#00ff88" strokeWidth="3" strokeLinecap="round" strokeLinejoin="round"/>
            <line x1="16" y1="22" x2="24" y2="22" stroke="#00ff88" strokeWidth="3" strokeLinecap="round"/>
          </svg>
          <span>tene</span>
          <span className="text-border">|</span>
          <span>Agentic Secret Runtime</span>
        </div>
        <div className="flex items-center gap-6 text-sm text-muted">
          <TrackedGithubLink
            href="https://github.com/tomo-kay/tene"
            location="footer"
            className="transition-colors hover:text-foreground"
          >
            GitHub
          </TrackedGithubLink>
          <TrackedGithubLink
            href="https://github.com/tomo-kay/tene/issues"
            location="footer"
            className="transition-colors hover:text-foreground"
          >
            Issues
          </TrackedGithubLink>
          <span className="text-border">MIT License</span>
        </div>
      </div>
    </footer>
  );
}
