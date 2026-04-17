# Graph Report - .  (2026-04-17)

## Corpus Check
- 10 files · ~9,301 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 46 nodes · 66 edges · 8 communities detected
- Extraction: 61% EXTRACTED · 39% INFERRED · 0% AMBIGUOUS · INFERRED: 26 edges (avg confidence: 0.8)
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

## God Nodes (most connected - your core abstractions)
1. `TestAlertStore_Subscribe()` - 6 edges
2. `AlertStore` - 6 edges
3. `TestRegressionTestRoundTrip()` - 5 edges
4. `TestAlertStore_Reset()` - 5 edges
5. `NewAlertStore()` - 5 edges
6. `EncodeYAML()` - 4 edges
7. `TestSilenceStore_Reset()` - 4 edges
8. `SilenceStore` - 4 edges
9. `TestAlertStore_Get()` - 4 edges
10. `alertIterator` - 4 edges

## Surprising Connections (you probably didn't know these)
- `DecodeMsgPack()` --calls--> `TestRegressionTestRoundTrip()`  [INFERRED]
  internal/codec/msgpack.go → internal/codec/codec_test.go
- `EncodeMsgPack()` --calls--> `TestRegressionTestRoundTrip()`  [INFERRED]
  internal/codec/msgpack.go → internal/codec/codec_test.go
- `TestRegressionTestRoundTrip()` --calls--> `EncodeYAML()`  [INFERRED]
  internal/codec/codec_test.go → internal/codec/yaml.go
- `TestRegressionTestRoundTrip()` --calls--> `DecodeYAML()`  [INFERRED]
  internal/codec/codec_test.go → internal/codec/yaml.go
- `TestBehavioralTestRoundTrip()` --calls--> `EncodeYAML()`  [INFERRED]
  internal/codec/codec_test.go → internal/codec/yaml.go

## Communities

### Community 0 - "Community 0"
Cohesion: 0.29
Nodes (4): NewSilenceStore(), TestSilenceStore_Mutes(), TestSilenceStore_Reset(), SilenceStore

### Community 1 - "Community 1"
Cohesion: 0.33
Nodes (6): TestBehavioralTestRoundTrip(), TestRegressionTestRoundTrip(), DecodeMsgPack(), EncodeMsgPack(), DecodeYAML(), EncodeYAML()

### Community 2 - "Community 2"
Cohesion: 0.5
Nodes (5): NewAlertStore(), TestAlertStore_Get(), TestAlertStore_Put(), TestAlertStore_Reset(), TestAlertStore_Subscribe()

### Community 3 - "Community 3"
Cohesion: 0.33
Nodes (5): AlertSample, BehavioralExpect, BehavioralTest, Silence, SystemState

### Community 4 - "Community 4"
Cohesion: 0.6
Nodes (2): newAlertIterator(), AlertStore

### Community 5 - "Community 5"
Cohesion: 0.67
Nodes (1): alertIterator

### Community 6 - "Community 6"
Cohesion: 1.0
Nodes (0): 

### Community 7 - "Community 7"
Cohesion: 1.0
Nodes (1): RegressionTest

## Knowledge Gaps
- **6 isolated node(s):** `RegressionTest`, `BehavioralTest`, `SystemState`, `AlertSample`, `Silence` (+1 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **Thin community `Community 6`** (2 nodes): `main.go`, `main()`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 7`** (2 nodes): `regression.go`, `RegressionTest`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `EncodeYAML()` connect `Community 1` to `Community 5`?**
  _High betweenness centrality (0.223) - this node is a cross-community bridge._
- **Why does `TestAlertStore_Subscribe()` connect `Community 2` to `Community 4`, `Community 5`?**
  _High betweenness centrality (0.212) - this node is a cross-community bridge._
- **Are the 5 inferred relationships involving `TestAlertStore_Subscribe()` (e.g. with `NewAlertStore()` and `.Put()`) actually correct?**
  _`TestAlertStore_Subscribe()` has 5 INFERRED edges - model-reasoned connections that need verification._
- **Are the 4 inferred relationships involving `TestRegressionTestRoundTrip()` (e.g. with `EncodeYAML()` and `DecodeYAML()`) actually correct?**
  _`TestRegressionTestRoundTrip()` has 4 INFERRED edges - model-reasoned connections that need verification._
- **Are the 4 inferred relationships involving `TestAlertStore_Reset()` (e.g. with `NewAlertStore()` and `.Put()`) actually correct?**
  _`TestAlertStore_Reset()` has 4 INFERRED edges - model-reasoned connections that need verification._
- **Are the 4 inferred relationships involving `NewAlertStore()` (e.g. with `TestAlertStore_Put()` and `TestAlertStore_Get()`) actually correct?**
  _`NewAlertStore()` has 4 INFERRED edges - model-reasoned connections that need verification._
- **What connects `RegressionTest`, `BehavioralTest`, `SystemState` to the rest of the system?**
  _6 weakly-connected nodes found - possible documentation gaps or missing edges._