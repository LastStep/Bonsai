---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task
id: T-5600
title: "The formats set and the two words' places: settings, attach and line formats, bonsai_release, settings and labels registered (5.6.0)"
status: done
lane:
done_when:
  - "One formats set after the last landed: bonsai.settings/1, bonsai.attach/1, bonsai.line/1, bonsai_release at the end of status, the memory schema's words on the personal layer; their Go types and writers; status writing bonsai_release: null"
  - "settings and labels registered with their sub-words not built yet: flag tables, exit codes, help, the error words both need; the agent-session refusal both call, with the stamped build's home confinement; the operating skill's lines for both (operating-setup)"
  - "The formats test (manifest, docs, examples); the schema compare (additions only); each new example validated by its writer; the operating skills' test; --help --json listing both words; the refusal and its confinement (a stamped build free only with its home inside its scratch root; a plain build never)"
  - "Human text in style C through internal/textui (5.6.8); check 10, the ladder, CI"
depends_on: [T-5608]
blocked_by:
created: 2026-10-16
started: 2026-10-16
finished: 2026-10-18
labels:
  bonsai.ladder: []
  bonsai.allows: []
  bonsai.wants: []
  bonsai.branch: T-5600
---

# The formats set and the two words' places (5.6.0)

Piece 5.6.0 of step 5.6: its row in `design/plan-5.md`'s "#### The pieces and their order" and its note in "#### Notes
per piece" (from 6368). Built on `T-5608`'s branch so its text is in style C; it lands after `T-5608`.

- **Owns:** as the plan's owns table gives 5.6.0: `formats/` (three new schemas, `status`'s and `memory`'s, their
  examples, README, manifest); the three outputs' Go types and writers and the new error words in `internal/format/`;
  `internal/status/` for `bonsai_release: null` only; `cmd/bonsai/settings.go` and `cmd/bonsai/labels.go`; the file
  holding the agent-session check; the operating skill's lines for both words; `docs/reference/lists.md`.
- **Grants:** none: every path it owns is free.
- **Proof:** the lines of `done_when`.
