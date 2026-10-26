---
format: bonsai.state/1   # fields: bonsai check --schema bonsai.state; how to fill: skill base:state
updated: 2026-10-26
updated_by: orchestrator (opus)
labels: {}
---

# Bonsai: where it stands

Rewritten, never appended (`base:state`). Last rewritten 26 Oct 2026 (a day session, the director present).

## Now (26 Oct), two sessions side by side

- **Session 1 (the LFS session; it holds T-6002, T-6003, T-6005).** The studio's B-053: a Git LFS checkout's climbs
  read `dirty: true` (BUG-054). **T-6002** fixes it: an LFS file reads clean when its bytes hash to the SHA-256 and
  size HEAD's committed pointer names (`internal/vergit/lfs.go`; rung 0 the same), the attribute from HEAD only, the
  executable bit checked; plan `records/plans/P-T-6002.md` (reviewed, approved), verifier PASS
  `M-2026-10-26-91a81fe9`. **T-6003** clubs BUG-091, 112, 113, 114's half, 022. **Both done, landed 11:15** at
  `df549bbd` (CI green, every job), `main` at `c91891aa` and after (the other session's two records commits replayed
  on top, content unchanged). A 1.0.1 release is the director's to decide.
  **T-6005** (running, sonnet builder, branch `t-6005-link-fixes-small`, green at `7051a440`, rebasing onto `main`): BUG-119, 120, 125, 129, 127
  (the director's option 3 of the split). Run reports `records/runs/R-2026-10-26-T-6002-lfs-dirty.md`,
  `R-2026-10-26-T-6003-small-fixes.md`.
- **Session 2 (fresh, to start now): `T-6004`**, the rest of the first link's bugs: BUG-121 to BUG-124, BUG-126,
  BUG-128, BUG-130. Start with group B's options (BUG-124 rung 0 on the base branch; BUG-126 a result per climb, a
  format change; BUG-121 never_edit and reads; BUG-123 future times), one question at a time to the director, no code
  needed; then group C (BUG-130, BUG-122: guard and link, a fresh Opus verifier); then BUG-128 (the workflow pack).
  `main` now holds T-6002/T-6003 (`internal/rung0`, `internal/ladder`): build on it. Keep out of T-6005's files (`internal/engine/checkdocs.go`, `internal/message/pointer.go`, base's
  message and memory skills, the workflow repo's orchestrate and verification-ladder skills).
- **Rules for two sessions here:** every task has empty `bonsai.allows`, so several `running` tasks grant nothing
  either way; always climb with `--task`. Before any push, read `origin/main`'s new commits; never stage a file the
  session did not write; each session lands its own branches; records commits only for its own files. The machine is
  loaded: timing tests (TestStatusLineSpeed, TestCleaning) went red under load on 26 Oct and passed alone.

**Bonsai 1.0 is released, and step 5 (the full 1.0) is done** (25 Oct). Every part of the plan is done: part 0, the
skeleton's six parts, and steps 5.1 to 5.7, each closed by its end verifier.

- **1.0** (`T-5739`): the director's word ("lets go"); the date commit `48d5f3f0` (CHANGELOG's heading only, the notes
  byte-identical to rc.3's, which is 1.0's code); their `v1.0.0` tag and approval; release run 40172190980, **published
  07:37:52 UTC**, a full release, GitHub's latest; read back on both sides (all six programs rebuilt byte-identical).
  Installed by the director: WSL `/usr/local/bin/bonsai` `f9b1cd76...f5caf`, Windows `C:\Program Files\Bonsai\bonsai.exe`
  `d1bba8dd...0df3e0`, both `bonsai 1.0.0 (commit 48d5f3f0d387)`. The way back: rc.3 in
  `~/bonsai-checks/prerelease/previous/`, rc.2 in `older-rc2/`. Their memory line in `~/.claude/CLAUDE.md`; `/memory`
  lists `~/.bonsai/personal/INDEX.md`.
- **The packs at 1.0** (`T-5706`, 5.7.6): the template's and the workflow repo's CI take `bonsai` from the 1.0 archive
  by its SHA-256 (base's `ci` skill the same); `base` and `workflow` at `1.0.0`; landed `4662da5e`; the workflow repo
  `0290f701`. The director's tags: `base-v1.0.0` at `4662da5e`, the workflow repo's `v1.0.0` at `0290f701` (its release
  made by its own job). Bonsai's repo still names both packs by commit (moving to the tags waits for their word).
- **The floors** (`T-5740`): `go_tests` **5067**, `go_tests_tagged` **2341**, climbed by the installed 1.0 (`97a04e18`).
- **5.7's end verifier** (`T-5741`, `M-2026-10-25-f289b34c`, fresh Opus): every check on the code, the release, the
  installs and the packs PASS, Windows natively included (test 5001 passed in 34 packages, tagged 2335 in 7); FAIL on
  three records findings: F3 (STATE's "being rebuilt") fixed; F1 (5.7 about 48 h, past its re-ask line of 24) on the
  director's word "Accept and close"; F2 (the rc.1 trial's task-file moves and two stages not reached) on "Stand, with
  a note" (the note in `records/trial-rc1.md`). The re-check skipped on the director's word ("skip the verifier").
- **The road there in 5.7:** rc.1 and the trial on Bonsai (`records/trial-rc1.md`), one system for time (`T-5795`),
  the README and demos (`T-5711`), rc.2 and its fixes (`T-5710`), the director's trial on rc.2 (`records/trial-rc2.md`),
  the workflow split (`T-5732`), rc.3 and its test (`T-5734`, `T-5738`), the pi research and contract (`T-5797`,
  `T-5798`; `design/agent-neutral.md`, `design/pi-contract.md`).

**No person's name** goes into any file or commit message (the director, 15 Oct; `CLAUDE.md`'s Safety section): "the
director" in this project's files, "the person" in the product. Files already under `records/` and `design/` keep
what they hold, on the director's word.

## In short

Bonsai 1.0 (released 25 Oct) is one small Go program that gives every project the same formats, packs, guards,
recorder and proof ladder (`design/one-pager.md`). The plan was approved 8 Oct. Part 0 and all six parts of the walking skeleton
are done (gate report `records/gate-skeleton.md`). At the gate (9 Oct) the director chose path (a), the full Bonsai
1.0, and approved the plan for step 5, `design/plan-5.md`; every part's section, 5.1 to 5.7, is written and approved.

- **Part 0, the formats** (`3770d04`): a JSON Schema for each format, examples, trick files, a raw-byte manifest and
  its Go test. Now **set 9** (`0377c20`, 15 Oct: set 8's descriptions without a person's name; no new format).
- **The skeleton, parts 1-6** (`17f2938` to `5633721`). The test pack `LastStep/bonsai-test-pack` is public, `main` at H
  (`19e3ca5`).
- **Step 5.1, formats and engine to 1.0: done** (code `11d871c`; verified on `5fa9bd8`).
- **Step 5.2, the recorder, logs and asks: done** (code at `8243800`), the redactor with known limits (below).
- **Step 5.3, the guards: done** (code at `750ca07`).
- **Step 5.4, the ladder runner and the switch: done** (13 Oct). The pre-release installed at `/usr/local/bin/bonsai`
  (binary at `a88436a`), the link `c1bb21f`, the floors `6f25dd6` (`go_tests` 4188, `go_tests_tagged` 4194).
- **Step 5.5, the packs, the roles and their messages: done** (16 Oct). `base` in Bonsai's repo (`packs/base`) and
  `workflow` public at `LastStep/bonsai-workflow` (`ef5b5ae`); both linked in Bonsai's repo (`16fb557b`), both plugins installed;
  the pre-release at `bf7cc3d` installed (sha256 `70516eb8...a2478`; the way back in `~/bonsai-checks/prerelease/previous/`);
  the floors `92de82c8`; the end verifier `T-5522` (report `~/bonsai-checks/verify-5.5-end/report.md` and `report-r2.md`).
  Every 5.5 task `T-5500`-`T-5522` and `T-5590`-`T-5599` is `done`; run reports `records/runs/R-2026-10-1[456]-T-55*.md`.
- **Step 5.6, the machine pieces: done** (19 Oct). Settings and labels, personal memory, the status line, the release
  read, the installers, the CLI look, the operating split, `workflow`'s changelog, prototype and correct skills; the
  end verifier `T-5611` (FAIL on records), then `T-5616`'s re-check. Tasks `T-5600`-`T-5617`, `T-5690`-`T-5699`; run
  reports `records/runs/R-2026-10-1[689]-T-56*.md`.
- **Step 5.7, the release: done** (19-25 Oct). 5.7.0-5.7.6; rc.1, rc.2 and rc.3, each published, checked and
  installed; the trial on Bonsai and the director's trial on rc.2; **1.0 published 25 Oct**; the packs tagged at 1.0;
  the floors; the end verifier `T-5741`. Run reports `records/runs/R-2026-10-{19..25}-T-57*.md`.

**Hours** (each run's minutes from the run reports): part 0 45 minutes; the skeleton 6 h 18 min; 5.1 866 minutes; 5.2
1,061; 5.3 about 2,013 (33 h 33 min); 5.4 about 2,372 (39 h 32 min); **5.5 about 78 h** with its planning and reviews (the end
verifier's count, under its re-ask line of 138). Questions put to the director in 5.5: twenty-three before 16 Oct, and on 16 Oct
eight (the workflow limits, the walls' Windows half, the login check, the history, check 5, check 5 on Windows, and the
morning's install and Windows steps as asks). **5.6 about 21 h** (T-5611's count of about 19 h, a floor since many rows
were not written, plus about 2 h on 19 Oct's day session), under its re-ask line of 51. **5.7 about 48 h**
wall-clock (`.bonsai/sessions.md`, minutes any session or subagent was active, 19 Oct to 25 Oct 11:44 UTC, less 5.6's
rows: 2,882; T-5741's count), **past its re-ask line of 24 with the director's word** (25 Oct, "Accept and close"):
three pre-releases, two trials and the director's added pieces (the README and demos, the time system, the pi
research, the workflow split); the plan's figure was 9.5-18.75 h. **Step 5's Windows-only tally:** 5.5's "about 143 minutes plus 16 Oct's Windows check-10 helpers (about 6 h, side
by side)", plus **5.6's about 187 minutes** (207 if T-5696's run `09b28f0b` was a second run; every row named in
`R-2026-10-19-T-5615`); **5.7's about 130 minutes** (rc.1 about 20, rc.2 about 40, rc.3 about 40, 1.0 about 5, the end verifier's about 25;
`R-2026-10-25-T-5741`).

## The director's decisions

- **22 Oct (a day session):** "okay lets go ahead with the readme and demo"; the README **restructured**; demos of
  **both kinds**; media in the repo, **"option 1 now, and then i want to record my screen and do a voice over as a
  youtube video later"**; the session's story "a task, start to done"; the picture a drawn SVG; then **"i also want an
  audit of all bonsai features, and a specific document on github which is linked on the readme ... with proper
  transperance"** (one page, `docs/features.md`); on how the recordings are made (OPT-0002): **"lets do option 1. but i
  also want to make sure that we record clips, and then we can edit and make a good video using them"**; the video
  **in the repo**; **"remove the trinetra paths from the workflow pack. also close any unnessary monitors"** (and "Yes,
  you do it" to moving this repo's pin); **"once done, push everything to main, from my side its approved"**; **"i want
  you to take ownership of this task here of building the demo, and see it to completion"** (the orchestrator filmed
  nothing itself: the director ran `film.sh` and the dry run's update in their own terminal); "once all done, wrap up
  cleanly". Their answers in `~/bonsai-checks/T-5711/notes.md`.
- **20 Oct:** **"in a previous conversation i discussed about converting bonsai for team development along with the
  studio, move that research item in the studio research folder, as we wont be focusing on that here"**; undo the
  commits of the files it made; **"then we just fully focus on the current planned bonsai release order"**; the aim
  for the session, **"finish 5.7"**. The research, its task and run report, and OPT-0001 now live in the studio's
  repo, in its research folder (solo-to-team-ai, 19 Oct); the freeze review it proposed is not a 5.7 piece. Then
  the scratch folder: on the first audit's options, **"we need a proper structure which decides what a scratch folder
  is and what type of authority it can contain. it should be declared in bonsai.yaml ... bonsai local maybe?"**; asked
  when, **"okay we will tackle this idea after 1.0"** (`T-5793`, `todo`; facts and a draft shape kept).
  Later: **"add a step before 1.0 release, i want to improve the readme, and create some demo videos showcasing how
  bonsai works (and does for a project)"**; the demos' form: **"we will discuss it thorughly when the time comes"**
  (`T-5711`, before rc.2). Go for rc.1: **"Stay on 1.27.2"** (go1.27.3 breaks govulncheck; its fixes are to
  crypto/x509, which bonsai does not link; checked again before rc.2 and 1.0). Batch 1 done (read back), batch 2 done:
  v1.0.0-rc.1 published from 73919ddc.
  At the session's end, two pointers for rc.2's list (`T-5710`, items 10-11): **"bonsai agents command shows all agents
  at once, need some filter there"**; **"bonsai answer needs to have an options key, like where we can answer just A,
  B etc; currently we use words"**.
- **19 Oct (afternoon):**
  - On "show me the status board for 5.7": **"no i dont want a page, just show me a table here"** (status boards as a
    table in the session, not an artifact).
  - `workflow:correct`'s three questions (saved in `~/bonsai-checks/5.7-prep/decisions-2026-10-19.md`): (1) Windows
    before a landing, **"C: CI on branch"**: the batch's branch pushed, CI (its windows jobs among them) green on it,
    then `main` fast-forwarded; needs `ci.yml` to run on pushed branches and `CLAUDE.md`'s landing rule reworded, a
    task before 5.7's first landing; (2) files outside a piece's Owns, **"B: strict, stop and ask"**: a builder stops
    with `needs-input` and the orchestrator amends Owns in the run report first; (3) a scratch folder for the roles that
    write nothing, A, with **"if we go with option A, then we also need some sort of new agent like a cleaner"**, then
    **"1: A now, cleaner after 1.0"**: write-nothing roles may write, clone and build in a scratch folder outside every
    linked project, named in the brief; a briefed runner cleans at each batch's and the part's end; a cleaner role is
    the first piece after 1.0. The orchestrator first said the guard needs no change for A; the guard refused an
    auditor's writes there, so A needs a guard change (an Opus builder and a verifier).
  - The records-database talk (work files when many people's AI agents share a project; the studio as the team
    layer; an enterprise product): researched as `T-5791` and moved to the studio on 20 Oct (above).
- **19 Oct (day session):** "we will finish up 5.6 in this session"; step 7 asked about ("why do we need the
  sandbox?"): Windows 11 Home has no Sandbox, so **7a-7d into CI's throwaway runner, 7e (declining the admin prompt) at
  their real install in 5.7** ("yes, go ahead"); ran steps 4, 5, 10 themselves; "update the roadmap artifact" (version
  39); on CI's windows job **"Both 1 and 2", then "actually dont turn off defender scanning"**: the parallel jobs only. The cleanup: "option 1" (done). On the tagged pass: "yes look into that so we don't
  do same work twice", then **"Yes, narrow it"** (`T-5790`; they ran `bonsai update` to accept it).
- **19 Oct (overnight, then asleep):** "i opened this session to be running overnight"; on 5.7 tonight, **"5.7.0 lands,
  5.7.4 drafted"** (5.7.0 built and landed after 5.6's work; 5.7.4 on its branch until the records-database talk and
  5.6's close); "also check for any other known bugs, and if they can be clubbed"; **one bugs register** ("record all
  the known bugs in that one place, and clear them up from other contexts ... properly linked"): `records/bugs.md`,
  STATE's lists replaced by a pointer, a row in CLAUDE.md; "fix the bugs which you can rightfully tonight, no need to
  overdo it"; **"make sure to check for this every time an subagent messages you"** (leftover shells, monitors and
  processes closed after every agent's report).
- **18 Oct:** resumed the session after the plugins' install; **"make sure that each agent first does an audit of whats
  actually in the codebase, and not fully trust the state, as previous runs were stopped midway"** (every brief now
  opens with an audit report and a stop; `~/bonsai-checks/briefs/common-2026-10-18.md`); **"be extra careful with any
  windows related work ... if you think there are a lot of things going on already, then wait for some to finish before
  queing something next"** (Windows runs one at a time, by the orchestrator only); **"close any unnessary proccesses
  running"**; then, at 15:00, **"we will consider that test as passed and move on to next things. now can we club windows
  checks later in the end and try to finish up the builds and wsl checks for 5.6"**: no Windows run per landing in 5.6;
  one combined Windows check at 5.6's end (every piece's Windows steps listed in its run report); `TestStrandedMoved`'s
  Windows failure counted as passed (its cause since found and fixed in `84533b07`; GitHub CI's windows job is its proof).

  Later: "you can now queue up parallel work if possible" (V1 started beside 5.6.5); the cleanup **"Parts + caches,
  after 5.6"** (the 5.1-5.5 scratch folders on both sides, ~27 GB, the 5.3.7 worktree by `git worktree remove`, and both
  Go build caches, ~28 GB; the rest kept until they look); the end Windows run **"Not today"**. The orchestrator's calls:
  5.6.5's published-only check not built (Go's `net` would break 5.3's `TestTheBinaryLinksNoResolver`): a known risk,
  **to raise with the director at 5.7** (plan-5.md:8231's hand-off).
- **16 Oct:** overnight run on their word; Cursor's pstack surveyed, to discuss after 5.5 and before 5.6 builds; the
  Windows trust sitting; installed the pre-release (09:28); **`workflow` added** with builder minutes 240 by default and 480
  at most ("option 1"; limits are lines in `bonsai.yaml`, a set command noted for 5.6-5.7); **the walls' Windows half on
  5.5.6's rows**; **the Windows login size check dropped** ("Skip the check": agents never touch the credentials file; the
  warm-up stays); ran `bonsai update --allow-exec --yes` and the plugin installs; one history item left as is (their word);
  check 5 run by a builder, its Windows half on payload replays plus 5.5.5's Windows sessions; **the human CLI output to be
  improved**. **Then, in the 5.6 talk (`T-5690`, on a private artifact "Bonsai 5.6 Ideas Board"; decisions in
  `~/bonsai-checks/5.6-prep/decisions-2026-10-16.md`):** from pstack, `workflow:correct`, observe-don't-ask with
  `workflow:prototype`, and four short skill lines into 5.6; **playbooks after 1.0** ("playbook we will consider after
  1.0 is done"), hillclimb and pickup with them; **a changelog skill, a skill only** ("i dont want a command line
  feature, just a normal skill"), in `workflow`, in 5.6; **`operating-bonsai` split in two**; **the CLI look prototyped
  first, under the new skill, shown as images** ("i dont need to test it myself"; "the prototype can happen as the first
  step"); the library, colour and the redesign's place asked after they see the images.
- **15 Oct, this session:**
  - About 10:45: "for now just finish all the steps which are at verify". Asked whether to split V1 so 5.5.0-5.5.2
    close today (split recommended): **"Keep the plan"**: one V1 after 5.5.6.
  - About 10:55: made `LastStep/bonsai-workflow` public themselves; its `main` was the commit P passed. Then: "this looks fine. but ... my name was used in some documents ... make sure that my name is no where
    present. use some generic desigation ... this is a general product which will be used by everyone. also make sure
    my name does not come up in bonsai repo as well"; "send a sonnet agent to check for that, and then read its
    message and tell it to remove accordingly". On the history (asked): **files only** ("i dont want my name within
    the files"): commit author lines and history stay. Then **"don't rewrite the files in the records folder, keep
    them as is"**, **"also don't remove from already made plan files"**, and on the four approved design files
    **"option C and also claude.md and state.md"**: the product (code, formats, docs, packs, LICENSE, CONTRIBUTING, the
    release workflow's comment), `CLAUDE.md`, `STATE.md` and `workflow` lose the name; `records/` and `design/` keep
    it. The orchestrator's calls: "the director" and "the person"; LICENSE lines "The Bonsai authors" and "The
    bonsai-workflow authors"; formats set 9 for three descriptions.
  - About 11:45: **`records/` stays in git "for now"** (asked whether it should be local; the take: tasks hold the
    guard's grants, the stop gate reads a named session's run report from its checkout, the contract says committed).
    At step 5.7 the director's records-database idea was raised first: researched on 19 Oct as `T-5791`, moved to the studio on 20 Oct.
  - About 12:10: "are we good to start 5.5.3 in parallel? if so, then go".
  - About 13:05: VG round 2 found M1; asked (fix and a short third check, recommended; fix without one; a known
    limit): **"A: fix + short 3rd check"**. VG3's must-fix (S2's other-system place) was fixed and re-checked by the
    same verifier under that choice.
  - 10:44: the size-check rule added (`/permissions`); about 11:00: logged in on Windows again (the warm-up had found
    the login lapsed: the credentials file read 363 bytes; 524 after the login).
- **15 Oct, earlier:** "land in batches" (`CLAUDE.md`'s landing rule, `T-5597`); "wrap up after current agents are
  done"; "close any unnecessary process left running"; Windows sessions: "Allow the size check"; "im leaving the
  session running for the night. work autonomosly".
- **14 Oct:** "take a pause once 5.5.9 and 5.5.8 are done"; Defender, "CI counts for those" (Defender quarantines
  `cmd/bonsai`'s test-built `bonsai.exe`, a false positive; CI's `windows` job is their proof, the gap written on every
  landing); 5.5's re-planned section approved (63.5-106 h, re-ask 138); 5.5 reshaped 11:13-14:26 (six roles: leads
  and shared helpers; the message format; generous stall budgets; `bonsai agents` with raw counts only; messages local
  for 30 days; the guard as two lines where roles are declared; every plain session the orchestrator; Bonsai's own
  team links `workflow` at 5.5's end).
- **13 Oct:** D1 fixed now; stall stops for subagents into 5.5's redesign; the director ran the switch's steps
  themselves; K2 a known limit after 1.0 (CI to run the ladder; a task done only with it green); K1 fixed now (looser
  committed settings refused until a person re-records them); Claude Code's memory stays on until 5.6 ("A. Keep them
  on"); build several pieces, then one verification pass where efficient.
- **12 Oct:** past records left as they are; builders stay subagents ("A. Keep the plan"); 5.3 closes on two known
  limits (CDPATH and `cdable_vars`; Git Bash's `/proc/cygdrive/`); Sonnet helpers for mechanical runs by default.
- **11 Oct:** verifiers stay adversarial, at most two rounds a piece, more only if necessary; no long-lived token for
  Windows sessions; the guard's fix loops' finish lines; the delete check's refusals as planned; how agents work
  (helpers one level down; a fresh agent past about 500k tokens; designs shown as diagrams).
- **10 Oct:** the redactor's finish line; 5.6's section approved with six decisions; from then on a section comes as
  one question per decision, then the approval.
- **9 Oct:** 5.5's section approved without the pin rule (a pack's version is the person's step); the gate: path (a);
  Bonsai has no screens of its own (a pure command-line tool; every visual the studio's); the 7 Oct scope split stands;
  no self-update; Bonsai is managed by AI agents inside projects ("top notch" support); 5.3's (a) and (ii); a pack
  plugin that runs code asked for on each machine; later sections reach the director under (B); the studio links at
  step 7, after 5.5.
- **8 Oct:** Bonsai's design and records live in this repo; the schemas and trick files are Bonsai's; Bonsai first;
  "the proper way, no shortcuts"; Bonsai joins the studio's dashboard at step 6; Haiku for bookkeeping; the plan and
  one-pager approved; hand checks by a Sonnet agent where an agent can; the roadmap is the artifact "Trinetra Roadmap"
  (https://claude.ai/artifact/XXKTi6geneu1pdQ4h4miw2), Bonsai's cards updated after each part.

## GitHub

- `LastStep/Bonsai`: `main` the only branch of ours (a piece's branch lives on GitHub only until it lands; a
  dependabot branch is GitHub's). At `57fa4f97` and the records commit after it, CI green. Tags: `v1.0.0-rc.1` (on `73919ddc`; a published pre-release, immutable) and
  the old `v0.4.3`. `release.yml` **active**, on version tags only, publishing inside the `release` environment once the
  director approves; immutable releases on; rulesets `main-protection` and `pack-tags` (base-v*); no secret.
- `LastStep/bonsai-workflow`: the ruleset `release-tags` (v*) and immutable releases on (batch 1).
- `LastStep/bonsai-workflow`: **public**, `main` at `612bc333` (no studio walls, 22 Oct), CI green on both sides; this repo pins it there.
- `LastStep/bonsai-test-pack`: public, `main` at H (`19e3ca5`).
- The old website stays on GitHub Pages until the director turns Pages off.

## Step 5.7, the release (done)

| Piece | Task | State | Where |
|---|---|---|---|
| 5.6, every piece, its end verifier and records | `T-5600`-`T-5617`, `T-5690`-`T-5699` | done | CI green at `ba2f4c5e`; `T-5616`'s re-check passed |
| The tagged pass narrowed | `T-5790` | done | `e85aa6e6`; rungs 1 and 4 and CI's tagged lines name the 7 packages; windows (fault tag) 16 min |
| 5.7.0, CI's own checks | `T-5700` | done | `a82d18b5`; the pin test's one exception `release.yml` (5.7.2's, with the director) |
| CI on the branch | `T-5792` | done | `aa866b3a`; the first landing by it |
| 5.7.3, the signed check in the install lines | `T-5703` | done | `11943f30`; seven lines a side, `docs/install.md` |
| 5.7.4, the words for 1.0 | `T-5704` | done | `4ebee8cc` (batch with 5.7.3) |
| 5.7.1, the release build | `T-5701` | done | `137e419f`; GoReleaser v2.18.3 by `go run`; `release-check` on every push |
| 5.7.2, the release workflow | `T-5702` | done | `4ec5c432`; V1 failed b93dd719 (the tag check passed an unreadable answer), fixed, V1 again PASS |
| CI's tagged pass narrowed; Go 1.27.3 | `T-5790`; `T-5794` | done; cut | the director: "Stay on 1.27.2" for rc.1 |
| Batch 1, the settings; batch 2, rc.1; every release checked | `T-5705` | done | read back; rc.1 run 39410125869; V2 PASS; every pre-release and 1.0 checked on both sides |
| 5.7.7, the trial on rc.1 | `T-5707` | done | climb of main by rc.1 green; proposals; the pick B |
| 5.7.8, the trial's feature | `T-5708` | done | `7afb9faa`; verifier PASS |
| 5.7.9, the trial's analysis | `T-5709` | done | `578a8724`; `records/trial-rc1.md`; batch 4: "looks good overall" |
| **One system for time** (the trial's rough 11; the director: "option 3", before 1.0) | `T-5795` | done | `T-5712`, `T-5713` (`97be0d99`), `T-5714`, `T-5716` (`f0ba50c9`), `T-5717` (`1b2c67f8`) landed; `T-5715` cut (UTC-only: the sessions table already UTC); verifiers PASS; the director's `bonsai update --allow-exec --yes` after rc.2 is installed |
| rc.2's fixes and rc.2 | `T-5710` (`T-5725`-`T-5730`) | done | landed `1cbd0aa4`; rc.2 published, checked both sides, installed in WSL 23 Oct; check 10's Windows half on rc.2 green |
| The public files' audit; the cleanup | `T-5799`; in `T-5730` | done | `assets/demos/` and CHANGELOG's old sections out; design/ kept |
| The stop gate reads the run report as check does (the pi brief's B3) | `T-5731` | done | landed `06ed179c`, verifier FAIL then PASS; for rc.3 |
| The README improved and demo videos (the director's added step) | `T-5711` (`T-5718`-`T-5724`) | done | landed `57fa4f97`; `T-5720` cut; facts verifier PASS (round 2); the director filmed the clips (`~/bonsai-clips`) |
| The workflow pack names no studio's files | `T-5796` | done | pack `612bc333`, pin `6f8783ea` |
| **Bonsai with any agent, pi first: the research** (the director, 22 Oct) | `T-5797` | done | `design/agent-neutral.md` (verifiers PASS, two rounds); walked through with the director 23 Oct, C1-C9 answered (`R-2026-10-22-T-5797`, 23 Oct lines) |
| **Bonsai under pi: the contract plan** (the director, 23 Oct) | `T-5798` | done | `design/pi-contract.md`: the pi side (one plain-JS extension in the director's `~/Servers/pi-harness`, a stopgap mode against today's binary, an envelope mode `bonsai.hook/1` later); four review rounds (FAIL, FAIL, FAIL, PASS), approved by the director 23 Oct; theirs to hand to pi |
| rc.3's fixes; rc.3 tested | `T-5734` (`T-5735`-`T-5737`); `T-5738` | done | rc.3 at `474ae7f7`; every check green on both sides, check 10's Windows half included |
| **Bonsai 1.0** (batch 5 to the memory line) | `T-5739` | done | `v1.0.0` at `48d5f3f0`, published 25 Oct 07:37 UTC; read back both sides; installed both sides; memory line loads |
| 5.7.6, the packs at 1.0; the pack tags | `T-5706` | done | landed `4662da5e`; workflow `0290f701`; tags `base-v1.0.0`, workflow `v1.0.0` (its release made) |
| The floors | `T-5740` | done | `97a04e18`: 5067 and 2341 |
| 5.7's end verifier | `T-5741` | done | FAIL on F1-F3 (records); F3 fixed, F1 and F2 on the director's words; re-check skipped on their word |
| Scratch folders declared in bonsai.yaml | `T-5793` | todo | after 1.0 (the director); `~/bonsai-checks/T-5793/notes.md` |
| **The first existing project linked: ZenGarden-AI-API** (the director, 26 Oct, "A, AI-API first") | `T-6001` | done | linked `d6355c7` in that project, old agent system retired, one fix landed locally (not pushed, the director's); `records/trial-zengarden-ai-api.md`; issues BUG-119 to BUG-130 |
| Git LFS checkouts climb clean (the studio's B-053, BUG-054) | `T-6002` | done | verifier PASS; landed `df549bbd` with T-6003, CI green |
| Five small open bugs clubbed with it (BUG-091, 112, 113, 114's half, 022) | `T-6003` | done | in T-6002's batch |
| The first link's small fixes (BUG-119, 120, 125, 129, 127), split from T-6004 | `T-6005` | running | session 1; the director's option 3 |
| Fix what that link found: the rest (BUG-121 to 124, 126, 128, 130) | `T-6004` | todo | **for a fresh session, now**, beside session 1 (see "Now (26 Oct)") |

**Now: after 1.0.** Next for a fresh session: `T-6004` (see "Now (26 Oct)" at the top). ZEN's link waits on the director's word, with the lessons in `records/trial-zengarden-ai-api.md`. What else waits, each on the director's word in turn:
scratch folders declared in `bonsai.yaml` (`T-5793`, `todo`); the agent work (the pi brief's B1, B4, B5; the
envelope `bonsai.hook/1`, B6, B7; `design/pi-contract.md`); a cleaner role (the first piece after 1.0, 19 Oct);
playbooks, hillclimb and pickup (16 Oct); the general rewording of `workflow`; `bonsai agents` filters, release.yml's
notes A and B (BUG-102), BUG-099 to BUG-101, BUG-107's and BUG-108's rest, BUG-114's first half; Bonsai's repo naming its
packs by their tags. Go 1.27.2 stays until govulncheck loads a newer patch (GO-2026-6809, crypto/x509, unreached).
A named session starts with `claude -p --agent workflow:builder` in its worktree, launcher
`~/bonsai-checks/scripts/T-5708/run.sh`. Task numbers: T-5790 to T-5799 are all taken; 5.7's last were T-5739 to T-5741.
**Proposed for 5.7's planning** (`workflow:correct`, unproven until shown failing on old input): C3, C4, C5, C6 as
`R-2026-10-19-T-5615` names them. Every known bug: `records/bugs.md`.

**How work runs now:** agents by role, briefed with `bonsai message`; the shared rules
`~/bonsai-checks/briefs/common-2026-10-25.md` (over `common-2026-10-24.md`, `common-2026-10-23.md`, `common-2026-10-22.md`, `common-2026-10-21.md`, `common-2026-10-20.md`, `common-2026-10-19-day.md`, `common-2026-10-19.md` and `common-2026-10-18.md`) (audit first;
Windows runs by the orchestrator only, one at a time; load modest), plus the director's 19 Oct afternoon rules above. A
verifier, runner or auditor cannot yet write a scratch file by the Write tool (BUG-013; the sanctioned folder is not
built): it returns its work in its final reply, or the orchestrator or a builder makes the scratch. Run reports' times
come from the messages' `created` stamps, never estimates (5.6's M26 recurred on 19 Oct and was corrected).

**Time (21 Oct, the director):** every recorded time is UTC (`date -u` until rc.2 brings `bonsai now`); Bonsai converts for display. The time line costs ~3.8 ms a call on WSL and 65-87 ms on Windows (mostly Git Bash's start; `--version` alone 53-84 ms).

**Facts measured that later work rests on:** Claude Code 2.1.301-2.1.302. `--agent <role>` does not preload a role's
`skills:`. On Windows the PowerShell tool ignores Read and Edit deny rules. Moving a pack's ref renames the plugin
marketplace: a fresh Claude Code start before `update` installs. The guard's cost a call: ~8 ms WSL, ~49 ms Windows. Go's
`testing.T.Chdir` keeps the folder it leaves open (Windows renames no such folder). GitHub's ubuntu runner leaves
`/usr/local/bin` writable by others (the install job readies it). PowerShell's `-Filter 'x.*'` matches `x` itself.
Windows PowerShell 5.1's Add-Type compiles into %TEMP%. The auto-mode classifier refuses an orchestrator's `taskkill` of
another agent's Windows run, and an agent's `pgrep -f` loop can match itself. The guard refuses a builder starting `claude` (a runner runs scripted sessions). Claude Code 2.1.301 on Windows refuses an edit of a `\\\\wsl` path under acceptEdits as "a sensitive file" with no rule set. CI's windows test job: 53 min on `7a05c12d` (plain 26.5, tagged 25.5); after T-5617 and T-5790, windows 26 min and windows (fault tag) 16 min. The fault tag changes only `internal/guard`'s files; 7 packages reach it; a package reaching guard only on Windows is caught by Windows runs only. The Go build cache grows ~20 GB per side in 5 days of this work (worktrees, two tags, two systems; trimmed after 5 days unused). Status line through Git Bash: p50 55.5 ms,
p95 60.4 ms. Bonsai's production Go names no Anthropic model and imports no `net`; a tool name the guard does not know is
allowed unguarded (rule `tool-not-judged`); a pulled `running` task leaves the guard granting nothing
(`internal/workspace/active.go:175`), while a named session skips that rule (`internal/workspace/active.go:131-150`).

## Waiting on the director

- **Now:** what comes after 1.0, and in what order (the list under "Now: after 1.0" above).
- **Their call:** whether the pi side's brief (pi-bonsai-work.md in design/, untracked) is committed.
- **Whenever they like, now 1.0 is installed:** film again whatever changed on screen (`sh docs/demos/clips/film.sh
  ~/bonsai-clips <clip>`, then `sh docs/demos/edit.sh`; T-5710 item 12, clip 10 among them, BUG-098).
- **Their own step, from 5.6's close:** take the `"autoMemoryEnabled": true` line out of their local settings file
  (theirs only; the guard refuses agents). Also for them to read: the pin test's one exception, `release.yml` (5.7.2's).
- **Can wait:** V1's design question: an agent past its budget cannot stop a process it started (exempt it, or accept).
- **Whenever they like:** whether to allow deleting git's own lock files by exact name; whether the redactor's shared `=>` and
  `:=` limits go in a named part before the release; the plan's text on 5.5.11's limits is looser than the code;
  whether check should find a missing field in a format-1 document (T-5731's option C).

## Known bugs

**Every known bug, open and fixed, is in `records/bugs.md`** (the register, skill `workflow:bugs`; 19 Oct: 63 open,
22 fixed, from the audit `M-2026-10-18-39aecfea`). This page keeps no bug list. Not bugs, kept here: `bonsai init
--allow-exec` and `update` run `claude plugin install` on the person's machine (on Windows a warm session first).

**Housekeeping:**
- The roadmap artifact ("Trinetra Roadmap", shared with the studio's sessions): Bonsai's cards updated 25 Oct (version 47: Bonsai 1.0 and 5.7 done, the after-1.0 card up next, move 2's wait on 1.0 over). The agent design:
  "Workflow Pack Redesign" (version 4). The work-files research's page, "Solo to Team AI" (private), is the
  studio's now (a copy in its research folder, solo-to-team-ai).
- Another of the director's sessions may also work in `~/Servers/Bonsai`: read `origin/main`'s new commits before each
  push, and never stage a file this session did not write.
- The director's own local settings file (settings.local.json in .claude, out of git) holds their `skillOverrides` and `"autoMemoryEnabled": true` (out of
  git; the guard refuses agents' edits of it).
- Scratch under `~/bonsai-checks/` and `%USERPROFILE%\bonsai-checks\`: every piece's and verifier's folder (this
  session's: `T-5503`, `T-5511`, `T-5513`, `T-5514`, `T-5595`, `T-5598`, `T-5599`, `batch-1`, `verify-5.5-VG2`,
  `verify-5.5-VG3`), the launchers, the pre-release's source clone and build. Some hold the Windows user name or raw
  payloads (`%USERPROFILE%\bonsai-checks\T-5511\win2\raw\`): never paste from them. `T-5599`'s unused backup and script
  stay in `~/bonsai-checks/T-5599/`. 19 Oct's: `~/bonsai-checks/5.7-prep/decisions-2026-10-19.md` (the three
  answers) and `~/bonsai-checks/T-5791/` (the team research's scratch; its findings copied to the studio, the rest to clean).
