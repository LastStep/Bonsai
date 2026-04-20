---
tags: [log, plan-18]
description: Plan 18 — bonsai guide multi-topic + orphan-doc cleanup — shipped 2026-04-20 via PR #25 with post-merge MDX hotfix.
---

# 2026-04-20 — Plan 18: `bonsai guide` multi-topic + legacy docs cleanup

## Shipped

PR #25 squash-merged as `e448140`. Post-merge hotfix `e336ccb` for MDX build.

**Scope delivered:**
- `bonsai guide` becomes multi-topic: no-args opens a Huh picker; `bonsai guide <topic>` renders directly; unknown topic errors with the allowlist; `cobra.MaximumNArgs(1)` rejects extra args.
- Three new terminal cheatsheets (all ≤120L, glamour-clean, footer deep-link to Starlight): `docs/quickstart.md` (93L), `docs/concepts.md` (113L), `docs/cli.md` (119L). `docs/custom-files.md` kept unchanged.
- Root `embed.go` now exposes four explicit embed vars (`GuideQuickstart`, `GuideConcepts`, `GuideCli`, `GuideCustomFiles`) — `GuideContent` renamed/retired.
- `cmd.Execute(fsys fs.FS, guides map[string]string)` signature threaded through `main.go` → `cmd/root.go` → `cmd/guide.go`.
- Three orphan legacy docs deleted: `HANDBOOK.md` (515L), `docs/working-with-agents.md` (403L), `docs/triggers.md` (295L) — 1,213 lines total. All superseded by the Starlight site (Plan 10).
- CLAUDE.md doc-drift fix rolled into the same PR (tree line 19 `main.go` → `cmd/bonsai/main.go` + `embed.go`; line 109 embed reference updated). P1 doc-drift backlog entry retired.
- Ancillary: `README.md` commands table, `website/src/content/docs/commands/guide.mdx`, `station/code-index.md`, `.github/workflows/docs.yml` path filter (HANDBOOK.md removed since file is deleted).

## Decision record

- **Terminal cheatsheets, not full mirrors.** Original backlog items (quickstart/concepts/CLI-usage) pre-dated Starlight (Plan 10); writing full-length versions would duplicate the site. Chose distilled 80–120-line cheatsheets deep-linking to the website.
- **Both invocation modes.** Huh picker *and* `bonsai guide <topic>` direct-arg — covers browsing and muscle-memory use.
- **Delete orphans in same PR.** Considered keeping them with redirect notes, but zero live code refs + Starlight canonical made deletion cleaner than maintaining redirect stubs.
- **Kept `docs/custom-files.md` unchanged** (exceeds 120L cap) — it pre-existed as the sole embedded guide and was intentionally out of scope per Phase B step 9. Phase G's line-cap check applied only to the three new files.

## Post-merge incident

- **Deploy Docs workflow failed on main** after PR #25 merged. Root cause: `<https://laststep.github.io/Bonsai/>` autolink in `guide.mdx:20`. MDX parses `<` as JSX, so Astro's Vite build failed with `Unexpected character '/'`.
- **Why the PR check didn't catch it:** the `CI` workflow (test + lint + GitGuardian) runs on `pull_request`, but `Deploy Docs` runs only on `push` to main. So `website/**` changes aren't validated pre-merge.
- **Fix:** `e336ccb` replaced the autolink with `[laststep.github.io/Bonsai](https://...)`. `npm run build` clean locally (45 pages). Deploy Docs green on main.
- **Prevention:** backlog entry added to run Astro build as a PR check on `website/**` paths.

## Verification trail

- CI on PR #25: test + lint + GitGuardian — all green.
- Post-merge on main: `make build`, `go test ./...`, `./bonsai guide quickstart|concepts|cli|custom-files` renders cleanly, `./bonsai guide unknown` exits 1 with allowlist error, `./bonsai guide a b` exits 1 with cobra's "accepts at most 1 arg".
- `git grep` confirms zero live code/doc references to `HANDBOOK`, `working-with-agents.md`, `docs/triggers.md`, or `GuideContent`.
- Deploy Docs on `e336ccb`: success.

## Follow-ups captured

- Backlog Group A retired (quickstart/concepts/cli-usage + multi-topic command — all shipped).
- New Group C entry: run Astro build on PRs touching `website/`.
- Memory notes added: PR CI ≠ main CI; MDX autolink gotcha.

## Learning — for next time

1. **When a PR touches `website/`, run `npm --prefix website run build` locally before merging.** The PR's CI workflow does not cover Astro.
2. **Inside `.mdx`, never use `<url>` autolinks.** Use `[label](url)`. MDX ≠ GFM.
