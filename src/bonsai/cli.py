"""Bonsai CLI — scaffold Claude Code agent workspaces."""

from __future__ import annotations

import shutil
from pathlib import Path

import typer
from InquirerPy import inquirer
from rich.console import Console
from rich.panel import Panel
from rich.rule import Rule
from rich.text import Text
from rich.tree import Tree

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
        console.print(
            Panel(
                f"No {CONFIG_FILE} found.\nRun [bold]bonsai init[/bold] first.",
                border_style="red",
                title="Error",
            )
        )
        raise typer.Exit(1)
    return config_path, ProjectConfig.load(config_path)


def _prompt_text(message: str, default: str = "", validate: bool = False) -> str:
    """Text prompt with InquirerPy."""
    kwargs: dict = {"message": message, "default": default}
    if validate:
        kwargs["validate"] = lambda x: len(x.strip()) > 0 or "Cannot be empty"
    result = inquirer.text(**kwargs).execute()
    if result is None:
        raise typer.Abort()
    return result


def _pick_items(
    label: str,
    available: list[CatalogItem],
    defaults: list[str],
) -> list[str]:
    """Show an interactive checkbox for a list of catalog items."""
    if not available:
        return []

    console.print()
    console.print(Rule(f"[bold]{label}", style="dim"))
    console.print()

    choices = [
        {
            "name": f"{item.name} — {item.description}",
            "value": item.name,
            "enabled": item.name in defaults,
        }
        for item in available
    ]
    selected = inquirer.checkbox(
        message=f"Select {label.lower()}:",
        choices=choices,
        cycle=True,
        instruction="(space to toggle, enter to confirm)",
        transformer=lambda results: ", ".join(
            r.split(" — ")[0] for r in results
        ) if results else "none",
    ).execute()
    if selected is None:
        raise typer.Abort()
    return selected


def _build_file_tree(files: list[str], root_label: str) -> Tree:
    """Build a Rich Tree from a flat list of relative file paths."""
    tree = Tree(f"[bold]{root_label}")
    nodes: dict[str, Tree] = {}

    for filepath in sorted(files):
        parts = filepath.split("/")
        current_key = ""
        parent = tree
        for i, part in enumerate(parts):
            current_key = f"{current_key}/{part}" if current_key else part
            if i == len(parts) - 1:
                # Leaf file
                parent.add(f"[dim]{part}")
            elif current_key not in nodes:
                nodes[current_key] = parent.add(f"[bold cyan]{part}/")
                parent = nodes[current_key]
            else:
                parent = nodes[current_key]

    return tree


# ── commands ─────────────────────────────────────────────────────────────────


@app.command()
def init() -> None:
    """Initialize Bonsai in the current project."""
    config_path = Path.cwd() / CONFIG_FILE

    if config_path.exists():
        console.print(
            Panel(
                f"{CONFIG_FILE} already exists. Skipping init.",
                border_style="yellow",
                title="Warning",
            )
        )
        raise typer.Exit(0)

    console.print()
    console.print(Rule("[bold]Initialize Project", style="blue"))
    console.print()

    project_name = _prompt_text("Project name:", validate=True)
    description = _prompt_text("Description (optional):")
    docs_path = _prompt_text(
        "Docs directory (blank for root, e.g. 'docs/'):"
    )
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

    with console.status("[bold]Generating project files..."):
        generate_root_claude_md(project_root, config)
        created = generate_scaffolding(project_root, config)

    console.print()
    if created:
        tree = _build_file_tree(created, docs_path or ".")
        console.print(Panel(tree, title="Created Files", border_style="green"))

    console.print()
    console.print(f"  [green]✓[/green] Initialized [bold]{project_name}[/bold]")
    console.print(f"  [dim]Next: run [bold]bonsai add[/bold] to add an agent.[/dim]")
    console.print()


@app.command()
def add() -> None:
    """Add an agent to the project."""
    config_path, config = _require_config()
    cat = Catalog()

    console.print()
    console.print(Rule("[bold]Add Agent", style="blue"))
    console.print()

    # 1. Pick agent type
    agent_choices = [
        {
            "name": f"{a.display_name} — {a.description}",
            "value": a.name,
        }
        for a in cat.agents
    ]
    agent_type = inquirer.select(
        message="Agent type:",
        choices=agent_choices,
        cycle=True,
    ).execute()
    if agent_type is None:
        raise typer.Abort()

    if agent_type in config.agents:
        console.print(
            Panel(
                f"Agent [bold]{agent_type}[/bold] is already installed.",
                border_style="yellow",
                title="Warning",
            )
        )
        raise typer.Exit(0)

    agent_def = cat.get_agent(agent_type)
    if agent_def is None:
        console.print(f"  [red]✗[/red] Unknown agent type: {agent_type}")
        raise typer.Exit(1)

    # 2. Workspace directory
    existing_workspaces = {a.workspace for a in config.agents.values()}
    workspace = inquirer.text(
        message="Workspace directory (e.g. backend/):",
        validate=lambda x: len(x.strip()) > 0 or "Workspace cannot be empty",
        default=f"{agent_type}/",
    ).execute()
    if workspace is None:
        raise typer.Abort()
    workspace = workspace.strip().rstrip("/") + "/"

    if workspace in existing_workspaces:
        console.print(
            Panel(
                f"Workspace [bold]{workspace}[/bold] is already in use.",
                border_style="red",
                title="Error",
            )
        )
        raise typer.Exit(1)

    # 3. Pick components
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

    # 4. Confirm — build a rich summary
    def _item_desc(name: str) -> str:
        item = cat.get_item(name) or cat.get_sensor(name)
        return item.description if item else ""

    summary_tree = Tree(
        f"[bold cyan]{agent_def.display_name}[/bold cyan]  [dim]→ {workspace}[/dim]"
    )

    categories = [
        ("Skills", selected_skills),
        ("Workflows", selected_workflows),
        ("Protocols", selected_protocols),
        ("Sensors", selected_sensors),
    ]
    for cat_label, items in categories:
        if items:
            branch = summary_tree.add(f"[bold]{cat_label}[/bold] [dim]({len(items)})[/dim]")
            for item in items:
                desc = _item_desc(item)
                branch.add(f"{item}  [dim]{desc}[/dim]")
        else:
            summary_tree.add(f"[dim]{cat_label} — none[/dim]")

    console.print()
    console.print(Panel(summary_tree, title="Review", border_style="blue", padding=(1, 2)))
    console.print()

    proceed = inquirer.confirm(message="Generate files?", default=True).execute()
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

    with console.status("[bold]Generating workspace..."):
        generate_agent_workspace(project_root, agent_def, installed, config, cat)
        generate_root_claude_md(project_root, config)
        generate_settings_json(project_root, config, cat)

    console.print()
    console.print(
        f"  [green]✓[/green] Added [bold]{agent_def.display_name}[/bold] "
        f"at [dim]{workspace}[/dim]"
    )
    console.print()


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
        console.print(f"  [red]✗[/red] Agent [bold]{agent_name}[/bold] is not installed.")
        raise typer.Exit(1)

    agent = config.agents[agent_name]
    workspace = agent.workspace

    proceed = inquirer.confirm(
        message=f"Remove {agent_name} (workspace: {workspace})?",
        default=False,
    ).execute()
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
            console.print(f"  [dim]Deleted {agent_dir}[/dim]")
        if claude_md.exists():
            claude_md.unlink()
            console.print(f"  [dim]Deleted {claude_md}[/dim]")

    console.print(f"  [green]✓[/green] Removed [bold]{agent_name}[/bold]")
    console.print()


@app.command("list")
def list_agents() -> None:
    """Show installed agents and their components."""
    _, config = _require_config()

    if not config.agents:
        console.print()
        console.print(
            Panel(
                "No agents installed.\nRun [bold]bonsai add[/bold] to get started.",
                border_style="dim",
                title="Empty",
            )
        )
        return

    console.print()
    for name, agent in config.agents.items():
        header = Text()
        header.append(f"{name}", style="bold cyan")
        header.append(f"  {agent.workspace}", style="dim")

        tree = Tree(header)

        if agent.skills:
            skills_branch = tree.add("[bold]Skills")
            for s in agent.skills:
                skills_branch.add(f"[dim]{s}")

        if agent.workflows:
            wf_branch = tree.add("[bold]Workflows")
            for w in agent.workflows:
                wf_branch.add(f"[dim]{w}")

        if agent.protocols:
            proto_branch = tree.add("[bold]Protocols")
            for p in agent.protocols:
                proto_branch.add(f"[dim]{p}")

        if agent.sensors:
            sensor_branch = tree.add("[bold]Sensors")
            for s in agent.sensors:
                sensor_branch.add(f"[dim]{s}")

        console.print(
            Panel(tree, title=config.project_name, border_style="blue")
        )
    console.print()


@app.command("catalog")
def show_catalog(
    agent: str = typer.Option(
        None, "--agent", "-a",
        help="Filter to items compatible with this agent type",
    ),
) -> None:
    """Browse available agents, skills, workflows, and protocols."""
    cat = Catalog()
    suffix = f" [dim](for {agent})[/dim]" if agent else ""

    console.print()

    # Agents
    console.print(Rule(f"[bold]Agents", style="blue"))
    console.print()
    for a in cat.agents:
        console.print(f"  [bold cyan]{a.name}[/bold cyan]  [dim]{a.description}[/dim]")
    console.print()

    # Items
    for label, items_fn in [
        ("Skills", cat.skills_for if agent else lambda _: cat.skills),
        ("Workflows", cat.workflows_for if agent else lambda _: cat.workflows),
        ("Protocols", cat.protocols_for if agent else lambda _: cat.protocols),
    ]:
        items = items_fn(agent) if agent else items_fn(None)
        console.print(Rule(f"[bold]{label}{suffix}", style="blue"))
        console.print()
        for item in items:
            agents_str = (
                "[dim]all[/dim]" if item.agents == "all"
                else f"[dim]{', '.join(item.agents)}[/dim]"
            )
            console.print(f"  [bold]{item.name}[/bold]  {item.description}  {agents_str}")
        console.print()

    # Sensors
    sensors = cat.sensors_for(agent) if agent else cat.sensors
    console.print(Rule(f"[bold]Sensors{suffix}", style="blue"))
    console.print()
    for s in sensors:
        agents_str = (
            "[dim]all[/dim]" if s.agents == "all"
            else f"[dim]{', '.join(s.agents)}[/dim]"
        )
        event_str = f"[cyan]{s.event}[/cyan]"
        if s.matcher:
            event_str += f" [dim]({s.matcher})[/dim]"
        console.print(f"  [bold]{s.name}[/bold]  {s.description}  {event_str}  {agents_str}")
    console.print()


def main() -> None:
    app()
