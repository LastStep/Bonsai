# Bonsai — Code Index

Quick-nav for the developer agent. Jump to what you need.

---

## Entry Point

| What | Where |
|------|-------|
| Embed catalog FS | `main.go:15` — `//go:embed all:catalog` |
| Embed guide content | `main.go:18` — `//go:embed docs/custom-files.md` |
| Main | `main.go:21` — `main()` → `cmd.Execute(sub, guideContent)` |

---

## CLI Commands (`cmd/`)

| Command | File | Entry Function |
|---------|------|----------------|
| `bonsai` (root) | `cmd/root.go:27` | `rootCmd` — shared helpers below |
| `bonsai init` | `cmd/init.go:28` | `runInit()` — creates `.bonsai.yaml`, scaffolding, settings |
| `bonsai add` | `cmd/add.go:57` | `runAdd()` — interactive wizard → agent type, workspace, items → generates files |
| `bonsai remove` | `cmd/remove.go:39` | `runRemove()` — removes agent or individual items |
| `bonsai list` | `cmd/list.go:24` | `runList()` — table of installed agents + components |
| `bonsai catalog` | `cmd/catalog.go:20` | `runCatalog()` — browse available agents, skills, workflows, etc. |
| `bonsai update` | `cmd/update.go:28` | `runUpdate()` — detect custom files, re-render abilities, refresh CLAUDE.md |
| `bonsai guide` | `cmd/guide.go:22` | `runGuide()` — render custom files guide in terminal |

### Shared Helpers (`cmd/root.go`)

| Helper | Line | Purpose |
|--------|------|---------|
| `loadCatalog()` | `:32` | Load embedded catalog or exit |
| `requireConfig()` | `:41` | Load `.bonsai.yaml` or exit |
| `resolveConflicts()` | `:65` | TUI for handling user-modified files (skip / overwrite / backup) |
| `showWriteResults()` | `:107` | Display categorized file trees (created / updated / skipped) |

### Add Helpers (`cmd/add.go`)

| Helper | Line | Purpose |
|--------|------|---------|
| `toItemOptions()` | `:29` | Convert `CatalogItem` slice to TUI picker options |
| `toSensorOptions()` | `:37` | Convert `SensorItem` slice to TUI picker options |
| `toRoutineOptions()` | `:45` | Convert `RoutineItem` slice to TUI picker options |

### Remove Helpers (`cmd/remove.go`)

| Helper | Line | Purpose |
|--------|------|---------|
| `runRemoveItem()` | `:198` | Remove a single skill/workflow/protocol/sensor/routine |
| `agentItemList()` | `:366` | Get an agent's installed items by type |
| `itemIsRequired()` | `:415` | Check if item is required for agent type |
| `itemDisplayName()` | `:441` | Look up display name from catalog |

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
| `DisplayNameFrom()` | `:14` | Convert kebab-case name to title-case display name |
| `New(fsys)` | `:177` | Load full catalog from embedded FS |

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
| `loadItems()` | `:281` | Load skills/workflows/protocols from `meta.yaml` + `.md` |
| `loadSensors()` | `:332` | Load sensors from `meta.yaml` + `.sh.tmpl` |
| `loadRoutines()` | `:383` | Load routines from `meta.yaml` + `.md.tmpl` |
| `loadScaffolding()` | `:434` | Load scaffolding from `manifest.yaml` |
| `loadAgents()` | `:451` | Load agent defs from `agent.yaml` + `core/` |

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
| `IsModified()` | `:84` | Check if file has changed since generation |

---

## Generator (`internal/generate/`)

### `generate.go` — Core Generation Functions

| Function | Line | Purpose |
|----------|------|---------|
| `Scaffolding()` | `:332` | Generate INDEX.md, Playbook/, Logs/, Reports/ |
| `SettingsJSON()` | `:439` | Generate `.claude/settings.json` with sensor hooks |
| `WorkspaceClaudeMD()` | `:615` | Generate workspace `CLAUDE.md` with nav tables |
| `AgentWorkspace()` | `:1159` | Full agent workspace — core templates + items + CLAUDE.md |
| `RoutineDashboard()` | `:884` | Generate `agent/Core/routines.md` dashboard |
| `EnsureRoutineCheckSensor()` | `:846` | Auto-manage routine-check sensor |
| `PathScopedRules()` | `:1032` | Generate `.claude/rules/skill-{name}.md` for path-scoped auto-loading |
| `WorkflowSkills()` | `:1070` | Generate `.claude/skills/{name}/SKILL.md` for curated workflows |

### Write System

| Type / Function | Line | Purpose |
|-----------------|------|---------|
| `WriteResult` | — | Tracks all file operations (created, updated, skipped, conflict) |
| `FileResult` | — | Single file operation result |
| `writeFile()` | `:258` | Lock-aware file write (detects conflicts) |
| `writeFileChmod()` | `:290` | Same as writeFile but sets file permissions (for scripts) |
| `ForceConflicts()` | `:208` | Overwrite all conflicted files |
| `ForceSelected()` | `:230` | Overwrite only user-selected conflict files |

### Helpers

| Function | Line | Purpose |
|----------|------|---------|
| `titleCase()` | `:46` | Custom template func — capitalize each word |
| `renderTemplate()` | `:63` | Render a `.tmpl` file with Go template |
| `renderContent()` | `:303` | Render or copy file content based on `.tmpl` extension |
| `descFor()` | `:79` | Build name→description map for nav tables (supports custom items) |
| `scenariosDesc()` | `:117` | Trigger-aware description for CLAUDE.md tables |
| `CuratedSlashWorkflows` | `:130` | Package-level set of workflows that get slash-command files |
| `hasScaffolding()` | `:318` | Check if scaffolding item is selected |
| `howToWorkLines()` | `:518` | Generate "How to Work" heuristics section |
| `quickTriggersLines()` | `:573` | Generate Quick Triggers reference table |
| `triggerSection()` | `:1122` | Generate trigger header for ability files |
| `parseFrequencyDays()` | `:872` | Parse frequency string (e.g. "5 days") to int |

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
| `Banner()` | `:77` | Bonsai ASCII banner |
| `Success/Error/Warning/Hint/Info()` | `:97–117` | Styled single-line messages |
| `Heading/Section/SectionHeader()` | `:122–132` | Section headers |
| `SuccessPanel/ErrorPanel/WarningPanel/InfoPanel()` | `:141–164` | Boxed panels |
| `EmptyPanel()` | `:167` | Dim panel for empty states |
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
