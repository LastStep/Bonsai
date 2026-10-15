---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task
id: T-5503
title: The operating Bonsai skill in base, kept in step with the commands by a test (5.5.3)
status: done
lane:
done_when:
  - "packs/base/skills/operating-bonsai/ written with messages, budgets and the guard's new refusals; its line in base's block.md and its row in base's README"
  - "cmd/bonsai/operating_test.go: a skill line naming an unknown word or flag fails; a word the registry has and the skill never names fails"
  - "The skill's size within the section's bound; its real sessions run in T-5505"
  - "Check 10 on both sides; the installed bonsai's ladder green at the branch's HEAD; landed by the landing rule; CI green"
depends_on: [T-5502, T-5510, T-5511, T-5512]
blocked_by:
created: 2026-10-14
started: 2026-10-15
finished: 2026-10-15
labels:
  bonsai.branch: "t5503"
  bonsai.ladder: []
  bonsai.allows: []
  bonsai.wants: []
---

# The operating Bonsai skill in base, kept in step with the commands by a test (5.5.3)

Piece 5.5.3 of step 5.5 (approved 14 Oct): its row in `design/plan-5.md`'s "### Step 5.5" "The pieces and their order", and its note in "Notes per piece".

- **Owns:** `packs/base/skills/operating-bonsai/`; its line in base's `block.md` and its row in base's README; `cmd/bonsai/operating_test.go`.
  Added 15 Oct (the builder's findings): base's README intro and its note on the skills that are not templates, where
  they leave the operating skill out; `packs/base/skills/message/SKILL.md`'s PowerShell line (`--body-file`, refused past
  a budget) and its `body` row, kept true to the report form past a budget.
- **Grants:** none: no protected path among its files, so `bonsai.allows` is empty.
- **Proof:** the lines of `done_when`; the installed `bonsai`'s ladder green at the branch's HEAD; the landing rule; CI.
