export type PricingTier = {
  name: string;
  price: string;
  period: string;
  description: string;
  features: string[];
  cta: { label: string; action: "install" | "waitlist" | "signup" };
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
    cta: { label: "Install CLI — Free", action: "install" },
    highlighted: false,
    comingSoon: false,
  },
  {
    name: "Pro",
    price: "$5",
    period: "per month",
    description:
      "Sync your vault, share with your team. Everything included.",
    features: [
      "Everything in Free",
      "Cross-machine vault sync",
      "Encrypted cloud backup",
      "Device management",
      "Team secret sharing",
      "Role-based access control",
      "Environment-level permissions",
      "Audit log & dashboard",
    ],
    cta: { label: "Open Dashboard", action: "signup" },
    highlighted: true,
    comingSoon: false,
  },
];
