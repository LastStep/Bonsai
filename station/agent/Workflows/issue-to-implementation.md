---
tags: [workflow, orchestration]
description: End-to-end autonomous workflow — issue intake to shipped code via research, planning, agent dispatch, review loops, and structured logging.
---

# Workflow: Issue to Implementation

> The Tech Lead's primary orchestration workflow. Issue → shipped code.

---

## Prerequisites

Load before starting:

- `agent/Skills/issue-classification.md` — issue types, importance levels, Bonsai domain labels
- `agent/Skills/planning-template.md` — plan format and tier rules
- `agent/Skills/review-checklist.md` — code review passes (referenced in review loop)

---

## Overview

```
Pre-Flight → Intake → Analysis → Research Loop → Clarify → Plan → Self-Review → Triage → Execute → Review Loop → Logging → Final Audit → Merge & Close
```

---

## Autonomy Modes

This workflow supports two modes:

- **Supervised** (default) — pause at Clarify and Triage for user input
- **Autonomous** — skip Clarify if research resolves all questions; self-dispatch if triage criteria are met

The user sets the mode at the start. When in doubt, default to supervised.

---

## Phase 0: Pre-Flight

1. Run `git status`. If the working tree has uncommitted changes, **stop and warn the user**. Suggest committing or stashing before proceeding.
2. Check `Playbook/Status.md` — if there's in-progress work that could conflict, flag it before starting.

---

## Phase 1: Intake

**Trigger:** User assigns a specific issue, asks to scan for work, or points to a Backlog.md item.

### From GitHub Issues

- **Repo:** `LastStep/Bonsai`
- Use `gh issue list -R LastStep/Bonsai` / `gh issue view N -R LastStep/Bonsai` to read the issue
- Extract: title, body, labels, comments, linked issues

### From Backlog.md

- Read the item, extract context and priority
- Check if there's a linked GitHub issue

### Classify

Use `agent/Skills/issue-classification.md`:

- **Type:** bug, feature, change, debt, research
- **Domains:** which parts of the codebase are affected
- **Importance:** critical, high, medium, low

---

## Phase 2: Analysis

Understand what the issue touches before planning anything.

1. **Trace** — use Explore agents or Grep/Read to find affected files, functions, modules. Key entry points:
   - CLI commands: `cmd/*.go`
   - Catalog loading: `internal/catalog/catalog.go`
   - Config/lock: `internal/config/config.go`, `internal/config/lockfile.go`
   - Generation: `internal/generate/generate.go`
   - TUI: `internal/tui/styles.go`, `internal/tui/prompts.go`
   - Catalog items: `catalog/{agents,skills,workflows,protocols,sensors,routines,scaffolding}/`
   - Code index: `station/index.md`
2. **Blast radius** — what files change, what tests are affected (`internal/generate/generate_test.go`, `internal/config/*_test.go`), what depends on the changed code
3. **Architecture** — check `station/INDEX.md` for the architecture overview; check research docs (`RESEARCH.md`, `RESEARCH-concepts.md`, `RESEARCH-catalog-expansion.md`) if touching design decisions
4. **Overlap** — check `Playbook/Status.md` for in-progress work that conflicts
5. **Related items** — check `Playbook/Backlog.md` for items that should be bundled with this work
6. **Prior decisions** — check `Logs/KeyDecisionLog.md` for constraints on the approach

---

## Phase 3: Research Loop

Deepen understanding until confident the plan will be correct.

### Each pass:

1. **Web research** — best practices, library documentation, known issues, similar implementations
2. **Codebase patterns** — how does the project handle similar cases? What conventions exist?
3. **Dependency check** — will this require new dependencies? Are there version constraints?
4. **Assumption verification** — test each assumption against actual code

### Confidence gate:

After each pass, rate confidence 1–5:

- **>= 4** — proceed to Clarify
- **< 4** — identify the specific gaps driving low confidence, research those gaps, re-rate
- **Max 3 passes** — if still < 4 after 3 passes, proceed but flag low-confidence areas explicitly in the plan

> The goal is not exhaustive research. The goal is enough confidence that the plan won't need revision mid-execution.

---

## Phase 4: Clarify

Ask the user about things that analysis and research couldn't resolve.

- Open design decisions
- Ambiguous requirements
- Priority conflicts
- Scope boundaries

> Don't ask questions you can answer from the codebase or research.
> One question at a time — each answer may resolve the next question.
> **Autonomous mode:** Skip entirely if confidence >= 4 and no unresolvable ambiguities exist.

---

## Phase 5: Plan

Write `Playbook/Plans/Active/NN-kebab-case-name.md` using the planning-template skill.

- **Tier 1 (Patch):** bug fix, config tweak, simple addition — single-domain, well-scoped
- **Tier 2 (Feature):** new capability, multi-step, multi-domain, architectural

**Multi-domain:** Write separate step sections per agent. Mark parallel vs sequential.

**Include:**

- Verification steps — concrete commands or checks that prove the work is correct
- Security references — every plan must reference SecurityStandards.md
- Research findings — anything that constrains the implementation

---

## Phase 6: Self-Review

- [ ] Steps are specific — no design decisions left to the agent
- [ ] File paths, function names, data shapes are explicit
- [ ] Security standards referenced
- [ ] Verification is concrete and testable
- [ ] No scope creep beyond the issue
- [ ] Edge cases addressed
- [ ] Multi-domain dependencies sequenced correctly
- [ ] Research findings incorporated — no assumptions contradicted by evidence

Fix every issue before proceeding. Do not carry known problems into dispatch.

---

## Phase 7: Triage

### Self-dispatch when ALL true:

- Limited scope, well-defined changes
- No project structure or architecture changes
- Self-review passed clean
- Not critical importance
- Single domain, or multi-domain with clear sequencing
- Research confidence >= 4

### Escalate to user when ANY true:

- Touches project structure or architecture
- Cross-domain with design interdependencies
- Critical importance
- Unconfirmed assumptions in the plan
- Open questions from self-review
- Low-confidence areas flagged in Phase 3

**When escalating:** Present the plan and point to the specific decisions that need attention. Wait for approval before dispatching.

---

## Phase 8: Execute

Dispatch implementation agent(s) in isolated worktrees.

> **Bonsai is currently a single-agent project** (tech-lead). Sub-agent dispatch means spinning up general-purpose agents in worktrees to do the implementation work while you orchestrate.

### Dispatch syntax

```
Agent(subagent_type: "general-purpose", isolation: "worktree", prompt: "...")
```

For independent parallel work:

```
Agent(subagent_type: "general-purpose", isolation: "worktree", run_in_background: true, prompt: "task A...")
Agent(subagent_type: "general-purpose", isolation: "worktree", run_in_background: true, prompt: "task B...")
```

### Agent prompt structure

Include in this order:

1. **Bootstrap** — "Read `CLAUDE.md` at the project root first, then `station/CLAUDE.md`."
2. **Context** — the problem being solved (from the issue)
3. **Plan steps** — the specific steps to execute, copied verbatim from the plan
4. **Plan location** — `station/Playbook/Plans/Active/NN-name.md`
5. **Verification** — "Run `make build` and `go test ./...` before reporting completion."
6. **Constraints:**
   ```
   - Don't modify files outside the scope of the plan
   - Don't make design decisions — if the plan is ambiguous, stop and report
   - Don't add features, refactor code, or make improvements beyond what the plan specifies
   - If something is unclear, stop and report — don't guess
   - Run verification steps before reporting completion
   ```

Do NOT include: conversation history, unrelated context, or vague instructions.

After dispatch, wait for the agent notification — don't poll.

---

## Phase 9: Review Loop

When the execution agent finishes, enter the review cycle.

### Step 1 — Self-review the output

- Read the agent's summary
- Diff the worktree against the plan — every step followed, nothing improvised
- Check for scope creep (changes not in the plan)

### Step 2 — Dispatch review agent(s)

For substantial changes, dispatch an independent review agent:

```
Agent(subagent_type: "general-purpose", prompt:
  "Review the changes on branch {branch}.
   Read the plan at {plan-path}.
   Use the review checklist at agent/Skills/review-checklist.md.
   Check: correctness, security, test coverage, standards compliance.
   Report pass/fail with specific issues found.")
```

For security-sensitive changes, also dispatch a security review:

```
Agent(subagent_type: "general-purpose", prompt:
  "Security review of changes on branch {branch}.
   Check against station/Playbook/Standards/SecurityStandards.md.
   Focus on: input validation, secrets in templates, error handling, embed.FS safety, dependency safety.
   Report pass/fail with specific findings.")
```

### Step 3 — Evaluate review results

| Result | Action |
|--------|--------|
| All checks pass | Proceed to Logging |
| Minor issues found | Dispatch fix agent on the same worktree, then re-review from Step 2 |
| Major issues found | Escalate to user with specific problems listed |

### Iteration limits

- **Max 3 execute-review cycles** before mandatory escalation to user
- Track the iteration count explicitly
- If hitting the limit, the plan likely needs revision — not just the code

---

## Phase 10: Logging

Update all tracking systems. Do all of the following that apply:

### 1. Execution log

Append to `Logs/RoutineLog.md`:

```markdown
### YYYY-MM-DD — Issue #N: Title
- **Plan:** Plans/Active/NN-name.md
- **Iterations:** N execute-review cycles
- **Issues found:** (list any issues caught during review)
- **Result:** completed | partial | escalated
```

### 2. GitHub Issue

If the issue came from GitHub, comment with:

- What was implemented
- Key decisions made during execution
- Test results
- Any caveats or follow-up items

### 3. Status

Update `Playbook/Status.md`:

- Move to Recently Done with today's date, or
- Update In Progress if partially complete

### 4. Backlog

Update `Playbook/Backlog.md`:

- Remove the item if it was sourced from there
- Add any new items discovered during implementation

### 5. Completion report

If reports scaffolding is installed, submit a report to `Reports/Pending/` using the report template.

---

## Phase 11: Final Audit

Before merging, verify holistically:

1. **Build** — `make build` — must succeed with no errors
2. **Tests** — `go test ./...` — full suite, not just new tests
3. **CLI smoke test** — if CLI behavior changed, test interactively in a temp dir:
   ```bash
   mkdir /tmp/bonsai-test && cd /tmp/bonsai-test && git init
   /path/to/bonsai init   # walk through the flow
   /path/to/bonsai add    # verify new items appear
   /path/to/bonsai list   # verify installed items
   ```
4. **Scope check** — `git diff` should match the plan; flag anything extra
5. **Security scan** — no secrets in templates, no hardcoded paths, no `.env` files committed
6. **Stale references** — no broken imports, no references to removed or renamed catalog items, no dangling template vars
7. **Catalog consistency** — if catalog items changed, verify `meta.yaml` fields (`name`, `description`, `agents`) are correct and `bonsai catalog` renders properly
8. **Documentation** — if behavior changed, are `station/INDEX.md`, `CLAUDE.md`, or `station/CLAUDE.md` updated?

If any check fails: fix it, re-verify the fix, and document what was caught in the execution log.

---

## Phase 12: Merge & Close

1. **Merge** worktree branch(es) into the working branch
2. **Post-merge verification** — run tests again after merge to catch integration issues
3. **Close GitHub Issue** — if applicable, close with `gh issue close N -R LastStep/Bonsai -c "..."`:
   - Confirmation that work is complete
   - Link to the plan
   - Follow-up items filed as new issues (if any)
4. **Update Status.md** — ensure Recently Done entry exists
5. **Update memory** — if significant architectural decisions were made, update `agent/Core/memory.md`
6. **Notify user** — concise summary: what was done, how many iterations, any follow-ups

---

## Quick Reference

| Phase | Key Tools | Exit Criteria |
|-------|-----------|---------------|
| Pre-Flight | `git status` | Clean working tree |
| Intake | `gh`, Backlog.md | Issue classified |
| Analysis | Explore agents, Grep, Read | Blast radius mapped |
| Research | WebSearch, WebFetch, Read | Confidence >= 4 (or 3 passes exhausted) |
| Clarify | Direct conversation | No unresolved ambiguities |
| Plan | Write tool | Plan file written |
| Self-Review | Internal checklist | All checks pass |
| Triage | Decision tree | Dispatch or escalate decided |
| Execute | Agent tool (worktree) | Agent(s) complete |
| Review | Review agents, diff | All reviews pass (max 3 cycles) |
| Logging | Status.md, GitHub, logs | All systems updated |
| Final Audit | Tests, lint, build | All green |
| Merge & Close | Git, GitHub | Issue closed, branch merged |

---

## Failure Modes

| Situation | Action |
|-----------|--------|
| Research confidence stays < 4 after 3 passes | Proceed with explicit caveats in plan; flag gaps to user at Triage |
| Execute-review loop hits 3 iterations | Stop and escalate — the plan likely needs revision, not just the code |
| Post-merge tests fail | Revert merge, fix in worktree, re-audit, merge again |
| Agent produces output outside plan scope | Reject the output. Re-dispatch with tighter constraints |
| Conflicting in-progress work discovered | Stop. Coordinate with user before proceeding |
| Agent fails or times out | Check agent summary for partial work. Decide: resume on same worktree or start fresh |
