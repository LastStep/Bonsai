# Custom Files Guide

Bonsai installs catalog items (skills, workflows, protocols, sensors, routines) from its built-in catalog. You can also create **custom files** — your own items that live alongside catalog items and are tracked by Bonsai.

## How It Works

1. Create a file in the appropriate directory (e.g., `agent/Workflows/my-workflow.md`)
2. Add YAML frontmatter with at least a `description` field
3. Run `bonsai update` — it detects untracked files and offers to track them
4. Once tracked, custom items appear in CLAUDE.md nav tables and `.bonsai.yaml`

## Frontmatter Format

Every custom file needs YAML frontmatter at the top, delimited by `---`:

### Skills, Workflows, Protocols

```yaml
---
description: What this item does — shown in CLAUDE.md nav tables
display_name: My Custom Skill
---

# Content starts here
...
```

| Field | Required | Notes |
|-------|----------|-------|
| `description` | Yes | Shown in CLAUDE.md nav tables |
| `display_name` | No | Human-readable name. If omitted, derived from filename (`my-skill` becomes "My Skill") |

### Sensors

Sensors are shell scripts that run as Claude Code hooks. They need extra fields:

```yaml
---
description: Blocks risky operations on Fridays
display_name: Friday Guard
event: PreToolUse
matcher: Bash
---

#!/usr/bin/env bash
# Script content here
...
```

| Field | Required | Notes |
|-------|----------|-------|
| `description` | Yes | Shown in CLAUDE.md nav tables |
| `display_name` | No | Human-readable name |
| `event` | Yes | Hook event: `SessionStart`, `PreToolUse`, `PostToolUse`, `Stop`, `UserPromptSubmit` |
| `matcher` | No | Tool filter (e.g., `Bash`, `Edit`, `Write`, `Edit\|Write`) |

### Routines

Routines are periodic maintenance tasks with a frequency:

```yaml
---
description: Review and clean up stale feature branches
display_name: Branch Cleanup
frequency: 14 days
---

# Branch Cleanup

**Frequency:** Every 14 days

1. **List stale branches:**
   - Run `git branch --merged` to find branches already merged
   ...
```

| Field | Required | Notes |
|-------|----------|-------|
| `description` | Yes | Shown in CLAUDE.md nav tables |
| `display_name` | No | Human-readable name |
| `frequency` | Yes | How often to run (e.g., `5 days`, `7 days`, `14 days`) |

## File Naming

- Use **kebab-case** for filenames: `my-custom-skill.md`, not `MyCustomSkill.md`
- The filename (minus extension) becomes the item's machine name
- Skills, workflows, protocols, routines use `.md` extension
- Sensors use `.sh` extension

## Directory Placement

Place files in the correct directory under your agent workspace:

| Type | Directory | Extension |
|------|-----------|-----------|
| Skill | `agent/Skills/` | `.md` |
| Workflow | `agent/Workflows/` | `.md` |
| Protocol | `agent/Protocols/` | `.md` |
| Sensor | `agent/Sensors/` | `.sh` |
| Routine | `agent/Routines/` | `.md` |

## What `bonsai update` Does

When you run `bonsai update`:

1. **Scans** each agent's workspace for files not yet tracked by Bonsai
2. **Validates** frontmatter — shows warnings for files missing required fields
3. **Prompts** you to select which files to track (multi-select, all pre-selected)
4. **Tracks** selected files in `.bonsai.yaml` (config) and `.bonsai-lock.yaml` (lock)
5. **Re-renders** catalog items from the latest embedded templates
6. **Refreshes** CLAUDE.md nav tables (includes both catalog and custom items)
7. **Syncs** `.claude/settings.json` hooks (includes custom sensors)
8. **Updates** routine dashboard (includes custom routines)

## Example: Adding a Custom Workflow

```bash
# 1. Create the file
cat > station/agent/Workflows/session-wrapup.md << 'EOF'
---
description: End-of-session verification — verify work, check for mistakes, cleanup
display_name: Session Wrap-Up
---

# Session Wrap-Up

## When to Use

Run this workflow at the end of every working session.

## Steps

1. **Verify completed work** — check all changes compile and tests pass
2. **Review for mistakes** — scan recent edits for bugs or oversights
3. **Update status** — move items in Status.md
4. **Update memory** — persist any new context to memory.md
EOF

# 2. Track it
bonsai update
```

After running `bonsai update`, the workflow appears in:
- `.bonsai.yaml` under the agent's `workflows` list
- `.bonsai.yaml` under `custom_items` with its metadata
- CLAUDE.md navigation table
- `.bonsai-lock.yaml` with a `custom:workflows/session-wrapup` source

## Catalog vs Custom

| Aspect | Catalog Items | Custom Items |
|--------|--------------|--------------|
| Source | Built into the Bonsai binary | Created by you in your workspace |
| Metadata | Separate `meta.yaml` file | YAML frontmatter in the file itself |
| Re-renderable | Yes — templates re-rendered on update | No — your file is yours, never overwritten |
| Tracked in config | Item name in skills/workflows/etc. list | Item name in list + metadata in `custom_items` |
| Lock source | `catalog:skills/name` | `custom:skills/name` |
