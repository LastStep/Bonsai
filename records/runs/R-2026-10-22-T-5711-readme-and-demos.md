---
format: bonsai.run/1   # fields: bonsai check --schema bonsai.run
id: R-2026-10-22-T-5711
task: T-5711
role: orchestrator
model: opus
started: 2026-10-22 06:04
finished: 2026-10-22 13:18
outcome: merged
commits: []
labels: {}
---

# Run: T-5711, the README and demo videos

- Who: the orchestrator, with the agents it briefs
- Why: the director, 20 Oct: "add a step before 1.0 release, i want to improve the readme, and create some demo videos
  showcasing how bonsai works (and does for a project)"; 22 Oct: "okay lets go ahead with the readme and demo"

> Opened before the first edit and appended to as the work goes. Every time UTC (`date -u`; the director, 21 Oct).
> Hours come from the log; agents by their message ids.

## Log

- 06:04 The talk with the director; five answers kept in `~/bonsai-checks/T-5711/notes.md`: README restructured on the
  research's outline; demos of both kinds (a scripted GIF of bonsai's commands, a real Claude Code session recorded
  by an agent); both in `docs/demos/` ("option 1 now, and then i want to record my screen and do a voice over as a
  youtube video later"); the session's story "a task, start to done"; the diagram a drawn SVG, light and dark.
- 06:07 Auditor (sonnet) `M-2026-10-22-dba39583` briefed: the facts the demo story rests on. T-5711 moved `plan`.
- 06:10 Auditor done (final reply; it may not write). Facts the plan rests on: bonsai.yaml is person-only on the
  guard's floor in every project (`internal/guard/floor.go:48`), its refusal three lines (why, next, rule); **the stop
  gate holds only a session with `BONSAI_TASK` set** (`internal/stopgate/stopgate.go:249-253`, rule
  `stop-not-named`), and needs a run report and a ladder result before it reads red; `bonsai logs` lists sessions and
  prints raw JSON lines (no readable view; `jq` is installed); README.md and docs/ are not protected (`bonsai.yaml:51-70`),
  so `bonsai.allows` stays empty; no `base-v*` tag yet; rc.1 has no `bonsai now` (main only); vhs 0.11.0, ttyd, ffmpeg
  and tmux are installed, asciinema and agg not; a ttyd not ours runs on port 7681. The orchestrator's own checks:
  `.gitignore:31` ignores `*.png` outside `assets/demos/`; `assets/demos/` holds 0.4.3's demo (3.5 MB, its `.tape`s).
- 06:12 Plan drafted, `records/plans/P-T-5711.md` (four pieces, T-5718 to T-5721; the dry run and an Opus facts
  verifier; final recordings after T-5710's fixes, before rc.2's tag). Verifier (opus) `M-2026-10-22-177afda3` briefed:
  the plan's facts review.
- 06:17 Verifier done: **FAIL**, four blocking findings, all held: (1) the person's `bonsai update --yes` is refused by
  the installed `bonsai` from any agent's session, scripts and vhs included (`cmd/bonsai/agent.go:5-43`); (2) the plan
  had a builder start Claude Code through a script, routing round `call-refused` (`internal/guard/rules.go:162-167`);
  (3) no scratch Bonsai home; (4) C's GIF could hold a home path (fix 13), and two wrong claims (the README does not
  point at `assets/demos/`; the protected lists are `bonsai.yaml:42-67`). Facts-lens count: 4 blocking, 4 held, 0 not.
  Revision 2: a stamped build of `main` with a scratch home for C and D; builders write scripts, a runner runs any that
  starts Claude Code; every frame of both checked; the dry run's step 6 the director's; setup committed (`stop-dirty`),
  permission mode and folder trust off camera, `CLAUDE_CODE_*` cleared, the log through `jq`.
- 06:19 Verifier round 2 `M-2026-10-22-831e602e`: **FAIL**, round 1's four answered; three new blocking
  findings, all held: a `/tmp` project is refused by a stamped build (`test-build`); `claude-here` sets
  `~/bonsai-checks/home` and puts an unstamped build first on PATH; clearing `CLAUDE_CODE_*` would change
  `CLAUDE_CODE_CHILD_SESSION`, a person's command (`internal/guard/rules.go:149`). Facts-lens count so far: 7
  blocking, 7 held. Revision 3 fixes each as the verifier proposed (projects under `~/bonsai-checks/`, the prompt short,
  the result-path line out of frame until fix 13; D's home `~/bonsai-checks/home` and `CLAUDE_HERE_FIRST`; only
  `CLAUDE_CODE_BRIDGE_SESSION_ID` unset), plus its notes. No third round: two is the director's limit "more only if
  necessary" (11 Oct); the fixes are the verifier's own wording, and the final facts verifier reads the made scripts.
- 06:50 The director: **"Approve (Recommended)"**: P-T-5711 approved (revision 3), T-5711 `approved`. Tasks
  written: T-5718 (A, README, opus), T-5719 (B, picture, fable), T-5720 (C, scripted GIF, sonnet) `running`; T-5721 (D,
  session, sonnet) `todo` until C's stamped build is ready. Worktrees `~/Servers/Bonsai-T-57{18,19,20}` from
  `ae9119f0`. Builders started side by side: `M-2026-10-22-081f8b9c` (opus), `M-2026-10-22-25b82dbc` (fable),
  `M-2026-10-22-d3a9ec4e` (sonnet).
- 06:59 T-5719 (fable) done, report `M-2026-10-22-5c48e747`: branch T-5719 at `4f4d86d4`, two SVGs 6,655 bytes
  each. Climb read by the orchestrator: `.bonsai/local/ladder/T-5719.json` mode local, green, git.sha `4f4d86d4`,
  dirty false, sha256 `1b8a91af...99beae`; rungs 0-5 green (tests 4,874, tagged 2,173). The orchestrator looked at
  both renders at 830 px; the diff is the two files only. T-5719 `verify`; the renders sent to the director.
- 07:01 T-5718 (opus) done, report `M-2026-10-22-36687838`: `32863772`, README 10,584 bytes; climb read: local,
  green, git.sha `32863772`, dirty false, sha256 `37d92ef4...014bde` (tests 4,883, tagged 2,176). Steps 1-3 and 6
  walked on T-5720's stamped build with a scratch home (`~/bonsai-checks/scripts/T-5718/`); 4-5 not run (Claude Code).
  Install lines copied from `docs/install.md` into `<details>` (not three lines: the plan's "three lines" read as
  three steps). The orchestrator's read of the diff: one finding, the Install NOTE "until it is, the download fails"
  would need removing after 1.0.0 is out, when main takes only records (note 5.7.10): fix briefed
  `M-2026-10-22-4d77ae4d`. Step 6's "rc.2 puts that first" goes in the final pass after T-5710.
- 07:04 T-5718's fix done, `M-2026-10-22-8dcf5062`: `c9c89ab0`, the NOTE now "These lines install 1.0.0. For
  another release or a pre-release, put its version ... in place of 1.0.0 on every line." Climb green, sha256
  `cbc01198...4879ca`. T-5718 `verify`.
- 07:13 T-5720 (sonnet) done, report `M-2026-10-22-23b07ccc`: `1d450532`, GIF 785 KB, 41 s; climb read: local,
  green, git.sha `1d450532`, dirty false, sha256 `92b32f10...13a87e`. Stamped build for D at
  `~/bonsai-checks/T-5720/bin/bonsai` (main at `ae9119f0`). **Deviation:** recorded with a second stamped build whose
  scratch root is `/tmp/bd` (outside the scratch folders CLAUDE.md names), so no home path is in frame. The
  orchestrator looked at the stills (`~/bonsai-checks/T-5720/stills4/`): no home path or name; but init's output shows
  "This test build's hook lines run the bonsai at /tmp/bd/bin/bonsai", a test build's line 1.0 never prints, and
  `update` leads "nothing to change" (fix 11, rc.2). Both go to the final pass; the way to make it (the director
  running the tapes on the installed rc.2, or a stamped build) to be put to the director. T-5720 `verify`.
- 07:13 T-5721 `running`; worktree `~/Servers/Bonsai-T-5721` from `ae9119f0`.
- 07:21 T-5721 builder (sonnet) started, `M-2026-10-22-c5c40d49`. Drafts sent to the director (the GIF, the
  README). Asked how the final pass is made, the director: "walk me through every part of the recording. i also want an
  audit of all bonsai features, and a specific document on github which is linked on the readme ... and then i want to
  see what all we can show in the recordings, and how" (notes, answer 6). Four auditors (sonnet), one a domain, side by
  side: setup and packs `M-2026-10-22-2a9bbef6`, guard and stop gate `M-2026-10-22-2b4d4b60`, records and recorder
  `M-2026-10-22-1fcea704`, ladder and proof `M-2026-10-22-bcb8999e`. T-5721's take continues as evidence.
- 07:25 Auditor, ladder and proof, done (`M-2026-10-22-bcb8999e`'s reply, kept at
  `~/bonsai-checks/T-5711/features-ladder.md`, 30.7 KB). Found: CLAUDE.md lists "a rung ... added" as a stricter change,
  but the code reads it as looser (`internal/workspace/checks.go`, rungLoosenings; base:workspace agrees with the
  code): a stale CLAUDE.md line for T-5710's list. Nothing runs `bonsai ladder` in CI (base:ci runs `check`; BUG-053).
- 07:26 Auditor, setup and packs, done (`M-2026-10-22-2a9bbef6`'s reply, kept at
  `~/bonsai-checks/T-5711/features-setup.md`). Notes for the document: no migration from 0.4.3 (the refusal says it
  "comes later"); no CLI to add or remove packs (edit bonsai.yaml, then update); `mcp`/`shell` needs listed but not
  written by this build; the workflow pack's 12 deny rules name the studio's own folder (`~/.trinetra/...`), a path no
  other user has (for the director); BUG-022's status unverified.
- 07:27 Auditor, guard and stop gate, done (`M-2026-10-22-2b4d4b60`'s reply, kept at
  `~/bonsai-checks/T-5711/features-guard.md`): 68 rules on main; the guard is a tripwire, not a wall (spec :678,
  Decision D); Read/Grep/web/MCP unjudged (walls only); the sandbox is in the spec, not built; the spec's hook table is
  out of date against the code (Monitor, roles' tools, the second line, the time lines).
- 07:28 Auditor, records and recorder, done (`M-2026-10-22-1fcea704`'s reply, kept at
  `~/bonsai-checks/T-5711/features-records.md`). All four inventories in. Notes: no code found that writes the
  personal-memory import the skill describes; `status --line` is not wired by init (the person adds it).
- 07:28 The director: **"remove the trinetra paths from the workflow pack. also close any unnessary
  monitors"**. Monitors: ten background `sleep 299; ls ~/bonsai-checks/T-5721/takes/1` waits of T-5721's builder had
  piled up; stopped (kill of each bash and its sleep), the builder told to keep one. Left: the runner's `pgrep vhs
  session.tape` loop (a take in progress) and pid 231's ttyd (not ours). The Trinetra paths: task T-5796.
- 07:30 T-5796's builder (sonnet) done (final reply; `bonsai message` not used): workflow `612bc333`, 2 files,
  +8 -43: twelve deny rules out, `deny: []`, README's "The walls" rewritten; `ci/check.sh check` "ok: no finding",
  `validate` "ok: only the missing version"; no mention left. The orchestrator read the diff; branch T-5796 pushed to
  LastStep/bonsai-workflow, CI awaited.
- 07:34 Workflow CI on `612bc333` green (validate, check ubuntu and windows, release); its `main`
  fast-forwarded to `612bc333` and pushed, branch deleted on GitHub, worktree removed. Bonsai's half of T-5796 waits
  until T-5721 no longer reads `running`.
- 07:34 The director, on the feature document's layout: **"One page (Recommended)"**: `docs/features.md`, seven
  areas (setting up, the guard, the stop gate, the ladder, records, the log, agents and the person), each a table, then
  "Not built, and known limits", one README link. Piece E, task T-5722.
- 07:45 T-5722 (opus) done, report `M-2026-10-22-1774a8c0` (its own log `~/bonsai-checks/T-5711/run-T-5722.md`):
  `35cf7264` on T-5718's tip; `docs/features.md` 42,269 bytes, seven tables (101 rows) and "Not built, and known
  limits"; README link under "What you get". Climb read: local, green, git.sha `35cf7264`, dirty false, sha256
  `f46f30c5...ef8cb7b` (tests 4,874, tagged 2,173). Corrections from the audits: 8, among them BUG-022 looks stale
  (the warning exists, `internal/engine/checkpersonal.go:14`) and CLAUDE.md's "rung added is stricter" (for T-5710).
  Left by the builder: a test build `bonsai` in this session's scratchpad the guard will not let it delete
  (`delete-protected`, a stand-in name); not on the PATH. T-5722 `verify`.
- 08:02 T-5721 (sonnet) needs-input, `M-2026-10-22-fff85979`: three takes by runners (briefs `M-2026-10-22-d60ac340`,
  `-7319fb8b`, `-78f40332`, `-e2d7447e`, `-1dc62fc4`); take 3 showed beats 1-4 for real, stalled on a Claude Code
  permission prompt; every frame holds the test build's path with the user name; nothing committed. T-5721 `blocked`.
- 08:03 Options write-up `records/options/OPT-0002-recordings.md` (what to show, how): A terminal only, B the
  terminal GIF plus one real session filmed in the director's terminal on the installed Bonsai (recommended), C B side
  by side, D agents on a test build with boxes.
- 08:18 The director on OPT-0002: **"lets do option 1. but i also want to make sure that we record clips, and then
  we can edit and make a good video using them ..."**: B, as clips and an edited showcase. OPT-0002 `decided`, chosen B;
  plan amendment 2 (T-5721 re-scoped to the clip set; T-5723 the showcase edit).
- 08:30 T-5796 done (its own run report; `6f8783ea` on main, the plugins installed after two restarts by the
  director). T-5721 re-scoped to the clip set (amendment 2) and `running`; builder (sonnet) `M-2026-10-22-eac07f8a`.
- 09:56 T-5721 (sonnet) done, `M-2026-10-22-d7122a21`: `ee88a00c`, eleven tapes, film.sh and README in
  `docs/demos/clips/`; all eleven practised on the test build (runners `M-2026-10-22-4885fb0c`, `-4b23d21d`,
  `-80fbc6b5`), each reached its moment; practice in `~/bonsai-checks/T-5721/practice/` (user name in frames: scratch
  only). Climb green, sha256 `b781dc79...113b96`. The orchestrator looked at clips 06 and 09's sheets. **To fix at
  landing:** the branch commits a run report under `records/runs/` (outside its Owns; moved to the main checkout then)
  and its commits lack the co-author lines (rewrite at the batch's rebase). **Finding:** the ladder's "leftovers not
  started by this run" line names other programs on the machine (chrome seen in clips 03-05): a privacy point for
  rc.2's list. T-5721 `verify`.
- 09:56 T-5723 (the showcase edit, fable) written `running`, worktree from T-5721's tip; brief
  `M-2026-10-22-94ca258b`.
- 10:32 T-5723 (fable) done, `M-2026-10-22-5b26518f` (its run report in the main checkout,
  `records/runs/R-2026-10-22-T-5723-showcase-edit.md`): `8ac57e1e`, `docs/demos/edit.sh` and `showcase.txt`; climb green,
  sha256 `4016fc0a...3997d9`. Practice cut (scratch, user name in frames): first-project.gif 0.77 MB 36 s, session.gif
  3.08 MB 42 s, showcase.mp4 4.79 MB 178 s, five chapters with captions and the picture beside Claude Code. The
  orchestrator looked at the showcase's contact sheet. Its notes: clips 01-05 too small at 830 px (film at width
  1100, font 22); showcase.txt's in/out times are the practice take's. T-5723 `verify`; the cut sent to the director.
- 10:47 The director: the showcase video **"In the repo (Recommended)"**. Plan amendment 3: the batch lands at
  the final pass with the filmed media, before rc.2's tag; T-5720 superseded by clip 01 (cut on the director's word).
- 10:48 Follow-ups, side by side: T-5721 (sonnet) `M-2026-10-22-fdd33713`: clips 01-05 larger, the run report off
  its branch (copied to the main checkout); T-5722 (sonnet) `M-2026-10-22-3362d3aa`: the README's "A real session"
  points at the clip set, the edit and the showcase. Both `running` again.
- 10:51 T-5722's follow-up (sonnet) done (hand-back only): `6730e359`, "A real session" names session.gif,
  showcase.mp4, the clip set's README and edit.sh; no session.md left. Climb read: local, green, at `6730e359`, clean.
  T-5722 `verify`.
- 10:52 T-5724 (the README's dry run) written `running`; builder (sonnet) `M-2026-10-22-c64c18ae` on T-5722's
  README with the installed rc.1, a scratch folder and home; step 6's update the director's.
- 10:56 T-5721's follow-up (sonnet) done, `M-2026-10-22-cbf4d427`: `edb96b4d`, clips 01-05 at 1100x900, font 22,
  legible at 830 px (stills read); the run report off the branch. Climb read: local, green at `edb96b4d`, clean,
  sha256 `abb60c7e...205be0b7`. T-5721 `verify`.
- 10:56 Facts verifier (opus, fresh) `M-2026-10-22-6a181316`: README and features (T-5722), the SVGs (T-5719),
  the clip set (T-5721), the edit (T-5723); tasks and diffs only, no builder's report.
- 11:04 Facts verifier (opus) `M-2026-10-22-24470d00`: **FAIL**, six blocking: features.md:79 the stop gate's
  rule-word range and the block's shape (three lines, `internal/hookin/frame.go:98-100`); features.md:98 BUG-029 on the
  wrong row; showcase.txt clip 07's caption gives the wrong reason (`--allow-exec` is consent to code); edit.sh crops
  1600x900 and fails on clips 01-05 at 1100x900 (ran: ffmpeg exit 1), T-5723 based before the resize; the clips README
  names the plan. README, SVGs and every count held. Notes: film.sh's PACK_SOURCE unused by init; raw lengths disagree;
  Decide's options are "up to four"; not-linked records nothing; MultiEdit is judged by the time check; loose bug links.
  Facts-lens count: 6 blocking, 6 held, 0 not. Fix round briefed.
- 11:10 T-5722's fix round (sonnet) `M-2026-10-22-3d5528b8`: `4dea6fc5`, findings 1-3 and the notes fixed (seven
  loose bug links dropped); climb green, sha256 `e35e0d09...d9ef781`. One gap it named (row 134) briefed back,
  `M-2026-10-22-34b9433b`.
- 11:13 T-5721 and T-5723's fix round (sonnet) `M-2026-10-22-6daec507`: T-5721 `3aecf419` (plan name, PACK_SOURCE
  used by init, "four lines", one set of raw lengths), T-5723 `0c5f7265` rebased on it (only edit.sh and showcase.txt;
  clip 07's caption; edit.sh by each clip's own size). Both climbs green. edit.sh re-run on the practice clips: all three
  outputs; first-project.gif 1.11 MiB, over the plan's 1 MB aim: its colours or fps go down at the final pass, where the
  clips change anyway. Both `verify`.
- 11:14 T-5722 row 134 fixed, `M-2026-10-22-8baceefa`: `386e4cb0`, climb green. T-5722 `verify`. Verifier round 2 briefed.
- 11:16 Verifier round 2 `M-2026-10-22-65f93a0a`: **PASS**, all six fixed and true (its own ffmpeg run on a
  1100x900 clip through edit.sh's crop exited 0); no new wrong claim. Notes for the final pass: first-project.gif under
  1 MB; clip 01 types `$PACK_SOURCE` and `$CLIP_REF` on camera (show the URL and commit); BUG-017/028/082 still loose
  on the Formats row. Facts-lens total: 6 blocking, 6 held, 0 not.
- 11:24 The director: "once done, push everything to main, from my side its approved"; then "i want you to take
  ownership of this task here of building the demo, and see it to completion". Amendment 4: film now on rc.1, land
  with the films in; T-5720 `cut`. Final-pass prep briefed (clip 01 shows the URL and commit; first GIF under 1 MB).
- 11:36 Prep done, `M-2026-10-22-adba56c8`: T-5721 `17c0b30c` (film.sh writes the real URL and commit into each
  tape's copy; out folder default ~/bonsai-clips; last line "Send back this folder"; every typed command in rc.1's help),
  T-5723 `cecce558` (gif line 830 20 7: first-project.gif 954,269 bytes; the picture in the showcase). Both climbs
  green. Both `verify`. Waiting on the dry run's step 6.
- 11:40 T-5724 (sonnet) needs-input, `M-2026-10-22-fcea63c4` (full report `~/bonsai-checks/T-5724/report.txt`):
  steps 1-3 as the README says; step 4 passed after harness fixes (claude-here's plugin cache and home differ from a
  plain session's: not README faults); step 5's headless session had no permission to write (the README's own line);
  step 6 refused `config-looser` as written. README snags: the trust screen defaults to "No, exit" (say choose Yes);
  no task content given; check's Notes block unmentioned; tables-lag warning not under step 6. The orchestrator's call:
  step 5's stand-in (README.md written by hand, marked a deviation). The director's update lines and the film line
  sent together.
- 11:44 The director ran the dry run's step-6 update (its home's record changed at 11:40 UTC) and started
  film.sh in their terminal ("currently we are here. can you watch the folder which has this, and then resume your
  process yourself once done"). A background wait on film.sh's end; the dry run resumed, `M-2026-10-22-5d518adb`.
- 11:47 T-5724 (sonnet) `M-2026-10-22-720f8704` (and its hand-back's correction: its own script called
  `bonsai now`, which rc.1 lacks, so its scratch run report was misnamed; the dry run's slip, not the README's). The
  climb **green** as the README says (`GREEN: 1 of 1 rungs green`). The stop-gate part **not seen**: claude-here forces
  `~/bonsai-checks/home`, not the scratch home the director's update recorded, so every Stop was `main-not-verified`
  (9 and 10 blocks; Claude Code's cap of eight, then the turn ends). Not worked round. The stop gate's hold and release
  on the installed Bonsai is proved instead by the director's clip 09. README snags to fix: the trust screen's default
  ("Yes, I trust this folder"); a task file example in step 6; `check --write` under step 6. Fixes after the filming
  (no climbs while it runs).
- 12:17 The director: "film is done". All eleven clips in ~/bonsai-clips (01-05 1100x900, 06-11 1600x900; 09 is
  181 s). The orchestrator's own background wait (`while pgrep -f film.sh`) had matched itself (STATE's known trap)
  and was stopped. Final cut and the README's dry-run fixes briefed side by side.
- 12:20 T-5722's dry-run fixes (sonnet, hand-back only): `e98b6abf`, README only (+32 -5): the Notes block, the
  trust screen's "Yes, I trust this folder", a minimal task file (checked: no finding in a scratch copy of the dry
  run's project), `check --write` under step 6. Climb read: local, green at `e98b6abf`, clean. T-5722 `verify`.
- 12:36 T-5723 final cut (sonnet) `M-2026-10-22-8c70bfe8`: `fc0afc6b`, climb green; ten of eleven clips reached their
  moment, no private detail in a used frame (paths /tmp/bonsai-film.*; the plan-name box fixed to follow the mascot's
  feet after it missed in clip 11's scrolled frames); outputs first-project.gif 958,828 B, session.gif 3,100,626 B,
  showcase.mp4 5,056,448 B 171 s. Clip 10 stalled on a Claude Code permission dialog before Claude read the answer.
  The orchestrator's call (the director's "take ownership"): drop scene 10 (a harness fault on screen for 30 s), keep
  the tape with a note; Claude-Session links in 09 and 11 kept (public in this repo's commits). Re-cut briefed,
  `M-2026-10-22-4981c172`.
- 12:48 T-5723 re-cut (sonnet) `M-2026-10-22-7cb851ed`: scene 10 out; showcase.mp4 4,498,742 B, 147 s; the GIFs
  unchanged. The orchestrator read a contact sheet of the showcase and clip 05's last frame (the caption "two climbs,
  the ask and its answer" is what is on screen).
- 12:48 **The batch.** Rebased in turn onto main `6f8783ea`: T-5719 → T-5722 (holds T-5718) → T-5723 (holds
  T-5721), every commit given the session's co-author lines (`~/bonsai-checks/scripts/T-5711-trailer.sh`, by
  `rebase --exec`): 22 commits, only README.md and docs/. Climbed at each tip with /usr/local/bin/bonsai, each green,
  local, clean, its log record naming /usr/local/bin/bonsai with its sha256:
  T-5719 `2da9c00a` sha256 `a21a3d9c...`; T-5722 `4fd00789` `6ff74977...`; T-5723 `57fa4f97` `3913d5b0...`; rungs 0-5
  green, about 4.4, 0.8, 0.4, 59, 53, 0.0 s; go_tests 4874, go_tests_tagged 2173. T-5723 pushed; CI awaited.
- 13:18 CI on `57fa4f97` (branch T-5723): nine jobs green, the two windows jobs among them. `main` fast-forwarded to
  `57fa4f97` and pushed; branch T-5723 deleted on GitHub. Landed: T-5719 (`2da9c00a`), T-5718 and T-5722 (to
  `4fd00789`), T-5721 and T-5723 (to `57fa4f97`). The orchestrator's read of the batch's diff against the pieces' Owns:
  README.md and docs/ only (27 files, +1942 -70), nothing under records/, design/ or .bonsai/. Tasks T-5711, T-5718,
  T-5719, T-5721, T-5722, T-5723, T-5724 `done` (T-5711 a full lane: closed on the director's "from my side its
  approved" and "see it to completion"); T-5720 `cut` (amendment 4). The pieces T-5718, T-5719, T-5722 and T-5724 kept
  no run report of their own: this report is their log. New in the bug register: BUG-097, BUG-098; BUG-022 marked
  maybe stale. rc.2's list (T-5710) item 12.

## Models

Orchestrator Opus. Builders: Opus (T-5718, T-5722's first draft), Fable (T-5719, T-5723's first cut), Sonnet (T-5720,
T-5721, the fix rounds, T-5723's final cut, T-5724, T-5796's pack change). Auditors Sonnet (six). Verifiers Opus (the
plan's review, two rounds; the facts check, two rounds). Runners Sonnet (the clip takes).

## What is left

- After rc.2's fixes land: README step 6's "(rc.2 puts that first)" line; film again what changed on screen (T-5710
  item 12), clip 10 among them.
- The director's YouTube video with a voice-over (later, theirs); `docs/demos/clips/` and `edit.sh` are its material.

## Processes left running

None of this session's (checked with `ps -eo pid,etime,cmd` at the end; the ttyd on port 7681 is not ours).
