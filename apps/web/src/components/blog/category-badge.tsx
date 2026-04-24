// Small category badge for post cards and post headers. Color scheme
// intentionally uses Tailwind's named colors with low-opacity surfaces so
// the badge reads as a soft status pill, not a button.
import Link from "next/link";
import { getCategoryLabel, type CategoryKey } from "@/lib/tags";

type Props = {
  category: CategoryKey;
  size?: "sm" | "md";
  asLink?: boolean; // when true, renders as Link to /blog/category/{slug}
};

const COLOR_BY_CATEGORY: Record<CategoryKey, string> = {
  tools: "bg-blue-500/10 text-blue-300 border-blue-500/30",
  engineering: "bg-purple-500/10 text-purple-300 border-purple-500/30",
  "vibe-coding": "bg-emerald-500/10 text-emerald-300 border-emerald-500/30",
  philosophy: "bg-amber-500/10 text-amber-300 border-amber-500/30",
};

export function CategoryBadge({
  category,
  size = "sm",
  asLink = false,
}: Props) {
  const label = getCategoryLabel(category);
  const colorClass = COLOR_BY_CATEGORY[category];
  const sizeClass =
    size === "sm" ? "px-2 py-0.5 text-xs" : "px-3 py-1 text-sm";
  const base = `inline-flex items-center rounded border font-medium ${sizeClass} ${colorClass}`;

  if (asLink) {
    return (
      <Link
        href={`/blog/category/${category}`}
        className={`${base} hover:opacity-80 transition-opacity`}
      >
        {label}
      </Link>
    );
  }

  return <span className={base}>{label}</span>;
}
