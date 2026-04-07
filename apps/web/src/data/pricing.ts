export type PricingTier = {
  name: string;
  price: string;
  period: string;
  description: string;
  features: string[];
  cta: { label: string; action: "install" | "waitlist" };
  highlighted: boolean;
  comingSoon: boolean;
};

export const pricingTiers: PricingTier[] = [
  {
    name: "Free",
    price: "$0",
    period: "forever",
    description: "Local encrypted secrets for individual projects.",
    features: [
      "Unlimited secrets",
      "XChaCha20-Poly1305 encryption",
      "AI runtime injection (5 agents)",
      "OS keychain integration",
      "12-word recovery key",
      "Multi-environment support",
    ],
    cta: { label: "Install now", action: "install" },
    highlighted: false,
    comingSoon: false,
  },
  {
    name: "Solo",
    price: "$5",
    period: "per month",
    description: "Sync your vault across machines. No repeated setup.",
    features: [
      "Everything in Free",
      "Cross-machine vault sync",
      "Encrypted cloud backup",
      "Device management",
      "Personal audit log",
    ],
    cta: { label: "Join waitlist", action: "waitlist" },
    highlighted: true,
    comingSoon: true,
  },
  {
    name: "Team",
    price: "$10",
    period: "per user / month",
    description: "Share secrets securely across your team.",
    features: [
      "Everything in Solo",
      "Team secret sharing",
      "Role-based access control",
      "Environment-level permissions",
      "Team audit log & dashboard",
    ],
    cta: { label: "Join waitlist", action: "waitlist" },
    highlighted: false,
    comingSoon: true,
  },
];
