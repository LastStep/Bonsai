---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task
id: T-5736
title: "bonsai check warns on a project with no commit or bonsai.yaml uncommitted; init and update say a running session reloads"
status: done
lane:
done_when:
  - "bonsai check warns (a new warning, next step a person's: commit) on a linked project whose repository has no commit, and on one whose bonsai.yaml differs from HEAD's or is not in it, by test; the warning in check's list and the reference regenerated"
  - "init and update, when a pack's plugin is added or its marketplace renamed, say that a Claude Code session already open in the checkout loads the pack's skills only after a reload or a fresh start, by test"
  - "the climb at the branch's HEAD green, CI on the branch green"
depends_on: []
blocked_by:
created: 2026-10-24
started: 2026-10-24
finished: 2026-10-24
labels:
  bonsai.branch: T-5736
  bonsai.ladder: []
  bonsai.allows: []
  workflow.model: opus
---

# check on no commit; the words after a pack is added (T-5736)

T-5734 item 2: BUG-106 and BUG-107's words (`records/bugs.md`; `records/trial-rc2.md`). In LinkSink the repository
held no commit for 2 h 25 min and `bonsai check` said nothing, though a climb and the stop gate refuse
(`config-uncommitted`); after the person added the workflow pack, the open session showed every pack skill as
"Unknown skill" until `/plugin` was run twice.
