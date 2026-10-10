# The formats set

This folder holds Bonsai's formats as files any program can test against: a JSON Schema for each of the ten formats
of the formats contract (`design/contract.md` §2) the eight that set 4 added (Bonsai spec §5, §6, §10, §16 row
29) and the one that set 5 added (`help`, plan-5 5.1.10), one example document for each, the **trick files**: small, deliberately tricky YAML and markdown files, each with
the one right outcome a reader must reach (accepted with this value, or refused for this reason), and contract §13's
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
                                status, lock; workspace, pack, tasks, sessions, memory, error, check, changes; help
  examples/<name>.json      one valid document per format, as a reader returns it
  examples/<name>.md|.yaml  for task, run, state, tasks, sessions, memory (markdown) and labels, lanes,
                            workspace, pack (YAML): the source file, also a case
  trick/yaml-1/<case>/case.md|.yaml   a case for one rule of the format-1 YAML grammar (contract §2.4)
  trick/yaml-0/<case>/case.md|.yaml   one of today's format-0 oddities
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
result of `init`, `update` and `unlink` with `--json`); and, from set 5, `help` (`bonsai.help/1`, `bonsai --help
--json`: every command word with its flags, exit codes and examples, the exit codes every word shares, and every `error`
word with its meaning and usual `who`, so an agent learns the tool without reading a page). Each documents itself: a top-level `description` (what the
format is for, who writes and reads it, where it lives) and a `description` and `examples` on every property. Later
steps add their commands' outputs (`asks`, `logs`, `ladder`) the same way, each a new schema file: an addition.

- **A schema describes what a writer writes** (contract §2.2): every field the version knows, in the contract's fixed
  order (the order of `properties`), all of them `required`, with `null` or `[]` where a field does not apply. A
  reader is more lenient: a missing field reads as `null`, an unknown field is kept, so `additionalProperties` stays
  open.
- **Lists.** A closed list (it never grows inside a major) is an `enum`; an open list (it may grow; readers show an
  unknown value as `other`) is a plain string, its known values named in the description. These schemas are the one
  home of every list in the formats (contract §2.2, format review R2.8): Bonsai's code embeds them and keeps no
  second copy. **One kind of open list lives in Bonsai's code instead:** a list of fixed words that code reads at many
  places, the `error` object's `code` and `check`'s finding and warning `code` (and, from step 5.2, the log's
  events). Its known words live in one Go table, their one home, which `bonsai check --schema <format>` prints (and the
  reference page of lists, `docs/reference/lists.md`, generated from the code by `go generate ./...`); the schema's description names those commands and does not
  copy the words, and the words in its examples are illustrative.
- **Copies held to their home.** The schema checker has no `$ref`, so a shape two schemas share is written in both,
  and a test holds the copy equal to its home: the `error` object, inline in `status`, `check` and `changes`, is
  `error.schema.json`'s (its `required` and `properties`, word for word), and so is every `next` object (a finding's,
  a warning's, the plugin step's); a ladder rung's `kind` in `bonsai.yaml` is the ladder result's closed list.
- **Spec §16's additions** are in from the start: the lock's `declares` (row 18), the ladder's third rung kind
  `ver-git` (row 23), and `status --json`'s additions (row 24), which arrive as values, not new fields: a `needs`
  entry of kind `plugin` for a pack not installed on the machine, the Claude Code floor as a `needs` entry of kind
  `tool`, plugin drift and a swapped `bonsai` binary as `problems`; `--line` prints another output, not a field.
- **Typed open**, because the sources do not fix their shape (plan part 0: a shape still unclear is typed open):
  - `ladder`: a rung's `status` (a string; today's values are `green`, `red`, `pending`, `skipped`, `error`; the
    contract lists none), a rung's `tests` (an object or `null`; today's inner fields are described, none required),
    `leftovers` (an object, a text or `null`, as today's three shapes).
  - `status`: `mode` (the contract shows only `offline`), each `documents` entry (only `kind` and `from` required:
    contract §7.3 says `format` is absent for a pack's kind and a kind has `path` or `file`), and `checks` (an object
    or `null`).
  - `lock`: each pack's `declares` (an object: the four kinds it holds are fixed, their key names and inner layout are
    not; format review 6.3).
  - `ask`: `data` (an object or `null`: a pack type's own payload).
  - `workspace`: a ladder rung's `ratchet` and `capture` have no `type` at all, so any value reads: spec §6 names
    them as today's rung shape, whose form is not in Bonsai's sources, and a type given now could only be widened
    in a later major. The ladder runner (step 5.4) reads them and fixes their shape, in a later major if it must. A
    rung's `tests` is an open string: spec §9 names its three forms (TAP, `go test -json`, JUnit) but not their
    words, which step 5.4 fixes.
  - The workspace id is `ws-` and 26 lowercase base32 characters (contract §3); which base32 alphabet is not fixed,
    so the schemas take `[a-z0-9]{26}`.
- **A name not invented.** Spec §16 row 27 adds the `bonsai` binary's own path and SHA-256 to the log's
  `session_start` records, but neither the spec nor the format review (4.2) fixes the two field names. They are not
  in `log.schema.json`; they are added, as an addition inside `bonsai.log/1`, once the names are fixed.

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
  label definition writes every field, `grants: false` among them. Rohan, 8 Oct: label-definition fields stay
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
  its repository's root. A lock written before set 4 has none, which a reader reads as `null`, the root (contract
  §2.2's "a new optional field"): an engine that compares it with `bonsai.yaml` sees such a pack's folder as changed.
  Until step 5.1.5's engine writes it, Bonsai's lock writer leaves it out of a pack it has none for.
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
  `sessions.md`'s rows as `sessions`, each with `kind` (`session` or `subagent`), and its hours as `hours`, sessions
  and subagent runs apart (contract §7.5: never added), in hours to one decimal place.
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
  a settings line's and a code item's `kind`, the plugin step's `result`, the lock's word in `changes`, the `error`
  and finding codes (above), and a ladder rung's `tests`.
- **`help`, set 5.** One document, `bonsai --help --json`, read from the same tables the human `--help` and the
  behaviour are made from (the word registry in `cmd/bonsai`, `format.ExitCodes`, `format.ErrorWords`), so it cannot
  drift; a test holds it to every word, flag and error word the tables have, and to this schema. Only the bare
  `bonsai --help --json` prints it: a word's own `--help` stays text, with or without `--json`. `words` lists the sub-words
  (`hook guard`) as entries of their own, a sub-word not built yet with `later` set and its text fields empty; `later`
  on a word or a flag names the step that builds it. `exit_codes` is the codes every word shares and `hook`'s own;
  each word's `exits` says which it returns and what each means for it. `error_words` is `format.ErrorWords`, an
  open list like the `error` object's `code` it describes. An addition (a new schema file), so the schema-compare test
  passes unchanged.
- **Bounds in set 4**, chosen so a later set never needs to lower one: `keep_days` and `keep_newest` take 0 and up
  (0 cleans as soon as the protections allow; they always win); a rung's `timeout_s` takes 1 and up.
- **Patterns in set 4**, each where a source or Bonsai's own reader fixes the form: a pack id (a lower-case letter,
  then lower-case letters, digits and dashes, as Bonsai's reader of `pack.yaml` and `bonsai.yaml` holds it); a
  document kind's name (a key's form, contract §2.4, since it is a key in `bonsai.yaml`); a memory id (`M-` and a
  slug); a fixed word (`error` and finding codes); the session id's first 8 characters; a pack's `source` and `ref`,
  which never start with `-` (git would read an option); a `never_edit` path, which as the deny rule `Edit(<path>)`
  holds no parenthesis and does not start with `~`.

## The examples

`examples/<name>.json` is one made-up, valid document per format, stored as the JSON a reader returns; its schema
validates it, and its fields keep the schema's order. For the ten YAML and markdown formats (`task`, `labels`,
`lanes`, `run`, `state`, `workspace`, `pack`, `tasks`, `sessions`, `memory`) the source sits beside it (`<name>.md`
or `<name>.yaml`) and is also a case in `expect.json`: its format-1 value is exactly `<name>.json`, except for the
two tables, whose format-1 value is their frontmatter, `<name>.json` without the fields a reader reads from the body
(`active` and `tasks`; `sessions` and `hours`), which come last. So the test checks every schema without a reader,
and a reader proves the YAML. The pack files (`labels.yaml`, `lanes.yaml`, `pack.yaml`) document themselves in `#`
comments, as every pack file does (contract §2.8), and `workspace.yaml` has a comment on every line, as `init`
writes `bonsai.yaml` (format review 1.4); the markdown examples carry the one-line pointer on their `format:` line
that a file made from a template carries, and the tables a line saying they are generated.

**The two tables' bodies**, which a JSON Schema cannot describe; the writer writes them so, and a reader reads them so:
- `tasks.md`: after the frontmatter, the line `Active task when none is named: <id>`, or `Active task when none is
  named: none (<why>)`; a blank line; then the table `| Task | Title | Status | Lane | Started | Finished |`, its
  `|---|` line, and one row per task, newest id first.
- `sessions.md`: after the frontmatter, the table `| Session | Kind | Task | Role | Model | Start | End | Minutes |`,
  its `|---|` line and one row per session or subagent run; a blank line; the line `Hours per task and role (a
  subagent run lies inside its session's row, so the two are never added):`; a blank line; then the table `| Task |
  Role | Kind | Hours |`, its `|---|` line and one row per task, role and kind.
- In both, a `null` is an empty cell, a `|` inside a cell is written `\|`, and nothing else is escaped.
- A task file that does not parse under its format gives no row in `tasks.md`: `bonsai check` reports it as a
  finding (spec §6), and the active line says none, naming it (contract §13: such a file makes the answer none).

`examples/status.json` and `examples/changes.json` print absolute paths, as `status --json` and `init`, `update` and
`unlink` with `--json` do (printed locally, never forwarded). The contract's own example (§12) uses a path under a
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
- The CRLF case (`trick/yaml-1/lines-crlf/case.md`) ends every line in CRLF, the lone-CR case
  (`trick/yaml-1/lines-cr/case.yaml`) every line in a CR alone, with no LF anywhere, and two cases start with a BOM
  (`bom-frontmatter`, `bom-definition`). The repo's `.gitattributes` sets `formats/** -text`, so no checkout changes
  their bytes.

A case's folder name starts with its rule (`key-reserved-on` is a `key-reserved` case). The rule, a sentence saying
what it tests, and both outcomes are in `expect.json`.

## expect.json

One file, UTF-8, LF, giving every case two outcomes. Example (one case, shortened):

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
| `cases` | One entry per case, sorted by `path`; every file under `trick/` and the ten example sources has one. |
| `path` | The input file, relative to `formats/`, forward slashes. |
| `rule` | The rule the case tests: one of the rule names the Go test lists (`rules` in `formats_test.go`), or `example`. |
| `settled_by` | `rule` when contract §2.4's words decide the format-1 outcome; `libraries` when they do not and §2.4's own test decides it: accepted only if a YAML 1.1 and a YAML 1.2 library both read the file and read the same value, otherwise refused (plan part 0). |
| `about` | One sentence: what the case tests. |
| `format0` | What the studio's frozen format-0 reader reaches, which reads every file as format 0, `format:` line or not: `{"outcome": "accepted", "value": <JSON>}` or `{"outcome": "refused", "message": "<the reader's own message>"}`. |
| `format1` | What a reader that dispatches as §2.4 says reaches: `{"outcome": "accepted", "value": <JSON>}`; `{"outcome": "refused", "reason": "<code>"}` with a code from the table below; or `{"outcome": "format-0"}` when the file has no top-level `format:` key, which sends it to format 0. |
| `value` | The YAML file's mapping, or for markdown the frontmatter's mapping (the body is not compared), as JSON. Under format 1 a date stays text and an integer has at most 15 digits. |

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

**A case folder** holds `case.json` and one folder per project: `main/` (always), `other/` (a second project), and,
in a worktree case whose worktree's files differ from main's, `worktree/`. Each project folder holds `bonsai.yaml`, whose `documents.task` is `work/tasks`,
and its task files. No case holds a `.git`: a case that needs git is a layout each reader's test builds.

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
`<name>.json` (a table's, its frontmatter); that each copy of a shared shape is its home's (above); that the schemas
`embed.go` embeds are exactly the files in `schemas/`; and that no file in the set holds a private string: an absolute
home-folder path, a WSL drive path, a path from a home folder's tilde, a Windows drive letter, an email address or a
tailnet host name. `fixtures_test.go` checks the active-task fixtures and `answers.json` against the tables above:
every case has its `case.json` and the folders it names, none holds a `.git`, every section answers every case once in
order, every field has its form, every reason is in the table and every reason in the table is given, and rung 0's
answer is the function's named task. `compare_test.go` is the schema-compare test (below). Run them with `go test
./formats/`. Bonsai's reader (`internal/reader`) is held to every case's two outcomes by its own tests, and reads every
task file and `bonsai.yaml` of the fixtures as `case.json` says.

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
commit, and a second test holds it there: the base's `manifest.json` has `set` one below this one's. A checkout without the base commit (no git, or a shallow clone) skips the test with that reason, except under
CI (the `CI` variable set), where it fails; Bonsai's CI checks out the history it needs. From step 5.4 the same test
is a rung of Bonsai's own ladder.
