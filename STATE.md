# Bonsai: where it stands

Rewritten, never appended. Last rewritten 10 Oct 2026 at 12:29, when Rohan closed the session ("find a good stop point
for this session. and then wrap up. we will continue in next fresh session. also update the roadmap once done"): every
5.1 piece landed, 5.6 and 5.7 approved, the Go toolchain at go1.27.2, his roadmap at version 33; the 5.1 end verifier
was stopped part-way and runs again, fresh, at the next session's start (`records/runs/R-2026-10-10-*`).

## In short

Bonsai is being rebuilt as one small Go program that gives every project the same formats, packs, guards, recorder and
proof ladder (`design/one-pager.md`). The plan is approved (8 Oct). **Part 0 and all six parts of the walking skeleton
are done; the skeleton's last fresh verifier passed it. At the gate (9 Oct) Rohan chose path (a), the full Bonsai
1.0. He approved the plan for step 5, `design/plan-5.md`, on 9 Oct; step 5.1 (formats and engine to 1.0, 30-47 h,
re-ask at 61): every piece, 5.1.0 to 5.1.10, has landed; its end verifier runs next. The sections of 5.2 to 5.7 are all
written and approved (5.6's and 5.7's on 10 Oct), so every part of step 5 has its plan.** The gate report is `records/gate-skeleton.md`. Everything below is on `main`, pushed.

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

- **10 Oct, about 10:00: 5.6's section approved, with (A).** Asked "for each decision in plan 5.6, give me proper
  context and then ask me with good options", he took six decisions one at a time and chose the recommended one each
  time: (1) notes about him in his personal memory are his to save, from an agent's draft (not agents writing them);
  (2) this machine's settings and the studio's label files are all his, even how long pack copies are kept; (3) a
  newer release is known as planned (`status --full` at most daily, kept in the home's `cache/release.json`, shown
  offline); (4) the installers first run on his machines at 5.7 with 1.0; (5) `status --line` kept; (6) the personal
  index at 40 lines and 4 KB, a note 4 KB. Then "Approve": 13-20 h, re-ask 26. Spec §4's dated note on `status`'s home
  cache written. From now on a section comes to him as one question per decision, then the approval.
- **9 Oct, 18:32: 5.5's section approved, without the pin rule.** Asked whether he approves it with the new public
  `workflow` repo and a rule letting agents move a project to a newer pack version when nothing loosens, he chose
  "Approve, no pin rule": moving a pack's version, adding a pack to a linked project and taking one out stay a
  person's step; an agent prepares the change and hands him the exact line, and makes it only on his word (5.5's
  "Moving a pack's version: a person's step"). Spec §5's "Adopting a release stays a person's step" and 5.4's rule
  stand.
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
- **9 Oct, 15:46: no self-update.** He chose a self-updating program at 15:42, then: "actually scratch out the self
  update thing. i dont want that". The program stays his install per release on each side; from 5.6 `status` and
  `check` say when a newer release exists and an agent hands him the lines. Spec §3's dated note says so, and why the
  guard cannot be served online.
- **9 Oct, 15:35: Bonsai is managed by AI agents inside projects** ("bonsai will mostly (and probably completely) be
  managed by the ai agents within a project. so make sure the support for that is top notch"). Asked whether that
  includes the program on each machine, he chose "inside projects": agents link, update, fix, check, read status and
  edit; the `bonsai` program stays his install per release (root/admin), which keeps his 5.3 (a) safe. Three additions
  follow: `bonsai --help --json`, a machine-readable list of every command, flag, exit code and error word (5.1.10);
  every `check` finding's `next.do` the exact command that fixes it where one exists (5.1.6); an "operating Bonsai"
  skill in `base` (5.5). (A self-updating program was chosen at 15:42 and dropped at 15:46, above.)
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

## The one thing to do next: the 5.1 end verifier, then 5.2

`design/plan-5.md`, "Step 5.1". Landed, each with CI green (run reports `records/runs/R-2026-10-0*-5.1.*`):
- **5.1.0** (`6a3a419`): `TestEachFaultBlocks` no longer races a 200 ms budget (test file only).
- **5.1.1** (`6c6fc33`): consent to code: `--allow-exec` real for `init` and `update`; test-pack commits E and F; its
  fresh verifier passed the fix round.
- **5.1.2** (`c4816eb`): the format-0 reader, a hand port of `yaml.mjs` at `4a05eac`; its fresh verifier passed it.
- **5.1.3** (`254572f`; set 4 is `22dd08a`): formats set 4 (eight new schemas, 125 trick cases, the schema-compare
  test); its fresh verifier passed it.
- **5.1.4a** (`37125a4`): a Go type for each of the eighteen formats; `check --schema`.
- **5.1.4b** (`c2864c0`): every command's `--json` in set 4's shapes; the error words; one flag-and-exit table.
- **5.1.5** (`a9148e4`): the lock's `declares`, `format0` and `path`; the moved tag refused; the active-task function;
  the block; `bonsai.yaml` with every field; `check` offline from the lock.
- **5.1.6** (`107f843`): `check` and `status` to 1.0. `format.CheckWords` holds 33 words (24 findings, 9 warnings),
  one test walking every word, every `next.do` an exact `run: <command>` checked against the flag table; a pack's local
  `source` in `bonsai.yaml` is an `absolute-path` finding; `status --json` with every field, `--full` (newer tags,
  plugins, Claude Code's version) and `--active`; the Claude Code floor 2.1.294; this machine's `workspace.json`;
  the forged-lock loose end fixed. Tests: WSL 1976 run (2 skipped), Windows 1980 (1 skipped).
- **5.1.7** (`f070474`): `bonsai unlink`; `update` taking a pack out; the project-scope uninstall after Bonsai's
  write (measured on both sides); `update` removing this checkout's stale records under an older marketplace name;
  trust's `waiting` in every output and a `plugin-trust` warning; Claude Code's key order kept (`git diff` holds only
  Bonsai's lines on both sides); check 12's `git revert` run on both sides (the record stays disabled, no plugin
  loads); `TestPluginScopeOnly`. Tests: WSL 2012 run (2 skipped), Windows 2016 (1 skipped).

- **5.1.8** (`c0d7cd6`, 10 Oct): `check --write` rebuilds `.bonsai/tasks.md` from format-0 and format-1 tasks (the
  active task, newest id first), in the main checkout only (a worktree exits 4 naming it); with `--write` exit 0 or 3;
  a stale table the `tables` warning, never a finding (in a worktree, the main checkout's table is judged); `init`
  writes both tables. Tests: WSL 2018 run (2 skipped), Windows 2022 (1 skipped, twice tagged).

- **5.1.9** (`a034db5`, 10 Oct): `bonsai check --pack <folder>`, eleven words in `format.PackCheckWords` (a pack's
  files held to their formats as written, every key commented, template fields tables, closed lists, deny `why`s, no
  `plugin.json` version, the block's 40 lines, document kinds, protected globs, no `bash` by name, `runs`); test-pack
  commit G (`cf356d3`, documentation only), pushed. Tests: WSL 2047 run, Windows 2051.
- **5.1.10** (`8a45729`, 10 Oct): `docs/reference/lists.md`, generated by `go generate ./...` from the code's tables,
  held by `TestPageIsCurrent`; `bonsai --help --json` (`bonsai.help/1`, every word, flag, exit code and error word).
  **Formats set 5** (the help format added); **5.2.0's set is now 6**, not 5 as the plan's 5.2 section says.
- **The Go toolchain** (`3b6d599`, 10 Oct, Rohan: "we can go with go 1.27 latest"): `toolchain go1.27.2` (the `go
  1.25.0` line kept); `govulncheck` `v1.8.0` (33 standard-library findings on go1.25.9, none now), `golangci-lint`
  `v2.14.0`; CI runs 1.27.2. Counted in 5.7's hours (5.7.0's bump, done early).

**The next session starts here** (nothing is running; `main` is pushed):
1. **A fresh 5.1 end verifier** (Opus), on `main`'s newest commit (5.1's code is unchanged since `11d871c`; later
   commits are design and records). Its brief is the 12:04 entry of `records/runs/R-2026-10-10-5.1-verify.md`, in full:
   "5.1 done" checks 1-21 in its own clones on both sides, the go1.27.2 switch, the items the run reports placed for
   it, real Claude Code only through `claude-here` in scratch targets with the settings hashes before and after. The
   first verifier was stopped when Rohan closed the session, after passing every check it finished (no defect found);
   the next one finishes it: 6 on Windows, 7, 8, 19's timing on a quiet machine, 21's Mimas and studio look and private
   grep, and the placed items it lists, plus check 10 on the newest commit; the passed checks stand as evidence on the
   same code. Keep Windows' `claude.exe` (in `%USERPROFILE%\.local\bin`) off the PATH: the first verifier's filter
   missed it once (read-only `plugin list` calls; the settings file unchanged).
   On PASS: 5.1 is done; `STATE.md` rewritten; the roadmap's 5.1 card marked done. On FAIL: the fixes it names, then a
   verifier again.
2. **Then 5.2** (`design/plan-5.md`, "Step 5.2", approved 9 Oct; "5.2 starts once 5.1's end verifier has passed 5.1"):
   5.2.0, 5.2.1 and 5.2.2 side by side once the orchestrator has checked their files against what 5.1 landed. **5.2.0's
   formats set is 6**, not 5 as the section says (5.1.10 took set 5 for `bonsai.help/1`). Rohan's look on the log's
   field names is done (10 Oct, "Both fine"): `bonsai_path` and `bonsai_sha256` stand. 5.1.10's hand-off: the log's
   events and categories as a Go table registered on the `log` format, the reference page's two "not built yet" lists
   filled, `go generate ./...`.
3. **Carried into later briefs:** 5.6.5's brief counts a tag as a release only once its release is published (5.7's
   review; 5.7's section names it in "Stale or in tension"); 5.7's runs so far (84 minutes of planning and review, 33
   of the Go switch) go into 5.7.0's run report.

Placed by 5.1.6, 5.1.7 and 5.1.8 for later: 5.2.3 (the sessions rows in `check --write`, its help's
"joins in step 5.2.3" taken out); 5.2.0's set 5 (`uninstalled` in `changes.plugins[].result`; unlink's "left in
place" in JSON; `needs.mcp`; the lock's `declares` holding `needs`, `hooks` and `deny`; a place for `check`'s
notes; `status`'s `mode: full` and `checks`' layout; a lock `ref` for the moved-tag gap; the lock README's wording on
an old lock's `path`; `bonsai --help --json`'s shape with 5.1.10); 5.1.8 (the tables in `TestUnlink`); 5.3 (the guard
judging `unlink` as a person's under (ii)); 5.5 (the operating skill covers `unlink` and taking a pack out); the 5.1 end verifier and
5.2.0's start check (a plugin at the same commit and folder asked for `--allow-exec` again when the marketplace name
moves: kept strict); 5.6 (the machine folder `unlink` leaves, with `stranded`; `install.json`'s keys `path`,
`version`, `sha256`; the spec §10 Windows personal-memory warning, in no 5.1 piece); 5.1.10 (the
new lists: `CheckWords` with kinds, `checkLater`, `status`'s needs kinds, `checks.claude_code.state`, Claude Code's
measured key orders, the plugin results, `plugin-trust`).

How this session runs, for the next orchestrator: pieces and later parts' plans side by side; every brief carries its
rules in full; each builder rebases on `main` before it lands; a part's section goes through a fresh review, fixes and
a Haiku audit before it lands, then to Rohan under his (B); times in the run reports come from `date`; no test or
script reaches the real `claude` except a builder's scripted run in a scratch target through `claude-here`, with the
settings hashes before and after.

5.1's hours: 842 minutes (14 h 2 min) of 61 h, the stopped verifier's 25 included. 5.5's planning: 84
minutes. 5.6's planning: 67 minutes. 5.7's so far: 117 minutes (84 of planning and review, 33 of the Go switch), of
24 h. Step 5's Windows-only tally: 0. Option rounds in 5.1: one.

## Waiting on Rohan

- Nothing now. Every plan section is approved, and his looks are done (10 Oct).
- Later, as the plan has them: at 5.4, a pre-release `bonsai` in WSL (about 5 minutes); at 5.5, reading the `workflow`
  repo and making it public; at 5.7, the release settings, the `rc.1` tag and his approval, the installs, the trial's
  analysis, his own trial on a fresh project, his word for 1.0 (about 1 h 45 min in all, 5.7's section).
- Whenever he likes: turn GitHub Pages off (the old website).

## Loose ends

- Rohan's roadmap artifact is at version 33 (10 Oct, on his word at the session's close): Bonsai's cards for 5.1
  (built, its verifier next), 5.2 (up next), 5.3-5.7, the 1.0 trial and the after-1.0 items; the studio's cards
  untouched. Next update at 5.1's end (plan item 9).
- Another of Rohan's sessions also works in `~/Servers/Bonsai` (10 Oct: two commits taking Jev out of Bonsai's plans,
  pushed). Read `origin/main`'s new commits before each push, and never stage a file this session did not write.

- The WSL user settings file's baseline is `c3978abe...e73ee7` since Rohan's `/plugin` and `/reload-plugins` at the
  start of the 10 Oct session (told him; before it `9e049dea...80d8d6`, from his `/plugin` of 9 Oct). Windows' is
  unchanged, `2b6295c1...4ff6c9`.
- From 5.1.1: S3's rest (a hook running a pack file by a glob or a built name escapes the `runs` scan) is the limit
  of the plan's rule 4, kept; `update --yes` says "nothing to change" on a forged lock while `check` finds it (5.1.6);
  the test pack's header named A-D until G (`cf356d3`, 5.1.9), which fixed it.

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
