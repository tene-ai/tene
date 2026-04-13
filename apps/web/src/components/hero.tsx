import { heroData } from "@/data/hero";
import { CopyCommand } from "./copy-command";
import { Terminal } from "./terminal";

export function Hero() {
  return (
    <section className="relative flex min-h-[90vh] flex-col items-center justify-center px-4 pt-20 pb-12 sm:px-6">
      <div className="absolute inset-0 bg-[radial-gradient(ellipse_at_center,_var(--tw-gradient-stops))] from-accent/5 via-transparent to-transparent" />

      <div className="relative z-10 mx-auto flex w-full max-w-6xl flex-col items-center gap-12 lg:flex-row lg:items-center lg:gap-10">
        {/* Left: Hero content */}
        <div className="flex-1 text-center lg:text-left">
          <div className="mb-6 inline-flex items-center gap-2 rounded-full border border-border bg-surface px-4 py-1.5 text-sm text-muted">
            <span className="h-2 w-2 rounded-full bg-accent animate-pulse" />
            {heroData.badge}
          </div>

          <h1 className="text-3xl font-bold leading-tight tracking-tight whitespace-nowrap sm:text-4xl md:text-5xl lg:text-[3.25rem] xl:text-6xl">
            {heroData.h1}
            <br />
            <span className="text-accent">{heroData.h1Accent}</span>
          </h1>

          <p className="mx-auto mt-6 max-w-xl text-base text-muted leading-relaxed sm:text-lg lg:mx-0">
            {heroData.sub}
          </p>

          <div className="mt-8 flex flex-col items-center gap-3 sm:mt-10 lg:items-start">
            <CopyCommand command={heroData.cta.install} className="relative w-full justify-center sm:w-auto" />

            <a
              href={heroData.cta.primary.href}
              target="_blank"
              rel="noopener noreferrer"
              className="inline-flex items-center justify-center gap-2 rounded-lg border border-accent/30 bg-accent/10 px-5 py-3 text-sm font-medium text-accent transition-colors hover:bg-accent/20"
            >
              <svg className="h-4 w-4" fill="currentColor" viewBox="0 0 24 24">
                <path d="M12 0C5.37 0 0 5.37 0 12c0 5.3 3.438 9.8 8.205 11.385.6.113.82-.258.82-.577 0-.285-.01-1.04-.015-2.04-3.338.724-4.042-1.61-4.042-1.61-.546-1.385-1.335-1.755-1.335-1.755-1.087-.744.084-.729.084-.729 1.205.084 1.838 1.236 1.838 1.236 1.07 1.835 2.809 1.305 3.495.998.108-.776.417-1.305.76-1.605-2.665-.3-5.466-1.332-5.466-5.93 0-1.31.465-2.38 1.235-3.22-.135-.303-.54-1.523.105-3.176 0 0 1.005-.322 3.3 1.23.96-.267 1.98-.399 3-.405 1.02.006 2.04.138 3 .405 2.28-1.552 3.285-1.23 3.285-1.23.645 1.653.24 2.873.12 3.176.765.84 1.23 1.91 1.23 3.22 0 4.61-2.805 5.625-5.475 5.92.42.36.81 1.096.81 2.22 0 1.606-.015 2.896-.015 3.286 0 .315.21.69.825.57C20.565 21.795 24 17.295 24 12 24 5.37 18.63 0 12 0z" />
              </svg>
              {heroData.cta.primary.label}
            </a>
          </div>
        </div>

        {/* Right: Terminal demo */}
        <div className="w-full flex-1 lg:max-w-lg xl:max-w-xl">
          <Terminal />
        </div>
      </div>
    </section>
  );
}
