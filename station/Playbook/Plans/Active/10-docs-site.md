# Plan 10 — Documentation Site (Starlight)

**Tier:** 2 (Feature)
**Status:** Draft
**Source:** Roadmap Phase 1 "Usage instructions" + Backlog Group A + user request
**Phases:** A (scaffold & migrate), B (fill gaps), C (catalog generation & LLM layer), D (deploy & CI)

---

## Goal

Ship a public documentation site for Bonsai using Astro Starlight — the primary learning resource for both humans and AI agents. The site serves as the public gateway to the project, covering installation, concepts, CLI reference, catalog browsing, guides, and configuration reference. It must also generate `llms.txt` files so AI agents can consume Bonsai's documentation efficiently.

### Success Criteria

- Docs site is live at a public URL (GitHub Pages or Cloudflare Pages)
- All existing documentation (README, HANDBOOK, guides) is migrated and restructured
- Every CLI command has a dedicated reference page with flags, examples, and output
- Core concepts (agents, abilities, sensors, routines, scaffolding, workspaces) each have a page
- The full catalog is browsable in the docs (agent types, skills, workflows, protocols, sensors, routines)
- A glossary page defines all Bonsai-specific terms
- Configuration reference covers `.bonsai.yaml`, `.bonsai-lock.yaml`, `meta.yaml`, and `agent.yaml` schemas
- `llms.txt` and `llms-full.txt` are auto-generated at build time via the `starlight-llms-txt` plugin
- CI auto-deploys on merge to main

### Relationship to Backlog

This plan **supersedes Backlog Group A items 1-3** (quickstart guide, concepts guide, CLI usage guide). Those items are absorbed into Phases A-B — the docs site delivers them as web pages rather than standalone markdown files. The 4th Group A item (`bonsai guide` multi-topic command) remains in the backlog as a standalone item — it becomes a CLI pointer to the docs site rather than a standalone renderer. Phase D includes updating the existing `bonsai guide` output to reference the live docs URL.

---

## Context

Bonsai is approaching public release. The existing documentation is high-quality but scattered across the repo root (README, HANDBOOK, CONTRIBUTING) and `docs/` (working-with-agents, triggers, custom-files). There is no unified, searchable, browsable documentation site. Users and AI agents have no single entry point for learning Bonsai.

Starlight (Astro's documentation framework) is the chosen platform. Key reasons:
- Static site generation (zero JS for content pages, sub-second loads)
- Built-in Pagefind search (zero infrastructure)
- MDX support with rich components (Tabs, Steps, FileTree, Cards, Code blocks)
- `starlight-llms-txt` plugin for AI-consumable docs
- Free hosting on GitHub Pages / Cloudflare Pages
- Dark mode, responsive, accessible out of the box

---

## Phase A — Scaffold & Migrate

Scaffold the Starlight project and migrate all existing content into the new structure. No new content written — only restructuring.

### A1. Scaffold Starlight project (`website/`)

Create `website/` directory in repo root (not `docs/` — that already contains existing guides that we'll migrate from).

```
website/
├── astro.config.mjs          ← Starlight config (sidebar, social, plugins)
├── package.json               ← dependencies
├── tsconfig.json              ← TypeScript config (Astro default)
├── public/
│   ├── favicon.svg            ← Bonsai favicon
│   └── og-image.png           ← Default Open Graph image
└── src/
    ├── content.config.ts      ← Content collection schema
    ├── content/
    │   └── docs/              ← All documentation pages (MDX/MD)
    ├── assets/                ← Images, diagrams (optimized by Astro)
    └── styles/
        └── custom.css         ← Brand colors, font overrides
```

**Dependencies:**
```json
{
  "@astrojs/starlight": "latest",
  "astro": "latest",
  "starlight-llms-txt": "latest",
  "starlight-links-validator": "latest",
  "js-yaml": "latest"
}
```

**Note:** Commit `package-lock.json` alongside `package.json` — the CI workflow uses `npm ci` which requires a lock file.

**`astro.config.mjs` configuration:**
- `site`: `https://laststep.github.io/Bonsai` (or custom domain if available)
- `base`: `'/Bonsai'` (required for GitHub Pages subdirectory hosting — remove if using a custom domain)
- `title`: `Bonsai`
- `description`: `A structured language for working with AI agents`
- `logo`: Bonsai icon (light/dark variants)
- `social`: array format — `[{ icon: 'github', label: 'GitHub', href: 'https://github.com/LastStep/Bonsai' }]`
- `editLink.baseUrl`: `https://github.com/LastStep/Bonsai/edit/main/website/`
- `lastUpdated`: `true`
- `customCss`: `['./src/styles/custom.css']`
- `plugins`: `[starlightLlmsTxt(), starlightLinksValidator()]`
- `sidebar`: manually configured (see A2)

**`content.config.ts`** is optional — Starlight manages its own content collection schema automatically. Only needed if extending the frontmatter schema with custom fields.

### A2. Configure sidebar structure

```js
sidebar: [
  {
    label: 'Start Here',
    items: [
      { slug: 'getting-started' },
      { slug: 'installation' },
      { slug: 'why-bonsai' },
    ],
  },
  {
    label: 'Core Concepts',
    items: [
      { slug: 'concepts/how-bonsai-works' },
      { slug: 'concepts/agents' },
      { slug: 'concepts/abilities' },
      { slug: 'concepts/sensors' },
      { slug: 'concepts/routines' },
      { slug: 'concepts/scaffolding' },
      { slug: 'concepts/workspaces' },
    ],
  },
  {
    label: 'Commands',
    items: [
      { slug: 'commands/init' },
      { slug: 'commands/add' },
      { slug: 'commands/remove' },
      { slug: 'commands/list' },
      { slug: 'commands/catalog' },
      { slug: 'commands/update' },
      { slug: 'commands/guide' },
    ],
  },
  {
    label: 'Guides',
    items: [
      { slug: 'guides/your-first-workspace' },
      { slug: 'guides/working-with-agents' },
      { slug: 'guides/triggers-and-activation' },
      { slug: 'guides/customizing-abilities' },
      { slug: 'guides/creating-custom-skills' },
      { slug: 'guides/creating-custom-sensors' },
      { slug: 'guides/creating-custom-routines' },
      { slug: 'guides/dogfooding' },
    ],
  },
  {
    label: 'Catalog',
    collapsed: true,
    items: [
      { slug: 'catalog/overview' },
      { slug: 'catalog/agent-types' },
      { slug: 'catalog/skills' },
      { slug: 'catalog/workflows' },
      { slug: 'catalog/protocols' },
      { slug: 'catalog/sensors' },
      { slug: 'catalog/routines' },
    ],
  },
  {
    label: 'Reference',
    collapsed: true,
    items: [
      { slug: 'reference/configuration' },
      { slug: 'reference/lock-file' },
      { slug: 'reference/template-variables' },
      { slug: 'reference/meta-yaml-schema' },
      { slug: 'reference/agent-yaml-schema' },
      { slug: 'reference/glossary' },
    ],
  },
],
```

### A3. Migrate existing content

Map existing files to Starlight pages. Content is restructured (split, reorganized) but **not rewritten** in this phase — tone and substance stay the same.

| Source | Destination | Transform |
|--------|-------------|-----------|
| `README.md` "The Problem" + "What Bonsai Does" | `src/content/docs/why-bonsai.mdx` | Extract, add frontmatter, add hero component |
| `README.md` "Install" section | `src/content/docs/installation.mdx` | Extract, expand with OS-specific tabs |
| `README.md` "Quick Start" section | `src/content/docs/getting-started.mdx` | Extract, expand into Steps component |
| `HANDBOOK.md` "The Mental Model" section | `src/content/docs/concepts/how-bonsai-works.mdx` | Extract, add FileTree component for structure |
| `HANDBOOK.md` agent sections | `src/content/docs/concepts/agents.mdx` | Extract, add Card components per agent type |
| `HANDBOOK.md` "Understanding Sensors" | `src/content/docs/concepts/sensors.mdx` | Extract |
| `HANDBOOK.md` "Understanding Routines" | `src/content/docs/concepts/routines.mdx` | Extract |
| `HANDBOOK.md` "How the Pieces Fit Together" | `src/content/docs/concepts/workspaces.mdx` | Extract |
| `docs/working-with-agents.md` | `src/content/docs/guides/working-with-agents.mdx` | Migrate, add Aside callouts |
| `docs/triggers.md` | `src/content/docs/guides/triggers-and-activation.mdx` | Migrate |
| `docs/custom-files.md` | Split into 3 pages under `guides/creating-custom-*` | Split by type (skills, sensors, routines) |
| `docs/assets/graph-view.png` | `src/assets/graph-view.png` | Move to Astro-optimized assets |
| `CONTRIBUTING.md` | `src/content/docs/contributing.mdx` (linked from footer, not sidebar) | Migrate |
| `SECURITY.md` | Footer link or `src/content/docs/security.mdx` | Migrate or link |
| `CODE_OF_CONDUCT.md` | Footer link or `src/content/docs/code-of-conduct.mdx` | Migrate or link |

### A4. Create stub pages for new content

Create placeholder pages with frontmatter and a "Coming soon" note for content that will be written in Phase B:

- `src/content/docs/concepts/abilities.mdx` — Skills vs. workflows vs. protocols decision matrix
- `src/content/docs/concepts/scaffolding.mdx` — Playbook, Logs, Reports, INDEX
- `src/content/docs/commands/*.mdx` — 7 command pages (init, add, remove, list, catalog, update, guide)
- `src/content/docs/guides/your-first-workspace.mdx` — End-to-end tutorial
- `src/content/docs/guides/customizing-abilities.mdx` — Edit installed content
- `src/content/docs/guides/dogfooding.mdx` — How Bonsai manages itself
- `src/content/docs/catalog/overview.mdx` — Catalog landing page
- `src/content/docs/reference/*.mdx` — 6 reference pages
- `src/content/docs/faq.mdx` — Frequently asked questions
- `src/content/docs/troubleshooting.mdx` — Common issues and recovery steps

Each stub follows this format:

```mdx
---
title: Page Title
description: One-line description for SEO and llms.txt
---

import { Aside } from '@astrojs/starlight/components';

<Aside type="caution">This page is under construction.</Aside>
```

### A5. Create landing page

`src/content/docs/index.mdx` with `template: splash` (no sidebar):

- Hero section: title, tagline, two CTAs ("Get Started", "Browse Catalog")
- Feature cards: 6-layer architecture, agent types, TUI screenshots
- Links to key sections: Concepts, Commands, Guides

### A6. Brand styling (`src/styles/custom.css`)

- Accent color palette (Bonsai green tones)
- Custom font if desired (or Starlight defaults)
- Content width adjustment if needed
- No heavy customization — stay close to Starlight defaults for maintainability

### A7. Add `.gitignore` entries

Add to repo root `.gitignore`:
```
website/node_modules/
website/dist/
website/.astro/
```

### Phase A Verification

- [ ] `cd website && npm install && npm run dev` — site runs locally with no errors
- [ ] All migrated content renders correctly (no broken markdown, no missing images)
- [ ] Sidebar navigation matches the structure defined in A2
- [ ] Landing page renders with hero and feature cards
- [ ] `starlight-links-validator` reports no broken internal links
- [ ] Stub pages are clearly marked as "under construction"
- [ ] Original `docs/` files are untouched (migration is copy, not move — cleanup later)

---

## Phase B — Fill Content Gaps

Write all new content. Each page follows Starlight conventions: frontmatter with `title` and `description`, Starlight components where appropriate.

### B1. Core Concepts pages (5 new pages)

**`concepts/abilities.mdx`** — The decision matrix page. This is the most important concept page for new users.
- What are abilities? (skills, workflows, protocols — the three types)
- Decision matrix: "When do I use a skill vs. a workflow vs. a protocol?"
- Table comparing: purpose, activation method, loaded when, example
- How abilities are installed (`bonsai add`), listed (`bonsai list`), browsed (`bonsai catalog`)
- Link to each catalog page for browsing

**`concepts/scaffolding.mdx`** — Project infrastructure.
- What scaffolding is: INDEX, Playbook, Logs, Reports
- What each piece does and why it exists
- Required vs. optional (Reports is optional)
- FileTree component showing the generated structure
- How scaffolding relates to routines (backlog-hygiene needs Backlog.md, etc.)

**`concepts/agents.mdx`** — Complete rewrite (not just HANDBOOK extraction).
- The 6 agent types with role descriptions
- Card component for each agent type with icon, description, default abilities
- "Which agent type do I need?" decision guide
- The Tech Lead orchestration model (plans, dispatches, reviews)
- Multi-agent workspace patterns

**`concepts/sensors.mdx`** — Expand from HANDBOOK content.
- How sensors work (hook events, matchers, scripts)
- Available events: SessionStart, PreToolUse, PostToolUse, Stop, UserPromptSubmit, SubagentStop
- Required sensors (context-guard, scope-guard-files, session-context, status-bar)
- Table of all 12 catalog sensors with event, matcher, description

**`concepts/routines.mdx`** — Expand from HANDBOOK content.
- How routines work (frequency, dashboard, opt-in execution)
- The routine lifecycle: check → approve → execute → log → update dashboard
- Table of all 8 catalog routines with frequency, description
- Note: Tech Lead only (other agents don't have routines by default)

### B2. Command reference pages (7 pages)

Each command page follows this template:

```
## Synopsis
## Description (what it does, when to use it)
## Interactive Flow (what the TUI prompts for, in order)
## Flags (table: flag, type, default, description)
## Examples (2-3 common use cases with code blocks)
## What Gets Generated (FileTree showing output for init/add)
## See Also (links to related commands and concepts)
```

Pages: `commands/init.mdx`, `commands/add.mdx`, `commands/remove.mdx`, `commands/list.mdx`, `commands/catalog.mdx`, `commands/update.mdx`, `commands/guide.mdx`

Source of truth for flags and descriptions: `cmd/*.go` Cobra command definitions.

### B3. Guide pages (4 new pages)

**`guides/your-first-workspace.mdx`** — End-to-end tutorial.
- Prerequisites (Go, Claude Code)
- Install Bonsai
- `bonsai init` walkthrough (each prompt explained with Steps component)
- What got generated (FileTree)
- Adding your first code agent (`bonsai add`)
- Starting your first session with the Tech Lead
- What to try next

**`guides/customizing-abilities.mdx`** — How to modify installed content.
- Editing installed skills/workflows directly (lock file detects changes)
- What happens on `bonsai update` (conflict resolution flow)
- When to customize vs. when to create a new custom ability
- Link to creating custom abilities guides

**`guides/creating-custom-sensors.mdx`** — Split from existing `custom-files.md`.
- Sensor-specific frontmatter (`event`, `matcher`)
- Script template format (`.sh.tmpl`)
- Available template variables
- Example: creating a custom pre-commit check
- Testing sensors locally

**`guides/creating-custom-routines.mdx`** — Split from existing `custom-files.md`.
- Routine-specific frontmatter (`frequency`)
- Content template format (`.md.tmpl`)
- The routine dashboard and how it tracks execution
- Example: creating a weekly documentation check

### B4. Catalog pages (7 pages)

**`catalog/overview.mdx`** — Landing page with CardGrid linking to each category.

**`catalog/agent-types.mdx`** — All 6 agents with:
- Card per agent: name, display name, description
- Default abilities table per agent
- "Best for" guidance

**`catalog/skills.mdx`** — All 17 skills with:
- Table: name, description, compatible agents, path triggers (if any)
- Grouped by domain (coding, infrastructure, management, design)

**`catalog/workflows.mdx`** — All 10 workflows with:
- Table: name, description, compatible agents, slash command (if any)
- Note which are curated for slash-command invocation

**`catalog/protocols.mdx`** — All 4 protocols with:
- Table: name, description, required status
- Note that all are currently required for all agents

**`catalog/sensors.mdx`** — All 12 sensors with:
- Table: name, description, event, matcher, compatible agents, required status
- Grouped by function (guards, context injection, review, monitoring)

**`catalog/routines.mdx`** — All 8 routines with:
- Table: name, description, frequency, compatible agents
- Note: currently Tech Lead only

### B5. Reference pages (6 pages)

**`reference/configuration.mdx`** — `.bonsai.yaml` full schema.
- Every field: type, default, description, example
- Example configs for common setups (solo dev, team, multi-agent)

**`reference/lock-file.mdx`** — `.bonsai-lock.yaml` format.
- Purpose: content hashing, conflict detection
- Schema: file path, hash, source
- How conflicts are detected and resolved

**`reference/template-variables.mdx`** — All Go template variables.
- `{{ .ProjectName }}`, `{{ .ProjectDescription }}`, `{{ .Routines }}`, etc.
- Full TemplateContext struct fields with descriptions
- Custom template functions (`{{ title .AgentType }}`)
- Examples of each variable in use

**`reference/meta-yaml-schema.mdx`** — Field reference for ability metadata.
- Base fields (all types): `name`, `description`, `display_name`, `agents`, `required`
- Sensor additions: `event`, `matcher`
- Routine additions: `frequency`
- Trigger additions: `triggers.scenarios`, `triggers.examples`, `triggers.paths`
- Tabs component showing schema per ability type

**`reference/agent-yaml-schema.mdx`** — Agent definition format.
- Fields: `name`, `display_name`, `description`, `defaults`
- Defaults sub-fields: skills, workflows, protocols, sensors, routines
- Core directory structure (`core/identity.md.tmpl`, `core/memory.md.tmpl`, etc.)

**`reference/glossary.mdx`** — All Bonsai-specific terms.
- Ability, Agent, Catalog, Core, Dashboard, Dispatch, Identity, Lock File, Matcher, Protocol, Routine, Scaffolding, Sensor, Skill, Station, Workspace, Workflow
- Each entry: 1-2 sentence definition + link to relevant concept/reference page
- Alphabetical order

### Phase B Verification

- [ ] All stub pages replaced with real content
- [ ] Every page has `title` and `description` frontmatter
- [ ] `starlight-links-validator` reports no broken links
- [ ] Catalog pages are accurate against current `catalog/*/meta.yaml` files
- [ ] Command pages match current `cmd/*.go` Cobra definitions
- [ ] Glossary covers all terms used in concepts and guides
- [ ] Configuration reference matches current `internal/config/config.go` struct definitions

---

## Phase C — Catalog Auto-Generation & LLM Layer

### C1. Catalog page generation script

Create `website/scripts/generate-catalog.mjs` — a Node.js script that reads `catalog/*/meta.yaml` files and generates/updates the catalog pages in `src/content/docs/catalog/`.

**Behavior:**
1. Walk `catalog/agents/`, `catalog/skills/`, `catalog/workflows/`, `catalog/protocols/`, `catalog/sensors/`, `catalog/routines/`
2. Parse each `meta.yaml` (and `agent.yaml` for agents)
3. Generate structured MDX content for each catalog page:
   - Skills page: table with name, description, agents, triggers.paths, triggers.scenarios
   - Workflows page: table with name, description, agents, slash command status
   - Sensors page: table with name, description, event, matcher, agents, required
   - Routines page: table with name, description, frequency, agents
   - Agent types page: card per agent with defaults listed
4. Write to catalog page files, preserving any hand-written intro/outro sections
5. Output a summary of changes
6. Also emit `website/public/catalog.json` — a structured JSON dump of all catalog metadata (agent types, skills, workflows, protocols, sensors, routines with full fields). This gives AI agents a machine-readable alternative to the prose catalog pages for structured queries like "what sensors are compatible with the backend agent?"

**Integration:** Add `"generate:catalog": "node scripts/generate-catalog.mjs"` to `package.json`. Run before `astro build` in CI.

**Why a script, not a content collection loader:** Starlight's content collection expects files in `src/content/docs/`. A script that writes MDX files is simpler than a custom loader, works with Starlight's existing routing, and the output is inspectable in git. The catalog changes rarely — a manual `npm run generate:catalog` before committing is fine.

### C2. Configure `starlight-llms-txt` plugin

The plugin auto-generates `llms.txt` and `llms-full.txt` at build time from Starlight content. Configuration in `astro.config.mjs`:

```js
import starlightLlmsTxt from 'starlight-llms-txt';

// In starlight() config:
plugins: [
  starlightLlmsTxt(),
  // ... other plugins
],
```

The plugin uses page titles and descriptions from frontmatter to build the index. This is why every page must have a `description` field (enforced in Phase B).

### C3. Custom LLM topic bundles

Beyond the auto-generated `llms.txt`, create curated topic-bundle files for targeted LLM consumption. These are static files in `website/public/_llms/`:

| File | Content | Purpose |
|------|---------|---------|
| `concepts.txt` | All concept pages concatenated | LLM asking "what is Bonsai?" |
| `commands.txt` | All command reference pages concatenated | LLM asking "how do I use bonsai add?" |
| `catalog.txt` | Full catalog listing with descriptions and metadata | LLM asking "what skills are available?" |
| `configuration.txt` | Config reference + template variables + schemas | LLM asking "how do I configure .bonsai.yaml?" |

Each file starts with a descriptive blockquote following the `llms.txt` spec:
```markdown
> This is the Bonsai documentation for [topic]. Bonsai is a CLI tool
> for scaffolding Claude Code agent workspaces.
```

**Generation:** Create `website/scripts/generate-llm-bundles.mjs` that concatenates rendered page content into these bundle files. Add `"generate:llm-bundles": "node scripts/generate-llm-bundles.mjs"` to `package.json`. Run alongside `generate:catalog` in the build pipeline.

### C4. Enhance root `llms.txt` with curated index

After the plugin generates the base `llms.txt`, post-process it (or use the plugin's customization options) to ensure the root file includes:

```markdown
# Bonsai

> Bonsai is a CLI tool for scaffolding Claude Code agent workspaces. It generates
> structured instruction files — identity, memory, protocols, skills, workflows,
> sensors, and routines — so AI agents work like teammates, not tools.

- Install: `go install github.com/LastStep/Bonsai@latest` or `brew install LastStep/tap/bonsai`
- 6 agent types: tech-lead, backend, frontend, fullstack, devops, security
- Abilities are modular: skills (reference), workflows (multi-step), protocols (rules), sensors (hooks), routines (periodic)

## Core Documentation

- [Concepts](/_llms/concepts.txt): How Bonsai works — agents, abilities, sensors, routines, scaffolding, workspaces
- [Commands](/_llms/commands.txt): CLI reference for all 7 commands with flags and examples
- [Configuration](/_llms/configuration.txt): .bonsai.yaml, .bonsai-lock.yaml, meta.yaml, and agent.yaml schemas

## Catalog

- [Full Catalog](/_llms/catalog.txt): All 6 agent types, 17 skills, 10 workflows, 4 protocols, 12 sensors, 8 routines with descriptions and compatibility

## Optional

- [Full Documentation](/llms-full.txt): Complete documentation in one file
```

### Phase C Verification

- [ ] `npm run generate:catalog` produces accurate catalog pages matching current `meta.yaml` files
- [ ] Adding a new skill `meta.yaml` and re-running the script adds it to the skills page
- [ ] `npm run build` generates `llms.txt` and `llms-full.txt` in the output
- [ ] Topic bundle files exist in `dist/_llms/` and contain correct content
- [ ] `llms.txt` root file has the curated Bonsai-specific index
- [ ] Each `_llms/*.txt` file starts with a descriptive blockquote

---

## Phase D — Deploy & CI

### D1. GitHub Actions workflow

Create `.github/workflows/docs.yml`:

```yaml
name: Deploy Docs
on:
  push:
    branches: [main]
    paths:
      - 'website/**'
      - 'catalog/**'
      - 'docs/**'
      - 'README.md'
      - 'HANDBOOK.md'
  workflow_dispatch:

permissions:
  contents: read
  pages: write
  id-token: write

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: 22
          cache: npm
          cache-dependency-path: website/package-lock.json
      - run: npm ci
        working-directory: website
      - run: npm run generate:catalog
        working-directory: website
      - run: npm run generate:llm-bundles
        working-directory: website
      - run: npm run build
        working-directory: website
      - uses: actions/upload-pages-artifact@v3
        with:
          path: website/dist

  deploy:
    needs: build
    runs-on: ubuntu-latest
    environment:
      name: github-pages
      url: ${{ steps.deployment.outputs.page_url }}
    steps:
      - id: deployment
        uses: actions/deploy-pages@v4
```

### D2. Update README

- Replace the `docs/` guide links with links to the live docs site
- Add a "Documentation" badge linking to the site
- Keep the inline Quick Start section (it serves as a fast on-ramp in the repo itself)
- Add a note: "For comprehensive documentation, visit [docs site URL]"

### D3. Update `docs/` directory

After the site is live:
- Add a `docs/README.md` that redirects to the live site ("This directory contains source content. Visit [URL] for the documentation site.")
- Keep `docs/assets/` (referenced by README)
- The original `docs/*.md` files can be kept as source or removed — the Starlight site is now authoritative

### D4. Update `bonsai guide` output

Update `cmd/guide.go` to include a reference to the docs site URL in its output. The existing `bonsai guide` command renders `custom-files.md` — add a header line or footer note like: "For comprehensive documentation, visit [docs site URL]". This is a minimal change (not the full multi-topic expansion from Backlog Group A — that remains a separate item).

### D5. Update HANDBOOK.md

Add a note at the top redirecting to the docs site for the latest version:
```markdown
> **Note:** The latest version of this handbook is at [docs site URL]/concepts/how-bonsai-works.
```

### Phase D Verification

- [ ] `git push` to main triggers the docs workflow
- [ ] Site is accessible at the configured URL
- [ ] Search (Pagefind) works — can find "bonsai init", "sensor", "routine"
- [ ] `llms.txt` is accessible at `{site}/llms.txt`
- [ ] All internal links resolve (no 404s)
- [ ] README badge links to live site
- [ ] Dark mode works
- [ ] Mobile layout is functional

---

## Security

> [!warning]
> Refer to SecurityStandards.md for all security requirements.

- No secrets in the docs site (no API keys, no auth tokens)
- GitHub Pages deployment uses OIDC tokens (no stored secrets needed)
- `starlight-links-validator` catches broken links before deploy
- No user input or dynamic content — static HTML only
- Generated `llms.txt` files contain only public documentation content

---

## Dependencies

- Node.js 22+ (for Astro build — CI only, not needed locally for Go development)
- Astro + Starlight (pinned versions in `package.json`)
- `starlight-llms-txt` plugin
- `starlight-links-validator` plugin
- GitHub Pages (or Cloudflare Pages) for hosting
- No changes to Go source code

---

## Dispatch

| Phase | Agent | Isolation | Notes |
|-------|-------|-----------|-------|
| A | general-purpose | worktree | Scaffold + content migration — no Go changes |
| B | general-purpose | worktree | Content writing — pure MDX, no Go changes |
| C | general-purpose | worktree | Scripts + LLM layer — Node.js scripts, no Go changes |
| D | general-purpose | worktree | CI workflow + README updates |

**Note:** All phases are pure documentation/web work — no Go application code changes (except the minor `cmd/guide.go` update in D4). Each phase can be a single PR.

**Parallelization note:** Phase B4 (catalog pages) and Phase C1 (catalog generation script) have a dependency — C1 auto-generates the pages that B4 writes by hand. If C1 is built first, B4 becomes unnecessary (the script generates the pages). Consider building C1 early and letting it handle B4 automatically. The rest of Phase B and Phase C are independent and can proceed in parallel.

---

## Known Limitations

1. **No doc versioning.** Starlight does not have built-in versioning. The community plugin `starlight-versions` exists but is not official. For now, docs deploy from `main` and reflect the latest version. A version indicator on the landing page will note which Bonsai release the docs cover. If breaking changes ship in a future release, evaluate versioning at that point.

---

## Open Questions

1. **Custom domain vs. GitHub Pages default?** — `laststep.github.io/Bonsai` vs. `docs.bonsai.dev` or similar. Custom domain is nicer but requires DNS setup and ongoing cost. If using a custom domain, remove `base: '/Bonsai'` from the Astro config.
2. **VHS terminal recordings?** — Charm's VHS tool can generate GIFs from `.tape` scripts. Great for showing TUI flows but adds build complexity. Could be a follow-up task rather than blocking launch.
3. **Catalog generation: build-time vs. manual?** — Script could run in CI (always fresh) or manually (inspectable diffs in git). Plan assumes manual with CI as a check. Revisit if catalog changes frequently.
4. **What happens to `docs/` originals?** — Keep as source alongside Starlight, or remove after migration? Plan assumes keep with a redirect README, but open to removing.
