---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task
id: T-5609
title: "operating-bonsai split in two: a core every agent reads and a second skill for linking, updating and settings (5.6.9)"
status: done
lane:
done_when:
  - "packs/base/skills/operating-bonsai/SKILL.md is the core every agent reads before any bonsai call (--json and refusals, status, check, the ladder, asks, briefs and budgets, the guard, the person's gates); a second base skill holds linking, updating, editing bonsai.yaml, status moves and this machine's settings"
  - "base's block line names both; base's README and every base skill line that names operating-bonsai (message's two) point to the right half"
  - "cmd/bonsai/operating_test.go holds every bonsai word to one of the two and each skill to its cap, and its failing copies fail (a word in neither; either past its cap)"
  - "bonsai check clean; check 10 on both sides; a green climb by the installed bonsai at the branch's HEAD; CI after the landing"
depends_on: [T-5690]
blocked_by:
created: 2026-10-16
started: 2026-10-16
finished: 2026-10-16
labels:
  bonsai.ladder: []
  bonsai.allows: []
  bonsai.wants: []
  bonsai.branch: T-5609
---

# operating-bonsai split in two (5.6.9)

Piece 5.6.9 of step 5.6: its row in `design/plan-5.md`'s "Added 16 Oct, before the build" (the director's decisions,
16 Oct; `T-5690`). The skill is at 12,262 of its 12,288-byte cap and 5.6 adds lines to it.

- **Owns:** `packs/base/skills/operating-bonsai/SKILL.md`, the new skill's folder, `packs/base/bonsai/block.md`, base's
  README, `cmd/bonsai/operating_test.go`, and every base skill line that names `operating-bonsai`
  (`packs/base/skills/message/SKILL.md`'s two).
- **Grants:** none: every path it owns is free.
- **Proof:** the lines of `done_when`. Bonsai's own `base` ref moves to it in `T-5691` onward, on the director's yes.
