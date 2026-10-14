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

## Runs

| Run | Model | Start | End | Minutes |
|---|---|---|---|---|
| Builder, phase A | Opus | 01:47 | 02:08 | 21 |
