---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task; how to fill: skill base:task
id: T-6006
title: "Bonsai 1.0.1's notes: CHANGELOG's [1.0.1] section from T-6002 to T-6005, landed with CI green, for the director's tag"
status: done
lane: light
done_when:
  - "CHANGELOG.md holds `## [1.0.1] - 2026-10-26` above 1.0.0, every line traced to a bug fixed in T-6002, T-6003, T-6004 or T-6005 (records/bugs.md's Fixed table), in the 1.0.0 section's form"
  - "`sh .github/scripts/release-notes.sh v1.0.1 <out>` writes the section; landed on main with CI green, every job; the director given the tag's commands"
depends_on: [T-6002, T-6003, T-6004, T-6005]
blocked_by: null
created: 2026-10-26
started: 2026-10-26
finished: 2026-10-26
labels:
  bonsai.ladder: []
  bonsai.allows: []
  bonsai.wants: []
  bonsai.branch: t-6006-notes-1-0-1
  workflow.model: opus
  workflow.estimate_h: 0.5
---

# Bonsai 1.0.1's notes

The director, 26 Oct 14:08: "ok, land it once windows is green. and then we will tag to 1.0.1". The release takes
its notes from CHANGELOG's version section (`.github/scripts/release-notes.sh`), so the section lands before the tag.
Only the director tags `v1.0.1` and approves the release run.
