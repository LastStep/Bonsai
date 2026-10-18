---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task
id: T-5698
title: "Small fixes from the bugs register: BUG-010, 012, 024, 030-032, 082 (code side), 084, 085"
status: done
lane:
done_when:
  - "each bug fixed in the code, with a test shown failing first where a test can hold it; a bug that turns out not to be one, or bigger than small, reported and left alone"
  - "BUG-082: only the code, the comments and the regenerated docs; formats/ untouched (the formats are final: the schema's words wait)"
  - "WSL check 10, lint, a green climb"
depends_on: [T-5696]
blocked_by:
created: 2026-10-19
started: 2026-10-19
finished: 2026-10-19
labels:
  bonsai.ladder: []
  bonsai.allows: []
  bonsai.wants: []
  bonsai.branch: T-5698
---

# Small fixes from the bugs register (T-5698)

The director, 19 Oct: "fix the bugs which you can rightfully tonight, no need to overdo it, and leave the ones you wont
fix". Each row in `records/bugs.md`.

- **Owns:** `internal/engine/checkdocs.go` (BUG-010); `cmd/bonsai/update.go` (BUG-012); the guard's refusal writer
  (BUG-024); `internal/recorder` and `internal/guard`'s task-allows read, `underHome`/`fromHome`,
  `internal/workspace`'s VerifyMain (BUG-030-032); `cmd/bonsai/word.go`, `internal/format/error.go`,
  `internal/workspace/checkout.go`, `internal/record/self.go`, `docs/reference/` (BUG-082, regenerated);
  `internal/guard/shell.go` (BUG-084); `internal/reference/gen/main.go` (BUG-085); and their tests.
- **Grants:** none.
