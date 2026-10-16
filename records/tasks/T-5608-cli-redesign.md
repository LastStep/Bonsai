---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task
id: T-5608
title: "The CLI redesign in style C: colour on a terminal, laid out to its width with golang.org/x/term (5.6.8)"
status: verify
lane:
done_when:
  - "Every word's human output in style C (the director, 16 Oct, from T-5607's images): summary first, columns for lists fitted to the terminal's width through golang.org/x/term, one what/why/next shape for refusals, no repeated footers, init's deny rules as counts with a new --explain flag (an addition); all drawn by one small package the words describe to"
  - "Colour only on a terminal, off under NO_COLOR and in pipes; piped output byte-stable on two runs and on both sides; a Windows console's colour"
  - "Every --json output but --help --json byte-identical before and after on fixture projects, and --help --json differing only by --explain (docs/reference/lists.md regenerated); exit codes and error words unchanged"
  - "Each changed text assertion moved to the new text, none dropped; before/after captures of each word's human output in ~/bonsai-checks/T-5608/captures/"
  - "CLAUDE.md's rule lines and design/bonsai-spec.md's section 3 lines changed on the director's word only as far as C needs (golang.org/x/term allowed; colour codes on a terminal); the start still under 5 ms; the operating skill's test; check 10, the ladder, CI"
depends_on: [T-5607]
blocked_by:
created: 2026-10-16
started: 2026-10-16
finished:
labels:
  bonsai.ladder: []
  bonsai.allows: ["CLAUDE.md", "design/bonsai-spec.md"]
  bonsai.wants: []
  bonsai.branch: T-5608
---

# The CLI redesign in style C (5.6.8)

Piece 5.6.8 of step 5.6: its row in `design/plan-5.md`'s "Added 16 Oct, before the build". The director chose style C
from T-5607's images ("okay lets go with C, color + x/term") and its place ("option 1": next), 16 Oct.

- **Owns:** the human writers in `cmd/bonsai/` and the text lines in `internal/status/` and `internal/engine/` that
  print them, a new `internal/textui/` (its name the builder's), the tests that assert human text, `go.mod` and
  `go.sum` (for `golang.org/x/term`), `CLAUDE.md`'s rule lines and `design/bonsai-spec.md`'s section 3 lines (on the
  director's word, recorded in T-5607's run report), `docs/reference/lists.md`, the operating skills' lines on
  `--explain`.
- **Grants:** `go.mod`, `go.sum`, `CLAUDE.md`, `design/bonsai-spec.md` (protected, none person-only): a change the
  director's choice names.
- **Proof:** the lines of `done_when`.
