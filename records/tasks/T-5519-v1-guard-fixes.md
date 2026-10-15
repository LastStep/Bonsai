---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task
id: T-5519
title: "V1's fix round (5.5.5e): git rm's --no-cached, a bonsai call inside script -c, the minutes count, a protocols folder that is a file"
status: running
lane:
done_when:
  - "M1: `git rm --cached` with `--no-cached` or any prefix of it, in either order, judged as a deletion of .bonsai/local again (git's last-wins); V1's TestV1Adversarial cases as Go tests that fail first"
  - "M2: a command line run by `script -c`, `script -qec` and the like read as sh -c and bash -c are, so a bonsai call or env change in it is judged; a Go test that fails first"
  - "Should-fix 1: a minutes budget spent once the elapsed time passes the limit, not a whole minute after; should-fix 2: commands.go's comment and update's help say what a preview without --yes still writes (the machine record, the plugin install); should-fix 3: a documents.protocols that names an existing file refused as bad-config before anything is written; each with a test that fails first"
  - "Check 10 on both sides; the installed bonsai's ladder green at the branch's HEAD; landed by the landing rule; CI green; V1 re-runs its reproducers (its second round) and passes T-5515 and this"
depends_on: [T-5515]
blocked_by:
created: 2026-10-16
started: 2026-10-16
finished:
labels:
  bonsai.branch: "t5519"
  bonsai.ladder: []
  bonsai.allows: []
  bonsai.wants: []
---

# V1's fix round (5.5.5e)

Piece 5.5.5e: V1 (`~/bonsai-checks/verify-5.5-V1/report.md`) failed `T-5515` on M1, a regression it brought, and found M2,
an older guard hole, and three should-fixes. Fixed forward by a fresh Opus builder; V1's second round re-runs its
reproducers.

- **Owns:** `internal/guard/`, `internal/engine/` where should-fix 3 leads, `cmd/bonsai/` (update's help), and their tests.
- **Grants:** none: no protected path.
- **Proof:** the lines of `done_when`.
