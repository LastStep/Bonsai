# The formats set

This folder holds Bonsai's formats as files any program can test against: a JSON Schema for each of the ten formats
of the formats contract (`design/contract.md` §2), the eight that set 4 added (Bonsai spec §5, §6, §10, §16 row 29),
the one that set 5 added (`help`, plan-5 5.1.10), the two that set 6 added (`asks` and `logs`, plan-5 5.2.0), the
one that set 7 added (`climb`, plan-5 5.4.0), the three that set 8 added (`message`, `messages` and `agents`, plan-5
5.5.9) the three that set 10 added (`settings`, `attach` and `line`, plan-5 5.6.0) and the one that set 11 added (`now`,
plan P-T-5795, contract §23), one example document for each, the **trick files**: small, deliberately tricky YAML and markdown files (and,
from set 8, message files; from set 11, run reports), each with the one right outcome a reader must reach (accepted with this value, or refused
for this reason), and contract §13's
**active-task fixtures**, each with the answer every reader of the active task must give. `manifest.json` fixes every
file's bytes.

**Who reads it.** Bonsai's own readers test against it (its Go code embeds the schemas from plan part 2 on). Any other
reader of these formats, the studio's among them, tests against the set of the Bonsai release it pins. Nothing here
reads or writes a project, a log or an ask: these are tests and their answers, not code.

```
formats/
  README.md                 this page
  expect.json               the expected outcome of every case (below)
  manifest.json             the set version and the SHA-256 of every file's bytes (below)
  formats_test.go           the Go tests that check the set is whole (below; Go files are not in the manifest)
  fixtures_test.go          ... that the active-task fixtures and their answers are well formed and agree
  compare_test.go           ... that each schema changed from the set's base commit only by additions
  embed.go                  embeds the schemas for Bonsai's code, formats.Schema(<name>) (not in the manifest)
  schemas/<name>.schema.json    one JSON Schema per format: task, labels, lanes, run, state, log, ask, ladder,
                                status, lock; workspace, pack, tasks, sessions, memory, error, check, changes; help;
                                asks, logs; climb; message, messages, agents; settings, attach, line; now
  examples/<name>.json      one valid document per format, as a reader returns it
  examples/<name>.md|.yaml  for task, run, state, tasks, sessions, memory, message (markdown) and labels, lanes,
                            workspace, pack (YAML): the source file, also a case
  trick/yaml-1/<case>/case.md|.yaml   a case for one rule of the format-1 YAML grammar (contract §2.4)
  trick/yaml-0/<case>/case.md|.yaml   one of today's format-0 oddities
  trick/message-1/<case>/case.md      a message file for one rule of bonsai.message/1's reader (set 8)
  trick/run-1/<case>/case.md          a run report for its recorded time and bonsai.run/1's schema (set 11)
  active-task/<case>/       one of contract §13's fixtures: case.json and the project folders it names
  active-task/answers.json  every reader's answer for every fixture (below)
```

## The schemas

One file per format, JSON Schema draft 2020-12, named for the format: `task` (`bonsai.task/1`, contract §4), `labels`
(§5.2), `lanes` (§6), `run` (§7.1), `state` (§7.2), `log` (§8.1), `ask` (§9.1), `ladder` (§11), `status` (§12),
`lock` (§14); and, from set 4, `workspace` (`bonsai.workspace/1`, `bonsai.yaml`, spec §6), `pack` (`bonsai.pack/1`,
a pack's `bonsai/pack.yaml`, spec §5), `tasks` and `sessions` (`bonsai.tasks/1`, `bonsai.sessions/1`, the two
generated tables, contract §7.5), `memory` (`bonsai.memory/1`, a memory note and the index, spec §10), `error` (the
`error` object, `bonsai.error`, spec §3 and §16 row 29), and the two command outputs the studio reads now that Bonsai
has no screens: `check` (`bonsai.check/1`, `bonsai check --json`) and `changes` (`bonsai.changes/1`, the preview or
result of `init`, `update` and `unlink` with `--json`); from set 5, `help` (`bonsai.help/1`, `bonsai --help --json`:
every command word with its flags, exit codes and examples, the exit codes every word shares, and every `error` word
with its meaning and usual `who`, so an agent learns the tool without reading a page); and, from set 6, `asks`
(`bonsai.asks/1`, the `--json` of `ask`, `answer` and `asks`) and `logs` (`bonsai.logs/1`, the `--json` of `logs` and
`log append`); and, from set 7, `climb` (`bonsai.climb/1`, the `--json` of `bonsai ladder`: one climb of the ladder,
named for the act, since `bonsai.ladder/1` is the result's name); and, from set 8, `message` (`bonsai.message/1`, a
message between agents: a brief or a report, one markdown file each in `.bonsai/local/messages/`, the schema being
the message as `bonsai message --show --json` prints it), `messages` (`bonsai.messages/1`, the `--json` of `bonsai
messages` and of `bonsai message`) and `agents` (`bonsai.agents/1`, the `--json` of `bonsai agents`, the one
published view of what Bonsai has recorded about each agent); and, from set 10, `settings` (`bonsai.settings/1`,
the `--json` of `bonsai settings show` and `set`: this machine's settings, for the workspace or the home), `attach`
(`bonsai.attach/1`, the `--json` of `bonsai labels attach` and `detach`: label definitions attached on this machine)
and `line` (`bonsai.line/1`, the `--json` of `bonsai status --line`: the statusline's workspace half as fields); and,
from set 11, `now` (`bonsai.now/1`, the `--json` of `bonsai now`: the time now in every form a document or a record
takes, contract §23). Each documents itself: a top-level `description` (what the format is for, who writes and reads it,
where it lives) and a `description` and `examples` on every property. Later steps add their commands' outputs
the same way, each a new schema file: an addition.

- **A schema describes what a writer writes** (contract §2.2): every field the version knows, in the contract's fixed
  order (the order of `properties`), all of them `required`, with `null` or `[]` where a field does not apply. A
  reader is more lenient: a missing field reads as `null`, an unknown field is kept, so `additionalProperties` stays
  open.
- **Lists.** A closed list (it never grows inside a major) is an `enum`; an open list (it may grow; readers show an
  unknown value as `other`) is a plain string, its known values named in the description. These schemas are the one
  home of every list in the formats (contract §2.2, format review R2.8): Bonsai's code embeds them and keeps no
  second copy. **One kind of open list lives in Bonsai's code instead:** a list of fixed words that code reads at many
  places: the `error` object's `code`, `check`'s finding and warning `code`, and (from set 6) the log's `event` and
  `category`; (from set 7) a ladder result's rung `status` and a `ver-git` rung's findings, and a rung's `tests` form in
  `bonsai.yaml`, with each form's test-file convention. Its known words live in one Go table, their one home, which `bonsai check --schema <format>` prints (and
  the reference page of lists, `docs/reference/lists.md`, generated from the code by `go generate ./...`); the
  schema's description names those commands and does not copy the words, and the words in its examples are
  illustrative.
- **Copies held to their home.** The schema checker has no `$ref`, so a shape two schemas share is written in both,
  and a test holds the copy equal to its home: the `error` object, inline in `status`, `check`, `changes`, `asks` and
  `logs`, is `error.schema.json`'s (its `required` and `properties`, word for word), and so is every `next` object (a
  finding's, a warning's, the plugin step's); a ladder rung's `kind` in `bonsai.yaml` is the ladder result's closed
  list; the `workspace` object of `asks` and `logs` is `changes`'. From set 8: `messages` and `agents` hold the
  `error` and `workspace` copies too; a thread's `recipient` holds an agent's `counts` and `limits` as `agents` has
  them, their home; `status --json`'s role `kind` is `pack`'s closed list. The records `asks` and `logs` carry
  (`bonsai.ask/1`, `bonsai.log/1`) are not copied: they are open `object`s there, whose descriptions name their home,
  because a full copy would make the writer refuse an older or hand-written record; so are the messages
  `bonsai.messages/1` holds (`bonsai.message/1`).
- **Spec §16's additions** are in from the start: the lock's `declares` (row 18), the ladder's third rung kind
  `ver-git` (row 23), and `status --json`'s additions (row 24), which arrive as values, not new fields: a `needs`
  entry of kind `plugin` for a pack not installed on the machine, the Claude Code floor as a `needs` entry of kind
  `tool`, plugin drift and a swapped `bonsai` binary as `problems`; `--line` prints another output, not a field.
- **Typed open**, because the sources do not fix their shape (plan part 0: a shape still unclear is typed open):
  - `ladder`: a rung's `status` (a string; its known words are a Go table, `format.RungStatuses`, since set 7; the
    contract lists none), a rung's `tests` (an object or `null`; today's inner fields are described, none required),
    `leftovers` (an object, a text or `null`, as today's three shapes).
  - `status`: `mode` (`offline`, or `full` with `--full`; an open list), each `documents` entry (only `kind` and `from`
    required: contract §7.3 says `format` is absent for a pack's kind and a kind has `path` or `file`), and `checks`
    (an object or `null`; with `--full` it holds `packs`, `plugins`, `claude_code` and `mcp`).
  - `lock`: each pack's `declares` (an object: its kinds are `format.DeclaresKeys`, eight from set 8, their key names
    and inner layout are not fixed; format review 6.3).
  - `workspace`: a `packs` entry's `roles` (set 8) is a map from a role's name to any of its limits, each with
    `default`, `max` or both: the one place in the set where a field may be left out, since a person sets only what
    they change and what is left out keeps the pack's.
  - `ask`: `data` (an object or `null`: a pack type's own payload).
  - `workspace`: a ladder rung's `ratchet` and `capture` have no `type` at all, so any value reads, and a rung's
    `tests` is an open string. Set 7 fixes their shape **by description**, held by Bonsai's reader and `bonsai
    check` (step 5.4.2), not by the schema: `ratchet` is a name (`[a-z_][a-z0-9_]{0,39}`) or `null`, picking a floor
    in `ratchets` (none reads as 0), and its count is the rung's capture of the same name, or on a rung with `tests`
    and no such capture the number of tests the output names as passed; `capture` is a map from such a name to a Go
    (RE2) pattern, matched line by line against the rung's output, whose first group is the value; `tests` is one of
    the forms in `format.TestForms` (`go-test-json`, `tap`, `junit`), read from the rung's standard output, and an
    unknown word means the tests are not read and `check` warns, naming the known ones. A type in the schema was
    refused: the compare calls a `type` added a change, so it would need `bonsai.workspace/2`, a new major and a
    person's call (contract §2.2), for a shape today's files already have. **What lost:** a reader holding only the JSON
    Schema (the studio's) accepts a shape Bonsai's `check` refuses.
  - The workspace id is `ws-` and 26 lowercase base32 characters (contract §3); which base32 alphabet is not fixed,
    so the schemas take `[a-z0-9]{26}`.
- **The binary's two names, chosen in set 6.** Spec §16 row 27 adds the `bonsai` binary's own path and SHA-256 to the
  log's `session_start` records, and the format review (4.2) left the two field names open. Set 6 takes the names the
  guard has written since part 5, `bonsai_path` and `bonsai_sha256`, at the end of `bonsai.log/1` (see "Choices made in
  set 6" below).

**Choices made here**, where the sources leave a detail open (each the smallest that fits):
- Field order: each schema follows its contract table or example. The ladder result follows the format review's
  example (5.1: `format`, `task`, `workspace`, `mode`), then contract §11's list (`started`, `finished`, `git`,
  `requested`, `green`, `rungs`, `skipped`, `leftovers`, `proof`). A label definition puts `items`, `pattern` and
  `max` (contract §5.2's kind row, §5.6) between `values` and `kinds`.
- Lists typed closed (`enum`) because their source gives a fixed set and does not call them open: the label `kind`,
  `set_by` and `items`; a lane's `close`; the log's `decision`; the ask's `source` and `verdict` (today's ask
  writer refuses any other); the ladder's `mode` and rung `kind` (plan part 0 counts `ver-git` among the closed
  lists a major cannot grow); the status `packs[].state`, `status_writes` and `active_task.how`; the lock's file
  `kind`.
- A field that does not apply is `[]` when it holds a list (`values`, `options`, `skipped`) and `null` otherwise; a
  label definition writes every field, `grants: false` among them. The director, 8 Oct: label-definition fields stay
  required, because Bonsai's packs write every field; the contract's short examples (§5.2, §5.6, and the format
  review's 2.2), which leave out fields that do not apply, are out of date. A reader still reads a missing field as
  `null` (`grants` as `false`).
- A `status --json` need carries `kind`, `id`, `name`, `source` and `version`, each `null` where it does not apply
  (a pack by `id`, `source` and `version`; a tool by `name` and `version`): contract §12's two examples, written in
  full by the writer rule (contract §2.2). The Claude Code floor is named by `name`, as contract §12 and the format
  review (6.5, confirmed) have it; spec §7's "id" is the spec's slip.
- `status --json` exits 3 with only `format`, `bonsai`, `problems` and (from set 4) `error` filled, so every other
  field also takes `null`.
- Patterns are added only where a source fixes a form: ids (task, run, workspace, machine folder, UUID), dates and
  times, label names, hashes (the lock's and status's 40-character commits; a 64-character SHA-256; today's
  16-character input hash; a run report's commits have none, as contract §7.1 fixes no form), the Remote Control
  id, and project-relative paths (no leading slash, no backslash, no drive colon).
- `folded-deeper` holds for `>-` as for `>`: the strip indicator changes only the end of the value.

**Choices made in set 4**, where the sources leave a detail open:
- **The lock's `path`** (each pack's folder in its repository, as `bonsai.yaml` names it) is `null` for a pack at
  its repository's root. A lock written before set 4 has none, which a reader reads as unknown, not as the root
  (contract §2.2's "a new optional field"): a pack in a folder does not look moved after an upgrade, and the engine
  judges that pack by its content hash alone until the next `update` writes its path (step 5.1.5). Bonsai's lock
  writer records the path of every pack it writes.
- **`bonsai.yaml`'s `documents`** holds Bonsai's own kinds (`task`, `run`, `answers`, `memory`) and the `protocols`
  folder as required properties, and a pack's document kind as an entry under its own name (spec §6, contract
  §7.3); STATE and the two tables have fixed places and no entry. `generated` has one property per kind that takes a
  cleaning rule (`log`, `asks`, `ladder`, `run`, `sessions`), its one home; the tasks table is rebuilt whole and has
  none, and the pack cache's rule is a machine setting.
- **`pack.yaml`'s declarations** besides labels and lanes are named as `bonsai.yaml` names their kind: `documents`
  (each kind with contract §7.3's fields but `from`, which is the pack, and `format`, which a pack's kind has none
  of; its `path` or `file` is the default place, which `init` writes into `bonsai.yaml`) and `protected`.
- **Memory:** a note and `INDEX.md` share one format; the index has `kind: index`, and its `id` and `source` are
  `null`.
- **The tables:** a reader returns `tasks.md`'s active line as `active` (`id`, `why`) and its rows as `tasks`;
  `sessions.md`'s rows as `sessions`, each with `kind` (`session` or `subagent`) and, from set 6, `subagent`, and its
  hours as `hours`, sessions and subagent runs apart (contract §7.5: never added), in hours to one decimal place.
- **The `error` object** has no `format` field and no major: it is part of other documents and changes with them,
  only by additions. Its `next` is an object, `do` and `who` (format review 4.5's question, taken: `who` is its own
  field), and a finding's, a warning's and the plugin step's next step take the same shape, so an agent reads one.
- **`check` and `changes`** follow the plan's words where step 5.1.1's JSON differs: no `command`, `result` or `exit`
  in `check` (the format names the command, and the exit code is the process's); in `changes`, no `exit`, the
  workspace's fields in `status --json`'s order, `null` rather than `""` for a file's empty `why`, `own_hooks` for
  Bonsai's own hook lines at a first link (written with `--yes`), and a settings line's `runs_code` true only for a
  line that needs `--allow-exec`. `status` gains `error` at the end of its fields, an addition: `null` unless the
  command refuses or fails.
- **Lists in set 4.** Closed: `changes`' `command`, and a settings line's or a code item's `change`; a pack file's
  `kind` (`pack`, `once`); a memory `kind`; a sessions row's and an hours row's `kind`; `next.who` (`agent`,
  `person`); a ladder rung's `kind` in `bonsai.yaml` (the ladder result's list). Open, where what they name may
  widen: `changes`' `result` (no source fixes its words, and a run can stop part-way: `failed`), a file's `result`,
  a settings line's and a code item's `kind`, the plugin step's `result` (`installed`, `waiting`, `failed`, `skipped`,
  and unlink's `uninstalled`), the lock's word in `changes` (`written`, `unchanged`, and unlink's `removed`), the
  `error` and finding codes (above), and a ladder rung's `tests`. A finding's `code` has two tables, `CheckWords` (a
  project's) and `PackCheckWords` (`check --pack`'s); its `file` is project-relative, or under `--pack` relative to
  the pack's folder.
- **`help`, set 5.** One document, `bonsai --help --json`, read from the same tables the human `--help` and the
  behaviour are made from (the word registry in `cmd/bonsai`, `format.ExitCodes`, `format.ErrorWords`), so it cannot
  drift; a test holds it to every word, flag and error word the tables have, and to this schema. Only the bare `bonsai
  --help --json` prints it: a word's own `--help` stays text, with or without `--json`. `words` lists the sub-words
  (`hook guard`) as entries of their own, a sub-word not built yet with `later` set and its text fields empty; `later`
  on a word or a flag names the step that builds it. `exit_codes` is the codes every word shares and `hook`'s own;
  each word's `exits` says which it returns and what each means for it. `error_words` is `format.ErrorWords`, an open
  list like the `error` object's `code` it describes. An addition (a new schema file), so the schema-compare test
  passes unchanged.
- **Set 6** (plan-5 5.2.0), every change an addition:
  - **`bonsai_path` and `bonsai_sha256`** close `bonsai.log/1`, after `remote`. `bonsai_path` is the binary's path,
    forward slashes: under the person's home folder it starts with a tilde and a slash, so no user name is in the log
    (contract §2.6), else as it is, or null; `hook start` writes it at every `session_start` (step 5.2.4). The guard
    writes its own absolute form until step 5.3 changes guard code, so the schema holds no pattern on it that an
    absolute path would fail. `bonsai_sha256` is 64 lower-case hex, on every `session_start` and on the guard record
    that made its log file, else null. The log's `event` and `category` descriptions now name `bonsai check --schema
    bonsai.log` and the reference page, and copy no word; `text`, `kind`, `task`, `role`, `labels` and `target` say
    what Bonsai writes there now.
  - **`asks` and `logs`**, two command outputs. `bonsai.asks/1`: `format`, `workspace` (null when the command refused
    before reading one), `asks` (a list of `{key, state, filed, closed}`: `state` is `open`, `answered` or `resolved`, a
    closed list; `filed` is the key's latest `file` record and `closed` the `answer` or `resolve` record that closed
    it, or null, both `bonsai.ask/1` records as stored), `written` (the record this run appended, or null) and
    `error`. `bonsai.logs/1`: `format`, `workspace`, `files` (a list or null; each `{file, session, day, first, last,
    records, unreadable, ended, task, role}`, a day file with `session` null and `day` its date, and `ended`, `task`
    and `role` a session's, else null), `records` (one file's records as written, or null), `written` (the record
    `log append` appended, or null) and `error`. Their records are open objects, as said above.
  - **A sessions row's `subagent`**, at the end of the row: the subagent run's own id, its first 8 characters, or null
    for a session's row (`-` is not used: an empty cell). Without it two subagent runs of one session started in the
    same minute share a key (the session, the subagent and the start). `sessions.md` gains a ninth column, `Subagent`;
    a reader also reads the eight-column header a table written before set 6 has, every row's `subagent` null.
  - **`check`'s `notes`** (what check says to a person that is neither a finding nor a warning) and **`changes`'
    `left`** (what `unlink` leaves in place that is not among its files), each a list of strings at the end of its
    object, [] for none. They were text-only until now.
  - **Descriptions only**: the `uninstalled` plugin result and the `removed` lock word; the seven `declares` kinds;
    a lock written before set 4 reads its pack's `path` as unknown, not the root (above); `status`'s `mode`, `checks`
    and `formats` (every format with a major, so `bonsai.error` is not in it); `check`'s two code tables and `file`
    under `--pack`; `labels.yaml` and `lanes.yaml` carry a comment on every key, as a pack's files must.
- **Set 7** (plan-5 5.4.0, the ladder's shapes and lists, set before the runner is built), every change an addition:
  - **`bonsai.climb/1`**, a new schema with an example: the `--json` of `bonsai ladder`. `format`; `workspace` (null
    when the command refused before reading one); `task` (or null); `mode` (`local`, `ci` or null); `result`, the
    project-relative path written, or null; `sha256` of the result file's bytes (64 lower-case hex), or null; `ladder`,
    the result as written (`bonsai.ladder/1`; an open object here, as `asks` and `logs` hold theirs), or null; `bless`,
    **a list** of the ask records this run wrote (`bonsai.ask/1`, each a `file` or a `resolve`; one run can resolve
    several open `ladder:bless-*` asks and file one more), `[]` when none; `cleaned`, the project-relative paths the
    cleaner removed; `next`, `{do, who}` as the `error` object's, the step after a red climb (the first red rung's fix,
    then the climb again), or null when green; `error`. A reader of `status --json` finds it in `formats` (`read` `[1]`,
    `write` 1).
  - **`new_tests`** in a ladder result's ratchet ends with `passed_on_base` and `not_run`, beside `count` and
    `failed_on_base`; its `base` is `""` when no base could be found (a `null` would retype it, a new major).
  - **`status --json`'s `ladder`**, after `error`: the active task's last result `{task, path, green, mode, sha,
    finished, stale}`, `stale` true when `sha` is not this checkout's HEAD; null when there is no active task or no
    result. An agent reads where its proof stands without running anything.
  - **Go tables for the ladder's words**, printed by `bonsai check --schema bonsai.ladder` and `bonsai.workspace` and
    listed on the reference page: a rung's `status` words, git integrity's finding names (the keys of a `ver-git`
    rung's `captures`), the `tests` forms and each form's test-file convention. The schemas' descriptions name those
    commands and copy no word. Beside them, in `check`: the ladder-shape words (a table of their own, `format.LadderCheckWords`) and a **Machine** mark on every `check` word that is about this machine rather than the commit.
    Four error words are new (`ladder-busy`, `no-task`, `rung-not-defined`, `main-not-verified`), and the existing
    `test-build` now covers the ladder.
  - **Descriptions only**: a rung's `ratchet`, `capture` and `tests` (above, "Typed open"); a failed test's `run` and a
    `ver-git` rung's `captures`; what a `ladder` log record carries (`target` the result's path, `text` `sha256:` and
    its hash, `ok` its `green`, `task`, `kind` its `mode`, `category` `Ladder`, `bonsai_path` and `bonsai_sha256` the
    runner's own); `bonsai_path` (the recorder and the guard both write it from a tilde and a slash under the home; the guard only on
    the record that made the file); cleaning's two protections for `asks`, `ladder` and `sessions`; `keep_days` (any
    whole number from 0 up, **no maximum**: a bound is a new major, and a number past what a duration counts back,
    about 292 years, cleans nothing by age); `never_edit`'s rule `Edit(/<path>)`; the hook line the `help` and `changes`
    examples quote, in today's form.
- **Set 8** (plan-5 5.5.9, step 5.5's shapes and lists, set before the code that writes them; the 14 Oct design of
  roles, messages and agents), every change an addition:
  - **`bonsai.message/1`**, a new schema with its example (`message.md`, the file, and `message.json`, what `--show
    --json` prints) and five trick files (below): a brief or a report, `format`, `id` (`M-`, the UTC date, a dash and
    8 random hex), `thread`, `reply_to`, `task`, `kind` (`brief` or `report`, closed), `from` and `to` (addresses:
    `<role>`, `<role>@<task>`, `<role>@<task>#<n>`, held to a form a Windows folder can hold), `status` (`done`,
    `blocked`, `needs-input`, or null for a brief; closed), `budget` (a brief's `minutes`, `actions`, `red_runs`; null
    for a report), `pointers` (`path`, project-relative or from a tilde and a slash, and `sha256`), `created`,
    `labels`, then `body` and `trust`, the two a reader adds (the file's body, and the trust line the reader makes from
    `from`, never read from the file). Beside the schema the reader refuses a report with no status, a brief with one
    and a report with a budget (`format.ReadMessage`). The trust line's words have one home, `format.TrustLine`.
  - **`bonsai.messages/1`**: `format`, `workspace`, `threads` (each `id`, `task`, `from`, `to`, `state` (`open`,
    `answered`; closed), `last_at`, `flags` (open: `format.MessageFlags`), `messages` (each message's frontmatter, an
    open object) and `recipient` (the agent tied to the newest brief, its counts and limits as the guard wrote them, or
    null)), `malformed` (`path`, `thread`, `reason`), `message` (the one message `bonsai message` wrote or showed, an
    open object), `path` and `error`. One document for `bonsai message` and `bonsai messages`, as `bonsai.asks/1` is
    for `ask`, `answer` and `asks`: the plan names only the list's, and a command that refuses needs an `error`.
  - **`bonsai.agents/1`**: `format`, `workspace`, `agents` (each `id`, `session`, `parent`, `role`, `type`, `task`,
    `brief`, `model`, `started`, `ended`, `last_call`, `counts`, `limits`, `spent`, `report`, `log`, `thread`,
    `usage`; every one as recorded, none derived), `skipped` (`path`, `reason`) and `error`. `usage` is null, or one
    item a model as the newest `usage` records hold them, or one item `kind` `unknown` with its `reason`.
  - **Roles:** `pack.yaml` gains `roles` (each `name`, `kind` `lead` or `helper`, `session`, `writes`, `max_turns`,
    and `minutes`, `actions`, `red_runs`, each with `default` and `max`), so a pack with none writes `roles: []`; the
    lock's `declares` gains `roles`, its eighth kind; `bonsai.yaml`'s `packs` entry gains `roles` (null, or the
    project's own limits for the pack's roles), and the file gains `messages` (`unanswered_minutes`,
    `circling_briefs`, each null for the default; null for both), at its end. `status --json` gains `roles` and then
    `messages` at its end, null until steps 5.5.0 and 5.5.10 fill them, as set 7's `ladder` was.
  - **The log** gains `parent` (on a `spawn` record, the agent that started this one, `session` for the person's own
    session) and then `usage` (on a `usage` record, one model's token counts over an agent's whole transcript:
    `input`, `output`, `cache_read`, `cache_write`) at its end, each null elsewhere; its events gain `message`, `spawn`
    and `usage`.
  - **Generated kinds:** `messages` and `agents` (each 30 days by default; never an open thread, nor an agent's state
    while its log or its briefs last) join `generated`, the schema and `format.GeneratedKinds` alike.
  - **Go tables**, each listed on the reference page: the message flags (`format.MessageFlags`, which `bonsai check
    --schema bonsai.messages` prints with their thresholds' defaults) and the thresholds (`format.MessageThresholds`),
    the role limits (`format.RoleLimits`, held to every schema that lists them) and Bonsai's fixed helper limits
    (`format.HelperLimits`), `bonsai message`'s six error words, and the check words steps 5.5.0 and 5.5.10 raise,
    each in a table of its own so this set is final before those steps write the checks: `format.RoleCheckWords`,
    `format.MessageCheckWords` and `format.PackRoleCheckWords` (`bonsai check --schema bonsai.check` prints them). The closed lists (a message's kind and status, a thread's state, a role's kind, where a
    limit came from) live in the schemas, as every closed list does.
  - **Descriptions only:** a pack file's `path` may start with `<protocols>/` (step 5.5.0, note 1); the lock's
    `declares` names its eight kinds; a ladder result's `leftovers` is Bonsai's `{checkout, ms}`, `{unknown, ms}`, a
    text or null; `check`'s codes name every table.
  - **One more active-task fixture**, `named-garbage`: the named task's file holds only `garbage: [`, which format 0
    reads as an empty mapping, so it is no task (step 5.4.3's fix, now shared by every reader). Its `does_not_parse`
    is `[]`: the file reads under format 0, and is broken only as a task.
- **Set 10** (plan-5 5.6.0, step 5.6's shapes, set before the code that writes them; set 9 changed three descriptions
  only), every change an addition:
  - **`bonsai.settings/1`**: `format`; `scope` (`workspace` or `machine`, closed); `file` (the settings file's place,
    forward slashes, printed locally; null when the command refused before it knew it); `settings` (one entry per key
    of the scope: `key`, `value`, a text, a whole number or null, and `from`, `file` or `default`, closed); `changed`
    (the keys `settings set` changed, `[]` for `show` and for a set that changed nothing); `error`. The keys, their
    rules and defaults are a Go table, their one home (step 5.6.1), not a list in the schema.
  - **`bonsai.attach/1`**: `format`; `command` (`attach` or `detach`, closed); `namespace`; `version` (the
    definitions' own, or null); `file` (the machine folder's `labels/<namespace>.yaml`, forward slashes); `result`
    (`attached`, `updated`, `unchanged`, `detached`, `nothing`, closed, or null on a refusal); `labels` (the names
    attached after the command, `[]` after a detach); `error`.
  - **`bonsai.line/1`**: `format`; `workspace` (`id` and `name`, or null when the folder is not linked); `task` (`id`,
    `title`, `status`, or null); `running` (how many tasks read `running` when no one task is active, else null);
    `ladder` (`green`, `red`, `old`, `none`, closed, or null with no task); `verify` and `asks` (counts); `release` (a
    newer release's version, or null); `line` (the text, `""` when not linked); `error`.
  - **`status --json` gains `bonsai_release`** at its end, null until step 5.6.5 fills it: `version`, `newest`,
    `state` (`current`, `newer`, `unknown`, `not-a-release`; closed), `read` (RFC 3339 UTC, or null), `why` (one line
    when `unknown`) and `lines` (`linux` and `windows`, each a list of command lines, when `newer`).
  - **Descriptions only:** the memory schema says the personal layer is written by a person (an agent drafts and hands
    the note over), each layer has its own caps, named on the reference page rather than in the schema, and a
    project's notes are written through base's `memory` skill; `status`'s names `bonsai.line/1` for `--line`.
  - **Not given a schema**, each Bonsai's own file on one machine, read by Bonsai alone and documented in the Go file
    that writes it (as `workspace.json` is): the machine folder's and the home's `settings.json`, `install.json`, the
    home's release record, the Windows copy's record and the line's cache.
  - **Go**: the three outputs' types and writers (`format.Settings`, `format.Attach`, `format.Line`), status's
    `format.StatusRelease`, and two error words, `namespace-taken` and `label-redefined` (`format.ErrorWords`).
- **Set 11** (plan P-T-5795 step 1 and its Amendment 2, contract §23 "Time": the director's rule, every recorded time
  is UTC and display converts), every change an addition; no field and no pattern changes:
  - **`bonsai.now/1`**: `format`; `utc` (the time now in UTC, `YYYY-MM-DD HH:MM`, the record form a run report's
    `started` and `finished` hold); `day` (the UTC day); `iso` (ISO-8601 UTC with milliseconds, the log's form);
    `local` (`YYYY-MM-DD HH:MM +HH:MM`, the display form for a person, never recorded); `offset` (`+05:30`, for
    display); `zone` (the abbreviation Go prints, for a person's eye only, never parsed); `error`. All but `format` and
    `error` null on a refusal.
  - **Descriptions only:** every time field names its form and its zone (§23): machine records UTC, as always; a run
    report's `started` and `finished`, its id's day, and every day field (a task's dates, STATE's and a memory note's
    `updated`, the packs' stamps) UTC from set 11, the writer's local time with its offset unknown in a file written
    before it, never rewritten; the sessions table UTC, as it always was. The two unpatterned status times stay
    unpatterned.
  - **Trick files** under `trick/run-1/` (above), with their `schema` outcome.
  - **Go**: the output's type, reader and writer (`format.Now`) and the registry's row.
  - **Dropped before release:** Amendment 1's run `offset` field and the sessions table's offset column (the installed
    release refuses a tenth column); neither ever left a released set.
- **Bounds in set 4**, chosen so a later set never needs to lower one: `keep_days` and `keep_newest` take 0 and up
  (0 cleans as soon as the protections allow; they always win); a rung's `timeout_s` takes 1 and up.
- **Patterns in set 4**, each where a source or Bonsai's own reader fixes the form: a pack id (a lower-case letter,
  then lower-case letters, digits and dashes, as Bonsai's reader of `pack.yaml` and `bonsai.yaml` holds it); a
  document kind's name (a key's form, contract §2.4, since it is a key in `bonsai.yaml`); a memory id (`M-` and a
  slug); a fixed word (`error` and finding codes); the session id's first 8 characters; a pack's `source` and `ref`,
  which never start with `-` (git would read an option); a `never_edit` path, which as the deny rule `Edit(/<path>)`
  (anchored at the project's top folder) holds no parenthesis and does not start with `~`.

## The examples

`examples/<name>.json` is one made-up, valid document per format, stored as the JSON a reader returns; its schema
validates it, and its fields keep the schema's order. For the eleven YAML and markdown formats (`task`, `labels`,
`lanes`, `run`, `state`, `workspace`, `pack`, `tasks`, `sessions`, `memory`, `message`) the source sits beside it
(`<name>.md` or `<name>.yaml`) and is also a case in `expect.json`: its format-1 value is exactly `<name>.json`,
except for the two tables and a message, whose format-1 value is their frontmatter, `<name>.json` without the fields a
reader reads from the body or makes (`active` and `tasks`; `sessions` and `hours`; `body` and `trust`), which come
last. So the test checks every schema without a reader,
and a reader proves the YAML. The pack files (`labels.yaml`, `lanes.yaml`, `pack.yaml`) document themselves in `#`
comments, as every pack file does (contract §2.8), and `workspace.yaml` has a comment on every line, as `init`
writes `bonsai.yaml` (format review 1.4); the markdown examples carry the one-line pointer on their `format:` line
that a file made from a template carries, and the tables a line saying they are generated.

**The two tables' bodies**, which a JSON Schema cannot describe; the writer writes them so, and a reader reads them so:
- `tasks.md`: after the frontmatter, the line `Active task when none is named: <id>`, or `Active task when none is
  named: none (<why>)`; a blank line; then the table `| Task | Title | Status | Lane | Started | Finished |`, its
  `|---|` line, and one row per task, newest id first.
- `sessions.md`: after the frontmatter, the table `| Session | Kind | Task | Role | Model | Start | End | Minutes |
  Subagent |` (from set 6; before it, the same without `Subagent`, which a reader still reads), its `|---|` line and one
  row per session or subagent run; a blank line; the line `Hours per task and role (a
  subagent run lies inside its session's row, so the two are never added):`; a blank line; then the table `| Task |
  Role | Kind | Hours |`, its `|---|` line and one row per task, role and kind.
- In both, a `null` is an empty cell, a `|` inside a cell is written `\|`, and nothing else is escaped.
- A task file that does not parse under its format gives no row in `tasks.md`: `bonsai check` reports it as a
  finding (spec §6), and the active line says none, naming it (contract §13: such a file makes the answer none).

`examples/status.json` and `examples/changes.json` print absolute paths, as `status --json` and `init`, `update` and
`unlink` with `--json` do (printed locally, never forwarded); so do `examples/settings.json` and
`examples/attach.json` (from set 10), for the files in the home. The contract's own example (§12) uses a path under a
home folder; these use `/srv/projects/example` and `/srv/bonsai-home`, made-up absolute paths that the set's
private-string check (below) does not flag.

## The trick files

Each case is one folder holding one file, `case.md` (markdown with YAML frontmatter) or `case.yaml` (a YAML definition
file), with made-up content only.

- `trick/yaml-1/<case>/`: one case per rule of contract §2.4, accepted or refused. These start with a `format:` line,
  as a real file would, except the dispatch cases, which test where `format:` sits (first, not first, nested, absent).
- `trick/yaml-0/<case>/`: today's format-0 oddities (contract §2.4: a plain value holding `: `, unquoted hashes,
  `TRUE`, the last of two duplicate keys, document markers skipped). They have no `format:` line, so a format-1
  reader sends them to format 0.
- `trick/message-1/<case>/`: from set 8, a message file (`case.md`) for one rule of `bonsai.message/1`'s reader: a
  report with no status, a brief whose budget has a field of the wrong type, an id of the wrong shape, an unknown field
  kept (the major's rule), an address with a character a Windows folder name cannot hold. Their YAML reads under
  both formats (their `format0` and `format1` are accepted); their third outcome, `message`, is the message reader's.
- `trick/run-1/<case>/`: from set 11, a run report (`case.md`) for its recorded time (contract §23: UTC, no zone
  written): `started` in the record form (accepted); with an offset, quoted (its YAML reads; bonsai.run/1 refuses
  it); with a zone name, `IST`, quoted (the same); with an offset, unquoted (format 1 refuses the plain value). A case
  that format 1 accepts has a third outcome, `schema`: whether bonsai.run/1 accepts the value, read as a reader reads
  it (a missing field as null).
- The CRLF case (`trick/yaml-1/lines-crlf/case.md`) ends every line in CRLF, the lone-CR case
  (`trick/yaml-1/lines-cr/case.yaml`) every line in a CR alone, with no LF anywhere, and two cases start with a BOM
  (`bom-frontmatter`, `bom-definition`). The repo's `.gitattributes` sets `formats/** -text`, so no checkout changes
  their bytes.

A case's folder name starts with its rule (`key-reserved-on` is a `key-reserved` case). The rule, a sentence saying
what it tests, and both outcomes are in `expect.json`.

## expect.json

One file, UTF-8, LF, giving every case two outcomes (and a message's or a run report's trick file a third). Example (one case, shortened):

```json
{
  "about": "The expected outcomes of the formats set: ...",
  "cases": [
    {
      "path": "trick/yaml-1/plain-listed-0755/case.yaml",
      "rule": "plain-listed-refusal",
      "settled_by": "rule",
      "about": "A listed refusal of contract §2.4, bare: 0755. Quote this value.",
      "format0": { "outcome": "accepted", "value": { "format": "bonsai.labels/1", "namespace": "example", "mode": 755 } },
      "format1": { "outcome": "refused", "reason": "quote-this-value" }
    }
  ]
}
```

| Field | Meaning |
|---|---|
| `about` | What the file is. |
| `cases` | One entry per case, sorted by `path`; every file under `trick/` and the eleven example sources has one. |
| `path` | The input file, relative to `formats/`, forward slashes. |
| `rule` | The rule the case tests: one of the rule names the Go test lists (`rules` in `formats_test.go`; a message's start with `message-`), or `example`. |
| `settled_by` | `rule` when contract §2.4's words decide the format-1 outcome; `libraries` when they do not and §2.4's own test decides it: accepted only if a YAML 1.1 and a YAML 1.2 library both read the file and read the same value, otherwise refused (plan part 0). |
| `about` | One sentence: what the case tests. |
| `format0` | What the studio's frozen format-0 reader reaches, which reads every file as format 0, `format:` line or not: `{"outcome": "accepted", "value": <JSON>}` or `{"outcome": "refused", "message": "<the reader's own message>"}`. |
| `format1` | What a reader that dispatches as §2.4 says reaches: `{"outcome": "accepted", "value": <JSON>}`; `{"outcome": "refused", "reason": "<code>"}` with a code from the table below; or `{"outcome": "format-0"}` when the file has no top-level `format:` key, which sends it to format 0. |
| `value` | The YAML file's mapping, or for markdown the frontmatter's mapping (the body is not compared), as JSON. Under format 1 a date stays text and an integer has at most 15 digits. |
| `schema` | Only for a case under `trick/run-1/` that format 1 accepts (set 11): whether `bonsai.run/1` accepts the format-1 value, a missing field read as null (contract §2.2): `{"outcome": "accepted"}` or `{"outcome": "refused", "field": "<the field refused, as started>"}`; the schema checker refuses that field and no other. |
| `message` | Only for a case under `trick/message-1/` (set 8): what `bonsai.message/1`'s reader reaches (`format.ReadMessage`, which `bonsai message` and every reader of messages read through): `{"outcome": "accepted", "value": <the message as bonsai message --show --json prints it>}` or `{"outcome": "refused", "field": "<the field refused, as budget.minutes>"}`. |

## Reason codes

The format-1 refusals, by short code. A file can break more than one rule, so a reader reports one code by this order:
- **Dispatch first**, whatever else the file holds: `format-not-first`, then `format-too-new`. A top-level key is a key
  at the top level's indentation, plain or quoted, read up to its colon and a space, a tab or the line's end, as
  format 0's reader reads it: `format:` and a tab is the file's format key, and so is `"format":` and a tab (set 4,
  `dispatch-tab-quoted`), as `"format":` and a space already was; the file is then format 1, which refuses the quoted
  key.
- **Then the first problem reading top to bottom.**
- **On one line, in reading order, each problem found at the character where it starts**: the indentation
  (`tab-indent`) and a document marker at the line's start; any problem of a key at its first character, so the key
  comes before its value; an anchor, alias, tag, flow mapping, block scalar header, or plain scalar that fits no row at
  its first character; a quoted scalar or flow sequence that does not close on its line at its opening quote or `[`; a
  bad escape at its backslash; anything after a closing quote, after `{}` or after a flow sequence's last `]` where it
  follows; a `[` or `{` that opens inside a flow sequence where it opens; a flow sequence's items one by one, left to
  right, before what follows its last `]`; a character that `not-text` refuses where it stands; and a block scalar that
  ends the file with no line ending at the file's end.
- **Of two codes found at one character, the one listed first below.**

**Lines that end in a CR alone.** Format 1's lines end in LF or CRLF (§2.4), so a file whose lines end in a CR alone
is one line, read by the rules above: its first problem is most often a value holding the next line's key
(`quote-this-value`, `trick/yaml-1/lines-cr`). Whatever the code, a refusal on a line that holds a CR ending no line
names the line endings in its message and its next step (set 4). A markdown file whose lines end in a CR alone has no
frontmatter to either reader (its first line is not `---` alone), so it is format 0.

| Code | Refused because |
|---|---|
| `format-not-first` | A top-level `format:` key is not the file's first key (comments and blank lines may come before it). |
| `format-too-new` | The first key, `format:`, names a major this reader does not know (`bonsai.task/2`): contract §2.2's "format too new", and nothing else is read. |
| `not-text` | A character a YAML 1.1 and a YAML 1.2 library do not read alike (§2.4's own test): a byte that is not valid UTF-8; a control character but tab, a CR that does not end a line among them; U+FFFE or U+FFFF; U+0085, U+2028 or U+2029, which a YAML 1.1 reader takes as a line break. And a file that ends inside a `\|` or `>` block scalar with no line ending, where a YAML 1.1 library reads no final line feed and a YAML 1.2 one does. |
| `tab-indent` | A tab in a line's indentation. |
| `doc-marker` | A `---` or `...` line anywhere but a frontmatter's own two markers: a YAML file has none, and a frontmatter block holds none. |
| `line-not-read` | A line the grammar does not consume: a block sequence at its key's own indent, a sequence item where a mapping key belongs, a value on the line below its key, a plain value over two lines, text after `{}` or after a flow sequence's last `]`. Also an empty flow sequence item, a trailing comma's among them (`[a, ]`): split at its commas it is an empty plain scalar, null, where both libraries read no item. And a line of only spaces inside a block scalar, deeper than the block: both libraries keep its spaces, format 0 drops them, and no one reading the file can see them. |
| `key-complex` | A complex key (`? key` then `: value`). |
| `key-quoted` | A quoted key, single or double. |
| `key-merge` | The merge key `<<`. |
| `key-form` | Any other key that is neither `[a-z][a-z0-9_]*` nor a label name `<namespace>.<name>` (contract §5.1): `Title`, `done-when`. Also a key longer than 1024 characters: §2.4 fixes a key's form, not its length, and a YAML 1.1 and a YAML 1.2 library both refuse a longer key (§2.4's own test). |
| `key-reserved` | A key that is `y`, `n`, `yes`, `no`, `on`, `off`, `true`, `false` or `null`. |
| `key-twice` | A key its mapping already holds. |
| `seq-dash-space` | A block sequence item whose `-` is not followed by exactly one space. |
| `anchor` | An anchor (`&name`). |
| `alias` | An alias (`*name`). |
| `tag` | A tag (`!name`, `!!str`). |
| `flow-mapping` | A flow mapping with content (`{a: 1}`); `{}` alone, spaces inside or not (`{ }`), is an empty map. |
| `flow-nested` | A `[` or `{` that opens inside a flow sequence. |
| `flow-multiline` | A flow sequence that does not close on its own line. |
| `block-indicator` | A block scalar header other than `\|`, `\|-`, `>` and `>-`: `\|+`, `>+`, an indentation indicator such as `\|2`. |
| `block-hash-line` | A line inside a block scalar whose first character after the indentation is `#`. |
| `folded-deeper` | In a `>` or `>-` block scalar, a line indented deeper than its first line. |
| `quoted-multiline` | A quoted scalar that does not close on its own line. |
| `bad-escape` | In double quotes, a backslash before anything but `"`, `\`, `n`, `r` or `t`. |
| `unescaped-quote` | In double quotes, an unescaped `"` inside: the first unescaped `"` after the opening one closes the scalar, and the rest of the line before any comment holds another `"`. |
| `after-quote` | Anything but spaces and a comment after a quoted scalar's closing quote (with no further `"`, which is `unescaped-quote`). |
| `quote-this-value` | A plain scalar that fits no row of §2.4's table (`0755`, `1.2.3`, `True`, `on`, a value ending in `:`, a 16-digit integer, a value holding a tab or starting with one, as after `key:` and a tab), or, inside a flow sequence, a plain item holding `]`, `}` or `?`. A YAML 1.1 library cannot read a tab in a plain scalar, nor a `?` in a flow sequence's plain item, so §2.4's own test refuses both. A flow sequence runs from its `[` to the last `]` on its line, and its items are split at commas outside quotes. |

## How the outcomes were found

- **Format 0** means exactly what the studio's frozen YAML reader, `yaml.mjs` at commit `4a05eac`, reads (contract
  §2.3, §2.4). Each case was run through that file, taken read-only with `git show`, the way the
  studio's callers read files: `readFileSync(path, 'utf8')` (so a BOM and CRLF reach the reader), then `parseYaml` for
  a YAML file or `parseFrontmatter(...).data` for markdown. `format0` holds its value or its exact error message.
- **Format 1** outcomes follow contract §2.4's grammar rule by rule. Where §2.4's words do not settle a value
  (`settled_by: libraries`: two single quotes inside single quotes, the value of a block scalar, the CRLF case's block
  scalar at the end of a frontmatter; in set 4, `{ }` and `[ ]` with a space inside, and a sequence item whose value
  is the deeper block below a comment), two libraries were run and both read the same value: PyYAML 5.4.1 (YAML 1.1)
  and the npm package `yaml` 2.8.3 (YAML 1.2). A markdown file's frontmatter is the lines between its markers, each
  with its own line ending. Dates are settled by §2.4 itself: they stay text, though a YAML 1.1 library reads a date.

## The active-task fixtures

Contract §13's fixtures, "one folder per case, run against every reader": `active-task/<case>/`, and the answers in
`active-task/answers.json`, one section per reader of the active task. Step 5.1.5 builds the active-task function
(`status --active`, contract §13's read-only command) and tests it on the function's section; step 5.3 tests the
guard's and the stop gate's, step 5.4 rung 0's; the studio's readers test against them when the studio links (spec
§14 step 7). Every file is made up; the task files follow `bonsai.task/1` (one is format 0, with no `format:` line,
and the files `does_not_parse` names are refused by their format).

**A case folder** holds `case.json` and one folder per project: `main/` (always), `other/` (a second project), and, in
a worktree case whose worktree's files differ from main's, `worktree/`. Each project folder holds `bonsai.yaml`, whose
`documents.task` is `work/tasks`, and its task files. No case holds a `.git`: a case that needs git is a layout each
reader's test builds.

| `case.json` field | Meaning |
|---|---|
| `about` | One sentence: what the case tests. |
| `layout` | `plain`: each project folder is a checkout of its own (a reader's test that finds checkouts through git makes each one a repository). `worktree`: `main/` is the main checkout, a repository with every file committed, and the test adds a worktree of it on a branch of its own (`git worktree add`), then copies `worktree/`'s files, when the case has that folder, over that worktree's: its own, stale copies of the task files. |
| `checkout` | The project folder the reader runs for: the path the guard judges, the checkout the ladder runs in. `main`, `other` or `worktree`. |
| `session` | The project folder holding the session's working directory: the environment counts only for its project, and the stop gate runs for it. |
| `task` | What `--task` names, or `null`: only a command that takes it (the ladder) is given it. |
| `env` | What `BONSAI_TASK` names in the session's environment, or `null`. |
| `does_not_parse` | The case's files that do not parse under their format, by path in the case folder; `[]` for none. |

**`answers.json`** holds `about`, then four sections, each a list with one entry per case folder, in the folders'
order (sorted by name), each starting with `case`, the folder's name:

| Section | An entry's fields |
|---|---|
| `function` | The active-task function given every input of the case (contract §13's steps): `id` (the active task, or `null`), `how` (`named` or `running`, or `null` when there is none: `status --json`'s `active_task.how`, its one home) and `why` (`null` when there is one, else a reason code from the table below, never free text). |
| `guard` | `grants`: the task whose `bonsai.allows` the guard honours for the checkout, or `null`. The guard is a hook: it takes no `--task`, so the environment alone names a task, and it honours grants only from an active task that reads `running`. |
| `rung0` | `grants`: the task whose grants rung 0 judges the diff against, at any status, or `null`: the function's task when it is named. |
| `stop_gate` | `engages` (the gate engages only for a task the environment names, for the session's own project), `task` (`ok`, or `blocks` when the named task is missing, twice or does not parse; `null` when the gate does not engage) and `ladder` (where the gate reads the ladder result, which must prove the session's own checkout's HEAD: the main checkout's `.bonsai/local/ladder/<task id>.json`, by path in the case folder; `null` unless `task` is `ok`). A gate that does not engage allows. |

### Why there is none

The reasons the active-task function gives when there is no active task, by short code, their one home. When a task
is named, its reasons come first, in this order; when none is named, a task file that does not parse comes before
the count of running tasks (that file could be the running one).

| Code | There is none because |
|---|---|
| `named-differ` | `--task` and `BONSAI_TASK` name different tasks. |
| `named-missing` | No task file has the named id. A named task that is missing never falls through to the running task. |
| `named-duplicate` | Two or more task files have the named id: an id resolves to exactly one file. |
| `named-broken` | The named task's file does not parse under its format. |
| `broken-file` | Nothing is named, and a task file does not parse. |
| `none-running` | Nothing is named, and no task reads `running`. |
| `many-running` | Nothing is named, and two or more tasks read `running`. |

## manifest.json

The set version and the SHA-256 of each file's raw bytes, for every file in this folder except `manifest.json` itself
and the Go files. Raw bytes are never line-ending-normalised: the CRLF case is the point.

```json
{
  "about": "The formats set's manifest: ...",
  "set": 1,
  "files": [
    { "path": "README.md", "sha256": "<64 hex>" },
    { "path": "examples/ask.json", "sha256": "<64 hex>" }
  ]
}
```

| Field | Meaning |
|---|---|
| `about` | What the file is. |
| `set` | The set's version, an integer. It goes up by one with every change to the set. |
| `files` | Every file, `path` (relative to `formats/`, forward slashes) and `sha256` (64 lowercase hex characters of its raw bytes), sorted by path. |

## The Go test

`formats_test.go` (the standard library and Bonsai's schema checker, `internal/schema`; it needs no reader) checks that
the manifest matches every file's bytes and lists every file and nothing more, sorted; that the CRLF case holds CRLF,
the lone-CR case CR alone, and the BOM cases start with `EF BB BF`, as checked out; that every rule in its list has a
case and every case has both outcomes, each reason code in the table above and each code in the table used; that
every schema is JSON with no duplicate key, declares draft 2020-12 and documents itself, and that each example
validates under the schema checker (`internal/schema`, moved out of this test in plan part 2 so Bonsai's code uses the
same one: it implements the keywords the schemas use, skips only `title`, `description` and `examples`, and fails on
any other keyword) and keeps the schema's field order; that each YAML or markdown example's format-1 value equals its
`<name>.json` (a table's or a message's, its frontmatter); that a message's trick file has its `message` outcome, an
accepted value valid under `bonsai.message/1` and a refused field one it has; that a run report's trick file that format 1
accepts has its `schema` outcome and the schema checker agrees (set 11); that every time field's description names its
zone, UTC or local, and its form (set 11, contract §23: a field whose pattern holds a date, or one the test lists); that each copy of a shared shape is its
home's (above); that the schemas
`embed.go` embeds are exactly the files in `schemas/`; and that no file in the set holds a private string: an absolute
home-folder path, a WSL drive path, a path from a home folder's tilde, a Windows drive letter, an email address or a
tailnet host name. `fixtures_test.go` checks the active-task fixtures and `answers.json` against the tables above:
every case has its `case.json` and the folders it names, none holds a `.git`, every section answers every case once in
order, every field has its form, every reason is in the table and every reason in the table is given, and rung 0's
answer is the function's named task. `compare_test.go` is the schema-compare test (below). Run them with `go test
./formats/`. Bonsai's reader (`internal/reader`) is held to every case's two outcomes by its own tests, and reads every
task file and `bonsai.yaml` of the fixtures as `case.json` says; from set 8, `internal/format`'s `TestMessageTrickFiles`
holds the message reader (`format.ReadMessage`) to each message case's third outcome.

## How the set changes

A change to this folder is one commit that also rewrites `manifest.json` (every hash, and `set` up by one), never a
silent edit. A new case is a new folder under `trick/`, its outcomes in `expect.json` (its format-0 outcome from the
frozen reader, run as above), and, for a new rule, the rule's name in the test's list. A new fixture is a new folder
under `active-task/` with an entry in every section of `answers.json`. Bonsai's `STATE.md` names the commit of each
new set, and readers that pin the set move to it on their own word.

**A schema changes only by additions inside a major** (contract §2.2), and `compare_test.go` holds it to that: it
compares each schema with the same file at the set's base commit, read with `git show`, and an addition is
- a new schema file;
- a new property at the end of an object's `properties`, its name added at the end of `required` (a schema describes
  what a writer writes, so a new field is required there; a reader still reads it as `null` when an older writer
  left it out, which is contract §2.2's "a new optional field");
- a change to `description`, `examples` or `title`, anywhere.

Anything else fails: a schema file removed; a property removed, renamed or moved; a `type`, `const`, `enum`,
`pattern` or bound changed, added or taken away; a name taken out of `required`; any other keyword added or removed.
The base is a constant beside the test (`schemaBase`): the commit of the set before, moved forward in each set's own
commit, and a second test holds it there: the base's `manifest.json` has `set` one below this one's. A checkout
without the base commit (no git, or a shallow clone) skips the test with that reason, except under CI (the `CI`
variable set) or under Bonsai's ladder (the `BONSAI_LADDER` variable set, which the ladder runner puts in every rung's
environment), where it fails; Bonsai's CI checks out the history it needs, and a rung of Bonsai's own ladder can never
pass by skipping.
