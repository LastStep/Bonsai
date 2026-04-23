# Plan 18 — `bonsai guide` multi-topic + legacy docs cleanup

**Tier:** 2
**Status:** Shipped — PR #25 merged 2026-04-20 (squash `e448140`)
**Agent:** general-purpose (single worktree)

## Goal

Expand `bonsai guide` from a single-doc renderer into a multi-topic offline reference with four tight terminal cheatsheets (`quickstart`, `concepts`, `cli`, `custom-files`), and delete three orphaned legacy docs whose content has been superseded by the Starlight site.

## Context

The Starlight site (`laststep.github.io/Bonsai`) shipped in Plan 10 and is now the canonical docs source. Three long-form docs that pre-dated the site remain in the repo as orphans, not referenced from the current README or Go code:

| File | Lines | Replaced by |
|------|------|------------|
| `HANDBOOK.md` (root) | 515 | `concepts/*.mdx` + `guides/*.mdx` on Starlight |
| `docs/working-with-agents.md` | 403 | `guides/working-with-agents.mdx` |
| `docs/triggers.md` | 295 | `guides/triggers-and-activation.mdx` |

**Total orphan content:** 1,213 lines.

The original backlog entry (Group A) proposed writing full-length `docs/quickstart.md`, `docs/concepts.md`, `docs/cli-usage.md` files, but that was drafted before the Starlight site existed. Full mirrors would duplicate the website; dropping the feature would contradict the README's advertised `bonsai guide` surface.

**Decision (2026-04-20 session):** terminal-friendly **cheatsheets** — 80–120 lines each, task-oriented, ending with a deep-link to the website — served via interactive Huh picker or direct argument. Legacy orphans are deleted first to leave a clean `docs/` folder.

## Steps

### Phase A — Legacy cleanup (prep)

1. Delete `HANDBOOK.md` (repo root, 515L)
2. Delete `docs/working-with-agents.md` (403L)
3. Delete `docs/triggers.md` (295L)
4. Verify no remaining references:
   - `Grep "HANDBOOK|working-with-agents\.md|docs/triggers\.md"` across repo, excluding `.claude/worktrees/`
   - Expected remaining hits: plan files (history), backlog removed-items comments (history). Zero live code/doc references.
5. Rewrite `docs/README.md` (currently 7 lines saying "source files migrated to docs site" — inaccurate once cheatsheets land). Replace with:
   ```markdown
   # docs/

   Markdown files in this folder are embedded into the `bonsai` binary and rendered
   by `bonsai guide`. Current topics: `quickstart`, `concepts`, `cli`, `custom-files`.

   Full user documentation (tutorials, reference, catalog browser) lives on the
   Starlight site: <https://laststep.github.io/Bonsai/>. Source under [`website/`](../website/).

   `docs/assets/` holds images referenced from `README.md`.
   ```

### Phase B — Cheatsheet content

> [!note]
> All cheatsheets must render cleanly via glamour (no HTML tags, no MDX components). Target 80–120 lines each including frontmatter. End every file with a single line: `> Full guide: https://laststep.github.io/Bonsai/<path>/`.

6. Write `docs/quickstart.md` (target ~80–100 lines)
   - Frontmatter: `description: Post-install 5-step walkthrough for new Bonsai users.`
   - Sections: "You just ran `bonsai init` — what's next?" (5 numbered steps)
     1. Open `CLAUDE.md` — the generated nav
     2. Start a session in Claude Code — say "hi, get started"
     3. Add a code agent via `bonsai add`
     4. Understand the Status / Plans / Reports flow
     5. When and how to run routines
   - Footer link: `guides/your-first-workspace/`

7. Write `docs/concepts.md` (target ~100–120 lines)
   - Frontmatter: `description: The mental model — station, instruction stack, agents, sensors, routines, scaffolding.`
   - Sections:
     - The Station (single directory, Tech Lead home)
     - The Instruction Stack (6-layer diagram + one-line each)
     - Agents as Teammates (Tech Lead orchestrates; code agents implement)
     - Sensors (auto-enforced via hooks)
     - Routines (periodic, Tech Lead, opt-in)
     - Scaffolding (shared project state: Status, Roadmap, Plans, Logs, Reports)
     - "When to use what" table
   - Footer link: `concepts/how-bonsai-works/`

8. Write `docs/cli.md` (target ~100–120 lines)
   - Frontmatter: `description: Concise reference for every bonsai command.`
   - One section per command (`init`, `add`, `remove`, `list`, `catalog`, `update`, `guide`) with: one-line description, example invocation, "use when" trigger, "gotcha" line if relevant.
   - Footer link: `commands/init/`

9. Keep `docs/custom-files.md` unchanged (already terminal-friendly, actively embedded).

### Phase C — Embed the new content

10. Update `embed.go` — rename existing `GuideContent` → `GuideCustomFiles` and add three new directives. Final state:
    ```go
    // Package bonsai exposes the embedded catalog and guide content used by the
    // bonsai CLI. It exists so //go:embed directives stay at the repo root, where
    // the embedded paths (catalog/, docs/*.md) live.
    package bonsai

    import "embed"

    //go:embed all:catalog
    var CatalogFS embed.FS

    //go:embed docs/custom-files.md
    var GuideCustomFiles string

    //go:embed docs/quickstart.md
    var GuideQuickstart string

    //go:embed docs/concepts.md
    var GuideConcepts string

    //go:embed docs/cli.md
    var GuideCli string
    ```

    Use explicit per-file directives, not a glob — keeps the embed surface audit-able.

### Phase D — Plumbing: pass topics through `cmd.Execute`

> [!note]
> Current flow (see `cmd/bonsai/main.go:22` → `cmd/root.go:79` → `cmd/guide.go:23`): `main.go` passes a single `string` to `cmd.Execute(fsys, guide string)`, which stores it in the `guideMarkdown` package global for `cmd/guide.go` to read. This plan replaces that with a map of topics.

11. In `cmd/root.go`:
    - Change signature: `func Execute(fsys fs.FS, guides map[string]string)`.
    - Replace `var guideMarkdown string` with `var guideContents map[string]string`.
    - In the function body, replace `guideMarkdown = guide` with `guideContents = guides`.

12. In `cmd/bonsai/main.go:22`, replace the single-string call with:
    ```go
    cmd.Execute(sub, map[string]string{
        "quickstart":   bonsai.GuideQuickstart,
        "concepts":     bonsai.GuideConcepts,
        "cli":          bonsai.GuideCli,
        "custom-files": bonsai.GuideCustomFiles,
    })
    ```

### Phase E — CLI wiring in `cmd/guide.go`

13. Refactor `cmd/guide.go`:
    - Define an ordered topic list (separate from `guideContents` so picker order is deterministic and labels are attached):
      ```go
      type guideTopic struct {
          Key   string  // "quickstart"
          Label string  // "Quickstart — 5-step post-install walkthrough"
      }
      var guideTopics = []guideTopic{
          {"quickstart", "Quickstart — 5-step post-install walkthrough"},
          {"concepts", "Concepts — the mental model"},
          {"cli", "CLI — command-by-command reference"},
          {"custom-files", "Custom Files — add your own abilities"},
      }
      ```
    - `bonsai guide` (no args) → Huh `Select[string]` form listing topics by label, returning the selected `Key`; render `guideContents[key]` via shared helper.
    - `bonsai guide <topic>` → lookup `guideContents[topic]`. If absent, return `fmt.Errorf("unknown topic %q. Available: quickstart, concepts, cli, custom-files", topic)` (non-zero exit).
    - Extract existing frontmatter-strip + glamour logic into a `renderMarkdown(content string) (string, error)` helper; call from both paths.

14. Update Cobra metadata on `guideCmd`:
    - `Use: "guide [topic]"`
    - `Short: "View bundled guides in the terminal."`
    - `Long: "Render one of the bundled guides as styled terminal output. Run without a topic to pick interactively, or pass one of: quickstart, concepts, cli, custom-files."`
    - `Args: cobra.MaximumNArgs(1)` — rejects `bonsai guide foo bar` before `runGuide` is called.

### Phase F — Downstream documentation updates

15. Update `README.md` commands table (line ~132): `bonsai guide` → `View bundled guides: quickstart, concepts, cli, custom-files`.

16. Update `website/src/content/docs/commands/guide.mdx` to describe multi-topic behavior: picker on no-args, direct arg with topic, list of available topics, example for each. Strip or revise the "read-only with no interactive prompts" line (picker is interactive).

17. Update `station/code-index.md` — current line 12 (if still references `main.go:18 //go:embed docs/custom-files.md`) needs to point to `embed.go` and list all four embedded guide vars.

18. **Roll in CLAUDE.md doc-drift fix** (per user instruction):
    - `Bonsai/CLAUDE.md` line 19: `├── main.go ← entry point, embeds catalog/ via embed.FS` → `├── cmd/bonsai/main.go ← entry point` + add `├── embed.go ← root embed package (CatalogFS + guide vars)`
    - `Bonsai/CLAUDE.md` line ~109: `Catalog is embedded via embed.FS in main.go` → `Catalog is embedded via embed.FS in embed.go (package bonsai at repo root)`
    - After this fix, remove the matching P1 doc-drift backlog entry.

### Phase G — Verification

- [ ] `make build` succeeds
- [ ] `./bonsai --version` returns the expected version string
- [ ] `./bonsai guide` with no args opens the Huh picker; arrow keys + enter render the selected topic without error
- [ ] `./bonsai guide quickstart` renders directly
- [ ] `./bonsai guide concepts` renders directly
- [ ] `./bonsai guide cli` renders directly
- [ ] `./bonsai guide custom-files` renders directly (regression)
- [ ] `./bonsai guide unknown` prints `unknown topic "unknown". Available: ...` and exits non-zero
- [ ] `./bonsai guide a b` rejects with cobra's "accepts at most 1 arg" error
- [ ] Each cheatsheet file is ≤ 120 lines of source
- [ ] Glamour output contains no visible HTML tags (`<div>`, `<Aside>`, `<Steps>`, etc.)
- [ ] Each cheatsheet ends with a live `https://laststep.github.io/Bonsai/...` link
- [ ] `rg "HANDBOOK|working-with-agents\.md|docs/triggers\.md" -g '!.claude/worktrees'` returns only plan-history and backlog-history hits (no live references)
- [ ] `grep -r "GuideContent" --include="*.go" --exclude-dir=.claude` returns zero hits (rename complete)
- [ ] `go test ./...` passes
- [ ] `make lint` passes (no new lint violations from the refactor)

## Dependencies

None — self-contained feature. Does not touch Plan 15's BubbleTea work on `ui-ux-testing`.

## Security

> [!warning]
> Refer to `Playbook/Standards/SecurityStandards.md` for all security requirements.

Plan-specific notes:
- All guide content is embedded at compile time via `//go:embed` — no runtime file I/O
- `bonsai guide <topic>` validates `topic` against a static allowlist before lookup — no path traversal surface
- No user input is interpolated into shell commands, filesystem paths, or network calls

## Out of scope (explicit non-goals)

- Auto-generating cheatsheets from Starlight MDX (considered; deferred)
- Full-length mirrors of Starlight pages (considered; rejected in favor of cheatsheets)
- `CHANGELOG.md` backfill — separate OSS polish item
- Final pre-release docs audit across README / SECURITY / CONTRIBUTING / Starlight — separate follow-up (backlog entry added)
- `good first issue` seeding — separate OSS polish item
- Demo GIF recording — user task
- Spinner error swallowing (P1) — owned by Plan 15 harness migration
