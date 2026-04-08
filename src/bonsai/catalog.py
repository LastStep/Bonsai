"""Discovers and loads catalog items from the bundled catalog/ directory."""

from __future__ import annotations

from pathlib import Path

import yaml

from bonsai.models import AgentDef, CatalogItem, SensorItem

CATALOG_ROOT = Path(__file__).parent / "catalog"


def _load_items(category: str) -> list[CatalogItem]:
    """Load all items from a catalog category (skills, workflows, protocols)."""
    category_dir = CATALOG_ROOT / category
    items: list[CatalogItem] = []

    if not category_dir.exists():
        return items

    for item_dir in sorted(category_dir.iterdir()):
        if not item_dir.is_dir():
            continue

        meta_path = item_dir / "meta.yaml"
        if not meta_path.exists():
            continue

        try:
            meta = yaml.safe_load(meta_path.read_text())
        except yaml.YAMLError:
            continue

        if not meta or "name" not in meta or "description" not in meta:
            continue

        # Find the content .md file — prefer one matching the item name
        md_files = sorted(item_dir.glob("*.md"))
        if not md_files:
            continue

        # Prefer the file whose stem matches the item name
        content_path = next(
            (f for f in md_files if f.stem == meta["name"]),
            md_files[0],
        )

        items.append(
            CatalogItem(
                name=meta["name"],
                description=meta["description"],
                agents=meta.get("agents", "all"),
                content_path=content_path,
            )
        )

    return items


def _load_sensors() -> list[SensorItem]:
    """Load all sensor definitions from the catalog."""
    sensors_dir = CATALOG_ROOT / "sensors"
    items: list[SensorItem] = []

    if not sensors_dir.exists():
        return items

    for item_dir in sorted(sensors_dir.iterdir()):
        if not item_dir.is_dir():
            continue

        meta_path = item_dir / "meta.yaml"
        if not meta_path.exists():
            continue

        try:
            meta = yaml.safe_load(meta_path.read_text())
        except yaml.YAMLError:
            continue

        if not meta or "name" not in meta or "event" not in meta:
            continue

        # Find the script file — anything that's not meta.yaml
        script_files = sorted(
            f for f in item_dir.iterdir()
            if f.is_file() and f.name != "meta.yaml"
        )
        if not script_files:
            continue

        # Prefer file whose stem (minus .j2) matches the item name
        content_path = next(
            (f for f in script_files if f.stem.split(".")[0] == meta["name"]),
            script_files[0],
        )

        items.append(
            SensorItem(
                name=meta["name"],
                description=meta.get("description", ""),
                agents=meta.get("agents", "all"),
                event=meta["event"],
                matcher=meta.get("matcher"),
                content_path=content_path,
            )
        )

    return items


def _load_agents() -> list[AgentDef]:
    """Load all agent definitions from the catalog."""
    agents_dir = CATALOG_ROOT / "agents"
    agents: list[AgentDef] = []

    if not agents_dir.exists():
        return agents

    for agent_dir in sorted(agents_dir.iterdir()):
        if not agent_dir.is_dir():
            continue

        agent_yaml = agent_dir / "agent.yaml"
        if not agent_yaml.exists():
            continue

        try:
            data = yaml.safe_load(agent_yaml.read_text())
        except yaml.YAMLError:
            continue

        if not data or "name" not in data:
            continue

        agents.append(
            AgentDef(
                name=data["name"],
                display_name=data.get("display_name", data["name"]),
                description=data.get("description", ""),
                default_skills=data.get("defaults", {}).get("skills", []),
                default_workflows=data.get("defaults", {}).get("workflows", []),
                default_protocols=data.get("defaults", {}).get("protocols", []),
                default_sensors=data.get("defaults", {}).get("sensors", []),
                core_dir=agent_dir / "core",
            )
        )

    return agents


class Catalog:
    """Loaded catalog — agents + items, with filtering helpers."""

    def __init__(self) -> None:
        self.agents = _load_agents()
        self.skills = _load_items("skills")
        self.workflows = _load_items("workflows")
        self.protocols = _load_items("protocols")
        self.sensors = _load_sensors()

        # Build lookup dicts for O(1) access by name
        self._skills_by_name = {s.name: s for s in self.skills}
        self._workflows_by_name = {w.name: w for w in self.workflows}
        self._protocols_by_name = {p.name: p for p in self.protocols}
        self._sensors_by_name = {s.name: s for s in self.sensors}

    def get_agent(self, name: str) -> AgentDef | None:
        return next((a for a in self.agents if a.name == name), None)

    def get_skill(self, name: str) -> CatalogItem | None:
        return self._skills_by_name.get(name)

    def get_workflow(self, name: str) -> CatalogItem | None:
        return self._workflows_by_name.get(name)

    def get_protocol(self, name: str) -> CatalogItem | None:
        return self._protocols_by_name.get(name)

    def get_sensor(self, name: str) -> SensorItem | None:
        return self._sensors_by_name.get(name)

    def get_item(self, name: str) -> CatalogItem | None:
        """Look up any item by name across all categories."""
        return (
            self.get_skill(name)
            or self.get_workflow(name)
            or self.get_protocol(name)
        )

    def skills_for(self, agent_type: str) -> list[CatalogItem]:
        return [s for s in self.skills if s.compatible_with(agent_type)]

    def workflows_for(self, agent_type: str) -> list[CatalogItem]:
        return [w for w in self.workflows if w.compatible_with(agent_type)]

    def protocols_for(self, agent_type: str) -> list[CatalogItem]:
        return [p for p in self.protocols if p.compatible_with(agent_type)]

    def sensors_for(self, agent_type: str) -> list[SensorItem]:
        return [s for s in self.sensors if s.compatible_with(agent_type)]
