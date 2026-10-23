---
format: bonsai.run/1   # fields: bonsai check --schema bonsai.run
id: R-2026-10-23-T-5798
task: T-5798
role: orchestrator
model: opus
started: 2026-10-23 06:48
finished: 2026-10-23 07:22
outcome: needs-person
commits: []
labels: {}
---

# Run: T-5798, Bonsai under pi: the contract plan

- Who: the orchestrator (opus), with an auditor and a plan reviewer
- Why: the director, 23 Oct, in the walkthrough of `design/agent-neutral.md`: "create kind of a contract plan which
  includes all the things pi needs to build, then i will give that plan to pi and build it there in parallel"; built
  in their pi-harness repo; the Go side here after 1.0. The choices: `R-2026-10-22-T-5797` (23 Oct lines).

## Log

- 06:48 Task written at `running` (it grants nothing). auditor (opus) `M-2026-10-23-43472a70`: today's hook
  interface, pi's tools, what a forged payload meets, for the stopgap and the envelope.
- 07:20 auditor (opus) report `M-2026-10-23-2162eec5` (full reply kept in the session): rc.1 has four hooks, none
  prints JSON (exit code and stderr only); the guard needs `CLAUDE_PROJECT_DIR`; unknown tool names pass unjudged
  (pi's lowercase names would); session ids `[A-Za-z0-9_-]{1,128}`; the stop gate reads no payload beyond
  event/session/cwd and has no cap; usage reads only Claude's transcript; pi's eight tools and six path rules with
  their mapping. Every fact pointed, none run beyond `--help`.
- 06:59 `design/pi-contract.md` written (the extension event by event, the two modes, the envelope draft, limits, proof, order). A fresh plan reviewer, verifier (opus), facts lens with security: `M-2026-10-23-c8f7c0db`.
- 07:10 Review round 1: verifier (opus) `M-2026-10-23-14c384bf` FAIL: B1 the hold's entry type (a silent yes), B2 a guarded parent starting an unguarded child, B3 last in packages is not last, B4 the cap's reset; nine minor. All folded in; round 2 sent.
- 07:13 Review round 2: `M-2026-10-23-3afadbb6` FAIL: round 1's fixes hold; new N1 (a session switch judged against the old project); four minor. Fixed; round 3 sent.
- 07:15 Review round 3: `M-2026-10-23-683df5ac` FAIL: N1 still open (pi reloads and re-evaluates the extension on a switch, so a load-time constant is read again); one minor. Fixed: a pid marker beside every export plus a once-only globalThis stash; a two-load unit test and a live /resume check; the minor. The round-3 verifier is at its role's round limit: a fresh verifier for the one point.
- 07:17 Round 4, a fresh verifier (opus) on the one point: `M-2026-10-23-e6240f74` PASS (the stash survives pi's reloads: jiti runs in this context; children and grandchildren keep the parent's project). Its risks 1, 3 and 4 folded in (the parent session by the same rules; two tests; the :385 cite); 2 and 5 accepted as stated (a child switched interactively stays on the inherited project; pid reuse is remote). Facts lens catches across rounds: 4 + 1 blocking held, 0 not held.
- 07:22 The director: "Approve (Recommended)". `T-5798` done; the file is theirs to hand to pi (outside this repo).

## Left

- The director hands `design/pi-contract.md` to pi, to build in `~/Servers/pi-harness`.
- After 1.0: Bonsai's Go side (section 9), its plan reviewed and approved; section 5 moves into `design/contract.md`.
- Before 1.0: `T-5710` item 13 (the markers), with rc.2's fixes.

## Processes left running

None: the auditor and verifiers were read-only and reported they started none; this session started only short
foreground commands (`ps -eo pid,etime,cmd` checked at the end).
