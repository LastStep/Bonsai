---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task
id: T-5509
title: Formats set 8, the contract's new section and the spec's notes (5.5.9)
status: verify
lane:
done_when:
  - "Formats set 8: bonsai.message/1, bonsai.messages/1 and bonsai.agents/1 with their examples and trick files; roles in pack, workspace and lock; messages in workspace; roles and messages at the end of status --json; the message, spawn and usage log events and the log's parent and usage fields; the manifest and README"
  - "Each new field joins its object's required, and every writer and fixture it reaches is written to it (the guard's log line, init's template, the test pack's commit H)"
  - "The formats test and the schema compare pass on set 8; the compare fails in a temporary copy on a removal and on a new property left out of required; the reference page's test passes"
  - "check --schema bonsai.message and the changed formats print the new words; check --pack on the test pack finds one finding at G and none at H"
  - "design/contract.md's new section and design/bonsai-spec.md's dated notes, in one last commit made while T-5509 is the only task running, read by the orchestrator against the section"
  - "Check 10 on both sides; the installed bonsai's ladder green at the branch's HEAD; landed by the landing rule; CI green; VM passes"
depends_on: [T-5590]
blocked_by:
created: 2026-10-14
started: 2026-10-14
finished:
labels:
  bonsai.branch: "t5509"
  bonsai.ladder: []
  bonsai.allows: ["design/contract.md", "design/bonsai-spec.md", "formats/active-task/**/bonsai.yaml"]
  bonsai.wants: ["formats/active-task/**/bonsai.yaml"]
---

# Formats set 8, the contract's new section and the spec's notes (5.5.9)

Piece 5.5.9 of step 5.5 (approved 14 Oct): its row in `design/plan-5.md`'s "### Step 5.5" "The pieces and their order", and its note in "Notes per piece".

- **Owns:** `formats/` (the three new schemas and the additions to `pack`, `workspace`, `lock`, `log` and `status`; README; manifest); the Go tables beside the formats' types in `internal/format/`; the schema registry and its test; the new tables in `internal/reference/reference.go`; `docs/reference/lists.md`; the writers a new required field reaches (note 5.5.9, 3a); after 5.5.8 lands, in one commit, `internal/testpack/testpack.go` and the test-pack case in `cmd/bonsai/checktable_test.go`; commit H of the test pack (pushed by the orchestrator); its last commit: `design/contract.md` and `design/bonsai-spec.md`.
- **Grants:** `design/contract.md` and `design/bonsai-spec.md` (protected, not person-only), for its last commit alone, made while it is the only task `running`. And `formats/active-task/**/bonsai.yaml` (person-only by the guard's floor, which guards every file named `bonsai.yaml`): the active-task fixtures' workspace files, test data, which gain set 8's new required fields and the new `named-garbage` case (note 5.5.9, 3a and 4: every writer and fixture a new required field reaches lands with the set); additions only, so no check loosens (the stricter-only rule's "a change a section Rohan approved names").
- **Proof:** the lines of `done_when`; the installed `bonsai`'s ladder green at the branch's HEAD; the landing rule; CI. VM.
