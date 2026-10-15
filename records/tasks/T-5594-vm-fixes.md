---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task
id: T-5594
title: "VM's fix round for 5.5.0 and 5.5.10: a pointer read only from a regular file, and an unknown limit name named"
status: done
lane:
done_when:
  - "D1: a pointer is fingerprinted only from a regular file (a FIFO, a device, a link to either, a folder refused at writing as pointer-missing and shown missing at --show), hashed by streaming with the LF rule kept; Go tests for each, Windows' nul among them"
  - "D2: an unknown limit name under a packs entry's roles never silences check or blames the lock: check names bonsai.yaml's line, bonsai message and status name bonsai.yaml with a next step that fixes it, every other role finding still raised"
  - "The can-wait items in the same files: D3's threshold compared without overflow; D4's 16-digit budget refused bad-value; D5's committed copy that does not read never silently replaced by the working copy in message.go and status.go; D6's recipient folders that differ only by case refused or told apart; VM's four wording points"
  - "5.5.11's three follow-ups in 5.5.10's files: status's messages.flagged reads the agents' state through agents.Store.Thread; bonsai message --show --json's thread carries its recipient; the until-5.5.11 wording gone from internal/message/thread.go and format.MessageFlags' text, the reference page regenerated"
  - "Check 10 on both sides; the installed bonsai's ladder green at the branch's HEAD; landed by the landing rule; CI green; VM passes 5.5.0 and 5.5.10 on its second round"
depends_on: [T-5593]
blocked_by:
created: 2026-10-15
started: 2026-10-15
finished: 2026-10-15
labels:
  bonsai.branch: "t5594"
  bonsai.ladder: []
  bonsai.allows: []
  bonsai.wants: []
---

# VM's fix round for 5.5.0 and 5.5.10

Part 5.5's work outside its pieces: VM's (`T-5593`) must-fixes D1 and D2 and its can-wait items in the same files,
fixed forward by one fresh Opus fixer (both pieces' builders past the ~500k hand-over line). VM's report is in
`records/runs/R-2026-10-15-T-5593-verify-vm.md`.

- **Owns:** `internal/message/` (pointer, thread, write, address); `format.MessageFlags`' text in `internal/format/` and
  `docs/reference/lists.md` (regenerated); `cmd/bonsai/message.go`, `messages.go` and their
  tests; `internal/workspace/roles.go`; `internal/engine/checkroles.go` and `checkpack.go`'s deny-rule suggestion;
  `internal/status/`; the `.bonsai/local/` summary lines in `internal/engine/` and `internal/status/`; their tests.
- **Grants:** none: no protected path, so `bonsai.allows` is empty.
- **Proof:** the lines of `done_when`; the landing rule; CI; VM's second round.
