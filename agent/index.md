# Bonsai — Code Index

Quick-nav for the developer agent. Jump to what you need.

---

## Entry Point

| What | Where |
|------|-------|
| Package config | `pyproject.toml` — entry point: `bonsai = "bonsai.cli:main"` |
| CLI app | `src/bonsai/cli.py:340` — `main()` → Typer `app` |

---

## CLI Commands

| Command | Function | What it does |
|---------|----------|--------------|
| `bonsai init` | `cli.py:76` `init()` | Creates `.bonsai.yaml`, root `CLAUDE.md`, scaffolding |
| `bonsai add` | `cli.py:128` `add()` | Interactive wizard → picks agent type, workspace, items → generates files |
| `bonsai remove` | `cli.py:219` `remove()` | Removes agent from config, optionally deletes files |
| `bonsai list` | `cli.py:264` `list_agents()` | Table of installed agents + their components |
| `bonsai catalog` | `cli.py:292` `show_catalog()` | Browse all available agents, skills, workflows, protocols |

### CLI Helpers

| Helper | Location | Purpose |
|--------|----------|---------|
| `_require_config()` | `cli.py:35` | Load `.bonsai.yaml` or exit |
| `_pick_items()` | `cli.py:46` | Interactive checkbox for catalog items |

---

## Models (`src/bonsai/models.py`)

| Model | Line | Purpose |
|-------|------|---------|
| `CatalogItem` | `:11` | Skill/workflow/protocol — name, description, agents, content_path |
| `AgentDef` | `:25` | Agent type definition — name, display_name, defaults, core_dir |
| `InstalledAgent` | `:39` | Agent installed in a project — type, workspace, selected items |
| `ProjectConfig` | `:49` | Root config (`.bonsai.yaml`) — project_name, docs_path, agents dict |

### Key methods

- `CatalogItem.compatible_with(agent_type)` — checks if item works with an agent
- `ProjectConfig.save(path)` / `ProjectConfig.load(path)` — YAML serialization

---

## Catalog (`src/bonsai/catalog.py`)

| What | Location |
|------|----------|
| `Catalog` class | `catalog.py:100` — loads everything, provides lookup + filtering |
| `_load_items(category)` | `catalog.py:14` — loads skills/workflows/protocols from `meta.yaml` + `.md` |
| `_load_agents()` | `catalog.py:61` — loads agent defs from `agent.yaml` + `core/` dir |

### Catalog lookup methods

| Method | Returns |
|--------|---------|
| `get_agent(name)` | `AgentDef` or None |
| `get_skill(name)` | `CatalogItem` or None |
| `get_workflow(name)` | `CatalogItem` or None |
| `get_protocol(name)` | `CatalogItem` or None |
| `get_item(name)` | Any `CatalogItem` (searches all categories) |
| `skills_for(agent_type)` | Compatible skills |
| `workflows_for(agent_type)` | Compatible workflows |
| `protocols_for(agent_type)` | Compatible protocols |

---

## Generator (`src/bonsai/generator.py`)

| Function | Line | What it generates |
|----------|------|-------------------|
| `generate_scaffolding()` | `:48` | INDEX.md, Playbook/, Logs/, Reports/ in docs_path |
| `generate_root_claude_md()` | `:81` | Root `CLAUDE.md` — routing table + universal rules + triggers |
| `generate_workspace_claude_md()` | `:144` | Workspace `CLAUDE.md` — nav tables for core/skills/workflows/protocols |
| `generate_agent_workspace()` | `:233` | Full `agent/` dir — core templates + selected items + workspace CLAUDE.md |

### Generator helpers

| Helper | Line | Purpose |
|--------|------|---------|
| `_render_template()` | `:17` | Render a `.j2` file with Jinja2 |
| `_copy_or_render()` | `:28` | Copy file, rendering as Jinja2 if `.j2` extension |
| `_desc_for()` | `:38` | Build name→description map for nav tables |

---

## Catalog Contents

### Agent Types (`src/bonsai/catalog/agents/`)

| Agent | Role | Default Skills | Default Workflows | Default Protocols |
|-------|------|---------------|-------------------|-------------------|
| `tech-lead` | Architects, plans, reviews — no app code | planning-template | planning, code-review, session-logging | session-start, security, scope-boundaries |
| `backend` | API, database, server-side | coding-standards, testing, database-conventions | plan-execution, reporting, session-logging | session-start, security, scope-boundaries |
| `frontend` | UI, state, styling | coding-standards, testing, design-guide | plan-execution, reporting, session-logging | session-start, security, scope-boundaries |

### Skills (`src/bonsai/catalog/skills/`)

coding-standards, database-conventions, design-guide, planning-template, testing

### Workflows (`src/bonsai/catalog/workflows/`)

code-review, plan-execution, planning, reporting, session-logging

### Protocols (`src/bonsai/catalog/protocols/`)

memory, scope-boundaries, security, session-start

---

## Generation Flow

```
bonsai init
  → ProjectConfig created → .bonsai.yaml
  → generate_root_claude_md() → CLAUDE.md (routing table)
  → generate_scaffolding() → INDEX.md, Playbook/*, Logs/*, Reports/*

bonsai add
  → pick agent type → pick workspace → pick skills/workflows/protocols
  → InstalledAgent saved to .bonsai.yaml
  → generate_agent_workspace()
      → core/ templates rendered (.j2 → .md) into {workspace}/agent/Core/
      → skills/workflows/protocols .md copied into {workspace}/agent/*/
      → generate_workspace_claude_md() → {workspace}/CLAUDE.md
  → generate_root_claude_md() → updates root CLAUDE.md routing table
```

---

## File Layout (user's project after setup)

```
project/
├── .bonsai.yaml              ← project config
├── CLAUDE.md                 ← routing table (generated)
├── INDEX.md                  ← project snapshot (scaffolding)
├── Playbook/
│   ├── Status.md
│   ├── Roadmap.md
│   ├── Standards/SecurityStandards.md
│   └── Plans/Active/
├── Logs/
│   ├── FieldNotes.md
│   └── KeyDecisionLog.md
├── Reports/
│   ├── report-template.md
│   └── Pending/
└── backend/                  ← workspace (example)
    ├── CLAUDE.md             ← workspace nav (generated)
    └── agent/
        ├── Core/
        │   ├── identity.md
        │   ├── memory.md
        │   └── self-awareness.md
        ├── Skills/
        ├── Workflows/
        └── Protocols/
```
