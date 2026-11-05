---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task; how to fill: skill base:task
id: T-6012
title: "Bonsai 1.0.2's notes: CHANGELOG's Unreleased section becomes [1.0.2], the pack command (T-6010, T-6011), landed with CI green, for the director's tag"
status: done
lane: light
done_when:
  - "CHANGELOG.md holds `## [1.0.2] - 2026-11-05` above 1.0.1, in the 1.0.1 section's form (a lead note, Added, Changed, Fixed), every line traced to T-6010's or T-6011's landed code or BUG-017, BUG-139; no Unreleased heading left empty"
  - "`sh .github/scripts/release-notes.sh v1.0.2 <out>` writes the section; landed on main with CI green, every job; the director given the tag's line"
depends_on: [T-6011]
blocked_by: null
created: 2026-11-05
started: 2026-11-05
finished: 2026-11-05
labels:
  bonsai.ladder: []
  bonsai.allows: []
  bonsai.branch: t-6012-notes-1-0-2
  workflow.model: opus
  workflow.estimate_h: 0.5
---

# Bonsai 1.0.2's notes

The director, 5 Nov: "yes, start the 1.0.2 release prep". The release takes its notes from CHANGELOG's version
section (`.github/scripts/release-notes.sh`), so the section lands before the tag. Only the director tags `v1.0.2`
and approves the release run. After the release: the install lines at 1.0.2 and base 1.0.2 (its version, its ci
skill's pins to 1.0.2's checksums), a task of their own, as T-6008 was for 1.0.1.
