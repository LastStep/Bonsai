---
tags: [core, memory]
description: Tech Lead Agent working memory — flags, work state, notes.
---

# Working Memory

## Flags

<!-- Active flags go here. Format: - [FLAG] description -->

(none)

## Work State

**Current task:** Idle. Last ship (2026-05-04): Plan 34 — custom-ability discovery bug bundle. 4 fixes shipped in single PR. [plan](../../Playbook/Plans/Archive/34-custom-ability-discovery-bug-bundle.md) · [PR #92](https://github.com/LastStep/Bonsai/pull/92).

**Brevity rule:** this section follows [NoteStandards](../../Playbook/Standards/NoteStandards.md) — link out, don't re-state. Prior shipped work in [Status.md](../../Playbook/Status.md) Recently Done.

## Notes

<!-- Session-to-session durable gotchas. Drop event narratives; keep what will bite again. -->

- **Brevity rule for trackers.** All writes into `Playbook/Status.md`, `Playbook/Backlog.md`, `agent/Core/memory.md` Work State, and any project tracker follow [Standards/NoteStandards.md](../../Playbook/Standards/NoteStandards.md): 3 lines max per entry, link out for detail. Phase-by-phase walkthroughs go in the plan; commit walkthroughs in the PR; process narrative in `Logs/`. Rule established 2026-04-25 after Plan 32 row hit ~3KB single-row.
- **Worktrees inherit only committed HEAD.** Uncommitted plans/docs in main tree are invisible to dispatched agents. Commit station/ artifacts before dispatch. Agent worktrees base off `origin/main`, not local main — push local-only commits first.
- **Worktree creation cwd matters.** `git worktree add ../<name> <branch>` resolves `../` against the Bash tool's cwd (often `station/`, not repo root). Use absolute paths when creating worktrees.
- **Worktree-held-branch post-merge cleanup.** `gh pr merge --squash --delete-branch` silently skips local+remote branch delete when its worktree is checked out. After every squash-merge from a non-main worktree, manually: `git worktree remove -f -f <path>` + `git branch -D <br>` + `git push origin --delete <br>`. Pattern hits 10×+/month.
- **MDX autolink gotcha.** `<https://example.com>` is valid GFM but MDX parses `<` as JSX — Astro build fails "Unexpected character `/`". In `.mdx` files, always use `[label](url)`.
- **Local `go build`/`go test` miss `golangci-lint unused`.** Local golangci-lint v2.x; repo config is v1 → `golangci-lint run` errors locally with "unsupported version". Trust CI for dead-code on deletion/refactor PRs, OR install v1 locally (`go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.8`).
- **`statusLine.command` runs in different `$PWD` than hook commands.** Walk-up-to-`.bonsai.yaml` wrappers fail (walk terminates at `/`). Use absolute path at install time. StatusLine config isn't hot-reloaded — requires `/clear` or restart.
- **Subdirectory launch determines which `.claude/settings.json` is live.** When launched from `station/`, effective project settings is `station/.claude/settings.json`; repo-root is ignored. Check `pwd` at session start.
- **golangci-lint binary Go-version coupling.** `version: X.Y` on `golangci/golangci-lint-action` resolves to a specific prebuilt binary; binary's own Go build must be ≥ repo target Go. Bumping Go major requires bumping action pin to release whose notes say "built with Go ≥ target". Track: https://github.com/golangci/golangci-lint/releases.
- **GoReleaser Homebrew step is silently PAT-dependent.** Release workflow succeeds up to GH Release publish, then fails at brew step if `HOMEBREW_TAP_TOKEN` is expired — GH Release already published. Symptom: `GET https://api.github.com/repos/LastStep/homebrew-tap: 401 Bad credentials`. Recovery: rotate PAT, set secret on main repo (NOT tap), push formula manually via `gh api -X PUT /repos/LastStep/homebrew-tap/contents/Formula/bonsai.rb` with SHAs from `checksums.txt`. Safer than `goreleaser release --clean` on existing tag (which wipes+recreates GH Release, breaks download caches). Add `workflow_dispatch:` to `release.yml` BEFORE first release for clean retries.
- **Session-start catalog generator diffs are load-bearing.** When session starts with `website/public/catalog.json` + mdx modified, do NOT dismiss as stale regen. A prior audit agent may have run `website/scripts/generate-catalog.mjs` and left the fix. Check what the diff fixes before reverting.
- **`git add <file>` is all-or-nothing — pulls unrelated WIP hunks from parallel sessions.** When parallel-session WIP is in the same file, `git add <path>` stages every modified hunk. How to apply: (a) `git stash push -u -- <paths>` first, OR (b) `git add -p <file>` for hunk-level stage, OR (c) `git diff --staged` before every commit.
- **Parallel sessions can spontaneously switch your branch mid-session.** A parallel session may create a branch from your tip + check you out onto it. Check `git branch --show-current` immediately before any commit when parallel session is active; if on unknown branch, `git restore --staged .` + `git switch main`.
- **Parallel sessions can re-stage between `git status` recheck and `git add`.** Explicit-path `git add` does NOT prevent this — once in index, `git commit` rides along. How to apply: chain `git add <paths> && git diff --cached --stat && git commit ...` in a single bash call, OR use `git commit -o <paths>` (commits only explicit pathspec, ignores other staged content).
- **`git commit -o <paths>` with one half of a rename breaks rename detection.** Pathspec must include BOTH old AND new paths of a rename; otherwise git fragments into new-file + pending delete. Recovery: `git reset --soft HEAD~1` + plain `git commit` (preserves mixed index state; rename auto-detects at 98-100%). Safer for rename-heavy commits: skip `-o`, rely on verify-then-commit-in-same-bash-call pattern.
- **Dispatched agents' Edit tool writes to absolute paths, not worktree-relative.** Initial Edit calls may land in main worktree instead of agent worktree. In agent dispatch prompts, top-of-prompt: "your working directory is `{worktreePath}` — verify via `pwd`, use worktree-relative paths". Post-dispatch pre-merge: always `git status` on main worktree to confirm clean before `gh pr merge`.

## Feedback

<!-- User corrections and confirmed approaches that persist across sessions. Only record what isn't already in CLAUDE.md, workflows, or protocols. -->

- **Backlog.md at session start is a P0 scan, not a full read.** Use `grep -n "^## P0" Playbook/Backlog.md` or Read with `limit:` bracketed to P0 section. Full backlog review is the backlog-hygiene routine's job. *Why:* full-read on a 200-line file wastes ~4k tokens on a 200k window.
- **Autonomous Tier-1 bundle dispatch for small P2 knock-offs.** When user says "autonomous" / "knock-offs" / is dogfooding elsewhere, bundle 3-5 well-scoped independent P2 items into one Tier-1 patch plan, single worktree dispatch, independent code-review agent, same-session merge. Scope: Tier-1 only, crisp file-level specs, no item touches the track user is dogfooding in parallel.
- **Parallel dispatch for Tier-2 multi-phase plans when phases are file-disjoint.** User confirmed "parallel ok, no drop on quality of work." Pre-dispatch checklist: (1) verify file-disjoint (grep each phase spec for other phase's files), (2) no shared type/signature changes (signature extensions ship before dependent phases dispatch), (3) same quality-bar prompt to both agents, (4) independent code-review agent per PR in parallel, (5) fix-agents in parallel, (6) merge sequentially (first-landed auto-rebases second). Rule: parallel below 40% context, sequential above.

### Durable UX preferences (2026-04-17 dogfooding)

> Keep short. Extend when new patterns emerge — don't let it calcify.

#### Visual / UI

- **Palette first, visuals second.** For any TUI color/style change, check semantic tokens (`ColorPrimary`, `ColorAccent`, etc.) exist. Introduce as prereq if missing — otherwise every color change is codebase-wide find-replace.
- **Sleek and minimal over ornate.** "Bulky", "thrown together", flat spaced-letter wordmarks are taste-negative. Prefer tight wordmarks, compact boxes, meaningful glyphs. Strip and measure — don't ornament.
- **Visible state over hidden state.** Any answered prompt, selected item, or auto-processed section should leave visible evidence. Don't rely on Huh's default clear-on-submit — print a summary line after.
- **Rich guidance, not cramped.** Next-steps, file-structure views, showcase moments get dedicated real estate + substantive copy. Don't default to a single `Hint()` when user just finished something significant.
- **TUI should redraw, not stack.** On major step transitions (review, generate, complete), use AltScreen or explicit clear/redraw. Don't treat TUI output as scrollable log.

#### Planning / iteration

- **Fast iteration beats process for UX work.** "Test locally" = worktree for isolation, commit directly to main (or merge locally), skip PR creation. Save PR flow for code-correctness-heavy work.
- **Pick scope pragmatically, foundations first.** Group by dependency (palette before banner) + visible-win density. State deferred scope explicitly so user can redirect.
- **Propose scope before writing the plan.** For taste-heavy or ambiguous tasks, send scope summary (picked / deferred / rationale) and ask "OK to proceed?" before drafting plan file. Scope disagreement at plan-stage costs a rewrite.
- **Log findings as they surface, don't batch.** Maintain group tag in `Playbook/Backlog.md` during testing sessions. Each finding: category tag, group tag, fix options if known, source attribution.

#### Communication

- **Concise and direct wins.** User makes fast decisions with minimal elaboration. Mirror their energy — two sentences in, two sentences out.
- **Surface incidental findings proactively.** When hitting a workaround during setup/chores, flag it as a finding. Don't normalize broken behavior into your flow.

## References

<!-- Pointers to external resources not documented elsewhere in the project. -->

- **Foundational research docs** — Anchor for methodology/concept decisions.
    - [Research/RESEARCH-landscape-analysis.md](../../Research/RESEARCH-landscape-analysis.md) — Bonsai vs GSD/ECC/others; identity/coordination layer positioning
    - [Research/RESEARCH-concept-decisions.md](../../Research/RESEARCH-concept-decisions.md) — Ambient vs command-driven, authority hierarchy, catalog ownership, talents taxonomy
    - [Research/RESEARCH-eval-system.md](../../Research/RESEARCH-eval-system.md) — Eval system concept: scenarios, evaluators, benchmarks
    - [Research/RESEARCH-trigger-system.md](../../Research/RESEARCH-trigger-system.md) — Trigger section design research
    - [Research/RESEARCH-uiux-overhaul.md](../../Research/RESEARCH-uiux-overhaul.md) — UI/UX overhaul research
    - [Research/RESEARCH-proof-of-bonsai-effectiveness.md](../../Research/RESEARCH-proof-of-bonsai-effectiveness.md) — OSS launch proof-of-work pre-registration (cut-over `4dfd3f4` 2026-04-14). Pick up when ready — user answers §10 first.
