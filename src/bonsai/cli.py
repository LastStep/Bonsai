"""Bonsai CLI — scaffold Claude Code agent workspaces."""

from __future__ import annotations

import shutil
from pathlib import Path

import questionary
import typer
from rich import print as rprint
from rich.console import Console
from rich.table import Table

from bonsai.catalog import Catalog
from bonsai.generator import (
    generate_agent_workspace,
    generate_root_claude_md,
    generate_scaffolding,
    generate_settings_json,
)
from bonsai.models import CatalogItem, InstalledAgent, ProjectConfig

app = typer.Typer(
    name="bonsai",
    help="Scaffold Claude Code agent workspaces for your project.",
    no_args_is_help=True,
)
console = Console()

CONFIG_FILE = ".bonsai.yaml"


# ── helpers ──────────────────────────────────────────────────────────────────


def _require_config() -> tuple[Path, ProjectConfig]:
    """Load .bonsai.yaml or exit with an error."""
    config_path = Path.cwd() / CONFIG_FILE
    if not config_path.exists():
        rprint(
            f"[red]No {CONFIG_FILE} found. Run [bold]bonsai init[/bold] first.[/red]"
        )
        raise typer.Exit(1)
    return config_path, ProjectConfig.load(config_path)


def _pick_items(
    label: str,
    available: list[CatalogItem],
    defaults: list[str],
) -> list[str]:
    """Show an interactive checkbox for a list of catalog items. Returns selected names."""
    if not available:
        return []

    choices = [
        questionary.Choice(
            title=f"{item.name} — {item.description}",
            value=item.name,
            checked=item.name in defaults,
        )
        for item in available
    ]
    selected = questionary.checkbox(
        f"{label} (space to toggle, enter to confirm):",
        choices=choices,
    ).ask()
    if selected is None:
        raise typer.Abort()
    return selected


# ── commands ─────────────────────────────────────────────────────────────────


@app.command()
def init() -> None:
    """Initialize Bonsai in the current project."""
    config_path = Path.cwd() / CONFIG_FILE

    if config_path.exists():
        rprint(f"[yellow]{CONFIG_FILE} already exists. Skipping init.[/yellow]")
        raise typer.Exit(0)

    project_name = questionary.text(
        "Project name:",
        validate=lambda x: len(x.strip()) > 0 or "Name cannot be empty",
    ).ask()
    if project_name is None:
        raise typer.Abort()

    description = questionary.text(
        "Project description (optional):",
    ).ask()
    if description is None:
        raise typer.Abort()

    docs_path = questionary.text(
        "Where should project docs live? (leave blank for project root, or e.g. 'docs/'):",
    ).ask()
    if docs_path is None:
        raise typer.Abort()
    docs_path = docs_path.strip()
    if docs_path and not docs_path.endswith("/"):
        docs_path += "/"

    config = ProjectConfig(
        project_name=project_name.strip(),
        description=description.strip(),
        docs_path=docs_path,
    )
    config.save(config_path)

    project_root = Path.cwd()
    generate_root_claude_md(project_root, config)
    created = generate_scaffolding(project_root, config)

    rprint(f"\n[green]Initialized Bonsai for [bold]{project_name}[/bold].[/green]")
    if created:
        rprint(f"  Created {len(created)} project files in '{docs_path or '.'}':")
        for f in created[:8]:
            rprint(f"    {f}")
        if len(created) > 8:
            rprint(f"    ... and {len(created) - 8} more")
    rprint("\nNext: run [bold]bonsai add[/bold] to add an agent.")


@app.command()
def add() -> None:
    """Add an agent to the project."""
    config_path, config = _require_config()
    cat = Catalog()

    # 1. Pick agent type
    agent_choices = [
        questionary.Choice(
            title=f"{a.display_name} — {a.description}",
            value=a.name,
        )
        for a in cat.agents
    ]
    agent_type = questionary.select(
        "Which agent type?",
        choices=agent_choices,
    ).ask()
    if agent_type is None:
        raise typer.Abort()

    if agent_type in config.agents:
        rprint(f"[yellow]Agent '{agent_type}' is already installed.[/yellow]")
        raise typer.Exit(0)

    agent_def = cat.get_agent(agent_type)
    if agent_def is None:
        rprint(f"[red]Unknown agent type: {agent_type}[/red]")
        raise typer.Exit(1)

    # 2. Ask workspace directory
    existing_workspaces = {a.workspace for a in config.agents.values()}
    workspace = questionary.text(
        "Workspace directory (e.g. backend/):",
        validate=lambda x: (
            len(x.strip()) > 0 or "Cannot be empty"
        ),
    ).ask()
    if workspace is None:
        raise typer.Abort()
    workspace = workspace.strip().rstrip("/") + "/"

    if workspace in existing_workspaces:
        rprint(
            f"[red]Workspace '{workspace}' is already used by another agent.[/red]"
        )
        raise typer.Exit(1)

    # 3. Pick skills, workflows, protocols
    selected_skills = _pick_items(
        "Skills", cat.skills_for(agent_type), agent_def.default_skills
    )
    selected_workflows = _pick_items(
        "Workflows", cat.workflows_for(agent_type), agent_def.default_workflows
    )
    selected_protocols = _pick_items(
        "Protocols", cat.protocols_for(agent_type), agent_def.default_protocols
    )
    selected_sensors = _pick_items(
        "Sensors", cat.sensors_for(agent_type), agent_def.default_sensors
    )

    # 4. Confirm
    rprint(f"\n[bold]Summary:[/bold]")
    rprint(f"  Agent:     {agent_def.display_name}")
    rprint(f"  Workspace: {workspace}")
    rprint(f"  Skills:    {', '.join(selected_skills) or '(none)'}")
    rprint(f"  Workflows: {', '.join(selected_workflows) or '(none)'}")
    rprint(f"  Protocols: {', '.join(selected_protocols) or '(none)'}")
    rprint(f"  Sensors:   {', '.join(selected_sensors) or '(none)'}")

    proceed = questionary.confirm("Generate files?", default=True).ask()
    if not proceed:
        raise typer.Abort()

    # 5. Generate
    installed = InstalledAgent(
        agent_type=agent_type,
        workspace=workspace,
        skills=selected_skills,
        workflows=selected_workflows,
        protocols=selected_protocols,
        sensors=selected_sensors,
    )
    config.agents[agent_type] = installed
    config.save(config_path)

    project_root = Path.cwd()
    generate_agent_workspace(project_root, agent_def, installed, config, cat)
    generate_root_claude_md(project_root, config)
    generate_settings_json(project_root, config, cat)

    rprint(
        f"\n[green]Added [bold]{agent_def.display_name}[/bold] "
        f"at [bold]{workspace}[/bold][/green]"
    )


@app.command()
def remove(
    agent_name: str = typer.Argument(
        ..., help="Agent type to remove (e.g. backend)"
    ),
    delete_files: bool = typer.Option(
        False, "--delete-files", "-d",
        help="Also delete the generated agent/ directory",
    ),
) -> None:
    """Remove an installed agent from the project."""
    config_path, config = _require_config()

    if agent_name not in config.agents:
        rprint(f"[red]Agent '{agent_name}' is not installed.[/red]")
        raise typer.Exit(1)

    agent = config.agents[agent_name]
    workspace = agent.workspace

    proceed = questionary.confirm(
        f"Remove {agent_name} (workspace: {workspace})?", default=False
    ).ask()
    if not proceed:
        raise typer.Abort()

    # Remove from config
    del config.agents[agent_name]
    config.save(config_path)
    project_root = Path.cwd()
    generate_root_claude_md(project_root, config)
    generate_settings_json(project_root, config, Catalog())

    # Optionally delete generated files
    if delete_files:
        agent_dir = Path.cwd() / workspace / "agent"
        claude_md = Path.cwd() / workspace / "CLAUDE.md"
        if agent_dir.exists():
            shutil.rmtree(agent_dir)
            rprint(f"  Deleted {agent_dir}")
        if claude_md.exists():
            claude_md.unlink()
            rprint(f"  Deleted {claude_md}")

    rprint(f"[green]Removed [bold]{agent_name}[/bold].[/green]")


@app.command("list")
def list_agents() -> None:
    """Show installed agents and their components."""
    _, config = _require_config()

    if not config.agents:
        rprint("[yellow]No agents installed. Run [bold]bonsai add[/bold].[/yellow]")
        return

    table = Table(title=f"{config.project_name} — Installed Agents")
    table.add_column("Agent", style="bold cyan")
    table.add_column("Workspace", style="dim")
    table.add_column("Skills")
    table.add_column("Workflows")
    table.add_column("Protocols")
    table.add_column("Sensors")

    for name, agent in config.agents.items():
        table.add_row(
            name,
            agent.workspace,
            "\n".join(agent.skills) or "—",
            "\n".join(agent.workflows) or "—",
            "\n".join(agent.protocols) or "—",
            "\n".join(agent.sensors) or "—",
        )

    console.print(table)


@app.command("catalog")
def show_catalog(
    agent: str = typer.Option(
        None, "--agent", "-a",
        help="Filter to items compatible with this agent type",
    ),
) -> None:
    """Browse available agents, skills, workflows, and protocols."""
    cat = Catalog()

    # Agents table
    agent_table = Table(title="Agents")
    agent_table.add_column("Name", style="bold cyan")
    agent_table.add_column("Description")
    agent_table.add_column("Default Skills", style="dim")
    agent_table.add_column("Default Workflows", style="dim")
    agent_table.add_column("Default Protocols", style="dim")
    agent_table.add_column("Default Sensors", style="dim")
    for a in cat.agents:
        agent_table.add_row(
            a.name,
            a.description,
            ", ".join(a.default_skills) or "—",
            ", ".join(a.default_workflows) or "—",
            ", ".join(a.default_protocols) or "—",
            ", ".join(a.default_sensors) or "—",
        )
    console.print(agent_table)
    console.print()

    # Items tables
    for label, items_fn in [
        ("Skills", cat.skills_for if agent else lambda _: cat.skills),
        ("Workflows", cat.workflows_for if agent else lambda _: cat.workflows),
        ("Protocols", cat.protocols_for if agent else lambda _: cat.protocols),
    ]:
        items = items_fn(agent) if agent else items_fn(None)
        table = Table(title=f"{label}" + (f" (for {agent})" if agent else ""))
        table.add_column("Name", style="bold")
        table.add_column("Description")
        table.add_column("Compatible Agents", style="dim")
        for item in items:
            agents_str = (
                "all" if item.agents == "all" else ", ".join(item.agents)
            )
            table.add_row(item.name, item.description, agents_str)
        console.print(table)
        console.print()

    # Sensors table
    sensors = cat.sensors_for(agent) if agent else cat.sensors
    sensor_table = Table(title="Sensors" + (f" (for {agent})" if agent else ""))
    sensor_table.add_column("Name", style="bold")
    sensor_table.add_column("Description")
    sensor_table.add_column("Event", style="cyan")
    sensor_table.add_column("Matcher", style="dim")
    sensor_table.add_column("Compatible Agents", style="dim")
    for s in sensors:
        agents_str = "all" if s.agents == "all" else ", ".join(s.agents)
        sensor_table.add_row(
            s.name, s.description, s.event, s.matcher or "—", agents_str
        )
    console.print(sensor_table)
    console.print()


def main() -> None:
    app()
