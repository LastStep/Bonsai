---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task; how to fill: skill base:task
id: T-6005
title: "The first link's small fixes, split from T-6004: BUG-119 and BUG-120 (missing-path skips git-ignored paths), BUG-125 (a worktree's pointer into .bonsai/local), BUG-129 (base's words), BUG-127 (the workflow pack: land only what the verifier saw)"
status: done
lane: light
done_when:
  - "BUG-119, BUG-120: bonsai check's missing-path does not judge a backticked path git ignores, so a clone and the real checkout give the same findings; a path neither tracked nor ignored that is missing is still a finding; tests shown failing on the old code first"
  - "BUG-125: bonsai message --pointer with a relative path into .bonsai/local/ from a linked worktree reads it in the main checkout, where Bonsai keeps it; other relative pointers keep reading against the worktree; a test shown failing first"
  - "BUG-129: base's message skill says the new id is at .message.id in --json; base's memory skill says a title holding ': ' is quoted; the skills' docs tests pass"
  - "BUG-127: the workflow pack's orchestrate and verification-ladder skills say the orchestrator lands only the sha a verifier passed, or, for commits after it, gets the verifier's word on the diff between them (or a fresh verify); committed in ~/bonsai-checks/bonsai-workflow, its check.sh green, not pushed or tagged (the director's)"
  - "The Bonsai changes landed by the landing rule (CLAUDE.md), CI green on the branch, both windows jobs among them; each bug moved to records/bugs.md's Fixed table with its commit"
depends_on: [T-6002, T-6003]
blocked_by: null
created: 2026-10-26
started: 2026-10-26
finished: 2026-10-26
labels:
  bonsai.ladder: []
  bonsai.allows: []
  bonsai.wants: []
  bonsai.branch: t-6005-link-fixes-small
  workflow.model: sonnet
  workflow.estimate_h: 2.5
---

# The first link's small fixes (split from T-6004)

## Context

The director, 26 Oct, choosing option 3 of the split of T-6004's twelve bugs: "3, and then update the state so a
fresh session can start working parallely from group b". This session takes group A (BUG-119, BUG-120, BUG-125,
BUG-129) and BUG-127's skill rule, after T-6002 and T-6003 land; T-6004 keeps the rest for a fresh session. The split
and its reasons: `records/runs/R-2026-10-26-T-6002-lfs-dirty.md` (10:39-10:45).

## Scope

- `internal/engine/checkdocs.go` (missing-path) and its tests (BUG-119, BUG-120). Git-ignored is read with git's own
  answer (`git check-ignore`, or the ls-files read check already makes), once per check, not per path.
- `internal/message/pointer.go` and its tests (BUG-125).
- `packs/base/skills/message/SKILL.md`, `packs/base/skills/memory/SKILL.md` (BUG-129). Reaches other projects only
  with the next base tag (the director's).
- The workflow pack's `skills/orchestrate`, `skills/verification-ladder` in `~/bonsai-checks/bonsai-workflow`
  (BUG-127). Its release and tag are the director's.

## Out of scope

Everything T-6004 keeps (BUG-121 to BUG-124, BUG-126, BUG-128, BUG-130); no guard, ladder, rung-0 or format change.
