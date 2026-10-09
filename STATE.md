# Bonsai: where it stands

Rewritten, never appended. Last rewritten 9 Oct 2026 at 15:18, when 5.1.0 to 5.1.4b had landed and 5.1.5 was next
(`records/runs/R-2026-10-09-5.1.4b-outputs.md`).

## In short

Bonsai is being rebuilt as one small Go program that gives every project the same formats, packs, guards, recorder and
proof ladder (`design/one-pager.md`). The plan is approved (8 Oct). **Part 0 and all six parts of the walking skeleton
are done; the skeleton's last fresh verifier passed it. At the gate (9 Oct) Rohan chose path (a), the full Bonsai
1.0. He approved the plan for step 5, `design/plan-5.md`, on 9 Oct; step 5.1 (formats and engine to 1.0, 30-47 h,
re-ask at 61) is under way.** The gate report is `records/gate-skeleton.md`. Everything below is on `main`, pushed.

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

## Rohan's decisions

- **9 Oct, the gate: path (a), the full Bonsai 1.0** (spec §14, "Path (a) after the gate", 139-218 h, parts 5.1-5.7,
  each with a re-ask line at 1.3 times its high estimate).
- **9 Oct: Bonsai has no screens of its own.** "Completely remove the idea of bonsai's own screens ... this visual part
  of the job will be handled by the studio, while bonsai is a pure cli tool for overall simplicity." No web page, no
  `bonsai serve`; every visual of Bonsai's data is the studio's, from Bonsai's JSON outputs. Dated notes in the spec
  (§1, §2, §4, §11, §14, §16), the contract's reader lists, the one-pager and `CLAUDE.md`.
- **9 Oct: Bonsai's scope stands as split on 7 Oct.** Bonsai is everything inside one project (its rules, its guard,
  the record of what happened, the proof that work is done) and works with no studio; the studio is everything across
  projects and anything that acts on agents (dispatch, approvals, status moves, notifications, every visual). The
  recorder and the ladder runner stay in Bonsai; Bonsai has no scheduler.
- **9 Oct, 5.3's two questions:** the hook lines name the installed `bonsai` (Windows' place, then WSL's
  `/usr/local/bin/bonsai`), never the PATH, with check 2's one named exception ((a)); and an agent may pass
  `--allow-exec`, run `unlink` or change Bonsai's guard lines only under his tap's grant in a project the studio
  manages, only a person elsewhere ((ii)), so from 5.4 such an update of Bonsai's own repo is his command. Spec §3 and
  §18's dated notes; `design/plan-5.md`, "Step 5.3".
- **9 Oct: a pack plugin that runs code is asked for on each machine** (5.1's first option round): `update` installs a
  locked pack plugin with code parts only with `--allow-exec` on that machine; and Bonsai installs or removes no
  plugin but the project's own packs' ("it shouldnt install or remove other plugins"). Spec §5's dated note.
- **9 Oct: later parts' plans reach him under (B):** a part's section comes to him only when it changes what is his
  (hours or re-ask line, the order, his steps, a new public repo or content, an option round, a format change: a new
  major or a removal). He first said (A), then chose (B) once the scope question was settled.
- 9 Oct: a handoff for the studio's orchestrator, `records/handoff-studio-2026-10-09.md`; the studio still links at
  step 7, after 5.5 (his 8 Oct "Bonsai first" stands).

8 Oct:

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

## The one thing to do next: step 5.1.5, what the engine reads and writes

`design/plan-5.md`, "Step 5.1". Landed on 9 Oct, each with CI green (run reports `records/runs/R-2026-10-09-5.1.*`):
- **5.1.0** (`6a3a419`): `TestEachFaultBlocks` no longer races a 200 ms budget (test file only).
- **5.1.1** (`6c6fc33`): consent to code. `--allow-exec` real for `init` and `update`; every first link to the test pack
  needs it; test-pack commits E and F. Its fresh verifier failed the first round (a folder change, a pack named
  `bonsai`, pack files under `.claude/`) and passed the fix round. Rohan's answer (11:50): a pack plugin that runs code
  is installed on a machine only with `--allow-exec` there, and Bonsai installs or removes no plugin but the project's
  own packs'.
- **5.1.2** (`c4816eb`): the format-0 reader, a hand port of `yaml.mjs` at `4a05eac`; its fresh verifier passed it
  (about 2.6 million inputs and today's studio and Mimas files, 0 differences but Node's own stack limit past about
  1,700 levels of nesting, which the port reads: "no new refusals").
- **5.1.3** (`254572f`; set 4 is `22dd08a`): formats set 4, eight new schemas (`workspace`, `pack`, `tasks`,
  `sessions`, `memory`, `error`, `check`, `changes`), `status`'s `error` and the lock's `path` added, 125 trick cases,
  contract §13's active-task fixtures, the schema-compare test (base set 3's `4936b37`), CI's history for it and
  lint's online check off. Its fresh verifier passed it.
- **5.1.4a** (`37125a4`): a Go type for each of the eighteen formats (`internal/format`, one registry), the engine and
  `check` reading `bonsai.yaml` and `pack.yaml` in full while the guard's read stays lean (its code unchanged), `check
  --schema`.
- **5.1.4b** (`c2864c0`): every command's `--json` in set 4's shapes, 28 error words in one table each with who takes
  the next step, one flag-and-exit table per command word. Rohan's look at the error words and the two output shapes
  sent.

Next, **5.1.5** (the lock's `declares`, `format0` and `path`, the moved tag, document kinds, labels in force, the active
task, this machine's settings, the instruction block, `bonsai.yaml` written with every field), then 5.1.6 to 5.1.10 in
the plan's order.

Later parts' sections, planned beside 5.1 on Rohan's 12:27 word ("if you can orchestrate work in parallel do that
whenever possible"), each reviewed fresh, fixed and audited:
- **5.2** (on `main` at `da427de`; `records/runs/R-2026-10-09-plan-5.2.md`): nothing of Rohan's changes but one look at
  two log field names. 5.2 starts when 5.1 ends.
- **5.3** (on `main` at `b79d00f`; `records/runs/R-2026-10-09-plan-5.3.md`), with Rohan's two answers.
- **5.4** being planned now (`records/runs/R-2026-10-09-plan-5.4.md`); it comes to Rohan (his pre-release install).

5.1's hours so far: 488 minutes (8 h 8 min) of 61 h. Step 5's Windows-only tally: 0. Option rounds in 5.1: one.

## Waiting on Rohan

- Whenever convenient: Claude Code's login on Windows has expired (5.1.1's Windows sessions could not reach the
  model). In PowerShell: `claude`, then `/login`, then quit.
- In 5.1, a look (no vote): the error words and their two-part "what next", and the shapes of `check`'s and
  `update`'s JSON (5.1.3/5.1.4b). In 5.2, a look (no vote): the log's two new field names, `bonsai_path` and
  `bonsai_sha256`.
- Later: step 8 at 5.4 (a pre-release `bonsai`); at 5.7 the `release` environment and a new tap token (the working
  environment command is in spec §17 step 4's note).
- Whenever he likes: turn GitHub Pages off (the old website).

## Loose ends

- The WSL user settings file's baseline is `9e049dea...80d8d6` since Rohan's `/plugin` at 12:06 on 9 Oct (his word;
  before it `7b515457...a025a7`). Windows' is unchanged, `2b6295c1...4ff6c9`.
- From 5.1.1: S3's rest (a hook running a pack file by a glob or a built name escapes the `runs` scan) is the limit
  of the plan's rule 4, kept; `update --yes` says "nothing to change" on a forged lock while `check` finds it (5.1.6);
  the test pack's header still names A-D (fixed in G, 5.1.9).

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
