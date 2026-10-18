---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task
id: T-5694
title: "V1 round 1, B1: engine.RealPath and the guard's resolve fail closed past their 64-step limit"
status: done
lane:
done_when:
  - "engine.RealPath (internal/engine/hookline.go) and guard.resolve (internal/guard/paths.go) never return a path as written when the walk runs out: the limit counts link hops only, and a walk past it fails closed (FolderHolds false; the guard refuses), each with a test holding a path deeper than the old limit behind a link out of the root (shown failing first, V1's try in M-2026-10-18-8ba15229)"
  - "the two walks one home if it can be done without a dependency the guard must not take; else both, each naming the other"
  - "WSL check 10, lint, a green climb; V1 round 2 reads it"
depends_on: [T-5692, T-5693]
blocked_by:
created: 2026-10-18
started: 2026-10-18
finished: 2026-10-19
labels:
  bonsai.ladder: []
  bonsai.allows: []
  bonsai.wants: []
  bonsai.branch: T-5694
---

# V1 round 1, B1: the path walks fail closed (T-5694)

V1 (`T-5693`, report `M-2026-10-18-8ba15229`) reproduced on WSL: a stamped build's home behind a link out of its root,
then 70 folders not made yet, was let through (`settings set --machine` wrote outside the root); 60 folders were refused.
The orchestrator found the same walk in the guard's `resolve`.

- **Owns:** `internal/engine/hookline.go` (`RealPath`, `FolderHolds`) and its tests; `internal/guard/paths.go`
  (`resolve`, and `procLink` if it shares the flaw) and its tests.
- **Added 18 Oct (the audit `M-2026-10-18-7704b238`):** `internal/guard/rolecmds.go` and `internal/guard/attrwords.go` (a shell write target whose path forms cannot be read is refused, not allowed: Q3 (i), needed since this task makes such errors reachable); the guard's `finalPath` moved into `internal/engine` for one home and used by `RealPath` on Windows (junctions, Q3 (ii)); `install/install_test.go:192` (RealPath's new signature, one line); V1's N1 (a relative link target not made yet joined by spelling) in the same loop. Their tests.
- **Grants:** none.
