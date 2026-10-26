---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task; how to fill: skill base:task
id: T-6002
title: "A checkout with Git LFS climbs clean: an LFS file checked by its pointer's SHA-256 (BUG-054, the studio's B-053; a 1.0.1 candidate)"
status: done
lane: null
done_when:
  - "The plan records/plans/P-T-6002.md was reviewed by a fresh Opus agent and approved by the director"
  - "In a repository with a Git LFS file committed and unchanged, a climb reads git.dirty false, on WSL and natively on Windows with real git-lfs; the same file edited (same size or not) reads dirty"
  - "Every test the plan lists passes, in WSL and natively on Windows, and CI on the branch is green"
  - "git-crypt and other filters still read dirty, with a reason that says so; BUG-054 is closed and their limit is a bug of its own in records/bugs.md"
  - "A fresh Opus verifier (security and tests lenses) passes the change"
depends_on: []
blocked_by: null
created: 2026-10-26
started: 2026-10-26
finished: 2026-10-26
labels:
  bonsai.ladder: []
  bonsai.allows: []
  bonsai.wants: []
  bonsai.branch: t-6002-lfs-clean
  workflow.model: opus
  workflow.estimate_h: 6
---

# A checkout with Git LFS climbs clean

## Context

The studio's move 2 (26 Oct): in Mimas (Windows, 216 Git LFS files, linked at 1.0.0) every climb is green with
`dirty: true`. Reproduced on WSL and Windows with the installed 1.0 (run report
`records/runs/R-2026-10-26-T-6002-lfs-dirty.md`). The cause is the dirty check's read of the contents with filters
off (`internal/vergit/tree.go`), a known limit (BUG-054). The director decides whether the fix ships as 1.0.1.

## Scope

`internal/vergit` (a new `lfs.go`, `tree.go`), `internal/ladder/text.go` (the reason and next step), their tests, the
known-limits text in `vergit.go`, `docs/features.md`, `records/bugs.md`. No format, schema, contract or bonsai.yaml
change. The plan: `records/plans/P-T-6002.md`.
