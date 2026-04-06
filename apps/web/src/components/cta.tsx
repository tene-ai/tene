import { CopyCommand } from "./copy-command";

export function CTA() {
  return (
    <section className="px-4 py-24 sm:px-6">
      <div className="mx-auto max-w-2xl rounded-2xl border border-accent/20 bg-accent/5 p-8 text-center sm:p-12">
        <h2 className="text-2xl font-bold sm:text-3xl md:text-4xl">
          Stop worrying about secrets.
        </h2>
        <p className="mx-auto mt-4 max-w-md text-sm text-muted sm:text-base">
          Install Tene, init your project, and let Claude Code handle the rest.
          No signup. No server. Free forever.
        </p>

        <div className="mt-8 flex justify-center">
          <CopyCommand command="curl -sSfL https://tene.sh/install.sh | sh" className="relative border-accent/30 text-xs sm:text-sm" />
        </div>

        <div className="mt-6 flex flex-col items-center gap-3 sm:flex-row sm:justify-center">
          <a
            href="https://github.com/tomo-kay/tene/releases"
            target="_blank"
            rel="noopener noreferrer"
            className="flex w-full items-center justify-center gap-2 rounded-lg bg-accent px-6 py-3 text-sm font-semibold text-background transition-colors hover:bg-accent-dim sm:w-auto"
          >
            Download binary
          </a>
          <a
            href="https://github.com/tomo-kay/tene"
            target="_blank"
            rel="noopener noreferrer"
            className="flex w-full items-center justify-center gap-2 rounded-lg border border-border px-6 py-3 text-sm font-medium transition-colors hover:bg-surface sm:w-auto"
          >
            <svg className="h-4 w-4" fill="currentColor" viewBox="0 0 24 24">
              <path d="M12 0C5.37 0 0 5.37 0 12c0 5.3 3.438 9.8 8.205 11.385.6.113.82-.258.82-.577 0-.285-.01-1.04-.015-2.04-3.338.724-4.042-1.61-4.042-1.61-.546-1.385-1.335-1.755-1.335-1.755-1.087-.744.084-.729.084-.729 1.205.084 1.838 1.236 1.838 1.236 1.07 1.835 2.809 1.305 3.495.998.108-.776.417-1.305.76-1.605-2.665-.3-5.466-1.332-5.466-5.93 0-1.31.465-2.38 1.235-3.22-.135-.303-.54-1.523.105-3.176 0 0 1.005-.322 3.3 1.23.96-.267 1.98-.399 3-.405 1.02.006 2.04.138 3 .405 2.28-1.552 3.285-1.23 3.285-1.23.645 1.653.24 2.873.12 3.176.765.84 1.23 1.91 1.23 3.22 0 4.61-2.805 5.625-5.475 5.92.42.36.81 1.096.81 2.22 0 1.606-.015 2.896-.015 3.286 0 .315.21.69.825.57C20.565 21.795 24 17.295 24 12 24 5.37 18.63 0 12 0z" />
            </svg>
            View on GitHub
          </a>
        </div>
      </div>
    </section>
  );
}
