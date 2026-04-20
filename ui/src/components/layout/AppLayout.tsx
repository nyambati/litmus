import React from "react";
import { Sidebar } from "./Sidebar";
import { StatsSidebar } from "./StatsSidebar";

export const AppLayout = ({
  children,
  stats,
}: {
  children: React.ReactNode;
  stats?: React.ReactNode;
}) => (
  <div className="flex h-screen bg-[#181b1f] text-[#d9d9d9] w-full overflow-hidden">
    <Sidebar />
    <main className="flex-1 flex flex-col min-w-0 overflow-hidden">
      {children}
    </main>
    <StatsSidebar>{stats}</StatsSidebar>
  </div>
);
