"use client";

import { useSearchParams } from "next/navigation";
import { Suspense } from "react";

function LoginContent() {
  const apiUrl = process.env.NEXT_PUBLIC_API_URL || "https://api.tene.sh";
  const searchParams = useSearchParams();
  const intent = searchParams.get("intent");
  const error = searchParams.get("error");

  const oauthUrl = intent
    ? `${apiUrl}/api/v1/auth/github/authorize?redirect=dashboard&intent=${intent}`
    : `${apiUrl}/api/v1/auth/github/authorize?redirect=dashboard`;

  return (
    <div className="min-h-screen flex items-center justify-center p-4">
      <div className="w-full max-w-sm space-y-8">
        <div className="text-center">
          {/* eslint-disable-next-line @next/next/no-img-element */}
          <img src="/logo.svg" alt="Tene" width={48} height={48} className="rounded-xl mx-auto mb-3" />
          <h1 className="font-mono font-bold text-3xl mb-2">tene</h1>
          <p className="text-muted text-sm">Sign in to your Tene Cloud dashboard</p>
        </div>

        {error && (
          <div className="rounded-lg border border-red-500/30 bg-red-500/10 px-4 py-3 text-sm text-red-400 text-center">
            {error === "exchange_failed" ? "Sign in failed. Please try again." : "An error occurred."}
          </div>
        )}

        <div className="space-y-3">
          <a
            href={oauthUrl}
            className="flex items-center justify-center gap-3 w-full py-3 rounded-lg border border-border bg-surface text-sm font-medium transition-all hover:border-accent/30 hover:bg-surface-2 active:scale-[0.98]"
          >
            <svg viewBox="0 0 16 16" className="w-5 h-5 fill-current" aria-hidden>
              <path d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27.68 0 1.36.09 2 .27 1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.013 8.013 0 0016 8c0-4.42-3.58-8-8-8z" />
            </svg>
            Continue with GitHub
          </a>

        </div>

        <div className="text-center text-xs text-muted">
          <p>Your secrets are encrypted locally and never leave your device.</p>
          <p className="mt-1">The dashboard only shows metadata (key names, not values).</p>
        </div>
      </div>
    </div>
  );
}

export default function LoginPage() {
  return (
    <Suspense fallback={<div className="min-h-screen flex items-center justify-center"><p className="text-muted text-sm">Loading...</p></div>}>
      <LoginContent />
    </Suspense>
  );
}
