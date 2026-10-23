---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task; how to fill: skill base:task
id: T-5730
title: "rc.2's words: CHANGELOG since rc.1, CLAUDE.md's stale lines, the README's step 6, the old demo and old sections out"
status: done
lane: null
done_when:
  - "CHANGELOG.md's 1.0.0 section names every change since rc.1 a user would notice (the stop gate's rule line, bonsai now and UTC records, the time check and time line, answer by letter, other agents' markers, BUG-014, Bonsai's words), each traced to its commit"
  - "CHANGELOG.md holds no old-product section: one line points to tag v0.4.3 for them; the notes script and CI's notes step green"
  - "assets/demos/ and .gitignore's assets/demos line are gone; nothing references them"
  - "CLAUDE.md: the stricter-only list matches the code (an added rung is looser, internal/workspace/checks.go), the named-session line does not name claude-here, line 3 no longer says being rebuilt"
  - "README step 6's update line matches rc.2's update output"
  - "the climb at the branch's HEAD green, CI on the branch green"
depends_on: [T-5725, T-5726, T-5727, T-5728, T-5729]
blocked_by: null
created: 2026-10-23
started: 2026-10-23
finished: 2026-10-23
labels:
  bonsai.branch: T-5730
  bonsai.ladder: []
  bonsai.allows: [".gitignore", "CLAUDE.md"]
  bonsai.wants: []
---

# rc.2's words and the cleanup

## Context

T-5710 items 2, 3, 7 and 12's README line; T-5799's removals on the director's word (23 Oct). It grants protected
paths, so it runs with no other task `running`, on top of the code batch.

## Out of scope

design/ (the director: "Keep them all"); past records.
