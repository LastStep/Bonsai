---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task
id: T-5734
title: "rc.3's fixes: the rc.2 trial's (LinkSink) list (each its own task, T-5735 on)"
status: todo
lane:
done_when:
  - "each item below landed as its own task with its proof, or put after 1.0 on the director's word; then rc.3, tested by the orchestrator, then 1.0"
depends_on: []
blocked_by:
created: 2026-10-24
started:
finished:
labels:
  bonsai.ladder: []
  bonsai.allows: []
  workflow.model: opus
---

# rc.3's fixes (T-5734)

The director, 24 Oct: "analyze what bugs need fixing and add them to rc3 plans. and then start working on that. once
rc3 lands and you have tested it, then we go to 1.0". From the trial on rc.2 in LinkSink (`records/trial-rc2.md`,
T-5733, and the build stage read the same day). B3 (`T-5731`) is already on `main` for rc.3.

1. **The workflow pack's steering** (BUG-103, BUG-104, BUG-105, and the build stage's two tasks `running` at once):
   a run report for any work on a task, planning, research and decisions included; the orchestrator keeps STATE as it
   goes and its notes in the project's memory, not a folder of its own; `running` when any agent starts on a task;
   what two `running` tasks cost (no active task: nothing granted, hours under "none"). `T-5735`, in
   `LastStep/bonsai-workflow`; then Bonsai's pin moved and `bonsai update`, as `T-5732`.
2. **`bonsai check` on a project with no commit, or with `bonsai.yaml` uncommitted** (BUG-106), and **the words
   after a pack is added** (BUG-107's words: the running Claude Code session needs a reload or a fresh start before
   the pack's skills load). `T-5736`.
3. **The guard's role-writes-nothing words** (BUG-108's words): when the guard could not follow a `cd` and so read a
   relative write as inside the project, its refusal says so and names a next step that fits (write by an absolute
   path outside the project), not "write scratch files outside the project". The reading itself after 1.0. `T-5737`.
4. **Before the tag:** Go and govulncheck checked again (go1.27.2 stays unless govulncheck loads a newer one); CI on
   `main` green; the records (`records/trial-rc2.md` with the build stage, STATE).
5. **rc.3 itself:** the director's tag and approval, the WSL install; then **the orchestrator's test** (`T-5738`):
   the README's first project on the installed rc.3 by an agent in scratch, each fix seen working, a check of LinkSink
   read-only, check 10's Windows half. Then 1.0 (batch 5, 5.7.6, the floors, the 5.7 end verifier).

After 1.0, by their rows in `records/bugs.md`: BUG-107's cleanup, BUG-108's reading, BUG-109 to BUG-111, BUG-095.
