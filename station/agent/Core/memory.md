---
tags: [core, memory]
description: Tech Lead Agent working memory — flags, work state, notes.
---

# Working Memory

## Flags

<!-- Active flags go here. Format: - [FLAG] description -->

(none)

## Work State

**Current task:** Backlog hygiene pass on branch `claude/get-started-PlBaC`. Main at `5248212` (PR #43 Group G closeout merged).
**In flight (other tracks):** None.
**Blocked on:** Nothing.
**Loose ends (uncommitted):** Backlog hygiene edits in progress this session.
**Last completed:** PR #43 (`5248212`, 2026-04-21) — Group G Dependabot closeout + CodeQL v3→v4 deprecation note. Dependabot Group G (8 PRs squash-merged 2026-04-21): #39 ansi patch, #38 go-isatty patch, #37 x/term 0.36→0.42 (rebased 1×), #36 setup-node v6, #35 setup-go v6, #34 golangci-lint-action v9, #33 upload-pages-artifact v5, #32 goreleaser-action v7. Plan 20 core (6 PRs): #28 #29 #30 #31 #40 #41; wrap-up PR #42 `9cf577e`; gitleaks 0/156.

Bundle: 4 Tier 1 fresh-install blockers (CRLF defence via .gitattributes + `normalizeShellLF`; `showWriteResults` cross-workspace bucket-by-top-segment; `applyCustomFileSelection` dedup via `appendUnique`; spinner `errors.Join` at ~30 sites + Warning surfaces) with 4 Tier 2 harness polish items (SpinnerStep goroutine `recover`; `NewConditional` nil-predicate guard; Esc-back predicate re-eval via harness SetPrior-before-Reset + `Conditional.Reset` re-evaluation; `bonsai add` drop duplicate NoteStep). Issue-to-implementation workflow end-to-end, worktree-isolated dispatch (general-purpose agent, ~15min wall), 6 new tests + 3 subtests all green. Plan doc: `Plans/Active/19-pre-launch-bug-sweep.md`. Session log: `Logs/2026-04-21-plan-19-bug-sweep.md`.

## Notes

<!-- Session-to-session notes. Keep concise. -->

- **Plan 14 code shipped to main via PR #24 bundle (2026-04-17).** Iterations 1 + 2 (palette tokens, banner, answered-prompt persistence, required-only feedback, category counts, prompt polish, mustCwd error surfacing, width-aware TitledPanel) merged along with Plan 17's release-prep work. Phase 4+ scope (screen lifecycle, progressive disclosure, go-back nav, flow redesign) still deferred — Plan 15's BubbleTea harness is the likely vehicle for those.
- **Parallel session convention.** `ui-ux-testing` branch is where Plan 15 iterations land (separate session). Don't cross-pollinate planning-doc edits between branches — each branch's Status.md/memory.md should stay scoped to its own track. Merging to main collapses everything cleanly.
- **CI lint gotcha.** `golangci/golangci-lint-action@v6` with `version: latest` resolves to v1.x, not v2. `.golangci.yml` must use v1 schema (`linters.disable-all: true`, no `version:`/`formatters:` keys). If we ever move to v2, update both the action pin (`version: v2.x`) and the config schema together. Learned from PR #24 `66a6304`.
- **Plan 08 Phase C (new sensors) paused** — moved back to Pending while Plan 14 ships. Resume once UI/UX overhaul series wraps or explicitly requested.
- **Pre-flight learning:** Worktrees inherit only committed HEAD — uncommitted plans/docs in main tree are invisible to dispatched agents. Commit station/ planning artifacts before dispatch.
- **PR review memory hygiene:** "both reviews APPROVE" from prior session was dispatched review agents, not GitHub reviews. `gh pr view --json reviews` returned empty. When noting review status, distinguish agent-dispatched reviews (in `Reports/`) from GitHub formal reviews.
- **Squash-bundle rebase pattern.** When main has absorbed a feature branch's commits via squash-merge (e.g., Plan 14 individual commits → bundled into Plan 17's `bc565bf` squash), straight `git rebase main feature-branch` will conflict on every doc snapshot because main has moved past those states even though the *code* matches. Cleaner approach: `git reset --hard main && git cherry-pick <only-the-truly-new-commits>` after listing each commit and asking "is this code/content already on main via the squash?" Always preserve old tip as `<branch>-pre-rebase` first. Used 2026-04-20 to rebase `ui-ux-testing` (13 commits → 3 commits actually carried forward).
- **Worktree cwd gotcha from Claude Code.** `git worktree add ../<name> <branch>` runs in the bash tool's cwd, which is often a subdirectory (for us: `station/`), not the repo root. So `../Bonsai-uiux` resolves to `/home/rohan/ZenGarden/Bonsai/Bonsai-uiux` (nested inside the main worktree — violates the no-nesting rule), not the intended sibling `/home/rohan/ZenGarden/Bonsai-uiux`. **Always use absolute paths when creating worktrees from this session.** Hit this 2026-04-20 setting up the Plan 15 parallel-session worktree; caught, removed, recreated at the correct sibling path.
- **Subagent tool inheritance is flaky.** On 2026-04-17, the original executing agent (worktree, `gh` authed) successfully created draft PR #23, but every subsequent subagent I spawned for merge/verification reported `gh: command not found` — the environment wasn't inherited from the spawning agent or from my own shell (Windows Git Bash over WSL UNC path, no `gh` on PATH). Implication: for PR-flow tasks, bundle **all** gh operations (push, create PR, mark ready, merge, delete branch) into a **single** agent dispatch rather than splitting across agents. Second-best: ask the user to click-merge via web UI when a subagent-less step is needed.
- **PR CI ≠ main CI.** The `CI` workflow (test + lint + GitGuardian) runs on pull_request, but `Deploy Docs` only runs on `push` to main. So a broken `website/**` change can pass PR CI and then fail on main — happened 2026-04-20 with PR #25's MDX autolink. When reviewing PRs that touch `website/`, run `cd website && npm run build` locally before merging, or add a `website/**` path trigger to Deploy Docs on pull_request (separate backlog item).
- **MDX autolink gotcha.** `<https://example.com>` is valid GitHub-flavored markdown but MDX parses `<` as JSX — Astro build fails with "Unexpected character `/`". Always use `[label](url)` inside `.mdx` files. Incident: PR #25 `guide.mdx:20`, fixed in `e336ccb`.
- **Local `go build` + `go test` miss `golangci-lint unused`.** Plan 15 iter 3 left `resolveConflicts` in `cmd/root.go` as dead code after the harness migration. Local verification (`make build`, `go test ./...`, `go vet ./...`) all passed; only the CI `lint` step (golangci-lint v1.64.8 with the `unused` linter) caught it. Local golangci-lint on this machine is v2.11.4 and the repo config is v1, so `golangci-lint run` errors out locally with "unsupported version of the configuration". Implication: for PRs that heavily refactor/delete, trust CI to catch dead code; OR install a v1 golangci-lint binary locally (e.g. `go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.8`) for pre-push. PR #26 hit this; fixed in `070bcd0`.
- **Post-squash-merge local cleanup gotcha.** `gh pr merge --squash --delete-branch` successfully squashed to main and deleted the remote branch, but the local-branch-delete step failed with `'main' is already checked out at '/home/rohan/ZenGarden/Bonsai'` because the CLI tried to fast-forward the local sibling worktree's main and that worktree was on main. Remote branch deletion also failed silently in the same call — needed a follow-up `git push origin --delete ui-ux-testing`. Check `gh pr view --json state` + `git ls-remote origin | grep <branch>` after any squash-merge from a non-main worktree to confirm both halves completed. Confirmed again 2026-04-21 × 4 during Plan 20 (#28, #29, #31, #40, #41 all hit the same pattern) — always need manual `git push origin --delete <br>` + `git worktree remove -f -f <path>` + `git branch -D <br>`.
- **golangci-lint binary Go-version coupling.** The `version: X.Y` input on `golangci/golangci-lint-action` resolves to a specific pre-built binary. That binary must have been built with a Go version ≥ the repo's target. Moving a repo to Go 1.25 required golangci-lint v2.11.4+ (older v2.1 binary was still built on Go 1.24 and errors with `can't load config: the Go language version (go1.24) used to build golangci-lint is lower than the targeted Go version (1.25.8)`). Rule: when bumping Go major in go.mod, also pin golangci-lint-action to a version whose release notes say "built with Go ≥ target". Track latest v2.x: https://github.com/golangci/golangci-lint/releases. Hit this 2026-04-21 across PR #28 / #29 during Plan 20.
- **Agent worktrees base off origin/main, not local main.** `Agent(isolation: "worktree")` fetches origin/main as the worktree base — ANY uncommitted work on the local-only main branch is invisible to the agent and won't be in its PR. To ship local-only work, create a dedicated branch from the local tip, push, PR, merge first — THEN dispatch agents. Learned 2026-04-21 during Plan 20 pre-flight: my 65cfdd6 chore commit was on local main only; first Go-bump PR #28 came back clean (2 lines) precisely because agent worktree didn't see it. Had to ship as separate PR #30 before resuming.

## Feedback

<!-- User corrections and confirmed approaches that persist across sessions. Only record what isn't already in CLAUDE.md, workflows, or protocols. -->

- **Backlog.md at session start is a P0 *scan*, not a full read.**
    - **Why:** During the 2026-04-17 context audit, the CLI session-start payload was ~4k heavier than the app's, and the delta was traced to my having `Read` Backlog.md in full (~200 lines, ~4k tokens) when session-start protocol explicitly says "scan, look for P0 items only." Full backlog review is the backlog-hygiene routine's job. On a 200k window that's ~4% of headroom burned for zero benefit.
    - **How to apply:** At session start, use `Grep -n "^## P0" Playbook/Backlog.md` (or read with `limit:` bracketed to the P0 section) instead of `Read` on the whole file. Only full-read Backlog when explicitly running backlog-hygiene or planning across multiple groups.

## References

<!-- Pointers to external resources not documented elsewhere in the project. -->

- **Foundational research docs** — Anchor for methodology/concept decisions. Reference these when improvements need grounding in the core philosophy (ambient behavior, meta-layer, talents, evals).
    - [Research/RESEARCH-landscape-analysis.md](../../Research/RESEARCH-landscape-analysis.md) — How Bonsai compares to GSD/ECC/others; Bonsai's identity/coordination layer positioning
    - [Research/RESEARCH-concept-decisions.md](../../Research/RESEARCH-concept-decisions.md) — Ambient vs. command-driven, authority hierarchy, catalog ownership, talents taxonomy
    - [Research/RESEARCH-eval-system.md](../../Research/RESEARCH-eval-system.md) — Eval system concept: scenarios, evaluators, benchmarks for methodology rigor
    - [Research/RESEARCH-trigger-system.md](../../Research/RESEARCH-trigger-system.md) — Trigger section design research
    - [Research/RESEARCH-uiux-overhaul.md](../../Research/RESEARCH-uiux-overhaul.md) — UI/UX overhaul research (Plan 14/15 origin)
