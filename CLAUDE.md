# Bonsai — Developer Agent

**Codename:** Bonsai
**What:** CLI tool for scaffolding Claude Code agent workspaces — `pip install bonsai-agents`
**Stack:** Python 3.10+, Typer, Pydantic V2, Jinja2, questionary, Rich

> [!warning]
> **FIRST:** Read `agent/Core/identity.md`, then `agent/Core/memory.md`.

---

## Project Structure

```
Bonsai/
├── CLAUDE.md               ← you are here
├── pyproject.toml           ← package config, entry point: bonsai.cli:main
├── src/bonsai/
│   ├── __init__.py
│   ├── cli.py               ← Typer CLI — init, add, remove, list, catalog
│   ├── models.py             ← Pydantic models — ProjectConfig, InstalledAgent, CatalogItem, SensorItem, AgentDef
│   ├── catalog.py            ← loads YAML metadata from catalog/
│   ├── generator.py          ← renders templates, writes files to target project
│   └── catalog/              ← bundled catalog (ships with the package)
│       ├── agents/           ← agent type definitions + core templates
│       │   ├── tech-lead/
│       │   ├── backend/
│       │   └── frontend/
│       ├── skills/           ← à la carte skills (meta.yaml + content.md)
│       ├── workflows/        ← à la carte workflows
│       ├── protocols/        ← à la carte protocols
│       ├── sensors/          ← auto-enforced hooks (meta.yaml + script.sh.j2)
│       └── scaffolding/      ← project management infrastructure templates
│           ├── INDEX.md.j2
│           ├── Playbook/     ← Status, Roadmap, Plans, SecurityStandards
│           ├── Logs/         ← FieldNotes, KeyDecisionLog
│           └── Reports/      ← report-template, Pending/
├── tests/
├── agent/                    ← agent instructions (this agent)
└── .venv/
```

---

## Key Concepts

- **Catalog items** (skills, workflows, protocols) each have a `meta.yaml` with `name`, `description`, `agents` (list or `"all"`) and a companion `.md` content file
- **Sensors** are auto-enforced hooks — `meta.yaml` adds `event` (hook event) and optional `matcher` (tool filter), with a companion `.sh.j2` script template instead of `.md`
- **Agent definitions** have an `agent.yaml` with `name`, `display_name`, `description`, `defaults` and a `core/` directory with `.j2` identity templates
- **Scaffolding** templates use Jinja2 (`.j2` extension) with `{{ project_name }}`, `{{ project_description }}` context vars
- **`.bonsai.yaml`** is the project config generated in the user's target project — tracks installed agents and docs_path
- **`.claude/settings.json`** is auto-generated with hook entries for all installed sensors
- **Generator** never overwrites existing files — safe to re-run (except settings.json hooks, which are rebuilt from config)

---

## Development

```bash
source .venv/bin/activate
pip install -e .           # editable install
bonsai --help              # verify CLI works
```

### Testing changes to catalog items

Edit files in `src/bonsai/catalog/`, then test in a temp dir:
```bash
mkdir /tmp/test && cd /tmp/test
bonsai init
bonsai add
bonsai list
```

### Adding a new catalog item (skill, workflow, protocol)

1. Create `src/bonsai/catalog/{category}/{item-name}/meta.yaml`
2. Create `src/bonsai/catalog/{category}/{item-name}/{item-name}.md`
3. Set `agents:` in meta.yaml to control compatibility

### Adding a new sensor

1. Create `src/bonsai/catalog/sensors/{name}/meta.yaml` — must include `event` and optionally `matcher`
2. Create `src/bonsai/catalog/sensors/{name}/{name}.sh.j2` — script template
3. Available events: `SessionStart`, `PreToolUse`, `PostToolUse`, `Stop`, etc.
4. Template context includes: `project_name`, `agent_name`, `agent_display_name`, `workspace`, `docs_path`, `other_agents`, `protocols`, `skills`, `workflows`

### Adding a new agent type

1. Create `src/bonsai/catalog/agents/{name}/agent.yaml`
2. Create `src/bonsai/catalog/agents/{name}/core/identity.md.j2` (+ memory.md.j2, self-awareness.md)
3. Set `defaults:` in agent.yaml to pre-select items

---

## Conventions

- Keep CLI interactive — use questionary for all user input
- All catalog items use the same base `meta.yaml` shape: `name`, `description`, `agents` — sensors add `event` and `matcher`
- Generator functions go in `generator.py`, catalog loading in `catalog.py`, CLI in `cli.py`
- Pydantic models for all data shapes
- Don't break the existing CLI commands — they're the public API
