import { pricingTiers } from "@/data/pricing";
import { CopyCommand } from "./copy-command";
import { WaitlistForm } from "./waitlist-form";
import { GlowCard } from "./glow-card";

export function Pricing() {
  return (
    <section id="pricing" className="px-4 py-24 sm:px-6">
      <div className="mx-auto max-w-5xl">
        <h2 className="text-center text-3xl font-bold sm:text-4xl">
          Free locally.{" "}
          <span className="text-accent">Cloud when you need it.</span>
        </h2>
        <p className="mx-auto mt-4 max-w-xl text-center text-muted">
          Local CLI is free forever. Pay only for cloud sync and team
          features.
        </p>

        <div className="mt-16 grid gap-6 lg:grid-cols-3">
          {pricingTiers.map((tier) => (
            <GlowCard
              key={tier.name}
              className={`rounded-xl border p-8 ${
                tier.highlighted
                  ? "border-accent/40 bg-accent/5"
                  : "border-border bg-surface"
              }`}
            >
              <div className="relative z-10 flex h-full flex-col">
                <div className="flex items-center justify-between">
                  <h3 className="text-lg font-semibold">{tier.name}</h3>
                  {tier.comingSoon && (
                    <span className="rounded-full border border-yellow-500/30 bg-yellow-500/10 px-2.5 py-0.5 text-xs text-yellow-400">
                      Coming Soon
                    </span>
                  )}
                </div>

                <div className="mt-4 flex items-baseline gap-1">
                  <span className="text-4xl font-bold text-accent">
                    {tier.price}
                  </span>
                  <span className="text-sm text-muted">/ {tier.period}</span>
                </div>

                <p className="mt-2 text-sm text-muted">{tier.description}</p>

                <ul className="mt-6 flex-1 space-y-3">
                  {tier.features.map((feature) => (
                    <li key={feature} className="flex items-start gap-2 text-sm">
                      <svg
                        className="mt-0.5 h-4 w-4 shrink-0 text-accent"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                        strokeWidth="2"
                      >
                        <path d="M5 13l4 4L19 7" />
                      </svg>
                      {feature}
                    </li>
                  ))}
                </ul>

                <div className="mt-8">
                  {tier.cta.action === "install" && (
                    <CopyCommand
                      command="curl -sSfL https://tene.sh/install.sh | sh"
                      className="w-full justify-center text-xs"
                    />
                  )}
                </div>
              </div>
            </GlowCard>
          ))}
        </div>

        {/* Shared waitlist form */}
        <div className="mx-auto mt-12 max-w-lg text-center">
          <p className="mb-4 text-sm text-muted">
            Interested in Solo or Team? Join the waitlist for early access.
          </p>
          <WaitlistForm />
        </div>
      </div>
    </section>
  );
}
