export default function DevicesPage() {
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold">Devices</h1>
        <p className="text-muted text-sm mt-1">Registered devices for vault sync</p>
      </div>

      <div className="grid sm:grid-cols-2 lg:grid-cols-3 gap-4">
        <div className="rounded-xl border border-dashed border-border p-8 flex flex-col items-center text-center text-muted">
          <span className="font-mono text-3xl mb-3">⊞</span>
          <p className="text-sm font-medium">No devices registered</p>
          <p className="text-xs mt-1">
            Devices are registered when you run <code className="font-mono text-accent">tene login</code>
          </p>
        </div>
      </div>
    </div>
  );
}
