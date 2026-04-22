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
| `bonsai init` | `cmd/init.go:11` | `initCmd` → `runInit()` in `cmd/init_flow.go:26` — cinematic init flow |
| `bonsai add` | `cmd/add.go:26` | `addCmd` → `runAdd()` in `cmd/add.go:54` — cinematic add flow |
| `bonsai remove` | `cmd/remove.go:32` | `runRemove()` — removes agent or individual items |
| `bonsai list` | `cmd/list.go:19` | `runList()` — table of installed agents + components |
| `bonsai catalog` | `cmd/catalog.go:16` | `runCatalog()` — browse available agents, skills, workflows, etc. |
| `bonsai update` | `cmd/update.go:22` | `runUpdate()` — detect custom files, re-render abilities, refresh CLAUDE.md |
| `bonsai guide` | `cmd/guide.go:34` | `guideCmd` → `runGuide()` at `:42` — render embedded docs in terminal |

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
