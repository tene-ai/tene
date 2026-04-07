interface SecretKeyRowProps {
  name: string;
  env: string;
  updatedAt: string;
}

export function SecretKeyRow({ name, env, updatedAt }: SecretKeyRowProps) {
  return (
    <tr className="border-b border-border last:border-0 hover:bg-surface-2 transition-colors">
      <td className="px-4 py-3 font-mono text-sm text-accent">{name}</td>
      <td className="px-4 py-3">
        <span className="font-mono text-muted tracking-widest">••••••••</span>
      </td>
      <td className="px-4 py-3 hidden sm:table-cell">
        <span className={`badge ${env === "prod" ? "badge-danger" : env === "staging" ? "badge-push" : "badge-active"}`}>
          {env}
        </span>
      </td>
      <td className="px-4 py-3 hidden md:table-cell text-xs text-muted">
        {new Date(updatedAt).toLocaleDateString()}
      </td>
      <td className="px-4 py-3">
        <button
          disabled
          className="text-xs text-muted/40 cursor-not-allowed"
          title="Secret values can only be viewed via CLI: tene get KEY"
          aria-label={`View value for ${name} — disabled, use CLI`}
        >
          👁 View
        </button>
      </td>
    </tr>
  );
}
