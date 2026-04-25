---
title: Plan 32 — Website Concept Page Rewrite
status: Active
tier: 1
created: 2026-04-25
---

# Plan 32 — Website Concept Page Rewrite

**Tier:** 1
**Status:** Active
**Agent:** general-purpose (single dispatch)

## Goal

Rewrite three website copy surfaces to match the README's audience-first, mechanism-over-personality voice — pivoting away from the "structured language / structured instruction files" framing left over from pre-v0.2.0 docs.

## Context

- v0.2.0 audit (PR #65) did mechanical tagline replace across `index.mdx`, `astro.config.mjs`, `why-bonsai.mdx` — body paragraphs still carry old framing.
- Plan 25 (commit `d6785e7`) established the canonical voice: audience-first opener ("Solo devs and small teams who want to give their coding agent real responsibility"), mechanism bullets (every session starts from same context / project navigable not searchable / rules in files not prompts / plans before action / everything auditable), and a banned-phrase list.
- README at `/home/rohan/ZenGarden/Bonsai/README.md` is the canonical reference for tone + framing — match its register, not its length.

### Banned phrases (Plan 25, must not appear in output)

- `by design`
- `powerful out of the box`
- `One binary. No runtime.` / `One binary. No runtime. Works with any project.`
- `teammates, not tools`
- `structured language` (any "structured language" framing)
- `structured instruction files` (the AI-smell variant in the LLM description)

### Voice rules (extracted from README + Plan 25)

- **Audience-first.** Open with who it's for + what concrete responsibility they want their agent to have. Never open with abstract product description.
- **Mechanism over personality.** Don't describe what Bonsai "feels like" or how agents "behave" — describe what specific files/hooks/protocols do. Concrete file paths and event names beat adjectives.
- **Concrete artifacts beat metaphors.** Reference real generated files (`Playbook/Plans/Active/`, `agent/Protocols/security.md`, `Logs/KeyDecisionLog.md`) when explaining a behavior, not abstract layer names.
- **No salesy modifiers.** Strip "powerful", "complete", "battle-tested", "wired-up", "rock-solid".
- **Match register, not length.** README is dense; concept pages can stay shorter — but every sentence must add a fact, not a feeling.

## Steps

### Step 1 — `website/src/content/docs/concepts/how-bonsai-works.mdx` full rewrite

**Goal:** Replace abstract instruction-stack framing with mechanism-led explanation. Preserve the FileTree component (it's load-bearing for showing the generated layout) and preserve the existing imports.

**Required structure (use these section headings):**

1. **Frontmatter** — keep `title: How Bonsai Works`. Update `description` to remove "the mental model" phrasing — say what the page is, not what it represents. Suggested: `description: What bonsai init generates, how the pieces fit together, and how the Tech Lead orchestrates code agents.`
2. **Lede paragraph** — 2 sentences. State plainly: Bonsai is a CLI generator (not a runtime). It writes a workspace under `station/` plus per-agent workspaces; Claude Code agents read those files via hooks at session start and during tool use. No "mental model", no "structured instruction set" phrasing.
3. **`## What `bonsai init` generates`** — replace the existing "Generated layer / Purpose" table with a tighter version that uses **concrete file/directory references** in the left column (e.g. `agent/Core/identity.md` + `agent/Core/memory.md` instead of `Identity + Memory`). Keep the table to 7 rows max — one per generated category. Right column says what that file/directory does in **one sentence each**.
4. **`## The station`** — keep the section title (lowercase "station" inside, since it's a noun, not a proper noun). 2-3 sentence paragraph. Mention `station/` is the Tech Lead's workspace + the shared scaffolding that every code agent's plans/status/decisions reference. Remove the bulleted list (the FileTree component later in the page already shows the layout).
5. **`## The instruction stack`** — keep the existing 6-layer ASCII diagram exactly as-is. Replace the Aside block underneath ("Claude Code agents are only as good as the instructions they have access to...") with a one-paragraph explanation of WHY layers exist: **Core/Protocols load every session via the `session-context` SessionStart hook; Workflows/Skills are pulled in on demand; Sensors fire automatically on Claude Code tool-use events.** Reference the actual hook events (`SessionStart`, `PreToolUse`, `PostToolUse`, `Stop`) — those names are concrete, "automated enforcement" is fluff.
6. **`## How it works`** — keep the existing ASCII tree showing `You → Tech Lead → Backend / Frontend / DevOps / Security`. Keep both Asides (the "never writes application code" caution and the "you don't have to go through the Tech Lead" tip) — they're already mechanism-grounded. Tighten if you can do so without losing information.
7. **`## What gets generated`** — keep the FileTree component verbatim. Tighten the lede sentence above it to drop "After `bonsai init` + `bonsai add` (backend agent):" → just say `After bonsai init followed by bonsai add backend:`. Keep the closing "Every file reference is a clickable markdown link..." sentence.

**Hard constraints for this file:**
- Preserve all `import { ... } from '@astrojs/starlight/components';` lines.
- Preserve all `<Aside>`, `<FileTree>`, `<CardGrid>`, `<Card>` components — only edit their text content, never their tags or attributes.
- Do not introduce `<` autolinks (e.g. `<https://example.com>`) — Astro/MDX parses `<` as JSX. Use `[label](url)` for any URL.
- Do not add new sections beyond the 7 listed above.
- Do not add a `## Why this exists` or `## Philosophy` section — that's the job of `why-bonsai.mdx`.

### Step 2 — `website/src/content/docs/why-bonsai.mdx` body rewrite

**Goal:** Replace banned-phrase paragraphs with the README's audience-first mechanism-led framing. Preserve the `Card` / `CardGrid` components.

**Required structure:**

1. **Frontmatter** — keep `title: Why Bonsai`. Rewrite `description` to drop "as a layered system" phrasing. Suggested: `description: What problem Bonsai solves and what it gives your coding agent that a single CLAUDE.md doesn't.`
2. **`## The problem`** (lowercase "problem" in the heading is fine, Starlight title-cases display) — rewrite the existing 2 paragraphs. Open by mirroring the README's framing: out-of-the-box Claude Code is an assistant; the moment you want it to (a) pick up where it left off, (b) stay inside scope across sessions, or (c) follow team standards without re-briefing, a single `CLAUDE.md` hits its ceiling. Then one sentence on the multi-agent drift problem (parallel instruction sets diverge). **Strip "powerful out of the box"** — that's a banned phrase.
3. **`## What Bonsai gives you`** (rename from "What Bonsai Does") — rewrite. Open with one sentence: Bonsai generates a workspace under `station/` plus per-agent workspaces, plus Claude Code hook wiring that enforces them. Then a tightened version of the 6-layer diagram (keep the ASCII block exactly as it currently is — the diagram is fine, the surrounding copy is not). Replace the "You pick the components..." paragraph with one short paragraph naming the actual mechanisms: **session-start context injection, PreToolUse scope guards, lock-aware file tracking, version-controlled protocols.** Strip "One binary. No runtime. Works with any project." — banned phrase.
4. **`## Who it's for`** — keep the `<CardGrid>` + 3 `<Card>` blocks. Tighten card body text:
   - "Solo developers" — drop "behave consistently" (taste-negative, see memory). Suggested: `Want their agent to remember context across sessions, follow project standards, and stay inside scope without re-briefing.`
   - "Teams" — preserve the multi-agent shared-source-of-truth idea but drop "on the same page" cliché. Suggested: `Coordinating multiple agents (backend, frontend, devops) across one codebase — with shared plans, status, and standards in version control.`
   - "Anyone tired of copy-pasting" — keep the spirit. Drop "battle-tested" (banned register). Suggested: `Stop copying agent instructions between projects. Bonsai generates a workspace from a catalog you pick from.`

**Hard constraints for this file:**
- Preserve all `import { Card, CardGrid } from '@astrojs/starlight/components';` lines.
- Preserve all `<Card>`, `<CardGrid>` tags + their `title=`/`icon=` attributes — only edit body text inside the tags.
- Do not introduce `<` autolinks. Use `[label](url)` for any URL.
- Do not add new sections beyond the 3 listed above.

### Step 3 — `website/astro.config.mjs:28` LLM description rewrite

**Goal:** Strip the AI-smell "structured instruction files... so AI agents work like teammates, not tools" phrasing from the `starlightLlmsTxt` plugin's `description` field.

**The change:**

- Locate the `description: ` line inside `starlightLlmsTxt({...})` (currently line 28, beginning `description: \`Bonsai is a CLI tool for scaffolding...\``).
- Replace the value (preserving the surrounding template literal backticks + the `description: ` key + the trailing comma) with a mechanism-led description. Suggested:
  ```
  description: `Bonsai is a CLI tool that generates Claude Code agent workspaces. It writes a workspace under station/ plus per-agent workspaces, wires Claude Code hooks (SessionStart, PreToolUse, Stop) to enforce scope and inject context, and tracks generated files with content hashes so user edits are never silently overwritten.`,
  ```
- Do not modify any other field in `astro.config.mjs` — not the `details:` block, not the `customSets:` array, not the `sidebar:` definition, not the `social:` array, not the integrations list.
- Do not modify the file's import statements, `defineConfig` call, or `site:` / `base:` fields.

**Hard constraints for this file:**
- Single-field edit. Keep template literal backticks. Keep trailing comma.
- No banned phrases ("teammates, not tools", "structured instruction files", etc.) in the new description.

## Out of scope

- `website/src/content/docs/index.mdx` — already done in PR #65, do not touch.
- `website/src/content/docs/quickstart*`, `commands/*`, `catalog/*`, `reference/*`, `guides/*` — separate Backlog items.
- Any Go code (`cmd/`, `internal/`, `embed.go`, `main.go`) — zero changes expected.
- `catalog/` directory — not a doc page, separate concern.
- `station/` — workspace, not a public-doc page.
- README.md — already canonical, do not touch.
- `CHANGELOG.md` — copy-only PR, no version-bump-worthy change.
- Any new files — this is a 3-file edit only.

## Security

> [!warning]
> Refer to `Playbook/Standards/SecurityStandards.md` for all security requirements.

This is a docs-copy PR. No code changes, no dependency changes, no new files. Security surface is limited to:
- **No new dependencies** — do not add packages to `website/package.json`.
- **No external links to untrusted domains** — any new URL must be to `github.com/LastStep/Bonsai`, `laststep.github.io/Bonsai`, or already-cited references.
- **No `<https://...>` autolinks in MDX** — parses as JSX, breaks build (memory: "MDX autolink gotcha").

## Verification

- [ ] `cd website && npm run build` exits 0 with no errors and no broken-link warnings (the `starlightLinksValidator` plugin is enabled — if it fails, fix the broken link before reporting done).
- [ ] `grep -nE 'powerful out of the box|by design|teammates, not tools|structured language|structured instruction files|One binary\. No runtime' website/src/content/docs/concepts/how-bonsai-works.mdx website/src/content/docs/why-bonsai.mdx website/astro.config.mjs` returns no matches.
- [ ] `grep -cE '^<' website/src/content/docs/concepts/how-bonsai-works.mdx website/src/content/docs/why-bonsai.mdx` shows no MDX-incompatible `<https://` autolinks (only legit JSX components like `<Aside>`, `<FileTree>`, `<Card>`, `<CardGrid>`).
- [ ] Diff scope: only the 3 files listed above. `git diff --stat main...HEAD` should show exactly these 3 paths.
- [ ] No new files created, no files deleted.
- [ ] Frontmatter `title:` lines unchanged in the two `.mdx` files (only `description:` changes).
- [ ] All preserved JSX components (`<Aside>`, `<FileTree>`, `<CardGrid>`, `<Card>`) still parse — check by reading the rendered HTML in `website/dist/` after build, or by running `npm run dev` and visiting `/Bonsai/concepts/how-bonsai-works/` + `/Bonsai/why-bonsai/`.
