# LIT-2: Core Data Types & Serialization
**Summary:** Implement shared Behavioral and Regression test structs with MsgPack/YAML tags.
**Component:** internal/types, internal/codec
**Priority:** High

### Description:
Define the `BehavioralTest` and `RegressionTest` structs in the `types` package. Implement the serialization logic in the `codec` package to handle YAML (human) and MessagePack (machine) formats.

### Tasks:
*   [ ] Implement `BehavioralTest` struct in `internal/types/behavioral.go`.
*   [ ] Implement `RegressionTest` struct in `internal/types/regression.go`.
*   [ ] Implement MessagePack encoder/decoder in `internal/codec/msgpack.go`.
*   [ ] Implement YAML encoder/decoder in `internal/codec/yaml.go`.

### Acceptance Criteria:
*   [ ] `RegressionTest` can be round-tripped through MessagePack without loss of data.
*   [ ] `BehavioralTest` can be correctly parsed from YAML files.
*   [ ] Types are isolated and do not import any other internal packages.
