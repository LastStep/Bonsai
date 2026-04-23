# Plan 09 — Code Index Refresh

**Tier:** 1 (Patch)
**Status:** Complete — shipped 2026-04-16
**Source:** Backlog P1 — "Code index line number drift + missing entries"

## Goal

Bring `station/code-index.md` up to date with the current codebase — fix all drifted line numbers, add missing commands/files/functions/types, fix stale descriptions.

## Steps

Update `station/code-index.md` with the following corrections:

### 1. Entry Point (`main.go`)

| Current | Correct |
|---------|---------|
| `main.go:12` — embed | `main.go:15` — `//go:embed all:catalog` |
| `main.go:15` — `main()` → `cmd.Execute(sub)` | `main.go:21` — `main()` → `cmd.Execute(sub, guideContent)` |

Add: `main.go:18` — `//go:embed docs/custom-files.md` (guide content embed)

### 2. CLI Commands — add missing entries

| Command | File | Entry Function |
|---------|------|----------------|
| `bonsai update` | `cmd/update.go:28` | `runUpdate()` — detect custom files, re-render abilities, refresh CLAUDE.md |
| `bonsai guide` | `cmd/guide.go:22` | `runGuide()` — render custom files guide in terminal |

### 3. Shared Helpers (`cmd/root.go`) — fix line numbers

| Helper | Current | Correct |
|--------|---------|---------|
| `loadCatalog()` | `:27` | `:32` |
| `requireConfig()` | `:36` | `:41` |
| `resolveConflicts()` | `:53` | `:65` |
| `showWriteResults()` | `:103` | `:107` |

Also fix rootCmd: `cmd/root.go:22` → `cmd/root.go:27`

### 4. Remove Helpers (`cmd/remove.go`) — fix line numbers

| Helper | Current | Correct |
|--------|---------|---------|
| `runRemoveItem()` | `:194` | `:198` |
| `agentItemList()` | `:348` | `:366` |
| `itemIsRequired()` | `:397` | `:415` |
| `itemDisplayName()` | `:423` | `:441` |

### 5. Catalog types — add missing

Add `TriggerExample` (line 93) and `Triggers` (line 99) to the Types table.

### 6. Catalog functions — fix line numbers

| Function | Current | Correct |
|----------|---------|---------|
| `New(fsys)` | `:163` | `:177` |
| `loadItems()` | `:267` | `:281` |
| `loadSensors()` | `:318` | `:332` |
| `loadRoutines()` | `:369` | `:383` |
| `loadScaffolding()` | `:420` | `:434` |
| `loadAgents()` | `:437` | `:451` |

### 7. Config — fix line numbers, add missing type

Add `CustomItemMeta` type (line 10) to config.go types.

| Function | Current | Correct |
|----------|---------|---------|
| `Save(path)` | `:30` | `:40` |
| `Load(path)` | `:39` | `:49` |

### 8. Generator — fix line numbers, add missing functions

**Core Generation Functions:**

| Function | Current | Correct |
|----------|---------|---------|
| `Scaffolding()` | `:279` | `:332` |
| `SettingsJSON()` | `:385` | `:439` |
| `WorkspaceClaudeMD()` | `:452` | `:615` |
| `AgentWorkspace()` | `:737` | `:1158` |
| `RoutineDashboard()` | `:621` | `:884` |
| `EnsureRoutineCheckSensor()` | `:583` | `:846` |

Add missing core functions:
- `PathScopedRules()` — `:1032` — Generate `.claude/rules/skill-{name}.md` for path-scoped auto-loading
- `WorkflowSkills()` — `:1070` — Generate `.claude/skills/{name}/SKILL.md` for curated workflows

**Write System:**

| Function | Current | Correct |
|----------|---------|---------|
| `writeFile()` | `:205` | `:258` |
| `writeFileChmod()` | `:237` | `:290` |
| `ForceConflicts()` | `:181` | `:208` |

Add missing: `ForceSelected()` — `:230` — Overwrite only user-selected conflict files

**Helpers:**

| Function | Current | Correct |
|----------|---------|---------|
| `renderContent()` | `:250` | `:303` |
| `parseFrequencyDays()` | `:609` | `:872` |

Fix description: `descFor()` — "Build name→description map for nav tables (supports custom items)"

Add missing helpers:
- `scenariosDesc()` — `:117` — Trigger-aware description for CLAUDE.md tables
- `CuratedSlashWorkflows` — `:130` — Package-level set of workflows that get slash-command files
- `howToWorkLines()` — `:518` — Generate "How to Work" heuristics section
- `quickTriggersLines()` — `:573` — Generate Quick Triggers reference table
- `triggerSection()` — `:1122` — Generate trigger header for ability files
- `hasScaffolding()` — `:318` — Check if scaffolding item is selected

### 9. TUI Styles — add missing function

Add `EmptyPanel()` — `:167` — Dim panel for empty states

### 10. File Layout section

Add `.claude/rules/` and `.claude/skills/` to the layout tree (new from Phase A triggers).

## Security

> [!warning]
> Refer to SecurityStandards.md for all security requirements.

No security implications — documentation-only change within station/.

## Verification

- [ ] Every line number in the updated code-index.md matches the actual source
- [ ] All Go files in the project have corresponding entries
- [ ] All public functions are listed
- [ ] No stale descriptions remain
