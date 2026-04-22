# Plan 25 — README revamp

**Tier:** 1
**Status:** Complete
**Agent:** tech-lead
**Shipped:** 2026-04-22 via PR #61 squash `d6785e7`

## Goal

Replace the top-level `README.md` with an audience-first structure: logo placeholder, outcome-focused tagline, a new "Who Bonsai is for" section with mechanism-based feature bullets and a `station/` tree snippet, a merged "How it works" section that absorbs the old layer-stack and catalog content, and a lightweight footer that carries nav and attribution.

## Context

Current README leads with abstract framing ("a structured language for working with AI agents") and buries the audience pitch. Independent research agent flagged three high-impact issues: (a) the "who is this for" question isn't answered up front, (b) feature framing anthropomorphises the agent ("takes ownership") instead of showing mechanism (hooks, protocol files, checklists), and (c) a `tree station/` snippet is the single most persuasive visual and is missing. This plan lands all three plus the reorder, kills two duplicated sections (Documentation, Contributing body), and reserves a logo slot for art still in progress.

## Steps

1. **Reserve logo slot** at the top of `README.md` using an HTML comment (`<!-- Logo placeholder -->`). No empty `<div>` with fixed size — it renders as a broken gap on GitHub.
2. **Replace tagline** from "A structured language for working with AI agents." to **"A workspace for your coding agent."** Keep the center-aligned `<div>`, badges, and nav link row directly below.
3. **Add `status-early-stage` badge** to the badge row: `[![Status: early-stage](https://img.shields.io/badge/status-early--stage-orange.svg)](https://github.com/LastStep/Bonsai/issues)`.
4. **Remove** the old hero `docs/assets/graph-view.png` from the top of the file. It returns later as a supporting visual inside "How it works".
5. **Delete** the existing "Why Bonsai" section (lines ~29–47 of current file). Its layer-stack diagram is preserved — lifted into the new "How it works" section.
6. **Add new section `## Who Bonsai is for`** with:
   - One-line audience hook (solo devs + small teams, give coding agent real responsibility).
   - Two-sentence problem framing (single CLAUDE.md hits its ceiling fast).
   - Five mechanism bullets (bold lead word + concrete file reference):
     - **Every session starts from the same context.** `SessionStart` hook injects identity, memory, active plans, health warnings before first reply.
     - **The project is navigable, not just searchable.** Indexed code, cross-linked plans, Obsidian-compatible markdown.
     - **Rules live in files, not prompts.** Protocols (`security.md`, `scope-boundaries.md`, `memory.md`) version-controlled. Sensors fire on `PreToolUse`/`Stop` to block at the tool call.
     - **Plans before it acts.** Writes a plan to `Playbook/Plans/Active/NN-*.md` before any dispatch. Reviews run from `agent/Skills/review-checklist.md`.
     - **Everything is auditable.** Decisions → `Logs/KeyDecisionLog.md`. Out-of-scope findings → `Playbook/Backlog.md`. Agent reports → `Reports/`. `git log` is the audit trail.
   - **"Not just CLAUDE.md with extra steps" rebuttal** — one paragraph framing Bonsai as workspace vs single file.
   - **`station/` tree snippet** (code block, trimmed to essential dirs with inline annotation).
7. **Keep** `## Install`, `## Quick Start`, `## See it in action` (rename "See It In Action" → sentence case) unchanged in content. Tweak "say 'hi, get started'" copy to show what the agent does: reads identity, checks memory, scans active plans, reports status.
8. **Add merged section `## How it works`** containing:
   - Layer-stack diagram (preserved verbatim from old "Why Bonsai").
   - One paragraph on the "pick components at init/add time → Bonsai generates cross-linked workspace" mechanism.
   - Graph-view PNG (moved here from hero — `docs/assets/graph-view.png`, width 700).
   - `### Six agent types` table (preserved from old "What's Inside").
   - `### The catalog` bullet list with counts (preserved, copy-edited).
   - `### Extensible` paragraph (preserved, trimmed of "by design" boilerplate).
9. **Keep** `## Commands` table unchanged.
10. **Delete** `## Documentation` as its own section. Fold its links into the footer nav row.
11. **Delete** `## Contributing` prose. Keep the `CONTRIBUTING.md` link in the footer nav row and at the top nav.
12. **Rewrite footer** to a single nav row: Documentation · Catalog · Contributing · Releases · MIT License. Keep the "Built with Cobra / Huh / LipGloss / BubbleTea. Developed with Claude Code." attribution line.
13. **Cut stylistic AI-readme smells** flagged by the reviewer: "by design", "powerful out of the box", "One binary. No runtime. Works with any project.", "teammates, not tools", "structured language" framing everywhere.
14. **Preserve** asset paths: `assets/demos/init.gif` (hero demo), `docs/assets/graph-view.png` (moved, not removed).

## Security

> [!warning]
> Refer to `Playbook/Standards/SecurityStandards.md` for all security requirements. No code changes in this plan — documentation only. No secrets, no credentials, no example API keys introduced. All links use HTTPS to the existing project hosts (`github.com/LastStep/Bonsai`, `laststep.github.io/Bonsai`). Shield badge URLs go to `img.shields.io` (same host already in use).

## Verification

- [ ] `README.md` renders on GitHub without broken images or empty gaps (visual inspection after push).
- [ ] All markdown links resolve (`[text](target)` targets: `CONTRIBUTING.md`, `LICENSE`, `laststep.github.io/Bonsai/*`, `github.com/LastStep/Bonsai/*`, `img.shields.io/*`, `assets/demos/init.gif`, `docs/assets/graph-view.png`).
- [ ] Logo slot is an HTML comment (not a visible empty `<div>`).
- [ ] `station/` tree snippet matches what `bonsai init` actually generates (dirs: `Playbook/`, `Logs/`, `Reports/`, `agent/Core/`, `agent/Protocols/`, `agent/Skills/`, `agent/Workflows/`, `agent/Sensors/`, `agent/Routines/`).
- [ ] Badge row parses: 5 badges, no broken shields.
- [ ] Top nav row parses: 4 links.
- [ ] Footer nav row parses: 5 links + attribution line.
- [ ] Tagline is **"A workspace for your coding agent."** (not the old "structured language" line).
- [ ] "Why Bonsai" section no longer exists by name.
- [ ] "Documentation" and "Contributing" no longer exist as standalone sections.
- [ ] `make build && go test ./...` still passes (sanity check — no code touched, but run anyway).
- [ ] PR created as draft targeting `main` with `Closes` line and Plan 25 reference.
