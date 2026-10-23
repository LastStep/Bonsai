---
format: bonsai.run/1   # fields: bonsai check --schema bonsai.run; how to fill: skill base:run
id: R-2026-10-23-T-5730
task: T-5730
role: orchestrator
model: opus
started: 2026-10-23 09:00
finished: 2026-10-23 09:12
outcome: verify
commits: []
labels: {}
---

# Run: rc.2's words and the cleanup (T-5730)

- Who: the orchestrator (opus), itself (words only), on T-5726's tip `dae88ac0`, the one task `running`.

## Log

- 09:00 `assets/demos/` removed (0.4.3's demo; nothing reads it, T-5799's audit and the orchestrator's `git grep`)
  and `.gitignore`'s `!assets/demos/*.png`; CHANGELOG's old-product sections replaced by a pointer to tag `v0.4.3`
  (T-5710 item 2: the notes no longer end on the old line; `release-notes.sh v1.0.0-rc.2` prints the 1.0.0 section,
  ending on "A 0.4.3 workspace").
- CHANGELOG's 1.0.0 section takes what came since rc.1: the stop gate's rule line (T-5708, item 3), one clock in UTC
  (T-5795), answer by letter (T-5728), other agents' markers (T-5727), `docs/features.md`, the subfolder pack's tags
  (T-5725), a climb's printed lines (T-5729), the old demo out; the schema count 28 to 29 (set 11).
- CLAUDE.md (item 7): line 3 no longer "being rebuilt"; the named session through the T-5708 launcher, `claude-here`
  scratch only; the stricter-only list: a rung added reads looser (`internal/workspace/checks.go`).
- README step 6's update line matches T-5729's lead (item 12); features.md's BUG-014 note out.

## Ladder

At the branch's tip, by the orchestrator: below in the umbrella report `R-2026-10-23-T-5710`.

## What did not work

Nothing went red. The 1.0.0 section's "Changed" now mixes changes since rc.1 with changes since 0.4.3: one section
for 1.0 by the plan's rule (5.7.4).

## Processes left running

None (`ps -eo pid,etime,cmd`).

## Left for next time

None in this task.
