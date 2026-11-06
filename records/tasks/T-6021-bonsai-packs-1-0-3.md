---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task; how to fill: skill base:task
id: T-6021
title: "Bonsai's own repo moves to base-v1.0.3 and the workflow pack's v1.0.2 by bonsai pack set --all, climbed on main before its push"
status: done
lane: null
done_when:
  - "bonsai pack set --all previewed (no loosening, no hook line, no code), applied with --yes --expect <fingerprint>; bonsai pack outdated reads no newer release; bonsai check finds nothing"
  - "The move committed on main and climbed there (bonsai ladder --task T-6021) green before its push; CI green on it"
depends_on: [T-6019]
blocked_by: null
created: 2026-11-06
started: 2026-11-06
finished: 2026-11-06
labels:
  bonsai.ladder: []
  bonsai.allows: ["bonsai.yaml", ".bonsai/lock.json", ".claude/settings.json", "CLAUDE.md"]
  workflow.model: opus
---

# Bonsai's own repo on base 1.0.3 and workflow 1.0.2

After the director's tags of 6 Nov ("tagged both"), as T-6014 did for 1.0.2, by the director's standing word "move
bonsai to new packs" (5 Nov). Learnt in T-6014: a move committed on main had no climb before its push; this one does.
