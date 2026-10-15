---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task
id: T-5514
title: "VG's third check's fix for 5.5.11 (5.5.11c): past a budget only this system's installed place, and a rewritten count held spent"
status: done
lane:
done_when:
  - "S2 must-fix: in an unstamped build, past a budget only this system's hook-line place counts as the report or a brief's read (PlaceLinux on Linux, PlaceWindows in its forms on Windows); the other system's place refused; VG3's s2plain.py and replay-w4.ps1 as Go tests that fail first, unstamped"
  - "Should-fix: a calls file shorter than its window's start (total below the window's first call) held spent, as a second unreadable state is; budget.go's limit text says what a write word outside the tables can still do"
  - "Where small: the root folder reached by mayReach (a recursive chmod of the root, icacls on the drive root), and PowerShell's parenthesised and chained IsReadOnly assignments; else each named as a limit"
  - "Check 10 on both sides; the installed bonsai's ladder green at the branch's HEAD; landed by the landing rule; CI green; VG3 re-runs its reproducers and passes 5.5.11"
depends_on: [T-5513]
blocked_by:
created: 2026-10-15
started: 2026-10-15
finished: 2026-10-15
labels:
  bonsai.branch: "t5514"
  bonsai.ladder: []
  bonsai.allows: []
  bonsai.wants: []
---

# VG's third check's fix for 5.5.11 (5.5.11c)

Piece 5.5.11c, the next free number: VG's short third check (`T-5596`, the director's "A: fix + short 3rd check")
passed M1, S1 and the can-wait items and failed 5.5.11 on one must-fix in S2, older than the round (its report:
`~/bonsai-checks/verify-5.5-VG3/report.md`). Fixed forward by a fresh Opus fixer; the same verifier re-runs its
reproducers ("each must-fix fixed forward and passed by the same verifier").

- **Owns:** `internal/guard/rolecmds.go` (`hookPlaces`, `installedBonsai`) and its tests; `internal/guard/budget.go`
  (the count's backstop and its limit text) and its tests; `internal/guard/attrwords.go` where small; their tests.
- **Grants:** none: no protected path, so `bonsai.allows` is empty.
- **Proof:** the lines of `done_when`; the landing rule; CI; VG3's re-run.
