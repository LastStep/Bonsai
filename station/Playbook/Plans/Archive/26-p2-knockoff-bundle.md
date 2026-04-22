---
tags: [plan, patch]
description: P2 knock-off bundle — context-guard path bug, stale routines dashboard, 2 Plan-23 cosmetic refactors.
---

# Plan 26 — P2 Knock-Off Bundle

**Tier:** 1
**Status:** Draft
**Agent:** general-purpose (single dispatch, worktree isolation)

## Goal

Ship four independent P2 Backlog items as one bundled PR: (1) fix misleading `context-guard.sh` planning-reminder path prefix, (2) regenerate stale `station/agent/Core/routines.md` dashboard to drop blank rows mid-table, (3) replace `selected[:0]` aliasing in two conflict-apply helpers with explicit pre-allocation, (4) rename shadowed `installedSet` closures in `cmd/add.go`.

## Steps

### Step 1 — Fix `context-guard.sh` planning-reminder path prefix

**File:** `station/agent/Sensors/context-guard.sh`

**Current bug (lines 155-157):**

```bash
reminder = (
    '\nPLANNING DETECTED. Before drafting a plan:\n'
    f'1. Load planning workflow: {os.path.join(root, "")}agent/Workflows/planning.md\n'
    f'2. Load planning-template skill: {os.path.join(root, "")}agent/Skills/planning-template.md\n'
    '3. Follow the Tier rules and Verification requirements'
)
```

`os.path.join(root, "")` produces `root/` — but `agent/` lives under `root/station/`, so injected paths point to non-existent files.

**Fix:** swap both occurrences to use the `docs_path` variable already defined at line 24 (`docs_path = os.path.join(root, 'station/')`). The existing wrap-up checklist at lines 122-124 already uses `docs_path` correctly — this change brings the planning reminder into parity.

**Edit (lines 156-157):**

```python
f'1. Load planning workflow: {docs_path}agent/Workflows/planning.md\n'
f'2. Load planning-template skill: {docs_path}agent/Skills/planning-template.md\n'
```

**Verification:** grep confirms no remaining `os.path.join(root, "")` patterns in `context-guard.sh`. Shell-parse the file: `bash -n station/agent/Sensors/context-guard.sh`. Python syntax check: feed a dummy JSON with `prompt: "lets plan this feature"` and confirm the injection contains the `station/` prefix (manual observation acceptable — no test harness exists for this sensor).

### Step 2 — Regenerate stale routines dashboard

**File:** `station/agent/Core/routines.md`

**Current state:** Blank lines at lines 38 and 55 split two markdown tables into fragments — GitHub and Obsidian render the second half without headers.

**Root cause (revised from original Backlog entry):** The current generator `RoutineDashboard()` in `internal/generate/generate.go:917-1061` produces no blank rows between table body rows — the for-loops at lines 988-1028 and 1042-1052 append rows contiguously. The file on disk was produced by an older generator version and is now stale. No generator fix needed; the file simply needs to be regenerated.

**Fix:** Delete lines 38 and 55 (the two blank lines). Update `.bonsai-lock.yaml` hash for `station/agent/Core/routines.md` to match the new content. Alternative approach: run `make build && ./bonsai update` in the worktree from the project root; accept the conflict-picker's `Overwrite` for `station/agent/Core/routines.md`; commit the resulting file plus lockfile update. Either path produces the same result — prefer the direct-edit approach since it is smaller and fully deterministic (no interactive TUI in the agent's flow).

**Regression guard:** Add a test in `internal/generate/generate_test.go` — build an `InstalledAgent` with 7 routines at mixed frequencies (mirroring tech-lead defaults), call `RoutineDashboard` into a temp dir, read the resulting file, assert that between the `ROUTINE_DASHBOARD_START` and `ROUTINE_DASHBOARD_END` markers every non-empty non-comment line begins with `|` — i.e. no blank rows splitting the table. Same assertion applied to the Routine Definitions table in the same file (between its header row and the trailing blank before end-of-file).

### Step 3 — Replace `selected[:0]` aliasing in conflict-apply helpers

**Files:** `cmd/root.go:173`, `cmd/add.go:329`

**Current pattern (identical in both sites):**

```go
filtered := selected[:0]
for _, relPath := range selected {
    if dropped[relPath] {
        droppedList = append(droppedList, relPath)
        continue
    }
    filtered = append(filtered, relPath)
}
selected = filtered
```

Safe today because the write index never overtakes the read index in iteration order, but the pattern aliases `filtered` into `selected`'s backing array — a known anti-pattern that would silently corrupt data if the loop ever changed shape (e.g. reverse iteration, inner `append` to `selected`).

**Fix (both sites):**

```go
filtered := make([]string, 0, len(selected)-len(dropped))
for _, relPath := range selected {
    if dropped[relPath] {
        droppedList = append(droppedList, relPath)
        continue
    }
    filtered = append(filtered, relPath)
}
selected = filtered
```

In `cmd/add.go:329` the variable is named `toOverwrite` — apply the same transformation (`filtered := make([]string, 0, len(toOverwrite)-len(dropped))`).

**Verification:** Existing tests in `cmd/add_test.go` (9 table tests covering Keep/Overwrite/Backup/mixed/empty/backup-read-fail/backup-write-fail/all-dropped/dropped-list-contains-all) must continue to pass. `go test ./cmd/... -count=1` green.

### Step 4 — Rename shadowed `installedSet` closures

**File:** `cmd/add.go`

**Current shadowing:**

- File-scope function at `cmd/add.go:349` — `func installedSet(cfg *config.ProjectConfig) map[string]bool`
- Inner closure at `cmd/add.go:532` — `installedSet := func(items []string) map[string]bool` (inside `distributeAddItemPicks`)
- Inner closure at `cmd/add.go:619` — `installedSet := func(items []string) map[string]bool` (inside `distributeNewAgentPicks` or similar)

Go-legal (different signatures, scope boundaries prevent accidental misuse), but mildly confusing to read.

**Fix:** rename the two inner closures to `installedItems` at both declaration sites AND every call site within their respective enclosing functions. Grep the enclosing function body for every use of the shadowed name before renaming; confirm no call site escapes the closure scope. Do NOT rename the file-scope `func installedSet` — it is called from other places (`cmd/add.go:119` and potentially elsewhere). Verify via `grep -n 'installedSet' cmd/add.go` post-edit — only the file-scope function and its call sites should remain.

**Verification:** `make build` clean. `go vet ./...` clean. `go test ./cmd/... -count=1` green.

## Dependencies

None. All four items are independent; any can be reverted without affecting the others.

## Security

> [!warning]
> Refer to [Playbook/Standards/SecurityStandards.md](../../Standards/SecurityStandards.md) for all security requirements.

Step 1 touches a sensor script: no change to the input-handling path, no new subprocess or external call — only the string content of the `additionalContext` injection changes. No security impact.

Step 2 touches a `station/` data file and a regression test: no security impact.

Steps 3 and 4 are pure refactors within the conflict-apply + add-items logic: no change to I/O boundaries, no change to privilege handling, no change to the `.bak` backup flow. Existing security tests in `cmd/add_test.go` (including `backup-read-fail` with `chmod 0500` real-OS perm checks) cover the regression surface.

## Verification

- [ ] `make build` — clean build, no warnings
- [ ] `go test ./... -count=1` — full suite green
- [ ] `go vet ./...` — clean
- [ ] `gofmt -l $(git diff --name-only --diff-filter=AM origin/main..HEAD | grep '\.go$')` — empty (no drift)
- [ ] `grep -n 'os.path.join(root, "")' station/agent/Sensors/context-guard.sh` — empty
- [ ] Blank-row count in `station/agent/Core/routines.md` between dashboard markers — zero
- [ ] New generator regression test at `internal/generate/generate_test.go` passes
- [ ] `grep -nE 'filtered := (selected|toOverwrite)\[:0\]' cmd/root.go cmd/add.go` — empty
- [ ] `grep -c 'installedSet :=' cmd/add.go` — zero (both inner closures renamed)
- [ ] Draft PR created against `main` via `gh pr create --draft`

## Out of scope

- `bonsai remove` / `bonsai update` cinematic redesigns — deferred until Plan 23 dogfood findings surface
- `context-guard.sh` rewrite to Go (its python-over-stdin shape is adequate for now; a port lives in Backlog Group E Phase 2)
- Broader dead-code audit in `cmd/` — outside bundle scope
