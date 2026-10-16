---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task
id: T-5505
title: The packs in real Claude Code sessions, on both sides (5.5.5)
status: blocked
lane:
done_when:
  - "A Sonnet agent runs the packs in real sessions as note 5.5.5 lists: a role's preloaded skills, a role as a subagent, the always-on file, the operating skill's four scenarios, --agent workflow:builder --bg, an interactive session through a terminal multiplexer; WSL first, Windows at Rohan's sitting"
  - "Each session's transcript (stream JSON) and answers in the run report, with Claude Code's version and the user settings hashes before and after"
  - "Started after T-5503 has landed, VM and VG have passed and workflow's CI is green; a finding goes back to the piece that owns the file, as a fix round"
depends_on: [T-5503, T-5504]
blocked_by: "Waits for the director: the Windows step for S1 at the fixed commit (~/bonsai-checks/T-5505/rerun/windows-sitting.md)."
created: 2026-10-14
started: 2026-10-16
finished:
labels:
  bonsai.branch: null
  bonsai.ladder: []
  bonsai.allows: []
  bonsai.wants: []
---

# The packs in real Claude Code sessions, on both sides (5.5.5)

Piece 5.5.5 of step 5.5 (approved 14 Oct): its row in `design/plan-5.md`'s "### Step 5.5" "The pieces and their order", and its note in "Notes per piece". Hand checks an agent can run (Rohan's 8 Oct word).

- **Owns:** no repository file: scripts in `~/bonsai-checks/scripts/`.
- **Grants:** none: it changes no file in the repo.
- **Proof:** the run report's transcripts and answers, as `done_when` says.
