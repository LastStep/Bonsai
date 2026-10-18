---
format: bonsai.run/1   # fields: bonsai check --schema bonsai.run
id: R-2026-10-18-T-5692
task: T-5692
role: orchestrator
model: opus
started: 2026-10-18 12:26
finished:
outcome:
commits: []
labels: {}
---

# Run: T-5692, Windows CI fixed forward (TestPersonOnly)

- Who: the orchestrator, with the builder it briefs
- Why: CI's windows job red on `main` since `4b41790b` (5.6.0); the director, 18 Oct: "go", each agent auditing the code first

> Opened before the first edit and appended to as the work goes.

## Runs

| Run | Model | Start | End | Minutes |
|---|---|---|---|---|
| Builder, brief M-2026-10-18-851a78dc, report ac78bb66 | opus | 2026-10-18 12:25 | 2026-10-18 13:08 | 43 |
| Audit M-2026-10-18-7371b037 | opus | 2026-10-18 12:25 | 2026-10-18 12:38 | 13 |
| Batch 1 Windows runner M-2026-10-18-f6d59384 (shared) | sonnet | not written | 2026-10-18 14:16 | not written |

> Runs rows for 16-18 Oct filled on 19 Oct from the log by the count `M-2026-10-18-e11266d9` (haiku); "not written" where the log names no time.

## Log

- 12:26 T-5692 written `running` (no grants) beside T-5601, T-5602, T-5604, T-5606 (all `running`, no grants). The director, 18 Oct: "make sure that each agent first does an audit of whats actually in the codebase, and not fully trust the state, as previous runs were stopped midway"; written into `~/bonsai-checks/briefs/common-2026-10-18.md`.
- 12:29 Builder started (opus), brief `M-2026-10-18-851a78dc`; audit first, then stop for a newer brief.
- 12:39 Audit `M-2026-10-18-7371b037` (opus, 12:25-12:38): cause in the code, not the test. `realPath` (cmd/bonsai/agent.go:126) resolves only existing paths (EvalSymlinks); CI's temp root has the short name `RUNNER~1`, expanded for the existing root but not for the not-yet-made home, so `folderHolds` fails and both cases are refused `test-build` (fails closed). Reproduced natively on Windows with a throwaway program; the same on Linux through a symlinked root. The same pattern in `internal/engine/hookline.go` (`realFolder`, `Confine`): added to the Owns. Go sent (`M-2026-10-18-d58e0ba1`).
- 18 Oct 12:41 The director: "be extra careful with any windows related work ... if you think there are a lot of things going on already, then wait for some to finish before queing something next". Builders run no Windows half; the orchestrator runs one at each batch's top, alone, when load is low. Two Windows runs already going at the time (T-5604's, T-5606's helpers; load 18-20) left to finish.
- 13:09 Report `M-2026-10-18-ac78bb66` (opus, 12:25-13:08, 43 min, no helpers): done. `engine.RealPath` resolves the existing part of a path and joins the rest (8.3 short names made long, a link to a not-yet-made target followed); `engine.FolderHolds` used by `Confine` and `personOnlyAt`; cmd/bonsai's copy removed. Closes a hole too: a link inside the scratch root pointing outside, at a target not made yet, counted as inside; now refused. Tests TestPersonOnlyLinkedRoot, TestConfineLinkedRoot, both failing on main first (the ladder's new-tests check: 2 of 2). Check 10 WSL plain 4648/0/4 skip, tagged 4654/0/4. Climb read in the main checkout: local, green, git.sha `5648071d546f2d371e2fbc50213c1ccdc806c17c` = HEAD, dirty false, rungs 0-5 green (tests 99.9 s, tagged 138.1 s), sha256 `bc8a2998a42dad1dbe138eb7263e7deb9390f87b717a182d811273aaffbe745c`, log names `/usr/local/bin/bonsai`. Diff read (4 files, +96/-31, inside Owns). Follow-up for one home: guard.resolve and engine.RealPath are the same walk. V1 reads this change (the confinement is a tripwire). T-5692 to `verify`; batch with T-5602 (T-5692 first; T-5602 rebased onto it, re-climbed).
- 13:41 Batch 1 = T-5692 (5648071d) → T-5602 (f0d9c6e2) → T-5604 (b5995ea5). T-5604 re-climbed at `b5995ea586289dd9b95a980b58cdb5e76172be21`: local, green, dirty false, rungs 0-5 green (4667, 4673), sha256 `2f21b687f876e4386190be57d162756fdc962137f695785c72f9f2b9216daee6` (report `M-2026-10-18-90eca423`, sonnet). Windows half at the top: load 3.6, no Windows Go process; CRLF checkout by Windows git from a bundle at `%USERPROFILE%\bonsai-checks\batch-2026-10-18\src` (HEAD b5995ea5, autocrlf true).
- 14:16 **Batch 1 landed:** the Windows half at the top `b5995ea5` (runner `M-2026-10-18-f6d59384`, sonnet, one step at a time, alone): plain test 30 ok / 0 FAIL (15.1 min), vet clean; tagged 30 ok / 0 FAIL (17.5 min), vet clean; no Defender failure. Status line through Git Bash (plain build, the repo checkout, 50 runs): p50 55.5 ms, p95 60.4 ms, exit 0 (measured, not budgeted; the native 60 ms budget held by TestStatusLineSpeed passing there). The diff `main..b5995ea5` read by folder (cmd/bonsai 9, internal/engine 2, internal/status 1, internal/workspace 1, operating-setup 1): inside the three tasks' Owns, nothing under records/, design/ or STATE. origin/main was `7daf9f35` (no other session's push). `main` fast-forwarded `7daf9f35..b5995ea5` and pushed; CI to read.
- 15:10 **CI green at `b5995ea5`**, every check (test, windows, lint, govulncheck, packs, Analyze Go). Task to `done`.
