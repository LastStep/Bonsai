# Bonsai: where it stands

Rewritten, never appended. Last rewritten 9 Oct 2026 at 01:15, when part 6 closed and the walking skeleton reached
Rohan's gate (`records/runs/R-2026-10-09-gate.md`).

## In short

Bonsai is being rebuilt as one small Go program that gives every project the same formats, packs, guards, recorder and
proof ladder (`design/one-pager.md`). The plan is approved (8 Oct). **Part 0 and all six parts of the walking skeleton
are done; the skeleton's last fresh verifier passed it. Rohan's gate is next: he picks path (a), (b) or (d).** The
gate report is `records/gate-skeleton.md`; its section 1 is for him. Everything below is on `main`, pushed.

- **Part 0, the formats** (`3770d04`, now set 3 at `3a1f195`): `formats/` holds a JSON Schema for each of the
  contract's ten formats, an example of each, 116 trick files with their format-0 and format-1 outcomes in
  `expect.json`, a raw-byte manifest (138 files) and a Go test.
- **Part 1, the clear-out** (`17f2938`): the old product's code gone (358 files; 0.4.3 stays at its tag); CI has
  `test` (Linux), `windows`, `lint`, `govulncheck` (pinned to `v1.7.0`) and CodeQL; `release.yml` can only build.
- **Part 4a, the test pack**: `LastStep/bonsai-test-pack`, public, commits A `5062053`, B `ab09e4b`, C `abfb5de`, D
  `1d4f46f`; `main` at D.
- **Part 2, the reader** (`3a1f195`): the format-1 reader without a YAML library, reaching all 116 format-1 outcomes;
  `bonsai.yaml`, `pack.yaml`, the lock; `bonsai status [--json]`.
- **Part 3, the engine** (`ee50971`): `bonsai init`, `update`, `check`; staged writes, the lock last; a hook-line
  change refused (exit 4).
- **Part 5, the hook path** (`33a6122`): `bonsai hook guard`, one rule, failing closed under every fault.
- **Part 4b, packs as plugins** (`54b4fdd`): each locked pack installed at project scope at its commit; two projects
  and a worktree each load their own commit, on WSL and Windows.
- **Part 6, the gate report** (`5633721`): spec §15's list measured on both sides, the twelve checks, the four stop
  lines, the findings step 5 inherits (its section 5, by step 5 part), a three-case `claude plugin eval` of the test
  pack. The last verifier re-ran the tests on both sides, part 3's scripted checks on the final build, check 3's
  Windows half and check 11 on WSL, and passed every part, check and stop line; its findings are fixed (one spec line
  named a user folder: fixed forward in `5633721`).
- Tests at `5633721` (the Go code is `54b4fdd`'s), run by the last verifier: WSL 773 runs (2 skipped, Windows-only
  tests) and 779 with the fault tag; natively on Windows 777 (1 skipped, symbolic links) and 783; `go vet` clean.

Hours (AI, builder and verifier runs, from the run reports): part 0 45 minutes against 6-10 h; the skeleton, parts
1-6, 6 h 18 min against 30-47 h (stop line 61 h; the gate report says 6 h 17 min, its hours filled a minute before
its last round ended). Windows-only failures about 18 minutes against 8 h. Option rounds asked of Rohan inside the
skeleton: none. No change to Mimas or the studio's repo. No stop line crossed.

## Rohan's decisions, 8 Oct

- Bonsai's design and records live in this repo (`design/`, `STATE.md`, `records/`); Bonsai owns its spec and the
  formats contract. Work lands on `main` directly; the old product stays at tag `v0.4.3`.
- The JSON Schemas and the trick files are Bonsai's: the master lives in `formats/`.
- **Bonsai first**: the studio adopts the formats when it links (spec step 7), not before.
- "The proper way, no shortcuts": the spec's order inside Bonsai; parallel work only where truly independent.
- Bonsai joins the studio's dashboard as its own project at spec step 6.
- Haiku for small bookkeeping and audit jobs; it reports facts and judges nothing (`CLAUDE.md`).
- The plan and one-pager approved after a fresh Opus review and its fixes.
- Label-definition fields stay required in the formats; the contract's short examples are out of date.
- Hand checks in two sittings; and (his later word) a Sonnet agent runs hand checks where an agent can.
- His roadmap is the artifact "Trinetra Roadmap" (https://claude.ai/artifact/XXKTi6geneu1pdQ4h4miw2); Bonsai's cards
  are updated after each part.

## GitHub, read 9 Oct

- `main` is the only branch; the newest tag is `v0.4.3`; no open pull request.
- `release.yml` is disabled (`disabled_manually`) and can only build; no repository secret. The `release` environment
  and a new tap token wait for step 5.7.
- Ruleset `main-protection`: blocks force pushes and deletion only.
- `LastStep/bonsai-test-pack`: public, `main` at D, no tag.
- The old website stays on GitHub Pages until Rohan turns Pages off.

## The one thing to do next: Rohan's gate

Rohan reads `records/gate-skeleton.md` (section 1) and picks the path for step 5 (spec §14): **(a)** the full Bonsai
1.0 (139-218 h in the spec; 18.6-45.6 h at the skeleton's measured ratio, 0.1337-0.2094), **(b)** the smaller cut
(about 121-188 h; 16.2-39.4 h), or **(d)** 1.0 with its own screens (about 169-268 h; 22.6-56.1 h). The ratio counts
agent runs only; the report's "What limits the ratio" says why it may not hold for step 5.

After his pick: a plan for step 5.1 (formats and engine to 1.0, spec §14's table), reviewed by a fresh Opus agent
before it reaches him, built in a plain worktree as the skeleton was. Step 5.1 settles the 5.1 findings in the gate
report's section 5 (`--allow-exec` and the mixed update first).

## Waiting on Rohan

- The gate: (a), (b) or (d).
- Later: step 8 at 5.4 (a pre-release `bonsai`); at 5.7 the `release` environment and a new tap token (the working
  environment command is in spec §17 step 4's note).
- Whenever he likes: turn GitHub Pages off (the old website).

## Loose ends

- Every finding and loose end the skeleton leaves is in `records/gate-skeleton.md` section 5, grouped by the step 5
  part that must settle it (5.1 to 5.7, and the spec's own text). This file no longer repeats them.
- Outside step 5's parts: the spec's §17 step 6 still names `/agents` and one sitting.
- Notes from the last verifier, for the record: part 0's run report quotes a studio task id from the plan; the reader's
  comments name `yaml.mjs`, the contract's format-0 reference (the plan's exception lists only part 0's report and
  `formats/README.md`).
- Scratch under `~/bonsai-checks/` and `%USERPROFILE%\bonsai-checks\`: clones, bundles, builds, targets, the
  launchers (`claude-here`, `bonsai-here`), the eval suite (`eval/`) and the last verifier's runs (`verify-final/`).
  `verify-final/scripts/out/check11-all.log` holds a full PATH with the user name: scratch only, never to be pasted
  into the repo. Safe to delete once step 5 has its own checks, but for the launchers if it reuses them.
