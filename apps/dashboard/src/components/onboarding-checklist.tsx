"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { api, type OnboardingStatus } from "@/lib/api";
import { Check, X } from "lucide-react";

export function OnboardingChecklist({ status }: { status: OnboardingStatus }) {
  const queryClient = useQueryClient();

  const dismissMutation = useMutation({
    mutationFn: () => api.dismissOnboarding(),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["onboarding"] }),
  });

  const steps = [
    { done: true, label: "GitHub account connected", hint: null },
    { done: status.cli_installed, label: "Install CLI", hint: "curl -fsSL tene.sh/install.sh | sh" },
    { done: status.first_push, label: "Push your first vault", hint: "cd your-project && tene init && tene push" },
    { done: status.second_device, label: "Connect a second device", hint: "tene login && tene pull", optional: true },
  ];

  const completedCount = steps.filter((s) => s.done).length;

  return (
    <div className="rounded-xl border border-accent/20 bg-accent/5 p-5">
      <div className="flex items-start justify-between mb-3">
        <div>
          <h2 className="font-semibold text-sm">Get started with Tene Cloud</h2>
          <p className="text-xs text-muted mt-0.5">{completedCount}/{steps.length} completed</p>
        </div>
        <button
          onClick={() => dismissMutation.mutate()}
          className="text-muted hover:text-foreground transition-colors"
          aria-label="Dismiss onboarding"
        >
          <X size={16} />
        </button>
      </div>
      <div className="space-y-3">
        {steps.map((step, i) => (
          <div key={i} className="flex items-start gap-3">
            <div className={`mt-0.5 flex-shrink-0 w-5 h-5 rounded-full flex items-center justify-center text-xs ${
              step.done
                ? "bg-accent text-background"
                : "border border-border text-muted"
            }`}>
              {step.done ? <Check size={12} /> : i + 1}
            </div>
            <div>
              <span className={`text-sm ${step.done ? "line-through text-muted" : ""}`}>
                {step.label}
              </span>
              {step.optional && <span className="text-muted text-xs ml-1">(optional)</span>}
              {!step.done && step.hint && (
                <code className="block mt-1 text-xs font-mono text-accent bg-surface px-2 py-1 rounded">
                  {step.hint}
                </code>
              )}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
