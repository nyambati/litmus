# Graph Report - .  (2026-04-17)

## Corpus Check
- 36 files · ~21,203 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 203 nodes · 369 edges · 11 communities detected
- Extraction: 58% EXTRACTED · 42% INFERRED · 0% AMBIGUOUS · INFERRED: 156 edges (avg confidence: 0.8)
- Token cost: 0 input · 0 output

## Community Hubs (Navigation)
- [[_COMMUNITY_Community 0|Community 0]]
- [[_COMMUNITY_Community 1|Community 1]]
- [[_COMMUNITY_Community 2|Community 2]]
- [[_COMMUNITY_Community 3|Community 3]]
- [[_COMMUNITY_Community 4|Community 4]]
- [[_COMMUNITY_Community 5|Community 5]]
- [[_COMMUNITY_Community 6|Community 6]]
- [[_COMMUNITY_Community 7|Community 7]]
- [[_COMMUNITY_Community 8|Community 8]]
- [[_COMMUNITY_Community 9|Community 9]]
- [[_COMMUNITY_Community 10|Community 10]]

## God Nodes (most connected - your core abstractions)
1. `runSnapshot()` - 10 edges
2. `NewAlertStore()` - 10 edges
3. `runCheck()` - 8 edges
4. `runSanityChecks()` - 8 edges
5. `newInitCmd()` - 8 edges
6. `NewSilenceStore()` - 8 edges
7. `detectIssues()` - 8 edges
8. `newSnapshotCmd()` - 7 edges
9. `NewBehavioralTestExecutor()` - 7 edges
10. `newCheckCmd()` - 6 edges

## Surprising Connections (you probably didn't know these)
- `runSanityChecks()` --calls--> `NewShadowedRouteDetector()`  [INFERRED]
  cmd/litmus/check.go → internal/engine/sanity/shadowed.go
- `runSanityChecks()` --calls--> `NewOrphanReceiverDetector()`  [INFERRED]
  cmd/litmus/check.go → internal/engine/sanity/inhibition.go
- `runSanityChecks()` --calls--> `NewInhibitionCycleDetector()`  [INFERRED]
  cmd/litmus/check.go → internal/engine/sanity/inhibition.go
- `runSnapshot()` --calls--> `NewRouteWalker()`  [INFERRED]
  cmd/litmus/snapshot.go → internal/engine/snapshot/route_walker.go
- `runBehavioralTests()` --calls--> `NewBehavioralTestLoader()`  [INFERRED]
  cmd/litmus/check.go → internal/engine/behavioral/loader.go

## Communities

### Community 0 - "Community 0"
Cohesion: 0.1
Nodes (18): newAlertIterator(), NewAlertStore(), TestAlertStore_Get(), TestAlertStore_Put(), TestAlertStore_Reset(), TestAlertStore_Subscribe(), NewRunner(), Outcome (+10 more)

### Community 1 - "Community 1"
Cohesion: 0.13
Nodes (23): BehavioralTestExecutor, TestResult, NewBehavioralTestExecutor(), receiversMatch(), TestBehavioralTestExecutor_Execute_Active(), TestBehavioralTestExecutor_Execute_Inhibited(), TestBehavioralTestExecutor_Execute_OutcomeOnly(), TestBehavioralTestExecutor_Execute_Receivers_Mismatch() (+15 more)

### Community 2 - "Community 2"
Cohesion: 0.12
Nodes (20): TestBehavioralTestRoundTrip(), TestRegressionTestRoundTrip(), newInspectCmd(), runInspect(), TestInspectCommand_JSON(), TestInspectCommand_MissingFile(), TestInspectCommand_YAML(), LitmusConfig (+12 more)

### Community 3 - "Community 3"
Cohesion: 0.13
Nodes (17): BehavioralTestLoader, newCheckCmd(), printTextOutput(), runBehavioralTests(), runCheck(), runRegressionTests(), runSanityChecks(), TestCheckCommand_MissingConfig() (+9 more)

### Community 4 - "Community 4"
Cohesion: 0.13
Nodes (11): newRouteInspector(), NewRouteWalker(), TestRouteWalker_FindTerminalPaths(), TestRouteWalker_MatchersCapture(), routeInspector, sanityMatcher, sanityPath, ShadowedRouteDetector (+3 more)

### Community 5 - "Community 5"
Cohesion: 0.16
Nodes (10): NewLabelCombinationGenerator(), NewRegexExpander(), TestLabelCombinations_BalancedCovering(), TestRegexExpansion_ExpandAlternations(), verifyCoverage(), LabelCombinationGenerator, NewSnapshotSynthesizer(), RegexExpander (+2 more)

### Community 6 - "Community 6"
Cohesion: 0.23
Nodes (9): NewInhibitionCycleDetector(), NewOrphanReceiverDetector(), TestInhibitionCycleDetector_DirectCycle(), TestInhibitionCycleDetector_NoCycle(), TestOrphanReceiverDetector_HasOrphans(), TestOrphanReceiverDetector_NestedRoutes(), TestOrphanReceiverDetector_NoOrphans(), InhibitionCycleDetector (+1 more)

### Community 7 - "Community 7"
Cohesion: 0.45
Nodes (10): containsAny(), detectIssues(), isShadowedVictim(), mustMatcher(), mustRegexp(), TestShadowedRouteDetector(), TestShadowedRouteDetector_MatcherFormats(), TestShadowedRouteDetector_NegativeMatcher() (+2 more)

### Community 8 - "Community 8"
Cohesion: 0.25
Nodes (0): 

### Community 9 - "Community 9"
Cohesion: 0.33
Nodes (5): AlertSample, BehavioralExpect, BehavioralTest, Silence, SystemState

### Community 10 - "Community 10"
Cohesion: 1.0
Nodes (1): RegressionTest

## Knowledge Gaps
- **19 isolated node(s):** `CheckResult`, `SanityResult`, `RegressionResult`, `BehavioralResult`, `LitmusConfig` (+14 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **Thin community `Community 10`** (2 nodes): `regression.go`, `RegressionTest`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `runSnapshot()` connect `Community 2` to `Community 1`, `Community 4`?**
  _High betweenness centrality (0.206) - this node is a cross-community bridge._
- **Why does `runSanityChecks()` connect `Community 3` to `Community 4`, `Community 6`?**
  _High betweenness centrality (0.195) - this node is a cross-community bridge._
- **Why does `newCheckCmd()` connect `Community 3` to `Community 1`?**
  _High betweenness centrality (0.171) - this node is a cross-community bridge._
- **Are the 4 inferred relationships involving `runSnapshot()` (e.g. with `NewRouteWalker()` and `.FindTerminalPaths()`) actually correct?**
  _`runSnapshot()` has 4 INFERRED edges - model-reasoned connections that need verification._
- **Are the 9 inferred relationships involving `NewAlertStore()` (e.g. with `TestAlertStore_Put()` and `TestAlertStore_Get()`) actually correct?**
  _`NewAlertStore()` has 9 INFERRED edges - model-reasoned connections that need verification._
- **Are the 2 inferred relationships involving `runCheck()` (e.g. with `loadLitmusConfig()` and `loadAlertmanagerConfig()`) actually correct?**
  _`runCheck()` has 2 INFERRED edges - model-reasoned connections that need verification._
- **Are the 6 inferred relationships involving `runSanityChecks()` (e.g. with `NewShadowedRouteDetector()` and `.Detect()`) actually correct?**
  _`runSanityChecks()` has 6 INFERRED edges - model-reasoned connections that need verification._