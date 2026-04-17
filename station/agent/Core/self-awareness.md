---
tags: [core, self-awareness]
description: Behavioral guardrails and dynamic context monitoring reference.
---

# Self-Awareness

## Dynamic Context Monitoring

Your context usage is tracked automatically by two sensors:

- **Status bar** (after every response) — shows context %, turns, and session health to the user
- **Context guard** (before every prompt) — injects behavioral constraints when context grows high

Follow any context advisories injected into your prompt — they are calibrated to your actual usage.

## Behavioral Guardrails

These are not automated. Follow them yourself:

- If you're losing track of the task, re-read [agent/Core/memory.md](memory.md).
- If a task has more than 10 steps, break it into sub-tasks before starting.
- Never continue if you're unsure what the plan says — re-read it.
- If you've already read a file this session, don't re-read it unless it changed.

## User Preferences — UX & Collaboration

> Durable learnings captured during dogfooding sessions. These describe what the user values in work you deliver and how they prefer to iterate. Extend this section when new patterns emerge — don't let it calcify.

### On visual and UI work

- **Palette first, visuals second.**
    - **Why:** During the 2026-04-17 `bonsai init` review, user pushed back on "overall coloring" as the first complaint. Before any re-skinning, semantic tokens must exist — otherwise every color change becomes a codebase-wide find-replace.
    - **How to apply:** For any TUI change that touches color or style, check that semantic tokens (`ColorPrimary`, `ColorAccent`, etc.) exist and are used. Introduce them as a prereq step if missing.

- **Sleek and minimal over ornate.**
    - **Why:** User called the prior UI "bulky", "thrown together", and the flat `B O N S A I` banner "unprofessional". Default ASCII art, excessive padding, spaced-letter wordmarks, and busy layouts are taste-negative signals.
    - **How to apply:** Prefer tight wordmarks, compact boxes, meaningful glyphs over decoration. When in doubt, strip and measure — don't ornament.

- **Visible state over hidden state.**
    - **Why:** User noticed their project name disappeared after advancing to the description prompt, and that required-only ability sections silently skipped. Both broke their sense of "I did something — where did it go?"
    - **How to apply:** Any answered prompt, selected item, or auto-processed section should leave visible evidence on screen. Don't rely on Huh's default clear-on-submit behavior — print a summary line afterward.

- **Rich guidance, not cramped.**
    - **Why:** User said "the next to do hints should be more rich, own their own space, and can be bit more verbose" — terse one-liners at the bottom of panels feel disrespectful of the moment.
    - **How to apply:** Next-steps, file structure views, and any screen that showcases core value should get dedicated visual real estate and substantive copy. Don't default to a single `Hint()` call when the user just finished something significant.

- **TUI should redraw, not stack.**
    - **Why:** User complained that the review panel, generate confirmation, and success messages all piled on top of each other during init. Their mental model is that each major step should feel like a new canvas.
    - **How to apply:** On major step transitions (review, generate, complete), plan for AltScreen or explicit clear/redraw. Don't treat TUI output as a scrollable log.

### On planning and iteration

- **Fast iteration beats process for UX work.**
    - **Why:** User said "we don't need PRs for this, first we test locally" — for taste-heavy design work, the PR review loop is too slow to be useful until the visual direction is settled.
    - **How to apply:** When user signals "test locally," still use worktrees for isolation but commit directly to main (or merge to main locally). Skip PR creation. Save the PR flow for code-correctness-heavy work.

- **Pick scope pragmatically, foundations first.**
    - **Why:** User dropped 11 items at once (Group F) and said "pick some of these up." They trust me to sequence — which means wrong sequencing wastes their time.
    - **How to apply:** Group by dependency (foundations before consumers — e.g., palette before banner) and visible-win density. Defer architectural rewrites until taste has settled on smaller items. State the deferred scope explicitly so the user can redirect.

- **Propose scope before writing the plan.**
    - **Why:** Plan documents are verbose; scope disagreements at the plan stage cost a rewrite. A scope summary is ~10x cheaper to revise.
    - **How to apply:** For any taste-heavy or ambiguous task, send a scope summary (picked / deferred / rationale) and ask "OK to proceed?" before drafting the plan file.

- **Log findings as they surface, don't batch.**
    - **Why:** User said "add things to backlog as we go along, under ui ux testing category" — they want a running tally in real time, not a post-hoc summary.
    - **How to apply:** Maintain Group F (or equivalent) in `Playbook/Backlog.md` during any testing session. Each finding: category tag, Group F tag, specific fix options if known, source attribution. Don't fix inline and don't wait until the session ends.

### On communication

- **Concise and direct wins.**
    - **Why:** User makes fast decisions with minimal elaboration ("b, but i will init myself"). Long hedged answers waste their attention.
    - **How to apply:** Short options, direct recommendations, no preamble. Mirror their energy — if they write two sentences, don't respond with five paragraphs.

- **Surface incidental findings proactively.**
    - **Why:** The `go install` binary-name bug was discovered while setting up their test environment, not while looking for it. If I'd silently renamed the binary, the bug would have stayed hidden.
    - **How to apply:** When you hit a workaround while doing setup/chores, explicitly flag it as a finding. Don't normalize broken behavior into your flow.
