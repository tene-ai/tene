interface Device {
  id: string;
  device_name: string;
  last_seen_at: string;
  created_at: string;
}

interface DeviceCardProps {
  device: Device;
  onRevoke?: (id: string) => void;
}

export function DeviceCard({ device, onRevoke }: DeviceCardProps) {
  const isOnline = Date.now() - new Date(device.last_seen_at).getTime() < 5 * 60 * 1000;

  return (
    <div className="rounded-xl border border-border bg-surface p-5 transition-colors hover:border-accent/20">
      <div className="flex items-start justify-between">
        <div className="flex items-center gap-3">
          <span className="font-mono text-xl" aria-hidden>⊞</span>
          <div>
            <p className="font-medium text-sm">{device.device_name}</p>
            <p className="text-xs text-muted mt-0.5">
              Added {new Date(device.created_at).toLocaleDateString()}
            </p>
          </div>
        </div>
        <span
          className={`w-2 h-2 rounded-full mt-1 ${isOnline ? "bg-accent" : "bg-muted/50"}`}
          aria-label={isOnline ? "Online" : "Offline"}
        />
      </div>
      <div className="mt-3 flex items-center justify-between text-xs text-muted">
        <span>Last seen: {new Date(device.last_seen_at).toLocaleString()}</span>
        {onRevoke && (
          <button
            onClick={() => onRevoke(device.id)}
            className="text-danger hover:underline"
            aria-label={`Revoke device ${device.device_name}`}
          >
            Revoke
          </button>
        )}
      </div>
    </div>
  );
}

export function DeviceCardEmpty() {
  return (
    <div className="rounded-xl border border-dashed border-border p-8 flex flex-col items-center text-center text-muted">
      <span className="font-mono text-3xl mb-3" aria-hidden>⊞</span>
      <p className="text-sm font-medium">No devices registered</p>
      <p className="text-xs mt-1">
        Devices are registered when you run <code className="font-mono text-accent">tene login</code>
      </p>
    </div>
  );
}
