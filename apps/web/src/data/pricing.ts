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
    description: "Everything you need for local secret management.",
    features: [
      "Unlimited secrets & environments",
      "XChaCha20-Poly1305 encryption",
      "AI runtime injection (5 agents)",
      "OS keychain integration",
      "12-word recovery key",
      "Import from .env files",
      "Audit logging",
    ],
    cta: { label: "Install CLI — Free", action: "install" },
    highlighted: true,
    comingSoon: false,
  },
  {
    name: "Pro",
    price: "TBD",
    period: "",
    description:
      "Team sync, advanced audit, and compliance features. Coming soon.",
    features: [
      "Everything in Free",
      "Encrypted team sync (no server)",
      "Advanced audit & compliance export",
      "Team secret sharing",
      "Role-based access control",
      "Priority support",
    ],
    cta: { label: "Join Waitlist", action: "waitlist" },
    highlighted: false,
    comingSoon: true,
  },
];
