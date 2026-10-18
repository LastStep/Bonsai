---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task
id: T-5601
title: "`settings`, the stranded folder and the cache (5.6.1)"
status: verify
lane:
done_when:
  - "settings show and set, --machine, the keys in one Go table; the refusal in an agent session (5.6.0's personOnly); the stranded warning with a person's exact lines; update cleaning old pack clones by cache_keep_days"
  - "Go tests on t.TempDir() homes and projects: every key and bad value, all or nothing, byte-stable files, a stranded folder found and brought over, clones cleaned but the lock's and never an --adopt copy; the refusal with the variable set; a status_command outside its rule refused by the reader, the guard failing closed on it"
  - "Human text in style C (internal/textui); check 10 both sides, the ladder, CI"
depends_on: [T-5600]
blocked_by:
created: 2026-10-16
started: 2026-10-16
finished:
labels:
  bonsai.ladder: []
  bonsai.allows: []
  bonsai.wants: []
  bonsai.branch: T-5601
---

# `settings`, the stranded folder and the cache (5.6.1)

Piece 5.6.1 of step 5.6: its row in `design/plan-5.md`'s "#### The pieces and their order" and its note in "#### Notes
per piece" (lines 6415-6493). One of the four run side by side after 5.6.0 ("The order", step 2).

- **Owns:** `cmd/bonsai/settings.go` (the run functions); `internal/workspace/machine.go` (the writes and the keys table) and a new `internal/workspace/homesettings.go`; `internal/engine/checkmachine.go` (`stranded`) and `check.go`'s `checkLater`; `internal/format/check.go` (its word); a new `internal/engine/cacheclean.go` and the fetch's last-use mark; `internal/status/status.go`'s text (the stranded line); `docs/reference/lists.md`; the operating-setup lines for `settings`; `cmd/bonsai/word.go`'s help for a word whose sub-words are all Later (the empty "Answers" heading, `settings show --help` refused: T-5600's open items).
- **Added 18 Oct (the audit `M-2026-10-18-170e4b8d`):** the note's own tests (`cmd/bonsai/settings_test.go`, `stranded_test.go`, `internal/guard/statuscommand_test.go`, `internal/workspace/homesettings_test.go`, `internal/engine/cacheclean_test.go`); `internal/format/lists_test.go` (`stranded` added to the machine words); `cmd/bonsai/engine.go` (`runEngine`'s call to the cache clean after a written run) and `cmd/bonsai/update.go`'s help text, for item 6.
- **Grants:** none while others run.
- **Proof:** the lines of `done_when`.
