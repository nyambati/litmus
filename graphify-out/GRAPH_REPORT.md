# Graph Report - .  (2026-04-17)

## Corpus Check
- 34 files · ~16,049 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 184 nodes · 341 edges · 11 communities detected
- Extraction: 53% EXTRACTED · 47% INFERRED · 0% AMBIGUOUS · INFERRED: 159 edges (avg confidence: 0.8)
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
7. `newSnapshotCmd()` - 7 edges
8. `NewBehavioralTestExecutor()` - 7 edges
9. `NewShadowedRouteDetector()` - 7 edges
10. `newCheckCmd()` - 6 edges

## Surprising Connections (you probably didn't know these)
- `runSanityChecks()` --calls--> `NewShadowedRouteDetector()`  [INFERRED]
  cmd/litmus/check.go → internal/engine/sanity/shadowed.go
- `runBehavioralTests()` --calls--> `NewBehavioralTestLoader()`  [INFERRED]
  cmd/litmus/check.go → internal/engine/behavioral/loader.go
- `runSnapshot()` --calls--> `NewRouteWalker()`  [INFERRED]
  cmd/litmus/snapshot.go → internal/engine/snapshot/route_walker.go
- `runSnapshot()` --calls--> `EncodeMsgPack()`  [INFERRED]
  cmd/litmus/snapshot.go → internal/codec/msgpack.go
- `loadBaseline()` --calls--> `DecodeMsgPack()`  [INFERRED]
  cmd/litmus/snapshot.go → internal/codec/msgpack.go

## Communities

### Community 0 - "Community 0"
Cohesion: 0.11
Nodes (27): BehavioralTestExecutor, TestResult, newCheckCmd(), TestCheckCommand_MissingConfig(), TestCheckCommand_Success(), TestCheckCommand_TextOutput(), NewBehavioralTestExecutor(), receiversMatch() (+19 more)

### Community 1 - "Community 1"
Cohesion: 0.16
Nodes (11): NewRouteWalker(), TestRouteWalker_FindTerminalPaths(), ShadowedRouteDetector, NewShadowedRouteDetector(), TestShadowedRouteDetector_NoShadow(), TestShadowedRouteDetector_NotShadowedDifferentMatchers(), TestShadowedRouteDetector_ShadowedByParent(), TestShadowedRouteDetector_ShadowedByParentNoChild() (+3 more)

### Community 2 - "Community 2"
Cohesion: 0.16
Nodes (11): NewAlertStore(), NewRunner(), Outcome, TestPipeline_Execute_Active(), TestPipeline_Execute_Inhibited(), TestPipeline_Execute_Silenced(), NewSilenceStore(), TestSilenceStore_Mutes() (+3 more)

### Community 3 - "Community 3"
Cohesion: 0.18
Nodes (8): newAlertIterator(), TestAlertStore_Get(), TestAlertStore_Put(), TestAlertStore_Reset(), TestAlertStore_Subscribe(), Runner, alertIterator, AlertStore

### Community 4 - "Community 4"
Cohesion: 0.16
Nodes (10): NewLabelCombinationGenerator(), NewRegexExpander(), TestLabelCombinations_BalancedCovering(), TestRegexExpansion_ExpandAlternations(), verifyCoverage(), LabelCombinationGenerator, NewSnapshotSynthesizer(), RegexExpander (+2 more)

### Community 5 - "Community 5"
Cohesion: 0.17
Nodes (16): printTextOutput(), runBehavioralTests(), runCheck(), runRegressionTests(), BehavioralResult, CheckResult, LitmusConfig, RegressionConfig (+8 more)

### Community 6 - "Community 6"
Cohesion: 0.23
Nodes (10): runSanityChecks(), NewInhibitionCycleDetector(), NewOrphanReceiverDetector(), TestInhibitionCycleDetector_DirectCycle(), TestInhibitionCycleDetector_NoCycle(), TestOrphanReceiverDetector_HasOrphans(), TestOrphanReceiverDetector_NestedRoutes(), TestOrphanReceiverDetector_NoOrphans() (+2 more)

### Community 7 - "Community 7"
Cohesion: 0.2
Nodes (11): TestBehavioralTestRoundTrip(), TestRegressionTestRoundTrip(), newInspectCmd(), runInspect(), TestInspectCommand_JSON(), TestInspectCommand_MissingFile(), TestInspectCommand_YAML(), DecodeMsgPack() (+3 more)

### Community 8 - "Community 8"
Cohesion: 0.39
Nodes (4): BehavioralTestLoader, NewBehavioralTestLoader(), TestBehavioralTestLoader_LoadFromDirectory(), TestBehavioralTestLoader_LoadFromFile()

### Community 9 - "Community 9"
Cohesion: 0.33
Nodes (5): AlertSample, BehavioralExpect, BehavioralTest, Silence, SystemState

### Community 10 - "Community 10"
Cohesion: 1.0
Nodes (1): RegressionTest

## Knowledge Gaps
- **17 isolated node(s):** `CheckResult`, `SanityResult`, `RegressionResult`, `BehavioralResult`, `LitmusConfig` (+12 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **Thin community `Community 10`** (2 nodes): `regression.go`, `RegressionTest`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `runSanityChecks()` connect `Community 6` to `Community 1`, `Community 5`?**
  _High betweenness centrality (0.201) - this node is a cross-community bridge._
- **Why does `newCheckCmd()` connect `Community 0` to `Community 5`?**
  _High betweenness centrality (0.195) - this node is a cross-community bridge._
- **Why does `runSnapshot()` connect `Community 5` to `Community 0`, `Community 1`, `Community 3`, `Community 7`?**
  _High betweenness centrality (0.187) - this node is a cross-community bridge._
- **Are the 4 inferred relationships involving `runSnapshot()` (e.g. with `NewRouteWalker()` and `.FindTerminalPaths()`) actually correct?**
  _`runSnapshot()` has 4 INFERRED edges - model-reasoned connections that need verification._
- **Are the 9 inferred relationships involving `NewAlertStore()` (e.g. with `TestAlertStore_Put()` and `TestAlertStore_Get()`) actually correct?**
  _`NewAlertStore()` has 9 INFERRED edges - model-reasoned connections that need verification._
- **Are the 2 inferred relationships involving `runCheck()` (e.g. with `loadLitmusConfig()` and `loadAlertmanagerConfig()`) actually correct?**
  _`runCheck()` has 2 INFERRED edges - model-reasoned connections that need verification._
- **Are the 6 inferred relationships involving `runSanityChecks()` (e.g. with `NewShadowedRouteDetector()` and `.Detect()`) actually correct?**
  _`runSanityChecks()` has 6 INFERRED edges - model-reasoned connections that need verification._