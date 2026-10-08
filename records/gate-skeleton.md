# The walking skeleton's gate report

- Date: 9 Oct 2026
- For: Rohan's gate (spec §14, step 4). The skeleton's last fresh verifier read it first, against the run reports.
- Written by: part 6's builder (Opus), from the run reports and today's measurements; the builder re-measured nothing.
  The skeleton's last fresh verifier (Opus, 00:49-01:03) read it and passed the skeleton and this report; its results
  are carried in here, and `records/runs/R-2026-10-09-gate.md` holds its report (cited below as `verifier`).
- Bonsai: `main` at `b18f444`; its Go code is `54b4fdd`'s, only Markdown changed since (gate 00:27).
- Sources are cited as `<report> <log time>`: `hook 19:41` is `records/runs/R-2026-10-08-hook.md`, its log entry at
  19:41; `gate 00:35` is today's `records/runs/R-2026-10-09-gate.md`. The short names: setup, plan-review, formats
  (part 0), clear-out (part 1), test-pack (part 4a), reader (part 2), engine (part 3), hook (part 5), plugins (part 4b),
  gate (part 6). Section 6 lists them.

## 1. For Rohan, in plain words

**What the skeleton was for.** Before you choose how much of Bonsai to build, its risky pieces were built on throwaway
copies of projects, on WSL and on Windows, and measured. Three things had to hold.

**What it showed.**
- **Your edits are kept.** Bonsai writes a project's files and updates them later. When an update would replace a file
  you changed, it stops, writes nothing and prints the command to keep your copy; you ran that in the PowerShell console
  and the printed command worked. An update with nothing new changes no byte (run again today, five times on each side).
- **A pack arrives as a Claude Code plugin at an exact commit.** Two projects and a worktree beside one of them each
  loaded their own commit (A, B, B, then A again), on WSL and on Windows, with no login.
- **The guard fails closed.** It refuses an edit of a protected file and allows other edits. It also blocks when Bonsai
  is missing, crashes, runs too long or is not on the PATH: you saw all four on Windows; agents saw them on WSL, where a
  145 KB write with Bonsai missing was blocked too.

**What it did not show, or left open.**
- Hand check a was run by a Sonnet agent on your word, in scripted sessions (no one typing). `/agents` is gone from
  Claude Code 2.1.294, so nobody read the pack's answer inside an ordinary typed session. On Windows no person opened a
  session in the pack's folders, so the Windows trust question and a Windows background session were not seen.
- Before a pack can install, Claude Code must register the project's marketplace (where it finds the pack), and it
  does that only in a session in a trusted folder. A person's first session there does it; on Windows an agent did it
  in a scripted session, with no one there, by handing Claude Code the project's own entry.
- An update that changes a hook line is refused; the switch that allows it comes in step 5.1.
- Undoing a link with a plain `git revert` gives back exactly the project as it was (checked again today, with the
  plugin step in place), but only where the plugin was never installed. Where it is installed, Claude Code keeps its
  own record of the install outside git, which a revert would leave; that was not tried.
- The studio's half of check 7 waits for step 7.

**The last check.** A fresh verifier re-ran the tests on both sides, part 3's checks on the final build, the revert and
the guard on WSL, and read the stop lines, Mimas and the studio's repo. It passed the skeleton and this report; its
results are in sections 2 to 4.

**Hours.** The whole skeleton, parts 1-6 with the last verifier, took 6 h 17 min of agent runs, against 30-47 h
estimated (stop line 61 h); parts 1-5 alone took 5 h 36 min against their 26-41 h. Windows-only failures cost about 18
minutes (stop line 8 h). No option round was asked of you. The verifier found no change to Mimas (its last commit is
from 5 Oct) or to the studio's repo (last commit 8 Oct, before part 1). No stop line is crossed, by its judgment.

**A sample eval.** A three-case `claude plugin eval` of the test pack ran in about half a minute for under 3 cents. The
one case only the pack can answer passed with it (2 of 2) and failed without it (0 of 2). It shows how an eval runs, not
that packs help.

**The three paths.** The spec's hours after the gate (§14), and the same hours times the skeleton's measured ratio:
its 6 h 17 min against its 47 h estimate is 0.1337; against its 30 h, 0.2094 (section 2.2 has the arithmetic).

| Path | The spec's hours | Times 0.1337 | Times 0.2094 |
|---|---|---|---|
| (a) the full Bonsai 1.0 | 139-218 h | 18.6-29.1 h | 29.1-45.6 h |
| (b) the smaller cut | about 121-188 h | 16.2-25.1 h | 25.3-39.4 h |
| (d) 1.0 with its own screens | about 169-268 h | 22.6-35.8 h | 35.4-56.1 h |

**What limits the ratio.**
- It counts only the agents' own runs: not the orchestrating session, not your sittings, not waiting. The spec's
  estimates are session hours, and its 1.6-times record was measured from plan to merge, waits included.
- It comes from six parts over two days, and it swung by part from 0.06-0.09 (the engine) to 0.32-0.53 (the hook
  path). The parts that dealt with Windows and Claude Code ran highest; step 5 has more of that work (the guards, the
  ladder on Windows, the packs).
- It includes the last verifier, which sent back only wording in this report and one line of the spec, no code.
- Path (d)'s screens are design work (30-50 h, the spec's guess); nothing in the skeleton sampled that kind of work.

**The choice is yours:** (a), (b) or (d). Sections 2 to 4 hold every number and where it comes from.

## 2. Spec §15's list, item by item

### 2.1 Hours per part against its estimate

Minutes are the sums of each run report's "Runs" rows (builder, verifier, audit and hand-check runs), copied by a Haiku
run and read by the orchestrator (gate 00:28); part 6's are the last verifier's tally (verifier). Ratio = minutes
spent / estimated minutes, high estimate first.

| Part | Estimate | Runs (minutes) | Spent | Ratio |
|---|---|---|---|---|
| 0. The formats (own stop line, 13 h) | 6-10 h | builder 31, verifier 11, fix round 2, verifier 1 (formats) | 45 min | 0.075-0.125 |
| 1. The clear-out | 2-4 h | builder 11, verifier 6 (clear-out) | 17 min | 0.071-0.142 |
| 2. The reader | 7-10 h | builder 52, verifier 22, fix round 10, verifier 5 (reader) | 89 min | 0.148-0.212 |
| 3. The engine | 10-15 h | builder 50, Haiku audit 1 (engine) | 51 min | 0.057-0.085 |
| 4. Packs as plugins (4a + 4b) | 4-7 h | 4a: builder 12, Haiku audit 1 (test-pack); 4b: builder 62, hand check a by Sonnet 8 (plugins) | 83 min | 0.198-0.346 |
| 5. The hook path | 3-5 h | builder 49, verifier 21, fix round 21, verifier 5 (hook) | 96 min | 0.320-0.533 |
| **Parts 1-5** | **26-41 h** | | **336 min = 5 h 36 min** | **0.137-0.215** |
| 6. Measure and report | 4-6 h | measurements, Sonnet 8; bookkeeping, Haiku 1; eval, Sonnet 2 (gate 00:35, 00:28, 00:39); this report's builder, Opus 11 (00:36-00:47); the last verifier, Opus 14 (00:49-01:03); this report's fixes 0 (00:48) and 4 (01:05-01:09) | 41 min | 0.114-0.171 |
| **The skeleton, parts 1-6** | **30-47 h** | | **377 min = 6 h 17 min** | **0.1337-0.2094** |

Part 6's runs sum to 36 minutes before the last fix round; the verifier gives 36-37 (the 00:48 fix as 0 or 1 minute),
and this report uses 37. Outside any part (not in the skeleton's tally): the setup 19 min (setup), the plan's review
and fixes 24 min (plan-review). Rohan's sittings and the orchestrator's own session are not counted (section 4).

### 2.2 The ratio applied to step 5's estimates

The skeleton's whole actual, all six parts with the last verifier: 336 + 41 = 377 minutes (6 h 17 min), the last
verifier's 373 (about 6 h 13 min) plus this report's last fix round, 4. Against both ends of the skeleton's 30-47 h:

- 377 / (30 x 60) = 377 / 1800 = 0.2094
- 377 / (47 x 60) = 377 / 2820 = 0.1337

A second line, like for like: parts 1-5's 336 minutes against their own estimates (2-4 + 7-10 + 10-15 + 4-7 + 3-5 =
26-41 h): 336 / 1560 = 0.2154; 336 / 2460 = 0.1366.

Applied to the spec's hours after the gate (§14, "Path (a) after the gate" and "Totals"); each range's ends times each
ratio, rounded to 0.1 h here:

| Path | Spec | x 0.1337 | x 0.2094 | x 0.1366 (parts 1-5) | x 0.2154 (parts 1-5) |
|---|---|---|---|---|---|
| (a) | 139-218 h | 18.6-29.1 | 29.1-45.6 | 19.0-29.8 | 29.9-47.0 |
| (b) | about 121-188 h | 16.2-25.1 | 25.3-39.4 | 16.5-25.7 | 26.1-40.5 |
| (d) | about 169-268 h | 22.6-35.8 | 35.4-56.1 | 23.1-36.6 | 36.4-57.7 |

The arithmetic for the six-part columns:
- (a) 139 x 0.1337 = 18.584; 218 x 0.1337 = 29.147; 139 x 0.2094 = 29.107; 218 x 0.2094 = 45.649
- (b) 121 x 0.1337 = 16.178; 188 x 0.1337 = 25.136; 121 x 0.2094 = 25.337; 188 x 0.2094 = 39.367
- (d) 169 x 0.1337 = 22.595; 268 x 0.1337 = 35.832; 169 x 0.2094 = 35.389; 268 x 0.2094 = 56.119

Mapping low end to low end and high to high (139 x 0.2094 and 218 x 0.1337) gives about 29.1 h for path (a)
either way. The spec's own figures for comparison: path (a) 222-349 h on the 1.6-times record; Bonsai 1.0 with the
skeleton 169-265 h (§14, "Totals"). The limits on the ratio are in section 1.

### 2.3 Each check, passed or not

Section 3, one row per check. In short: checks 1, 2, 3, 4, 6, 9 and 10 passed; 5, 8, 11 and 12 passed with stated
limits; 7 passed for Bonsai's half, and the studio's half has not run. The last verifier re-ran checks 1-6 (5 to its
session clause), 10, 11 on WSL and 12 itself, and passed them (verifier).

### 2.4 Windows-only failures and their hours

About 18 minutes in all, every one in part 5's fix round: the verifier's findings 1-3 (a directory stream in a middle
path segment, junctions, `\\?\GLOBALROOT` device paths), all Windows only (hook 19:18, 19:41, "Windows-only failures").
Parts 1, 2, 3 and 4b: 0 minutes (clear-out, engine 17:40, plugins 21:14). Part 6: none, the last verifier's runs
included (verifier).

### 2.5 Hook start-up, p50 and p95

Milliseconds, p50 / p95. Today: two rounds per side, n = 2000 per round on WSL, n = 300 on Windows (gate 00:35). WSL
ran at load average 0.6 and 1.5, with another person's Claude Code session and services running. The minimal Go hook
was rebuilt today (part 5's was not kept).

| What | WSL | Windows, Git Bash |
|---|---|---|
| Baseline, spec §15 (7 Oct; one figure per call, no p95) | a minimal Go hook 1.5 | Go 65 |
| Part 5's own, final build (hook 19:41) | `bonsai hook guard` 2.38 / 2.90 | the hook line 63.4 / 67.7 |
| Part 5's first build (hook 18:33) | 2.21 / 2.57; a minimal Go hook 1.34 in the same harness | 61.7 / 65.8 |
| Today, a minimal Go hook | 1.60 / 2.33; 1.45 / 2.05 | not measured |
| Today, `bonsai hook guard`, typical call | 2.49 / 3.48; 2.47 / 3.46 | without a shell: 12.3 / 14.3; 12.6 / 14.5 |
| Today, a refusal | 2.47 / 3.44; 2.49 / 3.48 | not measured |
| Today, a session's first call (it hashes itself) | 5.73 / 6.81; 5.72 / 6.73 | 17.5 / 19.2; 12.9 / 14.5 |
| Today, the hook line through the shell | `/bin/sh -c`: 3.16 / 4.29; 3.17 / 4.32 | 66.0 / 70.7; 66.3 / 70.5 |
| Today, Git Bash alone | | 34.7 / 38.1; 35.1 / 38.0 |
| Today, `bonsai.exe --version` | | 9.9 / 12.1; 10.2 / 12.0 |

Read against the baseline: on WSL the guard's typical call is about 1 ms over the minimal hook's 1.5 ms and under spec
§3's 5 ms, but a session's first call (5.73 and 5.72 at p50) is over 5 ms. Today's WSL p95 is about 0.6 ms above part
5's (3.48 and 3.46 against 2.90, 0.56-0.58 ms; section 5). On Windows the hook line is 66.0 and 66.3 at p50 against
the baseline's 65 and part 5's 63.4; Git Bash alone takes about 35 of it.

### 2.6 Fail-closed and minimal-PATH results

| Case | WSL (agents, scripted sessions through `claude-here`) | Windows, Git Bash (Rohan's hand check c, sessions through `claude-here.cmd`) |
|---|---|---|
| Edit of `protected.txt` | refused, 44 bytes kept (hook 18:33; re-run on the final build, 19:41; the last verifier, with its reason, 44 bytes kept) | refused with the reason and a next step (hook 20:11) |
| Edit of `free.txt` | done (18:33, 19:41; verifier) | done |
| `missing` | refused (18:33); a 144,976-byte Write, past the 64 KiB pipe buffer, refused, `big.txt` never made (plugins 21:53); the verifier: refused, "/bin/sh: 1: bonsai: not found" | refused; the launcher said "no bonsai on this session's PATH", the agent saw "bonsai: command not found" |
| `crash` | refused (18:33; verifier) | refused: "test fault crash ... the hook line's `\|\| exit 2` blocks the call" |
| `slow` | refused (18:33); the verifier: refused, "ran past its own 5s limit" | refused twice: "the guard ran past its own 5s limit, so it blocks the call" |
| `minimal-path` | refused (18:33); the verifier: refused, "bonsai: not found" | the Edit and a PowerShell call refused: "bonsai: command not found" |
| The verifier's own runs | 122 hook-line inputs plus 16 and 8 fault runs, 0 wrong allows (19:18); the 122 again after the fix round, 0 wrong allows (19:47) | first round: 5 wrong allows of 29, fixed in the fix round (19:18, 19:41); then 49 inputs through the hook line, 0 wrong allows apart from the documented admin share, and 9 fault runs, all blocked (19:47) |

The last verifier's WSL runs were a second runner on today's scratch builds: headless through `claude-here` in
`project-guard`, with part 5's flags and Haiku, $0.48 (verifier). A hook that times out does not block in Claude
Code, so the guard's own 5 s timer (half the hook's 10 s) is what blocks `slow` (hook 18:33, 19:18). The large Write
under `missing` was asked of Rohan on Windows and not run (hook 20:11); it ran on WSL only.

### 2.7 In which session kinds the pack's roles and skills loaded

The test pack's role `marker` answers which commit it comes from; its skill is `hello`.

| Session kind | WSL | Windows |
|---|---|---|
| Scripted (`-p`) | loaded: `project-a` "commit A", `project-b` "commit B", `project-a-worktree` "commit B", `project-a` again "commit A", by part 4b's builder at project scope (plugins 21:14) and again by hand check a's Sonnet agent with `-p --agent test-pack:marker` (plugins 21:53) | loaded: A, B, B, A by hand check a's Sonnet agent with `-p --agent test-pack:marker`, after each project's marketplace was registered by a `-p` session given the project's own entry through `--settings` (plugins 21:53) |
| Skills | today in `project-a`, `-p`: the session's start lists `test-pack:marker` and the skill and command `test-pack:hello`; the skill was called through the Skill tool; the marker answered "commit A" as a subagent and in an `--agent` session (gate 00:35) | the same, today (gate 00:35) |
| Background (`--bg`) | untrusted folder: refused, "Workspace not trusted" (plugins 21:14); `project-a` once trusted: "backgrounded", `claude logs` showed "commit A", stopped with `claude stop` (plugins 21:53). The record does not quote the `--bg` command line | not run: no Windows folder was trusted, and an agent cannot answer the trust prompt (plugins 21:47) |
| Interactive (typed) | Rohan opened sessions in `project-a`, `project-b` and `project-a-worktree`; `/agents` printed "The /agents wizard has been removed" (2.1.294); no trust question in the worktree. The marker's answer inside an interactive session was not read (plugins, second sitting) | not run with the pack. Rohan's Windows sessions were hand check c's, in `project-guard`: it is linked to the test pack at A (its `bonsai.yaml` lists it, `.claude/settings.json` turns on `test-pack@bonsai-project-guard-f004a1d0`, `test-pack/` is present), but its plugin was never installed there (the install step came with part 4b, after the sitting), and the marker was not asked there (verifier). Whether a trust question came is not recorded |
| A fresh worktree (`project-a-worktree` at B beside `project-a` at A) | loaded B in `-p` (plugins 21:14, 21:53); in Rohan's interactive session, no trust question | loaded B in `-p` (plugins 21:53) |

Hand check a's pass condition named `/agents`; with it gone, the orchestrator read the marker by starting a session as
the role instead (`--agent test-pack:marker`, plugins 21:30), and on Rohan's word (21:47) a Sonnet agent ran the check.
Not an option round: no choice was asked of him.

### 2.8 Plugin install and update time

| What | WSL | Windows |
|---|---|---|
| First install at project scope (plugins 21:14) | 11.3 s | 10.1 s |
| The same install again (plugins 21:14) | 0.95 s | not recorded |
| `update` with the install, first time (plugins 21:14) | 13.1 s | 12.7 s |
| `update --yes` at the same commit, five runs today (gate 00:35) | 1.04, 1.01, 1.03, 3.91, 1.00 s; median 1.03 | 3.75, 1.44, 1.44, 1.42, 1.44 s; median 1.44 (through `bonsai-here.cmd` and PowerShell) |
| `status` (offline) and `check` (asks Claude Code) (plugins 21:14) | 0.01 s and 0.93 s | not recorded |

Today's `update` runs printed "nothing to change" and "installed ... at 506205354b75"; the lock and
`.claude/settings.json` hashes were the same before and after, and the same on both sides; `check` found nothing
(gate 00:35).

### 2.9 Two projects, and a worktree beside its checkout, kept two commits

Yes, on both sides: `project-a` at A, `project-b` at B, `project-a-worktree` at B beside `project-a` at A, read in that
order and then `project-a` again, A, B, B, A; `bonsai check` "no findings" in all three (plugins 21:14 on WSL; 21:53 on
both sides). The first try, at local plugin scope, failed: the worktree loaded both commits and answered "commit A";
project scope fixed it (plugins 21:14).

### 2.10 The Claude Code version on each side

2.1.294 on WSL and on Windows (gate 00:25, 00:35). This is the first floor for `check`'s version warning (spec §7,
§15). The baseline (spec §15, 7 Oct) had WSL 2.1.292 and Windows 2.1.291. Go: WSL 1.25.9 (the module's toolchain),
Windows 1.26.2 (gate 00:35).

### 2.11 Binary size

| Build (gate 00:35) | Bytes | MiB |
|---|---|---|
| Linux, plain (a fresh build of `b18f444`) | 5,503,226 | 5.25 |
| Linux, stripped (`-trimpath -s -w`) | 3,813,560 | 3.64 |
| Linux, with the fault tag | 5,508,744 | 5,518 bytes over plain |
| Windows `bonsai.exe`, plain (the scratch build in use) | 5,786,112 | 5.52 |
| Windows, stripped | 4,079,104 | 3.89 |
| The WSL scratch build in use, `~/bonsai-checks/bin/bonsai` | 5,503,266 | |

The WSL scratch build carries no commit stamp (`vcs.revision`: it was built where Go recorded no git data), so its
commit cannot be proved from the binary; a fresh build of `b18f444` is 40 bytes smaller, with a different hash. The
Windows one is stamped `54b4fdd`, unmodified, and a fresh build from that tree is byte-identical (gate 00:35). `go tool
nm` on the plain build finds no Bonsai symbol with "fault" in its name; the tagged build has `internal/guard.testFault`.

The unstamped WSL scratch builds, plain and fault, match the last verifier's builds of `09cff74` function for function
(`go tool nm -size`: 276 and 278 Bonsai functions, identical names and sizes), so the hashes they log belong to the
final code in substance. Builds made in these WSL worktrees get no commit stamp at all; the verifier's Windows build of
`09cff74`, from a Windows-git clone, is stamped `vcs.revision 09cff74` (verifier). The missing stamp stays a step 5
finding (section 5, 5.2).

### 2.12 Lines of Go per part

Added / deleted, from `git diff --numstat` over each part's range ("from main at" to "on main at" in its run
report), split into tests (`_test.go`) and other code (gate 00:28; this builder re-ran the same command and got the
same figures).

| Part | Range | Tests | Other code |
|---|---|---|---|
| 0 | `c146972..3770d04` | 909 / 0 | 0 / 0 |
| 1 | `d37a5fb..17f2938` | 67 / 17,886 | 30 / 24,750 |
| 2 | `a63f9ac..3a1f195` | 2,140 / 502 | 3,595 / 16 |
| 3 | `8c83e0d..ee50971` | 1,334 / 9 | 3,938 / 28 |
| 5 | `1ba4abe..33a6122` | 1,427 / 4 | 1,438 / 29 |
| 4b | `2e67b51..54b4fdd` | 463 / 0 | 498 / 15 |

Part 4a (the test pack) has no Go. Part 1 removed 42,636 lines of the old product's Go. At `b18f444`: 67 Go files,
9,423 lines of code and 5,825 of tests, 15,248 in all (gate 00:28). The parts' net lines sum to 15,236; the other 12
are lines of the old `cmd/bonsai/main.go` that part 1's stub kept unchanged (the old product had 42,648 lines of Go at
`c146972`, 12 more than part 1 deleted; this builder's count).

### 2.13 The three-case eval

A sample of how an eval runs, not proof that a pack helps (spec §15). The results are from gate 00:39; the command line
and the totals are from the eval run's report as the orchestrator sent it to this builder.

- The run: Sonnet, 00:36-00:38, on WSL through `claude-here`, on a `git archive` copy of the test pack at D (`1d4f46f`)
  in `~/bonsai-checks/eval/test-pack/`; the test pack's repository untouched.
- The suite: `evals/<case>/prompt.md` and `graders/*.md` inside the plugin folder (the default eval folder), plain files
  that ship with the plugin. Graders deterministic only (regex, `tool_used`); no model grader.
- The command: `claude-here plugin eval ./test-pack --no-publish --trust-plugin --model haiku --judge-model haiku --runs
  2 --max-cost-usd 2 --json ...`, two runs per case with the plugin and two without it (a baseline arm).
- The first attempt, the target by name, was refused in 0.8 s at no cost: four installs named `test-pack` sit in the
  shared plugin cache. The second, the target by path, ran in 33 s, exit 0.

| Case | Prompt | Grader | With the plugin | Without | Delta |
|---|---|---|---|---|---|
| `marker-commit` | ask the marker role which commit is loaded | regex "commit B" | 1.0 (2 of 2) | 0.0 (0 of 2) | +1.0 |
| `hello-skill` | "Say hello from the test pack." | regex for the skill's line "hello from the test pack"; `tool_used: Skill`, reported for the plugin arm only and not scored | 1.0; the Skill tool used 2 of 2 | 1.0 | 0 |
| `control` | "What is 17 plus 25?" | regex "42" | 1.0 | 1.0 | 0, as meant |

- Totals: 3 of 3 cases passed, mean delta +0.333, $0.0275 (the agent runs; no judge cost), 32-33 s, 18 turns over 12
  runs.
- `hello-skill`'s prompt itself held the phrase the skill prescribes, so the baseline repeated it: a flaw in the case,
  not in the tool. Only `marker-commit` separates the arms: it asks a fact that only the plugin carries.
- Odd: the JSON reports `runsPerCase: 3` though two runs per arm ran; `--help` does not list the grader fields (the
  agent found them from `init --bare` and the binary's strings).
- Nothing was published; the user settings hash was unchanged; `~/.claude/plugins` held 143 entries before and after.
  Part 4b's 4,686 entries there (plugins 21:14) is the same tree counted whole; the orchestrator counted both ways at
  00:48 today: `find ~/.claude/plugins | wc -l` 4,686, and `-maxdepth 3` 143, both unchanged.

### 2.14 Whether the studio's half of check 7 has run

It has not. It runs when the studio links (spec step 7); the studio's readers task was cut on 8 Oct (spec §14's note;
plan, "The twelve checks" readings; `STATE.md`).

### 2.15 The guard's log, path and hash per side

| Side | Path the log names | SHA-256 the log names | Matches the binary (gate 00:35) |
|---|---|---|---|
| WSL, today | `~/bonsai-checks/bin/bonsai` | `8a9398e28974...3a6068` | yes; no commit stamp, but the final code function for function (2.11) |
| Windows, today | `%USERPROFILE%/bonsai-checks/bin/bonsai.exe` | `17ad3ce35c5b...f99013` | yes; stamped `54b4fdd` |
| Windows, hand check d (8 Oct, the `33a6122` build) | `%USERPROFILE%/bonsai-checks/bin/bonsai.exe` | `8c8463eb5e4b...` | the scratch build of then (hook 20:11) |
| Windows, hand check c's `crash` and `slow` sessions | `%USERPROFILE%/bonsai-checks/bin/fault/bonsai.exe` | `6123c0e3a83b...` | the fault build (hook 20:11) |
| WSL, the last verifier's check 11 | `~/bonsai-checks/bin/bonsai` | `8a9398e28974...` | today's build (verifier) |
| WSL, its fault sessions | `~/bonsai-checks/bin/fault/bonsai` | `4aadb0f888ea...` | the fault build; no commit stamp, the final code function for function (verifier) |

The hashes are given as the run reports give them (shortened).

## 3. The twelve checks

Spec §14's list with the plan's readings. "Passed" means the run reports record it passing. The last verifier (Opus,
00:49-01:03) re-ran checks 1-6 (5 to its session clause), 10, 11 on WSL and 12 itself, on builds of `09cff74`, and
passed them (verifier); part 3's scripted checks it re-ran from a copy of the scripts in its own scratch folder, with
its own plugin cache.

| # | What it asks (short) | Result | Where it was proved | Limit, or where the evidence is thinner than the words |
|---|---|---|---|---|
| 1 | `init` into the drifted copy: exit 2 without the required values; with them `bonsai.yaml` first (a comment on every line), the lock last; the project's own hook kept, the old absolute line gone, Bonsai's lines in shell form, deny rules with `never_edit`, a 40-character commit in the plugin wiring, the plain-words block, `status` agreeing | Passed | engine 17:40 (the builder's scripted runs over the drifted copy, a fresh repo and the studio clone); engine 17:42 (the orchestrator's read of the diff and its re-run of the Go tests, 703 passed); hook 19:18 (part 5's verifier: `init` at C writes the hook line as specified and keeps the project's hooks); verifier (re-run on a build of `09cff74`: pass, with the same command counts as part 3's logs) | The re-run differs from part 3's logs only in the plugin step's new lines (`waiting` in untrusted folders, with its next step, never changing the exit code; `check` adds a "not installed for this checkout" warning, exit 0), the commit ids, and one `bonsai.yaml` hash from a different workspace id. The hook line calls `bonsai` by name, while spec §7's note asks for a fixed path: step 5.3 |
| 2 | Nothing written holds an absolute path, the lock included; the lock's source is the pack's remote URL | Passed | engine 17:40; verifier (re-run on `09cff74`: pass) | Row 1's notes on the re-run |
| 3 | The lock is JSON with forward-slash keys and verifies after a CRLF checkout and under Git Bash | Passed | engine 17:40 (a Windows-git CRLF checkout under Git Bash); verifier (re-run on a Windows build of `09cff74`, stamped `vcs.revision 09cff74`, under Git Bash with a Windows-git CRLF checkout: pass) | None recorded |
| 4 | `init` again, and `update` to the same commit, change no byte | Passed | engine 17:40; gate 00:35 (five `update --yes` per side today: "nothing to change", the lock and `.claude/settings.json` hashes unchanged, `check` clean); verifier (`init` and `update` again on all three targets: no byte and no mtime changed) | On Windows, `project-a`'s git shows ` M .claude/settings.json` before and after today's runs: Claude Code's rewrite at part 4b's first install, not Bonsai's write |
| 5 | `update` to B with no terminal and no `--yes`: a preview naming every settings line with its sentence, nothing written, exit 4; with `--yes` one updated, one created; a hook-line change waits for `--allow-exec`; a new session loads B's roles by name; `check` no drift | Passed with a stated limit | engine 17:40 (A to B; C to D refused, exit 4, nothing written over four runs); hook 19:18 (verifier: D refused, exit 4, under 11 flag combinations); hook 20:11 (hand check b: the preview with its settings lines and sentences, on Windows); plugins 21:14 (the worktree's update A to B installed B, a new session answered "commit B", `check` no drift); verifier (re-run to its session clause on `09cff74`; C to D: exit 4 naming `--allow-exec` in each of four runs, no byte or mtime changed) | `--allow-exec` itself is step 5.1 (it exits 2 now). A mixed update (`--yes` while a hook line also changes) is avoided by commit D and still unsettled (plan, "Stale or in tension"). The new session was scripted (`-p`), not typed |
| 6 | One pack file edited, then `update --yes`: exit 5, nothing written, the file named; `--keep` applies the rest; C conflicts again; `--adopt` puts the project's copy in the cache | Passed | engine 17:40 (`--keep`, then `--adopt` at C); hook 20:11 (hand check b on Windows: "Stopped: 1 conflict (test-pack/guide.md): nothing was written", the `--keep` command pasted and working); verifier (re-run on `09cff74`: pass) | Rohan's lines do not show the exit code; row 1's notes on the re-run |
| 7 | Bonsai's format-1 reader and the studio's `yaml.mjs` format-1 mode reach the same outcome on every trick file (Bonsai's half: every trick file's format-1 outcome) | Bonsai's half passed; the studio's half not run | reader 16:41 (116 of 116, WSL and Windows); reader 16:47 (the verifier's re-check, both sides) | The studio's half waits for step 7 |
| 8 | A fetch from the public test pack at a commit, no login, WSL and Windows, read-only; roles and skills load; two projects at two commits, and a worktree at B beside its checkout at A, each load their own, sessions alternating; the session kinds recorded; `claude --agent workflow:builder --bg` starts the plugin's role | Passed with stated limits | plugins 21:14 (WSL, project scope: A, B, B, A; pinning with no login on both sides); plugins 21:53 (hand check a: A, B, B, A on both sides; WSL `--bg` "commit A"); gate 00:35 (the skill listed and called on both sides) | Hand check a was run by a Sonnet agent on Rohan's word, not by Rohan. `/agents` is gone in 2.1.294: the interactive kind rests on Rohan's three WSL sessions starting, and the marker's answer inside an interactive session was not read. No interactive session with the pack on Windows; `--bg` on WSL only; the Windows marketplaces were registered through `--settings`, not by a person trusting the folder. `--agent test-pack:marker` stands in for `workflow:builder` (the `workflow` pack is step 5.5). The `--bg` record does not quote its command line. "Read-only" is not shown by a test in the records: the fetch is a plain git fetch of the public URL with prompts off (engine 17:40) |
| 9 | (Rohan) One run in the PowerShell 5.1 console: ASCII output, the y/N question, a command pasted from an error message works | Passed | hook 20:11 (hand check b: plain ASCII, `Write these changes? [y/N]`, the pasted `--keep` command worked) | The record says "the PowerShell console"; its version is not written down |
| 10 | `go test ./...` and `go vet` pass natively on Windows and on `windows-latest`; no Windows-only skip without a named reason | Passed at `54b4fdd` and at `09cff74` | plugins 21:14 (Windows `plugins-src` at `54b4fdd`: 777 runs plain, 783 with the fault tag; WSL 773 and 779); plugins 21:24 (CI on `a403acc`, the state commit on `54b4fdd`: `test`, `windows`, `lint`, `govulncheck`, CodeQL green); CI on `b18f444`, read by the orchestrator at 00:48 today with `gh api repos/LastStep/Bonsai/commits/b18f444/check-runs`: `test`, `windows`, `lint`, `govulncheck`, `Analyze Go` (CodeQL) all completed, success; verifier (at `09cff74`: WSL plain 773 runs, 771 pass, 0 fail, 2 skip; with the fault tag 779, 777, 0, 2; Windows, `verify-final-src`, a Windows-git clone of a bundle of `gate`: 777, 776, 0, 1; tagged 783, 782, 0, 1; `go vet` clean on all four; format 1, 116 of 116) | Each skip has its reason: WSL `TestDecideOnACaseInsensitiveSystem` and `TestDecideOnWindowsForms` (Windows-only tests); Windows `TestDecideFollowsLinks` (symbolic links need a privilege; Linux and CI run it). CI on the pushed gate commit is read after the push |
| 11 | In a real session on WSL and on Windows (Git Bash): an Edit of a protected path blocked, others allowed; missing, crash and over-time each block; a minimal-PATH session blocks with a clear reason | Passed with a stated limit | WSL: hook 18:33, 19:41; plugins 21:53. Windows: hook 20:11 (hand checks c and d). Part 5's verifier: hook 19:18, 19:47. The last verifier, a second runner on WSL on today's builds: protected refused (44 bytes kept), free done, all four faults refused (verifier). Section 2.6 | On WSL both runners' sessions were scripted (no one typing). The minimal-PATH reason shown is the shell's "not found" (Bonsai cannot speak when it is not found). The large Write under `missing` ran on WSL only |
| 12 | A plain `git revert` of the link commit restores the project's old state with no Bonsai binary present | Passed with a stated limit | engine 17:40 (revert on the scratch copies, at `ee50971`); verifier (re-run with the plugin step in place: with no `bonsai` on the PATH, `git revert` of each link commit gives exactly the tree from before the link, for the drifted copy, the fresh repo (the empty tree) and the studio copy; `git status --ignored` empty) | Claude Code records a project-scope install outside git, in the plugin root's `installed_plugins.json` with the `projectPath` (seen for `project-a`, `project-b`, `project-a-worktree`); a revert of an installed checkout would leave that record. Untested, since the scratch targets stay `waiting` (verifier, N6) |

## 4. The four stop lines

| Line | Tally from the run reports | Crossed? |
|---|---|---|
| 1. Hours over 61 h for parts 1-6 | Parts 1-5: 336 min (2.1). Part 6: 41 min, the last verifier's 37 (measurements 8, bookkeeping 1, eval 2, this report's builder 11, its 00:48 fix 0 or 1, the last verifier 14) plus this report's last fix round, 4 (01:05-01:09). In all: 377 min, 6 h 17 min | No (the last verifier: about 6 h 13 min before the last fix round) |
| 2. Windows-only failures over 8 h | About 18 min, part 5's fix round (2.4); none in part 6, the verifier's runs included | No (the last verifier) |
| 3. More than two option rounds asked of Rohan inside the skeleton | None. Each report's tally says none; the `/agents` change was recorded as not an option round (plugins 21:30); the Sonnet agent for hand check a was Rohan's own word (21:47). The two questions he was asked on 8 Oct came before part 1: the hand-check sittings (plan-review 13:07) and the label fields (formats 14:27). The last verifier counts 0 | No (the last verifier) |
| 4. Any change to Mimas or to the studio's repo | None. Read by the last verifier: the studio's checkout's last commit is 8 Oct 12:01, before part 1, its tree clean and even with origin, nothing in `tools/` since; Mimas, read only with Windows git: last commit 5 Oct, none since 8 Oct (verifier). Part 3 cloned the studio's repo into scratch (`--no-hardlinks`, origin removed, git hooks off) for scripted runs only (engine 17:40); the setup, before the skeleton, ran in a session started in the studio's checkout and wrote nothing of the studio's (setup). The user settings files kept their hashes from before part 3 to today (WSL `7b515457...a025a7`, Windows `2b6295c1...4ff6c9`; engine 16:49, gate 00:25 and 00:35) | No (the last verifier) |

What the hours do not count. The rule counts builder and verifier runs (the reports also count their Haiku audits and
the Sonnet hand-check run, and so does this tally). It does not count the orchestrator's own session time, Rohan's two
sittings, or waiting for CI. For context, 8 Oct's work spans 10:28 (the setup builder's start, its Runs row; its log
has no times) to 21:58 (the last log entry, plugins): 11 h 30 min elapsed. The skeleton's part, from part 1's first log
entry (clear-out 14:34) to 21:58: 7 h 24 min elapsed. Part 6 began today at 00:25 (gate); the last verifier
ended at 01:03.

A note on the rows: some start a minute or two before the log line that briefs the run (plan-review: builder 13:02,
briefed 13:04; hook: builder 17:44, briefed 17:46; verifier 18:57, briefed 18:58; plugins: builder 20:12, briefed
20:14). The rows are what the rule counts.

## 5. Findings step 5 inherits

Grouped by the part of step 5 that must settle them.

**5.1 Formats and engine to 1.0**
- `--allow-exec`: a hook-line change is only refused now (exit 4; the flag exits 2), and a mixed update (`--yes` while a
  hook line also changes) is unsettled between §6's "all or nothing" and check 5's words (engine 17:40; plan, "Stale or
  in tension").
- A plugin's move to a new commit is not counted as running code, though spec §5 lets plugins carry hooks (hook 19:18,
  item 8).
- With the lock deleted, `init --yes` links again and writes a changed hook line, its preview saying a link is consent;
  `update` now refuses that case (hook 19:41, 19:47).
- First-time trust: Claude Code registers a project's marketplace only in a session in a trusted folder, so the install
  step reports `waiting`, with a next step, until one has (plugins 21:14). A person's first session there does it; on
  Windows the agents did it headless, passing the project's own entry through `--settings` (plugins 21:47, 21:53).
- Unlink and revert: Claude Code records a project-scope install outside git, in the plugin root's
  `installed_plugins.json` with the `projectPath`; a `git revert` of an installed checkout would leave that record
  (untested; verifier, N6).
- Claude Code rewrites `.claude/settings.json` in its own key order at the first project-scope install; Windows
  `project-a` showed ` M .claude/settings.json` in git before and after today's `update`s (plugins 21:14, 21:53;
  gate 00:35).
- Local plugin scope leaks across worktrees (a worktree reads the main checkout's `settings.local.json`, and `install
  --scope local` there writes it); Bonsai uses project scope only (plugins 21:14).
- `check`'s version warning: Claude Code 2.1.294 on both sides is the first floor measured (2.10).
- `status --json`'s `formats` stays null (a read-only format cannot be listed); no `error` object (`STATE.md`).
- For a later formats set: a quoted `"format":` key followed by a tab still dispatches as format 0; a file with lone-CR
  line endings gets `quote-this-value`, whose next step does not help (reader 16:47).

**5.2 Recorder, logs, asks**
- The log's `bonsai_path` and `bonsai_sha256` names are Bonsai's own, outside the log schema; `input_hash` is null until
  the salt (hook 18:33; `STATE.md`).
- The binary's hash is logged once per session file (hook 19:18, item 6).
- A scratch build without a commit stamp: builds made in these WSL worktrees get no `vcs.revision`, so a logged hash
  cannot be tied to a commit from the binary alone. Today's WSL builds match the verifier's builds of `09cff74`
  function for function; the Windows builds, from Windows-git clones, are stamped (gate 00:35; verifier, N8).

**5.3 Guards**
- How the hook line finds `bonsai`: by name on the PATH, which a settings `env` could redirect; spec §7's note asks for
  a fixed path, §3 and check 2 for a name; Rohan's 8 Oct requirement is that hook lines cannot be redirected by files an
  agent may edit (engine 17:40; hook 19:18; plan, "What still links" item 5).
- A `.git` entry in the session's starting subfolder makes the guard read the project as unlinked; `bonsai.yaml` is read
  from the working tree; shell `rm` or `mv` of `bonsai.yaml` or the lock is not judged (hook 19:18).
- Windows: a dangling junction into a protected folder is allowed (the write through it fails); a junction swapped
  between check and write is not caught; the admin share `\\localhost\C$` is a documented limit (hook 19:18, 19:47).
- WSL start-up: the guard's p95 today is about 0.6 ms above part 5's (3.48 and 3.46 against 2.90), measured under load;
  the minimal Go hook also read higher than in part 5 (1.60 and 1.45 against 1.34). A session's first call, which hashes
  the binary, is 5.73 and 5.72 ms at p50, over spec §3's 5 ms (gate 00:35; hook 18:33, 19:41).
- The large-Write fail-closed test (a payload past the pipe buffer under `missing`) ran on WSL only (plugins 21:53;
  hook 20:11).
- Part 5 tried `golang.org/x/sys` and dropped it: about 3 ms more on every Windows start (hook 19:41). Spec §3 plans it
  for job objects (5.4).

**5.4 Ladder runner**
- Not measured by the skeleton (spec §15's "Not measured yet"): Windows job objects, and Windows Defender's first-run
  scan of a new binary.

**5.5 Packs**
- `/agents` is gone in Claude Code 2.1.294; a role is read by starting a session as it (`--agent <plugin>:<role>`).
  Spec §17 step 6 still names `/agents` and one sitting (plugins, second sitting; `STATE.md`).
- The plugin carries no version field: the version shows only as the commit folder, the commit's first 12 characters
  (gate 00:35; plugins 21:14); `claude plugin validate` warns only of the missing `version` (test-pack 15:09).
- Not seen in the skeleton: the pack's role answering in an interactive session, a Windows `--bg` session, the Windows
  trust prompt (2.7).
- Claude Code's own writes outside the scratch folders: session transcripts, folder-trust entries, and Claude.ai's
  plugin sync rewriting `~/.claude/plugins/synced/<account>/.marketplaces.json` (392 to 391 bytes) and
  `.last-complete-round` even with `CLAUDE_CODE_PLUGIN_CACHE_DIR` set. The user `settings.json` files never changed
  (plugins 21:14, 21:20; gate 00:35). The sync rewrote `.marketplaces.json` again during the last verifier's
  `minimal-path` session (00:56); the plugin tree held 4,686 entries before and after (Windows 3,886) (verifier, N9).
- `claude --agent workflow:builder --bg` waits for the `workflow` pack.
- `claude plugin eval` publishes its report to claude.ai unless given `--no-publish`; unattended it needs
  `--trust-plugin`, and a path target when several installs share a name; its JSON's `runsPerCase` read 3 for two runs
  per arm; `--help` does not list the grader fields (gate 00:39).

**5.6 Machine pieces**
- Nothing new from the skeleton.

**5.7 Release**
- Go 1.25.9 (the `toolchain` line) has standard-library vulnerabilities that govulncheck lists and our code does not
  reach; a later 1.25.x patch fixes them, a one-line bump (clear-out 14:47; `STATE.md`).
- CI's `lint` does not use the fault tag (`go vet` and the tests do) (hook 19:18, item 9).
- The fault switch stays out of a normal build: no fault symbol in today's plain build (2.11).
- Wording: `README.md` describes the designed product in the present tense; `CONTRIBUTING.md`'s toolchain note holds
  from Go 1.21; `.gitattributes` keeps old-product patterns; `CHANGELOG.md`'s rebuild section still describes part 1's
  stub (`STATE.md`).
- Rohan's at 5.7: the `release` environment and a new tap token (`STATE.md`).

**The spec's own text**
- §18 and §19 D give both 60 h / 30-46 h and 61 h / 30-47 h; the plan uses 61 and 30-47 (`STATE.md`).

## 6. Where the numbers come from

- The 8 Oct run reports in `records/runs/`, their logs and "Runs" tables: `R-2026-10-08-setup.md`,
  `-plan-review.md`, `-formats.md` (part 0), `-clear-out.md` (part 1), `-test-pack.md` (part 4a), `-reader.md`
  (part 2), `-engine.md` (part 3), `-hook.md` (part 5, with Rohan's first sitting), `-plugins.md` (part 4b, with the
  second sitting).
- Today's run report, `records/runs/R-2026-10-09-gate.md`: 00:25 the versions and settings hashes; 00:27 the base
  commit; 00:28 the Haiku run's hours and lines of Go; 00:35 the Sonnet run's measurements on both sides (its new files
  are under `bonsai-checks/gate/` on each side, scratch, not committed); 00:39 the Sonnet eval run (its suite under
  `~/bonsai-checks/eval/test-pack/evals/`).
- The last verifier's report (Opus, 00:49-01:03), in today's run report: the tests at `09cff74` on both sides, part
  3's scripted checks re-run, checks 3 (Windows), 11 (WSL) and 12, the stop lines, and its findings by letter and
  number (cited as `verifier`).
- `STATE.md` at `b18f444`: the findings and loose ends.
- `design/bonsai-spec.md` §14 (the parts, the checks, the stop lines, the paths and totals) and §15 (the baseline and
  the list); `design/plan.md` (the parts table, the checks' readings, the stop lines, the hand checks).
- Commands: lines of Go, `git diff --numstat <from>..<to> -- '*.go'`, split by `_test.go` (re-run by this builder);
  the CI reads, `gh api repos/LastStep/Bonsai/commits/<sha>/check-runs` (plugins 21:24); the ratios and scaled hours,
  the arithmetic in 2.2.
