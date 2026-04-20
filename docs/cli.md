---
description: Concise reference for every bonsai command.
---

# CLI Reference

One section per command. Interactive unless noted.

## `bonsai init`

Bootstrap a new project — creates `.bonsai.yaml`, installs the Tech Lead,
generates scaffolding.

```bash
cd your-project
bonsai init
```

**Use when:** first time in a project.
**Gotcha:** refuses to run if `.bonsai.yaml` already exists — no flag to
force. To reset, delete the config + generated files first.

## `bonsai add`

Two modes, auto-detected:

- **New agent** — pick an uninstalled agent type (backend, frontend,
  fullstack, devops, security) and walk the full ability pickers.
- **Add abilities** — pick an already-installed agent; Bonsai shows only
  uninstalled items.

```bash
bonsai add
```

**Use when:** adding a code agent, or expanding an existing agent's
ability set.
**Gotcha:** Tech Lead must exist before any code agent can be added.
Bonsai selects `routine-check` automatically if any routine is picked.

## `bonsai remove`

Remove a whole agent, or a single ability.

```bash
bonsai remove <agent>              # remove agent (config only)
bonsai remove <agent> -d           # agent + delete files
bonsai remove skill <name>         # or workflow/protocol/sensor/routine
```

**Use when:** un-installing an agent or ability.
**Gotcha:** cannot remove Tech Lead while other agents exist. Required
items are blocked. `routine-check` is auto-managed — it removes itself
when the last routine goes.

## `bonsai list`

Read-only table of installed agents, workspaces, and abilities by
category.

```bash
bonsai list
```

**Use when:** you forgot what's installed.
**Gotcha:** shows display names (e.g. "Scope Guard Files"), not the
kebab-case names used in `.bonsai.yaml`.

## `bonsai catalog`

Browse the full embedded catalog — every agent type, skill, workflow,
protocol, sensor, routine, scaffolding item.

```bash
bonsai catalog                     # everything
bonsai catalog --agent backend     # filter by compatibility
bonsai catalog -a security         # short flag
```

**Use when:** exploring before running `bonsai add`.
**Gotcha:** catalog is what's **available**. Use `bonsai list` for
what's **installed**.

## `bonsai update`

Sync the workspace: detect new custom files, re-render all generated
files, resolve conflicts.

```bash
bonsai update
```

**Use when:** you created a custom skill/workflow/etc., edited
`.bonsai.yaml` by hand, or upgraded the Bonsai binary.
**Gotcha:** if you edited a generated file, Bonsai offers a
skip/overwrite/backup picker. `.bak` files land next to the original.

## `bonsai guide`

Render a bundled cheatsheet in the terminal. Run with no args to pick
interactively, or pass a topic directly.

```bash
bonsai guide                       # interactive picker
bonsai guide quickstart            # this + concepts/cli/custom-files
bonsai guide | less -R             # page long output, preserve color
```

**Use when:** offline reference or a quick refresher.
**Gotcha:** renders via glamour — output adapts to your terminal theme.
Passing an unknown topic exits non-zero.

## Global

- `--help` on any command for full flags.
- `--version` (root) prints the build version.
- `--no-color` disables ANSI color globally.

> Full guide: https://laststep.github.io/Bonsai/commands/init/
