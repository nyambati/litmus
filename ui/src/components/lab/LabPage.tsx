import { useState, useEffect } from "react";
import { FlaskConical } from "lucide-react";
import { API, loadCache, saveCache } from "../../utils/persistence";
import { GfSpinner } from "../ui/Spinner";
import { LastUpdated } from "../ui/LastUpdated";
import { PrimaryButton } from "../ui/Buttons";
import { Header } from "../layout/Header";
import { EmptyState } from "../ui/EmptyState";
import { FilterTabs } from "../ui/Tabs";
import { TestCaseCard } from "./TestCase";

type FilterType = "all" | "unit" | "regression";

export const LabPage = ({
  onTestsRun,
}: {
  onTestsRun: (passed: number, failed: number) => void;
}) => {
  const [tests, setTests] = useState<any[]>([]);
  const _labCache = loadCache<Record<string, any>>("litmus:lab:results");
  const [results, setResults] = useState<Record<string, any>>(
    _labCache?.data ?? {},
  );
  const [lastRunTs, setLastRunTs] = useState<number | null>(
    _labCache?.ts ?? null,
  );
  const [loading, setLoading] = useState(false);
  const [running, setRunning] = useState(false);
  const [runningTest, setRunningTest] = useState<string | null>(null);
  const [filter, setFilter] = useState<FilterType>("all");

  const fetchTests = async () => {
    setLoading(true);
    try {
      const [unitSettled, regressionSettled] = await Promise.allSettled([
        fetch(`${API}/api/v1/tests`).then((r) => r.json()),
        fetch(`${API}/api/v1/regressions`).then((r) => r.json()),
      ]);

      const unitData: any[] =
        unitSettled.status === "fulfilled" ? unitSettled.value || [] : [];
      const regressionData: any[] =
        regressionSettled.status === "fulfilled"
          ? regressionSettled.value || []
          : [];

      if (unitSettled.status === "rejected")
        console.error("Failed to fetch unit tests:", unitSettled.reason);
      if (regressionSettled.status === "rejected")
        console.error("Failed to fetch regression tests:", regressionSettled.reason);

      const unitTests = unitData.map((t) => ({ ...t, type: "unit" }));
      const regressionTests = regressionData.map((t) => ({
        ...t,
        type: "regression",
      }));

      setTests([...unitTests, ...regressionTests]);
    } catch (err) {
      console.error("Failed to fetch tests:", err);
    } finally {
      setLoading(false);
    }
  };

  const applyResults = (data: any[]) => {
    setResults((prev) => {
      const next = { ...prev };
      data.forEach((res: any) => {
        next[res.name] = res;
      });
      saveCache("litmus:lab:results", next);
      return next;
    });
    setLastRunTs(Date.now());
  };

  const runAllTests = async () => {
    setRunning(true);
    try {
      const toRun: Promise<any[]>[] = [];
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

      const allResults = (await Promise.all(toRun)).flat();
      const incoming: Record<string, any> = {};
      allResults.forEach((res: any) => {
        incoming[res.name] = res;
      });

      setResults((prev) => {
        const merged = { ...prev, ...incoming };
        let passed = 0;
        let failed = 0;
        Object.values(merged).forEach((r: any) => {
          if (r.pass) passed++;
          else failed++;
        });
        onTestsRun(passed, failed);
        saveCache("litmus:lab:results", merged);
        return merged;
      });
      setLastRunTs(Date.now());
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

  const getTestType = (test: any): string => test.type || "unit";

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
    fetchTests();
  }, []);

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
          <PrimaryButton
            onClick={runAllTests}
            disabled={running || filteredTests.length === 0}
            loading={running}
            icon={<FlaskConical size={14} />}
          >
            {running ? "Running…" : "Run All"}
          </PrimaryButton>
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
            {filteredTests.map((test, testIdx) => (
              <TestCaseCard
                key={`${test.name}-${testIdx}`}
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
