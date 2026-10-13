# Bonsai: where it stands

Rewritten, never appended. Last rewritten 13 Oct 2026 at 14:00. **Step 5.4 is done: Bonsai guards its own repo and
proves its own work with `bonsai ladder`.** Rohan installed the pre-release (`a88436a`, SHA-256 `356f787c...afeef644`),
made the link and the first climbs himself; the floors are 4188 and 4194; the fresh end verifier passed every check but
one clause, its D1, which Rohan had fixed ("Fix now", `T-5491`, the first piece landed by the landing rule). **Next:
5.5's re-plan with the agent-pipelines redesign** ("Now: step 5.5" below). From here every change lands by
`CLAUDE.md`'s rules "from the link".

## In short

Bonsai is being rebuilt as one small Go program that gives every project the same formats, packs, guards, recorder and
proof ladder (`design/one-pager.md`). The plan is approved (8 Oct). Part 0 and all six parts of the walking skeleton
are done (gate report `records/gate-skeleton.md`). At the gate (9 Oct) Rohan chose path (a), the full Bonsai 1.0, and
approved the plan for step 5, `design/plan-5.md`; every part's section, 5.1 to 5.7, is written and approved.

- **Part 0, the formats** (`3770d04`): a JSON Schema for each format, examples, trick files, a raw-byte manifest and
  its Go test. Now set 7 (5.4.0, `6688f81`).
- **The skeleton, parts 1-6** (`17f2938` to `5633721`). The test pack `LastStep/bonsai-test-pack` is public, `main` at G
  (`cf356d3`).
- **Step 5.1, formats and engine to 1.0: done** (code `11d871c`; verified on `5fa9bd8`).
- **Step 5.2, the recorder, logs and asks: done** (code at `8243800`), the redactor with known limits (below).
- **Step 5.3, the guards: done** (code at `750ca07`): the hook adapter, the guard, the delete check, the stop gate,
  Bonsai's own deny rules, `check`'s settings findings, the hook lines' fixed form (Rohan's (a)) and the commands' own
  refusal of a person's steps in an agent's session ((ii)). Run reports `records/runs/R-2026-10-11-5.3.*`.
- **Step 5.4, the ladder runner and the switch: done** (13 Oct). `bonsai ladder` (rung jobs that end what a rung
  leaves, on both systems; rung 0 the guard rung; floors, ratchets and Bless; new tests must fail on the base; git
  integrity; the settings record that refuses looser checks until a person re-records them), then the switch: the
  pre-release installed by Rohan at `/usr/local/bin/bonsai`, the link `c1bb21f` (by Rohan, `bonsai init`), the floors
  `6f25dd6` (`go_tests` 4188, `go_tests_tagged` 4194, from his first climb), the end verifier (`T-5490`), and D1's fix
  `d7037dc` (`T-5491`). Code at `d7037dc` (the binary's at `a88436a`; `d7037dc` changes only tests). Run reports
  `records/runs/R-2026-10-1[23]-5.4.*` and, in format 1, `R-2026-10-13-T-5490-verify-5.4-end.md` and
  `R-2026-10-13-T-5491-named-session-tests.md`. Tasks: `T-5407`, `T-5490`, `T-5491`, all `done`.

Hours (each run's minutes from the run reports): part 0 45 minutes; the skeleton 6 h 18 min. **5.1: 866 minutes. 5.2:
1,061 minutes. 5.3: about 2,013 minutes (33 h 33 min). 5.4: about 2,372 minutes (39 h 32 min) of its 28-44 h
estimate, under its 57 h re-ask line** (the verifier's count, 2,362, plus `T-5491`'s 10; its 79 minutes of planning
included; the orchestrator's own time between runs only partly in them). Later parts' planning: 5.5 84, 5.6 67, 5.7 117.
Step 5's Windows-only tally: about 143 minutes of 8 h (5.3's 36; 5.4's about 107: 5.4.1b and 5.4.1c's Windows rung
jobs). **Questions put to Rohan in 5.4: five option rounds** (builders stay subagents; memory kept on; K1 fixed; K2 a
known limit; D1 fixed), **so stop line 3 (more than two in a part) is crossed**, each in the open (the switch's own
questions and the verifiers' findings, not a section that left too much open). No change to Mimas or the studio's
repo.

## Rohan's decisions

- **13 Oct, about 13:35: D1 fixed now** ("Fix now"): the end verifier's one must-fix (two tests read the caller's
  `BONSAI_TASK`, so a named session's climbs are red and its stop gate never lets it end) fixed as `T-5491`, test-only,
  no new pre-release, landed by the landing rule (`d7037dc`); offered: a known limit fixed at 5.5's start.
- **13 Oct, about 13:45: stall stops for subagents go into 5.5's redesign.** He asked whether an idle agent can be
  stopped, "even if we stop an idle agent, the message of that would propogate upwards to the orchestrator or any other
  agent that started it, and they would be able to handle whatever needs to be done". Read in Claude Code's docs: no
  hook fires on a timer or on idleness; hooks see a subagent's calls with its `agent_id`; a subagent definition's
  `maxTurns` returns its output, marked partial, to the parent; a parent's stop tells it only "cancelled". So a soft
  stop (the guard refusing a subagent's calls past a budget, telling it to report) carries its report up; a true idle
  watch needs a clock outside Bonsai. He declined an ad-hoc watchdog for this session (memory
  `agent-pipelines-redesign`).
- **13 Oct, 12:31-12:56: Rohan ran the switch's steps himself** ("now i would like to do some steps myself if possible,
  so i get familiar with the program"): the install, `bonsai status`, `bonsai init` (read and answered `y`), and both
  first climbs, `bonsai ladder --task T-5407`; the orchestrator read, committed, probed the guard and set the floors.
  Then: "okay i believe you can take this from here. and close out 5.4".
- **13 Oct, 10:59: K2 is a known limit after 1.0** ("C. Known limit only"): a climb hands the caller's environment to every
  rung, so a fake program first on an agent's PATH can make a red branch climb green; until fixed, CI after the push is
  the independent check. His direction for the fix: "task done needs a green ci, and ci would run the ladder so that
  should be enough" (a `bonsai ladder --ci` job in CI; a task moves to done only with it green).
- **13 Oct, about 08:20: K1 fixed now** ("A. Fix now"): Bonsai records the checks' settings it last accepted and refuses
  looser committed ones until a person re-records them (`config-looser`); tightening is automatic; an added rung counts
  as looser (its command runs inside the climb).
- **13 Oct, 07:08: Claude Code's memory stays on in Bonsai's repo until 5.6** ("A. Keep them on"): one line of his in
  `.claude/settings.local.json` at his sitting; it comes out at 5.6.
- **13 Oct, 00:00: build several pieces, then one verification pass**, where efficient and without lowering the bar:
  "whenever it is possible to build multiple steps together, and then just do one verification pass over then ... only
  do this when it seems like a good efficient choice, and not something that compromises the quality of work". 5.4
  already has three passes for eight pieces (V1 beside 5.4.4/5.4.5); merging V1 into V2 was weighed and not taken
  (memory `batch-verification`).
- **12 Oct, 23:58: past records are left as they are** ("dont need any action for past records anywhere, like logs,
  tasks, runs etc. just let them be. only focus on the main build"): no lock, edit or cleanup of past run reports, logs
  or tasks at the switch or later; part 0's report keeps its quoted id (memory `past-records-left-alone`).
- **12 Oct, about 23:42: builders stay subagents, as planned** ("A. Keep the plan"): the landing rule (a green result at
  the landed commit, written only by the installed `bonsai`, audited by the end verifier) is the lock; a subagent's end
  stays ungated; how builders start is for 5.5's agent-pipelines redesign. Offered: own `claude -p` sessions (about 1-2
  h) or a `SubagentStop` line (about 3-5 h); plan lines 2762-2789 stand.
- **12 Oct, 00:46: 5.3 closes on the end verifier's two holes as known limits** ("Log as known limits"; recommended was
  a fix and a short fresh check, about 1.5 h). CDPATH and `cdable_vars` cd-resolution (`CDPATH=.. cd s && rm -rf .git`
  deletes the checkout's own `.git`) and Git Bash's `/proc/cygdrive/` path form, both deliberate spellings, are listed
  in `internal/guard/shell.go`'s known limits and fixed after 1.0.
- **12 Oct, about 00:40: Sonnet helpers for mechanical runs by default.** He asked whether the end verifier used helpers
  ("sounds like a sonnet helper agent would be a good choice"): a verifier or big builder hands the Windows check 10,
  timings, table re-runs and scripted sessions to Sonnet helpers beside its own judgment work, and says why when it
  does not (`rules.md`; memory `nested-subagents`).
- **11 Oct, about 20:30: the verifiers' bar stays adversarial; at most two verifier rounds per piece**, more only if
  necessary. Asked whether verifiers should drop the hunt for deliberate escapes (recommended), he said no: "i like that
  we are testing for cases where an agent might deliberately want to break the system. as that is the biggest security
  flaw the system can have ... so i dont mind the current approach. we limit the verifications to 2 rounds, we do more
  rounds only if its necessary" (memory `finish-lines-on-fix-loops`).
- **11 Oct, 19:4x: no long-lived token for agents' Windows sessions** ("No new token"). The Windows login kept lapsing
  (cause below); the fix is the rule that `claude.exe` is never reachable from a `bonsai.exe` run and a warm-up session
  before Windows runs, not a year-long credential in agents' environments.
- **11 Oct, 19:19: fix round 3's bar, "A: what round 2 opened"** (the two regressions, `xargs`'s input reaching
  `bonsai`, `~-0`); V2c's other neighbours known limits.
- **11 Oct, 17:36: the guard's fix loop's finish line, "Scoped".** V2 and V2b had failed 5.3.2 and 5.3.3 on edge
  cases. Asked (scoped, recommended; as in 5.2; strict), he chose: fix round 2's items (V2b's regressions D1, D2, D5,
  the cheap holes D3, D6, D7, the over-refusal W2), then a third fresh verifier checks those fixes and looks for
  breakage, then they land; **V2b's D4 long tail and W1 are known limits, fixed after 1.0** (below). 5.3's third option
  round.
- **11 Oct, 11:42: the delete check's ordinary refusals stay as planned** (`rm *.log`, `rm -rf build/*`, `git clean -fd`,
  `find . -name '*.tmp' -delete`, `rm -rf "$tmpdir"`): "ok, leave it as planned"; crying wolf judged on the verifiers'
  evidence.
- **11 Oct: how agents work** (memory notes, `~/bonsai-checks/briefs/rules.md` carries them): a subagent with a big
  task may spawn helpers, one level down, a helper none (13:14); past about 500k tokens an agent's next work goes to a
  fresh agent with a hand-over brief (17:39). Also: he wants the agent-team design shown as diagrams, not text
  (artifact "Bonsai Agent Team", https://claude.ai/artifact/HdNVzmTDBNZA8nj4mFzgMU, version 2).
- **10 Oct, 22:54: the redactor's finish line, "One more round, then land".** Two fresh verifiers had failed 5.2.1 on
  rare constructed shapes, each round about 2 h, with no set end. Asked to choose (studio parity, recommended; strict;
  one more round, then land), he chose: the third fresh verifier checks by the plan's strict rule, and 5.2.1 lands
  after it whatever it finds; anything it still finds is logged as known limits and fixed after 1.0. 5.2's first
  option round.
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
- `LastStep/bonsai-test-pack`: public, `main` at G (`cf356d3`), no tag.
- The old website stays on GitHub Pages until Rohan turns Pages off.

## Now: step 5.5, the packs, re-planned with the agent pipelines

`design/plan-5.md`, "### Step 5.5" (~3430): 19-30 h, re-ask at 39, approved 9 Oct without the pin rule. **Before any
piece is built, its re-plan:** Rohan's 12 Oct agent-pipelines redesign changes the `workflow` pack's roles, so the
section comes back to him under (B) (memory `agent-pipelines-redesign`): work sorted into kinds, each a pipeline of
specialised agents (planning: a researcher, an auditor, a planner that questions the person, a plan reviewer; building:
workers, helpers, several verifiers); agents talking to each other; little reliance on Claude Code; and, from 13 Oct,
**stall stops for subagents** (a budget per role in minutes, turns or red climbs in a row, enforced by the guard as a
soft stop that tells the agent to report; `maxTurns` on each role; what the parent does with a partial report). The
research is in `~/bonsai-checks/research/` (notes and reports, 12 Oct); the current setup's diagram is the artifact
"Workflow Pack Agents" (https://claude.ai/artifact/KjzSYPJ9LdXhcbCJ7iKrrU). Proposals go to him as one question per
decision with diagrams, then 5.5 re-costed and approved.

**How work runs from here** (`CLAUDE.md`, "from the link"): every piece a task `T-55yy` in `records/tasks/` (the
orchestrator alone writes and moves them), its run report `R-<date>-T-55yy-<topic>.md` in format 1; builders are
subagents in plain worktrees from `main`, climbing `/usr/local/bin/bonsai ladder --task <id> --json`; a piece lands only
on a green local result at its HEAD whose record names the installed `bonsai`, plus check 10's Windows half before the
push (back from 5.5: `~/bonsai-checks/briefs/rules.md`) and CI after it; the part's end verifier audits every landing.
The builders' rules: `~/bonsai-checks/briefs/rules.md` (**update its `Claude-Session` line to the new session's
first**). The floors are raised once a part, after its last records commit.

**For 5.5 from 5.4's close:** the runner could clear `BONSAI_TASK` and `BONSAI_ROLE` from the rungs' environment, so no
test reads a session's own task again (D1's repeat; a change to the binary, so with the next pre-release); the gate's
next step names `bonsai ladder --task <id>` where `/usr/local/bin/bonsai ...` is meant (O1); `CLAUDE.md`'s scratch-link
line covers only `--allow-exec` and a person's refusals, while any scratch link typed on the shell line in a project
holding a `bonsai.yaml` is refused (D2: the guard reads the real home, not the command's `BONSAI_HOME`; the script
route works), or the refusal's next step names the script for a scratch home. **The next pre-release** (5.5's packs
need a newer `bonsai`): built the same way from a clean clone, the current one first copied to
`~/bonsai-checks/prerelease/previous/` (its number is in 5.4.7's run report), Rohan's install lines and number sent in
one batch.

**Facts measured that later work rests on:** Claude Code 2.1.296 on both sides (Windows updated 11 Oct, 19:45). A Stop
hook printing "No such file or directory" and exiting 2 lets the session end; Stop hooks keep a turn going eight times
in a row, the ninth overridden, and the count restarts when a subagent finishes. `"onFailure": "block"` blocks a guard
line that cannot start on both sides (2.1.296). In `claude --worktree` sessions `CLAUDE_PROJECT_DIR` is the worktree;
after the `EnterWorktree` tool it stays at main while the payload's `cwd` moves (the stop gate judges the `cwd`'s
checkout, tied by git's own link). `CLAUDE_CODE_SHELL_PREFIX` (both sides) and `BASH_ENV` (Windows) reach a hook line;
`SHELL` and `CLAUDE_CODE_SHELL` do not; a project `env` cannot blank `CLAUDE_CODE_CHILD_SESSION`. Git Bash runs a
two-line `#!` file named `bonsai`. `Edit(<path>)` deny rules do not hold after a shell `cd`; `Edit(/<path>)` and
`Edit(//**/...)` do. A session started in a subfolder, or after `/cd`, loads none of the project's settings. `-p`
sessions run the PowerShell tool on Windows. **The Windows login:** `bonsai.exe init --allow-exec`/`update` run with
Windows' own `Path` start the real `claude.exe plugin ...`, which, meeting a run-out access token on 2.1.294, spent the
refresh token without saving the new one and wiped the login (9, 10 and 11 Oct); the rule in `rules.md` keeps
`claude.exe` unreachable and adds a warm-up (the credentials file's size: about 524-591 bytes live, 296 wiped).
Timings (5.3's end): the guard's typical call on WSL 3.5-5 ms p50, under 5 ms at p95 direct; on Windows about 13-16 ms
direct, and about 130-150 ms through the Git Bash line, where Git Bash's own start (`bash -c true` about 80 ms) is the
cost on this machine (Defender on, low free memory), Bonsai adding 6-9 ms.
**Measured at the switch (13 Oct):** the guard took up the new hook lines in the running interactive session at once,
no restart (a file-tool edit of `go.mod` refused right after the link); a climb of `main` takes 94-107 s (the two test
rungs 46-59 s each, vet under 1 s warm); a named session with a green climb at its HEAD stops at once (`stop-proven`);
the main checkout's `.bonsai/local/` is the guard's: an agent's `cp` out of it is refused.

## Waiting on Rohan

- **Nothing now.** 5.5's re-plan comes to him as one question per decision.
- **Whenever he likes:** whether to allow deleting git's own lock files by exact name (`rm -f .git/index.lock`, a
  common agent fix, refused because the project's `.git` is guarded for deletes; told 11 Oct 16:34, left as planned);
  whether the redactor's shared `=>` and `:=` limits go in a named part before the release; turn GitHub Pages off (the
  old website).
- **Later, as the plan has them:** at 5.5, a newer pre-release `bonsai` in WSL (about 5 minutes), reading the
  `workflow` repo and making it public, about 3 minutes on Windows (trusting a scratch folder); at 5.6, taking his
  `"autoMemoryEnabled": true` line out of `.claude/settings.local.json` once the personal layer holds the notes; at
  5.7, the release settings, the `rc.1` tag and his approval, the installs, the trial's analysis, his own trial on a
  fresh project, his word for 1.0.

## Loose ends

**Known limits, after 1.0, noted at 5.4's close (Rohan, 13 Oct: "just note them down, no need to fix or verify"):**
**the rung 0 race** (V2d: rung 0 judges by main's `bonsai.yaml` as it reads when rung 0 runs, not the committed settings
checked in prepare; a looser file swapped in during the climb and put back makes a red branch green, the gate 0;
reproducers `~/bonsai-checks/verify-5.4-V2d/repro/r1-race.sh`, `r1b-race-commit.sh`; fix: rung 0 judges with the
committed settings, and the climb re-reads main's file once it holds the lock); **the gate's git timeout** (under heavy
load its one git call can pass its 4 s limit; the refusal then says to fix `bonsai.yaml` instead of naming the timeout,
and one stopgate test is timing-bound); **the machine record is writable** by an agent's shell (deleting it or nulling
its `checks` lets the agent's own `update` write a free first record); a task file's grants are read from main's
working tree (the landing rule's audit of each task's `bonsai.allows` is the check); **two CI flakes**: on Windows,
`internal/rungjob`'s `TestOutputAndExit` ("ended [{684  }]": a job member that has just exited is listed and, gone
before it can be opened, reported with no name; fix: `member` skips a pid whose process is gone; red once, on
`f9e28e3`), and on Linux, `internal/ladder`'s `TestLooserKindsRefused` in its temp folder's cleanup ("unlinkat
.../.git/objects: directory not empty"; likely cause: since git 2.47 the upkeep `git commit` starts runs detached and
locks the object database, and CI has git 2.55, this machine 2.34; fix, test-only: the fixtures' git calls set
`maintenance.auto=false`; red once, on `81cb8d7`); **D2** (a scratch link typed on
the shell line refused, above); the end verifier's observations: `T-5407`'s `bonsai.allows` leaves out
`.bonsai/.gitignore` and the two tables its Owns names (harmless: `init` wrote them, run by Rohan), Bonsai's block
imports `@records/memory/INDEX.md`, which does not exist (the plan allows it; `check` finds nothing).
**Known limits, after 1.0:** **K2** (a fake program first on the caller's PATH fools a rung; fix: CI runs the
ladder and a task is done only with it green, Rohan 13 Oct); Git LFS and git-crypt checkouts always read dirty; the packs'
protected paths in `.bonsai/lock.json` outside the settings record (5.5); a `documents` kind added moves a pack's place
(5.5); the machine record in the home is writable by an agent's shell; the redactor's (5.2: L1, and the shared `'password' => 'x'`, `password := "x"`, `"""x"""`,
unfolded dotted i and full-width letters, escaped JSON keys, bare secrets, a lower-case secret as a program name); the
guard's, listed in `internal/guard/guard.go`, `shell.go` and `shell_env.go`: the variable check (V2b's D4 and W1), env's
own `-S` rules (`\_`, a `-S` text holding `;`), PowerShell casts, parentheses and comma lists on `$env:`, the git icase
variable spelled with quotes or encoded, a planted `.git` with a garbage HEAD, `parallel` as a runner, bonsai's words
from a variable or substitution outside `xargs`, the program itself from `xargs`'s input, `env -C`/`sudo -D` twice,
`~-`/`~N`/`~-N` after a line's own `cd`, **CDPATH and `cdable_vars`, Git Bash's `/proc/cygdrive/`** (the end verifier's
two, by Rohan's word), printf/echo escapes piped to a shell, `Start-Process -ArgumentList`, `rm` of a link to itself
refused; the commands' own refusal is a tripwire against programs the agent runs (V3's O2: a build stamped with
`scratchRoot=/`, a forged `BONSAI_HOME`, a renamed old binary); shell writes and scripts run by name (question A; rung
0 from 5.4).

**Placed for later:**
- **The next formats set (5.5's):** a shared active-task fixture case for a task file holding only `garbage: [` (5.4.3
  fixes the reader with Go tests); `ladder.schema.json`'s `leftovers` description still telling today's `shared_deploy`
  shape (typed open, so Bonsai's `{checkout, ms}` fits).
- **From 5.4:** the worktree without its own `bonsai.yaml` (rung 0 judges against main; the guard's own limit,
  `guard.go:46-48`) and V3's O4 (the gate proves HEAD; only commits land) are known limits after 1.0;
  `golang.org/x/sys` stays out (Windows calls loaded lazily through `syscall`); `flock`, `chrt`, `taskset`, `strace`
  are not runner words; the start-up cost left (`internal/recorder` about 0.14 ms, `internal/clean`).
- **5.5:** `design/plan-5.md` ~4031 (the operating skill) says an agent may add a rung as a stricter change: since K1
  an added rung is a person's step; `bonsai init --allow-exec` and `update` run `claude plugin install` on the person's machine: on Windows with
  2.1.294, meeting a run-out access token, that call can sign the person out of Claude Code (a warm session first, or a
  `check` finding naming the version); where the orchestrator's rules live (spec §2 names it as pack content; 5.5's
  `workflow` has five roles and none); `needs-yes` is `who: person` while 5.5.3's operating skill has an agent run
  `--yes` after the preview; a pack's relative deny rule does not hold after a shell `cd`; `needs.mcp`; a lock `ref` for
  the moved-tag gap; the marketplace observation; 5.1.9's limits; contract §5.4's `workflow.model` lacks haiku (5.5.4).
- **5.6:** installers write the binary by rename and the line tests `[ -s ]` (V3's O1: a 0-byte file at a place makes
  every line exit 0); Claude Code shows the whole hook line before a refusal, so its "no bonsai is installed" words show
  when Bonsai is installed (V3's O5, Rohan saw it in his sitting); the machine folder `unlink` leaves; `install.json`'s
  keys; spec §10's Windows personal-memory warning; a per-id index of the machine records.
- **One home, later:** the recorder's `taskAllows` duplicates the guard's `allowsOf`; the recorder's `underHome` the guard's
  `fromHome`; `workspace.VerifyMain`'s "cannot be resolved" reason quotes the system's error text (the stop gate's
  `noPaths` removes its path).

**Housekeeping:**
- Rohan's roadmap artifact ("Trinetra Roadmap", https://claude.ai/artifact/XXKTi6geneu1pdQ4h4miw2): Bonsai's cards
  updated at 5.4's end (13 Oct), next at 5.5's end.
- Another of Rohan's sessions may also work in `~/Servers/Bonsai`: read `origin/main`'s new commits before each push, and
  never stage a file this session did not write.
- The user settings files' baselines: WSL `4ccccc6b...a0db` (changed 12 Oct 10:59, before 5.4's work; unchanged at its
  close), Windows `2b6295c1...4ff6c9`. Rohan's own `~/Servers/Bonsai/.claude/settings.local.json` holds his
  `skillOverrides` and `"autoMemoryEnabled": true` (out of git; the guard refuses agents' edits of it).
- The pre-release's source clone `~/bonsai-checks/prerelease/src-a88436a` and the build `~/bonsai-checks/prerelease/bonsai`
  stay for the end verifier's record; the 5.4 end verifier's scratch (`~/bonsai-checks/verify-5.4-end/`, scripts in
  `~/bonsai-checks/scripts/verify-5.4-end/`) and `T-5491`'s (`~/bonsai-checks/T-5491/`, `scripts/T-5491/`).
- Scratch under `~/bonsai-checks/` and `%USERPROFILE%\bonsai-checks\`: every piece's and verifier's folder, the launchers
  (`bin/`, unchanged: `8a9398e2...`, `17ad3ce3...`), the 5.3.7 scripts and their detached worktree
  `~/bonsai-checks/5.3.7/src` (at `97c19eb`; remove with `git worktree remove` when no longer needed). Files there hold
  the Windows user name: never paste from them.
