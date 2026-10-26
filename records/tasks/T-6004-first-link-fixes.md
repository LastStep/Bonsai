---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task; how to fill: skill base:task
id: T-6004
title: "Fix what the first existing project's link found: BUG-121 to BUG-124, BUG-126, BUG-128, BUG-130 (the design questions, the guard and link, the workflow pack's reporting)"
status: todo
lane: null
done_when:
  - "A plan records/plans/P-T-6004.md, reviewed by a fresh Opus agent and approved by the director, says for each of BUG-121, BUG-122, BUG-123, BUG-124, BUG-126, BUG-128 and BUG-130 one of: fixed in Bonsai, fixed in a pack, or left with the director's word (a known limit, or after a later piece)"
  - "Each bug the plan fixes has a test (or, for words, the changed skill text) shown failing on the old code first, where a test can show it, and is moved to records/bugs.md's Fixed table with its commit"
  - "Every Bonsai change landed by the landing rule (CLAUDE.md): a green local ladder at the branch HEAD by the installed bonsai, CI green on the branch, both windows jobs among them; guard, ladder and rung-0 changes passed by a fresh Opus verifier"
  - "The workflow pack's changes (BUG-128), if the plan fixes them, are committed in its repository's local copy and named in the run report; their release and tags stay the director's"
depends_on: [T-6001]
blocked_by: null
created: 2026-10-26
started: null
finished: null
labels:
  bonsai.ladder: []
  bonsai.allows: []
  bonsai.wants: []
  workflow.model: opus
---

# Fix what the first existing project's link found: the design questions, the guard and link, the pack's reporting

**Split, 26 Oct** (the director: "3, and then update the state so a fresh session can start working parallely from
group b"): BUG-119, BUG-120, BUG-125, BUG-129 and BUG-127 went to `T-6005`, built in the session that landed T-6002
and T-6003. This task keeps the seven below. Its order, as the split proposed it: first the options for the design
questions (group B: BUG-124, BUG-126, BUG-121, BUG-123), one question at a time to the director; then group C (the
guard and the link: BUG-130, BUG-122) with a fresh Opus verifier; then group D (BUG-128, the workflow pack).

## Context

Bonsai 1.0 was linked to its first existing project on 26 Oct (T-6001). The trial record
`records/trial-zengarden-ai-api.md` says what held and what did not. Its twelve issues are filed in `records/bugs.md`
as BUG-119 to BUG-130, each with where in the code it is (BUG-029 and BUG-111 were seen again; they are not this
task's). The director's word, 26 Oct: "file all these bugs, and then a fresh session needs to pick them up and fix".

## Scope

- **Bonsai's code:**
  - `never_edit`'s rules (BUG-121);
  - a second pack's link (BUG-122);
  - `time-not-now` and future times in `check` (BUG-123);
  - rung 0 on the base branch (BUG-124);
  - the ladder's result per climb (BUG-126);
  - `delete-unnamed` on moves (BUG-130).
- **The workflow pack's skills** (BUG-128), in `~/bonsai-checks/bonsai-workflow`. T-6005 changes its orchestrate
  and verification-ladder skills there (BUG-127): rebase onto that commit.

The plan decides which are fixed now. Several are design questions, not plain bugs:
- whether shell reads of walled files can be judged at all;
- what rung 0 should claim on the base branch;
- whether a result keeps one entry per climb.

Each of those goes to the director as options (`workflow:options`), one question at a time.

## Out of scope

- The guard judging shell writes in general (BUG-048's limit): name it where it bears, do not take it on.
- ZenGarden-AI-API itself; ZEN's link.
- Any release or tag (the director's).

## Notes

- **Runs beside T-6005** (the director, 26 Oct). Every task here has empty `bonsai.allows`, so two `running` tasks
  grant nothing either way; name the task in every climb (`--task`). T-6002 and T-6003 (LFS: `internal/vergit`,
  `internal/rung0`, `internal/ladder/text.go`) land first: build BUG-124 and BUG-126 on a `main` that holds them.
  T-6005 touches `internal/engine/checkdocs.go`, `internal/message/pointer.go` and two base skills: keep out of them.
- The options come first and need no code: the design questions can start at once, before T-6002 lands.
- The trial's auditors' reports, with file and line evidence for each issue: `M-2026-10-26-851dff1c` and
  `M-2026-10-26-a781cee9`.
- The real project's log, for reproductions (read-only):
  `~/ZenGarden/ZenGarden-AI-API/.bonsai/local/log/s-c8c280da-….ndjson`.
