---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task
id: T-5732
title: "The workflow split: the brief and playtest kinds and the interview leave workflow for the studio and game packs"
status: done
lane: light
done_when:
  - "LastStep/bonsai-workflow: the brief kind and the daily brief's template, the playtest kind and skills/playtest, and skills/interview gone; the producer's work kept as a general skill under a name that is not brief; every reference to workflow:brief, workflow:playtest and workflow:interview repointed; pack.yaml's documents comment says six kinds; ci/check.sh check and validate clean; CI green; main pushed by the orchestrator"
  - "Bonsai's bonsai.yaml pins workflow at that commit; bonsai update in the main checkout; bonsai check clean; the records commit lands by CI"
depends_on: []
blocked_by:
created: 2026-10-24
started: 2026-10-24
finished: 2026-10-24
labels:
  bonsai.ladder: []
  bonsai.allows: ["bonsai.yaml", ".bonsai/lock.json"]
  bonsai.wants: ["bonsai.yaml", ".bonsai/lock.json"]
  workflow.model: sonnet
---

# The workflow split (T-5732)

The director, 24 Oct: do the workflow split exactly as the studio's plan writes it
(`~/Servers/Trinetra-Game-Studio/studio/plans/P-T-0082.md`, "For Bonsai's sessions", first bullet; the director's word
there, "go with the split"). The studio's T-0082 shipped the studio pack (declares `brief`) and the game pack
(declares `playtest`, holds `interview`); Bonsai refuses two packs declaring one kind (`packs-overlap`), so no project
can take them until workflow gives those up.

Out: the `brief` kind and the daily brief's template (now `studio:brief`); the `playtest` kind and `skills/playtest`
(now `game:playtest`); `skills/interview` (now `game:interview`). Stays, reworded general: the rest of `skills/brief`
(the producer's work: STATE kept true, lanes applied, failures and notes triaged into tasks, the milestone date
watched) under a new name, with no daily brief and no playtest in it. The plan's last line ("the rest of workflow made
general, as answer 4 says") is after 1.0 by that answer's own words (`VISION-2026-10-23-bonsai-and-studio-aligned.md`,
answer 4), so not here.

Two halves, as `T-5796`: the pack's change in `~/bonsai-checks/bonsai-workflow` (a builder, worktree
`~/bonsai-checks/bonsai-workflow-T-5732`, branch T-5732; the orchestrator pushes); then Bonsai's pin and `bonsai
update` in the main checkout, which needs `bonsai.yaml` and `.bonsai/lock.json` granted and so waits for the
director's word: "A, go ahead" (24 Oct).
