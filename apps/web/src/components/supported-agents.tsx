const agents = [
  { name: "Claude Code", icon: "claude" },
  { name: "Cursor", icon: "cursor" },
  { name: "Windsurf", icon: "windsurf" },
  { name: "Gemini", icon: "gemini" },
  { name: "Codex", icon: "codex" },
];

export function SupportedAgents() {
  return (
    <section className="px-4 py-12 sm:px-6">
      <div className="mx-auto max-w-3xl text-center">
        <p className="text-sm font-medium text-muted">
          Works with your AI Agents
        </p>
        <div className="mt-4 flex flex-wrap items-center justify-center gap-x-8 gap-y-3">
          {agents.map((agent) => (
            <div
              key={agent.name}
              className="flex items-center gap-2 text-sm text-foreground/70 transition-colors hover:text-foreground"
            >
              <AgentIcon name={agent.icon} />
              <span>{agent.name}</span>
            </div>
          ))}
        </div>
        <p className="mt-4 text-xs text-muted">
          <code className="rounded bg-surface px-1.5 py-0.5 text-accent">tene init</code>
          {" "}auto-generates context files for each agent
        </p>
      </div>
    </section>
  );
}

function AgentIcon({ name }: { name: string }) {
  const cls = "h-4 w-4";

  switch (name) {
    case "claude":
      return (
        <svg className={cls} viewBox="0 0 24 24" fill="currentColor">
          <path d="M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5" stroke="currentColor" fill="none" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
        </svg>
      );
    case "cursor":
      return (
        <svg className={cls} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
          <path d="M4 4l7.07 17 2.51-7.39L21 11.07z" />
        </svg>
      );
    case "windsurf":
      return (
        <svg className={cls} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
          <path d="M9.59 4.59A2 2 0 1111 8H2m10.59 11.41A2 2 0 1014 16H2m15.73-8.27A2.5 2.5 0 1119.5 12H2" />
        </svg>
      );
    case "gemini":
      return (
        <svg className={cls} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
          <path d="M12 3v18M3 12h18M5.636 5.636l12.728 12.728M18.364 5.636L5.636 18.364" />
        </svg>
      );
    case "codex":
      return (
        <svg className={cls} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
          <polyline points="16 18 22 12 16 6" />
          <polyline points="8 6 2 12 8 18" />
        </svg>
      );
    default:
      return null;
  }
}
