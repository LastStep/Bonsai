---
tags: [workflow, review]
description: Review a pull request — context, scope, correctness, security, performance, standards.
---

# Workflow: PR Review

---

## When to Use

When reviewing a pull request before merge. This is a systematic review — work through each pass in order. Do not skip passes.

---

## Steps

### 1. Load Context

- Read the PR description and linked ticket or plan
- Read the full diff — understand every changed file
- Note the PR's stated scope and goals
- Identify which components and modules are affected

### 2. Scope Check

- [ ] Changes match the linked ticket and ONLY the ticket
- [ ] No unrelated refactors, formatting changes, or "drive-by" fixes
- [ ] If unrelated changes exist, request they be split into a separate PR
- [ ] PR size is reviewable — flag PRs over ~400 lines of meaningful changes

### 3. Correctness Pass

- [ ] Logic matches the intended behavior from the ticket/plan
- [ ] Edge cases handled: empty inputs, null/undefined, boundary values, max limits
- [ ] Error paths are correct — errors are caught, reported, and don't leave inconsistent state
- [ ] No swallowed exceptions or empty catch blocks
- [ ] Async operations handle failures and timeouts
- [ ] State changes are atomic where required

### 4. Security Pass (OWASP)

- [ ] No hardcoded secrets, API keys, or credentials
- [ ] SQL queries are parameterized — no string concatenation
- [ ] User input validated at system boundaries (API handlers, form inputs)
- [ ] Auth checks on every new protected endpoint — verify both authentication and authorization
- [ ] Error responses don't leak internal details (file paths, stack traces, SQL)
- [ ] SSRF: user-supplied URLs validated against allowlist
- [ ] No `eval()`, `exec()`, or `new Function()` with dynamic input

### 5. Performance Pass

- [ ] No N+1 queries — list endpoints fetch related data in batch
- [ ] Large collections are paginated or streamed, never loaded entirely into memory
- [ ] No synchronous blocking in async code paths
- [ ] Appropriate caching for expensive or repeated operations
- [ ] Database queries have supporting indexes for filter/sort columns
- [ ] No unnecessary re-renders or re-computations in UI code

### 6. Maintainability Pass

- [ ] Functions under ~50 lines, files under ~300 lines (guidelines, not hard limits)
- [ ] No code duplication — shared logic extracted if used 3+ times
- [ ] Clear naming — variables, functions, and types describe what they represent
- [ ] No dead code, commented-out code, or unreachable branches
- [ ] No TODOs without a linked ticket reference
- [ ] Tests exist for new behavior and are meaningful (not just asserting `true`)

### 7. Standards Pass

- [ ] Follows project coding standards (naming, formatting, file organization)
- [ ] API changes follow API design conventions (if applicable)
- [ ] Documentation updated: README, API docs, CHANGELOG, inline docs for public APIs
- [ ] Types/interfaces updated to match implementation changes
- [ ] Migration files present if schema changed

### 8. Verdict

- **Approve** — all passes satisfied, merge when ready
- **Request Changes** — list specific issues with file paths and line references; explain what to fix and why
- **Needs Discussion** — flag design concerns that need broader input before proceeding
