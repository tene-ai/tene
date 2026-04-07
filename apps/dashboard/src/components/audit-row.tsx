interface AuditEvent {
  id: string;
  action: string;
  vault_id?: string;
  ip_address?: string;
  created_at: string;
}

interface AuditRowProps {
  event: AuditEvent;
}

const actionBadge: Record<string, string> = {
  "vault.push": "badge-push",
  "vault.pull": "badge-pull",
  "vault.create": "badge-active",
  "vault.delete": "badge-danger",
  "auth.login": "badge-active",
  "auth.logout": "badge-free",
};

export function AuditRow({ event }: AuditRowProps) {
  const badgeClass = actionBadge[event.action] || "badge-free";

  return (
    <tr className="border-b border-border last:border-0 hover:bg-surface-2 transition-colors">
      <td className="px-4 py-3 text-xs text-muted font-mono">
        {new Date(event.created_at).toLocaleString()}
      </td>
      <td className="px-4 py-3">
        <span className={`badge ${badgeClass}`}>{event.action}</span>
      </td>
      <td className="px-4 py-3 hidden sm:table-cell text-xs text-muted font-mono">
        {event.vault_id ? event.vault_id.slice(0, 8) + "..." : "—"}
      </td>
      <td className="px-4 py-3 hidden md:table-cell text-xs text-muted">
        {event.ip_address || "—"}
      </td>
    </tr>
  );
}
