# The trial on rc.1: what happened, from the records

For the director (T-5709, plan-5.md note 5.7.9). Nothing here asks anything of you: read it, then reply "happy" or
say what to fix (note 5.7.9, item 3).

**Times.** Every time on this page is India time (IST, UTC+5:30), the clock the run reports use. Bonsai's own records
(the log, the asks, the ladder results, `.bonsai/sessions.md`) keep UTC; each such time here has had 5 h 30 min added.

**How to read the pointers.** Each claim ends with where it comes from, in brackets: a file and line, a record id
(`M-...` a message, `agent:h-...` an ask), a commit, or a command. "the log" is
`.bonsai/local/log/s-64096e53-5e58-446c-9441-188608e138e9.ndjson` in the main checkout (the named builder session's
log), read by line number; "the orchestrator's log" is the same folder's `s-f7c1ed9c-...ndjson`. `.bonsai/local/` is
never committed, so those pointers work only on this machine. "T-5707's report" is
`records/runs/R-2026-10-20-T-5707-trial-on-rc1.md`; "T-5708's report" is
`records/runs/R-2026-10-20-T-5708-stopgate-names-its-rule.md`.

## In short

On 20 Oct, Bonsai's own repo ran the first published pre-release, `v1.0.0-rc.1`, installed at `/usr/local/bin/bonsai`
(sha256 `9883e243...a307`) [T-5707's report :31; the log :186, `bonsai_sha256`]. Bonsai built one small real feature
for itself under that build, the one you picked: **when the stop gate holds a session back, it now says which rule
did it**, as the guard already does [T-5708's report :112-120]. A builder ran as a named session, met the guard and the
stop gate on purpose (two probes), asked you two questions you answered from your own terminal, and finished green. A
fresh verifier passed it, CI on the branch was green on every job, and it landed on `main` as `7afb9faa`
[T-5707's report :69-73; `M-2026-10-20-57b77ed9`]. It reaches you in rc.2: this trial's own blocks, printed by rc.1,
show no rule line [T-5708's report :62, :114-115].

Twelve rough edges came out of it: seven noted at the time and five found while gathering the evidence for this
page. Each is listed under "What was rough" with where it goes. **Four are fixed in rc.2** (items 1, 2, 4 and 8, all
on `records/tasks/T-5710-rc2-fixes.md`), **three wait until after 1.0** (items 3, 11 and 12, each a row in
`records/bugs.md`), and **five need nothing in code** (items 5, 6, 7, 9 and 10). One more, found while writing the
page, waits until after 1.0 too (a gap in the records, BUG-095).

A separate dry run, an agent following the README's "A first project" steps on the installed rc.1, went green at its
last step and proposed 13 fixes for rc.2, most to the README and three to Bonsai's own words. One more finding waits
until after 1.0 (below, "The README's dry run").

## What Bonsai created and changed, and why each exists

### In the repo (committed)

| What | Why it exists | Source |
|---|---|---|
| 11 commits on branch `T-5708`, `6ef7e56d` to `7afb9faa`: 5 change code, tests or a skill, 6 are the builder's run report as it went | the feature, built test-first; the run report is written as a log so a session that dies still leaves its story | `git log --oneline ecb65520..7afb9faa` |
| `records/tasks/T-5707-...md`, `T-5708-...md`: the two tasks, moved by the orchestrator | a task is the contract: what is built, what proves it, which protected files it may change | `ecb65520`, `8fdf63d2` |
| The two run reports | one per task, a log with a row for every run's minutes | as named above |
| `.bonsai/tasks.md`, `.bonsai/sessions.md`: two tables `bonsai check --write` rebuilds from the task files and the log | a quick read of every task's state, and of every session and subagent with its minutes; a table grants nothing | `8fdf63d2` (tasks.md +5/-3, sessions.md +4 lines); `.bonsai/sessions.md:1-3` |
| `records/bugs.md`: BUG-088 and BUG-089 moved to Fixed | the one register of known bugs | `records/bugs.md:114-115` |

### In `.bonsai/local/` (this checkout only, never committed)

Before and after the run, the orchestrator listed every file Bonsai keeps, with its size
[`~/bonsai-checks/T-5708/files-before.txt` (987 lines, taken 16:47:33), `files-after.txt` (1004 lines, 17:47:10)].
Compared name by name, every difference is under `.bonsai/local/`:

| Folder | Change | What it holds, and why |
|---|---|---|
| `log/` | 1 new file (the session's log, 147,666 bytes); 2 grown (the orchestrator's session log; `w-2026-10-20.ndjson`, the day's workspace log, 1,947 to 3,155 bytes) | the recorder's record of every call, guard decision, climb and ask, one line each: the evidence behind everything on this page |
| `agents/` | 8 more files: the session's `.calls` and `.state`, the verifier's, one claim, three `ties/` (one per brief), and `spawns.ndjson` grown | each agent's counts against its budget (minutes, actions, red climbs), and which brief ties which agent to its task |
| `asks/` | 1 new file, `2026-10-20.ndjson` (4,046 bytes, 4 records) | the two questions to you and your two answers |
| `ladder/` | 1 new file, `T-5708.json` (7,938 bytes) | the newest climb's result: the proof the stop gate and the landing rule read |
| `messages/` | 6 new files | the briefs and reports between the agents |

[`diff` of the two lists sorted by name; file sizes from the lists]

### In the Bonsai home (by kind only)

The home holds five kinds of thing: `cache/` (the packs' git copies), `install.json` (which build is installed),
`locks/`, `salt` and `workspaces/` (this machine's record of each linked repo). There is no `personal/` folder, so no
personal `INDEX.md` [`ls -a ~/.bonsai/`, 21 Oct]. The before and after lists cover only `cache/`: 89 entries before
and 89 after, with the same names and sizes [the two lists, sorted by name]. **T-5707's report is wrong here**: its
line 75 says the cache's git folders were "rewritten by a fetch", but the lists show no file added, removed or
resized in the home. Only the order the files were listed in changed. (Same names and sizes cannot rule out a file
rewritten byte for byte; nothing in the lists shows one.)

## What the log and the tables hold

The session's log has 198 records [the log; `wc -l`, and counted by kind]:

| Records | Count |
|---|---|
| tool calls started / ended | 60 / 59 |
| guard: allowed, a shell command it does not judge (`shell-not-judged`) | 41 |
| guard: allowed, a `bonsai` command (`bonsai-command`) | 16 |
| guard: allowed, a tool it does not judge; a brief not tied to a task | 2; 1 |
| **guard: refused, `stop-red`** (a stop while the ladder is red) | **2** |
| **guard: refused, `protected`** (an edit of a protected file) | **1** |
| guard: allowed the final stop (`stop-proven`) | 1 |
| ladder climbs | 3 |
| stops; asks filed; messages | 3; 2; 2 |
| prompts: the brief, and a task notification | 2 |
| session start, end, token usage | 1 each |

None of the 198 lines holds a home-folder path [a count of home-folder paths in the log: 0]. A few lines, shortened to their
telling fields (the log keeps more: the workspace id, the session id, the checkout `Bonsai-T-5708`, the branch):

```
{"at":"2026-10-20T11:21:13.460Z","event":"guard","decision":"deny","rule":"stop-red","target":".bonsai/local/ladder/T-5708.json","text":"the ladder result for T-5708 is not green (rung 3 tests red: exit 1 83 tests failed (each in tests.failed, with the command that runs it a...)","task":"T-5708","role":"builder"}
{"at":"2026-10-20T11:21:27.576Z","event":"guard","tool":"Edit","target":"CLAUDE.md","decision":"deny","rule":"protected","text":"Edit of \"CLAUDE.md\" refused: it is on bonsai.yaml's protected list (\"CLAUDE.md\"), and the running task T-5708 grants no path (its bonsai.allows is empty)","task":"T-5708","role":"builder"}
{"at":"2026-10-20T11:21:50.288Z","event":"ask","kind":"file","target":"agent:h-639ab913f88a","task":"T-5708","role":"builder"}
{"at":"2026-10-20T11:48:57.399Z","event":"ladder","kind":"local","target":".bonsai/local/ladder/T-5708.json","text":"sha256:691141c0f1200b7ae334e006bf8c156100d3f737a76cf70c51c60fe03c0ec839","bonsai_path":"/usr/local/bin/bonsai","bonsai_sha256":"9883e243508c9f3e2a6c2bb8ef812fa23631a980274eed38c0d14c406f86a307"}
{"at":"2026-10-20T11:49:39.653Z","event":"guard","decision":"allow","rule":"stop-proven","target":".bonsai/local/ladder/T-5708.json","task":"T-5708","role":"builder"}
```

[the log :60, :68, :80, :186, :196]. In order: the stop gate's first block (probe 1, 16:51), the guard's refusal of
`CLAUDE.md` (probe 2, 16:51), your first question filed (16:51), the final green climb tied to the installed rc.1 by
its sha256 (17:18), and the stop gate letting the session end (17:19).

**The tables.** `.bonsai/sessions.md` holds a row per session and per subagent with its minutes: the named builder,
16:47 to 17:19, 32 minutes, and the verifier, 17:20 to 17:25, 5 minutes [`.bonsai/sessions.md:430-431`; the table
prints UTC]. Its per-task sum gives T-5708 0.5 h for the builder and 0.1 h for the verifier [`.bonsai/sessions.md:491-492`].
`.bonsai/tasks.md` lists T-5707 and T-5708 as `done` [`.bonsai/tasks.md:15-16`]. Today `bonsai check` reports no
findings and two warnings: both tables lag behind the files (tasks.md differs from a rebuild; sessions.md lacks rows for
sessions that ended later), each with `next: run: bonsai check --write` [`bonsai check`, 21 Oct].

## How the agents worked, run by run

| Run | Who, model | Time | Minutes | Source |
|---|---|---|---|---|
| Set-up climb of `main` | the orchestrator, Opus | 16:42:24-16:44:38 | 2 (the climb) | `.bonsai/local/ladder/T-5707.json` |
| The proposals | an auditor, Opus | 16:37-16:41 | 4 | `.bonsai/sessions.md:429`; `M-2026-10-20-392e067a` |
| The feature | a named builder session, Opus 5.5 | 16:47:57-17:19:40 | 32 by its launcher (28.5 by Claude Code's own count) | `~/bonsai-checks/T-5708/start.txt`, `exit.txt`; `stream.jsonl`'s result |
| The check | a fresh verifier, Opus | 17:20-17:25 | 5 | `.bonsai/sessions.md:431` |
| The landing | the orchestrator | to 17:48 | | T-5707's report :70-73 |

The times for the auditor and the verifier come from the sessions table. T-5707's report, written by hand, gives them
a minute or two apart: 16:36-16:40 and 17:22-17:27 [T-5707's report :25, :27].

### 1. Set up on rc.1

`bonsai check --json` in the main checkout found nothing, so Bonsai's own hook lines were already rc.1's and no update
was needed [T-5707's report :31-32]. A climb of `main` by rc.1 under T-5707 was green at `67fa1ea3`: rung 0 3.8 s,
1 2.3 s, 2 0.5 s, 3 60.8 s, 4 55.2 s, 5 green; the result file's sha256sum `5bc9da27...e194a`
[`sha256sum .bonsai/local/ladder/T-5707.json`, matching T-5707's report :33]. Tests: 4,807 seen, 9 skipped, so 4,798 passed against the floor of 4,756;
with the fault switch 2,125 seen, 3 skipped, 2,122 passed against 2,120 [the same file, `tests`; `bonsai.yaml:110`
counts the tests that passed]. The result reads `dirty: true`: the orchestrator's task moves were not yet committed
[T-5707's report :34].

### 2. The proposals and your pick

The auditor proposed three: A, `bonsai status` reading old tags as newer releases (BUG-014, its recommendation); B,
the stop gate names its rule (BUG-089, BUG-088); C, two more key files walled (BUG-035). You picked **B**. A went on
rc.2's list regardless, since every project would otherwise read `v1.0.0` as a newer base [T-5707's report :36-39;
`T-5710-rc2-fixes.md:23-25`].

### 3. The named builder session

Started plainly as `claude -p --agent workflow:builder` in its worktree with `BONSAI_TASK=T-5708` and
`BONSAI_ROLE=builder`, not through `claude-here` as `CLAUDE.md` says (rough 1, below) [T-5707's report :40-45]. It
ran 63 turns, exited 0 with nothing on stderr, and cost 2.79 USD by Claude Code's count [`stream.jsonl`, its `result`
line; `exit.txt`; `stderr.txt` is 0 bytes]. The task and run report were committed on `main` before its worktree was
made (`ecb65520`, 16:47).

What it did, in order [T-5708's report :48-110]:
1. **A failing test first** (`6ef7e56d`, 16:49): every block must end with a `rule:` line. On the old code,
   `go test ./internal/stopgate/` failed on every block case.
2. **Climb 1, red as wanted** (16:50:52): rung 3 red, 83 failed, all in the stop gate; sha256 `db8a477c...d733c`
   [the log :48; T-5708's report :49-52].
3. **Probe 1, the stop gate** (16:51): it ended its turn on purpose with the ladder red. The gate **blocked** the stop:
   "bonsai stop gate: the ladder result for T-5708 is not green (rung 3 tests red: ...). next: fix what the red rungs
   name, commit, then run: /usr/local/bin/bonsai ladder --task T-5708." There was no rule line: that is the gap the
   feature closes [T-5708's report :55-62; the log :60].
4. **Probe 2, the guard** (16:51): one Edit of `CLAUDE.md`, a protected file the task does not grant. **Refused**,
   the file unchanged, ending "rule: protected." [T-5708's report :64-70; the log :68].
5. **Your first question** (filed 16:51:50, `agent:h-639ab913f88a`, Decide): should the gate's own error words take
   Bonsai's shared ones? A leave them; **B (recommended) take the two shared kinds, keep its extra ones, no word you
   see changes**; C move its extras into the shared list [`.bonsai/local/asks/2026-10-20.ndjson`, record 1]. It carried
   on with the fix while it waited.
6. **The fix** (16:52): one field, `Rule`, passed through to the shared frame. Five cases still failed: two rule names
   held the word "missing", which the gate must never print (Claude Code may read such a Stop message as a hook that
   failed to start, and let the session end). It did not loosen that safety test [T-5708's report :78-83].
7. **Your second question** (filed 16:53:14, `agent:h-ae479b969f78`, Decide): how should those two rules' line read?
   **A (recommended) no rule line for those two**; B rename the two rules everywhere; C print them anyway and loosen the
   safety test [the asks file, record 2].
8. **A second block** (16:53:52): it ended a turn to wait for your answers, and the gate blocked that stop too, with
   the same words as probe 1 (rough 3) [the log :126; T-5708's report :91].
9. **Your answers** (17:00:43 and 17:01:34), both typed in your own terminal as `--words "B"`: by terminal, `choice`
   empty [the asks file, records 3-4]. So **B on both**: take the shared words, and rename the two rules.
10. **B on question 1** (`3b390caa`): the shared kinds taken; no word changed.
11. **B on question 2** needed three files outside the task's list; the builder stopped and asked the orchestrator,
    which added them (`M-2026-10-20-2c8bee1f`). Renamed (`ed22223a`): `stop-task-missing` is now `stop-task-unknown`,
    `stop-rung-missing` is now `stop-rung-not-asked` [T-5708's report :97-104].
12. **The skill corrected** (`5a20aa87`, BUG-088): base's operating skill now says a refusal's rule shows on its own
    line, not only in `bonsai logs`.
13. **Climb 2, green** at `b18448cc` (17:05:11-17:11:37), sha256 `628463d2...14d7`; **climb 3, green** at `7afb9faa`
    (17:12:27-17:18:57), sha256 `691141c0...e839`: rungs 0 3.9 s, 1 0.8 s, 2 0.4 s, 3 62.8 s (4,807 seen, 9 skipped,
    0 failed), 4 52.4 s (2,125 seen, 3 skipped), 5 0.04 s; clean tree, nothing to Bless [the log :173, :186;
    `.bonsai/local/ladder/T-5708.json`; `M-2026-10-20-504c5465`].
14. **The stop gate let it end** (17:19:39, `stop-proven`) [the log :196].

### 4. The verifier

A fresh Opus verifier, given the task, the diff and its proof but not the builder's report
[T-5707's report :66], **passed** it (`M-2026-10-20-57b77ed9`, 17:25). It ran each check itself in scratch copies:
every block prints its rule; at `ecb3ac1c` (before the rename) the safety test fails on the old names, which shows the
rename was needed; the renames are complete (the reference page regenerates byte for byte); the error words are the
same on 13 kinds of error; the new test fails 12 tests on the old code; the diff stays inside the task's files; no test
weakened. Its three risks, not blocking, are rc.2's item 6 [`T-5710-rc2-fixes.md:31-32`]. Bonsai refused it twice,
since a verifier changes nothing: a `git init` (17:21) and a redirect writing a test file into its copy (17:24)
[the orchestrator's log :3597, :3640, rule `role-writes-nothing`].

### 5. The landing

CI run 39417827040 on `7afb9faa`: every job green, the two Windows jobs among them. `main` fast-forwarded
`ecb65520..7afb9faa` and pushed; the branch deleted on GitHub; the worktree removed. The landing's proof: `git.sha`
`7afb9faa`, sha256 `691141c0...e839`, green, its log record naming `/usr/local/bin/bonsai` [T-5707's report :70-73;
the log :186].

### The stages, reached or probed

| Stage (note 5.7.7, item 2) | How | Evidence |
|---|---|---|
| the task file and its moves | reached | `ecb65520`, `8fdf63d2` |
| a worktree | reached | `Bonsai-T-5708` in every log line |
| the builder as a named session | reached | the log :2 (`role: builder`, `task: T-5708`) |
| a protected file refused outside the grant | **probe 2** | the log :68 |
| a protected file changed under the grant | not reached: the task granted no protected path | `records/tasks/T-5708-...md`, `bonsai.allows` |
| the ladder, a new test failing first | reached | climb 1 red, climbs 2 and 3 green |
| its Bless | not reached: no floor to raise in this task (`bless: []`) | T-5708's report :108 |
| the stop gate blocking a red stop | **probe 1** (and once more while waiting) | the log :60, :126 |
| the recorder's log | reached | 198 records |
| the tables by `check --write` | reached | `8fdf63d2` |

**Correction, 25 Oct (5.7's end verifier, T-5741, finding F2; the director's word: the gaps stand, with this note).**
"The task file and its moves" overstates it: T-5708's file was written `running` (`ecb65520`) and then `done`
(`8fdf63d2`); no `todo` or `verify` was written, and the log holds no record of a task-file edit (the orchestrator's
shell writes carry no target) and the run report no tasks-table snapshot. The two stages "not reached" stay so in the
trial. Later, on the installed 1.0 and outside the trial: a protected file changed under a grant (`T-5706`,
`packs/base/bonsai/pack.yaml`, granted by its `bonsai.allows`; landed `4662da5e` on a green climb, result sha256
`70c1f115...bb66ba`), and floors raised to a climb's counts (`T-5740`, `bonsai.yaml`'s ratchets 5067 and 2341,
`97a04e18`, the tightening accepted with no person).
| an ask to you and your answer | reached, twice | the asks file |
| the verifier | reached | `M-2026-10-20-57b77ed9` |
| the landing rule and CI | reached | CI run 39417827040 |

## What was rough

"rc.2" means an item on `records/tasks/T-5710-rc2-fixes.md`, fixed before rc.2 is cut. "After 1.0" means a row in
`records/bugs.md`, left for after the release. BUG-092 to BUG-096 are the rows the orchestrator adds with this task's
records.

### Noted during the trial (T-5707's report :86-89)

1. **`CLAUDE.md` says a named session starts through `claude-here`**, but that launcher refuses any folder outside the
   scratch area, so no real named session has ever started that way [T-5707's report :40-45; `CLAUDE.md:59`].
   **rc.2, item 7** [`T-5710-rc2-fixes.md:33-35`].
2. **You could not see the named session** in your Claude Code agent list: it is a separate `claude -p` process. You
   were told to watch it through `ps`, `git log`, the run report and the stream [T-5707's report :56]. **rc.2, item
   7:** the rewritten named-session line in `CLAUDE.md` says how to watch one (`ps`, `git log`, the run report, the
   stream); a view of it in `bonsai agents` goes with item 10's filters [`T-5710-rc2-fixes.md:33-35`, `:38-40`].
3. **Waiting on a question spends stop-gate blocks.** A named session that ends its turn to wait is a stop, and
   the gate blocks it; Claude Code allows eight blocks in a row. The builder waited by making foreground checks instead
   [T-5708's report :91, :146-147; the log :126]. Note 5.7.7 item 4 expected exactly this to be recorded
   [`design/plan-5.md:7931`]. **After 1.0**, BUG-092: the gate could let a stop pass while the running task has an
   open question. That changes how the gate behaves, which is too big for a release candidate.
4. **Answering by option was too long to type.** `--choice` needs the option's whole text, so you typed `--words "B"`
   twice, and the records keep `choice` empty [T-5707's report :47-48; the asks file, records 3-4]. **rc.2, item 11**
   [`T-5710-rc2-fixes.md:41-44`].
5. **`source: "agent"` on your answers** reads as if an agent answered. It names who filed the question; who answered
   is `by: terminal` [T-5707's report :68, citing `internal/format/ask.go:15`]. **Nothing in code:** one sentence in
   the asks' docs, written with rc.2 item 11, which rewrites the same answer path.
6. **The orchestrator missed your second question** until you said "done": its watch ended at the first
   [T-5707's report :52-53]. **Nothing in code:** the orchestrator now watches until every open question of the task
   is answered (this task's run report); if it happens again, it becomes a `workflow:correct` case.
7. **The builder skipped its "audit first" stop.** It read the brief's ordered steps as go straight on, and a named
   session cannot stop before its ladder is green anyway [T-5708's report :35-38]. **Nothing in code:** the brief
   rules of 21 Oct say a named session goes straight on, its ladder being its gate.

### Found while gathering this page's evidence

8. **The named session loaded your own connectors**: 6 servers, among them mail and calendar, 65 tools in all
   [`stream.jsonl`, the `init` line]. A `claude -p` builder takes on every connector of your account. It used none of
   them [the log: its tool calls are Bash, Edit, Read and ToolSearch only]. **rc.2, item 7:** the named-session launch
   line tries Claude Code's `--strict-mcp-config` with no servers. rc.2's work has still to prove that this also keeps
   out your account's connectors. BUG-093.
9. **The ladder counted one of those connectors' processes as leftovers in the checkout**: three processes, a
   connector server started through npm and its shell and node [`.bonsai/local/ladder/T-5708.json`, `leftovers`]. They
   belonged to the session's own Claude Code, not to anything the builder left behind. **Nothing more:** it goes with
   8's fix; a line is added to BUG-029, the row about the `leftovers` description [`records/bugs.md:39`].
10. **A wrong line in T-5707's report:** "rewritten by a fetch" for the home's cache, where the lists show no change
    (above, "In the Bonsai home"). A records error, not Bonsai's. **Placed:** noted in this task's run report; past
    run reports stay as they are.
11. **Three clocks to line up:** `.bonsai/sessions.md`, the log and the asks keep UTC; the run reports and the
    launcher keep local time [`.bonsai/sessions.md:430`, 11:17; `start.txt`, 16:47:57]. This page converts them by
    hand. **After 1.0:** UTC in the log and the tables is the formats' rule (`design/contract.md`). A local-time view
    would change only how they are shown, so it is for later. A change to a format comes to you first.
12. **The builder's climbs took three times their rungs**: 6.4 and 6.5 minutes of wall time against about 2 minutes
    of rungs, while the set-up climb on `main` took 2.2 minutes for 2.0 of rungs [`T-5708.json` and `T-5707.json`,
    `started`/`finished` and `duration_ms`; T-5708's report :108, "not looked into"]. **After 1.0 unless it recurs**,
    BUG-094: not yet explained; the climbs in rc.2's work will show whether it repeats.

### Gaps in the trial's own records

- **No tasks-table snapshots** in T-5707's report at each task move, which note 5.7.7 item 3 asked for
  [`design/plan-5.md:7919`; T-5707's report, none]. The moves can still be read from the commits.
- **The home listing covers `cache/` only.** It omits `install.json`, the machine folder of Bonsai's own repo, and
  whether a personal index exists, which note 5.7.7 item 5 asked to list by kind [`design/plan-5.md:7941`; the two
  lists].
- **The first two climbs survive only as log lines** (`T-5708.json` holds the newest); their rung times are in
  T-5708's report alone [the log :48, :173].
- **`.bonsai/sessions.md:429` files the proposals' auditor (16:37-16:41) under T-5705**, though its message comes from
  `auditor@T-5707` [`M-2026-10-20-392e067a`]. Its records in the orchestrator's log carry no task: T-5707 did not
  read `running` until 16:47 [the orchestrator's log :3293 on; `ecb65520`]. **After 1.0**, BUG-095.

### Already placed

- BUG-014 (proposal A): **rc.2**, item 1, must land before 1.0 [`T-5710-rc2-fixes.md:23-25`].
- The changelog line for this feature (note 5.7.8): **rc.2**, item 3 [`T-5710-rc2-fixes.md:28`; T-5708's report
  :127-128].
- The verifier's three risks: **rc.2**, item 6 [`T-5710-rc2-fixes.md:31-32`].
- The verifier refused a scratch repository: BUG-013 [`records/bugs.md:25`], and the scratch-folder piece after 1.0
  (`records/tasks/T-5793-scratch-folders.md`).
- BUG-088 and BUG-089: fixed by this trial [`records/bugs.md:114-115`].

## The README's dry run

**Who, and on what.** A builder agent (Sonnet) followed the README's "A first project" steps exactly as written, with
the installed rc.1 (`1.0.0-rc.1`, commit `73919ddc`), in a scratch folder with a scratch Bonsai home, from 11:44:54 to
11:51:32 on 21 Oct, about 7 minutes of work. A runner was tried first and refused at step 1, because a runner may not
make a repository (BUG-013, `records/bugs.md:25`). [`M-2026-10-21-8afc2b72`]

| Step | What happened | Exit | Source |
|---|---|---|---|
| 1. make a folder, `git init` | worked; git printed its long hint about the branch name | 0 | the report, step 1 |
| 2. `bonsai init` | no `base-v*` tag exists yet, so it used `main`'s commit, which the README gives no way to find. The preview with no terminal ended in a refusal, `needs-yes`, writing nothing; with `--yes` it wrote `bonsai.yaml`, `CLAUDE.md`, `.claude/settings.json`, `.bonsai/`, and a personal memory index in the scratch home | 4, then 0 | `step12.out:75-81`, `:151`, `:196` |
| 3. `bonsai check`, `bonsai status` | "no findings" plus the plugin warning; no problems, as the README says | 0 | `step3.out:3`, `:13`, `:45` |
| 4. open Claude Code, `bonsai update --yes` | the session ran; update said "nothing to change", and the `base` plugin still read "waiting" | 0 | `step4.out:4`; `step4b.out:23-26` |
| 5. ask Claude Code for an edit | first refused by Claude Code's own permission setting (no one to grant it in a scripted session; not Bonsai's guard); with edits allowed, `README.md` got its line and the log recorded the guard allowing it (`not-protected`) | 0 | `step5.out:3`; `step5b.out:3`; the report, step 5 |
| 6. a rung, a task, `bonsai ladder` | refused `config-looser`, with its next step, as the README warns. The agent's own `bonsai update --yes` was refused as a person's command. Once the director ran `bonsai update --yes` in their own terminal (11:50, as the README says it must), the climb went **green**: 1 of 1 rungs, 0.1 s, and `check --write` wrote both tables | 4, then 0 | `step6.out:31-56`; `step6b.out:17-42` |

**What it could not see.** Claude Code's question about trusting the folder, and the `base` plugin's install, never
happen in a scripted session. So steps 4 and 5 are checked in the director's own trial [the report, "Launcher versus a
new user"; `step4b.out:26`].

**The 13 proposed fixes** [the report, "README fixes"; line numbers are the README's at that commit]:
1. Step 1: use `git init -b main`, so git prints no branch-name hint. *rc.2, with the README work (T-5711).*
2. Step 2: say how to find `main`'s commit (`git ls-remote ... refs/heads/main`) until a `base-v1.0.0` tag exists. *rc.2, with the README work (T-5711).*
3. Step 2: say that the preview with no terminal ends in a refusal (exit 4) and writes nothing. *rc.2, with the README work (T-5711).*
4. Step 2: list everything `init` writes: `STATE.md`, the two tables, `.bonsai/.gitignore`, and the personal memory index in Bonsai's home. *rc.2, with the README work (T-5711).*
5. Step 3/4: make the README and `check`'s next step agree on `bonsai update` with or without `--yes`. *rc.2, with the README work (T-5711).*
6. Step 4: say what to look for ("installed" against "waiting"), and what to do if it still says waiting. *rc.2, with the README work (T-5711).*
7. Step 5: say that the question before an edit is Claude Code's own, and that Bonsai's guard is a separate check. *rc.2, with the README work (T-5711).*
8. Step 5: mention `check`'s warning that the tables lag, and `bonsai check --write`. *rc.2, with the README work (T-5711).*
9. Step 6: say where a task file goes (`work/tasks/`), how it is named, and where its template is. *rc.2, with the README work (T-5711).*
10. Step 6: show the one-rung example that passed (`test -s README.md`, `ladder_floor: [1]`). *rc.2, with the README work (T-5711).*
11. Step 6: `update` leads with "nothing to change" though it has just recorded the looser settings; it should lead with what it recorded. *rc.2, Bonsai's words.* (The README's half, a line explaining the output today, goes with T-5711.)
12. Step 6: tell a fresh machine to set `git config user.name` and `user.email` first. *rc.2, with the README work (T-5711).*
13. `check`'s note names a plan step ("from step 5.6") a new user cannot know, and the ladder prints its result's path in full rather than relative. *rc.2, Bonsai's words.*

Items 11 and 13 hold the three changes to Bonsai's own words (13 holds two). The other eleven items are README
lines, and item 11 carries a README line too.

**The launcher's `remote` finding.** A session started from inside another Claude Code session records that
session's id in the log's `remote` field, taken from its environment. A person in a plain terminal would see it empty
(not tested). **After 1.0**, BUG-096 [the report, step 5 and "Launcher versus a new user"].

## Sources

The run reports `records/runs/R-2026-10-20-T-5707-trial-on-rc1.md` and `R-2026-10-20-T-5708-stopgate-names-its-rule.md`;
the tasks `records/tasks/T-5707-trial-on-rc1.md`, `T-5708-stopgate-names-its-rule.md`, `T-5710-rc2-fixes.md`,
`T-5793-scratch-folders.md`; `records/bugs.md`; `design/plan-5.md` notes 5.7.7-5.7.9; `CLAUDE.md:59`; `bonsai.yaml:110`;
the messages `M-2026-10-20-392e067a`, `M-2026-10-20-306324a4`, `M-2026-10-20-2c8bee1f`, `M-2026-10-20-504c5465`,
`M-2026-10-20-2c48f516`, `M-2026-10-20-57b77ed9` (`bonsai message --show <id>`); the asks `agent:h-639ab913f88a`,
`agent:h-ae479b969f78`; in the main checkout's `.bonsai/local/`: `asks/2026-10-20.ndjson`,
`log/s-64096e53-5e58-446c-9441-188608e138e9.ndjson`, `log/s-f7c1ed9c-3905-4e3a-bbb1-ef6b7a3bcc61.ndjson`,
`ladder/T-5707.json`, `ladder/T-5708.json`; `.bonsai/sessions.md`, `.bonsai/tasks.md`; in `~/bonsai-checks/T-5708/`:
`stream.jsonl`, `files-before.txt`, `files-after.txt`, `start.txt`, `exit.txt`, `stderr.txt`; the commits
`67fa1ea3`, `ecb65520`, `6ef7e56d` to `7afb9faa`, `8fdf63d2`; CI run 39417827040; the commands `bonsai check` and
`ls -a ~/.bonsai/` (21 Oct); for the dry run, the message `M-2026-10-21-8afc2b72`, `~/bonsai-checks/T-5709/step12.out`
to `step6b.out` and the scripts in `~/bonsai-checks/scripts/T-5709/`, and `README.md`'s "A first project".
