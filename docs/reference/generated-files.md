# Reference: generated files

What Bonsai and its agents generate inside a project, who writes each kind, how long it is kept, what is never
cleaned and when cleaning happens. Bonsai deletes only what this page says it may.

**This page is generated. Never edit it by hand.** It is written by `go generate ./...` from the table
`format.GeneratedKinds` (internal/format/generated.go); a test rebuilds it and fails on any difference, naming that
command. Change a kind in the table, run `go generate ./...` from the repository's root, and commit the page with the
change. At step 5.5 the `generated-files` skill of the `base` pack takes these words, and its table stays generated
from the same Go table.

## The kinds

### `log`

- What: the log, one file per session.
- Where it lives: .bonsai/local/log/.
- Written by: the recorder, from the hooks, one line at a time.
- Default: 30 days after a file's last line.
- Never cleaned: an open session's file, or one holding an ended session or subagent run with no row yet in .bonsai/sessions.md.
- Cleaned: at a session's end, after its session_end line, within a budget of 1 second, oldest first; what the budget leaves waits for the next end.
- Rule in bonsai.yaml: `generated.log` takes `keep_days` and `keep_newest`.

### `asks`

- What: questions for a person and their answers.
- Where it lives: .bonsai/local/asks/.
- Written by: bonsai ask and bonsai answer (and the ladder runner from step 5.4, as its Bless), one day file a day.
- Default: kept.
- Never cleaned: a day file holding an open ask, or the answer or withdrawal of an ask whose filing stays (else it would read open again).
- Cleaned: at a session's end, within the same budget; by default nothing, as the default keeps.
- Rule in bonsai.yaml: `generated.asks` takes `keep_days` and `keep_newest`.

### `ladder`

- What: ladder results, one per task.
- Where it lives: .bonsai/local/ladder/.
- Written by: the ladder runner (step 5.4), one file per task.
- Default: 7 days after the result's finished.
- Never cleaned: the result of a task that is not done or cut (a task not found, or whose file does not read, counts as not done).
- Cleaned: at a session's end, and by the ladder runner after each run.
- Rule in bonsai.yaml: `generated.ladder` takes `keep_days` and `keep_newest`.

### `run`

- What: run reports, committed history.
- Where it lives: the run reports' folder (documents.run).
- Written by: the agent that did the work, in the run report format.
- Default: kept.
- Never cleaned: any, by Bonsai: past a rule, bonsai check lists them and a person deletes them.
- Cleaned: never, by Bonsai.
- Rule in bonsai.yaml: `generated.run` takes `keep_days` and `keep_newest`.

### `sessions`

- What: rows of the sessions table.
- Where it lives: .bonsai/sessions.md.
- Written by: bonsai check --write, from the log.
- Default: kept.
- Never cleaned: rows of a task that is not done or cut (none is never protected), and rows whose span is still in the log (the next write would add them again).
- Cleaned: in bonsai check --write, the only writer of the table.
- Rule in bonsai.yaml: `generated.sessions` takes `keep_days` and `keep_newest`.

### `tasks`

- What: the tasks table.
- Where it lives: .bonsai/tasks.md.
- Written by: bonsai check --write, from the task files.
- Default: a rebuild: nothing to clean.
- Never cleaned: -.
- Cleaned: never: the table is rebuilt whole.
- Rule in bonsai.yaml: none; this kind takes no rule.

## The rules

1. **The protections come first.** A protected file or row is never cleaned, whatever the rule says, and it still
   counts among the newest. The protections are the "Never cleaned" line of each kind above.
2. **`keep_days`** cleans a file or row older than that many days. Age is a log or asks file's last record (its
   modification time when no record reads), a ladder result's `finished` (its modification time when the result
   does not read), or a row's end.
   A file is cleaned only when both its modification time and its content say it is older than the rule.
3. **`keep_newest`** keeps only that many of the newest of its kind and cleans the rest.
4. Either one cleans. `null` for both keeps everything. A kind or the whole `generated:` section left out of
   bonsai.yaml takes the defaults above; a rule that does not read cleans nothing of its kind.
5. **Only Bonsai's own files** are cleaned, by the names it writes, regular files only, inside the kind's folder.
   A link is never followed, and nothing is cleaned in a folder that is itself a link. Anything else in the folder
   is left alone and never named.
6. **A `clean` record** goes in the log for every file or row cleaned, naming its path (a row: `.bonsai/sessions.md#`
   and its key) and the rule that cleaned it, written once the file is deleted (a row: once the table is written).
   A delete that finds the file gone writes nothing. A file another process holds open (Windows), or one written to
   since it was judged, is left for a later run.
7. **A session's end** cleans ladder results, asks and the log, oldest first, within 1 second; what it leaves waits
   for the next end. A cleaning killed between a delete and its record leaves the file gone with no record.
8. **Run reports are never deleted by Bonsai.** Past their rule, `bonsai check` lists them as a warning and a person
   deletes them.
