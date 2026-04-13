---
tags: [workflow, reporting]
description: Completion report format — what was done, tested, issues encountered.
---

# Workflow: Reporting

---

## When to Use

After completing a plan or significant task. Submit this report for review.

---

## Steps

1. Read the report template at `Reports/report-template.md`
2. Fill in all sections — be specific
3. Save the report to `Reports/Pending/` with the naming convention: `YYYY-MM-DD-plan-NN-agent.md`

> [!note]
> The paths above are relative to the project docs location. Check your workspace CLAUDE.md → External References for the exact paths.

---

## Rules

- Be specific — "added tests" is not enough, list which tests
- If you deviated from the plan, explain why
- Include coverage numbers
- Include all files created and modified
- Note anything the next session or another agent needs to know
- If you discovered bugs, debt, or improvement opportunities outside the plan's scope, add them to `Playbook/Backlog.md` rather than noting them only in the report
