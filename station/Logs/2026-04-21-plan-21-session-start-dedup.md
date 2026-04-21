# 2026-04-21 — Plan 21: Session-start dedup + Phase C sensors

## Context

User flagged too much context injection at session start. Asked to audit and remove redundancy — from catalog sensors and from `context-guard` itself. Follow issue-to-implementation workflow. Also incorporated Plan 08 Phase C (C1 compact-recovery sensor + C2 context-guard verify/plan patterns) into the same ship.

## Decisions

- **Cut redundancy, not features.** Kept identity.md, memory.md, self-awareness.md, INDEX.md, Status.md, security + scope-boundaries protocols, health checks. Removed only what was duplicate, misfired, or circular.
- **session-start.md protocol was self-referential junk.** Told the agent to re-read what the sensor had just dumped. Proof: at this session's start I spent 6 extra `Read` calls (~10k tokens) re-reading sensor-injected content. Rewrote to short version that defers to the sensor and handles the no-sensor fallback case.
- **UX preferences belong in `memory.md` Feedback, not `self-awareness.md`.** Catalog source of self-awareness is already lean; station's copy had accreted 50 lines of project-specific UX learnings. Moved unchanged under `## Feedback` → `### Durable UX preferences (2026-04-17 dogfooding)`, preserving each bullet's **Why** / **How to apply** structure.
- **`compact-recovery` lives alongside `session-context`, not inside it.** Used SessionStart matchers: `session-context` = `startup|resume|clear`, `compact-recovery` = `compact`. Avoids double-dumping after `/compact`. Generator already supported `matcher` field (`internal/generate/generate.go:473,491,517`) — no Go change needed.
- **Planning reminder guarded against wrap-up stacking.** `context-guard` fires at most one category per prompt. Implemented via `if not triggered:` check so a "let's plan X, that's all" style prompt prefers wrap-up over planning. Word-boundary only (`\b...\b`), NOT end-anchored like wrap-up patterns, because planning/verification phrases typically appear at prompt start or middle.
- **C3 (prompt hook intent classification) stays deferred.** Original Plan 08 verification note holds: ship auto-invocation + phrase regex first, revisit only if they prove insufficient.

## Outcomes

- PR #46 squash-merged → main `d14edbe`. Close-out commit `4eceab6`.
- SessionStart dump 34.3KB → 30.9KB (~10% cut). Banned strings (`PROTOCOL: memory.md`, `PROTOCOL: session-start.md`, "REMINDER: Before ending this session") absent. Required strings (core files, security + scope-boundaries protocols, SESSION HEALTH CHECK) present.
- New sensor `compact-recovery` wired for matcher=`compact`. Dumps Quick Triggers + Work State only. 2854 bytes for station's current (verbose) Work State — over 2000 target but explicitly acceptable per plan when Work State itself is verbose.
- `context-guard` pattern tests pass: `"verify everything"` → checklist; `"let's plan the caching layer"` → planning pointer; `"that's all I need to plan for today"` → no planning reminder (word-boundary holds).
- Plan 08 marked complete in Status. C3 tracked as a P3 research item with a specific revisit trigger.
- Follow-up backlog item: Plan 21 cut 10%; remaining ~30KB is core-file auto-dumps. Future optimization ideas (diff-based, summary+link, session-type-conditional) logged.

## Process notes

- Agent dispatch base = origin/main, not local main (memory learning from Plan 20). Pushed Plan 21 doc as commit `525ee11` BEFORE dispatching so the agent's worktree had the plan. Worked first try.
- `bonsai update` needs a TTY for its Huh prompt harness. Agent couldn't run it in non-interactive mode; wrote a temporary in-tree helper that called `generate.AgentWorkspace`, `WorkflowSkills`, `PathScopedRules`, and `SettingsJSON` directly with `force=false`. Helper was deleted after use. Decision: acceptable workaround for one-off agent execution; not a general fix. If we see this more than once, consider adding a non-interactive `--yes` flag to `bonsai update` (potential backlog item, not logged — let it recur once before systemizing).
- Station's `.claude/skills/*/SKILL.md` (6 files) were lockfile-referenced generator output from Plan 08 Phase A that had never been committed. Regeneration surfaced them; agent committed as part of PR. Closeout of an older loose end, orthogonal to Plan 21 intent but in-scope as generator-output hygiene.
- `gh pr merge --delete-branch` cleanly deleted the remote branch this time (no worktree checked out for the feature branch). Historically flaky — called out in memory.md as a pattern; no occurrence this session.

## Carry-forward

None. Plan 08 fully closed (A + B + C shipped; C3 deferred with explicit revisit criteria). Memory and Status reflect current state.
