import Link from "next/link";
import { comparisonRows, comparisonPricing } from "@/data/comparison";

// Design Ref: §4.6 — Comparison table with data import, "Secrets hidden from AI" first
function Check() {
  return (
    <svg className="h-5 w-5 text-accent" fill="none" stroke="currentColor" viewBox="0 0 24 24" strokeWidth="2.5">
      <path d="M5 13l4 4L19 7" />
    </svg>
  );
}

function Cross() {
  return (
    <svg className="h-4 w-4 text-muted/40" fill="none" stroke="currentColor" viewBox="0 0 24 24" strokeWidth="2">
      <path d="M18 6L6 18M6 6l12 12" />
    </svg>
  );
}

export function Comparison() {
  return (
    <section className="px-4 py-24 sm:px-6">
      <div className="mx-auto max-w-4xl">
        <h2 className="text-center text-3xl font-bold sm:text-4xl">
          How Tene compares
        </h2>
        <p className="mx-auto mt-4 max-w-xl text-center text-muted">
          The only tool that hides secrets from AI agents while keeping everything local and free.
        </p>

        <div className="-mx-4 mt-12 overflow-x-auto px-4 sm:mx-0 sm:px-0">
          <table className="w-full min-w-[540px] text-sm">
            <thead>
              <tr className="border-b border-border">
                <th className="pb-4 pr-4 text-left font-normal text-muted" />
                <th className="pb-4 px-3 text-center font-bold text-accent sm:px-4">Tene</th>
                <th className="pb-4 px-3 text-center font-normal text-muted sm:px-4">.env</th>
                <th className="pb-4 px-3 text-center font-normal text-muted sm:px-4">Doppler</th>
                <th className="pb-4 px-3 text-center font-normal text-muted sm:px-4">Vault</th>
                <th className="pb-4 px-3 text-center font-normal text-muted sm:px-4">Infisical</th>
              </tr>
            </thead>
            <tbody>
              {comparisonRows.map((row) => (
                <tr key={row.feature} className="border-b border-border/50">
                  <td className="py-3 pr-4 whitespace-nowrap text-foreground">{row.feature}</td>
                  <td className="py-3 px-3 text-center sm:px-4">
                    <div className="flex justify-center">{row.tene ? <Check /> : <Cross />}</div>
                  </td>
                  <td className="py-3 px-3 text-center sm:px-4">
                    <div className="flex justify-center">{row.env ? <Check /> : <Cross />}</div>
                  </td>
                  <td className="py-3 px-3 text-center sm:px-4">
                    <div className="flex justify-center">{row.doppler ? <Check /> : <Cross />}</div>
                  </td>
                  <td className="py-3 px-3 text-center sm:px-4">
                    <div className="flex justify-center">{row.vault ? <Check /> : <Cross />}</div>
                  </td>
                  <td className="py-3 px-3 text-center sm:px-4">
                    <div className="flex justify-center">{row.infisical ? <Check /> : <Cross />}</div>
                  </td>
                </tr>
              ))}
              <tr>
                <td className="pt-4 pr-4 text-muted">Price</td>
                <td className="pt-4 px-3 text-center font-bold text-accent sm:px-4">{comparisonPricing.tene}</td>
                <td className="pt-4 px-3 text-center text-muted sm:px-4">{comparisonPricing.env}</td>
                <td className="pt-4 px-3 text-center text-muted sm:px-4">{comparisonPricing.doppler}</td>
                <td className="pt-4 px-3 text-center text-muted sm:px-4">{comparisonPricing.vault}</td>
                <td className="pt-4 px-3 text-center text-muted sm:px-4">{comparisonPricing.infisical}</td>
              </tr>
            </tbody>
          </table>
        </div>

        <div className="mt-10 flex justify-center">
          <Link
            href="/vs"
            className="group inline-flex items-center gap-2 text-sm font-medium text-accent transition-colors hover:text-accent/80 sm:text-base"
          >
            See in-depth comparisons against every secret manager
            <span aria-hidden className="transition-transform group-hover:translate-x-1">→</span>
          </Link>
        </div>
      </div>
    </section>
  );
}
