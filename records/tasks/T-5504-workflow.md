---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task
id: T-5504
title: workflow, its own repository, written around the six roles (5.5.4)
status: done
lane:
done_when:
  - "LastStep/bonsai-workflow made from the template and written around the six roles: six role files with their skills: preloads and maxTurns, declared in pack.yaml with their budgets; session-start as the always-on file; the skills the note lists"
  - "check --pack and claude plugin validate locally; the role declarations equal to the section's \"The six roles, written once\""
  - "A scratch project linked to it alone with --yes alone, and with base, its block under 40 lines"
  - "P, the privacy verifier, on every commit from the root before its first push; the repository created private and pushed by the orchestrator; its CI green on both sides"
  - "Every later commit grepped and read before Rohan is told it can go public; Rohan's read"
depends_on: [T-5500, T-5501, T-5591]
blocked_by:
created: 2026-10-14
started: 2026-10-15
finished: 2026-10-16
labels:
  bonsai.ladder: []
  bonsai.allows: []
  bonsai.wants: []
---

# workflow, its own repository, written around the six roles (5.5.4)

Piece 5.5.4 of step 5.5 (approved 14 Oct): its row in `design/plan-5.md`'s "### Step 5.5" "The pieces and their order", and its note in "Notes per piece". Rohan's content, written with judgment by an Opus builder.

- **Owns:** the `workflow` repository (every file), worked in `~/bonsai-checks/bonsai-workflow`; nothing in Bonsai's repo.
- **Grants:** none in Bonsai's repo: it changes no file here, so `bonsai.allows` is empty and it has no branch here.
- **Proof:** the lines of `done_when`; P; its own CI; Rohan's read.
