# The formats set

This folder holds Bonsai's formats as files any program can test against: a JSON Schema for each of the ten formats
of the formats contract (`design/contract.md` §2), one example document for each, and the **trick files**: small,
deliberately tricky YAML and markdown files, each with the one right outcome a reader must reach (accepted with this
value, or refused for this reason). `manifest.json` fixes every file's bytes.

**Who reads it.** Bonsai's own readers test against it (its Go code embeds the schemas from plan part 2 on). Any other
reader of these formats, the studio's among them, tests against the set of the Bonsai release it pins. Nothing here
reads or writes a project, a log or an ask: these are tests and their answers, not code.

```
formats/
  README.md                 this page
  expect.json               the expected outcome of every case (below)
  manifest.json             the set version and the SHA-256 of every file's bytes (below)
  formats_test.go           the Go test that checks the set is whole (below; not in the manifest)
  schemas/<name>.schema.json    one JSON Schema per format: task, labels, lanes, run, state, log, ask, ladder,
                                status, lock
  examples/<name>.json      one valid document per format, as a reader returns it
  examples/<name>.md|.yaml  for task, run, state (markdown) and labels, lanes (YAML): the source file, also a case
  trick/yaml-1/<case>/case.md|.yaml   a case for one rule of the format-1 YAML grammar (contract §2.4)
  trick/yaml-0/<case>/case.md|.yaml   one of today's format-0 oddities
```

## The schemas

One file per format, JSON Schema draft 2020-12, named for the format: `task` (`bonsai.task/1`, contract §4), `labels`
(§5.2), `lanes` (§6), `run` (§7.1), `state` (§7.2), `log` (§8.1), `ask` (§9.1), `ladder` (§11), `status` (§12),
`lock` (§14). Each documents itself: a top-level `description` (what the format is for, who writes and reads it,
where it lives) and a `description` and `examples` on every property.

- **A schema describes what a writer writes** (contract §2.2): every field the version knows, in the contract's fixed
  order (the order of `properties`), all of them `required`, with `null` or `[]` where a field does not apply. A
  reader is more lenient: a missing field reads as `null`, an unknown field is kept, so `additionalProperties` stays
  open.
- **Lists.** A closed list (it never grows inside a major) is an `enum`; an open list (it may grow; readers show an
  unknown value as `other`) is a plain string, its known values named in the description. These schemas are the one
  home of every list in the ten formats (contract §2.2, format review R2.8): Bonsai's code embeds them and keeps no
  second copy.
- **Spec §16's additions** are in from the start: the lock's `declares` (row 18), the ladder's third rung kind
  `ver-git` (row 23), and `status --json`'s additions (row 24), which arrive as values, not new fields: a `needs`
  entry of kind `plugin` for a pack not installed on the machine, the Claude Code floor as a `needs` entry of kind
  `tool`, plugin drift and a swapped `bonsai` binary as `problems`; `--line` prints another output, not a field.
- **Typed open**, because the sources do not fix their shape (plan part 0: a shape still unclear is typed open):
  - `ladder`: a rung's `status` (a string; today's values are `green`, `red`, `pending`, `skipped`, `error`; the
    contract lists none), a rung's `tests` (an object or `null`; today's inner fields are described, none required),
    `leftovers` (an object, a text or `null`, as today's three shapes).
  - `status`: `mode` (the contract shows only `offline`), each `documents` entry (only `kind` and `from` required:
    contract §7.3 says `format` is absent for a pack's kind and a kind has `path` or `file`), each `needs` entry (only
    `kind` required: a pack need carries `id`, `source`, `version` and a tool need `name`, `version`; spec §7 names
    the Claude Code need by `id` where the contract's and the format review's examples use `name`), and `checks` (an
    object or `null`).
  - `lock`: each pack's `declares` (an object: the four kinds it holds are fixed, their key names and inner layout are
    not; format review 6.3).
  - `ask`: `data` (an object or `null`: a pack type's own payload).
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
  label definition writes `grants: false` rather than leaving it out.
- `status --json` exits 3 with only `format`, `bonsai` and `problems` filled, so every other field also takes `null`.
- Patterns are added only where a source fixes a form: ids (task, run, workspace, machine folder, UUID), dates and
  times, label names, hashes (a commit in lowercase hex, 40 characters where the lock resolves one; a 64-character
  SHA-256; today's 16-character input hash), the Remote Control id, and project-relative paths (no leading slash, no
  backslash, no drive colon).
- `folded-deeper` holds for `>-` as for `>`: the strip indicator changes only the end of the value.

## The examples

`examples/<name>.json` is one made-up, valid document per format, stored as the JSON a reader returns; its schema
validates it, and its fields keep the schema's order. For the five YAML and markdown formats (`task`, `labels`,
`lanes`, `run`, `state`) the source sits beside it (`<name>.md` or `<name>.yaml`) and is also a case in
`expect.json`: its format-1 value is exactly `<name>.json`. So the test checks every schema without a reader, and a
reader proves the YAML. The pack files (`labels.yaml`, `lanes.yaml`) document themselves in `#` comments, as every
pack file does (contract §2.8); the markdown examples carry the one-line pointer on their `format:` line that a file
made from a template carries.

`examples/status.json` prints absolute paths, as `status --json` does (printed locally, never forwarded). The
contract's own example (§12) uses a path under a home folder; this one uses `/srv/projects/example` and
`/srv/bonsai-home`, made-up absolute paths that the set's private-string check (below) does not flag.

## The trick files

Each case is one folder holding one file, `case.md` (markdown with YAML frontmatter) or `case.yaml` (a YAML definition
file), with made-up content only.

- `trick/yaml-1/<case>/`: one case per rule of contract §2.4, accepted or refused. These start with a `format:` line,
  as a real file would, except the dispatch cases, which test where `format:` sits (first, not first, nested, absent).
- `trick/yaml-0/<case>/`: today's format-0 oddities (contract §2.4: a plain value holding `: `, unquoted hashes,
  `TRUE`, the last of two duplicate keys, document markers skipped). They have no `format:` line, so a format-1
  reader sends them to format 0.
- The CRLF case (`trick/yaml-1/lines-crlf/case.md`) ends every line in CRLF, and two cases start with a BOM
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
| `cases` | One entry per case, sorted by `path`; every file under `trick/` and the five example sources has one. |
| `path` | The input file, relative to `formats/`, forward slashes. |
| `rule` | The rule the case tests: one of the rule names the Go test lists (`rules` in `formats_test.go`), or `example`. |
| `settled_by` | `rule` when contract §2.4's words decide the format-1 outcome; `libraries` when they do not and §2.4's own test decides it: accepted only if a YAML 1.1 and a YAML 1.2 library both read the file and read the same value, otherwise refused (plan part 0). |
| `about` | One sentence: what the case tests. |
| `format0` | What the studio's frozen format-0 reader reaches, which reads every file as format 0, `format:` line or not: `{"outcome": "accepted", "value": <JSON>}` or `{"outcome": "refused", "message": "<the reader's own message>"}`. |
| `format1` | What a reader that dispatches as §2.4 says reaches: `{"outcome": "accepted", "value": <JSON>}`; `{"outcome": "refused", "reason": "<code>"}` with a code from the table below; or `{"outcome": "format-0"}` when the file has no top-level `format:` key, which sends it to format 0. |
| `value` | The YAML file's mapping, or for markdown the frontmatter's mapping (the body is not compared), as JSON. Under format 1 a date stays text and an integer has at most 15 digits. |

## Reason codes

The format-1 refusals, by short code. A file can break more than one rule, so a reader reports one code by this order:
dispatch first (`format-not-first`, whatever else the file holds); then the first problem reading top to bottom; on
one line, the key before its value; and of two codes that fit one problem, the one listed first below.

| Code | Refused because |
|---|---|
| `format-not-first` | A top-level `format:` key is not the file's first key (comments and blank lines may come before it). |
| `tab-indent` | A tab in a line's indentation. |
| `doc-marker` | A `---` or `...` line anywhere but a frontmatter's own two markers: a YAML file has none, and a frontmatter block holds none. |
| `line-not-read` | A line the grammar does not consume: a block sequence at its key's own indent, a sequence item where a mapping key belongs. |
| `key-complex` | A complex key (`? key` then `: value`). |
| `key-quoted` | A quoted key, single or double. |
| `key-merge` | The merge key `<<`. |
| `key-form` | Any other key that is neither `[a-z][a-z0-9_]*` nor a label name `<namespace>.<name>` (contract §5.1): `Title`, `done-when`. |
| `key-reserved` | A key that is `y`, `n`, `yes`, `no`, `on`, `off`, `true`, `false` or `null`. |
| `key-twice` | A key its mapping already holds. |
| `seq-dash-space` | A block sequence item whose `-` is not followed by exactly one space. |
| `anchor` | An anchor (`&name`). |
| `alias` | An alias (`*name`). |
| `tag` | A tag (`!name`, `!!str`). |
| `flow-mapping` | A flow mapping with content (`{a: 1}`); `{}` alone is an empty map. |
| `flow-nested` | A `[` or `{` that opens inside a flow sequence. |
| `flow-multiline` | A flow sequence that does not close on its own line. |
| `block-indicator` | A block scalar header other than `\|`, `\|-`, `>` and `>-`: `\|+`, `>+`, an indentation indicator such as `\|2`. |
| `block-hash-line` | A line inside a block scalar whose first character after the indentation is `#`. |
| `folded-deeper` | In a `>` or `>-` block scalar, a line indented deeper than its first line. |
| `quoted-multiline` | A quoted scalar that does not close on its own line. |
| `bad-escape` | In double quotes, a backslash before anything but `"`, `\`, `n`, `r` or `t`. |
| `unescaped-quote` | In double quotes, an unescaped `"` inside: the first unescaped `"` after the opening one closes the scalar, and the rest of the line before any comment holds another `"`. |
| `after-quote` | Anything but spaces and a comment after a quoted scalar's closing quote (with no further `"`, which is `unescaped-quote`). |
| `quote-this-value` | A plain scalar that fits no row of §2.4's table (`0755`, `1.2.3`, `True`, `on`, a value ending in `:`, a 16-digit integer), or, inside a flow sequence, a plain item holding `]` or `}`. A flow sequence runs from its `[` to the last `]` on its line, and its items are split at commas outside quotes. |

## How the outcomes were found

- **Format 0** means exactly what the studio's `tools/lib/yaml.mjs` reads (contract §2.3, §2.4), as the studio froze it
  for format 0 at commit `4a05eac`. Each case was run through that file, taken read-only with `git show`, the way the
  studio's callers read files: `readFileSync(path, 'utf8')` (so a BOM and CRLF reach the reader), then `parseYaml` for
  a YAML file or `parseFrontmatter(...).data` for markdown. `format0` holds its value or its exact error message.
- **Format 1** outcomes follow contract §2.4's grammar rule by rule. Where §2.4's words do not settle a value
  (`settled_by: libraries`: two single quotes inside single quotes, the value of a block scalar, the CRLF case's block
  scalar at the end of a frontmatter), two libraries were run and both read the same value: PyYAML 5.4.1 (YAML 1.1)
  and the npm package `yaml` 2.8.3 (YAML 1.2). A markdown file's frontmatter is the lines between its markers, each
  with its own line ending. Dates are settled by §2.4 itself: they stay text, though a YAML 1.1 library reads a date.

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

`formats_test.go` (standard library only; a test-only package that needs no reader) checks that the manifest matches
every file's bytes and lists every file and nothing more, sorted; that the CRLF case holds CRLF and the BOM cases start
with `EF BB BF`, as checked out; that every rule in its list has a case and every case has both outcomes, each reason
code in the table above and each code in the table used; that every schema is JSON with no duplicate key, declares
draft 2020-12 and documents itself, and that each example validates under a small checker (it implements the keywords
the schemas use, skips only `title`, `description` and `examples`, and fails on any other keyword) and keeps the
schema's field order; that each YAML or markdown example's format-1 value equals its `<name>.json`; and that no file
in the set holds a private string: an absolute home-folder path, a WSL drive path, a path from a home folder's tilde,
a Windows drive letter, an email address or a tailnet host name. Run it with `go test ./formats/`.

## How the set changes

A change to this folder is one commit that also rewrites `manifest.json` (every hash, and `set` up by one), never a
silent edit. A new case is a new folder under `trick/`, its outcomes in `expect.json` (its format-0 outcome from the
frozen reader, run as above), and, for a new rule, the rule's name in the test's list. A schema changes only by
additions inside a major (contract §2.2). Bonsai's `STATE.md` names the commit of each new set, and readers that pin
the set move to it on their own word.
