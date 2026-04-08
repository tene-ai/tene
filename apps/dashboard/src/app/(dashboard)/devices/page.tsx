"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { useAuthReady } from "@/hooks/use-auth-ready";
import { api } from "@/lib/api";
import { DeviceCard, DeviceCardEmpty } from "@/components/device-card";

export default function DevicesPage() {
  const authReady = useAuthReady();
  const queryClient = useQueryClient();

  const { data: devices, isLoading } = useQuery({
    queryKey: ["devices"],
    queryFn: () => api.listDevices(),
    enabled: authReady,
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => api.deleteDevice(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["devices"] });
    },
  });

  const handleRevoke = (id: string) => {
    if (confirm("Revoke this device? It will need to re-authenticate.")) {
      deleteMutation.mutate(id);
    }
  };

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold">Devices</h1>
        <p className="text-muted text-sm mt-1">Registered devices for vault sync</p>
      </div>

      {deleteMutation.isError && (
        <div className="rounded-lg border border-danger/30 bg-danger/10 p-3 text-sm text-danger">
          {deleteMutation.error instanceof Error ? deleteMutation.error.message : "Failed to revoke device"}
        </div>
      )}

      {isLoading ? (
        <div className="grid sm:grid-cols-2 lg:grid-cols-3 gap-4">
          {[1, 2, 3].map((i) => (
            <div key={i} className="rounded-xl border border-border bg-surface p-5 animate-pulse h-28" />
          ))}
        </div>
      ) : !devices || devices.length === 0 ? (
        <div className="grid sm:grid-cols-2 lg:grid-cols-3 gap-4">
          <DeviceCardEmpty />
        </div>
      ) : (
        <div className="grid sm:grid-cols-2 lg:grid-cols-3 gap-4">
          {devices.map((device) => (
            <DeviceCard key={device.id} device={device} onRevoke={handleRevoke} />
          ))}
        </div>
      )}
    </div>
  );
}
