---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task
id: T-5407
title: The switch, Bonsai guards its own repo and proves its work with bonsai ladder
status: done
lane:
done_when:
  - "bonsai.yaml validates under bonsai.workspace/1, a comment on every line, with the rungs, floor, protected and person_only lists plan-5's switch sets"
  - "The way back that CLAUDE.md gives (claude --continue with every hook off through --settings) tried in a scratch session where every call is refused: a free edit and a shell call run, and the earlier conversation is kept"
  - "A scratch clone of the branch, linked by the final code's stamped build on a scratch home: git diff shows bonsai.yaml unchanged, bonsai check has no finding, and bonsai ladder --task T-5407 is green on every rung"
  - "Bonsai's repo linked by the bonsai installed in WSL and committed on main: the lock, .claude/settings.json with every Bonsai line naming the two installed places, the block in CLAUDE.md, .bonsai/.gitignore and the two tables"
  - "bonsai check has no finding in the main checkout and in a fresh worktree"
  - "The floors in ratchets are the first green climb's counts, set in a commit under T-5407, and the installed bonsai's ladder --task T-5407 is green at that commit, each ratchet at its floor"
  - "CLAUDE.md holds the landing rule, agents first, the stricter-only rule, the task and name rules, the commands that stay a person's, the way back, the real home and the script rule for scratch links"
  - "CONTRIBUTING.md says a clone without bonsai installed is refused and how to work round it, and .gitignore holds .claude/settings.local.json"
  - "Nothing private in any file the task changes: no home path, user, machine, address or token"
depends_on: []
blocked_by:
created: 2026-10-13
started: 2026-10-13
finished: 2026-10-13
labels:
  bonsai.branch: "5.4.7"
  bonsai.ladder: []
  bonsai.allows: ["bonsai.yaml", ".bonsai/lock.json", ".claude/settings.json", "CLAUDE.md", ".gitignore"]
  bonsai.wants: ["bonsai.yaml", ".bonsai/lock.json", ".claude/settings.json"]
---

# The switch, Bonsai guards its own repo and proves its work with bonsai ladder

Piece 5.4.7 of `design/plan-5.md`, step 5.4: "The switch, written before it happens" and the piece's note. The first
task file in this repository, brought by the switch's own commit; from then on the orchestrator alone writes and moves
task files, in the main checkout.

- **Owns:** `bonsai.yaml`, `records/tasks/`, `CLAUDE.md`, `CONTRIBUTING.md` and `.gitignore` on the branch; at the link,
  on `main`, what `bonsai init` writes (the lock, `.claude/settings.json`, the block in `CLAUDE.md`, `.bonsai/.gitignore`
  and the two tables); then the floors in `bonsai.yaml`, raised from 0 to the first green climb's counts.
- **Grants:** `bonsai.allows` holds the protected paths those changes touch; `bonsai.wants` the person-only ones among
  them, copied into `bonsai.allows` under the stricter-only rule (the section the director approved names each change).
- **Proof:** the lines of `done_when`, checked by the 5.4 end verifier ("5.4 done", checks 16 to 23).
