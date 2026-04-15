---
tags: [skill, pr, workflow]
description: How to create well-structured pull requests — branch naming, title conventions, body template, draft workflow.
---

# Skill: PR Creation

---

## When to Use

After implementation is complete and verification passes. The implementing agent (subagent) creates a **draft PR** as the final step of its execution. The tech lead marks it ready-for-review after the review loop passes.

---

## Branch Naming

The worktree isolation creates branches automatically. If you need to name one manually:

```
{type}/{short-kebab-description}
```

Types: `fix/`, `feat/`, `refactor/`, `docs/`, `chore/`

Examples: `fix/doubled-path-prefix`, `feat/selective-conflict-picker`

---

## PR Title

```
{type}: {concise description of what changed}
```

- Under 70 characters
- Imperative mood ("fix path doubling" not "fixed path doubling")
- Type prefix matches branch type: `fix:`, `feat:`, `refactor:`, `docs:`, `chore:`
- No issue number in the title — that goes in the body

Examples:
- `fix: strip workspace prefix from file tree display`
- `feat: add multi-select conflict picker to bonsai update`

---

## PR Body Template

Use this structure. All sections are required unless marked optional.

```markdown
## Summary

{1-3 sentences: what this PR does and why. Link the motivation — don't just describe the diff.}

Closes #{issue_number}

## Changes

{Bulleted list of what changed, grouped by area if multi-file. Be specific — file paths, function names.}

- `path/to/file.go` — {what changed and why}
- `path/to/other.go` — {what changed and why}

## Plan

{Link to the plan file if one exists, otherwise "No plan — Tier 1 patch."}

## Verification

{What was run to verify correctness. Include actual command output or confirmation.}

- [ ] `make build` — passes
- [ ] `go test ./...` — passes
- [ ] {any additional verification from the plan}

## Notes (optional)

{Edge cases, trade-offs, things the reviewer should pay attention to, follow-up items.}
```

---

## Creating the PR

Use `gh pr create` with a heredoc for the body:

```bash
gh pr create --draft --title "{type}: {description}" --body "$(cat <<'EOF'
## Summary

{summary text}

Closes #{issue_number}

## Changes

- `file.go` — description

## Plan

station/Playbook/Plans/Active/NN-name.md (or "No plan — Tier 1 patch.")

## Verification

- [x] `make build` — passes
- [x] `go test ./...` — passes
EOF
)"
```

### Flags

- **`--draft`** — always create as draft. The tech lead promotes it after review.
- **`--base main`** — target branch (use if not default)
- **`--label`** — add labels if applicable (e.g., `bug`, `enhancement`)

---

## After Creation

1. Report the PR URL and branch name back to the orchestrator
2. Do not merge, promote, or request review — that's the tech lead's job
3. If verification failed, report the failure instead of creating a PR

---

## Rules

- **Always draft** — subagents never create ready-for-review PRs
- **One PR per issue** — don't bundle unrelated work
- **Link the issue** — use `Closes #N` in the summary to auto-close on merge
- **Be specific in Changes** — file paths and function names, not vague descriptions
- **Include verification output** — "tests pass" means nothing without showing you ran them
- **No conversation history** — the PR should be self-contained context for any reviewer
