---
tags: [workflow, logging]
description: End-of-session log — what was done, decisions made, open items.
---

# Workflow: Session Logging

---

## When to Use

At the end of every session, or when switching to a different task.

---

## Log Format

```markdown
## Session Log — YYYY-MM-DD

### Completed
- What was accomplished this session

### Decisions
- Any decisions made and their rationale

### Open Items
- What remains to be done
- Any blockers or questions for next session

### Files Modified
- List of files created or changed
```

## Rules

- Keep it concise — bullet points, not paragraphs
- Include file paths for anything modified
- Flag anything that the next session needs to know
- Status/memory rows for this session follow `Playbook/Standards/NoteStandards.md` (3 lines max, link out). The full session log lives here — link to it from Status/memory rather than restating.
