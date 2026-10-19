---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task
id: T-5612
title: "5.6 end check 15: the 11 narrowed human-text assertions held to the full new wording"
status: done
lane:
done_when:
  - "each of the 11 narrowings in the audit M-2026-10-19-ec5723dd holds the full new text again: its label words, its line ends and anchors, its adjacency, its indent where the layout fixes one; none of the ~180 others weakened"
  - "each tightened assertion shown failing on a temporary copy of the output with the dropped part changed"
  - "WSL check 10, lint, a green climb"
depends_on: [T-5611]
blocked_by:
created: 2026-10-19
started: 2026-10-19
finished: 2026-10-19
labels:
  bonsai.ladder: []
  bonsai.allows: []
  bonsai.wants: []
  bonsai.branch: T-5612
---

# The 11 narrowed assertions tightened (T-5612)

The director, 19 Oct: "A. Tighten all 11" (5.6 end check 15; `update`'s `--explain` beside `init`'s accepted as written
in the operating skill).

- **Owns:** the 11 test files' lines the audit names: `cmd/bonsai/engine_test.go`, `consent_test.go`,
  `checktable_test.go`, `message_test.go`, `main_test.go`, `agree_test.go`; `internal/status/status_test.go`; and the
  shared test helpers (`flat`, `nextOf`) if a tighter helper is needed.
- **Grants:** none.
