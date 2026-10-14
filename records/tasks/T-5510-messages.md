---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task
id: T-5510
title: "Messages between agents: the reader, bonsai message and bonsai messages (5.5.10)"
status: todo
lane:
done_when:
  - "internal/message reads bonsai.message/1; bonsai message writes a brief or a report into its recipient's folder, its pointers fingerprinted; bonsai messages lists them"
  - "Go tests on fixture folders: every field, kind and status; a malformed file skipped and listed; two writers at once never colliding; a pointer whose file changed since; a budget over its max refused; a report into a thread with no brief to its --from refused; an open thread never cleaned; the circling flag at its threshold and not below it, at the default and at a project's own"
  - "Every output against its schema; the trust line in every --show; Windows renames retried on busy errors; scripted runs on both sides"
  - "Its last commits (the limits through 5.5.0's function, --to against the lock's roles, the flags' thresholds, check's malformed-message warning, messages in status, the messages kind in the cleaning) made one at a time after T-5500 has landed"
  - "Check 10 on both sides; the installed bonsai's ladder green at the branch's HEAD; landed by the landing rule; CI green; VM passes"
depends_on: [T-5509]
blocked_by:
created: 2026-10-14
started:
finished:
labels:
  bonsai.branch: "t5510"
  bonsai.ladder: []
  bonsai.allows: []
  bonsai.wants: []
---

# Messages between agents: the reader, bonsai message and bonsai messages (5.5.10)

Piece 5.5.10 of step 5.5 (approved 14 Oct): its row in `design/plan-5.md`'s "### Step 5.5" "The pieces and their order", and its note in "Notes per piece". The shape is the section's "Messages between agents, written once".

- **Owns:** `internal/message/` (new); `cmd/bonsai/message.go` and `cmd/bonsai/messages.go`, their words in the registry and the flag table; after 5.5.0 lands: the limits read through its function, `--to` against the lock's roles, the flags' thresholds from `bonsai.yaml`, the malformed-message warning in `check`'s findings table, `messages` in `internal/status/`, the `messages` kind in `internal/clean/`.
- **Grants:** none: no protected path among its files, so `bonsai.allows` is empty.
- **Proof:** the lines of `done_when`; the installed `bonsai`'s ladder green at the branch's HEAD; the landing rule; CI. VM.
