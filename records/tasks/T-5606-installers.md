---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task
id: T-5606
title: "The installers and `install.json` (5.6.6)"
status: done
lane:
done_when:
  - "install/install.sh and install/install.ps1, each held to the two installed places and the record's keys; the Windows elevated child's code inline; an install over a running copy; scratch targets in tests"
  - "Go tests running each script on scratch targets and a scratch home (install, the record read by check, again, again while a copy runs, remove, each refusal)"
  - "A CI job running both at their real places on throwaway machines, its commit last and made only when the orchestrator says (it needs .github/workflows/ci.yml granted while T-5606 runs alone); check 10 both sides, the ladder, CI"
depends_on: [T-5600]
blocked_by:
created: 2026-10-16
started: 2026-10-16
finished: 2026-10-19
labels:
  bonsai.ladder: []
  bonsai.allows: [".github/workflows/ci.yml"]
  bonsai.wants: []
  bonsai.branch: T-5606-fix1
---

# The installers and `install.json` (5.6.6)

Piece 5.6.6 of step 5.6: its row in `design/plan-5.md`'s "#### The pieces and their order" and its note in "#### Notes
per piece" (lines 6682-6785). One of the four run side by side after 5.6.0 ("The order", step 2).

- **Owns:** `install/` (both scripts and their Go test); one `.gitattributes` line (protected: granted when it runs alone); its last commit: the `install` job in `.github/workflows/ci.yml` (protected, the same).
- **Grants:** none while others run; the `.gitattributes` line and the CI job are granted in a later window when it runs alone.
- **Proof:** the lines of `done_when`.
