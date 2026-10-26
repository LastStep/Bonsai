# The first existing project: ZenGarden-AI-API on Bonsai 1.0

For the director (T-6001). Nothing here asks anything of you but the last section. All times UTC. The evidence: two
auditors (opus), reports `M-2026-10-26-851dff1c` (Bonsai's machinery) and `M-2026-10-26-a781cee9` (the packs' steering
and the plan's two tasks), `bonsai message --show <id>`; "L<n>" is line n of the project's
`.bonsai/local/log/s-c8c280da-….ndjson`. The plan: `records/plans/P-T-6001.md`; the log of the work:
`records/runs/R-2026-10-26-T-6001-zengarden-pilot.md`.

## In short

**The link and the transition: a success. The proof: weaker than it reads.**
- **The link and the cleanup.** On a private Python service with six months of history and an agent system of its
  own, Bonsai linked in one rehearsed script, kept the project's CLAUDE.md byte for byte, and proved 305 tests on
  every climb. The project's first session then:
  - retired the old system exactly as the approved table said: one moves-only commit, the guides brought up to
    date, nine memory notes each traced to its source line;
  - took one real fix (CI's image push waits for the tests) through builder, verifier and a green climb on `main`;
  - pushed nothing.
- **Two cracks in the proof.** The guard judged none of the session's writes: every agent wrote by shell, which it
  does not judge. And rung 0 judges nothing on `main`, where the cleanup was built.
- **One process gap.** The commit that landed was not the one the verifier passed.
- **Bonsai issues.** Twelve: two known, ten new, none blocking. They are listed below for the register.

## Timeline

| When (26 Oct) | What |
|---|---|
| 07:27-07:50 | Three auditors probe ZEN and ZenGarden-AI-API; the director chooses "A, AI-API first" |
| 07:51-08:06 | Plan written; review round 1 FAIL (3 blocking, all held), round 2 PASS; approved |
| 08:07-08:37 | Rehearsal on a clone with a stamped build: link, climbs on `main` and in a worktree, all green |
| 08:38-08:43 | Security review of `link.sh` and the kickoff: PASS |
| 08:57 | The director's real link stopped at `bonsai check` (`missing-path`, CLAUDE.md line 170); one `sed`, run again: linked (`d6355c7`) |
| 09:02-09:23 | The project's first session: T-0001 and T-0002, landed on local `main` (`511a3e1`), 17 commits, not pushed |
| 10:00-10:07 | Two auditors read the project's log, records and history |

## The criteria

| | What | Verdict | Evidence |
|---|---|---|---|
| A | The link | pass | Base at `4662da5e` and workflow at `0290f701` match across `bonsai.yaml`, the lock and the plugins. The hook lines are on every event; the `.env` edit denies are in place; `autoMemoryEnabled: false`. CLAUDE.md is identical outside the block (apart from the one `sed` and a blank line before the block). One stop on the real run: the `missing-path` the rehearsal could not see (Bonsai issue 1 below) |
| B | The guard | pass on what it judged; partial on coverage | 159 decisions; one refusal, right by its rule (L39, `git mv` in a loop, `delete-unnamed`). Nothing written to `bonsai.yaml`, the lock, `.claude/**`, `.env*` or `.bonsai/local` (git history, file times). But 91 of the 159 read `shell-not-judged`: every write was a shell command, and no Edit or Write ran all session, so the edit-time checks (`time-not-now`, status moves, formats, grants) judged nothing |
| C | The log | pass | 576 lines in 4 files all parse; every call has its guard record and an end; 7 named subagents paired. The one secret-shaped value found reads `[redacted]` (the redactor worked). No `.env` value appears in the log, the messages or `work/`. It keeps a shell command's first word only, so a read of a `.env` file by shell could not be seen |
| D | Messages | partial | Every brief went by `bonsai message`. 3 of 4 auditors ended by hand-back and never reported, so their threads read `open`; one stray brief with the body "test" is open too |
| E | Documents | pass; task moves partial | `bonsai check`: no finding. T-0001 was born `running` (never `todo`). T-0002 went `running` to `done` with no `verify`; nothing enforces moves where status writes are the agents' |
| F | The ladder | pass, with a gap | Six climbs, all by the installed 1.0 (`f9b1cd76…`); pytest read whole as JUnit, 305 seen each time. The landing climb on `main` at `93d54e9` was green, `git.dirty` false. But rung 0 reads `checked: 0` on `main`, by design, and does not say so (`not_judged` null). Each climb overwrites the last result, so the branch climb's `git.sha` survives only as a sha256 in the log |
| G | Roles and briefs | partial | Every agent ran in its role (4 auditors on sonnet, a builder on sonnet, a verifier on opus, the orchestrator on opus), each briefed by message. That the helpers wrote nothing rests on git alone: the guard ruled on none of their shell calls |
| H | Run reports, STATE, one task at a time | partial | The T-0001 report was opened before its first move. The builder's first T-0002 report was wrong three ways (started late, finished 10 minutes in the future, `outcome: done`); rung 0 went red on it and it was fixed. The orchestrator's T-0002 report was first committed at landing. Only one task read `running` at a time. STATE was rewritten |
| I | T-0001 against the approved table | **pass** | `251d4c9`: 18 renames at 100%, no line changed. Every file is where the table says; FieldNotes, README, API.md, `prompts/` and `tools/` untouched; the archive README and the guides each in their own commit. Three memory notes traced to their source (one to a commit message, not a file). Dropped facts listed. The block between CLAUDE.md's markers has the same sha256 before and after |
| J | T-0002 end to end | partial | Builder in a plain worktree; green climbs; fresh verifier PASS; fast-forward; green climb on `main`; `origin/main` still `d24ecdf` (nothing pushed). But the verifier passed `fb95bc1` and the builder committed `814c566` (a run-report change) while it worked, so the landed commit was not the one verified. The builder's first two green climbs were at its code commit, not its HEAD |
| K | The director's side | smooth | The director did by hand: the link script twice plus one `sed`, Claude Code opened once to trust, `bonsai update --yes`, and the kickoff. Two questions in the session (CLAUDE.md's outline, which fix to take); 21 minutes of session time |

## Bonsai issues found

For `records/bugs.md`. Ids are given when they are filed: another session holds uncommitted rows there, so the ids
wait (see the last section).

| # | What goes wrong | Where | Known? |
|---|---|---|---|
| 1 | `missing-path` judges a backticked path only when its first folder exists. The same CLAUDE.md is clean in a clone and a finding in the real checkout (whose ignored `logs/` folder exists), so a rehearsal on a clone cannot see it | `internal/engine/checkdocs.go:681-684` (pointed) | new |
| 2 | `missing-path` flags a git-ignored path a project must name to its agents (`.venv/bin/python`, a runtime log); the project had to reword its instructions | `internal/engine/checkdocs.go:628-629` (pointed) | new |
| 3 | `never_edit` writes Edit denies only: no Read deny, and the guard does not judge shell reads, so an agent can read the `.env` files it walls | the project's `.claude/settings.json`; the field's docs (pointed) | new |
| 4 | Adding a second pack after the link makes `update` change Bonsai's guard line and ask `--allow-exec`; `init` takes one pack. The script wrote the whole `bonsai.yaml` before `init` (with a hand-made workspace id) to avoid it | the rehearsal (said; not reproduced on its own) | new |
| 5 | `time-not-now` judges only Edit and Write: a run report written by shell with `finished` 10 minutes in the future passed, and `bonsai check` has no check of future times | `internal/guard/rules.go:137` (pointed); the builder's report at L436 | new |
| 6 | Rung 0 judges no commit on the base branch and says nothing of it (`checked: 0`, `not_judged` null): work built straight on `main` (T-0001 here) is never judged by rung 0 | `internal/rung0/rung0.go:156`, `git.go:146-155` (pointed) | new (the silence) |
| 7 | A relative `--pointer` from a worktree is read against the worktree, while a ladder result lives in the main checkout's `.bonsai/local`; the builder had to send an absolute path | `internal/message/pointer.go:70-79` (pointed) | new, near BUG-109 |
| 8 | Each climb overwrites the task's result: a branch climb's `git.sha` is lost once `main` is climbed, so the landing's "green at the branch HEAD" survives only as a sha256 in the log | the ladder's result file (pointed) | new |
| 9 | The workflow pack lets a commit land that the verifier did not see: nothing ties the verifier's PASS to the landed sha | the workflow pack's orchestrate and verification-ladder skills | new (pack) |
| 10 | Helpers that end with Claude Code's hand-back leave their threads `open`; nothing holds them to `workflow:reporting` | the log (L222, L234, L255) | new (pack and harness) |
| 11 | Words: `base:message` reads as if `--json` puts the id at the top (it is `.message.id`), which cost a stray brief; `base:memory` does not say a title with ": " needs quotes; `delete-unnamed` refuses `git mv` in a loop, though a move can be undone | `internal/format/message.go:203`; `internal/guard/shell_words.go:70` (pointed) | new (docs) |
| 12 | Leftovers list the session's MCP servers; the session log files hold only a session's end | the ladder's leftovers; the log folder | known: BUG-029, BUG-111 |

## What it means for ZEN

- **Rehearse against the real checkout's state, not only a clone.** A read-only `bonsai check`-style pass over the
  real folder catches what a clone lacks. Issue 1 cost one stop here.
- **ZEN's docs name far more paths:**
  - 93,000 lines of markdown, CLAUDE.md routers in three subfolders, and ignored `.venv`, `node_modules` and
    `.next`. Issues 1 and 2 will be larger there.
  - Its CLAUDE.md files need a read for backticked paths before the link.
- **The session wrote everything by shell, so the guard saw almost nothing.**
  - Until Bonsai judges shell writes, the proof in a linked project is git's history and the climb on `main`, not
    the guard.
  - The kickoff for ZEN should tell agents to write documents with Edit and Write, so Bonsai's edit-time checks run.
  - The verifier should be re-run on the exact commit that lands (issue 9).
- **ZEN's own hooks** (wired in `Docs/.claude/settings.json`, bash, absolute paths) and its routines are retired
  under decision A. `.claude/**` is person-only once linked, so the hooks are switched off in the link script, by
  the director's run, not by an agent after.

## Open after this trial

1. **The push of ZenGarden-AI-API's `main` is the director's** (17 commits; it deploys). The project's own T-0005
   (the Dockerfile installs gliner2 unpinned in its first layer) is advised first.
2. **Filing issues 1 to 11 in `records/bugs.md`** waits for the other session's uncommitted rows there (T-6002 and
   T-6003, both `verify` at 10:08) to be committed, so the ids do not clash. Then `T-6001` moves to `done`.
3. **ZEN's link:** a task of its own, on the director's word, with the lessons above.
