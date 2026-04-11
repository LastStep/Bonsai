---
tags: [core, memory]
description: Bonsai developer working memory — flags, work state, notes.
---

# Working Memory

## Flags

(none)

## Work State

**Current task:** Catalog expansion (RESEARCH-catalog-expansion.md)
**Blocked on:** (nothing)

**Progress:**
- Phase 1 (skills): DONE, committed — api-design-standards, auth-patterns, test-strategy, review-checklist
- Phase 2 (agents): DONE, committed — fullstack, devops, security agents + all dependencies
- Phase 3 (cross-cutting): DONE — cli-conventions, mobile-patterns skills + api-development, test-plan, pr-review workflows + test-integrity-guard sensor

**Decision:** Defer reviewer, qa, docs agents — they overlap with tech-lead. Final agent lineup: tech-lead, fullstack, backend, frontend, devops, security.

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

- RESEARCH-catalog-expansion.md — full spec for all new agents, skills, sensors, workflows, routines
