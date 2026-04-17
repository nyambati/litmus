# CLI Design: `litmus sync` Command

> **Status**: Brainstorming  
> **Related**: `mimirtool alertmanager load` analysis

## 1. Purpose

The `sync` command enables litmus to process, validate, and push Alertmanager configurations to Grafana Mimir. This bridges local validation with remote state management.

## 2. Motivation

Mimir provides the `mimirtool alertmanager load` command to upload configs and templates:

```bash
mimirtool alertmanager load alertmanager.yaml templates/*.tpl \
  --address=https://mimir.example.com \
  --id=anonymous \
  --key=<api-key>
```

This command:
1. Reads the Alertmanager config YAML
2. Reads template files (strips paths, uses only filenames)
3. POSTs to `/api/v1/alerts` with:
   ```json
   {
     "alertmanager_config": "<yaml>",
     "template_files": {
       "my-template.tpl": "<content>"
     }
   }
   ```

Litmus needs equivalent functionality but with:
- Environment variable substitution before upload
- Local validation before push (fail-fast)
- Unified CLI experience with existing litmus commands

## 3. Command Specification

### Usage

```bash
litmus sync [flags]

# Examples (using config from .litmus.yaml)
litmus sync
litmus sync --skip-validate

# Override config via CLI or env
litmus sync --address=https://mimir.example.com --tenant-id=anonymous
LITMUS_MIMIR_ADDRESS=https://mimir.example.com litmus sync
```

### Arguments

The sync command uses configuration from `.litmus.yaml`:

| Config Key | Description |
|------------|-------------|
| `config.directory` | Directory containing alertmanager.yml |
| `config.file` | Alertmanager config filename |
| `config.templates` | Templates subdirectory |
| `mimir.address` | Mimir URL |
| `mimir.tenant_id` | Tenant ID |
| `mimir.api_key` | API key |

### Flags

| Flag | Env Variable | Description |
|------|--------------|-------------|
| `--address` | `MIMIR_ADDRESS` | Override Mimir URL |
| `--tenant-id` | `MIMIR_TENANT_ID` | Override tenant ID |
| `--api-key` | `MIMIR_API_KEY` | Override API key |
| `--skip-validate` | - | Skip validation before push |
| `--allow-missing-env` | - | Allow undefined env vars |

### Template Resolution

Templates are resolved from the configured directory:

1. Config directory: `{config.directory}` → `config/`
2. Templates subdir: `{config.templates}` → `templates/`
3. Full path: `config/templates/`

The command only uploads templates **explicitly listed** in the config's `templates:` field:

1. Parse config's `templates:` list (e.g., `templates: ["slack.tpl", "email.tpl"]`)
2. For each template, look up file in `config/templates/`
3. **Error** if any referenced template not found in dir
4. **Ignore** extra `.tpl`/`.tmpl` files in dir not referenced in config

Example:
```
config/alertmanager.yml:
  templates:
    - slack.tpl
    - email.tpl

config/templates/:
  ├── slack.tpl    ✓ uploaded
  ├── email.tpl    ✓ uploaded
  └── extra.tpl    ✗ ignored (not in config)
```

### Flags

| Flag | Env Variable | Required | Default | Description |
|------|--------------|----------|---------|-------------|
| `--address` | `MIMIR_ADDRESS` | Yes | - | Mimir URL |
| `--tenant-id` | `MIMIR_TENANT_ID` | Yes | - | Tenant ID |
| `--api-key` | `MIMIR_API_KEY` | No | - | API key for auth |
| `--skip-validate` | - | No | `false` | Skip validation before push |
| `--allow-missing-env` | - | No | `false` | Allow undefined env vars (empty string) |

### Auth Precedence

1. CLI flags (highest priority)
2. Environment variables (`LITMUS_MIMIR_*`)
3. `.litmus.yaml` values (with `{{ env "VAR" }}` substitution)
4. Error if mimir.address not available

## 4. Pipeline: Process → Validate → Push

The sync command reads configuration from `.litmus.yaml` and resolves paths using the config directory structure.

### Step 1: Process (Environment Variable Substitution)

1. Load `.litmus.yaml` with env substitution for mimir credentials
2. Load alertmanager config from `{config.directory}/{config.file}`
3. Parse config for `{{ env "VAR" }}` patterns and substitute with `os.Getenv("VAR")`

**Syntax**: Go text/template style
```yaml
global:
  smtp_smarthost: '{{ env "SMTP_HOST" }}:587'
  smtp_from: '{{ env "SMTP_FROM" }}'
```

**Behavior**:
- Undefined env var → error (unless `--allow-missing-env`)
- Template is executed in-memory, not written to disk
- Processed config passed to validation step

### Step 2: Validate

Load processed YAML into `github.com/prometheus/alertmanager/config.Config`.

Run existing sanity checks:
- **Shadowed route detection**: Warn on routes that will never match
- **Orphan receiver detection**: Warn on receivers not referenced
- **Inhibition cycle detection**: Warn on circular inhibit rules

**Behavior**:
- Validation fails → print issues, exit with code 3
- Exit unless `--skip-validate` is set

### Step 3: Push

1. Parse config's `templates:` field for list of required templates
2. Resolve templates directory: `{config.directory}/{config.templates}` = `config/templates/`
3. Read only those templates (strip paths, use filename)
4. POST to `{mimir.address}/api/v1/alerts`

**Request**:
```
POST /api/v1/alerts
Content-Type: application/json
X-Scope-OrgID: {tenant-id}
X-Mimir-API-Key: {api-key}
```

**Body**:
```json
{
  "alertmanager_config": "<processed_yaml>",
  "template_files": {
    "my-template.tpl": "<content>"
  }
}
```

**Response**:
- 201 Created → success
- 4xx/5xx → error with body

## 5. Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Validation failure |
| 3 | HTTP/API error |
| 4 | Missing required argument |

## 6. Implementation Notes

### New Files

| File | Purpose |
|------|---------|
| `cmd/litmus/sync.go` | Command implementation using config |
| `cmd/litmus/sync_test.go` | Table-driven tests |
| `internal/config/config.go` | Viper config loader with env substitution |
| `internal/processor/env.go` | Environment substitution |

### Dependencies

- `github.com/spf13/viper` - config management
- `text/template` (stdlib) - for env substitution
- `net/http` (stdlib) - for API calls
- `github.com/prometheus/alertmanager/config` - already used

## 7. Related Documentation

- [Configuration](configuration.md) - `.litmus.yaml` schema and path resolution

### Mimir API Reference

- **Endpoint**: `POST /api/v1/alerts`
- **Requirement**: Template filenames must be valid (no path separators)
- **Docs**: https://grafana.com/docs/mimir/latest/operators-guide/reference-http-api/#set-alertmanager-configuration

## 7. Open Questions

1. **Dry-run mode**: Should we support `--dry-run` to validate without pushing?
2. **Template validation**: Should we validate Go template syntax in template files?
3. **Config backup**: Should we fetch existing config from Mimir before overwriting?
4. **Interactive mode**: Should we prompt for confirmation before push?

## 8. Related Commands Comparison

| Command | Local | Remote | Env Substitution |
|---------|-------|--------|------------------|
| `litmus check` | ✓ | - | - |
| `litmus snapshot` | ✓ | - | - |
| `litmus sync` | ✓ | ✓ | ✓ |
| `mimirtool get` | - | ✓ | - |
| `mimirtool load` | - | ✓ | - |
