---
format: bonsai.state/1   # fields: bonsai check --schema bonsai.state; how to fill: skill base:state
updated: 2026-10-16
updated_by: orchestrator
labels: {}
---

# Bonsai: where it stands

Rewritten, never appended (`base:state`). Last rewritten 16 Oct 2026 at 11:45. **Step 5.5 is done** (the end verifier passed it
in its second round, 11:28). **Next: step 5.6, the machine pieces** (`design/plan-5.md` "### Step 5.6", from 5968). On `main` at
`0b99023d` (code at `bf7cc3d`, the installed pre-release). Bonsai's own repo now takes its packs `base` and `workflow` (5.5.7):
sessions here start agents by their `workflow:` roles; an undeclared agent type (`general-purpose`) is held as a read-only
helper. The floors: `go_tests` 4600, `go_tests_tagged` 4606. Every change lands by `CLAUDE.md`'s rules.

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

- **16 Oct:** overnight run on their word; Cursor's pstack surveyed, to discuss after 5.5 and before 5.6 builds; the
  Windows trust sitting; installed the pre-release (09:28); **`workflow` added** with builder minutes 240 by default and 480
  at most ("option 1"; limits are lines in `bonsai.yaml`, a set command noted for 5.6-5.7); **the walls' Windows half on
  5.5.6's rows**; **the Windows login size check dropped** ("Skip the check": agents never touch the credentials file; the
  warm-up stays); ran `bonsai update --allow-exec --yes` and the plugin installs; one history item left as is (their word);
  check 5 run by a builder, its Windows half on payload replays plus 5.5.5's Windows sessions; **the human CLI output to be
  improved: ask them first thing next session**.
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

`design/plan-5.md`, "### Step 5.6" (from 5968), approved 10 Oct with the director's (A) on notes about them. **Next session,
in this order:**

1. **Ask the director how they want `bonsai`'s human output to look** (their 16 Oct ask; a memory note too): within
   `CLAUDE.md`'s rules (ASCII, the standard library, `--json` unchanged); `update`'s plugin lines ran together.
2. **The pstack talk** (Cursor's plugin, surveyed 16 Oct), before 5.6 builds.
3. **5.6's tasks** written from its section (`T-56yy`), the plan's pieces in its order, builders as `workflow:builder`.

**How work runs now:** agents by role (`workflow:builder`, `workflow:verifier`, `workflow:runner`, `workflow:auditor`),
briefed with `bonsai message` (a brief's id in the Agent call's prompt; a newer brief in the thread resets a budget);
`general-purpose` agents write nothing. The Windows login size check is gone: a Windows run does the warm-up session
first and stops if a session reports no login. The briefs' shared rules: `~/bonsai-checks/briefs/common-2026-10-16.md`
(its closing lines name an earlier session). A verifier or runner cannot make a scratch repository (`git init`/`clone` are
refused to a role that writes nothing): give those steps to a builder.

**Facts measured that later work rests on:** Claude Code 2.1.301 on both sides. `--agent <role>` does not preload a role's
`skills:`. On Windows the PowerShell tool ignores Read and Edit deny rules, and Git Bash passes the WSL home's
network-path forms (base's README). Adding a pack renames a project's plugin marketplace, so a fresh Claude Code start in
the checkout comes before `update` installs. The guard's added cost a call: about 8 ms on WSL, about 49 ms on Windows.

## Waiting on the director

- **Next session, first:** how `bonsai`'s human output should look (above).
- **Can wait:** V1's design question: an agent past its budget cannot stop a process it started (exempt it, or accept).
- **Whenever they like:** whether to allow deleting git's own lock files by exact name; whether the redactor's shared `=>` and
  `:=` limits go in a named part before the release; turn GitHub Pages off; the plan's text on 5.5.11's limits is looser than
  the code.
- **Later, as the plan has them:** at 5.6 taking their `"autoMemoryEnabled": true` line out; at 5.7 the release settings, the
  `rc.1` tag, the installs, the trial, their word for 1.0, **and first the records-database idea**; and `base`'s tags: Bonsai's
  old `v0.1.1`-`v0.4.3` tags read as newer versions of `base` in `bonsai status` (a pack-scoped tag form, before the 1.0 tag).

## Loose ends

**From 5.5's end verifier, after 5.5:** `workflow`'s interview skill drops the studio's "Before you stop" paragraph (a
`workflow` change, then a ref move, the director's step); the guard reads a `~/` redirect target as relative to the working
folder (refuses more, not less); `check`'s missing-path finding counts a path git ignores; `update`'s human output runs two
plugin lines together; the verifier and runner roles cannot make scratch repositories; one Windows skip of
`TestMemberGonePid` in 20.
**From tonight, after 1.0 unless the director says otherwise:** the guard's unread runners and a git alias in the line (written
in shell.go's limits, `T-5520`); a spent agent cannot stop its own process (the director's design question); `script -q /dev/null
< cmds.txt`; the packs' release job could run only on push; `check.sh validate`'s "ok" wording; Bonsai's CI pins actions by tag;
`~/.pi/agent/models.json` and Gemini's other files unwalled; the operating skill at 12262 of its 12288 bytes; 2,167 subagent ids in
the log carry only a stop record (the recorder's look before 1.0).
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
- The director's own local settings file (settings.local.json in .claude, out of git) holds their `skillOverrides` and `"autoMemoryEnabled": true` (out of
  git; the guard refuses agents' edits of it).
- Scratch under `~/bonsai-checks/` and `%USERPROFILE%\bonsai-checks\`: every piece's and verifier's folder (this
  session's: `T-5503`, `T-5511`, `T-5513`, `T-5514`, `T-5595`, `T-5598`, `T-5599`, `batch-1`, `verify-5.5-VG2`,
  `verify-5.5-VG3`), the launchers, the pre-release's source clone and build. Some hold the Windows user name or raw
  payloads (`%USERPROFILE%\bonsai-checks\T-5511\win2\raw\`): never paste from them. `T-5599`'s unused backup and script
  stay in `~/bonsai-checks/T-5599/`.
