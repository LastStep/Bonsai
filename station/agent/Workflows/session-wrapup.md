---
tags: [workflow, session]
description: End-of-session verification, review, cleanup, and summary — triggered by session wrap-up phrases.
---

# Workflow: Session Wrap-Up

> [!warning]
> When this workflow is triggered, follow every step. Do not skip steps or compress them. The point is to catch mistakes before the session ends.

---

## Step 1 — Verify Work Done

- Review everything accomplished in this session
- For each task or change: confirm the intended outcome actually happened
- If code was written: check it builds (`make build`) and tests pass (`go test ./...`)
- If files were created/edited: confirm they exist and contain the right content
- If config was changed: verify the config is valid

## Step 2 — Check for Mistakes

- Re-read every file you modified — look for:
  - Wrong paths or stale references
  - Broken formatting (unclosed markdown tables, bad links)
  - Placeholder text left in (e.g., `(your backend stack)`, `TODO`)
  - Inconsistencies between related files (e.g., CLAUDE.md references a file that doesn't exist)
  - Logic errors, typos, or copy-paste artifacts
- Check git status — are there files that should have been modified but weren't?
- Check for anything you said you'd do but didn't

## Step 3 — Fix Issues

- Fix every issue found in Step 2
- Do not ask for permission — just fix them (these are corrections, not new work)
- If a fix is non-trivial or could break something, note it for the summary

## Step 4 — Verify Fixes

- If fixes were made in Step 3, re-verify those specific changes
- Confirm the fix didn't introduce new issues
- If code was touched: rebuild and retest

## Step 5 — Final Cleanup & Updates

1. **Memory** — update `agent/Core/memory.md`:
   - Set current task to `(none)` or the carry-forward task
   - Move completed work to the Completed list
   - Add any new notes or references discovered this session
   - Clean stale entries

2. **Status** — update `Playbook/Status.md`:
   - Move completed items to Recently Done with today's date
   - Update In Progress items
   - Add any new Pending items

3. **Backlog** — update `Playbook/Backlog.md`:
   - Add any bugs, improvements, or ideas discovered during the session
   - Don't add items already tracked elsewhere

4. **Git status** — run `git status`
   - If there are uncommitted changes, tell the user and ask if they want to commit

## Step 6 — Session Summary

Provide a concise summary:

```
## Session Summary

**Done:**
- (bullet list of what was accomplished)

**Fixed during wrap-up:**
- (any issues caught and fixed in Steps 2-4, or "None")

**Carry-forward:**
- (anything that needs to continue next session, or "None")

**State:**
- Uncommitted changes: yes/no
- Memory updated: yes/no
- Status updated: yes/no
```
