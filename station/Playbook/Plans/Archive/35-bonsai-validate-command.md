---
tags: [plan, feature, v0.4]
description: New `bonsai validate` CLI command — proactive read-only check for ability state mismatches (orphaned registrations, stale lock entries, untracked custom files, invalid frontmatter, missing custom_items metadata, wrong category placement). Justifies v0.4.0 minor bump.
---

# Plan 35 — `bonsai validate` Command

**Tier:** 1
**Status:** Complete
**Agent:** general-purpose (single worktree)

## Goal

Add a `bonsai validate` command that audits a Bonsai project's ability state and reports any inconsistencies between `.bonsai.yaml`, `.bonsai-lock.yaml`, and the workspace's `agent/<Category>/` directories. Read-only — fixes happen via `bonsai update`. Two output modes: human-readable tables (default), `--json` for CI/agents. Exit code 0 = clean, 1 = issues found, 2 = internal error.

## Background

Plan 34 added reactive recovery for orphaned registrations in `bonsai update`. `validate` is the proactive companion: a user (or their agent) can ask "is my workspace consistent?" without triggering the re-render pipeline. Becomes the headline new feature for v0.4.0.

## Detection categories

The command detects six classes of issue. Each maps to an `Issue.Category` enum value.

| Category | Severity | Trigger |
|----------|----------|---------|
| `orphaned_registration` | error | Name in `installed.<Cat>`, file exists on disk, but `relPath` not in lock OR `custom_items[name]` missing/empty. (Plan 34's repro condition.) |
| `missing_file` | error | Name in `installed.<Cat>` but file does not exist at expected `agent/<Dir>/<name>.<ext>`. |
| `stale_lock_entry` | warning | `lock.Files[relPath]` references `custom:<type>s/<name>` but file does not exist on disk. |
| `untracked_custom_file` | warning | File on disk under `agent/<Dir>/`, valid frontmatter, but name not in `installed.<Cat>` and `relPath` not in lock. (i.e., user dropped a file but never ran `bonsai update`.) |
| `invalid_frontmatter` | error | File on disk under `agent/<Dir>/`, but frontmatter missing/malformed/missing required fields. Reuses `ScanCustomFiles` validation rules. |
| `wrong_extension_in_category` | warning | `.md` file in `agent/Sensors/` OR `.sh` file in `agent/Skills/` / `Workflows/` / `Protocols/` / `Routines/`. (Top-level only — subdirs ignored, same as scan.) |

## Steps

### Step 1 — New package `internal/validate`

Create `internal/validate/validate.go` with:

```go
package validate

import (
    "github.com/LastStep/Bonsai/internal/catalog"
    "github.com/LastStep/Bonsai/internal/config"
)

type Severity string

const (
    SeverityError   Severity = "error"
    SeverityWarning Severity = "warning"
)

type Category string

const (
    CategoryOrphanedRegistration  Category = "orphaned_registration"
    CategoryMissingFile           Category = "missing_file"
    CategoryStaleLockEntry        Category = "stale_lock_entry"
    CategoryUntrackedCustomFile   Category = "untracked_custom_file"
    CategoryInvalidFrontmatter    Category = "invalid_frontmatter"
    CategoryWrongExtension        Category = "wrong_extension_in_category"
)

type Issue struct {
    Category  Category `json:"category"`
    Severity  Severity `json:"severity"`
    AgentName string   `json:"agent,omitempty"`
    AbilityType string `json:"ability_type,omitempty"` // skill|workflow|protocol|sensor|routine
    Name      string   `json:"name,omitempty"`
    Path      string   `json:"path,omitempty"`
    Detail    string   `json:"detail"`
}

type Report struct {
    Issues       []Issue `json:"issues"`
    AgentsScanned []string `json:"agents_scanned"`
}

func (r *Report) HasErrors() bool { /* any Severity == Error */ }
func (r *Report) HasIssues() bool { return len(r.Issues) > 0 }

// Run audits the project at projectRoot using cfg + cat + lock. Optional
// agentFilter restricts checks to a single installed agent.
func Run(projectRoot string, cfg *config.ProjectConfig, cat *catalog.Catalog, lock *config.LockFile, agentFilter string) (*Report, error)
```

Implementation outline:

1. Build sorted agent name slice (or filter to one if `agentFilter != ""`).
2. For each agent:
   a. For each ability type (skill/workflow/protocol/sensor/routine), category dir + extension already known from `generate.ScanCustomFiles` constants — reuse the same `categoryDef` shape (extract to `internal/generate` exported helper, OR duplicate locally as a small map; duplication acceptable here — 5-line map).
   b. **Orphan/missing check:** for each name in `installed.<Cat>`, check `agent/<Dir>/<name>.<ext>` exists. If missing → `missing_file` error. If exists but `relPath` not in `lock.Files` OR `custom_items[name]` missing or `Description == ""` → `orphaned_registration` error.
   c. **Stale lock check:** for each `lock.Files[relPath]` with `Source == "custom:<type>s/<name>"`, check file exists. If not → `stale_lock_entry` warning.
   d. **Untracked + invalid + wrong-extension scan:** read `agent/<Dir>/` entries (top-level only, same as `ScanCustomFiles`). For each entry:
      - If extension matches `info.ext` AND name not in `installed.<Cat>` AND `relPath` not in `lock.Files`: parse frontmatter. If frontmatter missing/invalid → `invalid_frontmatter` error. Else → `untracked_custom_file` warning.
      - If extension is `.md` in Sensors dir OR `.sh` in Skills/Workflows/Protocols/Routines dir → `wrong_extension_in_category` warning.
3. Return `Report` with all issues + sorted `AgentsScanned`.

The Validate package depends only on `internal/config`, `internal/catalog`, `internal/generate` (for `ParseFrontmatter`). No TUI dependency — keeps it CI-safe.

### Step 2 — Tests `internal/validate/validate_test.go`

Table-driven test with one subtest per category. Each sets up a temp dir, writes `.bonsai.yaml` + `.bonsai-lock.yaml` + agent files matching one specific failure mode, calls `Run`, asserts the right `Issue.Category` appears with the right `Name` / `Path` / `AgentName`. Plus:

- `TestRun_CleanProject` — fully consistent project → empty `Report.Issues`, `HasIssues()` false.
- `TestRun_AgentFilter` — two agents, filter to one, only that agent's issues returned.
- `TestRun_MultipleCategoriesAtOnce` — single project triggers 3 different categories — all reported.

### Step 3 — New command `cmd/validate.go`

```go
package cmd

import (
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"

    "github.com/spf13/cobra"

    "github.com/LastStep/Bonsai/internal/config"
    "github.com/LastStep/Bonsai/internal/validate"
)

func init() {
    rootCmd.AddCommand(validateCmd)
    validateCmd.Flags().StringP("agent", "a", "", "Restrict check to a single installed agent")
    validateCmd.Flags().Bool("json", false, "Emit issues as JSON (one Report object) — non-interactive")
}

var validateCmd = &cobra.Command{
    Use:   "validate",
    Short: "Audit ability state — detect orphaned registrations, stale lock entries, untracked custom files, and frontmatter problems.",
    Long:  "Read-only audit. Checks .bonsai.yaml + .bonsai-lock.yaml + agent/ workspace for inconsistencies. Exit 0 = clean, 1 = issues found, 2 = internal error. Run `bonsai update` to fix orphans/untracked. JSON output via --json.",
    RunE:  runValidate,
}

func runValidate(cmd *cobra.Command, args []string) error {
    cwd := mustCwd()
    cfg, err := requireConfig(filepath.Join(cwd, configFile))
    if err != nil {
        os.Exit(2)
    }
    cat := loadCatalog()
    lock, _ := config.LoadLockFile(cwd)
    if lock == nil {
        lock = config.NewLockFile()
    }

    agentFilter, _ := cmd.Flags().GetString("agent")
    jsonOut, _ := cmd.Flags().GetBool("json")

    report, err := validate.Run(cwd, cfg, cat, lock, agentFilter)
    if err != nil {
        if jsonOut {
            fmt.Fprintf(os.Stderr, "validate error: %s\n", err.Error())
        }
        os.Exit(2)
    }

    if jsonOut {
        return renderValidateJSON(report)
    }
    renderValidateText(report)

    if report.HasIssues() {
        os.Exit(1)
    }
    return nil
}
```

JSON renderer: marshal `*Report` with indent 2, print to stdout, return nil.

Text renderer (`renderValidateText`):
- If no issues: print `tui.Success("No issues found.")` and number of agents scanned.
- Else: group issues by `AgentName`. For each agent, print `tui.SectionHeader("agent: <name>")`, then a table with columns: `severity`, `category`, `ability_type`, `name`, `path`, `detail`. Use `tui.CatalogTable([]string{"Severity", "Category", "Type", "Name", "Path", "Detail"}, rows)`. Footer line: total issue count, broken into `N error(s), M warning(s)`.

Direct exit-code handling via `os.Exit` (matches existing patterns). The cobra `RunE` returning nil is fine for the success path; we exit before returning on error/issue paths.

### Step 4 — Tests `cmd/validate_test.go`

Lightweight smoke test. Build the cobra command with a `RunE` override OR call `runValidate` indirectly by invoking the binary in a sub-process? Existing `cmd/list_test.go` patterns — read it first; if existing tests use `os.Exec` smoke pattern, follow that. Otherwise:

- `TestValidate_CleanProject` — temp dir setup, run `validate.Run`, verify report empty.
- `TestValidate_JSONOutput` — temp dir with one orphan, capture stdout, parse JSON, assert structure.

Bias toward placing exhaustive coverage in `internal/validate/validate_test.go`; keep `cmd/validate_test.go` small (1-2 smoke tests).

### Step 5 — Documentation

Update `catalog/skills/bonsai-model/bonsai-model.md`:

1. In the command table (around line 90), add row:
   ```
   | `bonsai validate` | Read-only audit — detect orphaned registrations, stale lock entries, untracked custom files, frontmatter problems. JSON via `--json`. |
   ```

2. In the troubleshooting/state-management section, add reference: when in doubt about workspace state, run `bonsai validate` first.

Update `catalog/skills/workspace-guide/workspace-guide.md.tmpl` similarly — add `bonsai validate` to the command table.

### Step 6 — Build registration check

The new `cmd/validate.go` registers itself via `init()` calling `rootCmd.AddCommand(validateCmd)`, so no edits to `cmd/root.go` are needed. Confirm by `bonsai --help` listing `validate` after build.

## Security

> [!warning]
> Refer to [SecurityStandards.md](../../Standards/SecurityStandards.md) for all security requirements.

Specifics:
- `validate` is read-only — no file mutations. No lock writes, no config writes.
- All file paths derived from `cfg.Agents[name].Workspace` + category dir + entry name. Workspace already validated via `wsvalidate` (Plan 32). No path-traversal surface beyond what `Load`/`requireConfig` already guards.
- JSON output uses `encoding/json` — no template injection. Issue `Detail` strings are constants from `internal/validate/validate.go`. `Path`, `Name`, `AgentName` come from filesystem reads / config; treat as untrusted in human renderer (use `%s` formatting, not `Sprintf` with format chars).
- No new external dependencies.

## Verification

- [ ] `make build` succeeds.
- [ ] `go test ./internal/validate/... ./cmd/...` passes.
- [ ] `./bonsai --help` lists `validate` as a subcommand with the expected short description.
- [ ] `./bonsai validate --help` shows `--agent` and `--json` flags.
- [ ] Sandbox scenario `clean`:
  - Set up temp project: `bonsai init` (or hand-built equivalent), drop a valid custom skill, run `bonsai update` to register it.
  - `bonsai validate` → exit 0, prints "No issues found."
- [ ] Sandbox scenario `orphan`:
  - Temp project with `.bonsai.yaml` listing `skills: [foo]`, file `agent/Skills/foo.md` valid frontmatter, NO lock entry, NO `custom_items[foo]`.
  - `bonsai validate` → exit 1, reports `orphaned_registration` error for `foo`.
  - `bonsai validate --json` → JSON includes `{"category":"orphaned_registration","severity":"error","name":"foo",...}`.
- [ ] Sandbox scenario `missing_file`:
  - Temp project with `.bonsai.yaml` listing `skills: [ghost]`, no file on disk.
  - `bonsai validate` → exit 1, reports `missing_file` error.
- [ ] Sandbox scenario `untracked`:
  - Temp project with file `agent/Skills/new.md` (valid frontmatter), name not in `.bonsai.yaml`.
  - `bonsai validate` → exit 1, reports `untracked_custom_file` warning.
- [ ] Sandbox scenario `invalid_frontmatter`:
  - Temp project with file `agent/Skills/bad.md` (no frontmatter).
  - `bonsai validate` → exit 1, reports `invalid_frontmatter` error.
- [ ] Sandbox scenario `wrong_ext`:
  - Temp project with `agent/Sensors/foo.md` (markdown in sensors dir).
  - `bonsai validate` → exit 1, reports `wrong_extension_in_category` warning.
- [ ] `--agent <name>` flag filters output to that agent only (verify via 2-agent project).

## Out of Scope

- `--fix` flag — fixes happen via `bonsai update`. Adding `--fix` would duplicate the update flow.
- Validating catalog-tracked (non-custom) files for hash drift — that's `bonsai update`'s conflict resolver job.
- Walking subdirectories under `agent/<Category>/` — Plan 34 explicitly kept top-level-only behaviour.
- Case-insensitive directory matching.
- TUI / cinematic flow for `validate` — text + JSON output is sufficient; no progressive disclosure needed for a read-only audit.
- Integration into `bonsai update` (e.g., auto-run validate before update) — keep commands orthogonal.
