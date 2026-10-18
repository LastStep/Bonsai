---
format: bonsai.state/1   # fields: bonsai check --schema bonsai.state; how to fill: skill base:state
updated: 2026-10-19
updated_by: orchestrator
labels: {}
---

# Bonsai: where it stands

Rewritten, never appended (`base:state`). Last rewritten 19 Oct 2026 at 04:50 (an overnight session, the director
asleep). **Step 5.6's builds are all on `main`, CI green at `9124cdc7`** (every job, the windows job red since
`43e8b5fb` fixed by `T-5696`; one install (windows) flake passed on its re-run). 5.6.5, 5.6.6 and `T-5694` are `done`;
5.6.3 is at `verify` until the end verifier. Tonight also: **the bugs register `records/bugs.md`** (the director's word),
nine small bugs fixed (`T-5698`), and **5.7.4 drafted on its branch** (`T-5704`, green, not landed: the director, 19 Oct,
"5.7.0 lands, 5.7.4 drafted"). **Left for 5.6:** the end Windows run (agent half, then the director's half; the runbook
`~/bonsai-checks/end-5.6/runbook.md`); the floors; the 5.6 end verifier; the cleanup the director chose (below). The installed pre-release is still `bf7cc3d`. Bonsai's own repo takes `base`
(`4b41790b`) and `workflow` (`a42a4f4`), plugins under the marketplace `bonsai-bonsai-299659af`. The floors: `go_tests`
4600, `go_tests_tagged` 4606 (the ladder's recorded counts are higher: 4750 and 4756 at `192a29dc`). Every change lands by
`CLAUDE.md`'s rules, with the director's 18 Oct waiver of the per-landing Windows half for 5.6.

**No person's name** goes into any file or commit message (the director, 15 Oct; `CLAUDE.md`'s Safety section): "the
director" in this project's files, "the person" in the product. Files already under `records/` and `design/` keep
what they hold, on the director's word.

## In short

Bonsai is being rebuilt as one small Go program that gives every project the same formats, packs, guards, recorder and
proof ladder (`design/one-pager.md`). The plan is approved (8 Oct). Part 0 and all six parts of the walking skeleton
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

**Hours** (each run's minutes from the run reports): part 0 45 minutes; the skeleton 6 h 18 min; 5.1 866 minutes; 5.2
1,061; 5.3 about 2,013 (33 h 33 min); 5.4 about 2,372 (39 h 32 min); **5.5 about 78 h** with its planning and reviews (the end
verifier's count, under its re-ask line of 138). Questions put to the director in 5.5: twenty-three before 16 Oct, and on 16 Oct
eight (the workflow limits, the walls' Windows half, the login check, the history, check 5, check 5 on Windows, and the
morning's install and Windows steps as asks). Step 5's Windows-only tally: about 143 minutes plus 16 Oct's Windows check-10
helpers (about 6 h, side by side).

## The director's decisions

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
    **At step 5.7, raise with the director first:** they are thinking of putting records in some form of database,
    and maybe Bonsai's local logs and tasks too, so a project's agentic work is private yet shareable with a team and
    across devices (to weigh: `CLAUDE.md`'s "no daemon, no network service", and that a shared store may be the
    studio's).
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

- `LastStep/Bonsai`: `main` the only branch, at `bf7cc3d` (records after it); the newest tag `v0.4.3`; no open pull request;
  `release.yml` disabled and build-only; ruleset `main-protection` blocks force pushes and deletion. CI green at `bf7cc3d`.
- `LastStep/bonsai-workflow`: **public**, `main` at `ef5b5ae`, CI green on both sides.
- `LastStep/bonsai-test-pack`: public, `main` at H (`19e3ca5`).
- The old website stays on GitHub Pages until the director turns Pages off.

## Now: step 5.6, the machine pieces

| Piece | Task | State | Where |
|---|---|---|---|
| 5.6.0-5.6.2, 5.6.4-5.6.10 | `T-5600`-`T-5610` but 5.6.3 | done | CI green at `9124cdc7` |
| 5.6.3 personal memory | `T-5603` | verify | `7f2d76c0`; the end verifier and the end Windows run (its copy, the cross-side rule) |
| ref moves, Windows CI fixes | `T-5691`, `T-5692`, `T-5696` | done | T-5696: two engine tests' want; the guard test's folder; bad-path words; `C:foo` fails closed; a file link to a folder still bad-path (BUG-086, verifier round 2's junction case) |
| V1, its B1 | `T-5693`, `T-5694` | done | |
| The bugs register's row in CLAUDE.md | `T-5697` | done | uncommitted until tonight's records commit |
| Small fixes | `T-5698` | done | BUG-001-004, 010, 012, 024, 030-032, 082 (code), 084, 085; verifier PASS `M-2026-10-18-0f2afe05` |
| 5.7.4, the words for 1.0 | `T-5704` | verify | branch `T-5704` at `3fc559c6`, climb green; held until 5.6's end verifier passes and the records-database talk |

**Next, in this order:** the end Windows run's agent half from the runbook, at `main`'s top, one Windows run at a time;
5.7.0 (CI's action pins, lint, actionlint, a new selfcheck package), run alone (it grants `.github/**`), landed after 5.6's
work by the director's 19 Oct word; the floors (`T-5695`); the 5.6 end verifier (fresh Opus); the cleanup; then
`workflow:correct`'s first run (5.6.10). **Small follow-ups:** the three new deny lines reach this repo's settings at the
next pre-release's `update`. Every known bug: `records/bugs.md`.

**How work runs now:** agents by role, briefed with `bonsai message`; today's shared rules
`~/bonsai-checks/briefs/common-2026-10-19.md` (over `common-2026-10-18.md`) (audit first; Windows runs by the orchestrator only, one at a time; load
modest). A verifier or runner cannot make a scratch repository or a Windows clone (BUG-013): the orchestrator or a builder makes it.

**Facts measured that later work rests on:** Claude Code 2.1.301-2.1.302. `--agent <role>` does not preload a role's
`skills:`. On Windows the PowerShell tool ignores Read and Edit deny rules. Moving a pack's ref renames the plugin
marketplace: a fresh Claude Code start before `update` installs. The guard's cost a call: ~8 ms WSL, ~49 ms Windows. Go's
`testing.T.Chdir` keeps the folder it leaves open (Windows renames no such folder). GitHub's ubuntu runner leaves
`/usr/local/bin` writable by others (the install job readies it). PowerShell's `-Filter 'x.*'` matches `x` itself.
Windows PowerShell 5.1's Add-Type compiles into %TEMP%. The auto-mode classifier refuses an orchestrator's `taskkill` of
another agent's Windows run, and an agent's `pgrep -f` loop can match itself. Status line through Git Bash: p50 55.5 ms,
p95 60.4 ms.

## Waiting on the director

- **Next:** nothing; 5.6's end Windows run is the orchestrator's, alone, when the machine is quiet.
- **Can wait:** V1's design question: an agent past its budget cannot stop a process it started (exempt it, or accept).
- **Whenever they like:** whether to allow deleting git's own lock files by exact name; whether the redactor's shared `=>` and
  `:=` limits go in a named part before the release; turn GitHub Pages off; the plan's text on 5.5.11's limits is looser than
  the code.
- **Later, as the plan has them:** at 5.6 taking their `"autoMemoryEnabled": true` line out; at 5.7 the release settings, the
  `rc.1` tag, the installs, the trial, their word for 1.0, **and first the records-database idea**; and `base`'s tags: Bonsai's
  old `v0.1.1`-`v0.4.3` tags read as newer versions of `base` in `bonsai status` (a pack-scoped tag form, before the 1.0 tag).

## Known bugs

**Every known bug, open and fixed, is in `records/bugs.md`** (the register, skill `workflow:bugs`; 19 Oct: 63 open,
22 fixed, from the audit `M-2026-10-18-39aecfea`). This page keeps no bug list. Not bugs, kept here: `bonsai init
--allow-exec` and `update` run `claude plugin install` on the person's machine (on Windows a warm session first).

**Housekeeping:**
- The roadmap artifact ("Trinetra Roadmap"): Bonsai's cards updated at 5.4's end, next at 5.5's end. The agent design:
  "Workflow Pack Redesign" (version 4).
- Another of the director's sessions may also work in `~/Servers/Bonsai`: read `origin/main`'s new commits before each
  push, and never stage a file this session did not write.
- The director's own local settings file (settings.local.json in .claude, out of git) holds their `skillOverrides` and `"autoMemoryEnabled": true` (out of
  git; the guard refuses agents' edits of it).
- Scratch under `~/bonsai-checks/` and `%USERPROFILE%\bonsai-checks\`: every piece's and verifier's folder (this
  session's: `T-5503`, `T-5511`, `T-5513`, `T-5514`, `T-5595`, `T-5598`, `T-5599`, `batch-1`, `verify-5.5-VG2`,
  `verify-5.5-VG3`), the launchers, the pre-release's source clone and build. Some hold the Windows user name or raw
  payloads (`%USERPROFILE%\bonsai-checks\T-5511\win2\raw\`): never paste from them. `T-5599`'s unused backup and script
  stay in `~/bonsai-checks/T-5599/`.
