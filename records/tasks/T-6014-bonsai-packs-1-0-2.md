---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task; how to fill: skill base:task
id: T-6014
title: "Bonsai's own repo moves to the released packs: base from commit 4b41790b to base-v1.0.2, workflow from commit 1a9f6b4e to v1.0.1, by bonsai pack set"
status: done
lane: null
done_when:
  - "bonsai.yaml's two pack refs read base-v1.0.2 and v1.0.1 (their comments kept), applied in the main checkout by `bonsai pack set <id>@<ref> --yes --expect <the preview's fingerprint>`, each preview read first: no loosening, no hook line, no code"
  - "bonsai check finds nothing; the lock, CLAUDE.md's block and .claude/settings.json as the previews named, committed on main; CI green on that commit"
depends_on: [T-6013]
blocked_by: null
created: 2026-11-05
started: 2026-11-05
finished: 2026-11-05
labels:
  bonsai.ladder: []
  bonsai.allows: ["bonsai.yaml", ".bonsai/lock.json", ".claude/settings.json", "CLAUDE.md"]
  workflow.model: opus
---

# Bonsai's own repo on the released packs

The director, 5 Nov: **"move bonsai to new packs. and tell me how it went"**: their word for the move (CLAUDE.md:
"an update of Bonsai's link that adds no code and leaves Bonsai's lines as they are ... is the orchestrator's, in the
main checkout only, under a task granting bonsai.yaml and the lock"). Both packs were pinned to commits from 16 Oct
and 24 Oct, so `pack set --all` (which leaves a commit-pinned pack) does not move them: each is moved by name.
