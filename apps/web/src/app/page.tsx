import { Hero } from "@/components/hero";
import { Features } from "@/components/features";
import { Comparison } from "@/components/comparison";
import { HowItWorks } from "@/components/how-it-works";
import { Security } from "@/components/security";
import { CTA } from "@/components/cta";
import { Pricing } from "@/components/pricing";
import { Footer } from "@/components/footer";
import { Nav } from "@/components/nav";
import { FAQ } from "@/components/faq";
import { InteractiveGrid } from "@/components/interactive-grid";

export default function Home() {
  return (
    <>
      {/* Canvas-based interactive dot grid (desktop only) */}
      <InteractiveGrid />

      {/* CSS fallback dot grid for mobile */}
      <div className="dot-grid-fixed sm:hidden" />

      <Nav />
      <main className="relative z-10">
        <Hero />
        <Features />
        <HowItWorks />
        <Security />
        <Comparison />
        {/* <Pricing /> */}
        <FAQ />
        <CTA />
      </main>
      <Footer />
    </>
  );
}
