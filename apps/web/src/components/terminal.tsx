export function Terminal() {
  return (
    <section className="px-4 py-20 sm:px-6">
      <div className="mx-auto max-w-3xl">
        <div className="overflow-hidden rounded-xl border border-border bg-surface">
          <div className="flex items-center gap-2 border-b border-border px-4 py-3">
            <div className="h-3 w-3 rounded-full bg-[#ff5f57]" />
            <div className="h-3 w-3 rounded-full bg-[#febc2e]" />
            <div className="h-3 w-3 rounded-full bg-[#28c840]" />
            <span className="ml-3 text-xs text-muted font-mono">~/my-project</span>
          </div>
          <div className="overflow-x-auto p-4 font-mono text-xs leading-7 sm:p-6 sm:text-sm">
            <Line prompt>brew install tomo-kay/tap/tene</Line>
            <Line />
            <Line prompt>tene init</Line>
            <Line dim>  Master Password: ********</Line>
            <Line green>  ✓ .tene/vault.db created</Line>
            <Line green>  ✓ CLAUDE.md created — Claude Code will auto-detect tene</Line>
            <Line green>  ✓ .tene/ added to .gitignore</Line>
            <Line />
            <Line dim>  Recovery Key:</Line>
            <Line accent>  apple banana cherry dolphin eagle frost</Line>
            <Line accent>  grape harbor island jungle kite lemon</Line>
            <Line />
            <Line prompt>tene set STRIPE_KEY sk_test_51Hxxxxx</Line>
            <Line green>  ✓ STRIPE_KEY saved (encrypted)</Line>
            <Line />
            <Line prompt>tene set OPENAI_API_KEY sk-proj-xxxxx</Line>
            <Line green>  ✓ OPENAI_API_KEY saved (encrypted)</Line>
            <Line />
            <Line prompt>tene run -- claude</Line>
            <Line green>  ✓ 2 secrets injected as environment variables</Line>
            <Line green>  ✓ Starting: claude</Line>
            <Line />
            <Line dim>  {"// Claude Code reads CLAUDE.md and knows:"}</Line>
            <Line dim>  {"// \"This project uses tene for secret management.\""}</Line>
            <Line dim>  {"// \"Use tene get <KEY> to retrieve secrets.\""}</Line>
          </div>
        </div>

        <p className="mt-6 text-center text-sm text-muted">
          From install to first secret injection — under 1 minute. No account needed.
        </p>
      </div>
    </section>
  );
}

function Line({
  children,
  prompt,
  green,
  accent,
  dim,
}: {
  children?: React.ReactNode;
  prompt?: boolean;
  green?: boolean;
  accent?: boolean;
  dim?: boolean;
}) {
  if (!children && !prompt) return <div className="h-4" />;

  return (
    <div
      className={`whitespace-nowrap ${green ? "text-[#28c840]" : ""} ${accent ? "text-accent font-semibold" : ""} ${dim ? "text-muted" : ""}`}
    >
      {prompt && <span className="text-accent mr-2">$</span>}
      {children}
    </div>
  );
}
