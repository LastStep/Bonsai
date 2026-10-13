# Reference: the lists

Every list Bonsai owns: its values when Bonsai owns them, whether it is closed or open, where it is defined and the
command that prints it in a project. A closed list never grows inside a format's major; an open list may, and a reader
shows a value it does not know as `other`. Every list has one home (spec section 6): this page is read from the code.

**This page is generated. Never edit it by hand.** It is written by `go generate ./...` from the tables in Bonsai's
code and the schemas in `formats/`; a test rebuilds it and fails on any difference, naming that command. Change a list
in its home, run `go generate ./...` from the repository's root, and commit the page with the change.
For a program, `bonsai --help --json` is the same knowledge for the command words, their flags and exit codes, and the
error words.

## Contents

- Exit codes
- Error words
- Check words
- Check words about the ladder's shape
- Check --pack words
- Check words later steps build
- Rung statuses
- Tests forms
- Test files by tests form
- Git integrity findings
- Ask types
- Active task: reasons there is none
- Generated kinds
- Bonsai's document kinds
- Pack file kinds
- Needs kinds
- Claude Code states
- Plugin results
- Installed places
- The guard line's tools
- Claude Code's key order: the top-level keys of a settings file
- Claude Code's key order: the keys of permissions
- Claude Code's key order: the keys of an inline marketplace's source object
- Log events
- Log categories
- Guard rules
- The guard's floor
- Stand-in names
- Delete words
- Words that run another program
- Secret kinds
- Lane rules
- Task statuses
- Label value kinds
- Label items kinds
- Label set_by
- Run outcomes
- Guard decisions
- Ask ops
- Ask sources
- Ask verdicts
- Ladder modes
- Ladder rung kinds
- Status: a pack's state
- Status: status_writes
- Active task: how it was found
- Lock file kinds
- Sessions: a row's kind
- Sessions: an hours row's kind
- Memory kinds
- Who takes a next step
- Changes: commands
- Changes: a settings line's change
- Changes: a code item's change
- Asks: an ask's state

## Exit codes

- Closed list.
- Defined in: format.ExitCodes (internal/format/exit.go).
- Printed in a project by: `bonsai --help --json` (exit_codes, and each word's exits).
- Spec section 3. Every word's own table says which of these it returns and what each means for it.

| Code | In short | Means | For |
|---|---|---|---|
| 0 | ok | the command did what it was asked (check found nothing, status read the workspace, a preview was printed and nothing was due) | every word but hook |
| 1 | check findings | check found at least one finding (a warning never makes it 1) | every word but hook |
| 2 | bad input | the command line, or a file the command was given, is not one Bonsai reads: nothing was done | every word but hook |
| 3 | runtime | something failed that the command line did not cause: a file or folder could not be read or written, git is missing, a fetch failed | every word but hook |
| 4 | wrong state or no --yes | the project is in a state the command does not run in, or the command writes and was given no --yes with no terminal to ask at: it printed the preview and wrote nothing | every word but hook |
| 5 | conflicts | files edited here were changed by the pack too: nothing is written until each is settled | every word but hook |
| 0 | allow | the hook lets the call go on | hook |
| 2 | block | the hook stops the call; the reason and the next step are on stderr, which Claude Code shows to the agent | hook |

## Error words

- Open list.
- Defined in: format.ErrorWords (internal/format/error.go).
- Printed in a project by: `bonsai check --schema bonsai.error`, and `bonsai --help --json` (error_words).
- The word of a refusal's `error.code` in every word's `--json`; who is who usually takes the next step. A word is added, never renamed or taken out.

| Word | Who | Means |
|---|---|---|
| `unknown-command` | agent | no command word was given, or one this Bonsai does not have |
| `bad-flag` | agent | the command line has a flag the word does not take, a flag without its value, a value given to a switch, a flag given twice that is taken once, or a word left over |
| `missing-value` | agent | a value the command needs was not given (init's --name, --source and --ref; hook's name) |
| `bad-value` | agent | a value was given in a form the command refuses (one of init's values; a --keep or --adopt path that names no conflict or edited file, or names one twice) |
| `values-differ` | person | init was given values that say otherwise than the bonsai.yaml already in the checkout |
| `unknown-format` | agent | check --schema names a format Bonsai does not know |
| `not-built` | agent | a word, flag or step this build of Bonsai does not have yet (the rebuild adds it in a later step) |
| `not-a-checkout` | agent | the folder is not inside a git checkout |
| `not-linked` | person | the checkout has no bonsai.yaml, so it is not linked to Bonsai |
| `old-workspace` | person | the checkout is a Bonsai 0.4.3 workspace, which this Bonsai neither reads nor changes |
| `not-main-checkout` | agent | the step runs only in the project's main checkout, and this is a worktree |
| `no-lock` | person | bonsai.yaml is in the checkout but .bonsai/lock.json is not, so update cannot tell what was consented to, nor unlink what Bonsai wrote |
| `not-a-pack` | agent | check --pack was given a folder that is not a pack's: it is not there, is not a folder, or holds no bonsai/pack.yaml |
| `bad-config` | person | bonsai.yaml is not one Bonsai reads (a line the reader refuses, a field of the wrong kind, another format) |
| `bad-lock` | person | .bonsai/lock.json is not one Bonsai reads |
| `bad-file` | person | a project file Bonsai writes into is not in a form it reads (.claude/settings.json not a JSON object Bonsai reads, CLAUDE.md's Bonsai block markers broken) |
| `bad-pack` | person | a pack is not one Bonsai can link (no bonsai/pack.yaml, a pack.yaml refused, a file it names missing, an id other than bonsai.yaml's, a hook line's runs not matching what it runs) |
| `packs-overlap` | person | two packs write one path |
| `ref-not-found` | person | a pack's tag or commit is not in its source, or is not a commit |
| `tag-moved` | person | a pack's tag now resolves to another commit than the one the lock holds (spec section 5: a moved tag is refused): nothing was written |
| `fetch-failed` | agent | fetching a pack failed (the network, the server, or git) |
| `git-missing` | person | git is not on the PATH |
| `bad-home` | person | the Bonsai home cannot be found or used (BONSAI_HOME, or the pack cache in it) |
| `read-failed` | agent | a file or folder cannot be read |
| `write-failed` | agent | a file cannot be written: nothing in the project was written |
| `needs-yes` | person | the command writes, and was given no --yes with no terminal to ask at: it printed the preview and wrote nothing |
| `needs-allow-exec` | person | the plan writes code that runs on this machine, which needs --allow-exec as well as --yes: nothing was written |
| `conflicts` | person | files edited here were changed by the pack too: nothing is written until each is settled with --keep or --adopt |
| `person-command` | person | an agent's session (CLAUDE_CODE_CHILD_SESSION set) ran a command only a person runs: init or update with --allow-exec, unlink, an update that changes or takes out Bonsai's guard or stop line, or an init or update that would record another base branch, or settings for the checks looser than the ones this machine accepted; an agent may run one only in a project in command mode, under a running task whose bonsai.allows a person gave bonsai.yaml and .bonsai/lock.json. Nothing was written |
| `test-build` | agent | a test build of Bonsai (stamped with its hook lines' place and a scratch folder) was asked to link, update, unlink or run the ladder outside its scratch folder, or was stamped with one of the two only: it writes nothing there |
| `ask-not-open` | agent | bonsai answer or ask --resolve names a key that has no ask, or whose ask is already answered or resolved, or ask --status names a key that has no ask: the message says which, and nothing was written (exit 4) |
| `answer-own-session` | person | the session that filed the ask tried to answer it: only a person, or a session that did not ask, answers (contract section 9.3), and nothing was written (exit 4) |
| `label-not-defined` | agent | bonsai log append names a label that no definition in force has, or gives it a value of the wrong kind: nothing was written (exit 2) |
| `session-not-found` | agent | bonsai logs --session matches no session in the log, or matches several: the message names the matches (exit 4) |
| `ladder-busy` | agent | another ladder run holds this home's lock for the checkout and did not let go within the wait: nothing was run or written (exit 4); the message names the holder's process and when its budget ends, and the same command again after that is the next step |
| `no-task` | agent | a local ladder run names no task, or the task named is none: the message carries the active-task function's reason code (no task, two running, the file unreadable, no such task), and nothing was run (exit 2) |
| `rung-not-defined` | agent | a rung the command asked for is not in bonsai.yaml's ladder, or bonsai.yaml has no ladder: the message names the rung, and nothing was run (exit 2) |
| `main-not-verified` | person | the project's main checkout is not verified on this machine, or it is not on the workspace's base branch this machine recorded for it, so no result can be written where the stop gate reads it, and nothing was run (exit 4); the person's step is bonsai update in the main checkout, or switching it back to its base branch (contract section 13; the guard's rule of the same name) |
| `config-uncommitted` | person | the main checkout's bonsai.yaml differs from the one committed at its HEAD, or its HEAD holds none: the ladder reads its rungs, floor, ratchets and ci_marked_tests from the committed one only, so nothing was run (exit 4); the person's step is to commit the change in the main checkout or revert it |
| `config-looser` | person | the main checkout's bonsai.yaml, as committed at its HEAD, sets the checks looser than the settings this machine last accepted for the workspace (a rung out of the floor or the ladder, a rung's check changed, a ratchet's floor lowered, a CI mark added, a path out of protected, person_only or never_edit, a document moved; the message names each), or this machine's record holds none yet: nothing was run (exit 4); the person's step is bonsai update in the main checkout, which records them as accepted, or a commit that puts the settings back |
| `partly-written` | agent | the command stopped part-way through writing, the lock not yet written (unlink: not yet removed): the same command again finishes the rest |
| `unexpected` | agent | something failed that Bonsai does not expect (no random number from the system, its own document not fitting its schema): run it again, and report it if it repeats |

## Check words

- Open list.
- Defined in: format.CheckWords (internal/format/check.go).
- Printed in a project by: `bonsai check --schema bonsai.check`.
- The code of a finding (makes `check` exit 1, and is a problem in `status --json`) or a warning (printed, never the exit code); who is who usually takes the next step. Machine says the word is about this machine rather than the commit (the home, the PATH, Claude Code and its plugins): rung 0 does not judge it, since a proof is of a commit.

| Word | Kind | Who | Machine | Means |
|---|---|---|---|---|
| `config` | finding | person | - | bonsai.yaml is not one Bonsai reads |
| `lock` | finding | person | - | .bonsai/lock.json is missing or not one Bonsai reads |
| `packs` | finding | person | - | bonsai.yaml and the lock name other packs, sources, refs or folders |
| `changed` | finding | person | - | a file the lock lists was edited: a pack file, Bonsai's block in CLAUDE.md, or Bonsai's lines in .claude/settings.json |
| `missing` | finding | person | - | a file the lock lists is gone |
| `gitignore` | finding | person | - | .bonsai/.gitignore is missing or changed, so .bonsai/local/ could be committed |
| `local` | finding | agent | - | git tracks or has staged a file from .bonsai/local/, which is never committed |
| `format0` | finding | agent | - | a format-0 file the lock's format0 list fixes changed or is gone (contract section 2.3) |
| `format0-new` | finding | agent | - | a task, run report or STATE file with no format: line that is not on the lock's format0 list: it takes a format: line first (contract section 2.3) |
| `document` | finding | agent | - | a file of one of Bonsai's kinds (task, run, state, memory, the two tables) does not read under its format, 0 or 1 |
| `label` | finding | agent | - | a label's value is not of the kind its definition says, or it is on a kind of document its definition does not name (contract section 5.2) |
| `label-twice` | finding | person | - | two sources (packs, or a file attached on this machine) define one label name (contract section 5.1) |
| `approve-first` | finding | person | - | a task in a lane with approve_first reached running, verify or done without reading approved since it last read todo or plan, as git history shows (contract section 6) |
| `absolute-path` | finding | agent | - | a committed Bonsai file (bonsai.yaml, a document of one of Bonsai's kinds) holds an absolute path in a field, a pack's source among them unless it is a remote URL (contract section 2.6; spec section 14, check 2) |
| `block-size` | finding | person | - | Bonsai's block in CLAUDE.md is over its fixed 40 lines (spec section 6) |
| `memory-index-size` | finding | agent | - | the memory index is over its fixed 120 lines or 12 KB (spec section 10) |
| `memory-note-size` | finding | agent | - | a memory note is over 4 KB (spec section 10) |
| `missing-path` | finding | agent | - | a project path named in CLAUDE.md, STATE or a memory note does not exist (Bonsai's own block left out: the lock checks it) |
| `secret` | finding | agent | - | a memory note or the memory index in the working tree holds a secret-shaped string, as Bonsai's redactor finds one: the finding names the note, the line and the kind of secret (the secret kinds' list), never the value; a person decides whether to rotate it (spec section 6) |
| `settings-rule` | finding | person | - | a permission rule in .claude/settings.json or .claude/settings.local.json is not valid on its own |
| `hooks-off` | finding | person | - | disableAllHooks is true in .claude/settings.json or .claude/settings.local.json, which turns off every hook, Bonsai's guard among them |
| `settings-env` | finding | person | - | the env of .claude/settings.json or .claude/settings.local.json sets a name that wraps, redirects or steers a hook line, Bonsai's guard among them: PATH, BASH_ENV, ENV, CLAUDE_CODE_SHELL_PREFIX, CLAUDE_CODE_GIT_BASH_PATH, CLAUDE_PROJECT_DIR, CLAUDE_CODE_STOP_HOOK_BLOCK_CAP, CLAUDE_CODE_CHILD_SESSION, or any name starting BONSAI_; the finding names the file and the name, never the value (step 5.3.6) |
| `plugin-version` | finding | person | - | a pack plugin's entry in a Bonsai marketplace carries a version, which a plugin pinned by commit never does (spec section 5) |
| `plugin` | finding | person | yes | plugin drift: Claude Code turns on another commit of a locked pack's plugin for this checkout (spec section 5) |
| `id-changed` | finding | person | yes | bonsai.yaml's workspace id is not the one this machine last recorded for the checkout (contract section 3) |
| `bonsai-path` | finding | person | yes | the bonsai on the PATH is not the installed one that install.json records (spec section 3) |
| `git-bash` | finding | person | yes | on Windows, Git Bash is not where Claude Code looks for it, so Claude Code runs every hook line in PowerShell, where Bonsai's lines do not parse: the guard's line, with onFailure block, then refuses every call on Claude Code 2.1.295 or later and lets every call through on an older one, and the stop gate's and the recorder's lines never run (step 5.3.6) |
| `git-user` | finding | person | yes | on Windows, the Git Bash Claude Code runs hook lines in is a Git for Windows installed for one user (outside the Program Files folders), so its /usr/local/bin, where Bonsai's lines look for bonsai when Windows' installed place holds none, is a folder that user can write (step 5.3.6) |
| `machine-record` | finding | person | yes | a record of a checkout in Bonsai's home (workspaces/r-<hex>/workspace.json) is not one Bonsai reads: until a person mends it, the guard verifies no main checkout on this machine, so no task's grant counts in any project here (step 5.3.6) |
| `claude-code-old` | warning | person | yes | Claude Code is older than the floor: Bonsai's own or a pack's needs.claude_code, the higher (spec section 7) |
| `claude-code-unknown` | warning | person | yes | Claude Code's version could not be read (not on the PATH, or claude --version gave an answer Bonsai does not read) |
| `same-id` | warning | person | yes | another checkout on this machine holds the same workspace id (contract section 3) |
| `run-reports` | warning | person | - | run reports are past bonsai.yaml's generated.run rule, listed for a person to delete (spec section 6) |
| `approve-first-unchecked` | warning | agent | - | approve_first was not checked: git history is not available (a shallow clone) or could not be read; never passed |
| `plugin-missing` | warning | person | yes | Claude Code reports a locked pack's plugin not installed for this checkout (it has registered the workspace's marketplace, or could not say whether it has) |
| `plugin-trust` | warning | person | yes | Claude Code has not registered the workspace's marketplace, which a Claude Code session in the checkout does once a person has trusted the folder (first-time trust, spec section 5), so the lock's plugin cannot be installed yet |
| `plugin-unchecked` | warning | person | yes | this machine's plugins were not compared with the lock (Claude Code not on the PATH, its answer unread, a local settings file unread) |
| `cache` | warning | person | yes | a lock written before formats set 4 names a pack this machine's cache lacks, so Bonsai's lines in .claude/settings.json were not checked |
| `tables` | warning | agent | - | a generated table in the main checkout is stale: the tasks table (.bonsai/tasks.md) differs from a rebuild of the task files, or the sessions table (.bonsai/sessions.md) lacks a row for an ended session or subagent run in the log, or either is missing; the message names which; the tables lag between moves by design and grant nothing (spec section 6; contract section 7.5) |
| `local-unchecked` | warning | agent | - | git ls-files failed, so files from .bonsai/local/ in git's index were not looked for |
| `own-hooks` | warning | person | yes | Bonsai's own hook lines in .claude/settings.json are not this build's (the project was linked or last updated by another Bonsai), while Bonsai's lines are otherwise as the lock says; a new hook line runs code, so update writes them only with a person's --allow-exec (spec section 7; step 5.1.1) |

## Check words about the ladder's shape

- Open list.
- Defined in: format.LadderCheckWords (internal/format/check.go).
- Printed in a project by: `bonsai check --schema bonsai.check`.
- The code of a finding or a warning about the ladder in bonsai.yaml (plan-5 5.4.0; the checks are 5.4.2's), in the same open list as the check words above; who is who usually takes the next step.

| Word | Kind | Who | Machine | Means |
|---|---|---|---|---|
| `rung-twice` | finding | person | - | two rungs of bonsai.yaml's ladder carry one number, so a task's bonsai.ladder and the ladder_floor cannot name one of them |
| `floor-no-rung` | finding | person | - | ladder_floor names a rung number that no rung of the ladder has |
| `ratchet-no-rung` | finding | person | - | a floor in bonsai.yaml's ratchets that no rung's ratchet names, so it holds back nothing |
| `capture-pattern` | finding | person | - | a rung's capture holds a pattern that does not compile as a Go (RE2) regular expression: the finding names the rung, the capture and the error |
| `ratchet-no-count` | finding | person | - | a rung's ratchet has no count: no capture of the same name, and the rung has no tests form to count passed tests from |
| `rung-bash` | finding | person | - | a rung's command calls bash by name (spec section 3), so it would run a different shell on Windows than elsewhere |
| `tests-unknown` | warning | person | - | a rung's tests is a word Bonsai does not know, so its tests are not read: the warning names the known ones (TestForms, bonsai check --schema bonsai.workspace) |

## Check --pack words

- Open list.
- Defined in: format.PackCheckWords (internal/format/check.go).
- Printed in a project by: `bonsai check --schema bonsai.check`.
- The code of a finding from `bonsai check --pack <folder>`; every one is a finding, and none appears in a project's check.

| Word | Kind | Who | Means |
|---|---|---|---|
| `pack-schema` | finding | agent | bonsai/pack.yaml, bonsai/labels.yaml or bonsai/lanes.yaml does not fit its format as its writer writes it (format 1's YAML, the format line, every field present, each of a type and value the schema allows, in the schema's order), or the engine would refuse it at a link (a rule a schema cannot say, or a file pack.yaml names missing from the pack) |
| `pack-comment` | finding | agent | a key in bonsai/pack.yaml, labels.yaml or lanes.yaml has no # comment at the end of its line or on the line just above it (spec section 5: every key documents itself) |
| `pack-fields` | finding | agent | a template skill's fields table and its template differ: a field of the template with no row, a row naming no field the template holds, no template after the table, or, for a Bonsai format, fields other than the format's schema |
| `pack-values` | finding | agent | an allowed-values cell of a template skill's fields table lists values other than the closed list its field has (a Bonsai format's schema, or the pack's own lanes, a declared kind's statuses or a choice label's values) |
| `pack-why` | finding | agent | a deny rule in bonsai/pack.yaml has no why, the sentence update's preview prints for it: missing, null or blank |
| `pack-plugin-version` | finding | agent | .claude-plugin/plugin.json carries a version, which a pack pinned by commit never does (spec section 5), or is not a JSON object Bonsai reads |
| `pack-block` | finding | agent | the instruction block a project linked to this pack alone would get (its markers, imports, label definitions and bonsai/block.md) is over its fixed 40 lines (spec section 6) |
| `pack-documents` | finding | agent | a document kind bonsai/pack.yaml declares is not well formed (contract section 7.3): a field the schema does not allow, a name of Bonsai's own or given twice, not exactly one of path and file, an id pattern Go does not read, a status twice, or a move or stamp naming a status the kind does not have |
| `pack-protected` | finding | agent | a protected path bonsai/pack.yaml declares is not a well-formed glob: project-relative with forward slashes, no empty, . or .. segment, each segment one path.Match reads (** whole segments only) |
| `pack-bash` | finding | agent | a hook command calls bash by name (spec section 3): a hook line in bonsai/pack.yaml, or a hook of the plugin itself |
| `pack-runs` | finding | agent | a hook command in bonsai/pack.yaml names a file the pack writes (by its path in the project or its file name) that its runs does not list, so a change to that file would run unseen (step 5.1.1) |

## Check words later steps build

- Closed list.
- Defined in: engine.checkLater (internal/engine/check.go).
- Printed in a project by: none.
- Spec section 6's findings and warnings that are not built yet; each joins the check words above when its step lands.

| Word | Built in | Means |
|---|---|---|
| `stranded` | step 5.6 | a machine folder stranded under an old path (contract section 3) |

## Rung statuses

- Open list.
- Defined in: format.RungStatuses (internal/format/ladder.go).
- Printed in a project by: `bonsai check --schema bonsai.ladder`.
- How a rung of a ladder result ended, in `rungs[].status`; a reader shows a word it does not know as other.

| Word | Means |
|---|---|
| `green` | the rung ran and passed |
| `red` | the rung ran and failed: its exit, its tests, its ratchet's count or its timeout; the reason says which |
| `skipped` | the rung did not run: an earlier required rung was red, or the run did not ask for it |
| `error` | the rung could not run or its outcome could not be read (the command did not start, a test output did not parse, a ver-git rung could not read git) |
| `pending` | a rung not built yet, which today's results hold; Bonsai's runner never writes it |

## Tests forms

- Open list.
- Defined in: format.TestForms (internal/format/ladder.go).
- Printed in a project by: `bonsai check --schema bonsai.workspace`.
- The form of a rung's test output in bonsai.yaml (`ladder[].tests`); an unknown word means the tests are not read, and `check` warns.

| Word | Means |
|---|---|
| `go-test-json` | the rung's standard output is go test -json: one JSON event per line, tests named by package and test, subtests by their path |
| `tap` | the rung's standard output is TAP, as node's test runner prints it (node --test --test-reporter=tap): ok and not ok lines, subtests indented |
| `junit` | the rung's standard output is a JUnit XML report: testcase elements with their classname, name and file, failures and skips inside |

## Test files by tests form

- Closed list.
- Defined in: format.TestFileConventions (internal/format/ladder.go).
- Printed in a project by: `bonsai check --schema bonsai.workspace`.
- Which files are a project's test files for each tests form: what new tests must fail on the base reads (5.4.5), and git integrity's changed-test-file finding (5.4.6). Patterns are project-relative globs with forward slashes (`**` is a whole segment); a path matching any is a test file.

| Form | Patterns | Means |
|---|---|---|
| `go-test-json` | `**/*_test.go`, `**/testdata/**` | the *_test.go files, and the files under a package's testdata folder |
| `tap` | `**/*.test.js`, `**/*-test.js`, `**/*_test.js`, `**/test-*.js`, `**/test.js`, `**/test/**/*.js`, `**/*.test.mjs`, `**/*-test.mjs`, `**/*_test.mjs`, `**/test-*.mjs`, `**/test.mjs`, `**/test/**/*.mjs`, `**/*.test.cjs`, `**/*-test.cjs`, `**/*_test.cjs`, `**/test-*.cjs`, `**/test.cjs`, `**/test/**/*.cjs`, `**/*.test.ts`, `**/*-test.ts`, `**/*_test.ts`, `**/test-*.ts`, `**/test.ts`, `**/test/**/*.ts` | node's test runner's default patterns, for .js, .mjs, .cjs and .ts: *.test.*, *-test.*, *_test.*, test-*.*, test.*, and files under a test folder |
| `junit` | the files the report names | the files the report's testcase elements name (their file attribute); with none named, the test files cannot be told, and the check is not run |

## Git integrity findings

- Open list.
- Defined in: format.GitFindings (internal/format/ladder.go).
- Printed in a project by: `bonsai check --schema bonsai.ladder`.
- The names a `ver-git` rung gives what it found, as keys of the rung's `captures`; the rung informs a verifier and never turns red for one.

| Word | Means |
|---|---|
| `base` | recorded: the task's base commit, the merge base of HEAD with the base branch (on the base branch itself, the base an earlier climb of the task recorded); null when there is none |
| `stash-count` | recorded: the number of entries the stash holds, which the task's next climb compares (stash-added) |
| `exclude-hash` | recorded: sha256:<hex> of the common git folder's info/exclude with line ends read as LF (an absent file reads as empty), which the task's next climb compares (exclude-edited) |
| `test-file-changed` | a test file (by the tests form, TestFileConventions) differs between the task's base and the working tree: changed, added (untracked ones too) or deleted |
| `assume-unchanged` | the index marks a file assume-unchanged, so git stops looking for edits to it: it hides a change, and the run counts as dirty |
| `skip-worktree` | the index marks a file skip-worktree, so git keeps the index's copy whatever the working tree holds: it hides a change, and the run counts as dirty |
| `exclude-edited` | the common git folder's info/exclude (line ends read as LF) differs from the hash the task's last climb recorded |
| `stash-added` | the stash holds more entries than the task's last climb recorded |
| `base-rewritten` | the base the task's last climb recorded is no longer in HEAD's history (or no longer in a clone that is not shallow) |
| `exclude-lines` | the common git folder's info/exclude holds lines other than comments: patterns that ignore files for this clone alone, which no diff shows |
| `exclude-hidden` | a file or folder present that git ignores only by info/exclude or core.excludesFile (no .gitignore ignores it), so git status does not show it: it hides a change, and the run counts as dirty, but for one inside a .claude/ folder (Claude Code's own local files, which Claude Code ignores that way itself), named only |
| `untracked-gitignore` | a file or folder present that git ignores by a .gitignore git does not track (one that ignores itself too), so git status does not show it: it hides a change, and the run counts as dirty |
| `filter-local` | the repository's own git config defines a filter driver (filter.<name>.clean, smudge or process), which can change what git compares, so git status may not show a change |

## Ask types

- Open list.
- Defined in: format.AskTypes (internal/format/ask.go), held to the ask schema's description of `type`.
- Printed in a project by: `bonsai check --schema bonsai.ask`.
- A pack may define its own type (contract section 9.1).

| Word | Means |
|---|---|
| `Answer` | filed by an agent (contract section 9.2) |
| `Decide` | filed by an agent (contract section 9.2) |
| `Look` | filed by an agent (contract section 9.2) |
| `Play` | filed by an agent (contract section 9.2) |
| `Bless` | filed by the ladder alone, when a ratchet count rose (contract sections 9.2 and 11) |

## Active task: reasons there is none

- Closed list.
- Defined in: workspace.ActiveReasons (internal/workspace/active.go), held to formats/README.md's table "Why there is none".
- Printed in a project by: none; `bonsai status --active --json` prints the sentence of the one that applies (`why`).
- The codes of contract section 13's function and its fixtures' answers.

| Code | Means |
|---|---|
| `named-differ` | `--task` and `BONSAI_TASK` name different tasks. |
| `named-missing` | No task file has the named id. A named task that is missing never falls through to the running task. |
| `named-duplicate` | Two or more task files have the named id: an id resolves to exactly one file. |
| `named-broken` | The named task's file does not parse under its format. |
| `broken-file` | Nothing is named, and a task file does not parse. |
| `none-running` | Nothing is named, and no task reads `running`. |
| `many-running` | Nothing is named, and two or more tasks read `running`. |

## Generated kinds

- Closed list.
- Defined in: format.GeneratedKinds (internal/format/generated.go), held to bonsai.workspace/1's `generated` properties.
- Printed in a project by: none; `bonsai init` writes each kind's default rule into bonsai.yaml's `generated:` section.
- Spec section 6, "Generated files".

| Kind | Where | What | Written by | Default | Never cleaned | Cleaned | Takes a rule |
|---|---|---|---|---|---|---|---|
| `log` | .bonsai/local/log/ | the log, one file per session | the recorder, from the hooks, one line at a time | 30 days after a file's last line | an open session's file, or one holding an ended session or subagent run with no row yet in .bonsai/sessions.md | at a session's end, after its session_end line, within a budget of 1 second, oldest first; what the budget leaves waits for the next end | yes |
| `asks` | .bonsai/local/asks/ | questions for a person and their answers | bonsai ask and bonsai answer (and the ladder runner from step 5.4, as its Bless), one day file a day | kept | a day file holding an open ask, or the answer or withdrawal of an ask whose filing stays (else it would read open again) | at a session's end, within the same budget; by default nothing, as the default keeps | yes |
| `ladder` | .bonsai/local/ladder/ | ladder results, one per task | the ladder runner (step 5.4), one file per task | 7 days after the result's finished | the result of a task that is not done or cut (a task not found, or whose file does not read, counts as not done) | at a session's end, and by the ladder runner after each run | yes |
| `run` | the run reports' folder (documents.run) | run reports, committed history | the agent that did the work, in the run report format | kept | any, by Bonsai: past a rule, bonsai check lists them and a person deletes them | never, by Bonsai | yes |
| `sessions` | .bonsai/sessions.md | rows of the sessions table | bonsai check --write, from the log | kept | rows of a task that is not done or cut (none is never protected), and rows whose span is still in the log (the next write would add them again) | in bonsai check --write, the only writer of the table | yes |
| `tasks` | .bonsai/tasks.md | the tasks table | bonsai check --write, from the task files | a rebuild: nothing to clean | - | never: the table is rebuilt whole | no |

## Bonsai's document kinds

- Closed list.
- Defined in: workspace.BonsaiKindNames (internal/workspace/documents.go).
- Printed in a project by: none.
- The kinds of Bonsai's own documents, in contract section 7.3's order; no pack may declare one of these names, and a pack's own kinds are the open part.

Values, in order: `task`, `run`, `state`, `answers`, `memory`, `tasks`, `sessions`.

## Pack file kinds

- Closed list.
- Defined in: workspace.PackFileKinds (internal/workspace/pack.go), held to the lock schema's list.
- Printed in a project by: `bonsai check --schema bonsai.pack`.
- The kinds a pack gives its files in bonsai/pack.yaml (`files[].kind`).

Values, in order: `pack`, `once`.

## Needs kinds

- Open list.
- Defined in: status.NeedKinds (internal/status/needs.go).
- Printed in a project by: `bonsai status --json` (needs[].kind).
- What the workspace needs from a machine; a reader shows a kind it does not know as other.

| Word | Means |
|---|---|
| `pack` | a locked pack, with its source and version, whose plugin is installed here or was not asked about |
| `plugin` | a locked pack whose plugin Claude Code reports not installed for this checkout (status --full only) |
| `tool` | a tool by name with the version it needs: the Claude Code floor (name claude-code) |
| `mcp` | an MCP server a pack's needs name (not written by this build: pack.yaml's needs holds only claude_code) |
| `shell` | a shell the workspace needs (not written by this build) |

## Claude Code states

- Closed list.
- Defined in: engine.ClaudeStates (internal/engine/claude.go).
- Printed in a project by: `bonsai status --full --json` (checks.claude_code.state).
- Claude Code's version against the floor (spec section 7).

| Word | Means |
|---|---|
| `ok` | Claude Code's version was read and is at or above the floor |
| `old` | Claude Code's version was read and is below the floor (check warns: claude-code-old) |
| `unknown` | Claude Code's version could not be read: not on the PATH, failed, or an answer with no version (check warns: claude-code-unknown) |

## Plugin results

- Open list.
- Defined in: engine.PluginResults (internal/engine/lists.go).
- Printed in a project by: `bonsai init --json`, `bonsai update --json`, `bonsai unlink --json` (plugins[].result).
- What the plugin step did for one pack's plugin on this machine; a reader shows a result it does not know as other.

| Word | Means |
|---|---|
| `installed` | Claude Code has the plugin on, at the locked commit, for this checkout |
| `waiting` | not installed yet: the plugin carries code and --allow-exec was not given, or Claude Code has not registered the workspace's marketplace (a person's first session in the trusted folder does) |
| `failed` | Claude Code was asked and could not install (or list) the plugin; the next step is the command to run |
| `skipped` | Claude Code is not on the PATH, so nothing was installed (or, on unlink, nothing removed) on this machine |
| `uninstalled` | unlink only: the plugin was removed for this checkout, or was not installed (nothing to remove) |

## Installed places

- Closed list.
- Defined in: engine.PlaceWindows and engine.PlaceLinux (internal/engine/settings.go).
- Printed in a project by: `bonsai hook guard --help` (the line), `bonsai init --help`.
- Where Bonsai's hook lines run bonsai from, in this order, as the shell that runs a hook line names them (spec section 3; Rohan's answer (a)): the PATH is never read, and with neither there the guard's line refuses every call and the stop gate's every session's end. Check 2's one named exception: the same on every machine, naming no user. A test build stamped with its own place writes that place alone.

| Order | Place | Where |
|---|---|---|
| 1 | `/c/Program Files/Bonsai/bonsai.exe` | Windows: C:/Program Files/Bonsai/bonsai.exe, as Git Bash names it; only an administrator places it |
| 2 | `/usr/local/bin/bonsai` | WSL and Linux; only root places it (under Git Bash, a folder of Git's own install) |

## The guard line's tools

- Open list.
- Defined in: engine.GuardTools (internal/engine/settings.go).
- Printed in a project by: `bonsai hook guard --help` (its matcher).
- The tools the guard's hook line matches, joined by | as its matcher: every tool that writes a file or runs a command, read from Claude Code's tools reference on the version in use. A changed list is a changed hook line, which a project linked before takes only with --allow-exec --yes.

| Tool | What it does that the guard judges |
|---|---|
| `Edit` | edits a file |
| `Write` | creates or overwrites a file |
| `MultiEdit` | edits a file several times in one call (older Claude Code) |
| `NotebookEdit` | edits a Jupyter notebook's cells |
| `Bash` | runs a shell command |
| `PowerShell` | runs a PowerShell command |
| `Monitor` | runs a command in the background, under Bash's permission rules |

## Claude Code's key order: the top-level keys of a settings file

- Closed list.
- Defined in: engine.claudeKeyOrder (internal/engine/settings.go).
- Printed in a project by: none.
- The order Claude Code writes them in, measured on Claude Code 2.1.294 and 2.1.295 (step 5.1.7); a key Bonsai adds goes where Claude Code would put it. The order is the value.

Values, in order: `$schema`, `respectGitignore`, `cleanupPeriodDays`, `env`, `includeCoAuthoredBy`, `includeGitInstructions`, `permissions`, `model`, `enableAllProjectMcpServers`, `enabledMcpjsonServers`, `disabledMcpjsonServers`, `hooks`, `disableAllHooks`, `enabledPlugins`, `extraKnownMarketplaces`, `outputStyle`, `spinnerTipsEnabled`, `alwaysThinkingEnabled`, `autoMemoryEnabled`.

## Claude Code's key order: the keys of permissions

- Closed list.
- Defined in: engine.claudePermissionsOrder (internal/engine/settings.go).
- Printed in a project by: none.
- The order Claude Code writes them in, measured on Claude Code 2.1.294 and 2.1.295 (step 5.1.7); a key Bonsai adds goes where Claude Code would put it. The order is the value.

Values, in order: `allow`, `deny`, `ask`, `defaultMode`, `additionalDirectories`.

## Claude Code's key order: the keys of an inline marketplace's source object

- Closed list.
- Defined in: engine.claudeMarketplaceOrder (internal/engine/settings.go).
- Printed in a project by: none.
- The order Claude Code writes them in, measured on Claude Code 2.1.294 and 2.1.295 (step 5.1.7); a key Bonsai adds goes where Claude Code would put it. The order is the value.

Values, in order: `source`, `name`, `plugins`, `owner`.

## Log events

- Open list.
- Defined in: format.LogEvents (internal/format/log.go).
- Printed in a project by: `bonsai check --schema bonsai.log`.
- Contract section 8.2's event names, in the log record's `event`; a reader shows a word it does not know as other.

| Word | Means |
|---|---|
| `session_start` | a session began, resumed or was cleared (the agent's SessionStart); it carries the active task found, the model, and the bonsai binary's path and hash |
| `prompt` | a person's prompt was submitted (the agent's UserPromptSubmit); its words are never kept, only its kind |
| `tool_start` | a tool call is about to run (the agent's PreToolUse) |
| `tool_end` | a tool call finished (the agent's PostToolUse) |
| `tool_fail` | a tool call failed or was interrupted (the agent's PostToolUseFailure) |
| `permission` | the agent asked for a permission (the agent's PermissionRequest) |
| `notice` | the agent noticed something to tell a person, such as that it is waiting (the agent's Notification) |
| `subagent_start` | a subagent run began (the agent's SubagentStart); it carries the active task found |
| `subagent_stop` | a subagent run ended (the agent's SubagentStop) |
| `stop` | the agent finished a turn (the agent's Stop) |
| `session_end` | a session ended (the agent's SessionEnd); its reason says why |
| `guard` | the guard decided on a tool call, or the stop gate on a session's end (agent_event Stop), allow or deny, by a named rule |
| `ladder` | a ladder result was written (contract section 11) |
| `ask` | an ask was filed, resolved or answered; the target is its key and the kind its op |
| `event` | an outside event, written by bonsai log append with its labels (contract section 8.4) |
| `clean` | a generated file Bonsai deleted; the target is its project-relative path and the reason the rule that cleaned it (contract section 8.5) |

## Log categories

- Open list.
- Defined in: format.LogCategories (internal/format/log.go).
- Printed in a project by: `bonsai check --schema bonsai.log`.
- Contract section 8.2's tool categories, in the log record's `category`; a pack's own are `<namespace>.<Name>`, and a reader shows a word it does not know as other.

| Word | Means |
|---|---|
| `Read` | reading a file or a folder |
| `Search` | searching files or their contents (today's Grep and Glob) |
| `Edit` | changing a file in place |
| `Write` | writing a whole file |
| `Shell` | running a command line (today's Bash and PowerShell) |
| `Ladder` | running Bonsai's ladder |
| `MCP` | calling a tool of an MCP server |
| `Agent` | starting a subagent |
| `Web` | fetching or searching the web |
| `Other` | any other tool, or one the recorder cannot place |

## Guard rules

- Open list.
- Defined in: guard.Rules (internal/guard/rules.go).
- Printed in a project by: `bonsai logs --json` (a guard record's `rule`).
- Every rule a guard decision names, the log record's `rule` (contract section 8.1), and whether it allows or refuses the call; a reader shows a rule it does not know as other.

| Rule | Decision | Means |
|---|---|---|
| `protected` | deny | a path on a protected list (bonsai.yaml's protected, a pack's declared protected paths) that no running task grants |
| `person-only` | deny | a path on bonsai.yaml's person_only list, or one the floor counts as person-only (bonsai.yaml, .bonsai/lock.json, anything else in .bonsai/, and a bonsai.yaml or .bonsai/ at any depth), that no running task grants |
| `never-granted` | deny | a path of the floor no grant opens: .claude/settings.json, .claude/settings.local.json and .git at a checkout's top, and a .git at any depth (for a delete, the project's own top .git only) |
| `bonsai-files` | deny | Bonsai's own files, in every mode and under any grant: .bonsai/local/, .bonsai/.gitignore, the two tables (.bonsai/tasks.md, .bonsai/sessions.md), and anything in Bonsai's home (from WSL, a Windows user's too) |
| `bonsai-stand-in` | deny | a file named as Bonsai's program (bonsai, bonsai.exe, bonsai.cmd, bonsai.bat, bonsai.ps1, bonsai.com) outside every linked project, which a folder on the PATH could hold |
| `main-not-verified` | deny | a guarded path where no project holding it is a verified checkout, so no task's grant can count; for the stop gate, the session's checkout's main checkout is not verified, or not on the base branch this machine recorded for it, so its task and ladder result are not trusted |
| `other-project` | deny | a delete or move of another linked project's top folder or its .git, .bonsai or .claude folder (that project on this machine's record) |
| `granted` | allow | a guarded path the running task's bonsai.allows holds (the task read from the verified main checkout) |
| `not-protected` | allow | a path inside a linked project that nothing guards |
| `outside-project` | allow | a path in no linked project, in every form it takes |
| `task-outside-main` | deny | command mode: an edit of a task file outside the main checkout |
| `status-line` | deny | command mode: an edit of a declared document's status line (a new document may start at its kind's first status only) |
| `lane-line` | deny | command mode: an edit of a task's lane line |
| `grant-person-only` | deny | command mode: an edit that puts a person-only path into a task's bonsai.allows |
| `outside-label` | deny | command mode: an edit that sets a label whose definition reads set_by: outside |
| `bad-format` | deny | command mode: a write of a task, run report or STATE that fails its format |
| `bad-settings` | deny | this machine's settings for the workspace cannot be read when a judgment needs its mode |
| `delete-unnamed` | deny | a shell call's recursive or bulk delete that does not name what it deletes (an operand holding a glob, a variable or a substitution, or one that is ., .., /, ~ or a drive's root; xargs, a pipeline or a loop feeding it; git clean with no path; a delete of several files by a glob; a git pathspec with magic other than :/, :(top) or :(literal)), and always git clean -x or -X, git stash --all and rm --no-preserve-root |
| `delete-protected` | deny | a shell call's delete or move of a guarded path, or of a folder holding one, judged as an edit of that path (the reason says what guards it) |
| `delete-unsplit` | deny | a shell call's line the delete check cannot split (a quote or bracket left open, nesting past three levels) that holds a delete word, or whose runner words run past the walk's bound |
| `delete-named` | allow | a shell call whose deletes and moves all name what they take, none of it guarded |
| `person-command` | deny | bonsai with --allow-exec, bonsai unlink, or a line that unsets or changes CLAUDE_CODE_CHILD_SESSION: a person's commands, allowed to an agent only in a project in command mode under a running task's grant of bonsai.yaml and .bonsai/lock.json |
| `home-command` | deny | bonsai settings set, bonsai labels attach or detach: never in an agent session |
| `bonsai-command` | allow | a bonsai command an agent may run: init and update judged as an edit of bonsai.yaml and the lock, every other word free |
| `shell-not-judged` | allow | a Bash, PowerShell or Monitor call whose line changes nothing the guard judges: no delete or move, no bonsai command, CLAUDE_CODE_CHILD_SESSION left as it is |
| `tool-not-judged` | allow | a tool other than the file tools and the shells |
| `bad-input` | deny | the hook's payload cannot be read (for the stop gate, or it is not a Stop payload) |
| `bad-glob` | deny | a glob on a list is not a glob the guard reads |
| `bad-path` | deny | Windows cannot name where the path leads, or it runs through a junction or link that does not resolve |
| `bad-lock` | deny | the lock's declares cannot be read (or the lock is missing while bonsai.yaml names packs) |
| `bad-config` | deny | a bonsai.yaml is there but cannot be read (the session's own: blocked with no record) |
| `no-project` | deny | CLAUDE_PROJECT_DIR is not set, or its folder cannot be read (blocked with no record) |
| `log-failed` | deny | the guard would allow the call, or the stop gate the session's end, but cannot write its record |
| `over-time` | deny | the guard or the stop gate ran past its own time limit |
| `internal-error` | deny | the guard or the stop gate failed inside (a panic) |
| `not-linked` | allow | neither the session's folder nor the path is in a linked project: allowed with no record (spec section 7); for the stop gate, the session's folder is in no linked project |
| `stop-not-named` | allow | the stop gate: BONSAI_TASK names no task, so it does not engage: the session ends, with no record |
| `stop-task-missing` | deny | the stop gate: no task file in the main checkout has the id BONSAI_TASK names |
| `stop-task-twice` | deny | the stop gate: two or more task files have the named id |
| `stop-task-broken` | deny | the stop gate: the named task's file does not parse under its format |
| `stop-task-unreadable` | deny | the stop gate: the main checkout's task folder, or a file in it, cannot be read |
| `stop-bad-rungs` | deny | the stop gate: the task's bonsai.ladder (format 0's ladder) or the workspace's ladder_floor is not a list of rung numbers |
| `stop-no-run-report` | deny | the stop gate: the session's checkout holds no run report for the task in its run folder (a file named R-YYYY-MM-DD-T-nnnn.md for the task) |
| `stop-no-result` | deny | the stop gate: the main checkout holds no ladder result for the task (.bonsai/local/ladder/T-nnnn.json for the task) |
| `stop-bad-result` | deny | the stop gate: the ladder result cannot be read, or does not read as bonsai.ladder/1 |
| `stop-other-task` | deny | the stop gate: the ladder result at the task's path was made for another task: its task is another id or null, or its workspace is another project's |
| `stop-not-local` | deny | the stop gate: the ladder result's mode is not local (a CI run is never a task's proof) |
| `stop-red` | deny | the stop gate: the ladder result is not green |
| `stop-head-unreadable` | deny | the stop gate: the commit at the checkout's HEAD cannot be read from its git files |
| `stop-other-commit` | deny | the stop gate: the ladder result proves another commit than the checkout's HEAD |
| `stop-dirty` | deny | the stop gate: the ladder result was made on a tree with uncommitted changes |
| `stop-config-uncommitted` | deny | the stop gate: the main checkout's bonsai.yaml differs from the one committed at its HEAD, or its HEAD holds none, so the ladder's settings there are not trusted |
| `stop-config-looser` | deny | the stop gate: the main checkout's bonsai.yaml, as committed at its HEAD, sets the checks looser than the settings this machine last accepted for the workspace (a rung out of the floor or the ladder, a rung's check changed, a ratchet's floor lowered, a CI mark added, a path out of protected, person_only or never_edit, a document moved), or this machine's record holds none: a person records them with bonsai update |
| `stop-rung-missing` | deny | the stop gate: the result did not ask for a rung of the workspace's ladder_floor or of the task's bonsai.ladder |
| `stop-rung-not-green` | deny | the stop gate: a rung of the floor or of the task's bonsai.ladder is not green in the result |
| `stop-proven` | allow | the stop gate: a green local ladder result proves the checkout's HEAD, with the floor and the task's rungs green: the session ends |

## The guard's floor

- Closed list.
- Defined in: guard.Floor (internal/guard/floor.go).
- Printed in a project by: none.
- The paths the guard guards in every project whatever bonsai.yaml's lists say, matched as the lists' globs are, the first entry that matches deciding: person-only (a running task's grant opens it), never-granted, bonsai-files (no grant opens either), or free. A delete or move does not meet an entry marked so.

| Path | At | Class | For deletes | What it is |
|---|---|---|---|---|
| `.bonsai/STATE.md` | top | `free` | yes | where the project stands; agents rewrite it (spec section 6) |
| `.bonsai/local/` | top | `bonsai-files` | yes | the log, the asks and the ladder results, which only Bonsai writes |
| `.bonsai/.gitignore` | top | `bonsai-files` | yes | written by Bonsai, so .bonsai/local/ is never committed |
| `.bonsai/tasks.md` | top | `bonsai-files` | yes | a generated table: bonsai check --write rebuilds it |
| `.bonsai/sessions.md` | top | `bonsai-files` | yes | a generated table: bonsai check --write rebuilds it |
| `.bonsai/lock.json` | top | `person-only` | yes | the lock: what each pack wrote and declares |
| `.bonsai` | top | `person-only` | yes | Bonsai's folder in the project; nothing else of an agent's belongs in it |
| `bonsai.yaml` | top | `person-only` | yes | the project's Bonsai settings, its lists among them |
| `.claude/settings.json` | top | `never-granted` | yes | Claude Code's project settings, which hold Bonsai's hook lines and deny rules |
| `.claude/settings.local.json` | top | `never-granted` | yes | Claude Code's local settings, which beat the project's |
| `.git` | top | `never-granted` | yes | the checkout's git folder or file, which says where its main checkout is |
| `**/.git` | any depth | `never-granted` | no | a .git below the top would plant a checkout inside the project |
| `**/bonsai.yaml` | any depth | `person-only` | yes | a bonsai.yaml below the top would plant a project inside this one |
| `**/.bonsai` | any depth | `person-only` | yes | a .bonsai folder below the top is a planted project's |

## Stand-in names

- Closed list.
- Defined in: guard.StandInNames (internal/guard/own.go).
- Printed in a project by: none.
- The file names the guard refuses an agent's file tools writing outside every linked project, letter case ignored: a stand-in for Bonsai's program in a folder on the PATH.

Values, in order: `bonsai`, `bonsai.exe`, `bonsai.cmd`, `bonsai.bat`, `bonsai.ps1`, `bonsai.com`.

## Delete words

- Open list.
- Defined in: guard.DeleteWords (internal/guard/shell_words.go).
- Printed in a project by: none.
- The words the guard's delete check reads in a Bash, PowerShell or Monitor call (plan-5 5.3.3): matched by base name, letter case ignored, .exe dropped; in PowerShell a bare name is the cmdlet's alias, a name with a folder or .exe the program. A recursive or bulk delete that does not name what it deletes is refused (`delete-unnamed`), a delete or move of a guarded path judged as an edit of it (`delete-protected`).

| Word | Names | Read by | Deletes |
|---|---|---|---|
| `rm` | `rm` | every shell, as a program | its operands; recursive with -r, -R or --recursive (bundled, any prefix); always refused with --no-preserve-root |
| `rmdir` | `rmdir` | every shell, as a program | its operands (empty folders) |
| `unlink` | `unlink` | every shell, as a program | its operand |
| `shred` | `shred` | every shell, as a program | its operands (overwritten, and deleted with -u) |
| `mv` | `mv` | every shell, as a program | moves its sources away (with all they hold) onto its destination |
| `git rm` | `git` | every shell, as a program | its paths; recursive with -r; :/x and :(top)x at the checkout's top, other pathspec magic unnamed; its paths from --pathspec-from-file are unnamed; not with -n |
| `git clean` | `git` | every shell, as a program | the untracked files under its paths (bulk); with no path unnamed; pathspec magic read as git rm's; always refused with -x or -X; not with -n |
| `git stash` | `git` | every shell, as a program | with --all or -a, ignored files moved away (.bonsai/local/ among them): always refused |
| `git mv` | `git` | every shell, as a program | moves its sources away onto its destination; pathspec magic read as git rm's; not with -n |
| `find` | `find` | every shell, as a program | with -delete, or -exec, -execdir, -ok or -okdir running a delete word: what it finds under its starting points (bulk) |
| `rsync` | `rsync` | every shell, as a program | with any --delete option (--del): files in its destination (bulk); with --remove-source-files: moves its sources |
| `robocopy` | `robocopy` | every shell, as a program | with /MIR or /PURGE: files in its destination (bulk); with /MOV or /MOVE: moves its source; not with /L |
| `Remove-Item` | `remove-item`, `rm`, `ri`, `del`, `erase`, `rd`, `rmdir` | PowerShell, as a cmdlet or its alias | its -Path or -LiteralPath (-LiteralPath names it as written); recursive with -Recurse (any prefix); bulk when piped into or with -Include, -Exclude or -Filter; not with -WhatIf |
| `Move-Item` | `move-item`, `mv`, `mi`, `move` | PowerShell, as a cmdlet or its alias | moves its -Path or -LiteralPath onto its -Destination; piped into, unnamed |
| `Rename-Item` | `rename-item`, `ren`, `rni` | PowerShell, as a cmdlet or its alias | moves its -Path or -LiteralPath onto -NewName beside it; piped into, unnamed |
| `del` | `del`, `erase` | cmd.exe, as its own command | the files in its operands; recursive with /s |
| `rd` | `rd`, `rmdir` | cmd.exe, as its own command | its folders; recursive with /s |
| `move` | `move` | cmd.exe, as its own command | moves its sources onto its destination |
| `ren` | `ren`, `rename` | cmd.exe, as its own command | moves its source onto a new name beside it |

## Words that run another program

- Open list.
- Defined in: guard.RunnerWords (internal/guard/shell_words.go).
- Printed in a project by: none.
- The words the delete check passes over, with their options, to reach the program they run (plan-5 5.3.3 note 1).

| Word | Read by | What the check reads |
|---|---|---|
| `sudo` | bash | its options skipped; -D (--chdir) names the folder |
| `doas` | bash | its options skipped |
| `env` | bash | its options and NAME=value words skipped; -C (--chdir) names the folder; -S (--split-string) is split again |
| `command` | bash | skipped; with -v or -V it runs nothing |
| `builtin` | bash | skipped |
| `exec` | bash | its options skipped |
| `nice` | bash | its options skipped |
| `ionice` | bash | its options skipped |
| `nohup` | bash | skipped |
| `setsid` | bash | its options skipped |
| `stdbuf` | bash | its options skipped |
| `time` | bash | its options skipped |
| `timeout` | bash | its options and its duration skipped |
| `xargs` | bash | its options skipped; a delete it runs takes its operands from input, so it does not name them; a bonsai call it runs is a person's step unless its own words settle a command that is not a person's |
| `busybox` | bash | skipped: the word after it is the program |
| `chronic` | bash | skipped |
| `unbuffer` | bash | skipped |
| `fakeroot` | bash | its options skipped |
| `watch` | bash | its options skipped; its words joined and split again, as sh -c runs them, unless -x (--exec) |
| `wsl` | bash, PowerShell, cmd.exe | its options skipped; the program runs in bash, its words joined and split again unless -e (--exec); --cd names the folder |
| `start` | cmd.exe | its /options and a quoted title skipped; /d names the folder |
| `call` | cmd.exe | skipped |

## Secret kinds

- Open list.
- Defined in: redact.Kinds (internal/redact/redact.go), in the order its rules run.
- Printed in a project by: `bonsai check` (a `secret` finding names the kind it found).
- The kinds of secret Bonsai's redactor finds and takes out of every record; a `secret` finding on a memory note names the kind, never the value. A new rule of the redactor adds a kind.

| Kind | Takes out |
|---|---|
| `private-key` | a private-key block, from its BEGIN line to its END line or the text's end |
| `webhook-url` | a Discord or Slack webhook URL, whole |
| `url-credentials` | the user:password inside scheme://user:password@host |
| `anthropic-key` | an sk-ant- key |
| `openai-key` | an sk- key |
| `stripe-key` | an sk_live_, sk_test_, rk_live_ or rk_test_ key |
| `github-token` | a github_pat_, ghp_, gho_, ghu_, ghs_ or ghr_ token |
| `slack-token` | an xoxa-, xoxb-, xoxp-, xoxo-, xoxs- or xoxr- token |
| `npm-token` | an npm_ token |
| `aws-key-id` | an AKIA or ASIA key id |
| `google-api-key` | an AIza key |
| `google-oauth-token` | a ya29. token |
| `jwt` | a JSON web token (eyJ..., three dotted parts) |
| `secret-named-key` | the value after a secret-named key and its : or = (password: x, DB_TOKEN=x, "apiKey": "x") |
| `secret-flag` | the value after a flag ending in a secret word (--password x, --client-secret x) |
| `authorization-header` | the value of an Authorization header, after its scheme word |
| `bearer-token` | a Bearer value of eight or more token characters |
| `extra-header` | the value after extraheader= (git's http.extraHeader) |
| `random-run` | a run of 32 or more token characters that looks random |

## Lane rules

- Closed list.
- Defined in: formats/schemas/lanes.schema.json (`lanes[].approve_first`, `lanes[].close`).
- Printed in a project by: `bonsai check --schema bonsai.lanes`.
- The two rules Bonsai understands (contract section 6); everything else a lane says is for agents to read.

- `approve_first`: `true`, `false` (a move to running, verify or done needs the task to have read approved since it last read todo or plan).
- `close`: `person`, `agent` (who may move the task to done).

## Task statuses

- Closed list.
- Defined in: formats/schemas/task.schema.json (`status`).
- Printed in a project by: `bonsai check --schema bonsai.task`.

Values: `todo`, `plan`, `approved`, `running`, `verify`, `done`, `blocked`, `cut`.

## Label value kinds

- Closed list.
- Defined in: formats/schemas/labels.schema.json (`labels[].kind`).
- Printed in a project by: `bonsai check --schema bonsai.labels`.

Values: `choice`, `text`, `number`, `list`.

## Label items kinds

- Closed list.
- Defined in: formats/schemas/labels.schema.json (`labels[].items`).
- Printed in a project by: `bonsai check --schema bonsai.labels`.

Values: `text`, `number`, `null`.

## Label set_by

- Closed list.
- Defined in: formats/schemas/labels.schema.json (`labels[].set_by`).
- Printed in a project by: `bonsai check --schema bonsai.labels`.

Values: `agent`, `outside`.

## Run outcomes

- Closed list.
- Defined in: formats/schemas/run.schema.json (`outcome`).
- Printed in a project by: `bonsai check --schema bonsai.run`.

Values: `running`, `verify`, `needs-verifier`, `needs-person`, `merged`, `blocked`, `abandoned`.

## Guard decisions

- Closed list.
- Defined in: formats/schemas/log.schema.json (`decision`).
- Printed in a project by: `bonsai check --schema bonsai.log`.

Values: `allow`, `deny`, `null`.

## Ask ops

- Closed list.
- Defined in: formats/schemas/ask.schema.json (`op`).
- Printed in a project by: `bonsai check --schema bonsai.ask`.

Values: `file`, `resolve`, `answer`.

## Ask sources

- Closed list.
- Defined in: formats/schemas/ask.schema.json (`source`).
- Printed in a project by: `bonsai check --schema bonsai.ask`.

Values: `agent`, `ladder`.

## Ask verdicts

- Closed list.
- Defined in: formats/schemas/ask.schema.json (`verdict`).
- Printed in a project by: `bonsai check --schema bonsai.ask`.

Values: `pass`, `fail`, `null`.

## Ladder modes

- Closed list.
- Defined in: formats/schemas/ladder.schema.json (`mode`).
- Printed in a project by: `bonsai check --schema bonsai.ladder`.

Values: `local`, `ci`.

## Ladder rung kinds

- Closed list.
- Defined in: formats/schemas/ladder.schema.json (`rungs[].kind`).
- Printed in a project by: `bonsai check --schema bonsai.ladder`.

Values: `guard`, `command`, `ver-git`.

## Status: a pack's state

- Closed list.
- Defined in: formats/schemas/status.schema.json (`packs[].state`).
- Printed in a project by: `bonsai check --schema bonsai.status`.

Values: `ok`, `changed`, `missing`.

## Status: status_writes

- Closed list.
- Defined in: formats/schemas/status.schema.json (`status_writes`).
- Printed in a project by: `bonsai check --schema bonsai.status`.

Values: `agents`, `command`, `null`.

## Active task: how it was found

- Closed list.
- Defined in: formats/schemas/status.schema.json (`active_task.how`).
- Printed in a project by: `bonsai check --schema bonsai.status`.

Values: `named`, `running`, `null`.

## Lock file kinds

- Closed list.
- Defined in: formats/schemas/lock.schema.json (`files.*.kind`).
- Printed in a project by: `bonsai check --schema bonsai.lock`.

Values: `pack`, `once`, `block`, `keys`, `kept`.

## Sessions: a row's kind

- Closed list.
- Defined in: formats/schemas/sessions.schema.json (`sessions[].kind`).
- Printed in a project by: `bonsai check --schema bonsai.sessions`.

Values: `session`, `subagent`.

## Sessions: an hours row's kind

- Closed list.
- Defined in: formats/schemas/sessions.schema.json (`hours[].kind`).
- Printed in a project by: `bonsai check --schema bonsai.sessions`.

Values: `session`, `subagent`.

## Memory kinds

- Closed list.
- Defined in: formats/schemas/memory.schema.json (`kind`).
- Printed in a project by: `bonsai check --schema bonsai.memory`.

Values: `project`, `feedback`, `reference`, `index`.

## Who takes a next step

- Closed list.
- Defined in: formats/schemas/error.schema.json (`next.who`).
- Printed in a project by: `bonsai check --schema bonsai.error`.

Values: `agent`, `person`.

## Changes: commands

- Closed list.
- Defined in: formats/schemas/changes.schema.json (`command`).
- Printed in a project by: `bonsai check --schema bonsai.changes`.

Values: `init`, `update`, `unlink`.

## Changes: a settings line's change

- Closed list.
- Defined in: formats/schemas/changes.schema.json (`settings[].change`).
- Printed in a project by: `bonsai check --schema bonsai.changes`.

Values: `add`, `change`, `remove`.

## Changes: a code item's change

- Closed list.
- Defined in: formats/schemas/changes.schema.json (`runs_code[].change`).
- Printed in a project by: `bonsai check --schema bonsai.changes`.

Values: `add`, `change`, `remove`.

## Asks: an ask's state

- Closed list.
- Defined in: formats/schemas/asks.schema.json (`asks[].state`).
- Printed in a project by: `bonsai check --schema bonsai.asks`.

Values: `open`, `answered`, `resolved`.
