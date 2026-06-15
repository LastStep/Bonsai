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

Every mutating command (`init`, `add`, `update`, `remove`) runs fully non-interactive behind a pure, `Result`-returning core; every read command (`list`, `catalog`, `validate`) emits structured JSON; all seven share one **exit-code contract** and a documented **event/schema philosophy** (two serializations: JSONL for streaming mutation progress, single-doc JSON for read snapshots). Cores are shaped so a future `bonsai mcp` server (Plan 42) is a thin wrapper calling the same functions — **zero duplicated work between the CLI and MCP layers.**

"Done" = an AI agent (or CI script, or a headless hub) can drive the full Bonsai lifecycle through stdout/exit-codes with no TTY, and the headless cores return structured `Result` values ready for MCP tool wrapping.

## Context

### Why now
Three pressing consumers are blocked on a headless surface, and a fourth (native AI-in-editor) is gated behind it:

1. **Headless hub** (Backlog P1) — must drive Bonsai non-interactively to materialize/maintain workspaces.
2. **Bonsai-Eval rung-3** (Plan 38) — Python `subprocess.run(bonsai …)`; only `init`/`add` work today.
3. **Plan 40 dogfood** — needs an `update`-delivery path (this supersedes Plan 40 Phase 4's slice) so the self-hosted station can be re-synced.
4. **"Connect anywhere, drive with AI"** — native MCP in Claude Code / Cursor / Desktop. This plan builds the substrate; the MCP server itself is a **separate fast-follow plan (Plan 42, not started).**

> **Consumer-name discipline:** specific consumer names (hub product names etc.) appear in plan rationale ONLY — never in shipped code, flags, output fields, or the contract doc. ("Bonsai-Eval" already appears in shipped `internal/nonint` comments — pre-existing precedent.)

### Architecture decision — layered, not parallel (research-backed)
MCP is a thin wrapper that **calls the same non-interactive core**, not a second implementation (proven by `github-mcp-server`, `fly mcp server`, `claude mcp serve`). The contract layer is the prerequisite for MCP *and* serves consumers 1–3 with zero new dependencies. The one rule that guarantees no future duplication:

> **Each command's core is a pure function: typed-options in → structured `Result` out. No Huh prompts, no `os.Exit`, no `fmt.Println` for data, no `os.Stdin` reads.** The CLI adapter serializes `Result`→JSONL (stdout); the future MCP adapter serializes the same `Result`→`structuredContent`.

`internal/nonint` already does ~half of this for `init`/`add`, but its runners **stream** JSONL to an `io.Writer` mid-run instead of returning a `Result`. Phase 1 reshapes that so the emit moves to a CLI-only adapter — keeping the JSONL **byte-identical** (Bonsai-Eval rung-3 parses the exact stream; `events_test.go` pins the all-zero-counts contract).

### Current state (audit 2026-06-16, Reality-critic verified)
| Command | Headless today | Output | Conflict handling | Gate |
|---------|----------------|--------|-------------------|------|
| `init` | ✅ `nonint.RunInit` | JSONL | skip (count) | `--non-interactive`+`--from-config` |
| `add` | ✅ `nonint.RunAdd` | JSONL | skip (count) | `--non-interactive`+`--from-config` |
| `update` | ⚠️ `updateflow.RunStatic` (isatty fallback) | plain text | **errors out** | `isatty(stdin)` only, no flag |
| `remove` | ❌ none — cinematic harness only | none | n/a | always TTY (freezes on pipe) |
| `list` | partial (non-ANSI pipe) | text only | n/a | — |
| `catalog` | ✅ `--json` | single-doc JSON | n/a | `--json` flag |
| `validate` | ✅ `--json` | single-doc JSON | n/a | `--json` flag |

### Flag surface (decided 2026-06-16 — grill Round 1)
- `init`/`add`: keep `--non-interactive` + `--from-config` (declarative, shipped v0.4.2).
- `update`: `--non-interactive` (force headless on a TTY) + `--skip-conflicts` (skip+count vs exit 5). **No `--json`** — headless mode always emits JSONL (a `--json` on a mutating command is redundant with `--non-interactive`).
- `remove`: `--non-interactive`/`--yes` (bypass cinematic confirm) + `--from <agent>` (disambiguate multi-owner item removal) + keep `--delete-files`. **No `--json`** (same reason).
- `list`/`catalog`/`validate`: `--json` is the sole structured-output trigger (read commands have no "headless" mode).

**Input model (imperative for update/remove):** `bonsai update --non-interactive --skip-conflicts` reconciles from the existing `.bonsai.yaml` + workspace scan (no config file — matches what `RunStatic` does today). `bonsai remove backend --yes` / `bonsai remove skill <name> --yes [--from <agent>]` take positional args. Rationale: `remove` is inherently imperative; a declarative diff-to-target surface is more surface for no current consumer. (Challenged in grilling; retained.)

## Steps

> Package note: keep the package name `internal/nonint` (scope expands from init/add to all four mutating cores). No rename — pure churn.

### Phase 1 — Shared contract foundation + `Result` reshape (BLOCKING; lands and MERGES before 2/3 branch)

**Files:** `internal/nonint/result.go` (new), `internal/nonint/runner.go`, `internal/nonint/events.go`, `cmd/init_flow.go`, `cmd/add.go`, `internal/nonint/testdata/{init,add}_golden.jsonl` (new fixtures), `internal/nonint/result_test.go` (new).

1. **Golden baseline FIRST (pre-refactor).** Before changing any signature: add committed input fixtures and capture the CURRENT `RunInit`/`RunAdd` stdout (on `main`, unrefactored) into `internal/nonint/testdata/init_golden.jsonl` + `add_golden.jsonl`. Pin the catalog `version` param to a fixed string so the `WriteCatalogSnapshot` line is reproducible. These fixtures are the byte-identity oracle for step 3 — commit them in the FIRST commit of the PR, before the reshape. (Verification B1.)
2. **Exit codes** — in `internal/nonint/runner.go`, keep `ExitOK=0`, `ExitInvalidConfig=2`, `ExitRuntime=3`, `ExitWrongCWDForInit=4` **unchanged** (back-compat; `5` is verified free). Add `ExitConflict=5` ("unresolved file conflicts; re-run with --skip-conflicts or interactively"). Doc-comment each constant + which commands can emit it (see exit-code reachability table in Verification).
3. **`Result` type** — new `internal/nonint/result.go`:
   ```go
   type Result struct {
       Write    *generate.WriteResult // file outcomes (created/updated/unchanged/skipped/conflicts)
       Warnings []string              // non-fatal anomalies (lock-save failure, invalid discoveries)
   }
   func (r *Result) Counts() (created, updated, unchanged, skipped, conflicts int) // delegates to Write.Summary()
   ```
   The structured value both the CLI JSONL adapter and the future MCP adapter consume. **Note:** `Result` is intentionally thin (one extra field over `WriteResult`) until Plan 42 enriches it for MCP `structuredContent`. It is **headless-only** — the TTY path keeps using `updateflow.Result` (`ConfigChanged`/`Cancelled`/`SyncErr` flow-control fields the cinematic Yield stage reads); do NOT unify the two. (Architecture #4.)
4. **Emit adapter** — new exported `EmitJSONL(w io.Writer, r *Result) error` in `events.go` (effectively the current `emitResults` exported + walking `r.Write`). It emits `file` + `summary` events to stdout. **`warning` events are NEVER written to stdout** — warnings live in `Result.Warnings` and the CLI adapter prints them to **stderr** as plain text. (Drop the `warning` JSONL event from the stdout schema entirely; keep `EmitWarning` only if a stderr-targeted JSONL warning is wanted — default plain stderr text.) (Security #1/#2.)
5. **Reshape runners** — change `RunInit`/`RunAdd` from `(…, w io.Writer) (int, error)` to `(…) (*Result, int, error)` (drop the `io.Writer` param). They build and return `*Result`; they no longer emit. **Preserve two existing seams or the contract breaks:** (a) the **all-installed zero-summary short-circuit** (`runner.go:178-184`, currently `EmitSummary(w,0,0,0,0,0)`) must still produce a zero-count `Result` that `EmitJSONL` renders identically; (b) lock-save **warnings** (`runner.go:99,190`) move from raw `os.Stderr` into `Result.Warnings`. (Architecture #1.)
6. **CLI adapters** — `runInitNonInteractive` / `runAddNonInteractive` now: call runner → `nonint.EmitJSONL(os.Stdout, result)` → print `result.Warnings` to **stderr** → `os.Exit(code)`. Preserve exact existing exit-code branching.
7. **Stream hygiene is a tested invariant, not a review note.** Replace any "assert in code review" with the helper-boundary test in Verification (stdout parses as pure JSONL; stderr carries warnings). Prerequisite for the Plan 42 stdio MCP server (stdout must be pure protocol).

### Phase 2 — `update` headless core (depends on Phase 1 MERGED)

**Files:** `cmd/update.go`, `internal/nonint/update.go` (new), `internal/tui/updateflow/run.go` (collapse `RunStatic` to a shim), `internal/nonint/update_test.go` (new), `cmd/update_nonint_test.go` (new).

1. **Flags** — add `--non-interactive` (bool; force headless even on a TTY) and `--skip-conflicts` (bool) to `cmd/update.go`. No `--json`.
2. **One implementation, not two.** New `nonint.RunUpdate(cwd string, cfg *config.ProjectConfig, cat *catalog.Catalog, lock *config.LockFile, version string, skipConflicts bool) (*Result, int, error)` is the **single** scan+render implementation. Lift the body of `updateflow.RunStatic` (`run.go:192-266` — already TUI-free) into it, then **collapse `updateflow.RunStatic` to a thin delegating shim** (or delete it and point its caller at `RunUpdate`) so there is no duplicate. Behavior:
   - Auto-accept every valid discovery (as `RunStatic` does today).
   - Invalid discoveries → `Result.Warnings` (was raw `os.Stderr` at `run.go:225-227`).
   - Conflicts: `skipConflicts` → skip + count in `Result.Write`, exit `ExitOK`; else → return `ExitConflict` (5) with a `Result` whose `Write.Files` still lists the conflicting entries (so the agent sees *which* files). **This replaces today's behavior** where `RunStatic` folds conflicts into `SyncErr` and cobra returns a non-zero generic error — a deliberate semantic change (error → exit 5), logged in CHANGELOG.
3. **Persistence** — core saves config (when changed) + lock; save failures → `Result.Warnings` (matches init/add).
4. **Gate** — `runUpdate` routing: `--non-interactive` OR `!isTerminal()` → `RunUpdate` + `EmitJSONL` + exit code; else → `updateflow.Run` (cinematic, unchanged). Explicit flag forces headless on a TTY.
5. **Breaking-output note** — piped `bonsai update` changes from human prose (`tui.Success`/`showWriteResults`, current `cmd/update.go:74-95`) to JSONL, and conflict from generic error to exit 5. Both are **Changed** entries in CHANGELOG (Phase 5), not just Added. (Risk #3, Architecture #2.)

### Phase 3 — `remove` headless core (depends on Phase 1 MERGED; file-disjoint from Phase 2 → parallel-eligible)

**Files:** `cmd/remove.go`, `internal/nonint/remove.go` (new), `internal/nonint/remove_test.go` (new), `cmd/remove_nonint_test.go` (new).

1. **Flags** — add `--non-interactive`/`--yes` (bypass cinematic Confirm) and `--from <agent>` (string; scope item removal to one owner) to `removeCmd` + item subcommands. Keep `--delete-files`. No `--json`.
2. **Extract cores (two functions):**
   - `nonint.RunRemoveAgent(cwd, cfg, cat, lock, version, agentName string, deleteFiles bool) (*Result, int, error)` — agent removal logic is a self-contained harness closure today (`remove.go:124-141`, plus the post-harness `--delete-files` cleanup `remove.go:205-219`); lift it cleanly. **Preserve the tech-lead-in-use guard** (`remove.go:68`).
   - `nonint.RunRemoveItem(cwd, cfg, cat, lock, version, itemType, itemName, fromAgent string) (*Result, int, error)` — `fromAgent` may be empty. **Item-removal target resolution is currently entangled with the harness SelectStage** (`resolveRemoveTargets` reads the picker result; `runRemoveItemAction` at `remove.go:565-614` takes `capturedTargets`). The core must re-implement target resolution **outside the harness**: compute owning agents directly; if >1 own the item and `fromAgent==""` → return `ExitInvalidConfig` (2) with a message naming the owners; if `fromAgent` set, scope to it (exit 2 if that agent doesn't own it). **Preserve `filterRequired` short-circuit and the `routine-check` auto-managed-sensor block** (`remove.go:292`). (Architecture #3.)
   - Refactor the cinematic path to call these same cores for business logic; TTY only adds Observe/Confirm/Yield chrome + the multi-owner picker.
3. **Gate** — `--non-interactive`/`--yes` set OR non-TTY → core + `EmitJSONL` + exit. Else cinematic (unchanged).
4. **Exit codes** — item/agent not found → 2; multi-owner item with no `--from` → 2; last tech-lead while other agents depend → 2 (message contains `tech-lead`); runtime/FS error → 3.
5. **Safety** — `--yes` bypasses confirmation, NOT validation: reject empty target and a literal `*` (exit 2, zero FS mutation — tested). `--delete-files` stays a separate explicit flag. Under `--yes --delete-files`, `Lstat` the target dir and **refuse if it is a symlink** (exit 2) — this is where the deferred Backlog-P2 symlink hardening matters most, since the human confirm gate is gone. (Security #4.)

### Phase 4 — read-command JSON parity (depends on Phase 1 only; parallel-eligible)

**Files:** `cmd/list.go`, a **new TUI-free** serializer package (NOT `internal/tui/listflow` — that imports glamour/initflow chrome; mirror `validate`/`catalog` which serialize from TUI-free packages), `cmd/list_test.go`; audit `cmd/catalog.go` + `cmd/validate.go`.

1. **`list --json`** — add a `--json` flag to `listCmd`. Define an **explicit output struct** (pinned, mirroring `validate.Report`) — no map-vs-list ambiguity:
   ```go
   type ListSnapshot struct {
       Version  string      `json:"version"`
       DocsPath string      `json:"docs_path"`
       Agents   []ListAgent `json:"agents"`
   }
   type ListAgent struct {
       Type      string   `json:"type"`
       Workspace string   `json:"workspace"`
       Skills    []string `json:"skills"`
       Workflows []string `json:"workflows"`
       Protocols []string `json:"protocols"`
       Sensors   []string `json:"sensors"`
       Routines  []string `json:"routines"`
   }
   ```
   `SerializeJSON(cfg, cat, version, cwd) ([]byte, error)` in the TUI-free package; `json.MarshalIndent("", "  ")` to stdout, matching `validate --json` / `catalog --json`. (Verification B2, Architecture #6.)
2. **Consistency audit** — confirm `catalog --json` (`cmd/catalog.go:43`) and `validate --json` (`cmd/validate.go:62`) share conventions (indent-2, stdout-only, exit 0). Document the **two-format split** (JSONL for mutating, single-doc JSON for read) in the Phase-5 contract doc so Plan 42 doesn't assume uniform JSONL. Fix only trivial drift here; file larger gaps to Backlog.

### Phase 5 — contract doc + test sweep (depends on 1-4)

**Files:** `docs/agent-interface.md` (new file in the **existing** `docs/` dir), `docs/formats.md` (extend — this is the binary-embedded `bonsai guide` Formats page from Plan 40 P3, surfaced via `embed.go`+`guideflow`, NOT a website page), `CHANGELOG.md`, cross-command golden tests.

1. **Canonical contract doc** — `docs/agent-interface.md`: per-command flags; the **two serializations** (JSONL `file`/`summary` event shapes for mutating commands on stdout; single-doc indent-2 JSON for read commands); the exit-code reachability table (which of 0/2/3/4/5 each command emits); input shapes; stream discipline (data→stdout, diagnostics→stderr, warnings stderr-only). Single source for AI integrators and Plan 42. No consumer product names.
2. **`bonsai guide` Formats page** — extend `docs/formats.md` (embedded into the binary) with a condensed pointer to the contract. (MDX rule N/A — these are plain `.md` embedded docs, not website MDX.)
3. **CHANGELOG** `## [0.5.0] - Unreleased`: **Added** — `update`/`remove` non-interactive flags, `--from`, `list --json`, `ExitConflict=5`. **Changed** — piped `bonsai update` now emits JSONL (was prose); update conflicts now exit 5 (was generic error). (Risk #3.)
4. **Tests** (mirror `cmd/init_nonint_test.go` + `cmd/add_nonint_test.go`): see Verification for the concrete gate list.

## Dependencies

- **No new Go modules** in this plan. The official MCP SDK (`github.com/modelcontextprotocol/go-sdk`, GA per external research) is added by **Plan 42**, not here — keeps the binary lean.
- Builds on: `internal/nonint` (RunInit/RunAdd), `internal/generate.WriteResult` + `.Summary()` (verified returns the 5 counts), `internal/tui/updateflow.RunStatic`, the `removeflow`/`cmd/remove.go` business logic, `internal/config` load/save (which runs `wsvalidate` on every workspace — the headless paths inherit the workspace-escape guard for free).
- **Phase 1 must MERGE before Phases 2 & 3 branch** — it changes `RunInit`/`RunAdd` signatures (callers `cmd/init_flow.go:294`, `cmd/add.go:767`) and creates `result.go` that 2/3 import. Dispatch Phase 1 alone, merge its PR, branch 2/3/4 from merged `main`. Do NOT fan out 2/3/4 from a pre-merge Phase-1 branch (manual discipline — no harness gate). (Risk #2.)
- Phases 2, 3, 4 are mutually file-disjoint (Risk-critic verified) → parallel-dispatch eligible after Phase 1 merges. Phase 5 last.
- **Delivery clarification:** the new *flags/behavior* ship inside the **binary** → reach existing users via `brew upgrade` / `go install …@latest` (the release pipeline, out of scope here). `bonsai update` delivers regenerated **workspace/template** content, not the binary. The Plan 40 dogfood re-sync needs (a) the new binary installed, then (b) `bonsai update` run. (Risk #4.)
- **Unblocks but does not resolve:** the dogfood also needs the `.bonsai-lock.yaml` gitignore policy decided (Backlog P2 — `.gitignore:15` makes `validate` report every ability as `orphaned_registration` here). Separate decision; flag at dogfood time, do not fix here.
- **Fast-follow:** Plan 42 — `bonsai mcp` stdio server (go-sdk; `cmd/mcp.go`; one tool per core; `readOnlyHint`/`destructiveHint` — `remove` destructive, read cmds read-only; `outputSchema`+`structuredContent` from `Result`; elicitation for conflict prompts; client-registration docs). Out of scope here.

## Security

> [!warning]
> Refer to [Standards/SecurityStandards.md](../../Standards/SecurityStandards.md) for all security requirements.

- **No weakening of workspace boundaries.** Headless cores write through the existing lock-aware `generate` write path (force=false) and inherit `config.Load`→`Validate()`→`wsvalidate` (rejects `..`-escape in `workspace`/`docs_path` at load — verified same for headless `remove --yes` as cinematic). No new direct `os.WriteFile`. Symlink/TOCTOU hardening of the shared write path stays Backlog P2 — track, don't regress.
- **`remove --yes` bypasses confirmation, never validation.** Reject empty/`*` targets (tested, zero mutation). Multi-owner item removal errors without `--from`. `--yes --delete-files` `Lstat`s the target and refuses a symlink (the human gate is gone → this is where it bites).
- **Stream separation is a security/correctness property and a tested invariant.** Data (JSONL) → stdout only; all diagnostics/warnings → stderr. `warning` events never touch stdout. Prevents an injected warning string (derived from untrusted on-disk frontmatter scanning) from corrupting a downstream parser; mandatory for the Plan 42 stdio MCP server. `json.Marshal` escaping makes injection impossible even before this, but the boundary is enforced anyway.
- **No new untrusted-input surface.** update/remove take no config file. All JSON/JSONL is marshaled (`json.Marshal`/`MarshalIndent`), never string-concatenated. No new Go dependency.

## Verification

**Exit-code reachability (the matrix tests assert only the reachable subset per command):**

| Command | 0 | 2 | 3 | 4 | 5 |
|---------|---|---|---|---|---|
| init | ✅ | ✅ | ✅ | ✅ (config exists) | — |
| add | ✅ | ✅ | ✅ | ✅ (no config) | — |
| update | ✅ | ✅ | ✅ | ✅ (no `.bonsai.yaml`) | ✅ (conflict, no `--skip-conflicts`) |
| remove | ✅ | ✅ (not-found / multi-owner / last-tech-lead) | ✅ | ✅ (no `.bonsai.yaml`) | — |

- [ ] `go build ./...`, `go test ./...`, `go vet ./...`, and `GOOS=windows GOARCH=amd64 go build ./...` all pass.
- [ ] **Byte-identity (B1):** `internal/nonint/testdata/{init,add}_golden.jsonl` captured from `main` PRE-refactor (committed in the PR's first commit). A test asserts `EmitJSONL(result)` for the pinned input fixtures (fixed `version`) equals the golden bytes. **If the diff is non-empty, the refactor is wrong — do not merge** (rollback contingency for the back-compat-critical init/add stream).
- [ ] **Stream separation (C5):** each `*_nonint_test.go` drives the **helper** (not `os.Exit`), captures `stdout`/`stderr` buffers, asserts every non-empty stdout line `json.Unmarshal`s to a known event shape (`file`/`summary`) and stderr contains no `{`-leading JSON. Covers the deleted Phase-1.6 review gate.
- [ ] **`update` exit-5 negative control (C2):** in `cmd/update_nonint_test.go`, init a project, overwrite `station/agent/Core/identity.md` with user bytes WITHOUT a lock update (the `TestRunInit_ConflictEmittedNotForced` recipe, `runner_test.go:469`), then `RunUpdate(..., skipConflicts=false)` → assert exit `ExitConflict` (5) AND `Result.Write.Files` has an `Action==conflict` entry for that path; `skipConflicts=true` → exit 0, that file counted in `skipped`.
- [ ] **`remove` negative controls (C3):** init tech-lead + add backend; `RunRemoveAgent("tech-lead")` → exit 2, `err` contains `tech-lead`; remove backend first, then tech-lead → exit 0. `RunRemoveItem` for a skill owned by 2 agents with `fromAgent==""` → exit 2, message names the owners; with `--from <owner>` → exit 0. `bonsai remove "" --yes` and `bonsai remove "*" --yes` → exit 2, zero FS mutation.
- [ ] **`list --json` schema (B2):** `json.Unmarshal(out, &ListSnapshot{})` succeeds for a two-agent fixture and every field is populated; field names/types match the struct in Phase 4.1.
- [ ] **Flag registration** — per-command tests assert each new flag (`--non-interactive`, `--skip-conflicts`, `--yes`, `--from`, `list --json`) is registered (mirror `TestInitCmd_FlagsRegistered`).
- [ ] **Doc-drift (C1, mechanical — no "or review"):** a test asserts the documented exit-code constants in `docs/agent-interface.md` (0/2/3/4/5 + names) equal the `nonint.Exit*` constant values (string-scan or reflection). 
- [ ] **MCP-readiness:** a Go test (via `go/packages` over `internal/nonint`, run in `go test ./...`) asserts zero `huh`/`bubbletea`/`lipgloss`/`charm` imports.
- [ ] `go.mod` unchanged (no new module).

## Grilling Pass

<!-- populated below by /grill 41 -->
