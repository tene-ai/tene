"use client";

import { useEffect } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { useAuthStore } from "@/lib/auth-store";

export default function AuthCallbackPage() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const login = useAuthStore((s) => s.login);

  useEffect(() => {
    const accessToken = searchParams.get("access_token");
    const refreshToken = searchParams.get("refresh_token");

    if (accessToken && refreshToken) {
      // Set cookie so Next.js middleware can verify auth on server-side
      document.cookie = `tene_access_token=${accessToken}; path=/; max-age=${15 * 60}; SameSite=Lax`;
      // Store in Zustand (persisted to localStorage)
      login(accessToken, refreshToken);
      router.replace("/");
    } else {
      router.replace("/login");
    }
  }, [searchParams, login, router]);

  return (
    <div className="min-h-screen flex items-center justify-center">
      <div className="text-center">
        <h1 className="font-mono font-bold text-2xl text-accent mb-2">tene</h1>
        <p className="text-muted text-sm">Signing you in...</p>
      </div>
    </div>
  );
}
