# The Bonsai pack template (`example-pack`)

A minimal, valid Bonsai pack: one role, one skill, one documented template, and the checks, CI and release every pack
needs. A new pack starts as a copy of this folder in its own repository. Bonsai's own CI checks this folder on every
Bonsai commit with the same script a pack's CI runs, so the template stays in step with Bonsai.

A **pack** is a Claude Code plugin (its roles, skills and any hooks or servers, which Claude Code loads) plus a
`bonsai/` folder, which Bonsai's engine reads when a project links the pack (`bonsai init`, `bonsai update`): who the
pack is, what it writes into a project, and what it declares. Claude Code never reads `bonsai/`; Bonsai never reads
Claude Code's plugin cache.

## Start a pack from this template

1. Copy this folder, `.github/` and `.gitattributes` included, to the top of a new git repository.
2. Choose the pack's id: a lower-case letter, then lower-case letters, digits and dashes, never `bonsai`. Put it in
   place of `example-pack` everywhere: `grep -rn example-pack .` lists each place (`plugin.json`'s `name`,
   `pack.yaml`'s `id`, `labels.yaml`'s `namespace` and label names, the role's `skills:` line, the skills' text).
3. Replace the example role, skill, document kind, label and deny rule with the pack's own, keeping every file
   documented as below, and give each new role or skill file its row under "Role and skill files".
4. In `.github/workflows/pack.yml`, check `BONSAI_COMMIT` (the Bonsai the pack is checked with) and
   `CLAUDE_CODE_VERSION`; move either only on purpose.
5. Run the checks locally (below), commit and push: the pack's CI runs them on Linux and Windows.

A project then links the pack by its repository and a tag or commit, for example `bonsai init --name demo --source
https://github.com/<owner>/<repo> --ref v0.1.0 --yes`.

## The folder

| File | What it is |
|---|---|
| `.claude-plugin/plugin.json` | The plugin manifest Claude Code reads. JSON holds no comments, so it is documented here, under "plugin.json" |
| `agents/` | The pack's roles, one file each; a role loads as `<id>:<name>` (`example-pack:example`) |
| `skills/` | The pack's skills, one folder each with its `SKILL.md`, to the open Agent Skills standard; a skill loads as `<id>:<name>` |
| `bonsai/pack.yaml` | The Bonsai manifest (format `bonsai.pack/1`): id, version, needs, what the engine writes into a project, and the pack's document kinds, protected paths and roles. A `#` comment on every key documents it |
| `bonsai/labels.yaml` | The labels the pack defines (format `bonsai.labels/1`), in its own namespace. A `#` comment on every key |
| `ci/check.sh` | Every check the pack has, in one script; its header documents each subcommand and argument |
| `.github/workflows/pack.yml` | The pack's CI and release, calling `ci/check.sh`. A `#` comment on every key |
| `.gitattributes` | Every file is text with LF line endings on every checkout, Windows included, so `sh ci/check.sh` runs there |
| `README.md` | This page |

Where each list lives (none is copied here): the document kinds, the deny rules (the walls) and the roles with their
limits in `bonsai/pack.yaml`; the labels in `bonsai/labels.yaml`; a role's preloaded skills on its own `skills:` line;
lanes, in a pack that has them, in `bonsai/lanes.yaml`. `bonsai check --schema bonsai.pack` (and `bonsai.labels`,
`bonsai.lanes`) prints every field of those formats with its allowed values.

## Role and skill files

Each file in `agents/` and `skills/` has one row here, and each row names a file there: `ci/check.sh` fails otherwise.

| File | What it is |
|---|---|
| `agents/example.md` | The role `example-pack:example`: it keeps the project's notes. Declared in `bonsai/pack.yaml`'s `roles` with its limits |
| `skills/example/SKILL.md` | The skill `example-pack:example`, no template: how to keep a note. The role preloads it |
| `skills/note/SKILL.md` | The skill `example-pack:note`, the template of the document kind `note`: its fields table and the template to copy |

### How each kind of file documents itself

Every file a pack ships says its purpose, when to use it, and every field: its meaning, its allowed values (or where
the list is defined) and an example.

- **A YAML file** (`pack.yaml`, `labels.yaml`, `lanes.yaml`, `pack.yml`): a short header, and a `#` comment on every
  key, at its line's end or on the line just above. `bonsai check --pack` fails a key in `bonsai/` without one.
- **A template skill**: a fields table, `| Field | Meaning | Allowed values | Example |`, above its template, the first
  fenced block after the table. `bonsai check --pack` holds the table and the template's fields equal, both ways, and
  each allowed-values cell that lists values to the pack's own list (a document kind's statuses, a choice label's
  values). It knows a template skill only by that table: a template written without one is not checked at all.
- **A role file, and a skill that is no template**: an HTML comment just after the frontmatter, a line or two (its
  purpose, when it is used, where its fields are documented). The body is read into the session, so the comment stays
  short and the fields are documented once, below.
- **`ci/check.sh`**: its header. **`plugin.json`**: this page.

### Role files: frontmatter fields

| Field | Meaning | Allowed values | Example |
|---|---|---|---|
| `name` | The role's name; the file is `agents/<name>.md`, and the role loads as `<id>:<name>` | a lower-case letter, then lower-case letters, digits and dashes | `example` |
| `description` | When to use the role, as Claude Code shows it to the caller choosing an agent | one line of text, quoted | `"Keeps the project's notes: ..."` |
| `model` | The model Claude Code runs the role on; a caller may name another for one call | a Claude Code model alias (`opus`, `sonnet`, `haiku`), a full model id, or `inherit` (the caller's) | `sonnet` |
| `skills` | The skills loaded into the role's context when it starts, each by its name as the plugin loads it | a list of skill names | `[example-pack:example]` |
| `maxTurns` | Claude Code's limit on the role's turns: a backstop where Bonsai's guard does not run (hooks turned off), never the limit itself | a whole number, equal to the role's `max_turns` in `bonsai/pack.yaml` | `200` |

The role's body names the skills it reads as well, so a route that ignores `skills:` still finds them.

### Skill files: frontmatter fields

| Field | Meaning | Allowed values | Example |
|---|---|---|---|
| `name` | The skill's name; the folder is `skills/<name>/`, and the skill loads as `<id>:<name>` | a lower-case letter, then lower-case letters, digits and dashes | `note` |
| `description` | What the skill holds and when to open it: Claude reads it to decide whether to load the skill | one line of text, quoted | `"The note template: ..."` |

## plugin.json

| Field | Meaning | Allowed values | Example |
|---|---|---|---|
| `name` | The plugin's name, which Claude Code puts before every role and skill (`example-pack:example`) | the pack's id, exactly as `bonsai/pack.yaml`'s `id` | `"example-pack"` |
| `description` | What the pack is, in a sentence, as Claude Code's plugin lists show it | text | `"The Bonsai pack template: ..."` |
| `author` | Who makes the pack | an object; `name` is the maker's name | `{"name": "Bonsai"}` |

Claude Code reads other optional fields (`homepage`, `repository`, `license`, `keywords`, and the folders of
components); add one only with its row here. **Never add `version`** (below). Hooks, MCP or LSP servers and monitors
in `plugin.json` or beside it are code (below).

## Versions and releases

- **A project pins the pack by commit.** `bonsai.yaml` names a tag or commit; Bonsai's lock records the 40-character
  commit and refuses a tag that later moves. The commit is the version Claude Code sees: with a `version` in
  `plugin.json`, Claude Code would keep each machine on its cached copy whatever the commit, so `plugin.json` never
  carries one, and `claude plugin validate --json` warns about that one thing, which `ci/check.sh` allows.
- **`bonsai/pack.yaml`'s `version` names tags only.** A release is a tag `v` and that version (`v0.1.0`); nothing is
  uploaded, since a pack is fetched by commit. To release: raise `version`, commit, push, then tag that commit
  `v<version>` and push the tag. The CI checks everything again, checks that the tag equals the version, and makes the
  GitHub release with its notes.
- **Adopting a release is a person's step** in each project: change the tag in `bonsai.yaml` and run `bonsai update`,
  whose preview says what changes.

## What makes a pack run code

A pack runs code when it carries a part Claude Code or a shell runs on its own, without an agent's call: a hook line
in `bonsai/pack.yaml`'s `hooks`, a pack file such a line runs, or a part of the plugin itself (hooks, MCP or LSP
servers, monitors, a mod). Roles, skills and commands are prompts, not code. A pack with code is asked for on each
machine: `bonsai init` and `bonsai update` write its hook lines and install its plugin only with `--allow-exec` as
well as `--yes`, typed by a person; a pack without code, like this one, links with `--yes` alone.

`--allow-exec` gates Bonsai's own install of the plugin. In a scripted session Claude Code was seen installing a pack's
plugin from the project's settings by itself and running its hook before any `bonsai update --allow-exec`, so keep code
out of a pack unless it needs it.

### Adding a hook line

A `hooks` entry in `bonsai/pack.yaml` (the field docs: `bonsai check --schema bonsai.pack`):

```yaml
hooks:                                # hook lines the engine writes into .claude/settings.json
  - event: SessionStart               # the Claude Code hook event
    matcher: startup                  # its matcher, or null
    command: "sh example-pack/greet.sh"   # the shell line; never calls bash by name
    runs: ["example-pack/greet.sh"]   # every pack file the line runs, by its path in the project
    why: "Prints the project's greeting when a session starts; it blocks nothing."   # the preview's sentence
```

The file it runs is a `files` entry (`path: example-pack/greet.sh`, `from: greet.sh`, `kind: pack`) with its source in
`bonsai/files/`. A plugin's own hooks go in `hooks/hooks.json` (or inline in `plugin.json`), not in a file named from
there, so `bonsai check --pack` reads every command in them.

### Adding an always-on file

A file every session loads goes in the project's protocols folder, which the instruction block imports: a `files`
entry whose path starts `<protocols>/`, the engine putting it in the folder the project's `bonsai.yaml` names.

```yaml
files:                                # the files the engine writes into a project
  - path: "<protocols>/example-start.md"   # into the project's protocols folder, loaded in every session
    from: example-start.md            # its source, bonsai/files/example-start.md
    kind: pack                        # Bonsai's file: an edit in the project is a conflict
```

It is read in every session, so it keeps its docs to one HTML comment line pointing here, with the rest in this page.
The same holds for `block` (`bonsai/block.md`), the pack's part of the block in the project's `CLAUDE.md`.

## Roles and their limits

Each role is a file in `agents/` and an entry in `bonsai/pack.yaml`'s `roles`: its `kind` (`lead`, which may call
helpers, or `helper`, which calls no one), `session` (true only for the role a person's own session plays),
`writes` (false: Bonsai's guard refuses it every file edit), `max_turns`, and three limits, `minutes`, `actions` and
`red_runs`, each with a `default` (what an agent of the role gets when its brief names none) and a `max` (the most a
brief may give it). The pack ships each max equal to its default; a project sets its own default and max for any limit
under the pack's entry in its `bonsai.yaml`.

- The role file's `maxTurns` equals the role's `max_turns` (`bonsai check --pack` holds them equal). It is a generous
  backstop, about four times the default actions: Bonsai's guard enforces the limits, and `maxTurns` only stops an
  agent where the guard does not run.
- A pack that declares roles brings Bonsai's two guard lines for them into a project. A first link writes them with
  `--yes`; adding the pack to a project already linked changes Bonsai's own guard, so it needs `--allow-exec`, a
  person's.

## Deny rules

Each `deny` entry in `bonsai/pack.yaml` is a Claude Code permission rule with its `why`, the sentence `bonsai update`'s
preview prints. Write a project path from the project's top with a leading `/` (`Read(/example-pack/secret.txt)`), or
from the disk's root with `//` (`Read(//**/example.key)`): a relative rule, such as `Read(example-pack/secret.txt)`,
does not hold after a shell `cd`.

## Checks, CI and release

`ci/check.sh` is the one home of the pack's checks (its header documents each subcommand):

- `sh ci/check.sh check <bonsai> <folder> <source> <commit> [<path>]`: `bonsai check --pack` on the folder; this page's
  rows against `agents/` and `skills/`; a scratch project linked to the pack at the commit, with `claude` off the PATH;
  `bonsai check` there. Run locally from the pack's top with the commit to check (committed work only):
  `sh ci/check.sh check "$(command -v bonsai)" . . "$(git rev-parse HEAD)"`.
- `sh ci/check.sh validate <folder>`: `claude plugin validate --json`, failing on every error and warning but the
  missing `version`. It needs no Claude login (measured with an empty Claude configuration folder and no key), so it
  runs in CI with no secret.
- `sh ci/check.sh release <tag> <folder>`: fails unless the tag is `v` and `pack.yaml`'s `version`. It reads no git tag,
  so it can be tried with any name: `sh ci/check.sh release v0.1.0 .`.

`.github/workflows/pack.yml` runs on every push and pull request, a fork's included (it uses no secret): `check` on
Linux and Windows with a bonsai built from `BONSAI_COMMIT`; `validate` on Linux with the pinned Claude Code; then
`release`, which checks the tag the version needs on every push and, on a `v*` tag only, makes the GitHub release.

### On Windows

Before a `bonsai update` that installs a plugin, open one Claude Code session first, so its login is fresh: a plugin
install meeting a run-out access token has signed the person out of Claude Code.

## Evals

A pack may ship eval cases for `claude plugin eval` under `evals/`, one folder per case (a `prompt.md` with its
`graders/*.md`, or a `case.yaml`; `claude plugin eval init --bare <name>` writes a blank one). They call a model, so
they run on a maker's machine, never in CI:

```sh
claude plugin eval . --no-publish --trust-plugin --model haiku --judge-model haiku --runs 2 --max-cost-usd 2 --json
```

Give the target as a path: by name it is refused when several installed plugins share it. Read the report's
`runsPerCase` with care: it has shown the default (3) when `--runs 2` ran two runs per arm.
