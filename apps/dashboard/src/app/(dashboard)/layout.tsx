import { Sidebar, MobileNav } from "@/components/sidebar";

export default function DashboardLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex min-h-screen">
      <Sidebar />
      <main className="flex-1 min-w-0">
        <div className="max-w-6xl mx-auto p-4 sm:p-6 lg:p-8 pb-20 lg:pb-8">
          {children}
        </div>
      </main>
      <MobileNav />
    </div>
  );
}
