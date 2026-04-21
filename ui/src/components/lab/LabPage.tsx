import { useState, useEffect, useCallback } from "react";
import { FlaskConical, ChevronDown } from "lucide-react";
import { API, loadCache, saveCache, cn } from "../../utils/persistence";
import { GfSpinner } from "../ui/Spinner";
import { LastUpdated } from "../ui/LastUpdated";
import { Header } from "../layout/Header";
import { EmptyState } from "../ui/EmptyState";
import { FilterTabs } from "../ui/Tabs";
import { TestCaseCard, type Test, type TestResult } from "./TestCase";

type FilterType = "all" | "unit" | "regression";

interface TestWithType extends Test {
  type: "unit" | "regression";
}

interface TestRunResult extends TestResult {
  name: string;
}

export const LabPage = ({
  onTestsRun,
}: {
  onTestsRun: (passed: number, failed: number) => void;
}) => {
  const [tests, setTests] = useState<TestWithType[]>([]);
  const [results, setResults] = useState<Record<string, TestResult>>(() => {
    return loadCache<Record<string, TestResult>>("litmus:lab:results")?.data ?? {};
  });
  const [lastRunTs, setLastRunTs] = useState<number | null>(() => {
    return loadCache<Record<string, TestResult>>("litmus:lab:results")?.ts ?? null;
  });
  const [loading, setLoading] = useState(false);
  const [running, setRunning] = useState(false);
  const [snapshotting, setSnapshotting] = useState(false);
  const [runningTest, setRunningTest] = useState<string | null>(null);
  const [filter, setFilter] = useState<FilterType>("all");
  const [activeAction, setActiveAction] = useState<"run" | "snapshot">("run");
  const [showDropdown, setShowDropdown] = useState(false);

  const actions = [
    {
      id: "run" as const,
      label: "Run All",
      icon: FlaskConical,
      desc: "Run all tests",
    },
    {
      id: "snapshot" as const,
      label: "Snapshot",
      icon: FlaskConical,
      desc: "Generate regression tests",
    },
  ];

  const currentAction =
    actions.find((a) => a.id === activeAction) || actions[0];

  const handleAction = () => {
    if (activeAction === "run") {
      runAllTests();
    } else {
      runSnapshot();
    }
  };

  const fetchTests = useCallback(async () => {
    setLoading(true);
    try {
      const [unitSettled, regressionSettled] = await Promise.allSettled([
        fetch(`${API}/api/v1/tests`).then((r) => r.json()),
        fetch(`${API}/api/v1/regressions`).then((r) => r.json()),
      ]);

      const unitData: Test[] =
        unitSettled.status === "fulfilled" ? unitSettled.value || [] : [];
      const regressionData: Test[] =
        regressionSettled.status === "fulfilled"
          ? regressionSettled.value || []
          : [];

      if (unitSettled.status === "rejected")
        console.error("Failed to fetch unit tests:", unitSettled.reason);
      if (regressionSettled.status === "rejected")
        console.error("Failed to fetch regression tests:", regressionSettled.reason);

      const unitTests: TestWithType[] = unitData.map((t) => ({ ...t, type: "unit" }));
      const regressionTests: TestWithType[] = regressionData.map((t) => ({
        ...t,
        type: "regression",
      }));

      setTests([...unitTests, ...regressionTests]);
    } catch (err) {
      console.error("Failed to fetch tests:", err);
    } finally {
      setLoading(false);
    }
  }, []);

  const runSnapshot = async () => {
    setSnapshotting(true);
    try {
      await fetch(`${API}/api/v1/regressions/generate`, { method: "POST" });
      await fetchTests();
    } catch (err) {
      console.error("Failed to generate regressions:", err);
    } finally {
      setSnapshotting(false);
    }
  };

  const applyResults = (data: TestRunResult[]) => {
    // eslint-disable-next-line react-hooks/purity
    const now = Date.now();
    setResults((prev) => {
      const next = { ...prev };
      data.forEach((res) => {
        next[res.name] = res;
      });
      saveCache("litmus:lab:results", next);
      return next;
    });
    setLastRunTs(now);
  };

  const runAllTests = async () => {
    setRunning(true);
    try {
      const toRun: Promise<TestRunResult[]>[] = [];
      if (filter === "all" || filter === "unit") {
        toRun.push(
          fetch(`${API}/api/v1/tests/run`, { method: "POST" }).then((r) =>
            r.json(),
          ),
        );
      }
      if (filter === "all" || filter === "regression") {
        toRun.push(
          fetch(`${API}/api/v1/regressions/run`, { method: "POST" }).then((r) =>
            r.json(),
          ),
        );
      }

      const allResults = await Promise.all(toRun);
      const incoming: Record<string, TestRunResult> = {};
      allResults.flat().forEach((res) => {
        incoming[res.name] = res;
      });

      const now = Date.now();
      const merged = { ...results, ...incoming };
      let passed = 0;
      let failed = 0;
      Object.values(merged).forEach((r) => {
        if (r.pass) passed++;
        else failed++;
      });
      saveCache("litmus:lab:results", merged);
      setResults(merged);
      setLastRunTs(now);
      onTestsRun(passed, failed);
    } catch (err) {
      console.error("Failed to run tests:", err);
    } finally {
      setRunning(false);
    }
  };

  const runSingleTest = async (testName: string, testType: string) => {
    setRunningTest(testName);
    try {
      const endpoint =
        testType === "regression"
          ? `${API}/api/v1/regressions/run?name=${encodeURIComponent(testName)}`
          : `${API}/api/v1/tests/run?name=${encodeURIComponent(testName)}`;
      const resp = await fetch(endpoint, { method: "POST" });
      const data = await resp.json();
      applyResults(data);
    } catch (err) {
      console.error("Failed to run test:", err);
    } finally {
      setRunningTest(null);
    }
  };

  const getTestType = (test: TestWithType): string => test.type;

  const filteredTests = tests
    .filter((test) => {
      if (filter === "all") return true;
      return getTestType(test) === filter;
    })
    .sort((a, b) => {
      const resA = results[a.name];
      const resB = results[b.name];

      // If both have results, sort failures first
      if (resA && resB) {
        if (resA.pass !== resB.pass) {
          return resA.pass ? 1 : -1;
        }
      } else if (resA && !resB) {
        return resA.pass ? 1 : -1;
      } else if (!resA && resB) {
        return resB.pass ? -1 : 1;
      }
      return 0;
    });

  useEffect(() => {
    localStorage.removeItem("litmus:lab:results");
    // eslint-disable-next-line react-hooks/set-state-in-effect
    fetchTests();
  }, [fetchTests]);

  const filterTabs = [
    { label: "All", value: "all" as FilterType, count: tests.length },
    {
      label: "Unit",
      value: "unit" as FilterType,
      count: tests.filter((t) => getTestType(t) === "unit").length,
    },
    {
      label: "Regression",
      value: "regression" as FilterType,
      count: tests.filter((t) => getTestType(t) === "regression").length,
    },
  ];

  return (
    <div className="flex-1 flex flex-col min-h-0 bg-[#181b1f]">
      <Header title="Test Lab" />
      <main className="flex-1 p-6 overflow-y-auto">
        {/* Toolbar */}
        <div className="flex items-center justify-between mb-4">
          <div className="flex items-center gap-3">
            <h3 className="text-[#d9d9d9] font-semibold text-sm">
              {filter === "unit"
                ? "Unit Tests"
                : filter === "regression"
                  ? "Regression Tests"
                  : "All Tests"}
            </h3>
            <span className="text-[11px] text-[#8e9193]">
              {filteredTests.length} of {tests.length}
            </span>
            <LastUpdated ts={lastRunTs} />
          </div>
          <div className="flex items-center">
            <div className="relative flex items-stretch">
              <button
                onClick={handleAction}
                disabled={running || snapshotting}
                className="flex items-center gap-2 px-4 py-[7px] rounded-l bg-[#f46800] hover:bg-[#ff7f2a] disabled:opacity-40 text-white text-sm font-semibold transition-colors border-r border-black/20"
              >
                {running || snapshotting ? (
                  <GfSpinner size="sm" />
                ) : (
                  <currentAction.icon size={12} />
                )}
                {running
                  ? "Running…"
                  : snapshotting
                    ? "Snapshotting…"
                    : currentAction.label}
              </button>
              <button
                onClick={() => setShowDropdown(!showDropdown)}
                disabled={running || snapshotting}
                className="flex items-center px-2 py-[7px] rounded-r bg-[#f46800] hover:bg-[#ff7f2a] disabled:opacity-40 text-white transition-colors"
              >
                <ChevronDown size={14} />
              </button>

              {showDropdown && (
                <>
                  <div
                    className="fixed inset-0 z-10"
                    onClick={() => setShowDropdown(false)}
                  />
                  <div className="absolute right-0 mt-1 w-64 bg-[#1f2128] border border-[#34383e] rounded shadow-xl z-20 py-1">
                    {actions.map((action) => (
                      <button
                        key={action.id}
                        onClick={() => {
                          setActiveAction(action.id);
                          setShowDropdown(false);
                        }}
                        className={cn(
                          "w-full text-left px-4 py-2.5 hover:bg-[#ffffff08] transition-colors group",
                          activeAction === action.id && "bg-[#f46800]/10",
                        )}
                      >
                        <div className="flex items-center gap-2">
                          <action.icon
                            size={14}
                            className={
                              activeAction === action.id
                                ? "text-[#f46800]"
                                : "text-[#8e9193] group-hover:text-[#d9d9d9]"
                            }
                          />
                          <span
                            className={cn(
                              "text-sm font-medium",
                              activeAction === action.id
                                ? "text-[#f46800]"
                                : "text-[#d9d9d9]",
                            )}
                          >
                            {action.label}
                          </span>
                        </div>
                        <p className="text-[11px] text-[#8e9193] mt-0.5 ml-6">
                          {action.desc}
                        </p>
                      </button>
                    ))}
                  </div>
                </>
              )}
            </div>
          </div>
        </div>

        {/* Underline tabs */}
        <FilterTabs tabs={filterTabs} active={filter} onChange={setFilter} />

        {/* Content */}
        {loading ? (
          <div className="flex items-center justify-center h-48">
            <GfSpinner size="lg" />
          </div>
        ) : filteredTests.length === 0 ? (
          <EmptyState
            icon={FlaskConical}
            title="No tests found"
            description={
              filter === "regression"
                ? "Run 'litmus snapshot' to generate a regression baseline"
                : "Add YAML test files to the tests/ directory"
            }
          />
        ) : (
          <div className="space-y-2">
            {filteredTests.map((test) => (
              <TestCaseCard
                key={`${test.name}-${test.type}`}
                test={test}
                result={results[test.name]}
                isRunning={runningTest === test.name}
                globalRunning={running}
                onRun={() => runSingleTest(test.name, getTestType(test))}
              />
            ))}
          </div>
        )}
      </main>
    </div>
  );
};
