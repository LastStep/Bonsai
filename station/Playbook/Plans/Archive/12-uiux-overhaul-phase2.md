# Plan 12 — UI/UX Overhaul Phase 2: Consistency

**Tier:** 2 (Feature)
**Status:** Complete — shipped 2026-04-17 (PR #20)
**Source:** Roadmap Phase 1 "UI overhaul" + RESEARCH-uiux-overhaul.md (Phase 2 items 2.1–2.7) + Plan 11 follow-up

---

## Goal

Make every `bonsai` command feel like it comes from one tool: deterministic ordering, consistent next-step hints on mutations, structured non-fatal errors matching `FatalPanel`, accurate "nothing changed" detection in `update`, and at-a-glance counts in `list` and `catalog`.

### Success Criteria

- `bonsai list` output is deterministic (agents sorted alphabetically) and ends with a one-line summary
- `bonsai catalog` section headers show item counts (e.g. `Skills (17)`, `Skills (12 for backend)`)
- `bonsai add` (both flows) and `bonsai update` print a next-step `Hint` after success
- `bonsai update` against an in-sync workspace shows an "Up to date" panel and skips re-writing identical files
- All non-fatal `ErrorPanel` callsites use a new structured `ErrorDetail(title, detail, hint)` helper — same shape as `FatalPanel`, no exit
- `add.go:292` "all installed" message uses `EmptyPanel` (semantic empty state), not `SuccessPanel`
- Zero visual regressions on existing flows (dark or light terminals)

---

## Context

Plan 11 Phase 1 hardened the design system foundation (adaptive palette, NO_COLOR, FatalPanel, version banner). Phase 2 polishes the seven consistency gaps identified in `Research/RESEARCH-uiux-overhaul.md` Section 9:

1. Only `init` provides a next-step hint; mutation commands `add` and `update` don't
2. `add.go:292` uses `SuccessPanel` for "all already installed" — that's an empty state, not a success
3. `cmd/list.go:62` iterates `cfg.Agents` map — non-deterministic output ordering
4. `bonsai list` has no summary; users can't see scope at a glance
5. `bonsai catalog` section headers don't show counts
6. `bonsai update` always says "workspace synced" even when nothing changed; worse, it re-writes every file because `writeFile` doesn't compare content
7. Five non-fatal `ErrorPanel(msg)` sites use a flat one-line panel — no structure, no consistent next-command hint

Decisions captured during planning:

- **Hints stay on mutations only** — `add` (both flows) and `update`. Skip `remove` (post-delete "go install more" reads as pushy) and `list`/`catalog` (read-only — silence is fine).
- **Single-line summary for `list`** — total counts across all agents joined with `·`. Skip per-agent counts; the panels already show that.
- **"Up to date" detection requires generator-level no-op** — comparing rendered content to existing file content. Without this, every `update` shows every file as `ActionUpdated` and the panel can't fire.
- **`ErrorDetail` mirrors `FatalPanel`** — same render, no `os.Exit`. Five callsites migrate.

---

## Steps

### Step 1 — Sort `list.go` Agent Iteration

**File:** `cmd/list.go`

#### 1A. Sort agent names before the iteration loop

Replace `cmd/list.go:62`:

```go
// Before
for name, agent := range cfg.Agents {
```

With:

```go
agentNames := make([]string, 0, len(cfg.Agents))
for name := range cfg.Agents {
    agentNames = append(agentNames, name)
}
sort.Strings(agentNames)

for _, name := range agentNames {
    agent := cfg.Agents[name]
```

Add `"sort"` to the import block (verified absent today). Indent the existing loop body one level — no other changes inside.

---

### Step 2 — `bonsai list` Summary Line

**File:** `cmd/list.go`

#### 2A. Add summary helper next to the new sort logic

After the agent loop and before the trailing `tui.Blank()`, compute totals across all agents:

```go
totalSkills, totalWorkflows, totalProtocols, totalSensors, totalRoutines := 0, 0, 0, 0, 0
for _, a := range cfg.Agents {
    totalSkills += len(a.Skills)
    totalWorkflows += len(a.Workflows)
    totalProtocols += len(a.Protocols)
    totalSensors += len(a.Sensors)
    totalRoutines += len(a.Routines)
}

parts := []string{
    pluralize(len(cfg.Agents), "agent", "agents"),
    pluralize(totalSkills, "skill", "skills"),
    pluralize(totalWorkflows, "workflow", "workflows"),
    pluralize(totalProtocols, "protocol", "protocols"),
    pluralize(totalSensors, "sensor", "sensors"),
    pluralize(totalRoutines, "routine", "routines"),
}
fmt.Println("\n  " + tui.StyleMuted.Render(strings.Join(parts, " "+tui.GlyphDot+" ")))
```

Add `"fmt"` to imports (verified absent today — `list.go` currently imports only `os`, `path/filepath`, `strings`, `cobra`, `catalog`, `tui`), and define `pluralize` at the bottom of `list.go`:

```go
func pluralize(n int, singular, plural string) string {
    if n == 1 {
        return fmt.Sprintf("%d %s", n, singular)
    }
    return fmt.Sprintf("%d %s", n, plural)
}
```

Output shape (`·` is `tui.GlyphDot`):
```
  1 agent · 12 skills · 8 workflows · 4 protocols · 6 sensors · 7 routines
```

The summary replaces the trailing `tui.Blank()` semantically — keep `tui.Blank()` after the println so the command still ends with one blank line per the Phase 1 spacing contract.

---

### Step 3 — `bonsai catalog` Section Counts

**File:** `cmd/catalog.go`

#### 3A. Build header strings with counts

For each section in `runCatalog`, change the `tui.SectionHeader(...)` call to include the count. Folded into the existing `suffix` pattern so the agent filter still appears cleanly:

```go
// Pattern (apply to skills, workflows, protocols, sensors, routines):
suffix := fmt.Sprintf(" (%d)", len(skills))
if agentFilter != "" {
    suffix = fmt.Sprintf(" (%d for %s)", len(skills), agentFilter)
}
tui.SectionHeader("Skills" + suffix)
```

For the **Agents** and **Scaffolding** sections (no agent filter applies):

```go
tui.SectionHeader(fmt.Sprintf("Agents (%d)", len(cat.Agents)))
// ...
tui.SectionHeader(fmt.Sprintf("Scaffolding (%d)", len(cat.Scaffolding)))
```

Add `"fmt"` to imports if not present (verified absent today — currently the file only imports `cobra` and `tui`).

Apply to all 7 sections at the call sites: `catalog.go:25`, `:42`, `:54`, `:66`, `:78`, `:94`, `:102`.

The existing muted "(for X)" rendering via `StyleMuted` is dropped in favor of the merged form — the count gives the section more weight than the filter chrome did, and the heavier styling is unnecessary now.

---

### Step 4 — Next-Step Hints on Mutations

**Files:** `cmd/add.go`, `cmd/update.go`

#### 4A. `cmd/add.go` — runAdd success path (line 227–228)

After:
```go
tui.Success(fmt.Sprintf("Added %s at %s", agentDef.DisplayName, workspace))
```
Insert before the trailing `tui.Blank()`:
```go
tui.Hint("Run: bonsai list to see your full setup.")
```

#### 4B. `cmd/add.go` — runAddItems success path (line 395–396)

After:
```go
tui.Success(fmt.Sprintf("Added %d abilities to %s", totalSelected, agentDef.DisplayName))
```
Insert before the trailing `tui.Blank()`:
```go
tui.Hint("Run: bonsai list to see your full setup.")
```

#### 4C. `cmd/update.go` — runUpdate success path (line 184–189)

After the existing Success branch (either "custom files tracked" or "workspace synced"), insert before `tui.Blank()`:

```go
tui.Hint("Review changes with: bonsai list")
```

(Skipped if the new "Up to date" panel from Step 6 fires — see Step 6C.)

**Not modified:** `cmd/init.go` (already has a hint), `cmd/remove.go` (no hint per design), `cmd/list.go`, `cmd/catalog.go`, `cmd/guide.go`.

---

### Step 5 — Empty State Cleanup

**File:** `cmd/add.go`

#### 5A. Replace SuccessPanel with EmptyPanel at line 292

Current:
```go
tui.SuccessPanel("All available abilities are already installed.", "")
return nil
```

After:
```go
tui.EmptyPanel("All available abilities are already installed.\nBrowse more with: bonsai catalog")
return nil
```

**Not modified:**
- `add.go:322` (`tui.Info("No new abilities selected.")`) — this is a user-cancellation, not an empty state. Leave as `Info`.
- `list.go:44` — already uses `EmptyPanel`.
- `catalog.go` `(none)` cells via `CatalogTable` — that's per-table empty state inside a section, not a section-level empty state. Leave as-is.

---

### Step 6 — `bonsai update` "Up to date" Detection

**Files:** `internal/generate/generate.go`, `cmd/root.go`, `cmd/update.go`

This requires generator-level changes because `writeFile` currently rewrites identical files and reports them as `ActionUpdated`. Without short-circuiting, the no-op panel can never fire.

#### 6A. Add `ActionUnchanged` to the FileAction enum (`generate.go:139–147`)

```go
const (
    ActionCreated   FileAction = iota // new file written
    ActionUpdated                     // existing unmodified file overwritten with new content
    ActionUnchanged                   // existing file identical to rendered content — no write
    ActionConflict                    // file modified by user, not overwritten
    ActionForced                      // conflict overridden by user, overwritten
)
```

#### 6B. Short-circuit identical writes in `writeFile` (`generate.go:258–287`)

In the "exists, not modified by user" branch (after the conflict check, before the actual write), compare content:

```go
if existing, err := os.ReadFile(absPath); err == nil && bytes.Equal(existing, content) {
    return FileResult{RelPath: relPath, Action: ActionUnchanged, Source: source}
}
```

Add `"bytes"` to the import block of `generate.go` (verified absent today). Apply only to the path where `exists && !modified` — do not short-circuit when `force && modified` (those need a real write to persist user-acked changes).

Also update `Summary()` (`generate.go:190`) to add an `unchanged` return value:

```go
func (wr *WriteResult) Summary() (created, updated, unchanged, skipped, conflicts int) {
    for _, f := range wr.Files {
        switch f.Action {
        case ActionCreated:
            created++
        case ActionUpdated, ActionForced:
            updated++
        case ActionUnchanged:
            unchanged++
        case ActionConflict:
            conflicts++
        }
    }
    return
}
```

Verified: `Summary()` is called in exactly one place — `internal/generate/generate_test.go:363` (`created, updated, skipped, conflicts := wr.Summary()`). Update that destructure to `created, updated, _, skipped, conflicts := wr.Summary()` (or rename to use the new `unchanged` value if a relevant assertion is added). No cmd-side callers exist today.

#### 6C. Show "Up to date" panel in `cmd/update.go`

Replace the success block at `update.go:182–189`:

```go
created, updated, _, _, conflicts := wr.Summary()
hadChanges := configChanged || created > 0 || updated > 0 || conflicts > 0

if !hadChanges {
    tui.TitledPanel("Up to date",
        "Workspace is in sync with the catalog.\nNo files needed updating.",
        tui.Moss)
    tui.Blank()
    return nil
}

showWriteResults(&wr, ".")

if configChanged {
    tui.Success("Update complete — custom files tracked")
} else {
    tui.Success("Update complete — workspace synced")
}
tui.Hint("Review changes with: bonsai list")
tui.Blank()
return nil
```

The `TitledPanel` with `tui.Moss` (green) gives the success-shaped "everything is fine" feel without a redundant `tui.Success` line beneath it.

#### 6D. Skip `ActionUnchanged` in `showWriteResults` (`cmd/root.go:115–147`)

The existing switch already only handles the four old actions and silently skips others. Verify by re-reading — no change needed if the default case drops Unchanged. If a case for Unchanged is needed for completeness, leave it as a no-op.

---

### Step 7 — `ErrorDetail` Helper + 5 Migrations

**Files:** `internal/tui/styles.go`, `cmd/add.go`, `cmd/remove.go`

#### 7A. Add `ErrorDetail` to `styles.go` (after `FatalPanel`, ~line 184)

```go
// ErrorDetail renders a structured non-fatal error. Same shape as FatalPanel but does not exit.
// title: what happened (bold, error color). detail: why. hint: how to fix (muted).
func ErrorDetail(title, detail, hint string) {
    content := StyleError.Bold(true).Render(title)
    if detail != "" {
        content += "\n" + detail
    }
    if hint != "" {
        content += "\n" + StyleMuted.Render(hint)
    }
    fmt.Println("\n" + indent(PanelError.Render(content), 2))
}
```

Identical body to `FatalPanel` minus the `os.Exit(1)`. Refactoring the two to share an inner helper is tempting but adds indirection for a 3-line gain — keep them parallel.

#### 7B. Migrate the 5 non-fatal callsites

| File | Line | Title | Detail | Hint |
|------|------|-------|--------|------|
| `cmd/add.go` | 88 | `"Tech Lead required"` | `"No tech-lead agent is installed yet."` | `"Run: bonsai init"` |
| `cmd/remove.go` | 59 | `"Tech Lead in use"` | `"Other agents depend on Tech Lead. Remove them first."` | `"Run: bonsai list"` |
| `cmd/remove.go` | 200 | `"Auto-managed sensor"` | `"routine-check is added and removed automatically when routines change."` | `""` (no actionable hint) |
| `cmd/remove.go` | 228 | `it.singular + " not installed"` (e.g. "skill not installed") | `fmt.Sprintf("%q is not in any agent.", name)` | `"Run: bonsai list"` |
| `cmd/remove.go` | 272 | `"Required item"` | `fmt.Sprintf("%s is required by all agents that have it.", itemDisplayName(cat, name, it))` | `""` |

For each: replace the `tui.ErrorPanel(...)` line with the corresponding `tui.ErrorDetail(...)` call. Keep the `return nil` on the next line.

Empty `hint == ""` is supported (the conditional skips the line).

**Not migrated:** `add.go:95`, `:122`, `:236`, `init.go:56`, `:88`, `remove.go:54`, `root.go:35`, `:43` — all already use `FatalPanel` (Phase 1).

---

## Dependencies

- Zero new Go modules — `bytes` is stdlib, `sort` is stdlib, `fmt` already imported broadly.
- Steps are mostly independent. Recommended order: 1 → 2 (both touch `list.go`), 3, 4, 5, 7, then 6 (largest blast radius — generator changes).

---

## Security

> [!warning]
> Refer to SecurityStandards.md for all security requirements.

- No secrets, credentials, or API keys involved
- `writeFile` short-circuit reads files we already manage; no new file-system surface
- `ErrorDetail` does not exit and does not leak stack traces — same safety profile as `ErrorPanel`
- `bytes.Equal` on file contents is constant-time-irrelevant here (no security context)

---

## Verification

### Build & Test

- [ ] `go build ./...` — compiles
- [ ] `go vet ./...` — no issues
- [ ] `go test ./...` — all existing tests pass
- [ ] `gofmt -s -l .` — no formatting issues
- [ ] `go mod tidy` — no module changes
- [ ] `internal/generate/generate_test.go:363` — destructure updated to 5 values for the new `unchanged` return; existing assertions still pass

### Manual Testing

#### Determinism & summary
- [ ] `bonsai list` in a project with ≥2 agents → agent panels appear in alphabetical order
- [ ] `bonsai list` ends with a single muted line: `N agents · M skills · …`
- [ ] Singular/plural handled (`1 agent`, not `1 agents`)

#### Catalog counts
- [ ] `bonsai catalog` shows `Agents (6)`, `Skills (N)`, etc. on every section header
- [ ] `bonsai catalog --agent backend` shows `Skills (M for backend)` etc.

#### Hints
- [ ] `bonsai add` (new agent flow) ends with `Run: bonsai list to see your full setup.`
- [ ] `bonsai add` (add-items flow on existing agent) ends with same hint
- [ ] `bonsai update` (with changes) ends with `Review changes with: bonsai list`
- [ ] `bonsai remove` does NOT print a hint (silence intentional)
- [ ] `bonsai list`, `bonsai catalog` do NOT print hints

#### Empty states
- [ ] `bonsai add <agent>` for an agent with all abilities already installed → gray-bordered `EmptyPanel`, not green `SuccessPanel`

#### Update no-op
- [ ] Run `bonsai update` once on a clean workspace → normal flow
- [ ] Immediately run `bonsai update` again → green `Up to date` titled panel, no `Updated` tree, no `Success` line, no `Hint`
- [ ] Modify one tracked file (matching the lock — i.e. user has not edited it; e.g. by removing it and re-running) → second `update` shows the file as `Created` or `Updated`, not Unchanged
- [ ] Verify file mtimes do not change on the no-op run (ls -la before/after) — confirms short-circuit works

#### Structured errors
- [ ] `bonsai add backend` in a project without tech-lead → red panel with bold "Tech Lead required" title and `Run: bonsai init` hint line in muted gray
- [ ] `bonsai remove tech-lead` with another agent installed → "Tech Lead in use" structured panel
- [ ] `bonsai remove sensor routine-check` → "Auto-managed sensor" panel, no hint line
- [ ] `bonsai remove skill nonexistent-skill` → `"skill not installed"` structured panel
- [ ] `bonsai remove protocol security` (when required) → "Required item" structured panel

---

## Dispatch

| Phase | Agent | Isolation | Notes |
|-------|-------|-----------|-------|
| All | general-purpose | worktree | ~9 files, ~150 lines net. Generator change in Step 6 is the only non-trivial blast radius — verify `Summary()` callers before changing arity. |
