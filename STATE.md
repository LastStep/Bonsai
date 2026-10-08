# Bonsai: where it stands

Rewritten, never appended. Last rewritten 8 Oct 2026 when part 0 landed (`records/runs/R-2026-10-08-formats.md`).

## In short

Bonsai is being rebuilt as one small Go program that gives every project the same formats, packs, guards, recorder and
proof ladder (`design/one-pager.md`). The design is settled and the plan is approved (8 Oct). **Part 0, the formats,
is done**: `formats/` holds a JSON Schema for each of the contract's ten formats, an example of each, 97 trick files
with both outcomes in `expect.json`, a raw-byte manifest (124 files, set 1) and a Go test, on `main` at `3770d04`,
passed by a fresh Opus verifier. Nothing else of the new Bonsai is built: `main` still holds the old product's code
(0.4.3 and the unreleased 0.5.0 commits), which part 1 clears out.

## Rohan's decisions, 8 Oct

- Bonsai's design and records live in this repo now (`design/`, `STATE.md`, `records/`); the studio keeps a pointer.
  Bonsai owns its spec and the formats contract.
- Work lands on `main` directly: no marker tag, no `rebuild` branch, no pull-request process. The old product stays
  reachable at tag `v0.4.3` and in git history.
- The JSON Schemas and the trick files are Bonsai's ("those are kind of tests which other projects can use"): the master
  lives in `formats/`, and they are Bonsai's first job.
- **Bonsai first** (Rohan, 8 Oct), to avoid interim code: the studio no longer adopts the formats early in its own Node
  code. Its readers task (T-0074) is cut. The studio adopts the formats when it links (spec step 7): `bonsai init` there,
  its bridge reading Bonsai's outputs, its own guards, ladder and statusline retired. Its Desk stays on upkeep until
  then. Nothing in the studio waits on part 0; Bonsai's own order is unchanged.
- "The proper way, no shortcuts": the spec's order inside Bonsai; parallel work only where it is truly independent.
- Bonsai joins the studio's dashboard as its own project at spec step 6, as planned.
- Haiku joins the models, for small bookkeeping and audit jobs (listing, sorting, counting, checking a list against a
  source); it reports facts and judges nothing (`CLAUDE.md`).
- Rohan approved `design/plan.md` and `design/one-pager.md` (8 Oct), after a fresh Opus review and its fixes.
- Label-definition fields stay required in the formats (Rohan, 8 Oct): Bonsai's packs write every field; the
  contract's short examples are out of date (the plan's "Stale or in tension" list).
- Rohan's hand checks come in two sittings, not one: checks b, c and d (about 25 minutes) after part 5, check a (about
  20 minutes) after part 4. Parts 3 and 5 close once the first sitting passes, part 4 once the second does.

## GitHub, read 8 Oct

- `main`: part 0 (`3770d04`) on top of the approved plan, pushed 8 Oct. The only branch; no open pull request; the
  newest tag is `v0.4.3`. The old maintenance routine's cloud sessions are archived.
- `release.yml` is disabled. No repository secret (the old `HOMEBREW_TAP_TOKEN` is deleted, and Rohan removed the token itself from GitHub, 8 Oct). The only environment is
  `github-pages`; the `release` environment and a new tap token wait for step 5.7.
- Ruleset `main-protection`: blocks force pushes and deletion only (Rohan switched off the pull-request and
  required-check rules on 8 Oct), so pushes to `main` are plain pushes.
- Workflows on: CI (`test`, `lint`, `govulncheck`), CodeQL, Dependabot, and Deploy Docs: a push to `main` touching
  `README.md`, `docs/`, `catalog/` or `website/` redeploys the old website to GitHub Pages until part 1 removes it.

## The one thing to do next: part 1, the clear-out and the new layout

Part 0 is done (`records/runs/R-2026-10-08-formats.md`). Next, in the plan's order: part 1, then the test pack (4a),
then parts 2, 3, 5, the rest of 4, and 6 (`design/plan.md`, "Parts 1-6"). Part 1 removes the old product's code from
`main` in one commit, sets up the new layout, rewrites CI (a `windows` job; govulncheck pinned to a version that builds
with CI's Go) and keeps `release.yml` disabled. Its builder is Opus; a fresh Opus verifier checks it **before** its
push (CI and release). Open `records/runs/R-<date>-clear-out.md` (the orchestrator is its only writer), make the
worktree `~/Servers/Bonsai-<part>`, and brief the builder with part 1's row, "CI and release after part 1", and the
rules from `CLAUDE.md`.

For part 2's reader, from part 0: the frontmatter is its lines with their own line endings (`formats/README.md`); cut
any other way, YAML libraries disagree on a final block scalar's newline.

## Waiting on Rohan

- Before part 5's hook checks: spec §17 step 3, the old `bonsai` binaries (the lines are in the spec).
- Later: the two hand-check sittings (after part 5, about 25 minutes; after part 4, about 20); step 8 at 5.4; at 5.7
  the `release` environment and a new tap token (the working environment command is in spec §17 step 4's note).

## Loose ends

- CI's `govulncheck` job fails at its install step: `go install golang.org/x/vuln/cmd/govulncheck@latest` now needs
  Go 1.26 and CI runs 1.25.9 with `GOTOOLCHAIN=local` (red since 23 Sep, before this setup). `test`, `lint`, `build` and
  CodeQL are green on the setup commit. Part 1 rewrites CI: pin govulncheck to a version that fits the Go it uses.
- Spec step 5.3 (the guard): the hook lines must not be redirectable by files an agent may edit. They call the installed
  `bonsai` by a fixed path, and what the guard trusts to find the main checkout cannot be rewritten without the guard
  noticing (fail closed). See `design/plan.md`, "What still links Bonsai and the studio", and spec §7.
- `README.md` and `CONTRIBUTING.md` still describe the old product, apart from a pointer; part 1 rewrites them.
- The spec's §18 and §19 D give both question D's figures (60 h, 30-46 h) and the format review's (61 h, 30-47 h);
  §14 and §20 say 61 and 30-47. The plan uses 61 and 30-47.
- The spec's §17 step 6 still has the hand checks in one sitting; the plan follows Rohan's two (8 Oct).
- Scratch left from part 0: `~/bonsai-checks/formats0/`, `verify0/`, `yaml-libs/` (PyYAML is the system's; npm
  `yaml@2.8.3` there), and under `%USERPROFILE%\bonsai-checks\` the clones `src` and `verify-src` and two bundles. Safe
  to delete; a later Windows run makes its own folder.
