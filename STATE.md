# Bonsai: where it stands

Rewritten, never appended. Last rewritten 15 Oct 2026 at 17:00. **Step 5.5 is building.** On `main` at `d81f5d4` (code at `8d151a6`):
every piece but 5.5.4-5.5.7 has landed; this session closed **5.5.8, 5.5.3 and 5.5.11** (VG passed 5.5.11 after its
fix rounds `T-5598`, `T-5513`, `T-5514`), landed `T-5595` (the Windows timing tests) and took the director's name out
of the product, `CLAUDE.md` and `workflow` (`T-5599`). `workflow` (5.5.4) is **public** at `fed3bbf` (the director
made it public, 15 Oct). **5.5.4 closed 16 Oct 00:29 (the link with `base`: 21 block lines); next 5.5.5, then 5.5.6** ("Now: step 5.5"
below). About 61 h of 5.5's 63.5-106 (re-ask 138) logged. Every change lands by `CLAUDE.md`'s rules "from the link".

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
- **Step 5.5, the packs, the roles and their messages: building** (approved 14 Oct, 17:52). The design in the artifact
  "Workflow Pack Redesign" (https://claude.ai/artifact/4j5RgCaPQs9QDh8sShbKNN, version 4).

| Piece | Task | Status | Where |
|---|---|---|---|
| 5.5.0 the engine's part for packs | `T-5500` | verify | landed `e1437d0`; VM passed; waits for V1 (after 5.5.6, the director's "Keep the plan") |
| 5.5.1 the pack template, its CI | `T-5501` | verify | landed `c641c15`; waits for V1 |
| 5.5.2 `base` | `T-5502` | verify | landed `53fca0a`; waits for V1 |
| 5.5.3 the operating skill | `T-5503` | **done** | landed `9370976` (batch 1), CI green |
| 5.5.4 `workflow` | `T-5504` | **done** | public at `c48d534`, CI green; the link with `base` 21 block lines (16 Oct) |
| 5.5.5 real sessions | `T-5505` | todo | after 5.5.3 and 5.5.4 |
| 5.5.6 the walls | `T-5506` | todo | after 5.5.5 |
| 5.5.7 Bonsai takes the packs | `T-5507` | todo | after V1: the new pre-release, the director's sitting |
| 5.5.8 CI green, carry-overs | `T-5508` | **done** | H fixed in `T-5598`; VG round 2 re-read items 4 and 5 |
| 5.5.9 formats set 8 | `T-5509` | done | `f239f8f` |
| 5.5.10 messages | `T-5510` | done | `298ed35` |
| 5.5.11 the guard's role rules | `T-5511` | **done** | VG passed it (VG3's re-run, 16:12); CI green at `8d151a6` |
| 5.5.11b M1 (permissions) | `T-5513` | **done** | landed `d3a00bd` (batch 1) |
| 5.5.11c S2 (this system's place) | `T-5514` | **done** | landed `8d151a6`, CI green |
| 5.5.12 `bonsai agents` | `T-5512` | done | `7615ec7` |

Outside the pieces: `T-5590`-`T-5597` done (the re-plan, the safety line, two Windows flakes, VM and its fixes,
the Windows timing tests, the batch rule), `T-5598` done (VG's first fix round), `T-5596` done (VG, three rounds),
`T-5599` done (the name out of the product, `CLAUDE.md`, `STATE.md` and `workflow`; its fresh check found none). Run reports
`records/runs/R-2026-10-1[45]-T-55*.md`.

**Hours** (each run's minutes from the run reports): part 0 45 minutes; the skeleton 6 h 18 min; 5.1 866 minutes; 5.2
1,061; 5.3 about 2,013 (33 h 33 min); 5.4 about 2,372 (39 h 32 min). **5.5's building so far: about 3,635 minutes
(60.6 h)** of its 63.5-106 h (re-ask 138): 2,776 before this session, and this session's runs 859 (`T-5598` 144,
`T-5513` 146, `T-5514` 93, VG rounds 2 and 3 132, `T-5503` 151, `T-5599` 133, `T-5595` 50, `T-5511`'s Windows half 10),
runs side by side each counted; the orchestrator's own session (about 10:40-17:00) only partly in them. Later parts'
planning: 5.6 67, 5.7 117. **Questions put to the director in 5.5 so far: twenty-three** (six option rounds this
session), stop line 3 crossed in the open. Step 5's Windows-only tally: about 143 minutes of 8 h.

## The director's decisions

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

- `LastStep/Bonsai`: `main` the only branch, at `8d151a6`; the newest tag `v0.4.3`; no open pull request;
  `release.yml` disabled and build-only; ruleset `main-protection` blocks force pushes and deletion. CodeQL failed three
  times on `2b9d09f` on GitHub's side ("An action could not be found ... actions/setup-go ..."), green since.
- `LastStep/bonsai-workflow`: **public**, `main` at `fed3bbf`, CI green on both sides.
- `LastStep/bonsai-test-pack`: public, `main` at H (`19e3ca5`).
- The old website stays on GitHub Pages until the director turns Pages off.

## Now: step 5.5, the packs, building

`design/plan-5.md`, "### Step 5.5" (lines 3437-5967); "Notes per piece" from 4577; "5.5 done" from 5726. Tasks in
`records/tasks/`. Outside-piece numbers `T-5590`-`T-5599` are all taken: new outside work takes the next free number
as a lettered piece (as `T-5513`, `T-5514`), or the director's say on numbering. **Land in batches.** **The next
session, in this order (parallel where marked):**

1. **5.5.4's link test with `base`** (`T-5504`'s last `done_when` line: a scratch project linked to `workflow` and
   `base`, its block under 40 lines). An agent's job; `T-5504` then `done` (the director's read was given 15 Oct, and
   the repository is public).
2. **5.5.5, the packs in real sessions** (`T-5505`, a Sonnet agent; plan 5511-5572): S1-S4 on WSL, S1 on Windows, the
   brief, report and budget lines R7-R9 with the operating skill as landed. The Windows login rule holds: warm-up,
   then the exact size check rules.md names (allowed by the director's rule).
3. **5.5.6, the walls tried once on both sides** (`T-5506`, a Sonnet agent; plan 5573-5599), after 5.5.5.
4. **V1** (plan 5600-5611): the walls, the packs' CI and release, 5.5.0's `<protocols>/`; then `T-5500`-`T-5502` `done`.
5. **5.5.7** (plan 5612-): the new pre-release from a clean clone (the current one first copied to
   `~/bonsai-checks/prerelease/previous/`), the director's sitting (install, `bonsai update --allow-exec --yes`,
   their yes to `workflow` in `bonsai.yaml`, about 3 minutes on Windows), `CLAUDE.md` shrinks; then the end verifier
   and the floors.

**How work runs** (`CLAUDE.md`, "from the link"): every piece a task in `records/tasks/` (the orchestrator alone writes
and moves them), its run report `R-<date>-<task id>-<topic>.md` in format 1; builders are subagents in plain worktrees
from `main`, climbing `/usr/local/bin/bonsai ladder --task <id> --json`; a piece lands only on a green local result at
its HEAD whose record names the installed `bonsai`, plus check 10's Windows half before the push and CI after it. The
builders' rules: `~/bonsai-checks/briefs/rules.md`. **Its `Claude-Session` line still names an earlier session**: the
auto-mode classifier refused the orchestrator's edit of it on 15 Oct ("Instruction Poisoning"), so every brief gives
the session's closing lines and the no-name rule itself. The floors stand at `go_tests` 4188 and `go_tests_tagged`
4194 (5.4's close); climbs at `8d151a6` see about 4569 and 4575. **PowerShell 5.1 writes `>` output as UTF-16**: the
Windows check-10 scripts' own counts read 0; count from the JSON with NULs and the byte-order mark stripped.

**Facts measured that later work rests on:** Claude Code 2.1.296 on Windows, 2.1.300 on WSL. **The hook payloads on
both sides** (`~/bonsai-checks/T-5511/payloads.md`, fixtures in `fixtures/{wsl,windows}/`): equal field by field but
for Windows paths with backslashes, an Agent call with no `subagent_type` async on Windows, and a refusal's bracket
showing the hook command as typed; the caller's role is the payload's own `agent_type` (with `agent_id` inside a
subagent); no field ties a new subagent to its call before PostToolUse's `agentId`; `session_id` is shared; a brief's
id sits in `tool_input.prompt`. **The auto-mode classifier** refuses a runner with a permission-bypass flag ("Create
Unsafe Agents"; use `--allowedTools`), a bulk in-place rewrite of many files ("Irreversible Local Destruction"), and a
merge, push and cleanup typed as one command ("Blind Apply": run each step alone). The earlier facts (Stop hooks,
`onFailure`, worktree variables, the Windows login, the guard's timings, the switch's measurements) are in 5.4's and
5.5's run reports.

## Waiting on the director

- Nothing now.
- **Whenever they like:** whether to allow deleting git's own lock files by exact name (`rm -f .git/index.lock`);
  whether the redactor's shared `=>` and `:=` limits go in a named part before the release; turn GitHub Pages off; the
  plan's text on 5.5.11's limits is now looser than the code (`T-5513`: "fails open" and a second-line fault "lets the
  call through" no longer hold): the design files are kept as they are, so a change there is theirs to ask for.
- **Later, as the plan has them:** at 5.5.7 the sitting (above); at 5.6 taking their `"autoMemoryEnabled": true` line
  out of `.claude/settings.local.json`; at 5.7 the release settings, the `rc.1` tag, the installs, the trial's
  analysis, their own trial on a fresh project, their word for 1.0, **and first the records-database idea (above)**.

## Loose ends

**Known limits from 5.5.11's verifiers, after 1.0** (VG3, 15 Oct; none frees a budget): Git Bash's `icacls /c //deny
... //t` (`/c` read as icacls's switch while Git Bash makes it `C:/`; an ACL change fails closed); `(Get-Item
x).Delete()`/`.MoveTo()` on a state starts a fresh window (written in shell.go's and budget.go's limits); write words
outside the tables onto Bonsai's files (rsync, tar -x, unzip, curl -o, wget -O; written); a calls file emptied in its
first 11 s, or put back from a copy as long as its window's start, counts afresh (budget.go's limit text); a rare false
hold when a first call's line was never written; an agent with no meta file joins an open call of the last 10 minutes;
plain `chmod 755 .` at a project's top refused to a builder; the settle wait in the recorder; a `BONSAI_TASK` set in a
nested shell not followed; records attributed, not authenticated.
**Known limits from earlier, after 1.0:** the rung 0 race (V2d); the gate's git timeout under load; the machine record
writable by an agent's shell; a task file's grants read from main's working tree; D2 (a scratch link typed on the
shell line refused); K2 (a fake program first on the caller's PATH fools a rung; CI runs the ladder); Git LFS and
git-crypt checkouts read dirty; the packs' protected paths in the lock outside the settings record; a `documents` kind
added moves a pack's place; the redactor's (5.2's L1 and the shared forms); the guard's, listed in
`internal/guard/guard.go`, `shell.go`, `shell_env.go`, `rolecmds.go`, `attrwords.go`, `roles.go`, `skill.go` and
`budget.go`.
**Load-sensitive tests** seen 14 Oct (`internal/redact`'s `TestLinearTime`): fix only if CI meets them.

**Placed for later:**
- **5.5:** the operating skill's line on an added rung (since K1 a person's step: the skill says so; the plan's ~4031
  does not); `bonsai init --allow-exec` and `update` run `claude plugin install` on the person's machine (on Windows a
  warm session first); a pack's relative deny rule does not hold after a shell `cd`; contract §5.4's `workflow.model`
  lacks haiku; the guard's stderr refusal does not print its rule word (the skill points to `bonsai logs`).
- **The next formats set:** a shared active-task fixture for a task file holding only `garbage: [`; `ladder.schema.json`'s
  `leftovers` description.
- **5.6:** installers write the binary by rename and the line tests `[ -s ]`; the "no bonsai is installed" words Claude
  Code shows; the machine folder `unlink` leaves; `install.json`'s keys; the Windows personal-memory warning; a per-id
  index of the machine records.
- **One home, later:** the recorder's `taskAllows` and the guard's `allowsOf`; `underHome` and `fromHome`;
  `workspace.VerifyMain`'s quoted system error.

**Housekeeping:**
- The roadmap artifact ("Trinetra Roadmap"): Bonsai's cards updated at 5.4's end, next at 5.5's end. The agent design:
  "Workflow Pack Redesign" (version 4).
- Another of the director's sessions may also work in `~/Servers/Bonsai`: read `origin/main`'s new commits before each
  push, and never stage a file this session did not write.
- The director's own `.claude/settings.local.json` holds their `skillOverrides` and `"autoMemoryEnabled": true` (out of
  git; the guard refuses agents' edits of it).
- Scratch under `~/bonsai-checks/` and `%USERPROFILE%\bonsai-checks\`: every piece's and verifier's folder (this
  session's: `T-5503`, `T-5511`, `T-5513`, `T-5514`, `T-5595`, `T-5598`, `T-5599`, `batch-1`, `verify-5.5-VG2`,
  `verify-5.5-VG3`), the launchers, the pre-release's source clone and build. Some hold the Windows user name or raw
  payloads (`%USERPROFILE%\bonsai-checks\T-5511\win2\raw\`): never paste from them. `T-5599`'s unused backup and script
  stay in `~/bonsai-checks/T-5599/`.
