---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task
id: T-5599
title: "No person's name in Bonsai's files: the director's name, user name and email replaced by a generic designation"
status: verify
lane:
done_when:
  - "A Sonnet inventory of every hit (name, family name, Windows user name, email) in Bonsai's tracked and untracked files, the branches in flight, workflow, the test pack, the history's counts and GitHub, read by the orchestrator"
  - "Every file under design/ kept exactly as it is too (the director, 15 Oct about 13:45: \"also don't remove from already made plan files\"; asked about the four approved design files: \"option C and also claude.md and state.md\")"
  - "Every file under records/ kept exactly as it is (the director, 15 Oct about 13:35: \"don't rewrite the files in the records folder, keep them as is\"), and the generated .bonsai/tasks.md with them"
  - "Every other hit (the product: code, formats, docs, packs, LICENSE, CONTRIBUTING, the release workflow's comment; and CLAUDE.md and STATE.md) replaced by the inventory's rules: \"the director\" in the project's records, design and rules; \"the person\" in the product's code, formats, docs and packs; the director's pronouns neutral; a Windows path as %USERPROFILE% or <user>"
  - "CLAUDE.md's Safety section says no person's name goes into any file or commit message"
  - "A fresh check after the sweep: the inventory's search on every place finds nothing, and the changed lines read right"
  - "Code changes: check 10 on both sides; the installed bonsai's ladder green at the branch's HEAD; landed by the landing rule; CI green"
depends_on: []
blocked_by:
created: 2026-10-15
started: 2026-10-15
finished:
labels:
  bonsai.branch: "t5599"
  bonsai.ladder: []
  bonsai.allows: ["CLAUDE.md", ".github/workflows/release.yml"]
  bonsai.wants: [".github/workflows/release.yml"]
---

# No person's name in Bonsai's files

Part 5.5's work outside its pieces, on the director's word (15 Oct, about 11:00): "make sure that my name is no where
present. use some generic desigation if you need to. understand that this is a general product which will be used by
everyone. also make sure my name does not come up in bonsai repo as well"; then "send a sonnet agent to check for
that, and then read its message and tell it to remove accordingly".

- **Owns:** every file in Bonsai's repository that holds a hit in the inventory (`~/bonsai-checks/T-5599/inventory.md`).
- **Grants:** set once the inventory is read: the protected and person-only paths among the hits (`CLAUDE.md` and
  `.github/workflows/release.yml`; the approved design files are kept as they are), each changed only by the name's replacement, the director's
  word; then the task runs with no other task `running`.
- **Proof:** the lines of `done_when`; the landing rule for code; CI.
