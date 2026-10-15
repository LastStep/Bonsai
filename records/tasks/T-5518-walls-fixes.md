---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task
id: T-5518
title: "5.5.6's fix round for the walls (5.5.6b): the unwalled credential files walled; each side's limits written in the README"
status: done
lane:
done_when:
  - "base's walls cover the credential files 5.5.6's listing found present and unwalled (the Codex, Gemini and Hugging Face token files and the two tool folders), on both sides, each with its why, in the forms 5.5.6 found to hold; base's pack test green"
  - "base's README (and workflow's, for its studio walls) says, with why, what 5.5.6 measured does not hold: on Windows the PowerShell tool ignores Read and Edit deny rules, and Git Bash passes the WSL home's network-path forms; no rule form holds there"
  - "Check 10 on both sides; the installed bonsai's ladder green at the branch's HEAD; landed by the landing rule; CI green; workflow's change grepped and read before its push, its CI green; V1 reads it"
depends_on: [T-5506]
blocked_by:
created: 2026-10-16
started: 2026-10-16
finished: 2026-10-16
labels:
  bonsai.branch: "t5518"
  bonsai.ladder: []
  bonsai.allows: []
  bonsai.wants: []
---

# 5.5.6's fix round for the walls (5.5.6b)

Piece 5.5.6b: the walls tried on both sides (`T-5506`, report `~/bonsai-checks/T-5506/report.md`) held for every file
tool and for shells on `~/` forms, but found five credential files present and unwalled, and two Windows paths no rule
form holds (the PowerShell tool; Git Bash on the WSL home's network paths). Note 5.5.6 item 4: "a rule that does not
hold on a side gets the form that does ... or a README line saying it is one side's, with why".

- **Owns:** `packs/base/bonsai/pack.yaml`'s walls, `packs/base/README.md` and their tests; `workflow`'s `README.md`
  (its walls' lines only).
- **Grants:** none: no protected path.
- **Proof:** the lines of `done_when`.
