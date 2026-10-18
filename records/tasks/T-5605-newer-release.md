---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task
id: T-5605
title: "A newer release, known and handed over (5.6.5)"
status: verify
lane:
done_when:
  - "status --full reads Bonsai's release tags (at most daily, 10 s), the answer kept in the home's cache; bonsai_release in status --json, a check warning (bonsai-newer) and the line's marker, offline from it; both sides' install lines written by Bonsai"
  - "Go tests against a local bare repository with tags (a stamped source), the record's times faked; no network read within a day; unknown and not-a-release; the lines' exact text, their names held to .goreleaser.yaml; a forged record giving no lines; one read of the real repository by the plain build on a scratch home"
  - "Human text in style C; check 10 on WSL, lint, the ladder, CI; the Windows half at 5.6's end run, by the director's word of 18 Oct"
depends_on: [T-5603, T-5604]
blocked_by:
created: 2026-10-18
started: 2026-10-18
finished:
labels:
  bonsai.ladder: []
  bonsai.allows: []
  bonsai.wants: []
  bonsai.branch: T-5605
---

# A newer release, known and handed over (T-5605)

Piece 5.6.5 of `design/plan-5.md` (its row, `^| 5.6.5 |`; its note from line 6644; the owns table's row): the
director's 9 Oct addition (spec §3's note).

- **Owns:** a new `internal/engine/release.go` and its test (the names held to `.goreleaser.yaml`, read only);
  `cmd/bonsai/status.go` (`--full`'s release read); `internal/status/` (`bonsai_release` filled, its text line,
  `line.go`'s marker); `internal/engine/checkmachine.go` (`bonsai-newer`); `internal/format/check.go`; the operating
  skill's install line; `docs/reference/lists.md`; their tests.
- **Added 18 Oct (the audit `M-2026-10-18-412e8ad9`):** `cmd/bonsai/check.go` (the one field passing the running version to `CheckWith`); `cmd/bonsai/checktable_test.go` (the new word's case); `internal/engine/newer.go` (`RemoteTags` takes its timeout, so the release read and the packs' read share one reader).
- **Grants:** none (`.goreleaser.yaml` read only).
- **Proof:** the lines of `done_when`.
