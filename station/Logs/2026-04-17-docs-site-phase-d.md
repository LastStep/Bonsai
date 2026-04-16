## 2026-04-17 — Plan 10 Phase D: Deploy & CI (Plan 10 Complete)

**Plan:** 10-docs-site.md, Phase D
**PR:** #19 (merged)
**Duration:** Single session

### What Was Done

1. **Verified Phases A-C** — Full audit of all 45 pages, build output, catalog generation, LLM bundles. All clean.

2. **Phase D implementation** (dispatched to code agent, worktree isolation):
   - **D1** — Updated `.github/workflows/docs.yml`: added `docs/**`, `README.md`, `HANDBOOK.md` to path triggers; added `npm run generate:catalog` step before build. Skipped `generate:llm-bundles` (plugin handles it automatically).
   - **D2** — Skipped (README already had docs badge and links from prior phases).
   - **D3** — Created `docs/README.md` redirect notice pointing to live docs site.
   - **D4** — Appended docs site URL footer to `docs/custom-files.md` (shown by `bonsai guide`).
   - **D5** — Added redirect note to top of `HANDBOOK.md`.

3. **Post-merge verification** — `make build` and `go test ./...` pass on main.

### Key Decisions

- **Deploy-on-push-to-main kept** (not release-tag-only). Rationale: pre-1.0 project, main is the latest, `starlight-versions` plugin is immature. Revisit at 1.0 when multi-version support matters.
- **`generate:catalog` added to CI** — despite Phase C decision to keep it as a dev tool, it's also run in CI as a safety net to ensure catalog pages are current.

### Plan 10 Summary

All 4 phases complete. PRs: #13 (Phase A), #15-17 (Phase B), #18 (Phase C), #19 (Phase D).
Docs site: https://laststep.github.io/Bonsai/ — 45 pages, full catalog, LLM bundles, auto-deploy.
