# Bonsai: where it stands

Rewritten, never appended. Last rewritten 8 Oct 2026 at the end of the day, when part 4 closed
(`records/runs/R-2026-10-08-plugins.md`).

## In short

Bonsai is being rebuilt as one small Go program that gives every project the same formats, packs, guards, recorder and
proof ladder (`design/one-pager.md`). The plan is approved (8 Oct). **Part 0 and parts 1 to 5 of the walking skeleton
are done and closed; part 6, the gate report, is next.** Everything below is on `main`, pushed, CI green.

- **Part 0, the formats** (`3770d04`, now set 3 at `3a1f195`): `formats/` holds a JSON Schema for each of the
  contract's ten formats, an example of each, 116 trick files with their format-0 and format-1 outcomes in
  `expect.json`, a raw-byte manifest (138 files) and a Go test.
- **Part 1, the clear-out** (`17f2938`): the old product's code gone (358 files; 0.4.3 stays at its tag); CI has
  `test` (Linux), `windows`, `lint`, `govulncheck` (pinned to `v1.7.0`) and CodeQL; `release.yml` can only build.
- **Part 4a, the test pack**: `LastStep/bonsai-test-pack`, public, commits A `5062053`, B `ab09e4b`, C `abfb5de`, D
  `1d4f46f`.
- **Part 2, the reader** (`3a1f195`): the format-1 reader without a YAML library, reaching all 116 format-1 outcomes;
  `bonsai.yaml`, `pack.yaml`, the lock; `bonsai status [--json]`.
- **Part 3, the engine** (`ee50971`): `bonsai init`, `update`, `check`; staged writes, the lock last; a hook-line
  change refused (exit 4); checks 1-6 and 12 pass on scratch targets.
- **Part 5, the hook path** (`33a6122`): `bonsai hook guard`, one rule, failing closed under every fault (missing,
  crash, slow, minimal PATH, and a 145 KB Write under `missing`); on Windows the final path Windows reaches is judged.
- **Part 4b, packs as plugins** (`54b4fdd`): each locked pack installed with `claude plugin install --scope project`
  at its commit; `project-a`, `project-b` and a worktree beside `project-a` each load their own commit, on WSL and
  Windows; `check` reports drift.
- **Rohan's sittings**: the first (hand checks b, c, d) passed by Rohan; the second (hand check a) run by a Sonnet
  agent on his word, A, B, B, A on both sides.
- Tests at `54b4fdd`: WSL 773 runs plain and 779 with the fault tag; natively on Windows 777 and 783; `go.mod` has no
  dependencies.

Hours (AI, from the run reports): part 0 45 minutes against 6-10 h; parts 1-5 5 h 36 min against the skeleton's
30-47 h (stop line 61 h); Windows-only failures about 18 minutes against an 8 h stop line; option rounds asked of
Rohan inside the skeleton: none.

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
  are updated after each part (version 27 at the end of 8 Oct).

## GitHub, read 8 Oct

- `main` is the only branch; the newest tag is `v0.4.3`; no open pull request (Dependabot closed #252 itself, and
  alert #45 is fixed).
- `release.yml` is disabled and can only build; no repository secret; the old tap token is removed. The `release`
  environment and a new tap token wait for step 5.7.
- Ruleset `main-protection`: blocks force pushes and deletion only.
- `LastStep/bonsai-test-pack`: public, `main` at D, no tag.
- The old website stays on GitHub Pages until Rohan turns Pages off.

## The one thing to do next: part 6, measure and report

Part 6 (4-6 AI hours): the gate report against spec §15's list: hours per part against estimate and the ratio for step
5; each of the twelve checks; Windows-only failures with hours; hook start-up p50 and p95 on both sides (part 5: WSL
2.4 ms against 1.5, Git Bash 63 ms against 65); the fail-closed results; plugin install and update time (part 4b:
about 11-13 s first, under 1 s again); two commits kept; the Claude Code version on each side (2.1.294); binary size;
lines of Go per part; a three-case `claude plugin eval` of the test pack; whether the studio's half of check 7 has run
(it has not: step 7). Then the skeleton's last fresh Opus verifier over all twelve checks, then Rohan's gate.

Findings the gate report must carry:
- First-time setup: Claude Code registers a project's marketplace only in a session in a trusted folder, so `install`
  reports `waiting` until a person has opened Claude Code there once.
- `/agents` is gone in Claude Code 2.1.294; hand check a read the marker through `--agent test-pack:marker` instead,
  run by a Sonnet agent; the marker's answer inside an interactive session was not read.
- Local plugin scope leaks across worktrees; Bonsai uses project scope only.
- Claude Code's own writes outside the scratch folders: session transcripts, folder-trust entries, and Claude.ai's
  plugin sync rewriting `~/.claude/plugins/synced/<account>/.marketplaces.json`. The user `settings.json` files never
  changed (WSL `7b515457...a025a7`, Windows `2b6295c1...4ff6c9`).

## Waiting on Rohan

- Nothing now. Next: the gate, after part 6's report.
- Later: step 8 at 5.4 (a pre-release `bonsai`); at 5.7 the `release` environment and a new tap token (the working
  environment command is in spec §17 step 4's note).
- Whenever he likes: turn GitHub Pages off (the old website).

## Loose ends

- Go 1.25.9 (the `toolchain` line) has standard-library vulnerabilities that govulncheck lists but our code does not
  reach; a later 1.25.x patch fixes them. A one-line `toolchain` bump, as its own small commit.
- For step 5.3, from part 5's verifier: a `.git` entry in the session's starting subfolder makes the guard read the
  project as unlinked; `bonsai.yaml` (what is protected) is read from the working tree; `bonsai` is found by the PATH,
  which a settings `env` could redirect; shell `rm` or `mv` of `bonsai.yaml` or the lock is not judged; a dangling
  junction into a protected folder is allowed (the write fails); a junction swapped between check and write is not
  caught; the admin share `\\localhost\C$` is a documented limit. And Rohan's 8 Oct requirement: the hook lines must
  not be redirectable by files an agent may edit (`design/plan.md`, "What still links Bonsai and the studio", item 5;
  spec §7). Spec §7's note asks for a fixed path, §3 and check 2 for a name: step 5.3 settles it.
- For step 5.1: a plugin's move to a new commit is not counted as running code (spec §5 lets plugins carry hooks); with
  the lock deleted, `init --yes` links again and writes a changed hook line (its preview says a link is consent);
  `status --json`'s `formats` stays null (a read-only format cannot be listed); no `error` object; `--allow-exec`.
- For step 5.2: the log's `bonsai_path` and `bonsai_sha256` field names; `input_hash` null until the salt.
- For a later formats set: a quoted `"format":` key followed by a tab still dispatches as format 0; a file with lone-CR
  line endings gets `quote-this-value`, whose next step does not help.
- CI's `lint` does not use the fault tag (`go vet` and tests do).
- Small wording: `README.md` describes the designed product in the present tense; `CONTRIBUTING.md`'s toolchain note
  holds from Go 1.21; `.gitattributes` keeps old-product patterns (harmless); `CHANGELOG.md`'s rebuild section still
  describes part 1's stub.
- The spec's §18 and §19 D give both 60 h / 30-46 h and 61 h / 30-47 h; the plan uses 61 and 30-47. The spec's §17
  step 6 still has one sitting and `/agents`.
- Scratch under `~/bonsai-checks/` and `%USERPROFILE%\bonsai-checks\`: clones, bundles, builds, the YAML libraries,
  the scratch targets and the launchers. Keep the targets and launchers for part 6; the rest is safe to delete.
