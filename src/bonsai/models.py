"""Data models for Bonsai catalog items and project config."""

from __future__ import annotations

from pathlib import Path
from typing import Literal

from pydantic import BaseModel, ConfigDict


class CatalogItem(BaseModel):
    """Base for any catalog item (skill, workflow, protocol)."""

    model_config = ConfigDict(frozen=True)

    name: str
    description: str
    agents: list[str] | Literal["all"]
    content_path: Path  # resolved absolute path to the .md file

    def compatible_with(self, agent_type: str) -> bool:
        return self.agents == "all" or agent_type in self.agents


class SensorItem(CatalogItem):
    """A sensor (hook) — auto-enforced behavior via Claude Code hooks.

    content_path points to a script template (.sh.j2 or .sh).
    """

    event: str  # Hook event: SessionStart, PreToolUse, PostToolUse, etc.
    matcher: str | None = None  # Tool matcher for Pre/PostToolUse (e.g. "Edit|Write")


class AgentDef(BaseModel):
    """An agent type definition from the catalog."""

    model_config = ConfigDict(frozen=True)

    name: str
    display_name: str
    description: str
    default_skills: list[str]
    default_workflows: list[str]
    default_protocols: list[str]
    default_sensors: list[str]
    core_dir: Path  # resolved absolute path to core/ template dir


class InstalledAgent(BaseModel):
    """An agent installed in a project — stored in .bonsai.yaml."""

    agent_type: str
    workspace: str
    skills: list[str]
    workflows: list[str]
    protocols: list[str]
    sensors: list[str] = []


class ProjectConfig(BaseModel):
    """Root project config — serialized to .bonsai.yaml."""

    project_name: str
    description: str = ""
    docs_path: str = ""  # where project management docs live (e.g. "docs/")
    agents: dict[str, InstalledAgent] = {}

    def save(self, path: Path) -> None:
        import yaml

        data = self.model_dump(mode="json")
        path.write_text(yaml.dump(data, default_flow_style=False, sort_keys=False))

    @classmethod
    def load(cls, path: Path) -> ProjectConfig:
        import yaml

        data = yaml.safe_load(path.read_text())
        return cls.model_validate(data)
