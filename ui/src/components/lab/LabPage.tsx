import { useState, useEffect, useCallback } from "react";
import { FlaskConical, ChevronDown, CheckCircle2, AlertTriangle, X } from "lucide-react";
import { API, cn, minDelay } from "../../utils/ui";
import { useLabStore } from "../../stores/useLabStore";
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

export const LabPage = () => {
  const { results, setResults, lastRunTs, setLastRunTs } = useLabStore();
  const [tests, setTests] = useState<TestWithType[]>([]);
  const [loading, setLoading] = useState(false);
  const [running, setRunning] = useState(false);
  const [snapshotting, setSnapshotting] = useState(false);
  const [runningTest, setRunningTest] = useState<string | null>(null);
  const [filter, setFilter] = useState<FilterType>("all");
  const [activeAction, setActiveAction] = useState<"run" | "snapshot">("run");
  const [showDropdown, setShowDropdown] = useState(false);
  const [notification, setNotification] = useState<{
    type: "success" | "error";
    message: string;
  } | null>(null);

  useEffect(() => {
    if (notification) {
      const timer = setTimeout(() => setNotification(null), 5000);
      return () => clearTimeout(timer);
    }
  }, [notification]);

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
    setNotification(null);
    try {
      const fetchPromise = fetch(`${API}/api/v1/regressions/generate`, { method: "POST" });
      const resp = await minDelay(fetchPromise);
      if (!resp.ok) throw new Error(await resp.text());
      await fetchTests();
      setNotification({
        type: "success",
        message: "Regression snapshot generated successfully",
      });
    } catch (err: any) {
      console.error("Failed to generate regressions:", err);
      setNotification({
        type: "error",
        message: `Snapshot failed: ${err.message || String(err)}`,
      });
    } finally {
      setSnapshotting(false);
    }
  };

  const applyResults = (data: TestRunResult[]) => {
    const now = Date.now();
    const next = { ...results };
    data.forEach((res) => {
      next[res.name] = res;
    });
    setResults(next);
    setLastRunTs(now);
  };

  const runAllTests = async () => {
    setRunning(true);
    setNotification(null);
    try {
      const toRun: Promise<TestRunResult[]>[] = [];
      if (filter === "all" || filter === "unit") {
        toRun.push(
          minDelay(
            fetch(`${API}/api/v1/tests/run`, { method: "POST" }).then((r) => {
              if (!r.ok) throw new Error("Failed to run unit tests");
              return r.json();
            }),
          ),
        );
      }
      if (filter === "all" || filter === "regression") {
        toRun.push(
          minDelay(
            fetch(`${API}/api/v1/regressions/run`, { method: "POST" }).then(
              (r) => {
                if (!r.ok) throw new Error("Failed to run regression tests");
                return r.json();
              },
            ),
          ),
        );
      }

      const allResults = await Promise.all(toRun);
      const incoming: Record<string, TestRunResult> = {};
      const resultsArray = allResults.flat();
      resultsArray.forEach((res) => {
        incoming[res.name] = res;
      });

      const now = Date.now();
      const merged = { ...results, ...incoming };
      setResults(merged);
      setLastRunTs(now);

      const passedCount = resultsArray.filter((r) => r.pass).length;
      setNotification({
        type: "success",
        message: `Test run completed: ${passedCount}/${resultsArray.length} passed`,
      });
    } catch (err: any) {
      console.error("Failed to run tests:", err);
      setNotification({
        type: "error",
        message: `Test run failed: ${err.message || String(err)}`,
      });
    } finally {
      setRunning(false);
    }
  };

  const runSingleTest = async (testName: string, testType: string) => {
    setRunningTest(testName);
    setNotification(null);
    try {
      const endpoint =
        testType === "regression"
          ? `${API}/api/v1/regressions/run?name=${encodeURIComponent(testName)}`
          : `${API}/api/v1/tests/run?name=${encodeURIComponent(testName)}`;
      const resp = await minDelay(fetch(endpoint, { method: "POST" }));
      if (!resp.ok) throw new Error(await resp.text());
      const data = await resp.json();
      applyResults(data);
      const passed = data.every((r: any) => r.pass);
      setNotification({
        type: passed ? "success" : "error",
        message: passed
          ? `Test '${testName}' passed`
          : `Test '${testName}' failed`,
      });
    } catch (err: any) {
      console.error("Failed to run test:", err);
      setNotification({
        type: "error",
        message: `Failed to run test: ${err.message || String(err)}`,
      });
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
        {/* Notifications */}
        {notification && (
          <div
            className={cn(
              "mb-5 px-4 py-3 rounded border flex items-center justify-between animate-in fade-in slide-in-from-top-2",
              notification.type === "success"
                ? "bg-[#73bf69]/10 border-[#73bf69]/20 text-[#73bf69]"
                : "bg-[#f2495c]/10 border-[#f2495c]/20 text-[#f2495c]",
            )}
          >
            <div className="flex items-center gap-2 text-sm font-medium">
              {notification.type === "success" ? (
                <CheckCircle2 size={16} />
              ) : (
                <AlertTriangle size={16} />
              )}
              {notification.message}
            </div>
            <button
              onClick={() => setNotification(null)}
              className="opacity-50 hover:opacity-100 transition-opacity"
            >
              <X size={14} />
            </button>
          </div>
        )}

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
          <div className="space-y-2 animate-fade-in-up">
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
