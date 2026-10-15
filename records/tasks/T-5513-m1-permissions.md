---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task
id: T-5513
title: "VG's second fix round for 5.5.11 (5.5.11b): permission changes on Bonsai's files refused, a call never passed uncounted"
status: done
lane:
done_when:
  - "M1: chmod, chown, chgrp, setfacl and chattr, and on Windows attrib, icacls, Set-ItemProperty, Set-Acl and an IsReadOnly assignment, onto Bonsai's own files refused as bonsai-files from every agent, and inside the project from roles that change nothing; VG round 2's chmod.py and chmod2.py as Go tests that fail first"
  - "M1: no call passes unjudged when its line cannot be written (budget.go's count, roles.go's two fail-open lines): refused with a record, or the unwritable file moved aside and counted afresh"
  - "S2: past a budget, only the hook lines' place counts as the report or a brief's read, and the budget-spent refusal's next step names that place; or the bare form kept and written as an honest limit, with the reason"
  - "S1 where it is small: a fork with no named call joins only a call whose caller matches (the meta file's parent), or open calls expire; else roles.go's and skill.go's limits rewritten to what holds; and the can-wait hookin.SpawnCall meta read through a regular-file reader"
  - "Check 10 on both sides; the installed bonsai's ladder green at the branch's HEAD; landed by the landing rule; CI green; a fresh verifier's short third check on these fixes passes (the director's A, 15 Oct)"
depends_on: [T-5598]
blocked_by:
created: 2026-10-15
started: 2026-10-15
finished: 2026-10-15
labels:
  bonsai.branch: "t5513"
  bonsai.ladder: []
  bonsai.allows: []
  bonsai.wants: []
---

# VG's second fix round for 5.5.11 (5.5.11b)

Piece 5.5.11b, the next free number: VG round 2 (`T-5596`) failed 5.5.11 on one new must-fix, M1, with should-fixes
S1 and S2 (its report: `~/bonsai-checks/verify-5.5-VG2/report.md`; `R-2026-10-15-T-5596-verify-vg.md`, 13:01). The
director chose (15 Oct, about 13:05) "A: fix + short 3rd check": fixed forward by a fresh Opus fixer, then a fresh
verifier checks only these fixes.

- **Owns:** `internal/guard/` (the write tables, `budget.go`, `roles.go`, `rolecmds.go`, the known-limits comments);
  `internal/hookin/`'s meta read; `internal/agents/` where a reader is shared (and, added at the fixer's report, `aside.go`'s `Repair` and `Asides`, which part 2 needs); `docs/reference/lists.md`
  (regenerated); their tests.
- **Grants:** none: no protected path, so `bonsai.allows` is empty.
- **Proof:** the lines of `done_when`; the landing rule; CI; the third check.
