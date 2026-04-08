"""Generates agent workspace files in the target project."""

from __future__ import annotations

import json
from pathlib import Path
from typing import TYPE_CHECKING

import jinja2

if TYPE_CHECKING:
    from bonsai.catalog import Catalog
    from bonsai.models import AgentDef, InstalledAgent, ProjectConfig

SCAFFOLDING_ROOT = Path(__file__).parent / "catalog" / "scaffolding"


def _render_template(template_path: Path, context: dict) -> str:
    """Render a Jinja2 template file with the given context."""
    env = jinja2.Environment(
        loader=jinja2.FileSystemLoader(template_path.parent),
        keep_trailing_newline=True,
        undefined=jinja2.StrictUndefined,
    )
    template = env.get_template(template_path.name)
    return template.render(context)


def _copy_or_render(src: Path, dest: Path, context: dict) -> None:
    """Copy a file, rendering it as Jinja2 if it ends with .j2."""
    dest.parent.mkdir(parents=True, exist_ok=True)
    if src.suffix == ".j2":
        content = _render_template(src, context)
        dest.with_suffix("").write_text(content)  # strip .j2 from output name
    else:
        dest.write_text(src.read_text())


def _desc_for(names: list[str], catalog: Catalog, category: str) -> dict[str, str]:
    """Build a name→description map for a list of item names."""
    getter = getattr(catalog, f"get_{category}")
    result = {}
    for name in names:
        item = getter(name)
        result[name] = item.description if item else name.replace("-", " ").title()
    return result


def generate_scaffolding(project_root: Path, config: ProjectConfig) -> list[str]:
    """Generate project management infrastructure (INDEX, Playbook, Logs, Reports).

    Returns list of created file paths (relative to project root).
    """
    docs_root = project_root / config.docs_path if config.docs_path else project_root
    context = {
        "project_name": config.project_name,
        "project_description": config.description,
    }

    created: list[str] = []

    for src_file in sorted(SCAFFOLDING_ROOT.rglob("*")):
        if not src_file.is_file():
            continue

        # Compute relative path from scaffolding root
        rel = src_file.relative_to(SCAFFOLDING_ROOT)
        dest = docs_root / rel

        # Don't overwrite existing files
        # (strip .j2 for comparison)
        final_dest = dest.with_suffix("") if dest.suffix == ".j2" else dest
        if final_dest.exists():
            continue

        _copy_or_render(src_file, dest, context)
        created.append(str(final_dest.relative_to(project_root)))

    return created


def generate_settings_json(project_root: Path, config: ProjectConfig, catalog: Catalog) -> None:
    """Generate or update .claude/settings.json with hook entries for all installed sensors."""
    settings_path = project_root / ".claude" / "settings.json"

    # Load existing settings (preserve non-hook keys like enabledPlugins)
    existing: dict = {}
    if settings_path.exists():
        try:
            existing = json.loads(settings_path.read_text())
        except json.JSONDecodeError:
            existing = {}

    # Rebuild the hooks section from scratch based on current config
    # Group by (event, matcher) → list of commands
    hook_groups: dict[tuple[str, str | None], list[str]] = {}

    for _agent_name, installed in config.agents.items():
        for sensor_name in installed.sensors:
            sensor = catalog.get_sensor(sensor_name)
            if not sensor:
                continue
            key = (sensor.event, sensor.matcher)
            script_path = f"{installed.workspace}agent/Sensors/{sensor_name}.sh"
            hook_groups.setdefault(key, []).append(f"bash {script_path}")

    # Build the hooks config structure
    hooks_config: dict[str, list] = {}
    for (event, matcher), commands in hook_groups.items():
        entry: dict = {
            "hooks": [{"type": "command", "command": cmd} for cmd in commands],
        }
        if matcher:
            entry["matcher"] = matcher
        hooks_config.setdefault(event, []).append(entry)

    if hooks_config:
        existing["hooks"] = hooks_config
    elif "hooks" in existing:
        del existing["hooks"]

    # Write settings
    settings_path.parent.mkdir(parents=True, exist_ok=True)
    settings_path.write_text(json.dumps(existing, indent=2) + "\n")


def generate_root_claude_md(project_root: Path, config: ProjectConfig) -> None:
    """Generate or update the root CLAUDE.md routing file."""
    docs_prefix = config.docs_path or ""

    lines = [
        f"# {config.project_name} — Project Router",
        "",
    ]

    # Routing table
    if config.agents:
        lines.extend([
            "## Routing",
            "",
            "| Working in | Read | Do NOT read |",
            "|------------|------|-------------|",
        ])

        for name, agent in config.agents.items():
            read = f"`{agent.workspace}CLAUDE.md`"
            do_not_read = ", ".join(
                f"`{other.workspace}CLAUDE.md`"
                for other_name, other in config.agents.items()
                if other_name != name
            )
            lines.append(f"| `{agent.workspace}` | {read} | {do_not_read or '—'} |")

        lines.extend([
            "",
            "> Read ONLY the CLAUDE.md for your workspace. Each workspace has its own agent/ directory.",
            "",
        ])

    # Universal rules
    lines.extend([
        "## Universal Rules",
        "",
        "- **Never touch another workspace's files** — stay in your lane",
        f"- **Plans live in `{docs_prefix}Playbook/Plans/`** — read your assigned plan before writing code",
        f"- **Security rules live in `{docs_prefix}Playbook/Standards/SecurityStandards.md`** — read every session",
        f"- **Logs go to `{docs_prefix}Logs/`** — write a log after completing any plan",
        "- **Attribution required** — anything written under the user's name must end with:",
        "  ```",
        "  ---",
        "  Written by **[Agent Name]** · Initiated by [source]",
        "  ```",
        "",
    ])

    # Triggers
    lines.extend([
        "## Triggers",
        "",
        "| Trigger | Action |",
        "|---------|--------|",
        f"| `status` | Read `{docs_prefix}Playbook/Status.md` and show current In Progress / Pending |",
        f"| `verify` | Run the verification suite for the current workspace |",
        "",
    ])

    (project_root / "CLAUDE.md").write_text("\n".join(lines))


def generate_workspace_claude_md(
    workspace_root: Path,
    agent_def: AgentDef,
    installed: InstalledAgent,
    config: ProjectConfig,
    catalog: Catalog,
) -> None:
    """Generate the workspace CLAUDE.md with navigation tables."""
    docs_prefix = config.docs_path or ""

    lines = [
        f"# {config.project_name} — {agent_def.display_name}",
        "",
        f"**Working directory:** `{installed.workspace}`",
        "",
        "> [!warning]",
        "> **FIRST:** Read `agent/Core/identity.md`, then `agent/Core/memory.md`.",
        "",
        "---",
        "",
        "## Navigation",
        "",
        "> All agent instruction files live in `agent/`.",
        "",
        "### Core (load first, every session)",
        "",
        "| File | Purpose |",
        "|------|---------|",
        "| `agent/Core/identity.md` | Who I am, relationships, mindset |",
        "| `agent/Core/memory.md` | Working memory — flags, work state, notes |",
        "| `agent/Core/self-awareness.md` | Context monitoring, hard thresholds |",
        "",
    ]

    if installed.protocols:
        proto_descs = _desc_for(installed.protocols, catalog, "protocol")
        lines.extend([
            "### Protocols (load after Core, every session)",
            "",
            "| File | Purpose |",
            "|------|---------|",
        ])
        for p in installed.protocols:
            lines.append(f"| `agent/Protocols/{p}.md` | {proto_descs[p]} |")
        lines.append("")

    if installed.workflows:
        wf_descs = _desc_for(installed.workflows, catalog, "workflow")
        lines.extend([
            "### Workflows (load when starting an activity)",
            "",
            "| Activity | Read this |",
            "|----------|-----------|",
        ])
        for w in installed.workflows:
            lines.append(f"| {wf_descs[w]} | `agent/Workflows/{w}.md` |")
        lines.append("")

    if installed.skills:
        skill_descs = _desc_for(installed.skills, catalog, "skill")
        lines.extend([
            "### Skills (load when doing specific work)",
            "",
            "| Need | Read this |",
            "|------|-----------|",
        ])
        for s in installed.skills:
            lines.append(f"| {skill_descs[s]} | `agent/Skills/{s}.md` |")
        lines.append("")

    if installed.sensors:
        lines.extend([
            "### Sensors (auto-enforced via hooks)",
            "",
            "| Sensor | Event | What it does |",
            "|--------|-------|-------------|",
        ])
        for sensor_name in installed.sensors:
            sensor = catalog.get_sensor(sensor_name)
            if sensor:
                event_str = sensor.event
                if sensor.matcher:
                    event_str += f" ({sensor.matcher})"
                lines.append(
                    f"| `agent/Sensors/{sensor_name}.sh` | {event_str} | {sensor.description} |"
                )
        lines.extend([
            "",
            "> Sensors run automatically — they are configured in `.claude/settings.json`.",
            "",
        ])

    # External references — point to project docs
    lines.extend([
        "### External References",
        "",
        "| Need | Read this |",
        "|------|-----------|",
        f"| Project snapshot | `{docs_prefix}INDEX.md` |",
        f"| Current work status | `{docs_prefix}Playbook/Status.md` |",
        f"| Long-term direction | `{docs_prefix}Playbook/Roadmap.md` |",
        f"| Security standards | `{docs_prefix}Playbook/Standards/SecurityStandards.md` |",
        f"| Your assigned plan | `{docs_prefix}Playbook/Plans/Active/` |",
        f"| Prior decisions | `{docs_prefix}Logs/KeyDecisionLog.md` |",
        f"| Submit report | `{docs_prefix}Reports/Pending/` |",
        "",
    ])

    (workspace_root / "CLAUDE.md").write_text("\n".join(lines))


def generate_agent_workspace(
    project_root: Path,
    agent_def: AgentDef,
    installed: InstalledAgent,
    config: ProjectConfig,
    catalog: Catalog,
) -> None:
    """Generate the full agent/ directory in a workspace."""
    workspace_root = project_root / installed.workspace
    agent_dir = workspace_root / "agent"

    # Template context for Jinja2 rendering
    context = {
        "project_name": config.project_name,
        "project_description": config.description,
        "agent_name": agent_def.name,
        "agent_display_name": agent_def.display_name,
        "agent_description": agent_def.description,
        "other_agents": [
            a for name, a in config.agents.items() if name != agent_def.name
        ],
    }

    # 1. Copy core files
    if agent_def.core_dir.exists():
        core_dest = agent_dir / "Core"
        core_dest.mkdir(parents=True, exist_ok=True)
        for src_file in sorted(agent_def.core_dir.iterdir()):
            if src_file.is_file():
                _copy_or_render(src_file, core_dest / src_file.name, context)

    # 2. Copy selected skills
    for skill_name in installed.skills:
        item = catalog.get_skill(skill_name)
        if item:
            dest = agent_dir / "Skills" / f"{skill_name}.md"
            dest.parent.mkdir(parents=True, exist_ok=True)
            dest.write_text(item.content_path.read_text())

    # 3. Copy selected workflows
    for wf_name in installed.workflows:
        item = catalog.get_workflow(wf_name)
        if item:
            dest = agent_dir / "Workflows" / f"{wf_name}.md"
            dest.parent.mkdir(parents=True, exist_ok=True)
            dest.write_text(item.content_path.read_text())

    # 4. Copy selected protocols
    for proto_name in installed.protocols:
        item = catalog.get_protocol(proto_name)
        if item:
            dest = agent_dir / "Protocols" / f"{proto_name}.md"
            dest.parent.mkdir(parents=True, exist_ok=True)
            dest.write_text(item.content_path.read_text())

    # 5. Render selected sensors
    sensor_context = {
        **context,
        "workspace": installed.workspace,
        "docs_path": config.docs_path or "",
        "protocols": installed.protocols,
        "skills": installed.skills,
        "workflows": installed.workflows,
    }
    for sensor_name in installed.sensors:
        sensor = catalog.get_sensor(sensor_name)
        if sensor:
            dest = agent_dir / "Sensors" / sensor.content_path.name
            _copy_or_render(sensor.content_path, dest, sensor_context)
            # Make rendered scripts executable
            final = dest.with_suffix("") if dest.suffix == ".j2" else dest
            if final.exists():
                final.chmod(final.stat().st_mode | 0o111)

    # 6. Generate workspace CLAUDE.md
    generate_workspace_claude_md(
        workspace_root, agent_def, installed, config, catalog
    )
