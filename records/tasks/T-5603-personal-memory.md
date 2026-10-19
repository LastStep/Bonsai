---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task
id: T-5603
title: "The personal memory layer and its check (5.6.3)"
status: done
lane:
done_when:
  - "the empty personal index made by init and update; the five warnings (form, budgets, index and notes, secrets, the Windows copy); the Windows copy refreshed from WSL's; base's memory skill and the operating skill's memory line on who writes it; the memory caps in one table; the engine's deny rules over the other side's Bonsai home; the import line's form measured"
  - "Go tests on scratch homes for each warning, never the exit code, no secret printed; nothing written by a preview or by check; the copy refreshed and compared by update, the unreachable source within 2 s; the import tried in a scratch project's session; the cross-side rules tried once on each side"
  - "Human text in style C (internal/textui); check 10 on WSL, lint, the ladder, CI; the Windows half (the copy on Windows, the cross-side rule there, check 10) at 5.6's end run, by the director's word of 18 Oct"
depends_on: [T-5601]
blocked_by:
created: 2026-10-18
started: 2026-10-18
finished: 2026-10-19
labels:
  bonsai.ladder: []
  bonsai.allows: []
  bonsai.wants: []
  bonsai.branch: T-5603
---

# The personal memory layer and its check (T-5603)

Piece 5.6.3 of `design/plan-5.md` (its row, `^| 5.6.3 |`; its note from line 6513; the owns table's row). Under the
director's (A) of 10 Oct it needs no verifier of its own: the 5.6 end verifier breaks its warnings and the Windows copy
and reads its rules.

- **Owns:** new `internal/workspace/personal.go` and `internal/engine/checkpersonal.go`; `init`'s and `update`'s call to
  them in `internal/engine/`; `internal/engine/checkdocs.go` (the budgets become the caps' table); the engine's `Own`
  deny lines (5.3.5's table beside `ownHooks`, `internal/engine/`); `internal/format/check.go` (its five words);
  `packs/base/skills/memory/SKILL.md` and the operating skill's memory line; `docs/reference/lists.md`; their tests.
- **Added 18 Oct (the audit `M-2026-10-18-fd2c4800`):** `cmd/bonsai/engine.go` (init's and update's after-write steps, beside the cache clean); `internal/engine/check.go` (the one call to the new check); `internal/reference/reference.go` (the caps' section the reference page is generated from); `cmd/bonsai/personal_test.go` and `internal/format/lists_test.go`.
- **Grants:** none (no protected path).
- **Proof:** the lines of `done_when`.
