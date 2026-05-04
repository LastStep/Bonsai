# Bonsai — Code Index

Quick-nav for the developer agent. Jump to what you need.

---

## Entry Point

| What | Where |
|------|-------|
| Embed catalog FS | `embed.go:8` — `//go:embed all:catalog` → `CatalogFS` |
| Embed guide cheatsheets | `embed.go:11–21` — `GuideCustomFiles`, `GuideQuickstart`, `GuideConcepts`, `GuideCli` |
| Main | `cmd/bonsai/main.go:15` — `main()` → `cmd.Execute(sub, map[string]string{...})` |

---

## CLI Commands (`cmd/`)

| Command | File | Entry Function |
|---------|------|----------------|
| `bonsai` (root) | `cmd/root.go:28` | `rootCmd` — shared helpers below |
| `bonsai init` | `cmd/init.go:11` | `initCmd` → `runInit()` in `cmd/init_flow.go:27` — cinematic init flow |
| `bonsai add` | `cmd/add.go:28` | `addCmd` → `runAdd()` in `cmd/add.go:56` — cinematic add flow |
| `bonsai remove` | `cmd/remove.go:34` | `runRemove()` — removes agent or individual items |
| `bonsai list` | `cmd/list.go:18` | `runList()` — table of installed agents + components |
| `bonsai catalog` | `cmd/catalog.go:23` | `runCatalog()` — browse available agents, skills, workflows, etc. |
| `bonsai update` | `cmd/update.go:19` | `runUpdate()` — detect custom files, re-render abilities, refresh CLAUDE.md |
| `bonsai guide` | `cmd/guide.go:27` | `guideCmd` → `runGuide()` at `:44` — render embedded docs in terminal |
| `bonsai validate` | `cmd/validate.go:23` | `validateCmd` → `runValidate()` at `:43` — read-only audit (orphans, stale lock, untracked customs, frontmatter) |

### Shared Helpers (`cmd/root.go`)

| Helper | Line | Purpose |
|--------|------|---------|
| `loadCatalog()` | `:42` | Load embedded catalog or exit |
| `requireConfig()` | `:50` | Load `.bonsai.yaml` or exit |
| `mustCwd()` | `:61` | Resolve current working dir or exit |
| `Execute()` | `:80` | Wire catalog + guides into root command and run |
| `buildConflictSteps()` | `:101` | Harness steps for legacy conflict picker |
| `applyConflictPicks()` | `:147` | Apply per-file conflict choices to lock + disk |
| `showWriteResults()` | `:198` | Display categorized file trees (created / updated / skipped) |

### Init Helpers (`cmd/init_flow.go`)

| Helper | Line | Purpose |
|--------|------|---------|
| `runInit()` | `:26` | Cinematic init: Vessel → Soil → Branches → Observe → Generate → Planted |
| `buildGenerateAction()` | `:206` | Closure invoked by GenerateStage to write scaffolding + agent |
| `plantedSummary()` | `:277` | Ability counts rendered in the Planted stage summary |
| `scaffoldingToSoilOptions()` | `:292` | Map catalog scaffolding entries into Soil picker options |

### Add Helpers (`cmd/add.go`)

| Helper | Line | Purpose |
|--------|------|---------|
| `runAdd()` | `:54` | Cinematic add: Select → [Ground] → Graft → Observe → Grow → [Conflicts] → Yield |
| `applyCinematicConflictPicks()` | `:293` | Apply per-file conflict choices from ConflictsStage |
| `installedSet()` | `:349` | Build "already installed" lookup for agent picker |
| `buildAddGrowAction()` | `:371` | Closure invoked by GrowStage to generate selected abilities |
| `distributeAddItemPicks()` | `:531` | Split per-category picks into skills/workflows/protocols/sensors/routines |
| `availableAddItems()` | `:616` | Uninstalled-per-category catalog filter (used by Select + Grow) |

### Remove Helpers (`cmd/remove.go`)

| Helper | Line | Purpose |
|--------|------|---------|
| `runRemoveItem()` | `:248` | Remove a single skill/workflow/protocol/sensor/routine |
| `runRemoveItemAction()` | `:466` | Execute removal + regenerate affected agents |
| `agentItemList()` | `:519` | Get an agent's installed items by type |
| `itemIsRequired()` | `:568` | Check if item is required for agent type |
| `itemDisplayName()` | `:594` | Look up display name from catalog |

---

## Catalog (`internal/catalog/catalog.go`)

### Types

| Type | Purpose |
|------|---------|
| `AgentCompat` | Agent compatibility — `"all"` or list of agent type names |
| `CatalogItem` | Skill/workflow/protocol — name, description, agents, triggers, content path |
| `SensorItem` | Hook-based sensor — adds event, matcher, script path |
| `TriggerExample` | Prompt-action pair showing how an ability activates |
| `Triggers` | Activation metadata for skills and workflows — scenarios, examples, paths |
| `RoutineItem` | Periodic routine — adds frequency, content path |
| `ScaffoldingItem` | Project infrastructure — name, description, required, affects, files |
| `AgentDef` | Agent type definition — name, display_name, description, defaults, core files |
| `Catalog` | Root container — holds all loaded items, provides lookup + filtering |

### Key Functions

| Function | Line | Purpose |
|----------|------|---------|
| `DisplayNameFrom()` | `:49` | Convert kebab-case name to title-case display name |
| `New(fsys)` | `:220` | Load full catalog from embedded FS |

### Lookup Methods on `Catalog`

| Method | Returns |
|--------|---------|
| `GetAgent(name)` | `*AgentDef` or nil |
| `GetSkill(name)` | `*CatalogItem` or nil |
| `GetWorkflow(name)` | `*CatalogItem` or nil |
| `GetProtocol(name)` | `*CatalogItem` or nil |
| `GetSensor(name)` | `*SensorItem` or nil |
| `GetRoutine(name)` | `*RoutineItem` or nil |
| `GetScaffolding(name)` | `*ScaffoldingItem` or nil |
| `GetItem(name)` | `*CatalogItem` — searches skills, workflows, protocols |
| `SkillsFor(agentType)` | Compatible skills |
| `WorkflowsFor(agentType)` | Compatible workflows |
| `ProtocolsFor(agentType)` | Compatible protocols |
| `SensorsFor(agentType)` | Compatible sensors |
| `RoutinesFor(agentType)` | Compatible routines |

### Internal Loaders

| Function | Line | Purpose |
|----------|------|---------|
| `loadItems()` | `:324` | Load skills/workflows/protocols from `meta.yaml` + `.md` |
| `loadSensors()` | `:375` | Load sensors from `meta.yaml` + `.sh.tmpl` |
| `loadRoutines()` | `:426` | Load routines from `meta.yaml` + `.md.tmpl` |
| `loadScaffolding()` | `:477` | Load scaffolding from `manifest.yaml` |
| `loadAgents()` | `:494` | Load agent defs from `agent.yaml` + `core/` |

---

## Config (`internal/config/`)

### `config.go`

| Type / Function | Line | Purpose |
|-----------------|------|---------|
| `CustomItemMeta` | `:10` | Metadata for user-created custom items (parsed from frontmatter) |
| `ProjectConfig` | — | Root config struct (`.bonsai.yaml`) — project name, docs_path, agents, scaffolding |
| `InstalledAgent` | — | Agent installed in a project — type, workspace, selected items |
| `Save(path)` | `:40` | Write config to YAML |
| `Load(path)` | `:49` | Read config from YAML |

### `lockfile.go`

| Function | Line | Purpose |
|----------|------|---------|
| `NewLockFile()` | `:27` | Create empty lock file |
| `LoadLockFile()` | `:33` | Load `.bonsai-lock.yaml` from project root |
| `Save()` | `:53` | Write lock file to disk |
| `ContentHash()` | `:62` | SHA-256 hash of file content |
| `Track()` | `:68` | Record file path + hash + source |
| `Untrack()` | `:76` | Remove file from tracking |
| `ConflictAction` | `:86` | Enum — Skip / Overwrite / Backup choices for conflicted files |
| `IsModified()` | `:103` | Check if file has changed since generation |

---

## Generator (`internal/generate/`)

### `generate.go` — Core Generation Functions

| Function | Line | Purpose |
|----------|------|---------|
| `Scaffolding()` | `:359` | Generate INDEX.md, Playbook/, Logs/, Reports/ |
| `SettingsJSON()` | `:466` | Generate `.claude/settings.json` with sensor hooks |
| `WorkspaceClaudeMD()` | `:642` | Generate workspace `CLAUDE.md` with nav tables |
| `AgentWorkspace()` | `:1216` | Full agent workspace — core templates + items + CLAUDE.md |
| `RoutineDashboard()` | `:917` | Generate `agent/Core/routines.md` dashboard |
| `EnsureRoutineCheckSensor()` | `:879` | Auto-manage routine-check sensor |
| `PathScopedRules()` | `:1065` | Generate `.claude/rules/skill-{name}.md` for path-scoped auto-loading |
| `WorkflowSkills()` | `:1103` | Generate `.claude/skills/{name}/SKILL.md` for curated workflows |

### Write System

| Type / Function | Line | Purpose |
|-----------------|------|---------|
| `FileAction` | `:140` | Enum — Create / Update / Unchanged / Skipped / Conflict |
| `FileResult` | `:152` | Single file operation result |
| `WriteResult` | `:161` | Tracks all file operations (created, updated, skipped, conflict) |
| `writeFile()` | `:276` | Lock-aware file write (detects conflicts) |
| `writeFileChmod()` | `:316` | Same as writeFile but sets file permissions (for scripts) |
| `ForceConflicts()` | `:212` | Overwrite all conflicted files |
| `ForceSelected()` | `:234` | Overwrite only user-selected conflict files |

### Helpers

| Function | Line | Purpose |
|----------|------|---------|
| `titleCase()` | `:47` | Custom template func — capitalize each word |
| `renderTemplate()` | `:64` | Render a `.tmpl` file with Go template |
| `renderContent()` | `:330` | Render or copy file content based on `.tmpl` extension |
| `descFor()` | `:80` | Build name→description map for nav tables (supports custom items) |
| `scenariosDesc()` | `:118` | Trigger-aware description for CLAUDE.md tables |
| `CuratedSlashWorkflows` | `:131` | Package-level set of workflows that get slash-command files |
| `hasScaffolding()` | `:345` | Check if scaffolding item is selected |
| `howToWorkLines()` | `:545` | Generate "How to Work" heuristics section |
| `quickTriggersLines()` | `:600` | Generate Quick Triggers reference table |
| `triggerSection()` | `:1155` | Generate trigger header for ability files |
| `parseFrequencyDays()` | `:905` | Parse frequency string (e.g. "5 days") to int |

### `frontmatter.go` — Custom File Parsing

| Function | Line | Purpose |
|----------|------|---------|
| `ParseFrontmatter()` | `:13` | Extract YAML frontmatter from custom file content |

### `scan.go` — Custom File Discovery

| Type / Function | Line | Purpose |
|-----------------|------|---------|
| `DiscoveredFile` | `:12` | Represents a custom file found in a workspace |
| `ScanCustomFiles()` | `:22` | Find untracked custom files in an agent's workspace directories |

### `catalog_snapshot.go` — Agent-consumable Catalog Listing (Plan 31)

| Type / Function | Line | Purpose |
|-----------------|------|---------|
| `CatalogSnapshot` | `:21` | Stable JSON shape written to `.bonsai/catalog.json` (decoupled from internal catalog types) |
| `AgentEntry` / `AbilityEntry` / `SensorEntry` / `RoutineEntry` | — | Per-section row shapes for the snapshot |
| `WriteCatalogSnapshot()` | — | Render full catalog into `.bonsai/catalog.json` for downstream agent readers |

---

## Validate (`internal/validate/`) — Plan 35

Read-only ability-state audit — detects orphaned registrations, stale lock entries, untracked custom files, and frontmatter problems. Strictly read-only; fixes happen via `bonsai update`.

| Type / Function | File | Purpose |
|-----------------|------|---------|
| `Run()` | `validate.go` | Entry point — runs all detection categories per installed agent, returns a `Report` |
| `Report` / `Issue` / `Severity` | `validate.go` | Result shape — issues classified by severity (error / warning) |
| `Category` constants | `validate.go` | Six detection categories (orphans, stale lock entries, untracked customs, frontmatter, etc.) |

Dependencies kept to `internal/config`, `internal/catalog`, and `internal/generate.ParseFrontmatter` — no TUI import so `bonsai validate --json` works in headless CI.

---

## Workspace-path Validation (`internal/wsvalidate/`) — Plan 32

Single source of truth for workspace-path rules used by addflow + initflow + cmd. Stdlib-only (`path/filepath` + `strings`).

| Function | File | Purpose |
|----------|------|---------|
| `Normalise()` | `wsvalidate.go` | Trim + `filepath.Clean` + trailing-slash canonicalisation |
| `InvalidReason()` | `wsvalidate.go` | User-facing error string when path escapes root, is absolute, contains backslash, or reduces to root |

---

## TUI (`internal/tui/`)

### Styles (`styles.go`)

| Function | Line | Purpose |
|----------|------|---------|
| `Banner()` | `:152` | Bonsai ASCII banner |
| `Success/Error/Warning/Hint/Info()` | `:187–221` | Styled single-line messages |
| `Heading/Section/SectionHeader()` | `:226–236` | Section headers |
| `SuccessPanel/ErrorPanel/WarningPanel/InfoPanel()` | `:245–289` | Boxed panels |
| `EmptyPanel()` | `:297` | Dim panel for empty states |
| `TitledPanel()` | `:376` | Generic titled box with custom color |
| `Fields()` | `:383` | Key-value pair display |
| `CardFields()` | `:398` | Card-style key-value (returns string) |
| `ItemTree()` | `:429` | Categorized tree view (for catalog/list output) |
| `FileTree()` | `:475` | File tree view (for write results) |
| `CatalogTable()` | `:533` | Table display for catalog command |

### Prompts (`prompts.go`)

| Function | Line | Purpose |
|----------|------|---------|
| `BonsaiTheme()` | `:12` | Custom Huh form theme |
| `AskText()` | `:66` | Text input prompt |
| `AskSelect()` | `:96` | Single-select prompt |
| `AskMultiSelect()` | `:110` | Multi-select prompt |
| `AskConfirm()` | `:124` | Yes/no confirmation |
| `PickItems()` | `:142` | Multi-select with pre-selected defaults + required items |

### Harness (`harness/`) — step/reducer runtime for cinematic flows

| Function / Type | Line | Purpose |
|-----------------|------|---------|
| `Step` interface | `harness.go:49` | Contract every flow stage implements (Title/Done/Result/Update/View) |
| `Harness` | `harness.go:116` | BubbleTea reducer that drives a linear list of Steps |
| `Run()` | `harness.go:523` | Build + run a Harness, return per-step results or `ErrAborted` |
| `TextStep` | `steps.go:21` | Huh-backed text-input step |
| `SelectStep` | `steps.go:135` | Huh-backed single-select step |
| `MultiSelectStep` | `steps.go:193` | Huh-backed multi-select step (supports required + defaults) |
| `ConfirmStep` / `ReviewStep` / `NoteStep` | `steps.go` | Confirmation, review-table, and static-note steps |
| `SpinnerStep` | `steps.go` | Long-running action with spinner; captures action error |
| `Conditional` / `Lazy` / `LazyGroup` | `steps.go` / `harness.go` | Predicate-gated, deferred, and splicing step wrappers |

### InitFlow (`initflow/`) — `bonsai init` cinematic stages

| Stage / File | Entry | Purpose |
|--------------|-------|---------|
| `VesselStage` | `vessel.go:48` | Project name + docs_path text inputs |
| `SoilStage` | `soil.go:40` | Scaffolding multi-select (INDEX, Playbook, Logs, Reports) |
| `BranchesStage` | `branches.go:88` | Tabbed ability picker (skills/workflows/protocols/sensors/routines) |
| `ObserveStage` | `observe.go:56` | Pre-generation confirm — project + planting summary |
| `GenerateStage` | `generate.go` | Spinner stage running `GenerateAction` closure |
| `PlantedStage` | `planted.go:54` | Post-generation summary + written file tree |
| Chrome | `chrome.go:41` / `:129` | `RenderHeader()` / `RenderFooter()` — kanji rail + key hints |
| Layout | `layout.go:21` / `:48` / `:106` | `TerminalTooSmall`, `ClampColumns`, `Viewport` scroll helper |
| Design tokens | `design.go:29` | `PanelWidth`, focused/unfocused styles, conflict tone palette |
| Fallback | `fallback.go:33` / `:87` | `WideCharSafe()` detection + `StageLabels` (kanji + ASCII pair) |
| Enso glyph | `enso.go` | Circle SVG-ish glyph rendered in chrome |
| `StageContext` | `stage.go:258` | Shared context threaded into every stage (version, dirs, timings) |
| `Stage` base | `stage.go:22` | Embedded helper for rail, size, title, done state |

### AddFlow (`addflow/`) — `bonsai add` cinematic stages

| Stage / File | Entry | Purpose |
|--------------|-------|---------|
| `SelectStage` | `select.go:34` | Agent type picker (marks already-installed) |
| `GroundStage` | `ground.go:48` | Workspace path input (auto-complete for tech-lead) |
| `GraftStage` | `graft.go:76` / `:84` | Tabbed ability picker — `NewAgent` or `AddItems` variant |
| `ObserveStage` | `observe.go:43` | Pre-generation confirm — agent + abilities summary |
| `GenerateStage` | `grow.go:16` | Wraps `initflow.GenerateStage` with add-specific copy |
| `ConflictsStage` | `conflicts.go:53` | Per-file conflict picker (Skip / Overwrite / Backup) |
| `YieldStage` | `yield.go:57–85` | Terminal outcomes: Success, AllInstalled, TechLeadRequired, UnknownAgent |
| `StageLabels` | `addflow.go:49` | Per-stage kanji rail labels (SELECT / GROUND / GRAFT / …) |
| `BuildAgentOptions()` | `select.go:57` | Catalog → `AgentOption` list with installed markers |
| `NormaliseWorkspace()` | `ground.go:226` | Trim + slash-normalise user-entered workspace path |

### RemoveFlow (`removeflow/`) — `bonsai remove` cinematic stages (Plan 31)

Mirrors addflow's chromeless-stage shape with a 4-segment rail (択 SELECT · 観 OBSERVE · 確 CONFIRM · 結 YIELD). Conflicts stage spliced off-rail when the post-generate write produces user-modified files. Two entry shapes: agent-removal (Select skipped) and item-removal (Select fires when multiple agents share an item).

| Stage / File | Purpose |
|--------------|---------|
| `removeflow.go` | Public entry + stage labels + rail wiring |
| `select.go` | Item-disambiguation picker (multiple agents own the item) |
| `observe.go` | Pre-removal preview of installed ability tree |
| `confirm.go` | Gate before write; explicit confirmation |
| `conflicts.go` | Off-rail per-file conflict picker |
| `yield.go` | Terminal summary card |

### UpdateFlow (`updateflow/`) — `bonsai update` cinematic stages (Plan 31)

5-stage rail (探 DISCOVER · 択 SELECT · 同 SYNC · 衝 CONFLICT · 結 YIELD). All chrome primitives imported from initflow — never reimplemented. Conflict stage off-rail; spliced lazily when Sync surfaces a conflict list.

| Stage / File | Purpose |
|--------------|---------|
| `updateflow.go` | Public entry, stage indices, rail wiring |
| `discover.go` | Scan for untracked customs + parse frontmatter |
| `select.go` | Pick which custom files to absorb into the lockfile |
| `run.go` | Drive sync stage (re-render abilities) |
| `sync.go` | Apply chosen rendering pipeline; surface conflicts |
| `conflicts.go` | Off-rail per-file conflict picker |
| `yield.go` | Terminal summary card |

### ListFlow (`listflow/`) — `bonsai list` cinematic render (Plan 31)

Static, non-interactive output (no BubbleTea model). `RenderAll` is a pure function returning the full rendered output ready for one `fmt.Print`. Composes initflow chrome (header + min-size floor) + per-agent panels + muted counts footer (agents · skills · workflows · protocols · sensors · routines).

| File | Purpose |
|------|---------|
| `listflow.go` | `RenderAll()` entry; chrome composition |
| `agent_panel.go` | Per-agent panel + workspace tree/hint stack |
| `fs_helpers.go` | Workspace tree gathering helpers |

### CatalogFlow (`catalogflow/`) — `bonsai catalog` cinematic browser (Plan 28)

Tabbed BubbleTea browser — single stage cycling 7 tabs over the 7 catalog sections (Agents · Skills · Workflows · Protocols · Sensors · Routines · Scaffolding). All chrome primitives imported from initflow. Public entry: `NewBrowser`. TTY-only; non-TTY falls back to the static-render path in `cmd/catalog.go`.

| File | Purpose |
|------|---------|
| `catalogflow.go` | Public `NewBrowser` entry + tab orchestration |
| `browser.go` | BubbleTea model + tab cycling logic |
| `entry.go` | Per-row `Entry` shape used across all tabs |

### GuideFlow (`guideflow/`) — `bonsai guide` cinematic viewer (Plan 28/30)

Tabbed BubbleTea scroll viewport rendering bundled markdown via glamour inside shared initflow chrome. English-only labels (per Plan 28 Session 2026-04-23 decision D1). Topics supplied as a pre-ordered slice; viewer preserves order. Frontmatter stripped at render time so cache key stays pure.

| File | Purpose |
|------|---------|
| `guideflow.go` | Public entry + `Topic` shape |
| `viewer.go` | BubbleTea model — tab strip + scroll viewport |
| `render.go` | Glamour rendering with cache + frontmatter stripping |

### Hints (`hints/`) — yield-stage 3-layer renderer (Plan 31)

Catalog-driven hints block consumed by every cinematic yield stage (init's Planted, add/remove/update Yield). Three sections per command: NEXT STEPS (CLI commands), TRY THIS (in-workspace workflows), ASK YOUR AGENT (copy-paste AI prompts). Sourced from `catalog/agents/<type>/hints.yaml`, template-rendered against TemplateContext. Missing entries fall back to a zero `Block` silently.

| File | Purpose |
|------|---------|
| `hints.go` | `Block` shape + zero-value safe rendering |
| `load.go` | Loader for `catalog/agents/<type>/hints.yaml` |
| `render.go` | Three-section formatter with template execution |

---

## Generation Flow

```
bonsai init
  → ProjectConfig created → .bonsai.yaml
  → Scaffolding() → INDEX.md, Playbook/*, Logs/*, Reports/*
  → SettingsJSON() → .claude/settings.json (empty initially)

bonsai add
  → pick agent type → pick workspace → pick items (skills, workflows, protocols, sensors, routines)
  → InstalledAgent saved to .bonsai.yaml
  → AgentWorkspace()
      → core/ templates rendered (.tmpl → .md) into {workspace}/agent/Core/
      → skills/workflows/protocols .md copied into {workspace}/agent/Skills|Workflows|Protocols/
      → sensors rendered (.sh.tmpl) into {workspace}/agent/Sensors/
      → routines rendered (.md.tmpl) into {workspace}/agent/Routines/
      → WorkspaceClaudeMD() → {workspace}/CLAUDE.md
      → RoutineDashboard() → {workspace}/agent/Core/routines.md (if routines present)
      → EnsureRoutineCheckSensor() → auto-add/remove routine-check sensor
  → PathScopedRules() → .claude/rules/skill-{name}.md (if triggers.paths defined)
  → WorkflowSkills() → .claude/skills/{name}/SKILL.md (for curated workflows)
  → SettingsJSON() → updates .claude/settings.json with sensor hooks

bonsai update
  → ScanCustomFiles() → find untracked user-created files
  → ParseFrontmatter() → extract metadata from custom files
  → re-render all agents (AgentWorkspace, PathScopedRules, WorkflowSkills, SettingsJSON)
```

---

## File Layout (user's project after setup)

```
project/
├── .bonsai.yaml              ← project config
├── .bonsai-lock.yaml         ← file tracking (hashes + sources)
├── .claude/
│   ├── settings.json         ← auto-generated sensor hooks
│   ├── rules/                ← path-scoped skill auto-load rules
│   │   └── skill-{name}.md
│   └── skills/               ← curated workflow slash-command files
│       └── {name}/
│           └── SKILL.md
├── INDEX.md                  ← project snapshot (scaffolding)
├── Playbook/
│   ├── Status.md
│   ├── Roadmap.md
│   ├── Backlog.md
│   ├── Standards/SecurityStandards.md
│   └── Plans/Active/
├── Logs/
│   ├── FieldNotes.md
│   ├── KeyDecisionLog.md
│   └── RoutineLog.md
├── Reports/
│   ├── report-template.md
│   └── Pending/
└── station/                  ← workspace (example)
    ├── CLAUDE.md             ← workspace nav (generated)
    └── agent/
        ├── Core/
        │   ├── identity.md
        │   ├── memory.md
        │   ├── self-awareness.md
        │   └── routines.md   ← (if routines installed)
        ├── Skills/
        ├── Workflows/
        ├── Protocols/
        ├── Sensors/
        └── Routines/
```
