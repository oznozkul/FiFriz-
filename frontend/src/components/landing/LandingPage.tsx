import { AiShowcaseSection } from "./AiShowcaseSection";
import { FeatureSection } from "./FeatureSection";
import { HeroSection } from "./HeroSection";
import { HowItWorksSection } from "./HowItWorksSection";
import { LandingFooter } from "./LandingFooter";
import { LandingNavbar } from "./LandingNavbar";
import { StatsSection } from "./StatsSection";

export function LandingPage() {
  return (
    <div className="min-h-screen scroll-smooth bg-white text-slate-700">
      <LandingNavbar />
      <main>
        <HeroSection />
        <FeatureSection />
        <AiShowcaseSection />
        <HowItWorksSection />
        <StatsSection />
      </main>
      <LandingFooter />
    </div>
  );
}
