---
format: bonsai.run/1   # fields: bonsai check --schema bonsai.run
id: R-2026-10-20-T-5702
task: T-5702
role: orchestrator
model: opus
started: 2026-10-20 13:35
finished: 2026-10-20 15:45
outcome: merged
commits: ["4bd50c88", "c79143e8", "39a95367", "90c0ac6d", "72448f40", "b93dd719", "224f6263", "b66ea820", "4ec5c432"]
labels: {}
---

# Run: T-5702, 5.7.2, the release workflow

- Who: the orchestrator, with the builder it briefs
- Why: `design/plan-5.md`, step 5.7, piece 5.7.2 (approved 10 Oct; the director's word for its `release.yml` changes)

> Opened before the first edit and appended to as the work goes.

## Runs

| Run | Model | Start | End | Minutes |
|---|---|---|---|---|
| Builder M-2026-10-20-fbd94566 (audit, build, tag-check proof) | opus | 2026-10-20 13:36 | 2026-10-20 14:00 | 24 |
| Builder, resumed (the test with a real answer, climb) | opus | 2026-10-20 14:38 | 2026-10-20 14:51 | 13 |
| V1 M-2026-10-20-353b8ff5 (the release path) | opus | 2026-10-20 14:56 | 2026-10-20 15:10 | 14 |
| Builder, V1's fixes, then shell bash | opus | 2026-10-20 15:08 | 2026-10-20 15:23 | 15 |
| Orchestrator, V1's Windows dist run | opus | 2026-10-20 15:12 | 2026-10-20 15:15 | 3 |
| V1 again M-2026-10-20-3a1f249d | opus | 2026-10-20 15:24 | 2026-10-20 15:28 | 4 |

## Log

- 13:35 T-5701 to `verify` (its climb green; CI on its branch running), T-5702 `running` (grants release.yml,
  the tag-check script, the templates, CLAUDE.md; the one running task). Builder (opus: the release path) briefed,
  `M-2026-10-20-fbd94566`, audit first; base 137e419f; nothing writes to GitHub.
- 13:45 Audit `M-2026-10-20-015fd900` (`~/bonsai-checks/T-5702/audit.md`): ten real check names (install's matrix names; windows (fault tag)), the tagged pass on ./... in build, no queue key (actionlint refuses it), a gh>=2.97 guard and an existing-release refusal in publish; GoReleaser asks no token (measured); gh reads a draft (from source; rc.1 the real test, fails closed). Go: Owns amended with workflows_test.go (release.yml's exemption removed) and tagcheck_test.go with testdata (choice B).
- 13:58 Report `M-2026-10-20-092081da`, blocked (its run ended while waiting): 5 commits to 72448f40 (the tag-check script, release.yml, the pin test reading release.yml, the templates, CLAUDE.md's release lines); pin test and actionlint pass locally; tagcheck_test.go (19 cases) uncommitted, waiting for a real saved answer of a fully green commit. Its live check on 137e419f during main's own CI run exited 1 naming the ten jobs in progress (the running proof, saved). Note: on a landing the same commit runs CI twice (branch, then main); the check reads both. It ran `git fetch` in the main checkout once (refs only). The orchestrator's earlier note that 137e419f's CI was complete was wrong: only the branch's run was.
- 14:44 The director: "nothing is running". True: the builder had stopped after its blocked report, and the orchestrator had read its later "waiting on background work" notice as live. main's CI on 137e419f found complete, all success. Builder resumed, `M-2026-10-20-3ae28a21`, told never to wait in the background.
- 14:51 Report `M-2026-10-20-7f352a68`, done: b93dd719 (+ the tag-check test with main's real answer for
  137e419f, 19 runs). Climb green, git.sha b93dd719, dirty false, sha256
  a324a3fa9403c7ac5c261a618bf955a472a8a32c56787e3ababb0efe6ceb5a14 (rungs 0 3.6 s, 1 2.5 s, 2 0.5 s, 3 61.8 s 4798,
  4 55.6 s 2125, 5 0.04 s). Tag check: main complete exits 0; main running exits 1 naming ten jobs; a scratch branch's
  commit exits 1; v1.0 and v1.0.0-beta exit 1; own suite left out; three break-its each caught by a case. The
  orchestrator read: triggers the two tag patterns only; permissions {} at top; build contents/checks/actions read,
  publish contents/id-token/attestations write inside environment release; GH_TOKEN per step only; cache false;
  CLAUDE.md's new release lines. Branch pushed; CI run 39400056360. T-5702 `verify`.
- 14:52 V1 (opus, fresh) briefed, `M-2026-10-20-353b8ff5` with its correction `M-2026-10-20-32ed511d` (brief `~/bonsai-checks/briefs/V1-5.7-brief.md`), on b93dd719 while the branch's CI runs; the snapshot for it built by the orchestrator in `~/bonsai-checks/V1-5.7/` (go run goreleaser v2.18.3, current tag v0.4.3, exit 0, six archives and checksums.txt).
- 15:07 **V1 FAIL** `M-2026-10-20-920de17f`: must-fix, release-tag-check.sh:120-134 passes an unreadable answer (jq | sort under dash takes sort's status; Not Found and garbage exit 0); should-fix, gh's signer match is a prefix (release.yml-x.yml would match; the tagged commit's own release.yml runs, so an off-main v-tag could sign: both need write access, the same trust as tagging); the failure branch untested. Held: triggers, permissions, publish runs nothing from the build, order, pins (ls-remote), GoReleaser no token, the ten names, dist checks and 11 doctored copies, -ForceChild, install.sh, 5.7.3's line, decision 2 (B), batch 1 against REST docs (immutable-releases GET is 200 enabled:false when off; ruleset update takes no parameters). Gap: Windows dist run. The orchestrator: fix (1) and the pin rule forward on T-5702; --cert-identity left as a known limit for the director (no real Bonsai release to prove it on).
- 15:07 V1's gap closed by the orchestrator (the one Windows run): the b93dd719 snapshot and a Windows-git clone at b93dd719 in %USERPROFILE%\bonsai-checks\V1-5.7, Windows Go 1.27.2: TestDist PASS (checksums, six archives, installer), ok 4.522 s, exit 0. Builder resumed on the fix, `M-2026-10-20-a173149c`.
- 15:15 Fix report `M-2026-10-20-8f3a9d16`: 224f6263 (the tag check refuses an empty, non-JSON or list-less answer; jq's own status tested; four new cases, three exiting 0 on the old script), b66ea820 (no workflow named like release.yml; shown failing on a scratch clone); climb green, sha256 c8db7454a620649fb3d57029947a347d472f85d763730dfea286fe92b40c0912. The builder's catch: release.yml's run steps name no shell (bash -e, no pipefail), so the draft step's gh | awk passes a gh failure. Go: defaults shell bash, held by a test.
- 15:16 Report `M-2026-10-20-385dbf86`, blocked: the guard refused release.yml (person-only; T-5702 at verify, so no grant): the orchestrator's error, moving it to verify while a fix round was open. T-5702 back to `running`; the builder resumed.
- 15:23 Report `M-2026-10-20-9bcd698e`: 4ec5c432 (defaults shell bash, GitHub's page read 20 Oct; the test failing first on the real file and on two copies); climb green, sha256 cb9bbcfdc7535ab64165acd698b00c1fc3be715c2c2e15f8209ae84fbd0703d5 (3 59.5 s 4807, 4 68.2 s 2125). Branch pushed. A fresh verifier with V1's brief (V1b) briefed.
- 15:27 **V1 again (V1b): PASS** `M-2026-10-20-2a5d18eb` (opus, fresh) at 4ec5c432: the must-fix fixed (nine bad answers each exit 1 with a reason; the good one exits 0); every pipe and substitution read; release.yml:211 now stops on a gh failure; GitHub's shell page read 20 Oct; the diff since b93dd719 is four files, build files untouched, so the snapshot and dist results on both sides hold; new tests fail on the old code. Notes, none blocking: A release.yml:141 an echo's substitution (fails closed at :186); B release.yml:282 the cleanup loop passes on a gh failure (the draft would stay; the next run refuses while it exists); C the shell rule reads block style only. Gap: CI still running.
- 15:45 CI run 39403956118 on 4ec5c432 (branch T-5702): every job success. `main` fast-forwarded
  137e419f..4ec5c432 and pushed; the branch deleted on GitHub; the worktree removed by `git worktree remove`. Landing:
  git.sha 4ec5c432, sha256 cb9bbcfdc7535ab64165acd698b00c1fc3be715c2c2e15f8209ae84fbd0703d5, green. Task `done`.
  **V1 passed at 4ec5c432**: the release path is ready for the director's batch 1.

## Processes left running

None of this task's (the builder and both verifiers reported none; the CI watches exited; ps -eo pid,etime,cmd).

## What is left

V1's notes A-C (release.yml:141, :282, the shell rule's block-style reading) for a later task; B told to the director
with batch 2 (a failed run's draft may need deleting by hand). The `--cert-identity` exact form in the install lines,
a known limit, told to the director.
