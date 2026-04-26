import { Link, useLocation } from "react-router-dom";
import { Search, FlaskConical, History, Activity, Layers } from "lucide-react";
import { cn } from "../../utils/ui";

export const Sidebar = () => {
  const location = useLocation();

  const navItems = [
    { name: "Explorer", path: "/", icon: Search, description: "Route evaluator" },
    { name: "Lab", path: "/lab", icon: FlaskConical, description: "Test runner" },
    { name: "Regression", path: "/regression", icon: History, description: "Drift detection" },
    { name: "Fragments", path: "/fragments", icon: Layers, description: "Config packages" },
  ];

  return (
    <aside
      className="w-[220px] flex flex-col shrink-0 border-r border-[#1e2228]"
      style={{ background: "#111217" }}
    >
      {/* Logo */}
      <div className="h-12 flex items-center px-4 border-b border-[#1e2228] shrink-0">
        <div className="flex items-center gap-2.5">
          <div className="w-6 h-6 rounded flex items-center justify-center bg-[#f46800]/15">
            <Activity size={14} className="text-[#f46800]" />
          </div>
          <span className="text-[#d9d9d9] font-semibold text-[15px] tracking-tight">
            Litmus
          </span>
        </div>
      </div>

      {/* Nav */}
      <nav className="flex-1 py-2">
        {navItems.map((item) => {
          const active = location.pathname === item.path;
          return (
            <Link
              key={item.path}
              to={item.path}
              className={cn(
                "relative flex items-center gap-3 px-4 py-2.5 text-sm transition-colors group",
                active
                  ? "text-[#d9d9d9] bg-[#f46800]/8"
                  : "text-[#8e9193] hover:text-[#d9d9d9] hover:bg-[#ffffff06]",
              )}
            >
              {/* Active indicator */}
              {active && (
                <span className="absolute left-0 top-0 bottom-0 w-[3px] bg-[#f46800] rounded-r" />
              )}
              <item.icon size={16} className={active ? "text-[#f46800]" : ""} />
              <span className="font-medium">{item.name}</span>
            </Link>
          );
        })}
      </nav>

      {/* Footer */}
      <div className="px-4 py-3 border-t border-[#1e2228]">
        <span className="text-[11px] text-[#8e9193]/60 font-mono">v0.1.0-alpha</span>
      </div>
    </aside>
  );
};
