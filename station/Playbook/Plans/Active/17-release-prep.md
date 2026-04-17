---
tags: [plan, release, security, tooling]
description: Release-prep PR bundling Go toolchain bump, triggerSection frontmatter fix, and OSS polish (linter + Makefile targets).
---

# Plan 17 — Release Prep (Go toolchain + triggerSection frontmatter + OSS polish)

**Tier:** 2
**Status:** Draft
**Agent:** general-purpose (single dispatch — all phases sequential in one worktree / one PR)

## Goal

Ship a single `release-prep` PR that closes three release-blocking backlog items without touching UI/UX files that Plan 14 and Plan 15 are still iterating on:

1. Bump Go toolchain from `1.24.3` to `1.24.13` (stdlib CVEs).
2. Fix `triggerSection()` prepending above YAML frontmatter.
3. Add `.golangci.yml` + `test` / `lint` / `fmt` Makefile targets + CI lint job.

## Context

Phase 1 of the Roadmap (Foundation & Polish) is close to shippable, but three items remain that are either security-relevant or contributor-friction:

- **Item 1 — Go toolchain CVEs.** `go.mod` pins `toolchain go1.24.3`. Three symbol-level advisories (GO-2025-3956 os/exec, GO-2025-3750 O_CREATE|O_EXCL Windows, GO-2026-4602 FileInfo Root) affect the compiled binary. GO-2025-3956 + GO-2025-3750 are fixed in `go1.24.13` (released 2026-02-04); GO-2026-4602 requires `go1.25.8` and stays tracked.
- **Item 3 — `triggerSection()` frontmatter bug.** Found during `bonsai update` dogfooding (backlog, 2026-04-16). Callers at `internal/generate/generate.go:1289-1291` (skills) and `:1310-1312` (workflows) do `append([]byte(ts), content...)`. When `content` starts with YAML frontmatter (`---\n...---\n`), the trigger header lands above the frontmatter, breaking metadata parsing for any generated file that has both frontmatter and triggers.
- **Item 4 — OSS polish.** The repo ships a simple `Makefile` (`build`/`install`/`clean`) and a CI job that only runs `go test` + `go vet`. No lint config, no `test`/`lint`/`fmt` targets. Drops contributor onboarding quality.

**Out of scope for this PR (explicitly deferred):**

- Spinner error-swallowing (backlog P1) — 34 of 41 callsites live in `cmd/init|add|remove|update.go`, which Plan 15's BubbleTea harness rewrites. Fix it centrally during the harness migration, not here.
- Demo GIF (Group C item 2) — requires a user-recorded asciinema/VHS capture. Keep in backlog.
- All UI/UX files: `internal/tui/styles.go`, `internal/tui/prompts.go`, any of `cmd/*.go`. Plan 14 and Plan 15 own those.

## Architecture

Three mechanical, independent changes land in the same worktree / PR. No cross-phase dependencies — phases can be executed in any order, but verification is shared at the end. Single dispatch keeps the PR coherent (one release-prep artifact, one review pass).

## Steps

### Phase A — Go toolchain bump

**File:** `go.mod`

1. Change line 5 from `toolchain go1.24.3` to `toolchain go1.24.13`.
2. Leave line 3 (`go 1.24.2`) untouched — that's the minimum language version, not the toolchain.
3. Run `go mod tidy`. Expected: no changes to `go.sum` or require blocks. If anything does change, stop and report.

**Why no CI/goreleaser edits:**
- `.github/workflows/ci.yml:19` and `.github/workflows/release.yml:23` both use `go-version-file: go.mod` — they pick up the new toolchain automatically.
- `.goreleaser.yaml` has no explicit Go version pin.

### Phase B — triggerSection frontmatter fix

**File:** `internal/generate/generate.go`

1. Add a new unexported helper directly below `triggerSection()` (after the existing `return` at line 1173):

   ```go
   // injectTriggerSection inserts the trigger section after YAML frontmatter
   // when present, otherwise prepends it. A YAML frontmatter block is
   // recognized when content starts with "---\n" and contains a closing
   // "\n---\n" before any other content. Returns the original content
   // unchanged when ts is empty.
   func injectTriggerSection(ts string, content []byte) []byte {
       if ts == "" {
           return content
       }
       const open = "---\n"
       if bytes.HasPrefix(content, []byte(open)) {
           rest := content[len(open):]
           if idx := bytes.Index(rest, []byte("\n---\n")); idx >= 0 {
               end := len(open) + idx + len("\n---\n")
               out := make([]byte, 0, len(content)+len(ts))
               out = append(out, content[:end]...)
               out = append(out, []byte(ts)...)
               out = append(out, content[end:]...)
               return out
           }
       }
       return append([]byte(ts), content...)
   }
   ```

   If `bytes` is not already imported at the top of the file, add it (check imports first — package uses `bytes` elsewhere almost certainly, but verify).

2. Replace the skills call site at `internal/generate/generate.go:1289-1291`:

   ```go
   // Before:
   if ts := triggerSection(item, installed.Workspace, "skill", false); ts != "" {
       content = append([]byte(ts), content...)
   }
   // After:
   content = injectTriggerSection(triggerSection(item, installed.Workspace, "skill", false), content)
   ```

3. Replace the workflows call site at `internal/generate/generate.go:1310-1312`:

   ```go
   // Before:
   if ts := triggerSection(item, installed.Workspace, "workflow", CuratedSlashWorkflows[wfName]); ts != "" {
       data = append([]byte(ts), data...)
   }
   // After:
   data = injectTriggerSection(triggerSection(item, installed.Workspace, "workflow", CuratedSlashWorkflows[wfName]), data)
   ```

4. Add unit tests in `internal/generate/generate_test.go` (create if needed, or extend existing). Table-driven:

   ```go
   func TestInjectTriggerSection(t *testing.T) {
       ts := "## Triggers\n\n**Slash command:** `/foo`\n\n---\n\n"
       tests := []struct {
           name    string
           ts      string
           content string
           want    string
       }{
           {
               name:    "empty ts returns content unchanged",
               ts:      "",
               content: "---\nfoo: bar\n---\n# Title\n",
               want:    "---\nfoo: bar\n---\n# Title\n",
           },
           {
               name:    "no frontmatter prepends as before",
               ts:      ts,
               content: "# Title\nbody\n",
               want:    ts + "# Title\nbody\n",
           },
           {
               name:    "frontmatter present: ts lands after closing ---",
               ts:      ts,
               content: "---\nfoo: bar\n---\n# Title\nbody\n",
               want:    "---\nfoo: bar\n---\n" + ts + "# Title\nbody\n",
           },
           {
               name:    "opens with --- but no closing fence: prepend",
               ts:      ts,
               content: "---\nfoo: bar\n# Title\n",
               want:    ts + "---\nfoo: bar\n# Title\n",
           },
       }
       for _, tc := range tests {
           t.Run(tc.name, func(t *testing.T) {
               got := string(injectTriggerSection(tc.ts, []byte(tc.content)))
               if got != tc.want {
                   t.Errorf("got %q, want %q", got, tc.want)
               }
           })
       }
   }
   ```

### Phase C — OSS polish

1. **New file: `.golangci.yml`** at repo root. Standard conservative config — only rules with very low false-positive rates, so CI stays green and contributors aren't drowned in noise on day one.

   ```yaml
   # golangci-lint v2 config
   # https://golangci-lint.run/usage/configuration/
   version: "2"

   run:
     timeout: 3m
     tests: true

   linters:
     default: none
     enable:
       - errcheck      # unchecked errors
       - gosimple      # simplification suggestions
       - govet         # suspicious constructs
       - ineffassign   # ineffective assignments
       - staticcheck   # broad static analysis
       - unused        # unused code
       - misspell      # typos in comments/strings
       - gofmt         # formatting
       - goimports     # imports ordering

   formatters:
     enable:
       - gofmt
       - goimports

   issues:
     max-issues-per-linter: 0
     max-same-issues: 0
   ```

   If the installed `golangci-lint` version on the agent's system is v1 rather than v2, the agent should use the v1-compatible shape (no `version` key, `linters-settings` instead of the v2 shape). Agent is expected to run `golangci-lint --version` and pick the matching schema. Report which version was used in the PR body.

2. **Modify `Makefile`** — add `test`, `lint`, `fmt`, `tidy` targets. Keep existing `build`/`install`/`clean` untouched.

   ```makefile
   .PHONY: build install clean test lint fmt tidy

   VERSION ?= dev
   LDFLAGS := -s -w -X main.version=$(VERSION)

   build:
   	go build -ldflags "$(LDFLAGS)" -o bonsai .

   install:
   	go install -ldflags "$(LDFLAGS)" .

   clean:
   	rm -f bonsai

   test:
   	go test ./...

   lint:
   	golangci-lint run ./...

   fmt:
   	gofmt -s -w .
   	goimports -w .

   tidy:
   	go mod tidy
   ```

3. **Modify `.github/workflows/ci.yml`** — add a `lint` job that runs in parallel with `test`. Do NOT merge into the existing `test` job; parallel keeps feedback faster.

   ```yaml
   name: CI

   on:
     pull_request:

   permissions:
     contents: read

   jobs:
     test:
       runs-on: ubuntu-latest
       steps:
         - name: Checkout
           uses: actions/checkout@v4

         - name: Set up Go
           uses: actions/setup-go@v5
           with:
             go-version-file: go.mod

         - name: Run tests
           run: go test ./...

         - name: Run vet
           run: go vet ./...

     lint:
       runs-on: ubuntu-latest
       steps:
         - name: Checkout
           uses: actions/checkout@v4

         - name: Set up Go
           uses: actions/setup-go@v5
           with:
             go-version-file: go.mod

         - name: golangci-lint
           uses: golangci/golangci-lint-action@v6
           with:
             version: latest
   ```

4. **Do not edit README.md** — demo GIF is deferred.

## Dependencies

- go1.24.13 available on the agent's system (GitHub Actions `setup-go@v5` resolves it automatically in CI via `go-version-file`).
- `golangci-lint` installed locally for the agent to dry-run `make lint`. If not installed, the agent should `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest` inside the worktree and note it in the PR body. It does NOT need to be committed — CI uses the action.

## Security

> [!warning]
> Refer to [Playbook/Standards/SecurityStandards.md](../../Standards/SecurityStandards.md) for all security requirements.

- Phase A is a security patch — purpose is to pick up stdlib CVE fixes. Verify the built binary reports `go1.24.13` or later via `./bonsai --version` debugging or `go version -m ./bonsai`.
- Phase B changes file-generation output format. No new network I/O, no new untrusted input paths. Frontmatter detection is a pure byte-level prefix/scan — no regex, no eval.
- Phase C adds static analysis. `.golangci.yml` must not enable any linter that would block on the existing codebase. If `make lint` fails on first run, the agent must either (a) scope the linter set tighter, or (b) stop and report — do NOT "fix" lint findings inline, that's scope creep.

## Verification

Run from the worktree root before creating the PR:

- [ ] `go mod tidy` — produces no changes beyond the toolchain line
- [ ] `grep -n "toolchain" go.mod` — shows `toolchain go1.24.13`
- [ ] `make build` — succeeds, produces `./bonsai`
- [ ] `go test ./...` — all tests pass, including the new `TestInjectTriggerSection` cases
- [ ] `make test` — same as above, via the new target
- [ ] `make fmt` — no diff after run (code already formatted)
- [ ] `make lint` — passes clean (if it fails, stop and report — do not fix findings)
- [ ] `make tidy` — no diff
- [ ] `git diff --stat` touches only: `go.mod`, `internal/generate/generate.go`, `internal/generate/generate_test.go`, `Makefile`, `.golangci.yml`, `.github/workflows/ci.yml`. Anything else is scope creep — stop and report.
- [ ] Manual: in a scratch dir, `./bonsai init` then `./bonsai add` a skill that has both frontmatter AND triggers (e.g., `issue-classification`). Open the generated `agent/Skills/issue-classification.md` and confirm the `---\nfoo\n---\n` frontmatter appears first, `## Triggers` appears after it. Before this fix, the triggers would appear above the frontmatter.

## Dispatch

Single agent, `isolation: "worktree"`, `subagent_type: "general-purpose"`. Pass only this plan file — no other conversation context.

**Branching & PR target — READ CAREFULLY:**
- The worktree must branch **off `release-prep`** (not `main`). `release-prep` is the long-lived integration branch for the v0.2 release; nothing lands on `main` until the full release is ready.
- The draft PR must target **`release-prep`** (`gh pr create --base release-prep --draft ...`). Do NOT target `main`.
- Agent reports the PR URL.

> [!warning]
> Memory note (2026-04-17): subagent tool inheritance for `gh` was flaky — bundle ALL gh operations (push, create draft PR) into this single dispatch. Do not split. If the agent reports `gh: command not found`, ask them to fall back to pushing the branch and reporting the push URL so user/Tech Lead can open the PR via web targeting `release-prep`.

## Out of Scope (do not touch in this PR)

- `internal/tui/styles.go`, `internal/tui/prompts.go` — Plan 14 iterations
- `cmd/init.go`, `cmd/add.go`, `cmd/remove.go`, `cmd/update.go`, `cmd/root.go` — Plan 15 BubbleTea harness migration
- Spinner error-swallowing in any cmd/* — Plan 15 centralizes this
- README demo GIF — user-recorded, separate backlog item
- Any new lint-finding "cleanups" — if golangci-lint complains about existing code, stop and report
