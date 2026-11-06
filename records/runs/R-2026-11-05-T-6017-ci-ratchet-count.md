---
format: bonsai.run/1   # fields: bonsai check --schema bonsai.run; how to fill: skill base:run
id: R-2026-11-05-T-6017
task: T-6017
role: orchestrator
model: opus
started: 2026-11-05 12:06
finished: 2026-11-06 10:50
outcome: merged
commits: ["fd8df70c", "2fc9e5f9"]
labels: {}
---

# Run: the CI ratchet count (BUG-138)

> Times UTC (`bonsai now`). The director present.

## Log

- **12:06** The director: **"yes, build 1.0.3"** (T-6015's verdicts A, B, D, E). T-6017 written `running` with
  T-6016, T-6017 and T-6018 side by side: none grants a protected path (`bonsai.allows` empty), so the three running
  together lose no grant; their hours file under "none" in the sessions table while all three run. Worktree
  `~/Servers/Bonsai-T-6017` from `main` `543973ca`. The three land as one batch (1.0.3).
- **12:08** Brief `M-2026-11-05-6648d42d` to builder@T-6017 (**opus**, minutes=240). Started as a background subagent.
- **12:25** builder@T-6017 (opus) report `M-2026-11-05-956ce04a`, done: `28e9acf2` (internal/ladder: under --ci a
  capture count gets the accepted marked skips added, to Now and to Parsed alike, `plusSkips`; ratchet's red reason
  names it), `42e85d12` (CHANGELOG Unreleased, Fixed). `TestCaptureCountUnderCI` red on main ("n_tests went backwards: 3
  -> 2", `~/bonsai-checks/T-6017/failfirst.txt`), green after; `TestSkipsAndMarks` unchanged and green; a capture
  counting skips itself disagrees and stays red. Climb green at `42e85d12cb1ff6f4bc0d6143a749c2a7ea1f0d70`, dirty false,
  sha256 `ae5e3eb0f423969787d72f74b372df67a7347544679a30454b01ea09ae171d5e`; go_tests 5208, tagged 2424. One edit by a
  shell replace (internal/ratchet/ratchet.go, a free path). The orchestrator read the diff (gofmt clean) and the
  studio's rung 4 (read only): `capture` with `tests: tap` and one ci_marked_tests name, the case the fix covers.
  T-6017 `verify`; the verifier shared with T-6016.
- **12:44** **Paused on the director's word: "pause once these subagents are done".** This task waits at `verify` for 1.0.3's batch; the next steps are in T-6016's run report's 12:44 line.
- **2026-11-06 06:23** Resumed ("go"): T-6016's verifier (`M-2026-11-06-6d1cc5f5`) judges this task too.
- **2026-11-06 06:31** verifier@T-6016 (opus, fresh) report `M-2026-11-06-fb3e78c1`: **PASS** for T-6016 at `9a530a0a`
  and T-6017 at `42e85d12`, no blocking finding (status and check byte-identical to `543973ca`'s, with and without a
  cache; status --json median 8.55 s to 0.25 s cached, 1.75 s cold, in a copy of the repo; four runs at once left one
  valid cache; -race clean; T-6017's test red on main, no double count). Non-blocking: the cache's version not tied to
  the build; a temp file left by a killed write never cleaned; T-6016 had no CHANGELOG line (taken in the 1.0.3
  section).
- **2026-11-06 06:50** The batch: T-6018 `c100c2ef` (its climb green, unchanged), T-6017 rebased onto it, T-6016 onto
  T-6017 (patch-ids equal before and after), then `55c08d81` on top: CHANGELOG's `## [1.0.3] - 2026-11-06` (status
  fast, the CI ratchet count; `release-notes.sh v1.0.3` 12 lines). Climb at this task's tip: green, `git.sha` `2fc9e5f9a342356608cef2ce729f3fd40b117a6f`,
  dirty false, sha256 `5462d1f0ca01070e2e23c4ee5d87b6538a8ac823c13330eae2b4e44e1e9e88c0` (log record by `/usr/local/bin/bonsai`); rungs 0 8.0 s, 1 1.2 s, 2 0.6 s, 3 88.0 s (5217), 4 79.2 s (2427), 5 0.1 s. The top branch
  `T-6016-status-speed` pushed; CI awaited.
- **2026-11-06 07:20** CI on `55c08d81` (the batch's top, run on the branch): 9 of 9 green, both windows jobs.
  `origin/main` unmoved at `543973ca`; `main` fast-forwarded to `55c08d81` (8 commits: T-6018, T-6017, T-6016 and
  1.0.3's notes), pushed; the branch deleted on GitHub; the three worktrees removed. `main`'s own CI and CodeQL
  awaited before the `v1.0.3` tag line goes to the director.
- **2026-11-06 10:50** The director: "done" (tagged `v1.0.3`, approved the release). Release run 42184887487 success;
  **Bonsai 1.0.3 published** 10:48 UTC, `releases/latest` `v1.0.3`; read back in `~/bonsai-checks/release-1.0.3/`: six
  archives `sha256sum -c` OK, `gh attestation verify` exit 0, `bonsai 1.0.3 (commit 55c08d817d6c)`. The way back: the
  installed 1.0.2 in `~/bonsai-checks/prerelease/previous/` (1.0.1 moved to `older-1.0.1/`). This task `done`.

## Processes left running

None (`ps -eo pid,etime,cmd`: no climb, test or CI watch of this run).
