# Bonsai: where it stands

Rewritten, never appended. Last rewritten 8 Oct 2026 when Rohan approved the plan (`records/runs/R-2026-10-08-plan-review.md`).

## In short

Bonsai is being rebuilt as one small Go program that gives every project the same formats, packs, guards, recorder and
proof ladder (`design/one-pager.md`). The design is settled: the spec, the formats contract, and Rohan's own review of
every format (confirmed 7 Oct: the formats are final), all in `design/`. **Nothing of the new Bonsai is built yet.**
`main` still holds the old product's code (0.4.3 and the unreleased 0.5.0 commits), which the skeleton's part 1 clears
out. The old agent workspace (`station/`, the old `CLAUDE.md`, `.claude/`, `.bonsai.yaml`, `.bonsai/`) was removed on
8 Oct, so agents start clean from `CLAUDE.md`, this file and `design/plan.md`.

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
- Rohan's hand checks come in two sittings, not one: checks b, c and d (about 25 minutes) after part 5, check a (about
  20 minutes) after part 4. Parts 3 and 5 close once the first sitting passes, part 4 once the second does.

## GitHub, read 8 Oct

- `main`: the plan's fixes (`8c6eb6f`) on top of the setup (`3fa4982`), pushed 8 Oct. The only branch; no open pull
  request; the newest tag is `v0.4.3`. The old maintenance routine's cloud sessions are archived.
- `release.yml` is disabled. No repository secret (the old `HOMEBREW_TAP_TOKEN` is deleted, and Rohan removed the token itself from GitHub, 8 Oct). The only environment is
  `github-pages`; the `release` environment and a new tap token wait for step 5.7.
- Ruleset `main-protection`: blocks force pushes and deletion only (Rohan switched off the pull-request and
  required-check rules on 8 Oct), so pushes to `main` are plain pushes.
- Workflows on: CI (`test`, `lint`, `govulncheck`), CodeQL, Dependabot, and Deploy Docs: a push to `main` touching
  `README.md`, `docs/`, `catalog/` or `website/` redeploys the old website to GitHub Pages until part 1 removes it.

## The one thing to do next: part 0, the formats

`design/plan.md`, approved 8 Oct, part 0: a JSON Schema for each of the contract's ten formats, the trick files with their expected outcomes in
`formats/expect.json`, a raw-byte manifest, and a Go test. 6-10 AI hours; at 13 work stops and Rohan is asked.
1. Open `records/runs/R-<date>-formats.md` (the orchestrator is its only writer). Make the builder's worktree:
   `git -C ~/Servers/Bonsai worktree add ~/Servers/Bonsai-formats -b formats main`.
2. Brief one Opus builder: part 0 whole, the sources it lists (contract §2.4 and each format's section, spec §16's rows
   and the sections they cite, by grep), and the rules from `CLAUDE.md`. The format-0 outcomes come from the studio's
   frozen reader, taken read-only: `git -C ~/Servers/Trinetra-Game-Studio show 4a05eac:tools/lib/yaml.mjs >
   <scratch folder>/yaml0.mjs`. Nothing is written in the studio's checkout. Tell the builder: the private-pattern test
   flags `/home/`, so the `status --json` example needs a made-up absolute path the pattern does not match (contract
   §12's own example uses `/home/<user>/...`).
3. A fresh Opus verifier reads every case rule by rule and runs the frozen reader with a runner of its own.
4. Before the push: `go test ./formats/` natively on Windows in a Windows-git clone made from a `git bundle` of the
   branch under `%USERPROFILE%\bonsai-checks\`. Then fast-forward `main`, push, check CI (`test` and `lint` green;
   `govulncheck` stays red until part 1), name the commit here, and tell Rohan.

Then the walking skeleton, parts 1-6, in the plan's order (1, the test pack, 2, 3, 5, the rest of 4, 6). Part 1 gets
its own fresh verifier before its push.

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
