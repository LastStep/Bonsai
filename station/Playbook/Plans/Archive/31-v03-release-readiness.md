---
tags: [plan, v0.3, release, tier-2]
description: v0.3 release readiness bundle — correctness + agent-readable foundation + UX polish + CI mode.
status: Complete — shipped 2026-04-24 (PRs #75 `48e2678`, #76 `1af9402`, #77 `467b540`)
---

# Plan 31 — v0.3 Release Readiness Bundle

**Tier:** 2
**Status:** Draft
**Agent:** tech-lead (orchestrator) + general-purpose (code) + code-review + security-review

## Goal

Ship v0.3.0 as a hardening release. Three themes: (1) fix the cross-agent peer-awareness staleness bug, (2) make Bonsai agent-consumable so user's agent can self-customize workspaces without being hand-fed docs, (3) finish the cinematic TUI rollout so all commands feel cohesive. All items land across two PRs; release pipeline runs last.

## Context

### What prompted this

User session 2026-04-24: "finalize UI/UX for v0.3 release, no bugs, fully finish redesign." Dogfood confirmed `add` flow is clean — no user-facing bugs surfaced in tech-lead → backend → devops sequence. Deep-dive on the `OtherAgents` template revealed silent correctness bug (below).

### Why each item matters

1. **Peer-awareness staleness** — Templates in `catalog/sensors/scope-guard-files/*.sh.tmpl`, `catalog/sensors/dispatch-guard/*.sh.tmpl`, and all 6 agent `catalog/agents/*/core/identity.md.tmpl` files use `{{ range .OtherAgents }}`. `cmd/add.go` only re-renders the newly-added/augmented agent's workspace (`generate.AgentWorkspace(..., installedAgent, ...)` at lines 460 + 515). Already-installed peers never get their `OtherAgents` list refreshed. Net effect: tech-lead's `scope-guard-files.sh` may have NO `# Block writes to backend/` entry even after `bonsai add backend` — scope-guard silently fails open. Same with `dispatch-guard.sh`'s workspace→agent map. Silent correctness bug, user can't see, blame falls on their agent. Backlog item `[debt] Cross-agent OtherAgents template staleness on bonsai add` (added 2026-04-22) tracks this — promoted to release blocker.

2. **Agent-consumable foundation** — User pattern: create project → `bonsai init` → tell agent about project → agent recommends customizations. Today the agent has zero mental model of Bonsai. Gap matches pi.dev's approach: their system prompt points to `README.md` at a known path, agent reads on-demand. Bonsai needs the same: (a) a `bonsai-model` skill at `station/agent/Skills/bonsai-model.md` documenting the mental model, (b) `.bonsai/catalog.json` snapshot at workspace root for filesystem-discovery of available abilities, (c) pointer lines in generated `CLAUDE.md` so the agent knows where to look.

3. **Cinematic coverage** — Plans 22/23/27/28/29/30 ported `init`/`add`/`list`/`catalog`/`guide` to dedicated flow packages. Two commands still run via raw harness without a flow package: `bonsai remove` (625L, `cmd/remove.go`) and `bonsai update` (313L, `cmd/update.go`). Cohesion argument: user sees `init → add → update/remove` feel different from each other.

4. **CI/scripting mode** — `RESEARCH-uiux-overhaul.md` §4.2 flagged "No CI/Non-Interactive Mode" as Critical. Agent-consumption needs `bonsai catalog --json`. Also verify `NO_COLOR` honored (research §4.3 flagged Critical, unverified).

5. **Hints 3-layer overhaul** — Post-flow hints today are thin. New contract: (a) mechanical next CLI command, (b) workflow next step inside workspace, (c) copy-paste AI-prompt snippet. Drives user-agent onboarding per pi.dev pattern. Catalog-driven `hints.yaml` per agent type keeps it extensible.

### Dependencies

- Plan 28's `RenderHeader(version, projectDir, action, rightLabel, width, safe)` infrastructure — used by all new flow packages
- Plan 28's `StageContext.HeaderAction` + `HeaderRightLabel` fields with `ApplyContextHeader(ctx)` setter
- Plan 15's harness step/reducer primitives (`NewLazy`, `NewLazyGroup`, `NewConditional`, `NewSpinner`)
- Plan 27 Phase B2's `*ForAgent` scoping variants (`PathScopedRulesForAgent`, `WorkflowSkillsForAgent`, `SettingsJSONForAgent`)

### Out of scope (defer to v0.4)

- `tune-workspace` workflow (new headline feature — v0.4)
- `custom-item-creator` CLI/skill (v0.4 prerequisite)
- `/repair-bonsai` skill (v0.4)
- Plan-29 carry-over items (cosmetic/test-gap/security-hardening) — filed, non-blocking
- Demo GIF/asciinema (needs user recording)
- Homebrew PAT rotation calendar (ops, tracked in Backlog P1)

---

## PR Split

### PR1 — Correctness + Agent-Readable Foundation (phases A–D)

Single `general-purpose` agent, `isolation: worktree` off `main`. Phases sequential within one agent (A and C both touch `internal/generate/generate.go` + `cmd/add.go`).

### PR2 — UX Polish + CI Mode (phases E–H, parallel)

Two `general-purpose` agents in parallel, file-disjoint per memory feedback:
- **Agent α** — E (remove cinematic) + G (catalog --json + NO_COLOR)
- **Agent β** — F (update cinematic) + H (hints 3-layer overhaul)

Second merge rebases on first.

---

## Phase A — Peer-Awareness Silent Refresh (PR1)

**Why:** Silent correctness bug — already-installed peers never re-render their `OtherAgents`-using files after `bonsai add`.

**Files:**
- `internal/generate/generate.go` — add new function
- `cmd/add.go` — add call site in both new-agent + add-items branches
- `internal/generate/generate_test.go` — new unit tests

**Steps:**

1. **New function `RefreshPeerAwareness`** in `internal/generate/generate.go`:
   ```go
   // RefreshPeerAwareness re-renders the three OtherAgents-dependent files
   // (scope-guard-files.sh, dispatch-guard.sh, identity.md) for every installed
   // agent EXCEPT the one specified. Called after `bonsai add` so already-
   // installed peers pick up the newly-added agent in their awareness blocks.
   //
   // Lock-aware: user-edited copies trigger the standard conflict flow.
   // Agents without the relevant sensor/file are skipped silently (e.g. an
   // agent that didn't install dispatch-guard just has that file skipped).
   func RefreshPeerAwareness(projectRoot string, excludeAgent string, cfg *config.ProjectConfig, cat *catalog.Catalog, lock *config.LockFile, result *WriteResult, force bool) error {
       // For each installed agent !== excludeAgent:
       //   - If scope-guard-files in agent.Sensors → re-render ONLY that sensor
       //   - If dispatch-guard in agent.Sensors → re-render ONLY that sensor
       //   - Always re-render identity.md (all agents have one)
       // Uses existing template context builder; scoping is on the file-list
       // side, not the template side.
   }
   ```

   Implementation detail: reuse the existing sensor-rendering helper by extracting a per-file render path if needed. Do NOT duplicate template loading logic. File targets per agent:
   - `<workspace>/agent/Sensors/scope-guard-files.sh` (if installed)
   - `<workspace>/agent/Sensors/dispatch-guard.sh` (if installed)
   - `<workspace>/agent/Core/identity.md` (always)

2. **Call site A** in `cmd/add.go:460-468` (add-items branch, after existing `SettingsJSONForAgent`):
   ```go
   errs = append(errs, generate.RefreshPeerAwareness(cwd, installedAgent.AgentType, cfg, cat, lock, wr, false))
   ```

3. **Call site B** in `cmd/add.go:515-520` (new-agent branch, after existing `SettingsJSONForAgent`):
   ```go
   errs = append(errs, generate.RefreshPeerAwareness(cwd, installed.AgentType, cfg, cat, lock, wr, false))
   ```

4. **Tests** (new cases in `internal/generate/generate_test.go` or dedicated `refresh_peer_awareness_test.go`):
   - `TestRefreshPeerAwareness_UpdatesSiblingScopeGuard` — install tech-lead + backend, call `RefreshPeerAwareness(..., "backend", ...)`, assert tech-lead's `scope-guard-files.sh` contains `backend/` block entry.
   - `TestRefreshPeerAwareness_SkipsExcludedAgent` — call with `excludeAgent="backend"`, assert backend's files are NOT rewritten (check mtime or WriteResult).
   - `TestRefreshPeerAwareness_SkipsAgentsMissingSensor` — agent without dispatch-guard installed doesn't error.
   - `TestRefreshPeerAwareness_TrackedInWriteResult` — updated files appear in `result.Updated`.

**Verification:**
- [ ] Unit tests green
- [ ] Manual: fresh dir, `bonsai init` (tech-lead), `bonsai add backend`, grep `station/agent/Sensors/scope-guard-files.sh` for `backend/` → must exist.
- [ ] Manual: grep `station/agent/Sensors/dispatch-guard.sh` for `'backend/':` workspace entry → must exist.

---

## Phase B — `bonsai-model` Skill Registration (PR1)

**Why:** User's agent needs a stable mental model of Bonsai to reason about customization.

**Files already authored** by tech-lead pre-dispatch (committed alongside plan):
- `catalog/skills/bonsai-model/meta.yaml`
- `catalog/skills/bonsai-model/bonsai-model.md`

**Steps (for PR1 agent):**

1. **Add `bonsai-model` to default skills** for tech-lead agent only in `catalog/agents/tech-lead/agent.yaml`:
   ```yaml
   defaults:
     skills:
       - ...existing...
       - bonsai-model
   ```
   Confirm by running `bonsai catalog -a tech-lead` pre/post — `bonsai-model` should appear in Skills.

2. **Verify skill file embed:** no code changes — catalog uses `embed.FS` at repo root, new files auto-included.

3. **Regenerate station workspace** via `bonsai add` on tech-lead (in test fixture) to verify file writes correctly.

**Verification:**
- [ ] `bonsai catalog` lists `bonsai-model` under Skills
- [ ] Fresh `bonsai init` with tech-lead lands `station/agent/Skills/bonsai-model.md`
- [ ] File size reasonable (200–400 lines, not a 2k-line dump)

---

## Phase C — `.bonsai/catalog.json` Snapshot (PR1)

**Why:** Filesystem-discoverable catalog — matches pi's `.pi/extensions/` convention. Agent can read ability listings without invoking CLI.

**Files:**
- `internal/generate/generate.go` — new function + called from init/add/update
- `cmd/init_flow.go` — call site (post-generate success)
- `cmd/add.go` — call site (both branches, after other generate calls)
- `cmd/update.go` — call site (post-spinner success)
- `internal/generate/generate_test.go` — unit test

**Steps:**

1. **New function `WriteCatalogSnapshot`** in `internal/generate/generate.go`:
   ```go
   // WriteCatalogSnapshot serializes the full catalog to .bonsai/catalog.json at
   // the project root. Agent-consumable — provides a filesystem-discoverable
   // listing of every agent/skill/workflow/protocol/sensor/routine currently
   // shipped by the installed Bonsai binary.
   //
   // Shape is stable JSON:
   //   { "version": "<binary-version>",
   //     "agents": [{"name":"...", "display_name":"...", "description":"..."}],
   //     "skills": [{"name":"...", "display_name":"...", "description":"...", "agents":[...], "required":[...]}],
   //     "workflows": [...], "protocols": [...], "sensors": [...], "routines": [...]
   //   }
   //
   // Not lock-tracked (regenerated on every init/add/update; no user edits
   // expected). Parent `.bonsai/` directory created if missing.
   func WriteCatalogSnapshot(projectRoot string, version string, cat *catalog.Catalog, result *WriteResult) error
   ```

2. **Define stable JSON structs** (not reusing `catalog` internal types — those may change). Add to `internal/generate/catalog_snapshot.go` (new file) or inline. Ship with JSON tags.

3. **Call sites:**
   - `cmd/init_flow.go` — after existing `generate.*` calls in generate stage's action closure, call `generate.WriteCatalogSnapshot(cwd, Version, cat, wr)`. Error aggregated via `errors.Join`.
   - `cmd/add.go:468` and `cmd/add.go:520` — same pattern after `RefreshPeerAwareness`.
   - `cmd/update.go` spinner closure — append after existing `generate.SettingsJSON(...)`.

4. **Test** `TestWriteCatalogSnapshot_RoundTrip`:
   - Write snapshot to temp dir
   - Unmarshal back
   - Assert at minimum: `agents[].name` contains tech-lead, `skills[].name` contains planning-template, `routines[].frequency` field populated for e.g. backlog-hygiene.

**Verification:**
- [ ] `.bonsai/catalog.json` written on every init/add/update
- [ ] File is valid JSON (`jq . .bonsai/catalog.json`)
- [ ] Categories non-empty after fresh init
- [ ] Size reasonable (<100KB — current catalog is small)

---

## Phase D — Generated `CLAUDE.md` Pointer Lines (PR1)

**Why:** Pi-style system-prompt pointer. Tells user's agent where to find Bonsai's mental model + catalog snapshot.

**Files:**
- `catalog/agents/tech-lead/core/workspace-claude.md.tmpl` (or wherever the generated `station/CLAUDE.md` template lives — confirm exact path; likely under `catalog/agents/<type>/` or shared under `catalog/core/`)
- Other 5 agent types' workspace-claude templates (if separately templated)

**Steps:**

1. **Locate generated CLAUDE.md template.** Reads as `station/CLAUDE.md` post-init. Source may be in `catalog/core/workspace-claude.md.tmpl` (shared) or per-agent under `catalog/agents/<type>/core/`. Grep for existing content (e.g. "Quick Triggers") to find template.

2. **Add pointer block** near top of CLAUDE.md template, after the `agent/Core` navigation table, before Quick Triggers. Exact copy:

   ```markdown
   ---

   ## Bonsai Reference

   > Agents working in this workspace: read these when reasoning about Bonsai itself (what catalog items exist, how to customize, what `bonsai add`/`remove`/`update` do).

   | Need | Read |
   |------|------|
   | Bonsai mental model — catalog shape, customization, decisions | [agent/Skills/bonsai-model.md](agent/Skills/bonsai-model.md) |
   | Available abilities (agent/skills/workflows/protocols/sensors/routines) | [../.bonsai/catalog.json](../.bonsai/catalog.json) |
   | Current installed state | [../.bonsai.yaml](../.bonsai.yaml) |
   ```

   (Adjust relative paths based on template context — if template is rendered per agent, `../.bonsai/catalog.json` may need `{{ .DocsPath }}`-aware logic.)

3. **If only tech-lead installs `bonsai-model`:** pointer-block must be template-conditional — other agent types don't have the skill locally. Two options:
   - (a) Only render the bonsai-model row for tech-lead — use `{{ if eq .AgentType "tech-lead" }}` wrapping that row.
   - (b) Reference the tech-lead workspace: for non-tech-lead agents, point to `../station/agent/Skills/bonsai-model.md`.

   Decision: (b) is cleaner — Bonsai model is project-level knowledge, not per-agent. Every agent's CLAUDE.md points to tech-lead's copy. Uses `{{ .DocsPath }}` (which is tech-lead's workspace — typically `station/`). Template computes path as `{{ .DocsPath }}agent/Skills/bonsai-model.md`.

4. **Verify template-context vars** include `DocsPath` and `AgentType`. Grep `internal/generate/generate.go` for the struct — should be `TemplateContext` with those fields.

**Verification:**
- [ ] Fresh `bonsai init` produces `station/CLAUDE.md` with the new reference block
- [ ] `bonsai add backend` produces `backend/CLAUDE.md` with pointer to `../station/agent/Skills/bonsai-model.md`
- [ ] Markdown renders correctly (obsidian-markdown compatible, no broken wikilink syntax)

---

## Phase E — `bonsai remove` Cinematic Port (PR2 Agent α)

**Why:** Last destructive command without cinematic TUI. `cmd/remove.go` (625L) currently uses raw harness with `tui.FatalPanel`, `tui.ErrorDetail`, `tui.Heading`, etc.

**Files:**
- New package: `internal/tui/removeflow/` with:
  - `removeflow.go` — package entry, shared types, stage labels
  - `observe.go` — show what will be removed (preview panel)
  - `confirm.go` — confirmation prompt (chromeless)
  - `conflicts.go` — per-file conflict resolver (reuse `addflow/conflicts.go` pattern)
  - `yield.go` — success summary + hints 3-layer block (depends on Phase H; agent α implements yield skeleton with 2-layer hints initially; PR2 merge sequencing handles 3-layer)
  - corresponding `_test.go` files
- Rewrite: `cmd/remove.go` — replace both `runRemove` (agent removal) and `runRemoveItem` (item removal) bodies with `removeflow` wiring
- Potentially split into `cmd/remove.go` (entry point, cobra wiring) + delegation to `removeflow.Run(...)`

**Steps:**

1. **Rail stages** (2 modes of flow — agent-remove vs item-remove; share stages where possible):
   - Agent-remove: 観 OBSERVE → 確 CONFIRM (chromeless) → 衝 CONFLICT (chromeless, lazy) → 結 YIELD
   - Item-remove: 択 SELECT (agent picker, conditional) → 観 OBSERVE → 確 CONFIRM → 衝 CONFLICT → 結 YIELD

   StageLabels shared in `removeflow.go` per `addflow.StageLabels` pattern. Reuse kanji + English conventions.

2. **Header context:** `HeaderAction: "REMOVE"`, `HeaderRightLabel: "UPROOTING FROM"` (sticking with the plant metaphor). Stamped in `StageContext` at ctor time.

3. **Conflict stage:** copy `internal/tui/addflow/conflicts.go` wholesale — same per-file 3-way radio (Keep/Overwrite/Backup), lowercase `k/o/b` no-op, uppercase `K/O/B` batch. Avoid divergence.

4. **Yield stage:** success panel with removed-count + file-tree of untracked files. Hints block placeholder (Phase H integration lands during merge).

5. **cmd/remove.go rewrite:**
   - Keep cobra wiring + flag parsing
   - Keep `typeSkill` / `typeWorkflow` / etc. itemType descriptors (pure data)
   - Keep helper functions that are business logic (`filterRequired`, `agentItemList`, `itemIsRequired`, etc. — these aren't UI)
   - Replace harness-Run calls with `removeflow.NewAgentRemoveStage(cfg, agentName, cat, lock)` / `removeflow.NewItemRemoveStage(cfg, name, it, cat, lock, matches)` returning a BubbleTea model
   - Run via `tea.NewProgram(stage, tea.WithAltScreen()).Run()`
   - Non-TTY fallback via `removeflow.RenderStatic(cfg, agentName, cat)` — simple text preview + inherit pre-cinematic confirm-by-flag if possible (or error out saying `--yes` flag needed, to be filed as followup)

6. **Tests:**
   - `removeflow_test.go` — BubbleTea model update unit tests (focus move, confirm, cancel)
   - `conflicts_test.go` — mirror `addflow/conflicts_test.go`
   - `TestObserve_ShowsInstalledCounts` — install agent with 3 skills + 2 workflows, assert panel shows both groupings
   - `TestConfirm_EscAbortsRemoval` — no mutation on cancel
   - `TestYield_ShowsRemovedPaths` — post-success file list

**Verification:**
- [ ] `make build && go test ./...` green
- [ ] Manual: `bonsai remove backend` in multi-agent fixture — cinematic flow, removes cleanly, no regression in `cfg.Save` or `SettingsJSON`
- [ ] Manual: `bonsai remove skill coding-standards` — picker fires, removes, lockfile clean

---

## Phase F — `bonsai update` Cinematic Port (PR2 Agent β)

**Why:** Last mutating command without cinematic. `cmd/update.go` (313L) uses raw harness + stdout-warnings pre-harness.

**Files:**
- New package: `internal/tui/updateflow/` with:
  - `updateflow.go` — package entry, shared types
  - `discover.go` — custom-file scan stage (renders discovered files per agent in a preview panel)
  - `select.go` — per-agent multi-select (chromeless, tab-strip over agents)
  - `sync.go` — spinner-equivalent (progress indication during re-render)
  - `conflicts.go` — reuse `addflow/conflicts.go` pattern
  - `yield.go` — success summary (was-up-to-date vs changes-synced) + hints
  - `_test.go` counterparts
- Rewrite: `cmd/update.go` — delegate to `updateflow.Run(cwd, cfg, cat, lock)`

**Steps:**

1. **Rail stages:** 探 DISCOVER → 択 SELECT (chromeless, per-agent tabs) → 同 SYNC → 衝 CONFLICT → 結 YIELD

2. **Header:** `HeaderAction: "UPDATE"`, `HeaderRightLabel: "SYNCING"`.

3. **Discover stage:** replaces the pre-harness stdout warnings in `cmd/update.go:68-75` — move invalid-file warnings into a dedicated panel within the Discover stage. Better visibility than scrolling stdout.

4. **Select stage:** chromeless, per-agent tab strip. Each tab = one agent with discovered files. Arrow keys cycle agents, space toggles file selection, `↵` advances. Default: all files selected (mirror legacy).

5. **Sync stage:** visible progress indicator. Show each agent being re-rendered as a bullet (✓ tech-lead synced · ✓ backend synced · …). On completion, advance to conflicts or yield.

6. **Yield stage:** two modes — "up to date" panel (no changes, mimics current `tui.TitledPanel("Up to date", ...)`) vs "synced" panel (file counts created/updated/conflicts). Hints 3-layer integration in merge.

7. **cmd/update.go rewrite:**
   - Drop stdout warnings pre-harness (moved into Discover stage)
   - `runUpdate` body → `updateflow.Run(cwd, cfg, cat, lock, Version)` returning a summary the caller prints or handles
   - Non-TTY fallback: `updateflow.RunStatic(...)` — auto-select all discovered files, no conflict prompt (conflicts become errors), basic success line

8. **Tests:**
   - `TestDiscover_InvalidFilesSurface` — file with bad frontmatter shows error panel
   - `TestSelect_PerAgentSelection` — switch tabs, different agents' files
   - `TestSync_ErrorAggregates` — generator error visible in yield
   - `TestYield_UpToDateWhenNoDiscoveriesNoConflicts`

**Verification:**
- [ ] `make build && go test ./...` green
- [ ] Manual: `bonsai update` on clean workspace — up-to-date panel
- [ ] Manual: drop a custom skill into `station/agent/Skills/custom.md` with proper frontmatter, `bonsai update` — discover stage flags it, accept → lockfile tracks it

---

## Phase G — `bonsai catalog --json` + NO_COLOR Audit (PR2 Agent α)

**Why:** Agent-consumable catalog via stdout. NO_COLOR honoring verification per research §4.3.

**Files:**
- `cmd/catalog.go` — add `--json` flag, bypass TUI when set
- `internal/tui/styles.go` — NO_COLOR check (may already be handled by lipgloss)
- `internal/tui/styles_test.go` — add NO_COLOR-path test

**Steps:**

1. **`--json` flag:** add to `catalogCmd`:
   ```go
   catalogCmd.Flags().Bool("json", false, "Output catalog as JSON (agent-consumable, non-interactive)")
   ```

2. **Branch in `runCatalog`:** before the TTY path, check the flag. If set:
   ```go
   if jsonOut, _ := cmd.Flags().GetBool("json"); jsonOut {
       return renderCatalogJSON(cat, agentFilter)  // writes to stdout
   }
   ```
   Reuse `WriteCatalogSnapshot`'s serialization via a `SerializeCatalog(cat, version) ([]byte, error)` helper in `internal/generate/catalog_snapshot.go`. One source of truth for JSON format.

3. **NO_COLOR audit:**
   - Grep `internal/tui/styles.go` for existing env checks.
   - Verify lipgloss's default `termenv.NO_COLOR` honoring fires. Add explicit test: set env, render a Colored() string, assert no ANSI escapes.
   - If lipgloss doesn't honor automatically: add explicit check at palette init — if `os.Getenv("NO_COLOR") != ""` → downgrade all AdaptiveColor to plain foreground.
   - TERM=dumb same audit.

4. **Tests:**
   - `TestRenderCatalogJSON_SchemaStable` — marshal, unmarshal, assert struct round-trips
   - `TestRenderCatalogJSON_AgentFilter` — `-a backend` excludes incompatible skills from output
   - `TestStyles_NoColorStripsAnsi` — `t.Setenv("NO_COLOR", "1")`, render styled string, assert no `\x1b[` sequences

**Verification:**
- [ ] `bonsai catalog --json | jq .skills[0]` returns a valid object
- [ ] `NO_COLOR=1 bonsai list | cat` contains no ANSI escapes
- [ ] `TERM=dumb bonsai catalog | cat` contains no ANSI escapes
- [ ] Tests green

---

## Phase H — Hints 3-Layer Overhaul (PR2 Agent β)

**Why:** Post-flow onboarding. Per pi.dev mental model + user's "give users prompts for their agent" idea.

**Files:**
- `catalog/agents/<type>/hints.yaml` (new — 6 files, one per agent type)
- `internal/tui/initflow/yield.go` — integrate hint renderer
- `internal/tui/addflow/yield.go` — integrate hint renderer
- `internal/tui/removeflow/yield.go` (new in Phase E) — integrate
- `internal/tui/updateflow/yield.go` (new in Phase F) — integrate
- New: `internal/tui/hints/` package — shared renderer + YAML loader
- `internal/catalog/catalog.go` — load hints into catalog struct (if catalog owns them)

**Schema for `hints.yaml`** (per agent type):
```yaml
---
# Hints shown on yield stages per agent type
init:
  next_cli:
    - "bonsai add <agent>  → install another agent type"
    - "bonsai catalog      → browse available abilities"
  next_workflow:
    - "Edit {{ .DocsPath }}Playbook/Backlog.md — seed initial tasks"
    - "Write your first plan: {{ .DocsPath }}Playbook/Plans/Active/01-<name>.md"
  ai_prompts:
    - label: "Start working"
      body: "Hi, get started — read {{ .DocsPath }}CLAUDE.md and summarize this project."
    - label: "Customize Bonsai"
      body: "Read {{ .DocsPath }}agent/Skills/bonsai-model.md and .bonsai/catalog.json. Based on this project's nature, suggest agents/skills to add or remove. Don't execute yet."
add:
  next_cli:
    - "bonsai list  → verify new agent installed"
  # ... same structure
remove:
  # ...
update:
  # ...
```

**Steps:**

1. **Design `hints.yaml` schema** — single source per agent type, covers all 4 yield stages. Template-rendered at yield time (substitutes `{{ .DocsPath }}`, `{{ .AgentName }}`, etc. via existing template context).

2. **New `internal/tui/hints/` package:**
   ```go
   package hints

   type Block struct {
       NextCLI      []string
       NextWorkflow []string
       AIPrompts    []Prompt
   }
   type Prompt struct { Label, Body string }

   func Load(cat *catalog.Catalog, agentType, command string, ctx TemplateContext) (Block, error)
   func Render(block Block, width int) string  // lipgloss-styled output
   ```

3. **Render shape:** three labeled sub-sections in the yield panel. Prompts displayed as boxed code blocks (user can select-copy). Styled separator between sections.

4. **Integrate in each yield stage:**
   - Replace thin hint lines with `hints.Render(block, width)`
   - Wrap in existing `tui.TitledPanel` for consistency
   - For chromeless yield (e.g. `addflow.renderSuccess`), stack naturally after heroStats.

5. **Catalog loading:** `catalog.LoadCatalog()` grows a `Hints map[string]HintSet` (keyed by agent type). If `hints.yaml` missing for an agent type, fall back to sensible defaults (no crash).

6. **Tests:**
   - `TestHintsLoad_RendersTemplates` — substitute `{{ .DocsPath }}` correctly
   - `TestHintsLoad_MissingYamlDefaults` — no crash, empty block
   - `TestHintsRender_ThreeSectionsPresent` — smoke test output contains "Next" / "Try this" markers

**Verification:**
- [ ] Fresh `bonsai init` yield shows 3-layer hint block
- [ ] `bonsai add` yield shows add-specific hints
- [ ] `NO_COLOR=1 bonsai init | cat` still readable (no styling tax)
- [ ] AI-prompt bodies copy-pastable from terminal

---

## Security

> [!warning]
> Refer to [SecurityStandards.md](../../Standards/SecurityStandards.md) for all security requirements.

**Specific concerns:**

- **Phase A (RefreshPeerAwareness):** template context must not leak agent workspaces from other projects. Since each `ProjectConfig` is project-local, this is automatic — no cross-project risk.
- **Phase C (catalog.json):** `.bonsai/` directory creation — use `os.MkdirAll` with 0o755. No secret data in snapshot (catalog is public).
- **Phase G (`--json`):** output to stdout only. Do not include user-config content (`.bonsai.yaml`) in `--json` output — catalog data only. Prevents accidental exposure of `workspace` paths in CI logs.
- **Phase H (hints.yaml):** templated substitutions. Sanitize `{{ .ProjectName }}` before rendering in shell-safe contexts (though unused in current stages). Hint content is user-displayed, not executed.

Independent security-review agent required for PR1 (Phase A is correctness-critical) and PR2 (Phase F + G touch user-data paths).

---

## Verification (overall)

- [ ] All unit tests green across PR1 + PR2
- [ ] All 6 CI checks green (test, lint, Analyze-Go, govulncheck, CodeQL, GitGuardian)
- [ ] Manual dogfood in fresh `/tmp/v03-smoke` directory:
  - [ ] `bonsai init` → cinematic flow, `.bonsai/catalog.json` present, `station/CLAUDE.md` has Bonsai Reference block, `station/agent/Skills/bonsai-model.md` present
  - [ ] `bonsai add backend` → cinematic flow, tech-lead's `scope-guard-files.sh` lists `backend/` (Phase A verification)
  - [ ] `bonsai add devops` → both tech-lead and backend have devops in their `OtherAgents` blocks
  - [ ] `bonsai list` → cinematic output (Plan 28), shows 3 agents
  - [ ] `bonsai catalog --json | jq .` → valid JSON, categories populated
  - [ ] `bonsai update` → cinematic flow (new)
  - [ ] `bonsai remove devops` → cinematic flow (new), tech-lead and backend get devops removed from their awareness lists
  - [ ] `NO_COLOR=1 bonsai list | cat` → zero ANSI escapes
- [ ] `bonsai-model` skill readable by an agent (spot-check readability)
- [ ] Release pipeline green on tag (Plan 30's GoReleaser + Homebrew continues to work)

---

## Rollout Order

1. Tech-lead writes this plan + `bonsai-model` skill content
2. Tech-lead commits both to `main`
3. Tech-lead dispatches **PR1 agent** (single worktree, isolation=worktree) off committed HEAD
4. PR1 returns → independent code-review + security-review in parallel
5. Fix-agent if minors (single dispatch covering all review findings)
6. PR1 merged (squash)
7. Post-merge: worktree-held-branch cleanup per memory pattern (`git worktree remove -f -f` + `git branch -D` + `git push origin --delete`)
8. Tech-lead dispatches **PR2 Agent α** (remove + catalog --json) and **Agent β** (update + hints) in parallel — separate worktrees, both off merged PR1
9. Each PR returns → independent code-review per PR in parallel
10. Fix-agents per PR if needed
11. Merge PR2a first → PR2b rebases, merge
12. Post-merge cleanup
13. `CHANGELOG.md` entry, version bump refs, tag `v0.3.0`
14. Monitor `release.yml` — watch for Homebrew step (PAT rotation was 2026-04-22; should still be valid)
15. Archive this plan, update Status.md + Roadmap.md

---

## Open Questions

(To be resolved during execution — if any block dispatch, escalate to user.)

1. Does the `CLAUDE.md` template for generated agents live in `catalog/core/` or per-agent `catalog/agents/*/core/`? Confirm exact path at Phase D start.
2. Is `bonsai-model` compatible with `all` agents or only `tech-lead`? Decision: only `tech-lead`, other agents reference via relative path (`{{ .DocsPath }}agent/Skills/bonsai-model.md`). Revisit if dogfood shows awkwardness.
3. Does `update` non-TTY path matter today? It's typed as destructive (modifies workspace) but also as idempotent. Default assumption: non-TTY auto-accepts all defaults (same as existing huh behavior with stdin=closed).
