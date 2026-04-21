# Plan: Named Snapshots — Multi-Environment Baselines

## Context
A single baseline can't represent multiple environments. Add `config.multi_env: true` to `litmus.yaml` which makes litmus discover environment subdirectories and run snapshot/check/diff per-env. By default runs all discovered envs; `--environment=prod` targets one. Fully backward-compatible — single-env users see zero behaviour change.

---

## Folder Structure (Option A — by-concern)
```
config/
  production/alertmanager.yml
  staging/alertmanager.yml
tests/
  production/routing-critical.yml
  staging/routing-critical.yml
regressions/
  production/regressions.litmus.mpk
  production/regressions.litmus.yml
  staging/regressions.litmus.mpk
  staging/regressions.litmus.yml
```
`.litmus.yaml`:
```yaml
config:
  directory: config/
  file: alertmanager.yml
  multi_env: true
```

---

## Path Rules

| Mode | Alertmanager config | Tests dir | Regression baseline |
|---|---|---|---|
| `multi_env: false` (default) | `config/alertmanager.yml` | `tests/` | `regressions/regressions.litmus.mpk` |
| `multi_env: true`, env=`prod` | `config/prod/alertmanager.yml` | `tests/prod/` | `regressions/prod/regressions.litmus.mpk` |

---

## Files to Modify

### 1. `internal/config/types.go`
Add `MultiEnv bool` to `Config`:
```go
type Config struct {
    Directory string `yaml:"directory" mapstructure:"directory"`
    File      string `yaml:"file"      mapstructure:"file"`
    Templates string `yaml:"templates" mapstructure:"templates"`
    MultiEnv  bool   `yaml:"multi_env" mapstructure:"multi_env"`
}
```

### 2. `internal/config/loader.go`
- Add Viper default: `v.SetDefault("config.multi_env", false)`
- Add exported helper used by all CLI commands:
```go
func DiscoverEnvironments(dir string) ([]string, error) {
    entries, err := os.ReadDir(dir)
    // return names of subdirectories only
}
```

### 3. `internal/cli/helpers.go`
Add three unexported path helpers (reused by snapshot, check, diff):
```go
func envAlertConfigPath(cfg *config.LitmusConfig, env string) string {
    if env == "" { return filepath.Join(cfg.Config.Directory, cfg.Config.File) }
    return filepath.Join(cfg.Config.Directory, env, cfg.Config.File)
}

func envTestsDir(cfg *config.LitmusConfig, env string) string {
    if env == "" { return cfg.Tests.Directory }
    return filepath.Join(cfg.Tests.Directory, env)
}

func envBaselinePath(cfg *config.LitmusConfig, env string) string {
    if env == "" { return filepath.Join(cfg.Regression.Directory, "regressions.litmus.mpk") }
    return filepath.Join(cfg.Regression.Directory, env, "regressions.litmus.mpk")
}
```
`env == ""` preserves exact current paths — zero migration for existing users.

### 4. `cmd/snapshot.go`
Add `--environment/-e` string flag (default `""`):
```go
env, _ := cmd.Flags().GetString("environment")
return cli.RunSnapshot(update, strict, env)
cmd.Flags().StringP("environment", "e", "", "target a single environment (default: all when multi_env is true)")
```

### 5. `internal/cli/snapshot.go`
Update `RunSnapshot(update, strict bool, env string) error`:
- If `multi_env: false` OR `env != ""`: call `runSnapshotForEnv` once
- If `multi_env: true` AND `env == ""`: call `config.DiscoverEnvironments`, loop calling `runSnapshotForEnv` per env

Extract single-env logic into:
```go
func runSnapshotForEnv(litmusConfig *config.LitmusConfig, update, strict bool, env string) error {
    bp := envBaselinePath(litmusConfig, env)
    ymlPath := strings.TrimSuffix(bp, ".mpk") + ".yml"
    // current write logic unchanged
}
```

### 6. `cmd/check.go`
Add `--environment/-e` string flag (default `""`):
```go
env, _ := cmd.Flags().GetString("environment")
code, err := cli.RunCheck(format, diff, tags, env)
cmd.Flags().StringP("environment", "e", "", "target a single environment (default: all when multi_env is true)")
```

### 7. `internal/cli/check.go`
Update signatures:
```go
func RunCheck(format string, showDiff bool, tags []string, env string) (CheckExitCode, error)
func RunRegressionTests(ctx, litmusConfig, router, tags []string, env string) RegressionResult
func RunBehavioralTests(ctx, litmusConfig, router, inhibitRules, tags []string, env string) BehavioralResult
```

Multi-env orchestration in `RunCheck`:
- `multi_env: false` OR `env != ""`: single-env run (current flow)
- `multi_env: true` AND `env == ""`: discover envs, run each, aggregate into per-env results

Output format (multi-env):
```
[production]
  Sanity:     PASS
  Regression: 10/10
  Behavioral: 5/5

[staging]
  Sanity:     PASS
  Regression: 9/10  FAIL

SUMMARY: production PASS | staging FAIL
```

### 8. `cmd/diff.go`
Add `--environment/-e` string flag:
```go
env, _ := cmd.Flags().GetString("environment")
return cli.RunDiff(env)
cmd.Flags().StringP("environment", "e", "", "target a single environment")
```

### 9. `internal/cli/diff.go`
Update `RunDiff(env string) error`:
- Replace hardcoded paths with `envAlertConfigPath` and `envBaselinePath`
- If `multi_env: true` AND `env == ""`: discover envs, run diff per env

---

## Existing Functions to Reuse
- `LoadBaseline(path string)` — `internal/cli/helpers.go:12` — path-agnostic, no changes needed
- `LoadBaselineYAML(path string)` — `internal/cli/helpers.go:27` — same
- `RunSanityChecks` — no env concept, unchanged
- All executors (behavioral, snapshot) — unchanged; receive data, not paths

---

## Backward Compatibility
- `multi_env` defaults to `false` → all `env == ""` paths resolve identically to today
- No existing tests break
- `strings.Replace(baselinePath, "mpk", "yml", 1)` brittle pattern replaced with `strings.TrimSuffix(bp, ".mpk") + ".yml"`

---

## New Tests
- `TestSnapshotCommand_MultiEnv` — `multi_env: true` creates per-env `.mpk` files
- `TestSnapshotCommand_MultiEnv_SingleTarget` — `--environment=prod` creates only prod baseline
- `TestCheckCommand_MultiEnv` — check runs all envs, reports per-env
- `TestDiscoverEnvironments` — unit test for discovery helper
