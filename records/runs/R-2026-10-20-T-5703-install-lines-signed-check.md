---
format: bonsai.run/1   # fields: bonsai check --schema bonsai.run
id: R-2026-10-20-T-5703
task: T-5703
role: orchestrator
model: opus
started: 2026-10-20 12:01
finished: 2026-10-20 13:28
outcome: merged
commits: ["21c961fa", "52f9a580", "7692ee82", "44d0632b", "11943f30"]
labels: {}
---

# Run: T-5703, 5.7.3, the stronger check in the install lines

- Who: the orchestrator, with the builder it briefs
- Why: `design/plan-5.md`, step 5.7, piece 5.7.3 (approved 10 Oct)

> Opened before the first edit and appended to as the work goes.

## Runs

| Run | Model | Start | End | Minutes |
|---|---|---|---|---|
| Builder M-2026-10-20-5cdc09a2 (audit, build, gh proof, two climbs) | opus | 2026-10-20 11:58 | 2026-10-20 12:34 | 36 |
| Builder, the fix and two rebases with climbs | opus | 2026-10-20 12:37 | 2026-10-20 12:54 | 17 |

## Log

- 12:01 T-5792 moved to `verify` (its climb green), T-5703 to `running` (no grants). Builder (opus: the
  release lines and a doc test) briefed, `M-2026-10-20-5cdc09a2`, audit first; its one Windows run named in the brief
  (a `gh.exe attestation verify` set, no Go build).
- 12:03 Audit M-2026-10-20-425c77e3 read; go given (skill test yes, cmd/bonsai/operating_test.go added to Owns; the doc named by its GitHub URL; base aa866b3a).
- 12:35 Builder's report `M-2026-10-20-2260beef`, done: branch T-5703 at 35617152 (0b7b5213 release, 1b4d1bae, 500d3062
  docs, 35617152 packs); first climb red at rung 0 only (before main held T-5792), second green: git.sha 35617152,
  dirty false, sha256 d70134bb36218e0924bb646ebb61c50868e3c33a5569b497aa54dfc93df6820d, rungs 0 3.6 s, 1 0.8 s, 2 0.5 s,
  3 61.3 s (go_tests 4767), 4 53.5 s (2121), 5 green. The gh proof (2.102.0 both sides, cli/cli v2.102.0's own
  archives; logs in `~/bonsai-checks/T-5703/`): cli/cli signs from its trunk branch, so the exact form with
  --source-ref refs/tags/v2.102.0 fails on it (exit 1, ref mismatch); with its real ref WSL "Verification succeeded!"
  exit 0, Windows exit 0; the wrong repository 404 exit 1; a lookalike workflow "Sigstore verification failed" exit 1;
  --deny-self-hosted-runners adds the github-hosted requirement, its failing case not shown (no such file). For 5.7.2:
  Bonsai's line passes only on a tag-run release.yml (V1 to read). The orchestrator read the diff: one fix asked
  (the way out's fingerprint, `M-2026-10-20-2f11f14e`).
- 12:45 Fix report `M-2026-10-20-08f9a3f8`: a8a442aa (the way out's copy and its fingerprint, two lines); climb green, sha256 da598b17c6f26ed5ba9adcd431fceb5a76cbf90101528be5436cfcbe66e5fd3c. Batch with T-5704: rebase onto 2580e4c3 and climb again (`M-2026-10-20-d563c32f`).
- 12:46 Report `M-2026-10-20-8b665f37`: rebased onto T-5704 (3b3701f0), climb red at rung 0 only (sha256 8e1ed6df...; .gitattributes, T-5704's, judged against T-5703's empty grant). The orchestrator flips the batch: T-5703 at the bottom on dba4e7af, T-5704 on top (`M-2026-10-20-*`).
- 12:53 Report `M-2026-10-20-bb3647f9`: T-5703 on dba4e7af at 11943f30, climb green, sha256 072030091e05e3ce5660e6e32cf3d10f1ed1503815164336065b633a3accff88 (rungs 0 3.6 s, 1 0.8 s, 2 0.4 s, 3 58.5 s 4767, 4 53.9 s 2121, 5 green). T-5704 to go on top.
- 13:02 T-5704 rebased on top (4ebee8cc; its report `M-2026-10-20-1e633f24`, climb green, sha256
  bc79bb2998850cd728517fe2f66a776d7e2e4a30c1d6b8d73508e83097d702da). The orchestrator read both results (green, exact
  tips, clean) and the batch's paths against both tasks' Owns. Branch T-5704 pushed as the batch's top; CI run
  39387029868. T-5703 moved to `verify`, T-5701 to `running`. **A rule learnt:** rung 0 judges a branch's whole diff
  from main against the one named task's grants, so in a batch the piece that changes protected files goes on top,
  and a batch holds at most one such piece (for `workflow:correct` at 5.7's end).
- 13:28 CI run 39387029868 on 4ebee8cc (the batch's top, branch T-5704): every job success. `main` fast-forwarded dba4e7af..4ebee8cc and pushed; branch T-5704 deleted on GitHub; both worktrees removed by `git worktree remove`. Landings: T-5703 at 11943f30 (sha256 072030091e05e3ce5660e6e32cf3d10f1ed1503815164336065b633a3accff88, green), T-5704 at 4ebee8cc (sha256 bc79bb2998850cd728517fe2f66a776d7e2e4a30c1d6b8d73508e83097d702da, green). Task `done`.

## Processes left running

None of this task's (ps -eo pid,etime,cmd; the builder reported none).

## What is left

Nothing for T-5703. For 5.7.2 and V1: Bonsai's line passes only on a release.yml run from a tag.
