---
format: bonsai.run/1   # fields: bonsai check --schema bonsai.run
id: R-2026-10-20-T-5793
task: T-5793
role: orchestrator
model: opus
started: 2026-10-20 12:05
finished: 2026-10-20 12:12
outcome: blocked
commits: []
labels: {}
---

# Run: T-5793, scratch folders declared in bonsai.yaml (planning)

- Who: the orchestrator, with an auditor and a researcher
- Why: the director, 20 Oct (in the task); BUG-013

> Opened before the first edit and appended to as the work goes.

## Runs

| Run | Model | Start | End | Minutes |
|---|---|---|---|---|

## Log

- 12:05 T-5793 written `plan` (full lane). Auditor (opus) `M-2026-10-20-ce717ea7` on bonsai.yaml, .bonsai/local,
  the guard's authority kinds and what goes stale; researcher (sonnet) `M-2026-10-20-2821844b` on how other agent tools
  declare sandbox folders. The first audit, `M-2026-10-20-68e4b9a0`, is in T-5792's run report.
- 12:03 Researcher M-2026-10-20-6a934eca (sonnet), done: Claude Code, Codex, Gemini, Cursor, Aider, devcontainers, Bubblewrap, CrewAI. Lessons: a declared root minus a fixed deny-list (deny wins); resolve the real path of target and parents, judge both spellings, refuse symlinks out; declare in a layer the limited agent cannot write, scope only narrows; separate authorities, each least by default (execute the riskiest); say who cleans, no widening from inside. Gaps: no tool separates delete from write; no documented cleanup.
- 12:07 Auditor M-2026-10-20-d4f09499 (opus), done: .bonsai/local/ is closed by three layers (Claude Code's deny Edit(//**/.bonsai/local/**), the guard's floor entry floor.go:42, role-writes-nothing) and holds the ladder results the stop gate trusts; git write words refused to writes:false roles by word whatever the path (rolecmds.go:775-786, :207); any folder with a bonsai.yaml is a project to the guard (hookin.go:115-146), so a scratch clone of Bonsai stays closed; a new top-level bonsai.yaml key is the safe place (roles' names are a closed list); joins Checks/Loosenings (checks.go; readChecks needs every key: a missing key must read as empty); formats set 11, the director's word; enforceable by path: file tools, shell write words, rm/mv, chmod; by word: git, bonsai; not at all: a program's own writes, interpreters, network; cleaning by age needs a new generated kind. Questions to the director next, one at a time.
- 12:12 Asked the timing (now before rc.1, recommended; or after 1.0): **"okay we will tackle this idea after 1.0"**. T-5793 back to `todo`, marked after 1.0; facts and the draft shape kept (`~/bonsai-checks/T-5793/notes.md`).
