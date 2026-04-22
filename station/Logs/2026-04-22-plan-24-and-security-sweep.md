---
tags: [log, plan-24, security, oss-launch]
date: 2026-04-22
---

# 2026-04-22 — Plan 24 pre-launch polish + security sweep

## What shipped

### Plan 24 — Pre-Launch Polish Bundle (PR [#58](https://github.com/LastStep/Bonsai/pull/58) squash `4ef8271`)

Four OSS-readiness items in one bundle:

- **`CHANGELOG.md`** — new file at repo root, keep-a-changelog 1.1.0 format, curated v0.1.0–v0.1.3 backfill, `[Unreleased]` stub, 5 link-reference rows.
- **`.github/workflows/docs.yml`** — `pull_request` trigger added (same `paths:` as `push`), deploy job + upload-pages-artifact step gated on `if: github.event_name == 'push'`. Broken MDX now fails at PR time (reacts to PR #25 post-merge incident) while deploy stays push-only.
- **Root `Bonsai/CLAUDE.md`** — `internal/tui/` tree block refreshed to include `styles_test.go`, `filetree.go`, `filetree_test.go`, `harness/` (Plan 15), `initflow/` (Plan 22). Subdirs listed only; 22-file `initflow/` not per-file-enumerated.
- **Backlog.md consolidation** — Group C "CHANGELOG.md + richer release notes" + "Consolidate CHANGELOG backlog items" replaced with HTML-comment resolution markers. Group D "Changelog generation skill" retained as future work, suffixed `(refiled as good-first-issue via Plan 24 Step E)`. Root CLAUDE.md drift entry in Group E also marked resolved.

Plus **Step E** (outside the PR): 5 GitHub issues filed / relabeled for contributor on-ramp, all `good first issue` + `help wanted`:

| # | Title |
|---|-------|
| [#53](https://github.com/LastStep/Bonsai/issues/53) | statusLine catalog sensor port (pre-existed, relabeled) |
| [#54](https://github.com/LastStep/Bonsai/issues/54) | Shell completion via `bonsai completion [bash\|zsh\|fish]` |
| [#55](https://github.com/LastStep/Bonsai/issues/55) | `bonsai changelog` command + catalog skill |
| [#56](https://github.com/LastStep/Bonsai/issues/56) | Umbrella — propose and add a new catalog skill/workflow/protocol |
| [#57](https://github.com/LastStep/Bonsai/issues/57) | Post-update `.bak` merge hint in `bonsai update` |

### Post-Plan-24 security sweep (commits `ef55f4f` + hotfix `7c1dd49`)

User asked: "check all the security warnings we have on github, and fix them if needed."

Findings:

- **Dependabot:** 1 alert, already auto-fixed — Astro XSS via `define:vars` (CVE-2026-41067, GHSA-j687-52p2-xcff, severity medium). `website/package-lock.json` resolved `astro@6.1.7` ≥ patched `6.1.6`.
- **CodeQL:** 2 low-severity `go/useless-assignment-to-local` alerts, both fixed in `ef55f4f`:
    - `internal/tui/styles.go:211` — `val := value` → `var val string` (both branches overwrite, initial assignment dead).
    - `cmd/add.go:690` — dropped trailing `offset++` in final `routines` block (offset never read after).
- **Secret scanning:** 0 alerts, clean.
- **Branch protection:** ruleset `main-protection` active (required `test` check + required PR). Admin bypass preserved — fast-iter UX convention stays for polish commits, external contributors still hit the gate.

Also: widened `.github/workflows/ci.yml` to trigger on `push: branches: [main]` in addition to `pull_request`. Closes the "gofmt drift on main silently accumulates" pattern documented in memory — direct-to-main polish batches now surface CI failures immediately.

## Incidents

### Plan 24 pre-merge lint fail — pre-existing gofmt drift on main

PR #58's first lint run failed on `internal/tui/initflow/observe.go:466` + `planted.go:423` — unformatted since the 2026-04-22 dogfood polish run (commits `018966d` / `4a8fea9` / `975e15d`, all direct-to-main). CI never ran those pushes because `ci.yml` was PR-only. Fix: dispatched an agent to `gofmt -s -w` the two files on the PR branch. Merged clean. Motivated the push-CI widening later in the session.

### Security-sweep hotfix — `git add <file>` pulled in parallel-session WIP

`ef55f4f` intended to: drop one `offset++` in `cmd/add.go:690`, fix `val` in `styles.go:211`, add push trigger to `ci.yml`. But `git add cmd/add.go` also staged two Plan 23 WIP hunks from the working tree:

- `os` import addition
- `BONSAI_ADD_REDESIGN` env gate at `runAdd` top calling `runAddRedesign`

`runAddRedesign` lives in untracked `cmd/add_redesign.go` (Plan 23 WIP). Push-CI fired for the first time thanks to the widening we'd just shipped — caught `undefined: runAddRedesign` within 5 minutes on test/lint/govulncheck. Hotfix `7c1dd49` reverted only those two hunks; WIP restored to working tree uncommitted for Plan 23 owner. The widening paid for itself inside its own session.

## Final state

- **Main at:** `7c1dd49` (after `ef55f4f` security, `4974a60` Plan 24 bookkeeping, `4ef8271` PR #58 merge). Local + remote in sync. `make build` + `go test ./...` green.
- **Working tree:** Plan 23 WIP uncommitted (`cmd/add.go` gate + import, `internal/tui/initflow/{chrome,design,enso,generate,stage}.go`, `internal/tui/addflow/` untracked, `cmd/add_redesign.go` untracked, Backlog/StatusArchive edits). Parallel session's track; not mine to commit.
- **5 `good first issue` issues open** for contributor on-ramp (#53–#57).
- **CHANGELOG.md visible on repo homepage.**

## Remaining pre-announce gates

1. **Demo GIF / asciinema** for README hero — user recording (not agent-able).
2. **Pre-release docs audit** across README / SECURITY / CONTRIBUTING / CODE_OF_CONDUCT / Starlight pages / `bonsai guide` cheatsheets — user-flagged as final gate before announce.

## Durable takeaways (also captured in memory.md Notes)

1. **CI on `push: branches: [main]`** is structural, not optional — direct-to-main polish commits need the same safety net as PRs.
2. **`git add <file>` is all-or-nothing.** When parallel-session WIP exists in a file, always `git diff --staged` or `git add -p` before commit. Rule written into memory.
3. **Push-CI pays for itself fast.** The widening caught my own follow-on mistake within 5 minutes of shipping.
