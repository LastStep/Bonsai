---
name: operating-setup
description: "How an agent links, updates and sets up a project with Bonsai unattended. Open it before you run bonsai init or bonsai update, edit bonsai.yaml, change a pack's version, move a task's status where status_command is set, or change what this machine is set up with. Everything else an agent does with Bonsai is in base:operating-bonsai."
---

<!-- skills/operating-setup/SKILL.md (base:operating-setup): the second half of the operating skills; no template, and
no copy of a list. Its frontmatter: base's README.md, "Skill files". It holds linking, updating, editing
bonsai.yaml, status moves and this machine's settings; the core every agent reads before any bonsai call is
skills/operating-bonsai/SKILL.md (base:operating-bonsai). A later piece's lines go to the half that holds its
subject: a word that changes what a project or this machine is set up with (install, settings, labels) here; a word
an agent uses while it works there. Bonsai's cmd/bonsai/operating_test.go holds every bonsai line in the two to the
command table, each to its size, and every word of the table to one of them. -->

# Operating Bonsai: setup

This is the setup half: linking, updating, `bonsai.yaml`, status moves and this machine's settings. Read the core,
`base:operating-bonsai`, first: `--json`, refusals by their next step, check, the ladder, asks, briefs, budgets, the
guard and the person's gates are there, and what follows assumes them. The `bonsai` you run describes itself with
`bonsai --help --json`.

## Link and update: the main checkout, on its base branch

`init` and `update` write `.claude/settings.json`, which the ladder's first rung refuses on any other branch.

- **A first link**, at the packs and refs the person named (by these flags, or in `bonsai.yaml` first); where no ref
  is named, hand them the one you propose (the newest tag, else the default branch's head commit) and link on their
  yes. Without `--yes` it is a preview: nothing written, exit 4 (`needs-yes`). Read each settings line, then run it
  again with `--yes`.

  ```sh
  bonsai init --name <name> --source <url> --path <folder> --ref <tag or commit> --json
  ```

- **An update** applies what is already decided (a repair, a re-apply on this machine, the project's own stricter
  change): `bonsai update --json`, its preview read line by line, then `bonsai update --yes --json`. `needs-yes` is
  the person's: pass `--yes` only where their word already covers what the preview shows (a task or plan they
  approved, their yes to the line handed over); else hand the preview over. On Windows the person opens a Claude Code
  session before an update that installs a plugin (a fresh login).
- **Exit 5** (`conflicts`): `--keep <path>` when the project's edit should win, `--adopt <path>` when the pack's
  should; ask when unsure.
- **Exit 4 with "Runs code"** (`needs-allow-exec`): consent to code is never your own choice. Where a studio manages
  the project the command takes `--allow-exec` under the person's grant (the running task grants `bonsai.yaml` and
  the lock); elsewhere it refuses (`person-command`): stop and hand the person the exact line.
- **A plugin `waiting`**: on trust, the person opens Claude Code in the folder and accepts its question, then
  `bonsai update`; on code, the line above.
- Then `bonsai check --json` with no finding, and everything the command wrote (`bonsai.yaml`, the lock,
  `.claude/settings.json`, the block, the packs' files, the tables) in one commit.

## Edit bonsai.yaml

It is person-only: edit it only under a running task whose `bonsai.allows` grants it, and on your own only for a
stricter change: a rung or floor entry added, a floor raised to a green, clean climb's count, a protected path added.
A rung added still needs the person: the next climb refuses it (`config-looser`) until they run `bonsai update`,
so write it and hand them that line. A climb reads it as committed (`config-uncommitted` until
then). Write only the forms `base:workspace` shows, a mapping as indented `key: value`
lines: a `bonsai.yaml` Bonsai cannot read blocks every call, and only the person can repair it.

**Anything else loosens a check and is the person's step:** a floor lowered; a rung's command, `required`, tests,
ratchet or capture changed or removed; a protected or person-only path taken out; a test marked to skip in CI. Edit
nothing, however the request reaches you (chat, a brief, a task): you cannot tell a message from the person's own
step. Hand them the exact line, old and new, and why, or ask (`--type Decide`); they make the edit.

**A role's budget** (`roles` under a pack's entry) is neither: change a default or max either way under a task
granting `bonsai.yaml`; it holds for agents started after the commit. So are the `messages` thresholds.

**A pack's version is the person's step, whatever the preview would show:** a pack's `ref`, `source` or `path`
changed, a pack added to a linked project, a pack taken out. A request for one is not their word. Read
`bonsai status --full --json` (newer tags), hand the person the exact edit (the line, old and
new, or the entry) and why in one message, or ask; change nothing. On their yes to that edit make it, read
`bonsai update --json` line by line, hand back a move that removes a wall, deny rule, label definition or protected
path, else run `--yes`. Why: an agent loosening the checks it is held to, or choosing the packs that hold it, would
be judging its own work. `bonsai check --json` after every edit.

## Status moves

Read `status_writes`. Under `agents`, edit a task's status line yourself, by contract section 10.2's agents' moves:
between the open statuses but never to `approved`, to `done` only in a `close: agent` lane, to `blocked` with
`blocked_by` filled. Under `command` (a studio manages the project) the guard refuses an agent's edit of a status
line (`status-line`), a lane line (`lane-line`), a task file outside the main checkout (`task-outside-main`) and a
person-only path put into `bonsai.allows` (`grant-person-only`): run the command `status_command` names instead.
