---
tags: [core, memory]
description: Tech Lead Agent working memory — flags, work state, notes.
---

# Working Memory

## Flags

(none)

## Work State

**Current task:** (none)
**Blocked on:** (nothing)

**Completed:**
- Catalog expansion (all 3 phases) — see RESEARCH-catalog-expansion.md
- Final agent lineup: tech-lead, fullstack, backend, frontend, devops, security
- Managed Agents analysis — decided to defer cloud integration until local foundation is solid
- Lock file conflict handling — `.bonsai-lock.yaml` tracks generated files with sha256 hashes, prevents silent overwrites of user-modified files
- Awareness Framework — `status-bar` (Stop) + `context-guard` (UserPromptSubmit) sensors, updated self-awareness.md
- Dogfooding — ran `bonsai init` on itself, generated station/ workspace with tech-lead, migrated all content from hand-crafted agent/ to generated station/agent/
- Station customization — tailored INDEX.md, Roadmap.md, Status.md, KeyDecisionLog.md, SecurityStandards.md from generic templates to Bonsai-specific content
- Session wrap-up workflow — created agent/Workflows/session-wrapup.md, wired to context-guard trigger detection
- Stale artifact cleanup — rewrote agent/index.md for Go codebase, fixed Python references in RESEARCH.md

## Notes

- Go 1.24+ required — see `go.mod`
- Build: `make build` → `./bonsai`
- Install: `go install .` → `$GOPATH/bin/bonsai`
- Stack: Cobra (CLI), Huh (forms), LipGloss (styling), BubbleTea (TUI)
- security-audit workflow already created (shared between devops and security agents)
- Routine .md.tmpl format: `1. **Bold step:**` with sub-bullets, not H3 headers. Include `**Frequency:** Every N days` after H1.
- session-wrapup.md is a custom workflow (not in catalog) — `.bonsai.yaml` only tracks catalog items, so it's not listed there. This is expected until custom item detection is built (P2 backlog).

## Feedback

_(empty)_

## References

- RESEARCH-catalog-expansion.md — full spec for all new agents, skills, sensors, workflows, routines
- DESIGN-companion-app.md — Greenhouse design doc (architecture, tech stack, integration, UI, data model)
- RESEARCH-trigger-system.md — Trigger system research: determinism taxonomy, proposed hybrid design
- Claude Code Agent SDK docs — https://code.claude.com/docs/en/agent-sdk/overview
- Claude Managed Agents — https://platform.claude.com/docs/en/managed-agents/overview (future cloud integration)
