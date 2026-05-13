---
tags: [plan, active]
description: Plan 39 — Non-interactive `bonsai init` / `bonsai add` flags. Tier-2 feature. Ships v0.5.0.
---

# Plan 39 — Non-interactive `bonsai init` / `bonsai add`

**Tier:** 2
**Status:** Draft
**Agent:** general-purpose (single agent, sequential phases)

## Goal

`bonsai init` and `bonsai add` accept `--non-interactive` + `--from-config <path>` flags. Under those flags the cinematic TUI is fully bypassed, all answers are read from a `.bonsai.yaml`-shaped YAML file (or defaulted), conflicts are skipped (never overwritten), and per-file outcomes are emitted as one JSON object per line on stdout. Validation errors exit non-zero with a stderr message — no interactive fallback. Ships as **v0.5.0** (minor — new public flags).

## Context

Plan 38 P2 (Bonsai-Eval rung-3 solver, `~/ZenGarden/Bonsai-Eval/bonsai_eval/solvers/rungs.py:135`) needs to materialise a fresh `station/` workspace inside per-scenario sandboxes from a fixture YAML — without TUI prompts. Today both commands force interactive Huh+BubbleTea flows, which is unusable from a Python `subprocess.run()` call. Backlog P0 entry was promoted from the Plan 38 dispatch review.

Locked decisions (recorded in [`Logs/KeyDecisionLog.md`](../../../Logs/KeyDecisionLog.md) on plan acceptance):

1. **Q1=A — `--from-config` shape mirrors `.bonsai.yaml`.** Round-trip: `bonsai init --from-config station/.bonsai.yaml` reproduces an existing workspace. `.bonsai.yaml` already carries every field needed (`project_name`, `description`, `docs_path`, `scaffolding`, `agents[]` with all five ability lists). No schema additions.
2. **Q2=C — no `--target-dir` flag.** Caller manages cwd (Python's `subprocess.run(..., cwd=...)`, shell `cd`). Smallest CLI surface.
3. **Q3=Lenient — only one `agents:` entry is required.** Defaults applied:
   - `project_name` → `filepath.Base(cwd)` if missing
   - `description` → `""`
   - `docs_path` → `"station/"`
   - `scaffolding` → required-only items from catalog
   - per-agent ability lists → `agent.yaml` `defaults` for that agent type
   - per-agent `workspace` → `cfg.DocsPath` for tech-lead, otherwise `<agentType>/`
4. **Q4=JSONL stdout.** One event per line. Schema: `{"event":"<type>","path":"<relpath>","action":"<created|updated|unchanged|skipped|conflict>","source":"<source>"}`. Plus a final `{"event":"summary","created":N,"updated":N,"unchanged":N,"skipped":N,"conflicts":N}` line.
5. **Conflict resolution forced to skip.** Under `--non-interactive`, `wr.HasConflicts()` does not prompt — every conflict file is left untouched and emitted as `{"event":"file","path":"...","action":"conflict"}`. No `.bak`, no overwrite.
6. **Validation errors exit ≠ 0 with stderr message.** Exit codes:
   - `0` — success
   - `2` — invalid input config (missing required field, malformed YAML, shell-metachar trip in `config.Validate`)
   - `3` — runtime error (filesystem, generator panic)
   - `4` — `.bonsai.yaml` already exists for `init`, or no `.bonsai.yaml` for `add`
7. **TUI fully suppressed.** No `harness.Run`, no AltScreen, no LipGloss styled output to stdout. `tui.Warning` calls in the non-interactive path route to stderr as plain text. The interactive path is unchanged.
8. **v0.5.0** — minor bump for new public flags.

## Phases

> Single agent works phases sequentially. Each phase is one commit on the same branch. CI runs after each push.

### Phase A — Foundation: `internal/nonint` package

**Files (new):**
- `internal/nonint/nonint.go` — public API surface
- `internal/nonint/config.go` — input YAML loading + default resolution + validation
- `internal/nonint/events.go` — JSON Lines event emitter
- `internal/nonint/runner.go` — `RunInit` / `RunAdd` orchestrators
- `internal/nonint/config_test.go`
- `internal/nonint/events_test.go`
- `internal/nonint/runner_test.go`

**API:**

```go
// Package nonint drives bonsai init / bonsai add without TUI.
package nonint

// Event is one JSONL record. Marshalled with omitempty so per-event payloads stay tight.
type Event struct {
    Event   string `json:"event"`             // "file" | "summary" | "warning"
    Path    string `json:"path,omitempty"`
    Action  string `json:"action,omitempty"`  // created | updated | unchanged | skipped | conflict
    Source  string `json:"source,omitempty"`
    // summary-event fields
    Created    int `json:"created,omitempty"`
    Updated    int `json:"updated,omitempty"`
    Unchanged  int `json:"unchanged,omitempty"`
    Skipped    int `json:"skipped,omitempty"`
    Conflicts  int `json:"conflicts,omitempty"`
    // warning fields
    Message string `json:"message,omitempty"`
}

// Emit writes one event as a single JSON line + newline. Errors propagate so
// the caller can exit ≠ 0 if stdout is closed.
func Emit(w io.Writer, e Event) error

// LoadConfig reads <path> as YAML, applies defaults from cwd + cat, validates,
// returns a fully-resolved *config.ProjectConfig ready to pass to runners.
// Errors:
//   - "from-config: read <path>: ..." on I/O failure
//   - "from-config: parse YAML: ..." on bad YAML
//   - "from-config: missing required field 'agents' (need at least one entry)" on lenient-required miss
//   - validation errors from config.Validate() (shell-metachars, invalid workspace)
func LoadConfig(path, cwd string, cat *catalog.Catalog) (*config.ProjectConfig, error)

// RunInit performs a full init from cfg (already defaulted + validated).
// Emits one "file" event per WriteResult entry, then one "summary" event.
// Returns (exitCode, error). exitCode is 0 on success, 3 on runtime error,
// 4 if .bonsai.yaml already exists at cwd.
func RunInit(cwd, configPath string, cfg *config.ProjectConfig, cat *catalog.Catalog, w io.Writer) (int, error)

// RunAdd appends agents/abilities to an existing project. cfg here is the
// INPUT config (overlay) — runner loads the existing .bonsai.yaml, merges
// new agents + new abilities (skip-if-installed semantics, no overwrite).
// Returns (exitCode, error). exitCode is 0 on success, 3 on runtime error,
// 4 if .bonsai.yaml does not exist at cwd.
func RunAdd(cwd, configPath string, overlay *config.ProjectConfig, cat *catalog.Catalog, w io.Writer) (int, error)
```

**Implementation notes:**

1. `LoadConfig` reads the YAML, then walks defaults:
   - if `cfg.ProjectName == ""` → `filepath.Base(cwd)`
   - if `cfg.DocsPath == ""` → `"station/"`
   - if `cfg.Scaffolding == nil` → for each `cat.Scaffolding` item where `Required`, append `item.Name`
   - for each `agent` in `cfg.Agents`:
     - if `agent.Workspace == ""` → tech-lead: `cfg.DocsPath`; else: `agentType + "/"` (then `wsvalidate.Normalise`)
     - if every ability list is nil → use `agentDef.Defaults` (mirrors interactive `BranchesStage` initial selection)
     - call `generate.EnsureRoutineCheckSensor(agent)` so routine-check is wired iff routines present
   - finally `cfg.Validate()` (existing shell-metachar + wsvalidate scan)
2. `RunInit` mirrors `cmd.runInit`'s `buildGenerateAction` body **exactly**: `cfg.Save(configPath)` → `generate.Scaffolding` → `generate.AgentWorkspace` → `generate.PathScopedRules` → `generate.WorkflowSkills` → `generate.SettingsJSON` → `generate.WriteCatalogSnapshot`. Each is fed the same `*WriteResult` so the post-walk emit sees every file. Must early-exit if `os.Stat(configPath) == nil` returns existence (mirror interactive behaviour).
3. `RunAdd` mirrors `cmd.runAdd`'s `buildAddGrowAction`: per agent in overlay's `cfg.Agents`, decide new-agent (not in existing cfg) vs add-items (already present, append the diff of ability names not already installed), then call the same generator chain — `AgentWorkspace` + `PathScopedRulesForAgent` + `WorkflowSkillsForAgent` + `SettingsJSONForAgent` + `RefreshPeerAwareness` + `WriteCatalogSnapshot`. Tech-lead-required guard applies (non-tech-lead overlay agent without a tech-lead present → error). Unknown-agent (not in catalog) → error.
4. After all writes, walk `wr.Files` and `Emit` one event per file:
   - `ActionCreated` → `"action":"created"`
   - `ActionUpdated`, `ActionForced` → `"action":"updated"`
   - `ActionUnchanged` → `"action":"unchanged"`
   - `ActionSkipped` → `"action":"skipped"`
   - `ActionConflict` → `"action":"conflict"` (file untouched on disk — non-interactive never calls `ForceSelected`/`ForceConflicts`)
5. Emit final summary event from `wr.Summary()`.
6. `lock.Save(cwd)` runs after JSONL flush; on error emit `{"event":"warning","message":"..."}` to **stderr** (not stdout) and return exit 0 (lock-save failure is not fatal — matches interactive behaviour).

**Tests:**

- `TestLoadConfig_Minimal_AppliesAllDefaults` — input has only `agents: { tech-lead: {} }`; verify defaulted ProjectName, DocsPath, Scaffolding (matches required scaffolding from catalog), and agent.Workspace + abilities.
- `TestLoadConfig_RoundTrip` — write a generated `.bonsai.yaml` from interactive init, load it, verify equality after defaulting (no drift).
- `TestLoadConfig_MissingAgents_Errors` — empty `agents:` → error message includes "missing required field 'agents'".
- `TestLoadConfig_InvalidWorkspace_Errors` — workspace `../etc/passwd` → wsvalidate-driven error.
- `TestLoadConfig_ShellMetachar_Errors` — project_name `proj"$x` → error.
- `TestEmit_JSONLShape` — emit one event, parse line back through `json.Unmarshal`, verify shape.
- `TestEmit_NoStyling` — assert output is byte-identical to `json.Marshal(e) + "\n"` (no LipGloss codes).
- `TestRunInit_Smoke` — tmp dir, minimal cfg, run, assert: `.bonsai.yaml` present, station tree present, JSONL parseable, summary line emitted.
- `TestRunInit_ConfigExists_ExitCode4` — pre-create `.bonsai.yaml`, run, assert exitCode==4 + stderr message.
- `TestRunAdd_NewAgent_Smoke` — init tech-lead via RunInit, then RunAdd a backend agent overlay, verify backend workspace materialises.
- `TestRunAdd_AddItems_Smoke` — RunInit tech-lead with default abilities, then RunAdd same tech-lead with extra skill in overlay, verify only the new skill is appended (no duplicates).
- `TestRunAdd_TechLeadRequired_Errors` — RunInit skipped, RunAdd of backend → error.
- `TestRunAdd_UnknownAgent_Errors` — overlay names agent type `does-not-exist` → error.
- `TestRunInit_ConflictEmittedNotForced` — pre-place a user-edited file at a path the generator targets, run, verify event has `action:"conflict"` and the file content on disk is unchanged.

### Phase B — Wire `bonsai init --non-interactive --from-config <path>`

**Files (modify):**
- `cmd/init.go` — register both flags via `cobra.Command.Flags()`.
- `cmd/init_flow.go` (the existing one with `runInit`) — at the top of `runInit`, if both flags set, branch into `nonint.RunInit` and return its exit code via `os.Exit(...)` (or via the cobra `RunE` error wrapping pattern). If only one flag is set, error: "both --non-interactive and --from-config are required together".

**Files (new):**
- `cmd/init_nonint_test.go` — end-to-end test invoking the cobra command with flags set, asserting JSONL output + filesystem state.

**Pattern:**

```go
// cmd/init.go
var (
    initNonInteractive bool
    initFromConfig     string
)

func init() {
    rootCmd.AddCommand(initCmd)
    initCmd.Flags().BoolVar(&initNonInteractive, "non-interactive", false, "Skip TUI prompts; read answers from --from-config")
    initCmd.Flags().StringVar(&initFromConfig, "from-config", "", "Path to YAML config (.bonsai.yaml shape)")
}
```

```go
// cmd/init_flow.go runInit head
if initNonInteractive || initFromConfig != "" {
    if !initNonInteractive || initFromConfig == "" {
        return fmt.Errorf("--non-interactive and --from-config must be set together")
    }
    cwd := mustCwd()
    cat := loadCatalog()
    cfg, err := nonint.LoadConfig(initFromConfig, cwd, cat)
    if err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(2)
    }
    code, runErr := nonint.RunInit(cwd, filepath.Join(cwd, configFile), cfg, cat, os.Stdout)
    if runErr != nil {
        fmt.Fprintln(os.Stderr, runErr)
    }
    if code != 0 {
        os.Exit(code)
    }
    return nil
}
// existing interactive runInit body unchanged below
```

### Phase C — Wire `bonsai add --non-interactive --from-config <path>`

**Files (modify):**
- `cmd/add.go` — register flags + branch identical pattern at top of `runAdd`.

**Files (new):**
- `cmd/add_nonint_test.go`

**Notes:**
- Add must accept overlay configs that name multiple agents and/or extra abilities for the existing tech-lead. Skip-if-installed semantics are handled inside `nonint.RunAdd`.

### Phase D — Polish + release prep

**Files (modify):**
- `CHANGELOG.md` — new `## v0.5.0 — 2026-05-XX` block. Bullets:
  - Added `--non-interactive` + `--from-config` to `bonsai init` and `bonsai add`. Enables Bonsai-Eval rung-3 solver to materialise stations from fixture YAML without TUI.
  - JSON Lines progress output on stdout; validation errors exit ≠ 0 with stderr message.
  - Conflict resolution under `--non-interactive` is forced to skip (never overwrite).
- `README.md` — short usage block under "Usage" or new "Scripted use" section: 4-line example.
- `cmd/init.go` + `cmd/add.go` — `--help` text for new flags (already covered by `Flags().BoolVar` description above; verify cobra renders both).
- Version constant: leave as `dev` — release tagger picks up the v0.5.0 ldflags injection at `goreleaser` time. (Plan 36 already wired this.)

**Verification (manual, by tech-lead post-merge):**
- `mkdir /tmp/bonsai-nonint-test && cd /tmp/bonsai-nonint-test`
- Write minimal fixture: `cat > cfg.yaml <<EOF
project_name: t
agents:
  tech-lead: {}
EOF`
- `bonsai init --non-interactive --from-config cfg.yaml | jq -c .` — every line parseable.
- Verify `.bonsai.yaml` written, station tree present, summary line includes correct counts.
- `bonsai init --non-interactive --from-config /dev/null` in fresh dir — exit 2 + stderr.
- `bonsai init --non-interactive` (no --from-config) — exit ≠ 0 + stderr "must be set together".
- Run `./bonsai validate` against the produced workspace — clean exit.
- Re-run `bonsai init --non-interactive --from-config cfg.yaml` in same dir — exit 4.

## Dependencies

- `internal/config` — extending behaviour but no schema additions.
- `internal/catalog` — read-only.
- `internal/generate` — read-only (reuses existing public functions).
- `internal/wsvalidate` — read-only.
- No new third-party deps. `encoding/json` (stdlib) covers JSONL.

## Security

> [!warning]
> Refer to [Standards/SecurityStandards.md](../../Standards/SecurityStandards.md) for all security requirements.

Specific concerns for this plan:

1. **Path traversal via `--from-config` workspace fields.** Input YAML's `agent.workspace` and `docs_path` are user-controlled; both already flow through `wsvalidate.Normalise` + `InvalidReason` in `config.Validate` — `nonint.LoadConfig` MUST call `cfg.Validate()` after defaulting and propagate any error as exit 2. Do not skip validation under any flag.
2. **Shell metacharacter injection via `project_name` / agent names.** These strings are templated into shell scripts (sensor templates, status-bar). Existing `forbiddenShellChars` scan in `config.Validate` covers this; same enforcement path.
3. **File overwrite via crafted overlay.** Conflict resolution is forced to **skip** under `--non-interactive` — `wr.HasConflicts()` MUST NOT trigger any `ForceSelected`/`ForceConflicts` call in the non-interactive path. Defensive: `nonint.RunInit` and `nonint.RunAdd` must never call those two helpers; instead, walk `wr.Files`, emit conflict events, leave files untouched. Add a unit test that pre-places a user-edited file and asserts byte-identical content after run.
4. **`--from-config` reading arbitrary paths.** YAML loading uses `os.ReadFile(path)` — same as the existing `config.Load`. Caller (eval harness) is trusted to supply a sane path; we do not need to sandbox the read.
5. **JSONL output cannot leak templated content.** Events emit `path` + `action` + `source` only — never file contents.

## Verification

- [ ] `go test ./...` green (existing + new tests under `internal/nonint/`, `cmd/init_nonint_test.go`, `cmd/add_nonint_test.go`)
- [ ] `go vet ./...` clean
- [ ] `make build` succeeds; `./bonsai init --help` lists both new flags
- [ ] `./bonsai init --non-interactive --from-config <minimal.yaml>` in tmp dir → JSONL stdout, valid station tree, exit 0
- [ ] `./bonsai init --non-interactive` (no --from-config) → exit ≠ 0, stderr "must be set together"
- [ ] `./bonsai init --non-interactive --from-config /dev/null` → exit 2 + stderr error
- [ ] `./bonsai init --non-interactive --from-config <minimal.yaml>` re-run in same dir → exit 4 + stderr
- [ ] `./bonsai add --non-interactive --from-config <backend-overlay.yaml>` after init → backend workspace materialises, JSONL emitted
- [ ] Conflict-skipped semantics verified by integration test (pre-place edited file, assert bytes unchanged)
- [ ] Existing interactive `bonsai init` flow unchanged — sanity-check in tmp dir
- [ ] Windows cross-compile passes (CI gate per Plan 36 — `GOOS=windows GOARCH=amd64 go build`)
- [ ] CHANGELOG entry written under v0.5.0
- [ ] Plan 38 P2 unblocked: `bonsai_eval/solvers/rungs.py:135` can call `bonsai init --non-interactive --from-config <fixture>` and parse JSONL — confirmed by tech-lead in a follow-up Plan 38 dispatch.

## Out of scope

- `bonsai update --non-interactive` / `bonsai remove --non-interactive` — file Backlog row if Plan 38 needs them.
- `--target-dir` flag — caller manages cwd via subprocess.
- `bonsai catalog` / `bonsai list` — already non-interactive functionally; no flags needed.
- Replaying a `.bonsai.yaml` against a partial existing workspace beyond add-items semantics. Nominal use is `init` on empty dir + `add` on existing project.

## Dispatch

Single worktree, single general-purpose code agent. Phases A → B → C → D run sequentially in one branch. Independent code-review agent post-dispatch. Squash-merge as `v0.5.0` after CI green + review clean.

Worktree branch name: `agent-plan39-nonint`.
