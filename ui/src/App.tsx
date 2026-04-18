import React from "react";
import {
  BrowserRouter as Router,
  Routes,
  Route,
  Link,
  useLocation,
} from "react-router-dom";
import {
  FlaskConical,
  History,
  Search,
  Activity,
  CheckCircle2,
  AlertCircle,
} from "lucide-react";
import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";

function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

// --- Components ---

const Sidebar = () => {
  const location = useLocation();

  const navItems = [
    { name: "Explorer", path: "/", icon: Search },
    { name: "Lab", path: "/lab", icon: FlaskConical },
    { name: "Regression", path: "/regression", icon: History },
  ];

  return (
    <aside className="w-64 border-r border-slate-800 flex flex-col h-screen sticky top-0 bg-slate-900 text-slate-300">
      <div className="p-6 border-b border-slate-800">
        <h1 className="text-xl font-bold text-white flex items-center gap-2">
          <Activity className="text-blue-500" />
          Litmus
        </h1>
      </div>

      <nav className="flex-1 p-4 space-y-2">
        {navItems.map((item) => (
          <Link
            key={item.path}
            to={item.path}
            className={cn(
              "flex items-center gap-3 px-3 py-2 rounded-lg transition-colors",
              location.pathname === item.path
                ? "bg-blue-600/10 text-blue-400 border border-blue-600/20"
                : "hover:bg-slate-800 hover:text-white",
            )}
          >
            <item.icon size={20} />
            <span className="font-medium">{item.name}</span>
          </Link>
        ))}
      </nav>

      <div className="p-4 border-t border-slate-800 text-xs text-slate-500">
        v0.1.0-alpha
      </div>
    </aside>
  );
};

const Header = ({ title }: { title: string }) => (
  <header className="h-16 border-b border-slate-800 flex items-center px-8 bg-slate-900/50 backdrop-blur-sm sticky top-0 z-10">
    <h2 className="text-lg font-semibold text-white uppercase tracking-wider">
      {title}
    </h2>
  </header>
);

const StatsSidebar = ({ children }: { children?: React.ReactNode }) => (
  <aside className="w-80 border-l border-slate-800 p-6 flex flex-col h-screen sticky top-0 bg-slate-900 overflow-y-auto">
    <h3 className="text-sm font-bold text-slate-500 uppercase mb-6 tracking-widest">
      Statistics
    </h3>
    <div className="space-y-4">
      {children || (
        <div className="p-4 rounded-xl bg-slate-800/50 border border-slate-700">
          <p className="text-slate-400 text-sm">Select a page to see stats</p>
        </div>
      )}
    </div>
  </aside>
);

// --- Pages ---

const ExplorerPage = () => (
  <div className="flex-1 flex flex-col min-h-0">
    <Header title="Route Explorer" />
    <main className="flex-1 p-6 flex flex-col gap-6 overflow-hidden">
      {/* Route Path at the top - Scrollable */}
      <div className="flex-1 p-6 rounded-2xl bg-slate-800/30 border border-slate-800 overflow-y-auto">
        <div className="flex items-center justify-between mb-6">
          <h3 className="text-white font-medium flex items-center gap-2">
            <Activity size={18} className="text-blue-500" />
            Evaluation Path
          </h3>
          <span className="text-xs text-slate-500 font-mono uppercase tracking-widest">
            Live Preview
          </span>
        </div>

        <div className="space-y-6 relative">
          {[1, 2, 3].map((i) => (
            <div key={i} className="flex items-start gap-4 group relative">
              {/* Connector Line */}
              {i < 3 && (
                <div className="absolute left-4 top-8 bottom-0 w-0.5 bg-slate-800 group-hover:bg-blue-500/30 transition-colors -mb-6 z-0" />
              )}

              <div
                className={cn(
                  "w-8 h-8 shrink-0 rounded-full flex items-center justify-center text-sm font-bold z-10 transition-transform group-hover:scale-110",
                  i === 1
                    ? "bg-blue-600 text-white ring-4 ring-blue-600/20"
                    : "bg-slate-800 text-slate-500 border border-slate-700",
                )}
              >
                {i}
              </div>

              <div className="flex-1 p-4 rounded-xl bg-slate-900/50 border border-slate-800 group-hover:border-slate-600 transition-all shadow-sm">
                <div className="flex items-center justify-between mb-2">
                  <span className="text-slate-200 font-semibold tracking-tight">
                    Route Node {i}
                  </span>
                  <div className="flex gap-2">
                    <span className="px-2 py-0.5 rounded bg-blue-500/10 text-blue-400 text-[10px] font-bold uppercase tracking-tighter border border-blue-500/20">
                      continue
                    </span>
                    <span className="px-2 py-0.5 rounded bg-slate-800 text-slate-400 text-[10px] font-bold uppercase tracking-tighter border border-slate-700">
                      matching
                    </span>
                  </div>
                </div>
                <div className="text-sm text-slate-500 font-mono space-y-1">
                  <div className="flex gap-2">
                    <span className="text-blue-400/60">match:</span>{" "}
                    <span>{'{severity="critical"}'}</span>
                  </div>
                  <div className="flex gap-2">
                    <span className="text-blue-400/60">receiver:</span>{" "}
                    <span>pagerduty-critical</span>
                  </div>
                </div>
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* Input at the bottom - Fixed/Compact */}
      <div className="p-6 rounded-2xl bg-slate-900 border border-slate-800 shadow-2xl relative group ring-1 ring-slate-800 hover:ring-blue-500/30 transition-all">
        <div className="flex items-center justify-between mb-4">
          <div className="flex items-center gap-3">
            <Search
              size={18}
              className="text-slate-500 group-focus-within:text-blue-500 transition-colors"
            />
            <h3 className="text-white font-medium">Alert Labels</h3>
          </div>
          <div className="flex gap-2">
            <button className="px-3 py-1.5 text-xs font-semibold bg-slate-800 text-slate-300 rounded-md hover:bg-slate-700 transition-colors border border-slate-700">
              Clear
            </button>
            <button className="px-4 py-1.5 text-xs font-semibold bg-blue-600 text-white rounded-md hover:bg-blue-500 transition-colors shadow-lg shadow-blue-900/20">
              Run Query
            </button>
          </div>
        </div>
        <textarea
          className="w-full h-24 bg-slate-950 border border-slate-800 rounded-lg p-4 font-mono text-sm text-blue-400 focus:ring-1 focus:ring-blue-500/50 focus:outline-none resize-none shadow-inner"
          placeholder="severity: critical&#10;team: database"
          defaultValue="severity: critical&#10;team: database"
        />
        <div className="absolute bottom-8 right-8 pointer-events-none opacity-20">
          <Activity size={40} className="text-blue-500" />
        </div>
      </div>
    </main>
  </div>
);

const LabPage = () => (
  <div className="flex-1 flex flex-col min-h-0">
    <Header title="Test Lab" />
    <main className="flex-1 p-8 overflow-y-auto">
      <div className="space-y-4">
        {[1, 2, 3].map((i) => (
          <div
            key={i}
            className="flex items-center gap-4 p-5 rounded-xl bg-slate-800/30 border border-slate-800 hover:border-blue-500/50 hover:bg-slate-800/50 transition-all cursor-pointer group shadow-sm"
          >
            <div className="w-10 h-10 rounded-full bg-emerald-500/10 flex items-center justify-center border border-emerald-500/20">
              <CheckCircle2 className="text-emerald-500" size={20} />
            </div>
            <div className="flex-1">
              <h4 className="text-white font-medium text-lg tracking-tight">
                Critical alerts reach on-call
              </h4>
              <p className="text-slate-500 text-sm font-mono">
                tests/critical-alert.yml
              </p>
            </div>
            <div className="flex items-center gap-4">
              <span className="text-xs font-mono text-slate-500 uppercase tracking-widest bg-slate-900 px-3 py-1 rounded-full border border-slate-800">
                42ms
              </span>
              <button className="px-6 py-2 bg-blue-600 hover:bg-blue-500 text-white text-sm font-bold rounded-lg opacity-0 group-hover:opacity-100 transition-all translate-x-2 group-hover:translate-x-0">
                Run
              </button>
            </div>
          </div>
        ))}
      </div>
    </main>
  </div>
);

const RegressionPage = () => (
  <div className="flex-1 flex flex-col min-h-0">
    <Header title="Regression Diff" />
    <main className="flex-1 p-8 overflow-y-auto">
      <div className="bg-slate-800/30 border border-slate-800 rounded-2xl overflow-hidden shadow-xl">
        <table className="w-full text-left border-collapse">
          <thead>
            <tr className="bg-slate-900/80 backdrop-blur-md">
              <th className="p-5 text-xs font-bold text-slate-500 uppercase tracking-widest border-b border-slate-800">
                Alert Labels
              </th>
              <th className="p-5 text-xs font-bold text-slate-500 uppercase tracking-widest border-b border-slate-800">
                Old Receiver
              </th>
              <th className="p-5 text-xs font-bold text-slate-500 uppercase tracking-widest border-b border-slate-800">
                New Receiver
              </th>
              <th className="p-5 text-xs font-bold text-slate-500 uppercase tracking-widest border-b border-slate-800">
                Status
              </th>
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-800/50">
            {[1, 2, 3].map((i) => (
              <tr key={i} className="hover:bg-slate-800/20 transition-colors group">
                <td className="p-5 text-slate-300 font-mono text-sm group-hover:text-blue-400 transition-colors tracking-tight">
                  {'{severity="critical", service="db"}'}
                </td>
                <td className="p-5 text-slate-500 line-through decoration-slate-700/50 font-mono text-sm italic">
                  default-receiver
                </td>
                <td className="p-5">
                  <span className="px-3 py-1 rounded-md text-emerald-400 font-bold bg-emerald-500/10 border border-emerald-500/20 text-sm font-mono">
                    pagerduty-oncall
                  </span>
                </td>
                <td className="p-5">
                  <span className="flex items-center gap-2 text-xs font-bold text-amber-500/80 uppercase tracking-tighter">
                    <div className="w-1.5 h-1.5 rounded-full bg-amber-500 shadow-[0_0_8px_rgba(245,158,11,0.5)]" />
                    Drifted
                  </span>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </main>
  </div>
);

// --- Layout & Root ---

const AppLayout = ({
  children,
  stats,
}: {
  children: React.ReactNode;
  stats?: React.ReactNode;
}) => (
  <div className="flex min-h-screen bg-slate-950 text-slate-300 w-full overflow-x-hidden">
    <Sidebar />
    <div className="flex-1 flex min-w-0">
      {children}
      <StatsSidebar>{stats}</StatsSidebar>
    </div>
  </div>
);

function App() {
  return (
    <Router>
      <Routes>
        <Route
          path="/"
          element={
            <AppLayout
              stats={
                <div className="space-y-4">
                  <div className="p-4 rounded-xl bg-slate-800/50 border border-slate-700">
                    <p className="text-slate-500 text-xs uppercase font-bold mb-1">
                      Matched Receiver
                    </p>
                    <p className="text-white font-mono">database-oncall</p>
                  </div>
                  <div className="p-4 rounded-xl bg-slate-800/50 border border-slate-700">
                    <p className="text-slate-500 text-xs uppercase font-bold mb-1">
                      Match Confidence
                    </p>
                    <p className="text-white font-bold text-lg">100%</p>
                  </div>
                </div>
              }
            >
              <ExplorerPage />
            </AppLayout>
          }
        />
        <Route
          path="/lab"
          element={
            <AppLayout
              stats={
                <div className="space-y-4">
                  <div className="p-4 rounded-xl bg-slate-800/50 border border-slate-700 flex items-center justify-between">
                    <div>
                      <p className="text-slate-500 text-xs uppercase font-bold mb-1">
                        Passed
                      </p>
                      <p className="text-emerald-400 text-2xl font-bold">12</p>
                    </div>
                    <CheckCircle2 className="text-emerald-500/20" size={40} />
                  </div>
                  <div className="p-4 rounded-xl bg-slate-800/50 border border-slate-700 flex items-center justify-between">
                    <div>
                      <p className="text-slate-500 text-xs uppercase font-bold mb-1">
                        Failed
                      </p>
                      <p className="text-rose-400 text-2xl font-bold">0</p>
                    </div>
                    <AlertCircle className="text-rose-500/20" size={40} />
                  </div>
                </div>
              }
            >
              <LabPage />
            </AppLayout>
          }
        />
        <Route
          path="/regression"
          element={
            <AppLayout
              stats={
                <div className="space-y-4">
                  <div className="p-4 rounded-xl bg-slate-800/50 border border-slate-700">
                    <p className="text-slate-500 text-xs uppercase font-bold mb-1">
                      Total Regressions
                    </p>
                    <p className="text-white text-2xl font-bold">1</p>
                  </div>
                  <div className="p-4 rounded-xl bg-slate-800/50 border border-slate-700">
                    <p className="text-slate-500 text-xs uppercase font-bold mb-1">
                      Status
                    </p>
                    <div className="flex items-center gap-2 mt-1">
                      <div className="w-2 h-2 rounded-full bg-amber-500 animate-pulse" />
                      <span className="text-amber-500 font-medium">
                        Drift Detected
                      </span>
                    </div>
                  </div>
                </div>
              }
            >
              <RegressionPage />
            </AppLayout>
          }
        />
      </Routes>
    </Router>
  );
}

export default App;
