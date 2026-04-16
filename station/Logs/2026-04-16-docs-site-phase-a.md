# 2026-04-16 — Documentation Site: Research + Plan + Phase A

## What Happened

- **Research phase:** Three parallel research agents investigated Starlight capabilities, Bonsai content inventory, and best-in-class docs site patterns
- **Key findings:** Starlight is the right framework; `starlight-llms-txt` plugin for AI-consumable docs; Diataxis information architecture; multi-resolution LLM topic bundles (Astro's pattern)
- **Plan 10 drafted:** 4 phases (A: scaffold & migrate, B: fill gaps, C: catalog auto-gen & LLM layer, D: deploy & CI), ~40 pages planned
- **Independent review:** Two subagent reviews (correctness + intent/completeness) found 13 issues — all resolved
- **Phase A executed:** Dispatched code agent via issue-to-implementation workflow with worktree isolation
- **PR #13 merged:** 53 files, 8,275 additions — Starlight project scaffolded, 19 pages migrated, 28 stubs created
- **GitHub Pages deploy:** Workflow added (`.github/workflows/docs.yml`), site deployed to laststep.github.io/Bonsai

## Key Decisions

- **`website/` directory** (not `docs/`) — avoids conflict with existing `docs/` guide files
- **`base: '/Bonsai'`** — required for GitHub Pages subdirectory hosting
- **`starlight-llms-txt` plugin** — auto-generates llms.txt for AI agent consumption
- **Catalog auto-generation script** planned for Phase C — single source of truth from meta.yaml
- **`<SYSTEM>` tags replaced with blockquotes** — per llms.txt spec compliance (caught in review)
- **`catalog.json` structured dump** added to Phase C — machine-readable alternative for structured AI queries

## Review Notes

- Plan 10 supersedes Backlog Group A items 1-3 (quickstart, concepts, CLI usage guides)
- `bonsai guide` multi-topic command remains separate backlog item
- No doc versioning for now — known limitation documented in plan

## Next

- Phase B: Fill content gaps (concepts, commands, catalog, reference, glossary — 29 pages)
