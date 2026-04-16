# Plan 08 — Better Trigger Sections

**Tier:** 2 (Feature)
**Status:** Ready (verified — see Verification Notes at bottom)
**Source:** Roadmap P1 + RESEARCH-trigger-system.md (v2)
**Phases:** A (core mechanics), B (documentation), C (new sensors)

---

## Goal

Make ability activation consistent, documented, and as deterministic as possible. Add trigger metadata to the catalog, generate Claude Code native rules/skills files for mechanical activation, enrich CLAUDE.md with scenario-based tables and a quick-triggers reference, add per-ability trigger sections, and create human-facing trigger documentation.

### Success Criteria

- Skills auto-load when Claude reads matching files (via `.claude/rules/`)
- Key workflows are invocable as `/slash-commands` and auto-invocable by description (via `.claude/skills/`)
- CLAUDE.md tables show "Activate when..." scenarios instead of bare descriptions
- Each ability file has a trigger section explaining how to activate it
- Users have a `docs/triggers.md` guide documenting all trigger mechanisms
- Compaction recovery re-injects critical context after context compaction
- Context-guard detects verification and planning phrases, not just wrap-up

---

## Phase A — Core Mechanics

Catalog schema changes, generator additions, CLAUDE.md enrichment, per-ability trigger sections.

### A1. Add Triggers struct to catalog (`internal/catalog/catalog.go`)

Add new types after `RoutineItem` (after line 100):

```go
// TriggerExample is a prompt-action pair showing how an ability activates.
type TriggerExample struct {
    Prompt string `yaml:"prompt"`
    Action string `yaml:"action"`
}

// Triggers holds activation metadata for skills and workflows.
type Triggers struct {
    Scenarios []string         `yaml:"scenarios,omitempty"`
    Examples  []TriggerExample `yaml:"examples,omitempty"`
    Paths     []string         `yaml:"paths,omitempty"`
}
```

Add `Triggers` field to `CatalogItem` (line 76, before `ContentPath`):

```go
Triggers    *Triggers   `yaml:"triggers,omitempty"`
```

No loader changes needed — `yaml.Unmarshal` handles optional fields automatically. Nil pointer means "no triggers" (backward compat).

### A2. Add `scenariosDesc()` helper (`internal/generate/generate.go`)

Add after `descFor()` (after line 113):

```go
// scenariosDesc returns a trigger-aware description for CLAUDE.md tables.
// Uses triggers.scenarios (joined with "; ") if available, falls back to description.
func scenariosDesc(item *catalog.CatalogItem) string {
    if item != nil && item.Triggers != nil && len(item.Triggers.Scenarios) > 0 {
        return strings.Join(item.Triggers.Scenarios, "; ")
    }
    if item != nil {
        return item.Description
    }
    return ""
}
```

### A3. Enrich CLAUDE.md tables (`internal/generate/generate.go` — `WorkspaceClaudeMD()`)

**Workflows table** (lines 594-604): Change header and use scenarios.

Current:
```go
"### Workflows (load when starting an activity)", "",
"| Activity | Read this |",
"|----------|-----------|",
```
New:
```go
"### Workflows (load when starting an activity)", "",
"| Activate when... | Read this |",
"|------------------|-----------|",
```

Change the row format to use `scenariosDesc()` instead of `wfDescs[w]`:
```go
for _, w := range installed.Workflows {
    desc := wfDescs[w]
    if item := cat.GetWorkflow(w); item != nil {
        if sd := scenariosDesc(item); sd != "" {
            desc = sd
        }
    }
    lines = append(lines, fmt.Sprintf("| %s | [agent/Workflows/%s.md](agent/Workflows/%s.md) |", desc, w, w))
}
```

**Skills table** (lines 607-617): Same pattern.

Current:
```go
"### Skills (load when doing specific work)", "",
"| Need | Read this |",
"|------|-----------|",
```
New:
```go
"### Skills (load when doing specific work)", "",
"| Activate when... | Read this |",
"|------------------|-----------|",
```

Row format with scenarios fallback:
```go
for _, s := range installed.Skills {
    desc := skillDescs[s]
    if item := cat.GetSkill(s); item != nil {
        if sd := scenariosDesc(item); sd != "" {
            desc = sd
        }
    }
    lines = append(lines, fmt.Sprintf("| %s | [agent/Skills/%s.md](agent/Skills/%s.md) |", desc, s, s))
}
```

### A4. Extract curated workflow set and add Quick Triggers table

**Package-level variable** (top of `internal/generate/generate.go`):

```go
// CuratedSlashWorkflows is the set of workflows that get .claude/skills/ files
// and appear in the Quick Triggers table. Keep this small — each entry consumes
// ~1,536 chars of context budget at session start.
var CuratedSlashWorkflows = map[string]bool{
    "planning": true, "code-review": true, "pr-review": true,
    "security-audit": true, "issue-to-implementation": true,
    "test-plan": true, "plan-execution": true,
}
```

**Quick Triggers function** — derives descriptions from catalog metadata, not a hardcoded map:

```go
func quickTriggersLines(installed *config.InstalledAgent, cat *catalog.Catalog) []string {
    var lines []string
    lines = append(lines,
        "### Quick Triggers", "",
        "> Common phrases and commands that activate specific behaviors.", "",
        "| You want to... | Say or do this |",
        "|----------------|---------------|",
        "| Start a session | \"Hi, get started\" |",
    )

    // Add workflow-based triggers from curated set — derive descriptions from catalog
    for _, w := range installed.Workflows {
        if !CuratedSlashWorkflows[w] {
            continue
        }
        desc := catalog.DisplayNameFrom(w)
        if item := cat.GetWorkflow(w); item != nil {
            if item.Triggers != nil && len(item.Triggers.Scenarios) > 0 {
                desc = item.Triggers.Scenarios[0]
            } else {
                desc = item.Description
            }
        }
        lines = append(lines, fmt.Sprintf("| %s | \"[describe task]\" or `/%s` |", desc, w))
    }

    lines = append(lines,
        "| Self-review before shipping | \"Verify everything\" |",
        "| End session | \"That's all\" |",
        "", "",
    )
    return lines
}
```

Call this in `WorkspaceClaudeMD()` after the Core table and before Protocols (insert between lines 579 and 581).

### A5. Generate path-scoped rules (`internal/generate/generate.go`)

Add new function:

```go
// PathScopedRules generates .claude/rules/skill-{name}.md files for skills
// that have triggers.paths defined. These auto-load when Claude reads matching files.
func PathScopedRules(projectRoot string, cfg *config.ProjectConfig, cat *catalog.Catalog, lock *config.LockFile, result *WriteResult, force bool) error {
    for _, installed := range cfg.Agents {
        for _, skillName := range installed.Skills {
            item := cat.GetSkill(skillName)
            if item == nil || item.Triggers == nil || len(item.Triggers.Paths) == 0 {
                continue
            }

            // Build content
            var lines []string
            lines = append(lines, "---")
            lines = append(lines, "paths:")
            for _, p := range item.Triggers.Paths {
                lines = append(lines, fmt.Sprintf("  - \"%s\"", p))
            }
            lines = append(lines, "---", "")
            lines = append(lines, fmt.Sprintf("When working with files matching these paths, load and follow the **%s** skill at `%sagent/Skills/%s.md`.",
                item.DisplayName, installed.Workspace, skillName))

            if len(item.Triggers.Scenarios) > 0 {
                lines = append(lines, "", "Activate when:")
                for _, s := range item.Triggers.Scenarios {
                    lines = append(lines, "- "+s)
                }
            }
            lines = append(lines, "")

            content := []byte(strings.Join(lines, "\n"))
            relPath := filepath.Join(installed.Workspace, ".claude", "rules", "skill-"+skillName+".md")
            r := writeFile(projectRoot, relPath, content, "generated:rule-skill-"+skillName, lock, force)
            result.Add(r)
        }
    }
    return nil
}
```

### A6. Generate Claude Code skills for workflows (`internal/generate/generate.go`)

Add new function. Uses the package-level `CuratedSlashWorkflows` set (7 workflows):

```go
// WorkflowSkills generates .claude/skills/{name}/SKILL.md files for curated
// workflows, enabling /slash-command invocation and description-based auto-invocation.
func WorkflowSkills(projectRoot string, cfg *config.ProjectConfig, cat *catalog.Catalog, lock *config.LockFile, result *WriteResult, force bool) error {

    for _, installed := range cfg.Agents {
        for _, wfName := range installed.Workflows {
            if !CuratedSlashWorkflows[wfName] {
                continue
            }
            item := cat.GetWorkflow(wfName)
            if item == nil {
                continue
            }

            var lines []string
            lines = append(lines, "---")
            lines = append(lines, fmt.Sprintf("name: %s", wfName))

            // Description from scenarios or fallback
            desc := item.Description
            if item.Triggers != nil && len(item.Triggers.Scenarios) > 0 {
                desc = strings.Join(item.Triggers.Scenarios, ". ")
            }
            lines = append(lines, fmt.Sprintf("description: %s", desc))

            // when_to_use from scenarios
            if item.Triggers != nil && len(item.Triggers.Scenarios) > 0 {
                lines = append(lines, fmt.Sprintf("when_to_use: %s", strings.Join(item.Triggers.Scenarios, "; ")))
            }

            lines = append(lines, "---", "")
            lines = append(lines, fmt.Sprintf("Load and follow the **%s** workflow at `%sagent/Workflows/%s.md`.",
                item.DisplayName, installed.Workspace, wfName))

            // Examples section
            if item.Triggers != nil && len(item.Triggers.Examples) > 0 {
                lines = append(lines, "", "## Examples", "")
                for _, ex := range item.Triggers.Examples {
                    lines = append(lines, fmt.Sprintf("> **User:** \"%s\"", ex.Prompt))
                    lines = append(lines, fmt.Sprintf("> **Action:** %s", ex.Action))
                    lines = append(lines, "")
                }
            }
            lines = append(lines, "")

            content := []byte(strings.Join(lines, "\n"))
            relPath := filepath.Join(installed.Workspace, ".claude", "skills", wfName, "SKILL.md")
            r := writeFile(projectRoot, relPath, content, "generated:skill-workflow-"+wfName, lock, force)
            result.Add(r)
        }
    }
    return nil
}
```

### A7. Generate trigger sections in ability files (`internal/generate/generate.go`)

Add a function that prepends a trigger section to skill/workflow content before writing:

```go
// triggerSection generates a markdown trigger header from catalog triggers metadata.
func triggerSection(item *catalog.CatalogItem, workspace string, category string, isSlashCommand bool) string {
    if item.Triggers == nil {
        return ""
    }
    t := item.Triggers

    var lines []string
    lines = append(lines, "## Triggers", "")

    if isSlashCommand {
        lines = append(lines, fmt.Sprintf("**Slash command:** `/%s`", item.Name))
    }

    if len(t.Paths) > 0 {
        lines = append(lines, fmt.Sprintf("**Auto-loads when reading:** %s", strings.Join(t.Paths, ", ")))
    }

    if len(t.Scenarios) > 0 {
        lines = append(lines, "**Activate when:**")
        for _, s := range t.Scenarios {
            lines = append(lines, "- "+s)
        }
    }

    if len(t.Examples) > 0 {
        lines = append(lines, "", "**Examples:**")
        for _, ex := range t.Examples {
            lines = append(lines, fmt.Sprintf("> **User:** \"%s\"", ex.Prompt))
            lines = append(lines, fmt.Sprintf("> **Action:** %s", ex.Action))
        }
    }

    lines = append(lines, "", "---", "")
    return strings.Join(lines, "\n")
}
```

Modify `AgentWorkspace()` — in the skills loop (line 1062-1074), prepend trigger section:

```go
// 2. Skills
for _, skillName := range installed.Skills {
    item := cat.GetSkill(skillName)
    if item == nil {
        continue
    }
    content, err := renderContent(catFS, item.ContentPath, fullCtx)
    if err != nil {
        return fmt.Errorf("skill %s: %w", skillName, err)
    }

    // Prepend trigger section if triggers exist
    if ts := triggerSection(item, installed.Workspace, "skill", false); ts != "" {
        content = append([]byte(ts), content...)
    }

    relPath, _ := filepath.Rel(projectRoot, filepath.Join(agentDir, "Skills", skillName+".md"))
    r := writeFile(projectRoot, relPath, content, "catalog:skills/"+skillName, lock, force)
    result.Add(r)
}
```

Same for workflows (line 1077-1089) — check if it's in the curated slash-command set:

```go
// 3. Workflows
for _, wfName := range installed.Workflows {
    item := cat.GetWorkflow(wfName)
    if item == nil {
        continue
    }
    data, err := fs.ReadFile(catFS, item.ContentPath)
    if err != nil {
        return err
    }

    // Prepend trigger section if triggers exist
    if ts := triggerSection(item, installed.Workspace, "workflow", CuratedSlashWorkflows[wfName]); ts != "" {
        data = append([]byte(ts), data...)
    }

    relPath, _ := filepath.Rel(projectRoot, filepath.Join(agentDir, "Workflows", wfName+".md"))
    r := writeFile(projectRoot, relPath, data, "catalog:workflows/"+wfName, lock, force)
    result.Add(r)
}
```

### A8. Wire generators into commands

**`cmd/init.go`** (line 185-188) — add after `AgentWorkspace`, before `SettingsJSON`:

```go
_ = generate.PathScopedRules(cwd, cfg, cat, lock, &wr, false)
_ = generate.WorkflowSkills(cwd, cfg, cat, lock, &wr, false)
```

**`cmd/add.go`** — two places:

1. New agent flow (line 213-214, after `AgentWorkspace`):
```go
_ = generate.PathScopedRules(cwd, cfg, cat, lock, &wr, false)
_ = generate.WorkflowSkills(cwd, cfg, cat, lock, &wr, false)
```

2. Add items flow (line 380-381, after `AgentWorkspace`):
```go
_ = generate.PathScopedRules(cwd, cfg, cat, lock, &wr, false)
_ = generate.WorkflowSkills(cwd, cfg, cat, lock, &wr, false)
```

**`cmd/update.go`** (line 159-161, after `AgentWorkspace`):
```go
_ = generate.PathScopedRules(cwd, cfg, cat, lock, &wr, false)
_ = generate.WorkflowSkills(cwd, cfg, cat, lock, &wr, false)
```

**`cmd/remove.go`** — two changes:

1. In `runRemoveItem()` (after line 307), add cleanup for rule/skill files when removing skills or workflows:

```go
// Clean up generated trigger files
if it.singular == "skill" {
    rulePath := filepath.Join(t.agent.Workspace, ".claude", "rules", "skill-"+name+".md")
    lock.Untrack(rulePath)
    _ = os.Remove(filepath.Join(cwd, rulePath))
}
if it.singular == "workflow" {
    skillDir := filepath.Join(t.agent.Workspace, ".claude", "skills", name)
    skillPath := filepath.Join(skillDir, "SKILL.md")
    lock.Untrack(filepath.Join(t.agent.Workspace, ".claude", "skills", name, "SKILL.md"))
    _ = os.Remove(filepath.Join(cwd, skillPath))
    _ = os.Remove(filepath.Join(cwd, skillDir)) // remove empty dir
}
```

2. In `runRemove()` (full agent removal, after line 127), add `.claude/` directory cleanup alongside existing `agent/` and `CLAUDE.md` cleanup:

```go
claudeDir := filepath.Join(cwd, agent.Workspace, ".claude")
if err := os.RemoveAll(claudeDir); err == nil {
    tui.Info("Deleted " + claudeDir)
}
```

### A9. Add trigger metadata to all catalog skills (17 items)

Update each `catalog/skills/*/meta.yaml` to add `triggers:` field. Content per skill is defined in the research doc — see "Layer 2: Path-Scoped Rules" table and "Layer 3: Skill Auto-Invocation" table.

**Skills WITH paths** (11 items): coding-standards, database-conventions, api-design-standards, testing, test-strategy, auth-patterns, container-standards, iac-conventions, cli-conventions, mobile-patterns, design-guide

**Skills WITHOUT paths** (6 items): planning-template, review-checklist, workspace-guide, issue-classification, dispatch, pr-creation

Each gets `scenarios:` (2-5 entries). Skills with file associations also get `paths:` (2-6 globs). Optional `examples:` for complex skills.

### A10. Add trigger metadata to all catalog workflows (10 items)

Update each `catalog/workflows/*/meta.yaml` to add `triggers:` field.

All workflows get `scenarios:` (2-4 entries) and `examples:` (1-2 prompt/action pairs). Workflows without natural file associations don't get `paths:`.

Items: planning, code-review, pr-review, security-audit, session-logging, test-plan, plan-execution, api-development, reporting, issue-to-implementation

> **Note:** `session-wrapup` is a custom workflow in the Bonsai dogfooding project, not a catalog item. It does not need trigger metadata in the catalog.

### A11. Tests (`internal/generate/generate_test.go`)

1. **`TestScenariosDescFallback`** — verify `scenariosDesc()` returns description when triggers nil, scenarios when present
2. **`TestScenariosDescJoinsScenarios`** — verify multiple scenarios are joined with "; "
3. **`TestClaudeMDUsesScenarios`** — build test catalog with triggers.scenarios, generate CLAUDE.md, assert "Activate when..." header and scenario text in table
4. **`TestPathScopedRulesGenerated`** — build test catalog with triggers.paths on a skill, call `PathScopedRules()`, assert rule file exists with correct `paths:` frontmatter
5. **`TestPathScopedRulesSkippedWhenNoPaths`** — skill without paths → no rule file generated
6. **`TestWorkflowSkillsGenerated`** — build test catalog with triggers on a curated workflow, call `WorkflowSkills()`, assert SKILL.md exists with correct content
7. **`TestWorkflowSkillsSkippedWhenNotCurated`** — non-curated workflow → no SKILL.md generated
8. **`TestTriggerSectionPrepended`** — generate a skill with triggers, verify the output file starts with "## Triggers"
9. **`TestBackwardCompatNilTriggers`** — catalog with no triggers → everything works as before, no crashes

---

## Phase B — Documentation

### B1. Create `docs/triggers.md`

Comprehensive user-facing guide covering:

1. **How triggers work** — the 5-layer stack explained for humans (hooks → path rules → skill invocation → nav tables → prompt phrases)
2. **Complete trigger reference** — table of every ability with activation methods (slash command, path auto-load, phrases)
3. **Session workflow** — the 4-beat rhythm (ORIENT → COMMIT → VERIFY → SHIP) mapped to trigger mechanisms
4. **Slash commands** — list of available `/commands` with descriptions
5. **Path-based auto-loading** — which skills load for which file types
6. **Customization** — how to add custom triggers (edit meta.yaml, add .claude/rules/)
7. **Troubleshooting** — common issues: ability didn't activate, wrong ability loaded, how to check

The guide should also cover:
- **User customization:** Users can add their own `.claude/rules/` and `.claude/skills/` files alongside generated ones. These won't be tracked by the lock file or overwritten on `bonsai update`. Document this as the recommended customization path.
- **Version control:** Generated `.claude/rules/` and `.claude/skills/` files should be committed to version control — they are project-specific configuration, not user-local state.

### B2. Update `docs/working-with-agents.md`

- Add "Slash Commands" section in the Quick Reference area
- Cross-link to `docs/triggers.md` for complete reference
- Add note about path-scoped auto-loading in the session flow description
- Update "The Five Essential Phrases" table to include `/command` alternatives

### B3. Update README

Add `docs/triggers.md` to the guides table in the README.

---

## Phase C — New Sensors

### C1. Compact recovery sensor

**Catalog:** `catalog/sensors/compact-recovery/`
- `meta.yaml`: name: compact-recovery, event: SessionStart, matcher: compact, agents: all, required: all
- `compact-recovery.sh.tmpl`: Script that re-injects abbreviated context after compaction

**Script behavior:**
1. Read `{{.Workspace}}CLAUDE.md` — extract only the Quick Triggers table and nav table headers
2. Read `{{.Workspace}}agent/Core/memory.md` — extract Work State section only
3. Output as `additionalContext` JSON — keep under 2000 chars total
4. Exit 0

This sensor auto-installs for all agents (required: all). It fires only on `SessionStart` with matcher `compact` — meaning it only runs after context compaction, not on regular startup.

### C2. Context-guard expansion

Expand `catalog/sensors/context-guard/context-guard.sh.tmpl` with additional trigger categories.

**Verification triggers** (add after wrap-up patterns, same normalize+end-anchor approach):

```python
verify_patterns = [
    r'\bverify\s+(everything|it\s+all)',
    r'\bcheck\s+(your|the)\s+work',
    r'\bcheck\s+if\s+you\s+missed',
    r'\breview\s+(your|the)\s+changes',
    r'\breview\s+before\s+(commit|push|ship)',
    r'\bdoes\s+everything\s+look\s+(right|good)',
]
```

**Effect:** Inject verification checklist:
```
VERIFICATION REQUESTED. Before proceeding:
1. Re-read your own changes — check for bugs, edge cases, regressions
2. Verify all tests pass (if applicable)
3. Check for stale references in documentation
4. Confirm no security issues introduced
```

**Planning triggers:**

```python
plan_patterns = [
    r'\b(lets|let\s+us)\s+plan',
    r'\bplan\s+(this|the|a)\b',
    r'\bcreate\s+a\s+plan',
    r'\bdesign\s+(this|the|a)\b',
    r'\barchitect\s+(this|the|a)\b',
]
```

**Effect:** Inject reminder:
```
PLANNING DETECTED. Load the planning workflow at {workspace}agent/Workflows/planning.md and the planning-template skill at {workspace}agent/Skills/planning-template.md before proceeding.
```

**Note:** These are **not** end-anchored like wrap-up triggers. Planning/verification phrases typically appear at the start or middle of a prompt, not the end. Use word-boundary anchoring (`\b`) only.

### C3. Prompt hook for intent classification — DEFERRED

> **Status:** Deferred to post-Phase C. Ship auto-invocation (A6) first, measure whether the curated workflows activate reliably via `.claude/skills/` descriptions. If they do, prompt hooks are unnecessary complexity. If they don't, revisit.

The 3 target workflows (code-review, pr-review, security-audit) are all in the curated slash command set and already get auto-invocation via skill descriptions. Adding a ~$0.001/prompt Haiku evaluation on top adds a second detection layer but may not provide enough marginal value. Additionally, wiring prompt hook classification output into context injection (making one hook's output available to another) has uncertain implementation complexity.

**Backlog item:** If auto-invocation proves insufficient after Phase A ships, create a P2 backlog item to explore prompt hooks.

---

## Verification

### Build & Test

- [ ] `make build` — compiles with no errors
- [ ] `go test ./...` — all existing + new tests pass
- [ ] `gofmt -s -l .` — no formatting issues

### Manual Testing — Phase A

- [ ] `mkdir /tmp/test-a && cd /tmp/test-a && /path/to/bonsai init`
  - [ ] CLAUDE.md has "Activate when..." table headers (not "Need" or "Activity")
  - [ ] CLAUDE.md has Quick Triggers table
  - [ ] `.claude/rules/skill-*.md` files exist for skills with paths (e.g., skill-coding-standards.md)
  - [ ] `.claude/skills/*/SKILL.md` files exist for curated workflows (planning, code-review, pr-review, security-audit, issue-to-implementation, test-plan)
  - [ ] Skill files have "## Triggers" section at top
  - [ ] Workflow files have "## Triggers" section at top (with `/slash-command` for curated set)
- [ ] `bonsai add` adds new rules/skills files
- [ ] `bonsai remove skill coding-standards` → removes `.claude/rules/skill-coding-standards.md`
- [ ] `bonsai remove workflow code-review` → removes `.claude/skills/code-review/`
- [ ] `bonsai update` regenerates rules/skills files
- [ ] Backward compat: existing `.bonsai.yaml` without triggers → generation works, falls back to descriptions

### Manual Testing — Phase C

- [ ] After compaction → compact-recovery sensor fires, abbreviated context re-injected
- [ ] User says "verify everything" → verification checklist injected
- [ ] User says "let's plan the caching layer" → planning workflow pointer injected
- [ ] User says "check your work before we commit" → verification checklist injected
- [ ] False negative check: "that's all I need to plan" → should NOT trigger planning (it's mid-sentence)

---

## Security

> [!warning]
> Refer to SecurityStandards.md for all security requirements.

- Generated `.claude/rules/` and `.claude/skills/` files contain no secrets
- File paths in `triggers.paths` are globs, not executed — no injection risk
- Lock file tracking prevents unauthorized modification detection
- Prompt hook prompt is read-only classification — no tool access, no side effects

---

## Dependencies

- Existing catalog items must have valid `meta.yaml` — they do
- No external dependencies added
- Prompt hooks require Claude Code API access (already available in all environments)

---

## Dispatch

| Phase | Agent | Isolation | Notes |
|-------|-------|-----------|-------|
| A | general-purpose | worktree | Core Go changes + catalog meta.yaml updates |
| B | general-purpose | worktree | Documentation only — no Go changes |
| C | general-purpose | worktree | Sensor scripts + settings generator update |

---

## Verification Notes

Plan verified by two independent agents (2026-04-16). Issues found and resolved:

| Issue | Source | Resolution |
|-------|--------|------------|
| Full agent removal doesn't clean `.claude/` dir | Correctness agent | Added `.claude/` cleanup to A8 remove.go section |
| `session-wrapup` listed in A10 but not a catalog item | Intent agent | Removed from A10 list with explanatory note |
| Curated set hardcoded in two places (maintenance hazard) | Both agents | Extracted to package-level `CuratedSlashWorkflows` variable |
| Quick Triggers descriptions should come from catalog | Intent agent | Rewritten `quickTriggersLines()` to derive from triggers.scenarios |
| `plan-execution` missing from curated set | Intent agent | Added as 7th curated workflow |
| C3 (prompt hooks) uncertain implementation | Both agents | Deferred — auto-invocation via .claude/skills/ covers same workflows |
| User customization path undefined | Intent agent | Added to B1 documentation scope |
| Generated .claude/ files should be committed | Intent agent | Added to B1 documentation scope |
| Task-based skills (.claude/skills/) not in Phase A | Intent agent | Noted as future backlog item (post Phase C) |
