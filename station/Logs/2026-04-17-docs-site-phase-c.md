## 2026-04-17 — Plan 10 Phase C: Catalog Auto-Generation & LLM Layer

**Plan:** 10-docs-site.md, Phase C
**PR:** #18 (merged)
**Duration:** Single session

### What Was Done

1. **Catalog generation script** (`website/scripts/generate-catalog.mjs`)
   - Reads all 57 catalog items from `catalog/*/meta.yaml` and `agent.yaml`
   - Updates data tables between `{/* CATALOG-TABLE-START/END */}` markers in 7 catalog MDX pages
   - Outputs `website/public/catalog.json` for machine consumption
   - Idempotent — second run produces no changes

2. **LLM layer** — configured `starlight-llms-txt` plugin with:
   - Custom Bonsai description and install details in `llms.txt`
   - 4 topic bundles via `customSets`: concepts, commands, catalog, configuration

3. **Data drift fixes** — 3 inaccuracies corrected:
   - `design-guide`: wrong description and missing agents (tech-lead, backend)
   - `review-checklist`: missing agent (reviewer)
   - `test-strategy`: missing agent (qa)

### Key Decisions

- **Plan simplification:** Replaced planned custom scripts (C3 `generate-llm-bundles.mjs` + C4 post-processing) with native `starlight-llms-txt` plugin `customSets` config. Zero custom code, same result.
- **Marker approach for catalog pages:** Used `{/* CATALOG-TABLE-START/END */}` markers to auto-generate only data tables, preserving hand-written prose sections.
- **Option A for workflow:** `generate:catalog` is a dev tool, not a CI step. Run manually before committing when catalog changes. Catalog data tables are committed to git for PR visibility.
- **MDX comments:** Used `{/* */}` style markers instead of HTML `<!-- -->` since `.mdx` files don't support HTML comments reliably.

### Findings

- Site was already live at https://laststep.github.io/Bonsai/ from Phase B deployment
- `docs.yml` CI workflow already exists and auto-deploys — Phase D is mostly complete
- Remaining Phase D items: README docs link update (may be done), `bonsai guide` URL reference, HANDBOOK redirect note

### Backlog Items Added

- **Developer guide for Bonsai contributors** (P2, ungrouped) — covers build commands, `generate:catalog` workflow, release checklist
