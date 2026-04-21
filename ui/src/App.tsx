import { BrowserRouter as Router, Routes, Route } from "react-router-dom";

// Layout components
import { AppLayout } from "./components/layout/AppLayout";
import {
  ExplorerStats,
  LabStats,
  RegressionStats,
} from "./components/layout/PageStats";

// Page components
import { ExplorerPage } from "./components/explorer/ExplorerPage";
import { LabPage } from "./components/lab/LabPage";
import { RegressionPage } from "./components/regression/RegressionPage";

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
      </Routes>
    </Router>
  );
}

export default App;
