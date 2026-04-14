# Bonsai — Code Index

Quick-nav for the developer agent. Jump to what you need.

---

## Entry Point

| What | Where |
|------|-------|
| Embed catalog FS | `main.go:12` — `//go:embed all:catalog` |
| Main | `main.go:15` — `main()` → `cmd.Execute(sub)` |

---

## CLI Commands (`cmd/`)

| Command | File | Entry Function |
|---------|------|----------------|
| `bonsai` (root) | `cmd/root.go:22` | `rootCmd` — shared helpers below |
| `bonsai init` | `cmd/init.go:28` | `runInit()` — creates `.bonsai.yaml`, scaffolding, settings |
| `bonsai add` | `cmd/add.go:57` | `runAdd()` — interactive wizard → agent type, workspace, items → generates files |
| `bonsai remove` | `cmd/remove.go:39` | `runRemove()` — removes agent or individual items |
| `bonsai list` | `cmd/list.go:24` | `runList()` — table of installed agents + components |
| `bonsai catalog` | `cmd/catalog.go:20` | `runCatalog()` — browse available agents, skills, workflows, etc. |

### Shared Helpers (`cmd/root.go`)

| Helper | Line | Purpose |
|--------|------|---------|
| `loadCatalog()` | `:27` | Load embedded catalog or exit |
| `requireConfig()` | `:36` | Load `.bonsai.yaml` or exit |
| `resolveConflicts()` | `:53` | TUI for handling user-modified files (skip / overwrite / backup) |
| `showWriteResults()` | `:103` | Display categorized file trees (created / updated / skipped) |

### Add Helpers (`cmd/add.go`)

| Helper | Line | Purpose |
|--------|------|---------|
| `toItemOptions()` | `:29` | Convert `CatalogItem` slice to TUI picker options |
| `toSensorOptions()` | `:37` | Convert `SensorItem` slice to TUI picker options |
| `toRoutineOptions()` | `:45` | Convert `RoutineItem` slice to TUI picker options |

### Remove Helpers (`cmd/remove.go`)

| Helper | Line | Purpose |
|--------|------|---------|
| `runRemoveItem()` | `:194` | Remove a single skill/workflow/protocol/sensor/routine |
| `agentItemList()` | `:348` | Get an agent's installed items by type |
| `itemIsRequired()` | `:397` | Check if item is required for agent type |
| `itemDisplayName()` | `:423` | Look up display name from catalog |

---

## Catalog (`internal/catalog/catalog.go`)

### Types

| Type | Purpose |
|------|---------|
| `AgentCompat` | Agent compatibility — `"all"` or list of agent type names |
| `CatalogItem` | Skill/workflow/protocol — name, description, agents, content path |
| `SensorItem` | Hook-based sensor — adds event, matcher, script path |
| `RoutineItem` | Periodic routine — adds frequency, content path |
| `ScaffoldingItem` | Project infrastructure — name, description, required, affects, files |
| `AgentDef` | Agent type definition — name, display_name, description, defaults, core files |
| `Catalog` | Root container — holds all loaded items, provides lookup + filtering |

### Key Functions

| Function | Line | Purpose |
|----------|------|---------|
| `DisplayNameFrom()` | `:14` | Convert kebab-case name to title-case display name |
| `New(fsys)` | `:163` | Load full catalog from embedded FS |

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
| `loadItems()` | `:267` | Load skills/workflows/protocols from `meta.yaml` + `.md` |
| `loadSensors()` | `:318` | Load sensors from `meta.yaml` + `.sh.tmpl` |
| `loadRoutines()` | `:369` | Load routines from `meta.yaml` + `.md.tmpl` |
| `loadScaffolding()` | `:420` | Load scaffolding from `manifest.yaml` |
| `loadAgents()` | `:437` | Load agent defs from `agent.yaml` + `core/` |

---

## Config (`internal/config/`)

### `config.go`

| Type / Function | Line | Purpose |
|-----------------|------|---------|
| `ProjectConfig` | — | Root config struct (`.bonsai.yaml`) — project name, docs_path, agents, scaffolding |
| `InstalledAgent` | — | Agent installed in a project — type, workspace, selected items |
| `Save(path)` | `:30` | Write config to YAML |
| `Load(path)` | `:39` | Read config from YAML |

### `lockfile.go`

| Function | Line | Purpose |
|----------|------|---------|
| `NewLockFile()` | `:27` | Create empty lock file |
| `LoadLockFile()` | `:33` | Load `.bonsai-lock.yaml` from project root |
| `Save()` | `:53` | Write lock file to disk |
| `ContentHash()` | `:62` | SHA-256 hash of file content |
| `Track()` | `:68` | Record file path + hash + source |
| `Untrack()` | `:76` | Remove file from tracking |
| `IsModified()` | `:84` | Check if file has changed since generation |

---

## Generator (`internal/generate/generate.go`)

### Core Generation Functions

| Function | Line | Purpose |
|----------|------|---------|
| `Scaffolding()` | `:279` | Generate INDEX.md, Playbook/, Logs/, Reports/ |
| `SettingsJSON()` | `:385` | Generate `.claude/settings.json` with sensor hooks |
| `WorkspaceClaudeMD()` | `:452` | Generate workspace `CLAUDE.md` with nav tables |
| `AgentWorkspace()` | `:737` | Full agent workspace — core templates + items + CLAUDE.md |
| `RoutineDashboard()` | `:621` | Generate `agent/Core/routines.md` dashboard |
| `EnsureRoutineCheckSensor()` | `:583` | Auto-manage routine-check sensor |

### Write System

| Type / Function | Line | Purpose |
|-----------------|------|---------|
| `WriteResult` | — | Tracks all file operations (created, updated, skipped, conflict) |
| `FileResult` | — | Single file operation result |
| `writeFile()` | `:205` | Lock-aware file write (detects conflicts) |
| `writeFileChmod()` | `:237` | Same as writeFile but sets file permissions (for scripts) |
| `ForceConflicts()` | `:181` | Overwrite all conflicted files |

### Helpers

| Function | Line | Purpose |
|----------|------|---------|
| `titleCase()` | `:46` | Custom template func — capitalize each word |
| `renderTemplate()` | `:63` | Render a `.tmpl` file with Go template |
| `renderContent()` | `:250` | Render or copy file content based on `.tmpl` extension |
| `descFor()` | `:79` | Build name→description map for nav tables |
| `parseFrequencyDays()` | `:609` | Parse frequency string (e.g. "5 days") to int |

---

## TUI (`internal/tui/`)

### Styles (`styles.go`)

| Function | Line | Purpose |
|----------|------|---------|
| `Banner()` | `:77` | Bonsai ASCII banner |
| `Success/Error/Warning/Hint/Info()` | `:97–117` | Styled single-line messages |
| `Heading/Section/SectionHeader()` | `:122–132` | Section headers |
| `SuccessPanel/ErrorPanel/WarningPanel/InfoPanel()` | `:141–159` | Boxed panels |
| `TitledPanel()` | `:172` | Generic titled box with custom color |
| `Fields()` | `:220` | Key-value pair display |
| `CardFields()` | `:235` | Card-style key-value (returns string) |
| `ItemTree()` | `:266` | Categorized tree view (for catalog/list output) |
| `FileTree()` | `:311` | File tree view (for write results) |
| `CatalogTable()` | `:369` | Table display for catalog command |

### Prompts (`prompts.go`)

| Function | Line | Purpose |
|----------|------|---------|
| `BonsaiTheme()` | `:12` | Custom Huh form theme |
| `AskText()` | `:62` | Text input prompt |
| `AskSelect()` | `:92` | Single-select prompt |
| `AskMultiSelect()` | `:106` | Multi-select prompt |
| `AskConfirm()` | `:120` | Yes/no confirmation |
| `PickItems()` | `:138` | Multi-select with pre-selected defaults + required items |

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
  → SettingsJSON() → updates .claude/settings.json with sensor hooks
```

---

## File Layout (user's project after setup)

```
project/
├── .bonsai.yaml              ← project config
├── .bonsai-lock.yaml         ← file tracking (hashes + sources)
├── .claude/
│   └── settings.json         ← auto-generated sensor hooks
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
