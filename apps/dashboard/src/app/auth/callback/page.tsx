"use client";

import { Suspense, useEffect, useRef } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { useAuthStore } from "@/lib/auth-store";
import { api } from "@/lib/api";

function CallbackHandler() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const login = useAuthStore((s) => s.login);
  const exchanged = useRef(false);

  useEffect(() => {
    const code = searchParams.get("code");
    const intent = searchParams.get("intent");

    if (!code) {
      router.replace("/login");
      return;
    }

    // Prevent double exchange in React StrictMode
    if (exchanged.current) return;
    exchanged.current = true;

    api
      .exchangeAuthCode(code)
      .then(async ({ access_token, refresh_token }) => {
        // Set cookie so Next.js middleware can verify auth on server-side
        document.cookie = `tene_access_token=${access_token}; path=/; max-age=${15 * 60}; SameSite=Lax`;
        // Store in Zustand (persisted to localStorage)
        login(access_token, refresh_token);

        // Fetch user profile to check plan
        try {
          const me = await api.getMe();
          if (me.plan !== "pro") {
            router.replace("/upgrade");
            return;
          }
        } catch {
          // If /me fails, middleware will handle plan check
        }

        // Pro user — redirect to dashboard
        if (intent === "upgrade") {
          router.replace("/");
        } else {
          router.replace("/");
        }
      })
      .catch(() => {
        router.replace("/login?error=exchange_failed");
      });
  }, [searchParams, login, router]);

  return (
    <div className="text-center">
      <h1 className="font-mono font-bold text-2xl text-accent mb-2">tene</h1>
      <p className="text-muted text-sm">Signing you in...</p>
    </div>
  );
}

export default function AuthCallbackPage() {
  return (
    <div className="min-h-screen flex items-center justify-center">
      <Suspense fallback={<p className="text-muted text-sm">Loading...</p>}>
        <CallbackHandler />
      </Suspense>
    </div>
  );
}
