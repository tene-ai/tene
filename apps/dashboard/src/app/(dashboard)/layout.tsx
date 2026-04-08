import { Sidebar, MobileNav, MobileHeader } from "@/components/sidebar";
import { CommandPalette } from "@/components/command-palette";
import { InteractiveGrid } from "@/components/interactive-grid";

export default function DashboardLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex min-h-screen">
      <InteractiveGrid />
      <Sidebar />
      <main className="flex-1 min-w-0 relative z-10">
        <MobileHeader />
        <div className="max-w-6xl mx-auto p-4 sm:p-6 lg:p-8 pb-20 lg:pb-8">
          {children}
        </div>
      </main>
      <MobileNav />
      <CommandPalette />
    </div>
  );
}
