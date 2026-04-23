# Plan 03 — Cross-Link Generated Files with Obsidian-Compatible Markdown Links

**Tier:** 2
**Status:** Complete — shipped 2026-04-16
**Agent:** tech-lead

## Goal

All file references in generated workspace files become clickable markdown links (`[text](path)`) using correct relative paths. Compatible with Obsidian, GitHub, and VS Code markdown preview.

## Context

Currently, ~95 cross-file references across catalog templates use backtick code spans (`` `agent/Core/memory.md` ``). These render as monospace text but are NOT clickable. Zero markdown link syntax exists anywhere in the catalog today.

Only NAVIGATE references (52 instances) are converted. INSTRUCT references (agent tool calls) and CODE_BLOCK references stay as backticks — they're functional paths the AI agent needs verbatim.

## Steps

### Step 1 — Go code: `WorkspaceClaudeMD()` in `internal/generate/generate.go`

This is the highest-impact change — affects every generated workspace.

**1a. Navigation tables (lines 518-623)**

Change ALL backtick file references in `fmt.Sprintf` calls to markdown links.

Current pattern:
```go
fmt.Sprintf("| `agent/Protocols/%s.md` | %s |", p, protoDescs[p])
```

New pattern:
```go
fmt.Sprintf("| [agent/Protocols/%s.md](agent/Protocols/%s.md) | %s |", p, p, protoDescs[p])
```

Apply to every table row in these sections:
- **Core** (lines 521-523): hardcoded `agent/Core/{identity,memory,self-awareness}.md`
- **Protocols** (line 534): `agent/Protocols/%s.md`
- **Workflows** (line 547): `agent/Workflows/%s.md`
- **Skills** (line 560): `agent/Skills/%s.md`
- **Routines** (line 587-588): `agent/Routines/%s.md`
- **Sensors** (lines 607, 618): `agent/Sensors/%s.sh`

Also convert the routines dashboard reference (line 591): `` `agent/Core/routines.md` `` → `[agent/Core/routines.md](agent/Core/routines.md)`

**1b. Header reference (line 514)**

Current: `` > **FIRST:** Read `agent/Core/identity.md`, then `agent/Core/memory.md`. ``

New: `` > **FIRST:** Read [agent/Core/identity.md](agent/Core/identity.md), then [agent/Core/memory.md](agent/Core/memory.md). ``

**1c. Memory section (lines 629-631)**

The warning about auto-memory contains `` `agent/Core/memory.md` `` — convert the second reference (the one pointing to the project file) to a link. The `~/.claude/...` external path stays as backtick.

Current: `...goes in \`agent/Core/memory.md\` — version-controlled...`
New: `...goes in [agent/Core/memory.md](agent/Core/memory.md) — version-controlled...`

Same for the follow-up line: `...write to the appropriate section in \`agent/Core/memory.md\` instead.`
New: `...write to the appropriate section in [agent/Core/memory.md](agent/Core/memory.md) instead.`

**1d. External References (lines 636-646)**

These use `docsPrefix` (e.g., `station/`). Since CLAUDE.md sits at workspace root, links to docs files should be relative from workspace root. Compute the relative path:

```go
extRef := func(target string) string {
    full := filepath.Join(cfg.DocsPath, target)
    rel, err := filepath.Rel(installed.Workspace, full)
    if err != nil {
        return target
    }
    return rel
}
```

Then:
```go
fmt.Sprintf("| Project snapshot | [%sINDEX.md](%s) |", docsPrefix, extRef("INDEX.md"))
```

If `DocsPath == Workspace` (common case), this produces `[station/INDEX.md](INDEX.md)`.

Display text keeps the docsPrefix for agent context; href is the correct relative path.

### Step 2 — Catalog static files: Protocols

All protocols install to `agent/Protocols/`. Relative path base:
- To `agent/{X}/file.md` → `../{X}/file.md`
- To workspace-root files → `../../file.md`

**catalog/protocols/session-start/session-start.md** — Convert NAVIGATE refs only:
- Line 15: `` `agent/Core/identity.md` `` → `[agent/Core/identity.md](../Core/identity.md)`
- Line 16: `` `agent/Core/memory.md` `` → `[agent/Core/memory.md](../Core/memory.md)`
- Line 17: `` `agent/Core/self-awareness.md` `` → `[agent/Core/self-awareness.md](../Core/self-awareness.md)`
- Line 18: `` `INDEX.md` `` → `[INDEX.md](../../INDEX.md)`
- Line 19: `` `Playbook/Status.md` `` → `[Playbook/Status.md](../../Playbook/Status.md)`
- Line 20: `` `Playbook/Backlog.md` `` → `[Playbook/Backlog.md](../../Playbook/Backlog.md)`
- Line 21: `` `Logs/FieldNotes.md` `` → `[Logs/FieldNotes.md](../../Logs/FieldNotes.md)`
- Line 22: `` `Reports/Pending/` `` → `[Reports/Pending/](../../Reports/Pending/)`
- Line 23: `` `agent/Protocols/security.md` `` → `[agent/Protocols/security.md](security.md)` (same dir)
- Line 24: `` `agent/Protocols/scope-boundaries.md` `` → `[agent/Protocols/scope-boundaries.md](scope-boundaries.md)`
- Line 37: `` `Playbook/Standards/SecurityStandards.md` `` → `[Playbook/Standards/SecurityStandards.md](../../Playbook/Standards/SecurityStandards.md)`
- Line 38: `` `agent/Skills/` `` → `[agent/Skills/](../Skills/)`
- Line 42: `` `Playbook/Plans/Active/` `` → `[Playbook/Plans/Active/](../../Playbook/Plans/Active/)`
- Line 48: `` `Playbook/Standards/SecurityStandards.md` `` → `[Playbook/Standards/SecurityStandards.md](../../Playbook/Standards/SecurityStandards.md)`
- Lines 49: `` `Reports/Pending/` `` — INSTRUCT, keep backtick

**catalog/protocols/memory/memory.md** — Convert NAVIGATE refs only:
- Line 12: `` `agent/Core/memory.md` `` → `[agent/Core/memory.md](../Core/memory.md)`
- Line 20: `` `agent/Core/memory.md` `` — INSTRUCT, keep backtick
- Line 35: `` `agent/Core/memory.md` `` → `[agent/Core/memory.md](../Core/memory.md)`

**catalog/protocols/security/security.md**:
- Line 35: `` `Playbook/Standards/SecurityStandards.md` `` → `[Playbook/Standards/SecurityStandards.md](../../Playbook/Standards/SecurityStandards.md)`

### Step 3 — Catalog static files: Workflows

All workflows install to `agent/Workflows/`. Same relative path rules as protocols.

**catalog/workflows/issue-to-implementation/issue-to-implementation.md** — Convert NAVIGATE refs only:
- Line 21: `` `agent/Skills/issue-classification.md` `` → `[agent/Skills/issue-classification.md](../Skills/issue-classification.md)`
- Line 22: `` `agent/Skills/dispatch.md` `` → `[agent/Skills/dispatch.md](../Skills/dispatch.md)`
- Line 23: `` `agent/Skills/planning-template.md` `` → `[agent/Skills/planning-template.md](../Skills/planning-template.md)`
- Line 49: `` `Playbook/Status.md` `` → `[Playbook/Status.md](../../Playbook/Status.md)`
- Line 69: `` `agent/Skills/issue-classification.md` `` → `[agent/Skills/issue-classification.md](../Skills/issue-classification.md)`
- Line 84: `` `Playbook/Status.md` `` → `[Playbook/Status.md](../../Playbook/Status.md)`
- Line 85: `` `Playbook/Backlog.md` `` → `[Playbook/Backlog.md](../../Playbook/Backlog.md)`
- Line 86: `` `Logs/KeyDecisionLog.md` `` → `[Logs/KeyDecisionLog.md](../../Logs/KeyDecisionLog.md)`
- Line 162: `` `agent/Skills/dispatch.md` `` → `[agent/Skills/dispatch.md](../Skills/dispatch.md)`
- Line 191: `` `agent/Skills/dispatch.md` `` → `[agent/Skills/dispatch.md](../Skills/dispatch.md)`
- Line 398: `` `agent/Workflows/pr-review.md` `` → `[agent/Workflows/pr-review.md](pr-review.md)` (same dir)
- All INSTRUCT refs (lines 130, 337, 358, 365, 372, 437) and CODE_BLOCK refs: keep as backticks

**catalog/workflows/planning/planning.md**:
- Line 22: `` `agent/Skills/planning-template.md` `` → `[agent/Skills/planning-template.md](../Skills/planning-template.md)`
- Line 23: `` `Playbook/Backlog.md` `` → `[Playbook/Backlog.md](../../Playbook/Backlog.md)`
- Line 37: `` `agent/Skills/planning-template.md` `` → `[agent/Skills/planning-template.md](../Skills/planning-template.md)`

**catalog/workflows/reporting/reporting.md**:
- Line 18: `` `Reports/report-template.md` `` → `[Reports/report-template.md](../../Reports/report-template.md)`
- Lines 20, 34: INSTRUCT, keep backticks

**catalog/workflows/plan-execution/plan-execution.md**:
- Line 46: `` `agent/Workflows/reporting.md` `` → `[agent/Workflows/reporting.md](reporting.md)` (same dir)

### Step 4 — Core static files

Core files install to `agent/Core/`.

**catalog/core/self-awareness.md**:
- Line 21: `` `agent/Core/memory.md` `` → `[agent/Core/memory.md](memory.md)` (same dir)

### Step 5 — Routine templates

Routines install to `agent/Routines/`. Only convert NAVIGATE refs. All INSTRUCT refs (especially dashboard updates with `{{ .Workspace }}` or plain `agent/Core/routines.md`) stay as backticks.

**catalog/routines/memory-consolidation/memory-consolidation.md.tmpl**:
- Line 13: `` `agent/Core/memory.md` `` → `[agent/Core/memory.md](../Core/memory.md)`
- Line 22: `` `agent/Core/memory.md` `` → `[agent/Core/memory.md](../Core/memory.md)`
- Line 46: INSTRUCT, keep backtick

**catalog/routines/doc-freshness-check/doc-freshness-check.md.tmpl**:
- Line 27: Generic directory refs like `` `agent/Core/` `` — convert to `[agent/Core/](../Core/)`, `[agent/Protocols/](../Protocols/)`, etc.
- Line 34: INSTRUCT, keep backtick

All other routine files: only have INSTRUCT refs → no changes.

### Step 6 — Scaffolding templates

**catalog/scaffolding/INDEX.md.tmpl**:
INDEX.md installs at workspace root. All paths are already relative to workspace root.

Convert the document registry table entries (NAVIGATE context):
- `` `INDEX.md` `` stays as backtick (self-reference)
- `` `CLAUDE.md` `` → `[CLAUDE.md](CLAUDE.md)`
- `` `Playbook/Status.md` `` → `[Playbook/Status.md](Playbook/Status.md)`
- `` `Playbook/Roadmap.md` `` → `[Playbook/Roadmap.md](Playbook/Roadmap.md)`
- `` `Playbook/Backlog.md` `` → `[Playbook/Backlog.md](Playbook/Backlog.md)`
- `` `Playbook/Standards/SecurityStandards.md` `` → `[Playbook/Standards/SecurityStandards.md](Playbook/Standards/SecurityStandards.md)`
- `` `Logs/FieldNotes.md` `` → `[Logs/FieldNotes.md](Logs/FieldNotes.md)`
- `` `Logs/KeyDecisionLog.md` `` → `[Logs/KeyDecisionLog.md](Logs/KeyDecisionLog.md)`
- `` `Reports/report-template.md` `` → `[Reports/report-template.md](Reports/report-template.md)`
- `` `Playbook/Plans/Active/` `` → `[Playbook/Plans/Active/](Playbook/Plans/Active/)`
- `` `Reports/Pending/` `` → `[Reports/Pending/](Reports/Pending/)`

Line 47 hierarchy: `` `CLAUDE.md` `` → `[CLAUDE.md](CLAUDE.md)`, `` `agent/Core/` `` → `[agent/Core/](agent/Core/)`, etc.

**catalog/scaffolding/Playbook/Backlog.md.tmpl**:
Installs at `Playbook/Backlog.md`.
- Line 9-10: `` `Status.md` `` or `` `Playbook/Status.md` `` → `[Status.md](Status.md)` (same dir)
- Line 10: `` `Playbook/Roadmap.md` `` → `[Roadmap.md](Roadmap.md)` (same dir)
- Lines 18, 20: INSTRUCT, keep backticks

**catalog/scaffolding/Playbook/Status.md.tmpl**:
Installs at `Playbook/Status.md`.
- Line 9: `` `Plans/Active/` `` → `[Plans/Active/](Plans/Active/)` and `` `Plans/Archive/` `` → `[Plans/Archive/](Plans/Archive/)`

### Step 7 — Update tests

Run `go test ./...` and check `internal/generate/generate_test.go`. If tests assert on backtick format in CLAUDE.md output, update expected strings to match the new markdown link format.

## Dependencies

None — self-contained change.

## Security

> [!warning]
> Refer to `Playbook/Standards/SecurityStandards.md` for all security requirements.

- No user input involved in link generation (all paths are deterministic from catalog)
- No risk of path traversal (relative paths resolve within workspace)
- No secrets or credentials affected

## Verification

- [ ] `make build` — passes
- [ ] `go test ./...` — passes (including updated test assertions)
- [ ] CLI smoke test in temp dir: `bonsai init` → check generated CLAUDE.md has clickable markdown links, not backticks
- [ ] Open generated workspace in Obsidian — links click through to correct files
- [ ] Verify INSTRUCT/CODE_BLOCK references still use backticks (not converted)
- [ ] Verify relative paths are correct from each file depth (agent/{Category}/ → ../../ for workspace root, ../X/ for cross-category)
- [ ] `bonsai catalog` still renders properly
- [ ] `bonsai update` still works (lock file hashes will differ — expected)
