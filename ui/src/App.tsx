import { BrowserRouter as Router, Routes, Route } from "react-router-dom";

// Layout components
import { AppLayout } from "./components/layout/AppLayout";
import {
  ExplorerStats,
  LabStats,
  RegressionStats,
  FragmentsStats,
} from "./components/layout/PageStats";

// Page components
import { ExplorerPage } from "./components/explorer/ExplorerPage";
import { LabPage } from "./components/lab/LabPage";
import { RegressionPage } from "./components/regression/RegressionPage";
import { FragmentsPage } from "./components/fragments/FragmentsPage";

function App() {
  return (
    <Router>
      <Routes>
        <Route
          path="/"
          element={
            <AppLayout stats={<ExplorerStats />}>
              <ExplorerPage />
            </AppLayout>
          }
        />

        <Route
          path="/lab"
          element={
            <AppLayout stats={<LabStats />}>
              <LabPage />
            </AppLayout>
          }
        />

        <Route
          path="/regression"
          element={
            <AppLayout stats={<RegressionStats />}>
              <RegressionPage />
            </AppLayout>
          }
        />

        <Route
          path="/fragments"
          element={
            <AppLayout stats={<FragmentsStats />}>
              <FragmentsPage />
            </AppLayout>
          }
        />
      </Routes>
    </Router>
  );
}

export default App;
