---
tags: [core, memory]
description: Bonsai developer working memory — flags, work state, notes.
---

# Working Memory

## Flags

(none)

## Work State

**Current task:** None active
**Blocked on:** (nothing)

**Completed:**
- Catalog expansion (all 3 phases) — see RESEARCH-catalog-expansion.md
- Final agent lineup: tech-lead, fullstack, backend, frontend, devops, security
- Managed Agents analysis — decided to defer cloud integration until local foundation is solid

## Notes

- Go 1.24+ required — see `go.mod`
- Build: `make build` → `./bonsai`
- Install: `go install .` → `$GOPATH/bin/bonsai`
- Stack: Cobra (CLI), Huh (forms), LipGloss (styling), BubbleTea (TUI)
- security-audit workflow already created (shared between devops and future security agent)
- Routine .md.tmpl format: `1. **Bold step:**` with sub-bullets, not H3 headers. Include `**Frequency:** Every N days` after H1.

## Feedback

_(empty)_

## References

- Backlog.md — prioritized todo queue (bugs, features, debt, research). Add items here, not in memory.md
- RESEARCH-catalog-expansion.md — full spec for all new agents, skills, sensors, workflows, routines
- DESIGN-companion-app.md — Greenhouse design doc (architecture, tech stack, integration, UI, data model)
- Claude Code Agent SDK docs — https://code.claude.com/docs/en/agent-sdk/overview
- Claude Managed Agents — https://platform.claude.com/docs/en/managed-agents/overview (future cloud integration)
