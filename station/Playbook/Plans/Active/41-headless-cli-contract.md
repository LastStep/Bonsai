---
tags: [plan, cli, headless, mcp, contract]
status: draft
tier: 2
agent: backend
supersedes: Plan 40 Phase 4 (update-delivery slice); folds in Rung-2.5 multi-agent --from-config
---

# Plan 41 — Headless CLI Contract + MCP-Ready Cores

**Tier:** 2
**Status:** Draft
**Agent:** backend (code agents via worktrees)

## Goal

Every mutating command (`init`, `add`, `update`, `remove`) runs fully non-interactive behind a pure, `Result`-returning core; every read command (`list`, `catalog`, `validate`) emits structured JSON; all seven share one exit-code + event-schema contract documented for AI integrators. Cores are shaped so a future `bonsai mcp` server is a thin wrapper that calls the same functions — **zero duplicated work between the CLI and MCP layers.**

"Done" = an AI agent (or CI script, or the Odysseus hub) can drive the full Bonsai lifecycle through stdout/exit-codes with no TTY, and the headless cores return structured `Result` values ready for MCP tool wrapping.

## Context

### Why now
Three pressing consumers are blocked on a headless surface, and a fourth (native AI-in-editor) is gated behind it:

1. **Odysseus hub** (Backlog P1) — must drive Bonsai headless to materialize/maintain workspaces.
2. **Bonsai-Eval rung-3** (Plan 38) — Python `subprocess.run(bonsai …)`; only `init`/`add` work today.
3. **Plan 40 dogfood** — needs an `update`-delivery path (this supersedes Plan 40 Phase 4's slice) so the self-hosted station can be re-synced.
4. **"Connect anywhere, drive with AI"** — native MCP in Claude Code / Cursor / Desktop. This plan builds the substrate; the MCP server itself is a **separate fast-follow plan** (Plan 42, not started).

### Architecture decision — layered, not parallel (research-backed)
MCP is a thin wrapper that **calls the same non-interactive core**, not a second implementation (proven by `github-mcp-server`, `fly mcp server`, `claude mcp serve`). The contract layer is the prerequisite for MCP *and* serves consumers 1–3 with zero new dependencies. The one rule that guarantees no future duplication:

> **Each command's core is a pure function: typed-options in → structured `Result` out. No Huh prompts, no `os.Exit`, no `fmt.Println` for data, no `os.Stdin` reads.** The CLI adapter serializes `Result`→JSONL (stdout); the future MCP adapter serializes the same `Result`→`structuredContent`.

`internal/nonint` already does ~half of this for `init`/`add`, but its runners **stream** JSONL to an `io.Writer` mid-run instead of returning a `Result`. Phase 1 reshapes that so the emit moves to a CLI-only adapter — keeping the JSONL **byte-identical** (Bonsai-Eval rung-3 parses the exact stream — memory: "all-zero-counts contract").

### Current state (audit 2026-06-16)
| Command | Headless today | Output | Conflict handling | Gate |
|---------|----------------|--------|-------------------|------|
| `init` | ✅ `nonint.RunInit` | JSONL | skip (count) | `--non-interactive`+`--from-config` |
| `add` | ✅ `nonint.RunAdd` | JSONL | skip (count) | `--non-interactive`+`--from-config` |
| `update` | ⚠️ `updateflow.RunStatic` (isatty fallback) | plain text | **errors out** | `isatty(stdin)` only, no flag |
| `remove` | ❌ none — cinematic harness only | none | n/a | always TTY (freezes on pipe) |
| `list` | partial (non-ANSI pipe) | text only | n/a | — |
| `catalog` | ✅ `--json` | JSON | n/a | `--json` flag |
| `validate` | ✅ `--json` | JSON | n/a | `--json` flag |

### Input-model decision (surfaced for grilling)
`update` and `remove` use **imperative flags**, not `--from-config`:
- `bonsai update --non-interactive --skip-conflicts --json` — reconciles from the existing `.bonsai.yaml` + workspace scan (no config file; matches what `RunStatic` already does and what the hub wants: push via `init`/`add`, sync via `update`).
- `bonsai remove backend --yes --json` / `bonsai remove skill <name> --yes --json` — positional args (how `remove` already works).
- `init`/`add` keep `--from-config` (declarative, already shipped).

Rationale: `remove` is inherently imperative ("remove *this*"); a declarative diff-to-target surface is more surface area for no current consumer. **Grilling should challenge this** against the Odysseus hub's "push one desired-state config and converge" model — if the hub needs declarative remove, fold a `--from-config` reconcile into a later phase.

## Steps

> Package note: keep the package name `internal/nonint` (expanding its scope from init/add to all four mutating cores). Renaming to `internal/headless` is pure churn — skip it.

### Phase 1 — Shared contract foundation + `Result` reshape (BLOCKING; lands before 2/3)

**Files:** `internal/nonint/result.go` (new), `internal/nonint/runner.go`, `internal/nonint/events.go`, `cmd/init_flow.go`, `cmd/add.go`, `internal/nonint/result_test.go` (new).

1. **Exit codes** — in `internal/nonint/runner.go`, keep `ExitOK=0`, `ExitInvalidConfig=2`, `ExitRuntime=3`, `ExitWrongCWDForInit=4` **unchanged** (back-compat). Add `ExitConflict=5` ("unresolved file conflicts; re-run with --skip-conflicts or interactively"). Document each constant's meaning in a doc-comment block.
2. **`Result` type** — new `internal/nonint/result.go`:
   ```go
   type Result struct {
       Write    *generate.WriteResult // file outcomes (created/updated/unchanged/skipped/conflicts)
       Warnings []string              // non-fatal anomalies (lock-save failure, invalid discoveries)
   }
   func (r *Result) Counts() (created, updated, unchanged, skipped, conflicts int) // delegates to Write.Summary()
   ```
   This is the structured value both the CLI JSONL adapter and the future MCP adapter consume.
3. **Emit adapter** — move the existing `emitResults(w, *generate.WriteResult)` + `EmitSummary` calls **out of** `RunInit`/`RunAdd` into a new exported `EmitJSONL(w io.Writer, r *Result) error` in `events.go`. It walks `r.Write` in the **same order** as today and emits the same `file`/`summary`/`warning` event shapes. **Byte-for-byte identical** to current output (assert with a golden test — see Verification).
4. **Reshape runners** — change `RunInit` / `RunAdd` signatures from `(…, w io.Writer) (int, error)` to `(…) (*Result, int, error)` (drop the `io.Writer` param). They build and return `*Result`; they no longer emit. Warnings (lock-save failure) go into `Result.Warnings` instead of `fmt.Fprintln(os.Stderr, …)`.
5. **CLI adapters** — `cmd/init_flow.go`'s `runInitNonInteractive` and `cmd/add.go`'s `runAddNonInteractive` now: call the runner → on success `nonint.EmitJSONL(os.Stdout, result)` → print `result.Warnings` to **stderr** → `os.Exit(code)`. Preserve the exact existing exit-code branching.
6. **Stream hygiene** — assert in code review: **stdout carries only JSONL data; every warning/diagnostic goes to stderr.** (Prerequisite for the MCP plan — stdio MCP servers require pure-protocol stdout.)

### Phase 2 — `update` headless core (depends on Phase 1)

**Files:** `cmd/update.go`, `internal/nonint/update.go` (new), `internal/tui/updateflow/run.go` (extract logic), `internal/nonint/update_test.go` (new), `cmd/update_nonint_test.go` (new).

1. **Flags** — add to `cmd/update.go`: `--non-interactive` (bool; force headless even on a TTY), `--skip-conflicts` (bool; skip+count conflicts instead of erroring), `--json` (bool; emit JSONL — implied by `--non-interactive`).
2. **Core** — new `nonint.RunUpdate(cwd string, cfg *config.ProjectConfig, cat *catalog.Catalog, lock *config.LockFile, version string, skipConflicts bool) (*Result, int, error)`. Lift the scan+render pipeline body out of `updateflow.RunStatic` (`internal/tui/updateflow/run.go:192-266`) into this core (or have `RunStatic` delegate to it so there is ONE implementation). Behavior:
   - Auto-accept every valid discovery (as `RunStatic` does today).
   - Invalid discoveries → `Result.Warnings` (not raw stderr writes).
   - Conflicts: if `skipConflicts` → skip + count in `Result.Write` (exit `ExitOK`); else → return `ExitConflict` with a `Result` that still lists the conflicts (so the agent sees *which* files).
3. **Persistence** — core saves config (when changed) + lock, surfacing save failures as `Result.Warnings` (matches `init`/`add`).
4. **Gate** — `runUpdate` routing becomes: `--non-interactive` OR `!isTerminal()` → `RunUpdate` + `EmitJSONL` + exit code; else → `updateflow.Run` (cinematic, unchanged). Explicit flag forces headless on a TTY (clig.dev: scripted-with-TTY case).
5. **Back-compat** — the plain-text non-TTY path (current `runUpdate` lines 74-95) is **replaced** by JSONL. Note in CHANGELOG: piped `bonsai update` now emits JSONL, not prose. (No external consumer depends on the prose — only the dogfood, which moves to JSONL.)

### Phase 3 — `remove` headless core (depends on Phase 1; file-disjoint from Phase 2 → parallel-eligible)

**Files:** `cmd/remove.go`, `internal/nonint/remove.go` (new), `internal/nonint/remove_test.go` (new), `cmd/remove_nonint_test.go` (new).

1. **Flags** — add `--non-interactive` / `--yes` (bool; bypass the cinematic Confirm) and `--json` (bool) to `removeCmd` and the item subcommands (`removeSkillCmd` etc.). Keep `--delete-files`.
2. **Extract business logic** — lift the non-TUI logic out of `runRemove` (agent removal: tech-lead guard, lock-untrack, `SettingsJSON` regen, optional `--delete-files` cleanup) and `runRemoveItem` (item removal) into:
   - `nonint.RunRemoveAgent(cwd string, cfg *config.ProjectConfig, cat *catalog.Catalog, lock *config.LockFile, version, agentName string, deleteFiles bool) (*Result, int, error)`
   - `nonint.RunRemoveItem(cwd string, cfg *config.ProjectConfig, cat *catalog.Catalog, lock *config.LockFile, version, itemType, itemName string) (*Result, int, error)`
   The cinematic flow (`removeflow`) calls the SAME cores for its business logic (refactor `runRemove` so TTY and headless share one path; TTY only adds the Observe/Confirm/Yield chrome).
3. **Gate** — when `--non-interactive`/`--yes` set (or stdin not a TTY): call the core, `EmitJSONL`, exit. Else: cinematic flow (unchanged).
4. **Exit codes** — agent/item not found → `ExitInvalidConfig` (2); removing the last tech-lead while other agents depend on it → `ExitInvalidConfig` (2) with a clear message; runtime/FS error → `ExitRuntime` (3).
5. **Safety** — `--yes` bypasses confirmation but NOT validation: never accept an empty/wildcard target; `--delete-files` stays an explicit separate flag (no implicit dir deletion). See Security.

### Phase 4 — read-command JSON parity (independent; parallel-eligible)

**Files:** `cmd/list.go`, `internal/tui/listflow/` (add a serializer), `cmd/list_test.go`; audit `cmd/catalog.go` + `cmd/validate.go`.

1. **`list --json`** — add a `--json` flag to `listCmd`. Emit a stable structured snapshot of installed state (agents → abilities, versions, workspace, docs_path) — reuse `cfg` + catalog data `listflow.RenderAll` already reads; add a `listflow.SerializeJSON(cfg, cat, version, cwd)` pure function. Indent-2 JSON to stdout, like `validate --json` / `catalog --json`.
2. **Consistency audit** — confirm `catalog --json` (`cmd/catalog.go:43`) and `validate --json` (`cmd/validate.go:62`) use the same JSON conventions (indent, stdout-only, exit 0 on success). Document any divergence; fix only trivial drift in this phase, file larger gaps to Backlog.

### Phase 5 — contract doc + test sweep (depends on 1-4)

**Files:** `docs/agent-interface.md` (new, repo root `docs/`), website guide page (extend the Plan 40 Formats page), `CHANGELOG.md`, cross-command golden tests.

1. **Canonical contract doc** — `docs/agent-interface.md`: per-command flags, the JSONL event schema (`file`/`summary`/`warning` shapes), the exit-code table (0/2/3/4/5 with meanings), input shapes (`--from-config` for init/add; imperative flags for update/remove), and stream discipline (data→stdout, diagnostics→stderr). This is the single source AI integrators (and the Plan 42 MCP server) read.
2. **Website** — surface a condensed version on the guide site (extend the existing Formats page from Plan 40 P3); MDX-link rule applies (memory: use `[label](url)`, never `<url>`).
3. **CHANGELOG** — `v0.5.0 (unreleased)` section: new `update`/`remove` non-interactive flags, `list --json`, JSONL-on-pipe for update, exit-code 5.
4. **Golden tests** — a cross-command test asserting the JSONL byte-shape for each mutating command against committed fixtures; exit-code assertions for each (success / invalid-config / conflict / wrong-state / not-found). Mirror `cmd/init_nonint_test.go` + `cmd/add_nonint_test.go` structure.

## Dependencies

- **No new Go modules** in this plan. The official MCP SDK (`github.com/modelcontextprotocol/go-sdk`, GA v1.6.1) is added by the **Plan 42 MCP fast-follow**, not here — keeps the binary lean until MCP is greenlit.
- Builds on: `internal/nonint` (RunInit/RunAdd), `internal/generate.WriteResult` + `.Summary()`, `internal/tui/updateflow.RunStatic`, the `removeflow` business logic, `internal/config` load/save.
- **Phase 1 is a hard prerequisite** for Phases 2 & 3 (shared `Result` + `EmitJSONL`). Phases 2, 3, 4 are mutually file-disjoint → parallel-dispatch eligible once Phase 1 merges (per memory: file-disjoint parallel rule). Phase 5 last.
- **Unblocks but does not resolve:** the Plan 40 dogfood also needs the `.bonsai-lock.yaml` gitignore policy decided (Backlog P2 — `.gitignore:15` makes `validate` report every ability as `orphaned_registration` on this repo). That is a separate decision; flag it when the dogfood is attempted, do not fix it in this plan.
- **Fast-follow:** Plan 42 — `bonsai mcp` stdio server (official go-sdk; `cmd/mcp.go`; one tool per core; `readOnlyHint`/`destructiveHint` annotations — `remove` destructive, `list`/`catalog`/`validate` read-only; `outputSchema`+`structuredContent` from `Result`; elicitation for conflict prompts; client-registration docs). Out of scope here.

## Security

> [!warning]
> Refer to [Standards/SecurityStandards.md](../../Standards/SecurityStandards.md) for all security requirements.

- **No weakening of workspace boundaries.** Headless cores write through the existing lock-aware `generate` write path and `wsvalidate` checks — same path as the TUI flows. No new direct `os.WriteFile`. (Symlink/TOCTOU hardening of the shared write path is Backlog P2 from the Plan 40 grill — track, don't regress.)
- **`remove --yes` bypasses confirmation, never validation.** Reject empty/wildcard targets; require an explicit agent/item name. `--delete-files` stays a separate explicit flag — no implicit directory deletion under `--yes`.
- **Stream separation is a security/correctness property.** Data → stdout (JSON only); all diagnostics/warnings → stderr. Prevents an injected warning string from corrupting a downstream parser, and is mandatory for the Plan 42 stdio MCP server (stdout must be pure protocol).
- **`--from-config` already validated** (Plan 39 §3 overlay-match guards); update/remove take no config file so no new untrusted-YAML surface. JSON output is marshaled, never string-concatenated.

## Verification

- [ ] `go build ./...` and `go test ./...` pass; `GOOS=windows GOARCH=amd64 go build ./...` passes (memory: POSIX-only syscall class).
- [ ] **JSONL byte-identity:** a golden test proves `EmitJSONL(Result)` for `init`/`add` is byte-for-byte identical to the pre-refactor stream (Bonsai-Eval rung-3 contract preserved).
- [ ] `bonsai update --non-interactive --skip-conflicts --json` in a drifted workspace emits `file`+`summary` JSONL on stdout, exits 0; without `--skip-conflicts` on a conflict it exits 5 and lists the conflicting files.
- [ ] `bonsai remove backend --yes --json` removes the agent headless (no prompt), emits JSONL, exits 0; non-existent agent exits 2; piped stdin (no TTY) does NOT freeze.
- [ ] `bonsai list --json` emits a stable installed-state snapshot to stdout, exits 0; shape matches the documented schema.
- [ ] Every mutating command: data only on stdout, diagnostics only on stderr (assert via captured streams in tests).
- [ ] Exit-code matrix tested per command: success(0) / invalid-config(2) / runtime(3) / wrong-state(4) / conflict(5).
- [ ] `docs/agent-interface.md` documents all four mutating + three read commands; exit-code table and JSONL schema match the implementation (cross-checked in a doc-drift test or review).
- [ ] No new Go module in `go.mod`.
- [ ] Cores return `*Result` with no TUI import in `internal/nonint` (grep: zero `huh`/`bubbletea`/`lipgloss` imports) — confirms MCP-readiness.

## Grilling Pass

<!-- populated by /grill 41 -->
