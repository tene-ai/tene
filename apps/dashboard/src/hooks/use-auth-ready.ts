import { useAuthStore } from "@/lib/auth-store";

/** Returns true when access token is available (rehydrated or freshly logged in). */
export function useAuthReady(): boolean {
  return useAuthStore((s) => !!s.accessToken);
}
