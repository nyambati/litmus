# Litmus Configuration Design

## Overview

Litmus uses a structured configuration with sensible defaults to manage Alertmanager configurations and Mimir integration.

## Default Folder Structure

```
.
├── .litmus.yaml          # litmus configuration
├── config/
│   ├── alertmanager.yml  # alertmanager configuration
│   └── templates/
│       ├── slack.tpl
│       └── email.tpl
├── tests/                # behavioral unit tests
│   └── ...
└── regressions/          # regression test baselines
    └── regressions.litus.mpk
```

## .litmus.yaml Schema

```yaml
config:
  directory: config       # alertmanager config directory
  file: alertmanager.yml  # config filename
  templates: templates/    # templates directory (relative to config.directory)

mimir:
  address: "env(MIMIR_ADDRESS)"   # env substitution supported
  tenant_id: "env(MIMIR_TENANT_ID)"
  api_key: "env(MIMIR_API_KEY)"

regression:
  directory: regressions/ # baseline directory
  max_samples: 5

tests:
  directory: tests/       # behavioral tests directory
```

### Path Resolution

| Field | Expression | Example |
|-------|------------|---------|
| `config.file` | `{config.directory}/{config.file}` | `config/alertmanager.yml` |
| `templates` | `{config.directory}/{config.templates}` | `config/templates/` |
| `baseline` | `{regression.directory}/regressions.litmus.mpk` | `regressions/regressions.litmus.mpk` |

### Mimir Configuration

Mimir credentials support environment variable substitution using `env(VAR)` syntax:

```yaml
mimir:
  address: "https://mimir.example.com"
  tenant_id: "anonymous"
  api_key: "env(MIMIR_API_KEY)"
```

When `env(VAR)` is encountered, litmus replaces it with the value of the environment variable.

## Configuration Loading

### Precedence Order

1. **CLI Flags** (highest priority)
2. **Environment Variables** (`LITMUS_*` prefix)
3. **`.litmus.yaml`** (project config)
4. **Defaults** (lowest priority)

### Environment Variables

| Variable | Config Key | Description |
|----------|------------|-------------|
| `LITMUS_CONFIG_DIRECTORY` | `config.directory` | Config directory |
| `LITMUS_CONFIG_FILE` | `config.file` | Config filename |
| `LITMUS_CONFIG_TEMPLATES` | `config.templates` | Templates subdir |
| `LITMUS_MIMIR_ADDRESS` | `mimir.address` | Mimir URL |
| `LITMUS_MIMIR_TENANT_ID` | `mimir.tenant_id` | Tenant ID |
| `LITMUS_MIMIR_API_KEY` | `mimir.api_key` | API key |
| `LITMUS_REGRESSION_DIRECTORY` | `regression.directory` | Baseline dir |
| `LITMUS_TESTS_DIRECTORY` | `tests.directory` | Tests directory |

## Template File Handling

### Supported Extensions

Litmus reads template files with extensions: `.tpl`, `.tmpl`

### Template Resolution

1. Parse alertmanager config's `templates:` field
2. Resolve each template against the templates directory
3. **Error** if referenced template not found
4. **Ignore** extra template files not referenced in config

Example:
```
config/alertmanager.yml:
  templates:
    - slack.tpl
    - email.tpl

config/templates/:
  ├── slack.tpl    ✓ uploaded
  ├── email.tpl    ✓ uploaded
  └── unused.tmpl  ✗ ignored
```

## Implementation Notes

### Viper Configuration

```go
func LoadConfig() (*Viper, error) {
    v := viper.New()
    
    // Defaults
    v.SetDefault("config.directory", "config")
    v.SetDefault("config.file", "alertmanager.yml")
    v.SetDefault("config.templates", "templates/")
    v.SetDefault("regression.directory", "regressions")
    v.SetDefault("regression.max_samples", 5)
    v.SetDefault("tests.directory", "tests")
    
    // Config file
    v.SetConfigName(".litmus")
    v.AddConfigPath(".")
    v.AddConfigPath("config/")
    
    // Environment variables
    v.SetEnvPrefix("LITMUS")
    v.AutomaticEnv()
    
    // Try to read config, auto-create if missing
    if err := v.ReadInConfig(); err != nil {
        if _, ok := err.(viper.ConfigFileNotFoundError); ok {
            return v, autoCreateConfig(v)
        }
        return nil, err
    }
    
    return v, nil
}
```

### Related Commands

| Command | Config Used |
|---------|-------------|
| `litmus init` | Creates default structure |
| `litmus snapshot` | `config.file`, `regression.*` |
| `litmus check` | `config.*`, `regression.*`, `tests.*` |
| `litmus sync` | `config.*`, `mimir.*` |

### Files Modified

| File | Change |
|------|--------|
| `cmd/litmus/main.go` | Add config loader |
| `internal/config/config.go` | New: Viper loader |
| `internal/templates/files/litmus.yaml` | New format |
| `cmd/litmus/init.go` | Create subdirs |
| `cmd/litmus/sync.go` | Use resolved paths |
| `docs/cli/sync_command.md` | Update sync docs |
