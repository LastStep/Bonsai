---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task
id: T-5520
title: "V1's second round's should-fixes (5.5.5f): the unread runners and a git alias in the line written as known limits; a protocols folder that is CLAUDE.md refused at a first link"
status: done
lane:
done_when:
  - "internal/guard/shell.go's known-limits text names the runners V1 confirmed unread (tmux, screen, expect, systemd-run, nsenter, ssh, strace/ltrace, at, firejail, bwrap) and a git alias defined in the line (`git -c alias.x=!...`, GIT_CONFIG_PARAMETERS), beside parallel; no guard logic changed"
  - "A documents.protocols that equals or holds a reserved target (CLAUDE.md in any case, and the other files Bonsai writes at the top) refused as bad-config before anything is written, at a first link where the file does not exist yet as well as at update; a test that fails first (V1's proto-cases2.sh case r1-first)"
  - "Check 10 on both sides; the installed bonsai's ladder green at the branch's HEAD; landed by the landing rule; CI green; the orchestrator's read of the diff"
depends_on: [T-5519]
blocked_by:
created: 2026-10-16
started: 2026-10-16
finished: 2026-10-16
labels:
  bonsai.branch: "t5520"
  bonsai.ladder: []
  bonsai.allows: []
  bonsai.wants: []
---

# V1's second round's should-fixes (5.5.5f)

Piece 5.5.5f: V1's second round passed `T-5515` with `T-5519` and left three should-fixes
(`~/bonsai-checks/verify-5.5-V1/report-r2.md`). Two are written as known limits, as the guard's design treats the
interpreters it leaves (a tripwire, with rung 0 and the landing rule behind it); one is fixed.

- **Owns:** `internal/guard/shell.go`'s limits text; `internal/engine/` where the protocols check lives; their tests.
- **Grants:** none: no protected path.
- **Proof:** the lines of `done_when`.
