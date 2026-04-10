---
tags: [workflow, execution]
description: Execute an assigned plan step by step — implement, test, report.
---

# Workflow: Plan Execution

---

## When to Use

When you have been assigned a plan to implement.

---

## Steps

### 1. Read the Plan

- Read the full plan from start to finish
- Identify all files that need to be created or modified
- Note any dependencies or sequencing requirements

### 2. Execute Step by Step

- Work through the plan steps in order
- After each step, verify it works before moving on
- If a step is ambiguous, ask the user — do not guess

### 3. Test

- Run the full test suite
- Ensure no regressions
- Add new tests for new code

### 4. Self-Check

- [ ] Every plan step completed
- [ ] All tests pass
- [ ] No files modified outside the plan's scope
- [ ] No hardcoded secrets or credentials
- [ ] Code follows project coding standards

### 5. Report

- Write a completion report using `agent/Workflows/reporting.md`
- Include what was done, what was tested, any issues encountered
