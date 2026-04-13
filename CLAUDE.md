# Bonsai — Developer Agent

**Codename:** Bonsai
**What:** CLI tool for scaffolding Claude Code agent workspaces — single binary, `go install`
**Stack:** Go 1.24+, Cobra, Huh (forms), LipGloss (styling), BubbleTea (TUI)

> [!warning]
> **FIRST:** Read `agent/Core/identity.md`, then `agent/Core/memory.md`.

---

## Project Structure

```
Bonsai/
├── CLAUDE.md               ← you are here
├── Backlog.md               ← prioritized todo queue (bugs, features, debt, research)
├── main.go                  ← entry point, embeds catalog/ via embed.FS
├── go.mod / go.sum          ← module config
├── Makefile                 ← build, install, clean
├── cmd/                     ← Cobra commands
│   ├── root.go              ← root command, shared helpers
│   ├── init.go              ← bonsai init
│   ├── add.go               ← bonsai add
│   ├── remove.go            ← bonsai remove
│   ├── list.go              ← bonsai list
│   └── catalog.go           ← bonsai catalog
├── internal/
│   ├── catalog/
│   │   └── catalog.go       ← loads YAML metadata from embedded catalog/
│   ├── config/
│   │   └── config.go        ← ProjectConfig, InstalledAgent + YAML I/O
│   ├── generate/
│   │   ├── generate.go      ← renders templates, writes files to target project
│   │   └── generate_test.go ← tests for core file generation
│   └── tui/
│       ├── styles.go         ← LipGloss styles, panels, trees, display helpers
│       └── prompts.go        ← Huh form wrappers (text, select, multi-select, confirm)
├── catalog/                  ← bundled catalog (embedded into binary)
│   ├── core/                 ← shared core files (memory, self-awareness)
│   ├── agents/               ← agent type definitions + identity templates
│   │   ├── tech-lead/
│   │   ├── backend/
│   │   └── frontend/
│   ├── skills/               ← à la carte skills (meta.yaml + content.md)
│   ├── workflows/            ← à la carte workflows
│   ├── protocols/            ← à la carte protocols
│   ├── sensors/              ← auto-enforced hooks (meta.yaml + script.sh.tmpl)
│   ├── routines/             ← periodic maintenance routines (meta.yaml + content.md.tmpl)
│   └── scaffolding/          ← project management infrastructure templates
│       ├── manifest.yaml     ← scaffolding item definitions (name, description, required, affects)
│       ├── INDEX.md.tmpl
│       ├── Playbook/         ← Status, Roadmap, Backlog, Plans, SecurityStandards
│       ├── Logs/             ← FieldNotes, KeyDecisionLog, RoutineLog
│       └── Reports/          ← report-template, Pending/
└── agent/                    ← agent instructions (this agent)
    ├── Core/                 ← identity, memory
    └── Skills/               ← domain skills + references
```

---

## Skills (load when doing specific work)

| Need | Read this |
|------|-----------|
| BubbleTea TUI development | `agent/Skills/bubbletea.md` |

References for each skill live in a subdirectory (e.g. `agent/Skills/bubbletea/`) — load progressively as needed.

---

## Memory

> [!warning]
> **Do NOT use Claude Code's auto-memory system** (`~/.claude/projects/*/memory/`). All persistent memory goes in `agent/Core/memory.md` — version-controlled, auditable, inside the project.

When you would normally write to auto-memory (feedback, references, project context, flags), write to the appropriate section in `agent/Core/memory.md` instead.

When you discover bugs, improvement ideas, tech debt, or feature requests outside your current task scope, add them to `Backlog.md` instead of fixing them inline or noting them only in memory.

---

## Key Concepts

- **Catalog items** (skills, workflows, protocols) each have a `meta.yaml` with `name`, `description`, `agents` (list or `"all"`), optional `required` (same format) and a companion `.md` content file
- **Sensors** are auto-enforced hooks — `meta.yaml` adds `event` (hook event) and optional `matcher` (tool filter), with a companion `.sh.tmpl` script template instead of `.md`
- **Routines** are periodic self-maintenance tasks — `meta.yaml` adds `frequency` (e.g. `"5 days"`), with a companion `.md.tmpl` content template. Installed to `agent/Routines/` with a managed dashboard at `agent/Core/routines.md`
- **`routine-check` sensor** is auto-installed when any routines are present, auto-removed when the last routine is removed — parses the dashboard at session start and flags overdue routines
- **Shared core files** live in `catalog/core/` (memory, self-awareness) — used by all agents. An agent can override any shared file by placing a same-named file in its own `core/` directory. Generator checks agent first, falls back to shared.
- **Agent definitions** have an `agent.yaml` with `name`, `display_name`, `description`, `defaults` and a `core/` directory with agent-specific files (at minimum `identity.md.tmpl`)
- **`.tmpl` extension rule** — files ending in `.tmpl` contain Go template variables (`{{ }}`), are rendered at generation time, and have the `.tmpl` extension stripped from output. Files without `.tmpl` are copied as-is.
- **Templates** use Go `text/template` with `{{ .ProjectName }}`, `{{ .ProjectDescription }}`, `{{ .Routines }}` context vars
- **Scaffolding** is project infrastructure (INDEX, Playbook, Logs, Reports) — defined in `catalog/scaffolding/manifest.yaml` with `name`, `description`, `required`, `affects`, and `files`. Selected during `bonsai init`, some items are required
- **`.bonsai.yaml`** is the project config generated in the user's target project — tracks installed agents, scaffolding selections, and docs_path
- **`.claude/settings.json`** is auto-generated with hook entries for all installed sensors
- **`.bonsai-lock.yaml`** tracks generated files with content hashes — enables conflict detection on re-run
- **Generator** uses lock-aware writes: new files are created, unmodified files are updated silently, user-modified files trigger a conflict prompt (skip / overwrite / backup & overwrite). Scaffolding files are always write-once (skip if exists).
- **Catalog is embedded** via `embed.FS` in `main.go` — ships inside the binary

---

## Development

```bash
make build             # builds ./bonsai binary
./bonsai --help        # verify CLI works
go install .           # install to $GOPATH/bin
```

### Testing changes to catalog items

Edit files in `catalog/`, then rebuild and test in a temp dir:
```bash
make build
mkdir /tmp/test && cd /tmp/test
/path/to/bonsai init
/path/to/bonsai add
/path/to/bonsai list
```

### Adding a new catalog item (skill, workflow, protocol)

1. Create `catalog/{category}/{item-name}/meta.yaml`
2. Create `catalog/{category}/{item-name}/{item-name}.md`
3. Set `agents:` in meta.yaml to control compatibility

### Adding a new sensor

1. Create `catalog/sensors/{name}/meta.yaml` — must include `event` and optionally `matcher`
2. Create `catalog/sensors/{name}/{name}.sh.tmpl` — script template
3. Available events: `SessionStart`, `PreToolUse`, `PostToolUse`, `Stop`, etc.
4. Template context includes: `.ProjectName`, `.AgentName`, `.AgentDisplayName`, `.Workspace`, `.DocsPath`, `.OtherAgents`, `.Protocols`, `.Skills`, `.Workflows`, `.Routines`
5. Custom func: `{{ title .AgentType }}` capitalizes each word

### Adding a new routine

1. Create `catalog/routines/{name}/meta.yaml` — must include `frequency` (e.g. `"5 days"`)
2. Create `catalog/routines/{name}/{name}.md.tmpl` — procedure template (rendered with full TemplateContext)
3. Set `agents:` in meta.yaml to control compatibility
4. Procedure steps should be concrete, idempotent, and reference specific file paths (use template vars for project-specific paths)
5. The `routine-check` sensor is auto-managed — no manual wiring needed

### Adding a new agent type

1. Create `catalog/agents/{name}/agent.yaml`
2. Create `catalog/agents/{name}/core/identity.md.tmpl` (+ memory.md.tmpl, self-awareness.md)
3. Set `defaults:` in agent.yaml to pre-select items

---

## Conventions

- Keep CLI interactive — use Huh forms for all user input
- All catalog items use the same base `meta.yaml` shape: `name`, `description`, `agents`, `required` — sensors add `event` and `matcher`, routines add `frequency`
- **`required`** uses the same format as `agents` (`all` or list of agent types) — required items are auto-installed during `bonsai add` and can't be unchecked
- Generator functions in `internal/generate/`, catalog loading in `internal/catalog/`, commands in `cmd/`
- Go structs for all data shapes (config, catalog models)
- Don't break the existing CLI commands — they're the public API
- TUI styling uses LipGloss — styles defined in `internal/tui/styles.go`

### Naming Standard

**`name`** (machine identifier):
- Characters: `[a-z0-9-]` only — lowercase, digits, hyphens
- Style: kebab-case (`scope-guard-files`, `coding-standards`)
- Unique within its type — two types CAN share a name (e.g. `memory` protocol and `memory` skill)
- No type prefixes (`design-guide`, not `skill-design-guide`)
- No agent prefixes (`design-guide`, not `frontend-design-guide`) — the `agents:` field handles compatibility
- Used in: config files (`.bonsai.yaml`), file paths, template context lists, internal lookups

**`display_name`** (human-readable label):
- Optional in all `meta.yaml` / `agent.yaml` files
- If omitted, auto-derived from `name`: hyphens → spaces, title-cased (`scope-guard-files` → "Scope Guard Files")
- Derivation function: `catalog.DisplayNameFrom()` in `internal/catalog/catalog.go`
- Used in: TUI pickers, catalog tables, list output, generated CLAUDE.md headings
- Never stored in config — purely cosmetic

**Informal word patterns** (not enforced, but encouraged for consistency):
- `*-guard-*` — sensors that block/prevent actions
- `*-check`, `*-hygiene`, `*-accuracy` — routines that audit
- `*-standards`, `*-conventions`, `*-guide` — skills that define rules
- `*-logging`, `*-reporting`, `*-review` — workflows that produce output
