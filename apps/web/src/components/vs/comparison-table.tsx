import type { ComparisonRow } from "@/data/comparisons/types";

type Props = {
  competitorName: string;
  rows: ComparisonRow[];
};

export function ComparisonTable({ competitorName, rows }: Props) {
  return (
    <section className="px-4 py-12 sm:px-6">
      <div className="mx-auto max-w-4xl">
        <h2 className="text-2xl font-bold sm:text-3xl">Side-by-side</h2>
        <p className="mt-3 text-muted">
          Feature-by-feature comparison. Every row is sourced from the official
          docs of each product — if you find something stale, open an issue.
        </p>

        <div className="-mx-4 mt-8 overflow-x-auto px-4 sm:mx-0 sm:px-0">
          <table className="w-full min-w-[640px] text-sm">
            <thead>
              <tr className="border-b border-border">
                <th className="py-3 pr-4 text-left font-normal text-muted">
                  Dimension
                </th>
                <th className="py-3 px-4 text-left font-bold text-accent">
                  tene
                </th>
                <th className="py-3 px-4 text-left font-normal text-muted">
                  {competitorName}
                </th>
              </tr>
            </thead>
            <tbody>
              {rows.map((row) => (
                <tr
                  key={row.dimension}
                  className="border-b border-border/50 align-top"
                >
                  <td className="py-4 pr-4 font-medium whitespace-nowrap">
                    {row.dimension}
                  </td>
                  <td className="py-4 px-4 text-foreground">{row.tene}</td>
                  <td className="py-4 px-4 text-muted">{row.competitor}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </section>
  );
}
