---
format: bonsai.run/1   # fields: bonsai check --schema bonsai.run
id: R-2026-10-15-T-5501
task: T-5501
role: orchestrator
model: opus
started: 2026-10-15 01:45
finished:
outcome:
commits: []
labels: {}
---

# Run: T-5501, the pack template, its checks and Bonsai's CI job for packs (5.5.1)

- Who: the orchestrator; one Opus builder (CI and release); Sonnet runners for the Windows check 10; V1
- Branch: `t5501`, from `main` at `60545a7`, in the plain worktree `~/Servers/Bonsai-t5501`
- Why: `design/plan-5.md` Step 5.5, piece 5.5.1 and its note; Rohan's "go" (15 Oct, 01:44) for 5.5's next wave
- Proof: the task's `done_when`; the landing rule; CI; V1
- Beside it: `T-5500` and `T-5510`. Its role declaration passes `check --pack` only once 5.5.0 has landed, so its last commits wait for it; its last commit (`packs/template/.gitattributes` and `ci.yml`) is made while it is the only task `running`.

> Opened before the first edit and appended to as the work goes.

## Log

- 01:45 `T-5501` to `running` beside the other two of the wave (the section's order, item 2). Worktree made from
  `main` at `60545a7`. Brief `~/bonsai-checks/briefs/brief-T-5501.md`.
- 01:47 Builder started (Opus, a subagent of this session, unnamed), from the brief.
- 02:08 **Builder's phase A report** (Opus, 01:47-02:08, no helpers): five commits `493939f`..`56c55ce`, 9 files
  +752/-0, all in `packs/template/` (no `.gitattributes` yet). The template: `plugin.json` (`example-pack`, `author`, no
  `version`); `pack.yaml` (a comment on every key; one deny rule with its `why`, written with a leading `/`; one
  document kind `note`; one role `example`, a lead, `writes: true`, `max_turns: 200`, limits 30/30 minutes, 50/50
  actions, 3/3 red runs); `labels.yaml`; `agents/example.md` (`skills: [example-pack:example]`, `maxTurns: 200`);
  two skills (one a template with its fields table); README; `ci/check.sh` (`check`, `validate`, `release`, `version`;
  `sh`, documented); `pack.yml` (check on both systems, validate with Claude Code 2.1.300 on Linux, release on a `v*`
  tag only, actions pinned by commit, `contents: read` but the release job's write). Passed with a build of `60545a7`:
  `check --pack` no finding (roles not read before 5.5.0); `claude plugin validate --json` only the `version` warning;
  **validate needs no login** (an empty config folder and home, no key or token: same JSON, exit 0); the script green
  on WSL three ways and natively on Windows two ways (Git Bash `sh`, `claude` and `node` unreachable); the release
  subcommand on made-up names (`v9.9.9`, `v0.2.0`, `0.1.0` fail; `v0.1.0` passes; nothing made); item 7's broken
  copies each fail; shellcheck and actionlint (run with `go run`, nothing installed) clean. Drafts of the template's
  `.gitattributes` and the `ci.yml` edits (the `packs` job, one `windows` step) in `~/bonsai-checks/T-5501/drafts/`.
  Its choices: a local source copied into a bare repository beside the scratch project (naming the checkout gives
  `check` an `absolute-path` finding), `--update-shallow` for CI's depth-1 checkout; a `version` subcommand so the
  release job reads the version once; `author` added so validate shows only `version`; the prefixed `skills:` form,
  5.5.4 to measure which loads; Bonsai's own `ci.yml` keeps its `@v6` tags (pinning is 5.7's). 5.1.9's limits and the
  marketplace observation answered in the README, the engine's parts left to a later engine piece. Guard: `bonsai init
  --help` refused on the line (run from a script); an `rm -rf` of a variable path refused, run from a script with the
  path written out.
- 02:08 The orchestrator's read: phase A inside Owns. **One rule for the rest of the piece:** a delete the guard
  refuses for a variable path is typed again on the shell line with its literal path, so the guard judges it, never run
  from a script (the script rule is for a person's `bonsai` steps). The builder waits for 5.5.0's landing (phase B:
  rebase, `check --pack` with roles, `BONSAI_COMMIT` moved to 5.5.0's landing commit).
- 02:31 `T-5501` to `blocked` for `T-5500`'s lone window and landing (its last commits read 5.5.0's work).
- 02:56 `T-5501` to `running`, alone (5.5.0 landed at `e1437d0`): phase B and its protected commit (phase C) first, so no other
  task reads `blocked` while its builder works; `T-5510` and `T-5511` start right after that commit.
- 02:58 Phase B and C report: rebased on `2deaf2e` (phase A now `37ff60a`..`1e14b51`); `b9077ae` moves `BONSAI_COMMIT` to
  5.5.0's `e1437d0`; **`c641c15`, the protected commit** (the template's `.gitattributes` and the two `ci.yml` edits, equal
  to the drafts; the guard refused neither, `T-5501` the only task `running`). `check --pack` with a build of HEAD: no
  finding; 5.5.0's rules reach the template (scratch copies raise `pack-role-turns`, `pack-role-undeclared`,
  `pack-deny-relative`); `ci/check.sh check` green on WSL. The `ci.yml` diff read by the orchestrator: the `packs` job
  (Claude Code pinned from the template's `pack.yml`, no secret) and one `windows` step. Builder sent on to check 10
  and the climb; `T-5510` and `T-5511` start now.
- 03:18 **Builder's checks** (02:58-03:18; a Sonnet helper for Windows, 02:58-03:17): WSL plain 4284 pass, tagged 4290, 0
  fail, 2 skips; Windows natively plain 4263, tagged 4269, 0 fail, 18 skips (the known ones), vet clean, Defender
  blocked nothing; `ci/check.sh check` natively on Windows at HEAD exit 0, `release v9.9.9` 1 and `v0.1.0` 0;
  actionlint (by `go run`) and shellcheck clean. Climb green (03:02-03:05, load 6.4-7.3).
- 03:19 **Landed by the landing rule:** `.bonsai/local/ladder/T-5501.json` mode local, green, `git.sha` `c641c15` (the
  branch's HEAD), dirty false, rungs 0-5 green (0.8/0.9/0.5/72.2/66.7/0.04 s; `go_tests` 4188->4284, tagged
  4194->4290); its `ladder` record (21:35:28Z) names `/usr/local/bin/bonsai` and
  `sha256:67e7df23c51b712e93dec9937c43390d3512259fc34470310fb1b5b3c7a73b8e`. The diff: `packs/template/` and
  `.github/workflows/ci.yml` only. `main` fast-forwarded `2deaf2e..c641c15` and pushed; the worktree removed, the
  branch deleted. **`T-5501` to `verify`, until V1 passes** (and CI's new `packs` job and `windows` step green).
- **An observation for the end verifier (the ladder's, not this piece's):** each test rung's `new_tests` reads count
  96, `not_run` 96, against base `2deaf2e`, though this branch adds no test: 96 is the tests landed since 5.4's
  closing record (4284 - 4188), so the count reads against the floors' last recorded names rather than the base's own,
  and none runs on the base (`base_setup` null). It goes away when the floors are raised at the part's end.

## Runs

| Run | Model | Start | End | Minutes |
|---|---|---|---|---|
| Builder, phase A | Opus | 01:47 | 02:08 | 21 |
| Builder, phases B and C | Opus | 02:56 | 02:58 | 2 |
| Builder, check 10 and the climb | Opus | 02:58 | 03:18 | 20 |
| Builder's helper: Windows check 10 | Sonnet | 02:58 | 03:17 | 19 |
| Orchestrator: phase A's read | Opus | 02:08 | 02:10 | 2 |
| Orchestrator: the lone window, the ci.yml read | Opus | 02:56 | 02:58 | 2 |
| Orchestrator: the landing | Opus | 03:18 | 03:19 | 1 |
