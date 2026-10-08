# Run: the fresh review of the rebuild's plan

- Date: 2026-10-08
- Who: one reviewer (Opus, fresh context), briefed by the orchestrator; the orchestrator fixes what it finds
- Branch: `main` at `713ae04` (no code; the review reads, the fixes are edits to `design/`)
- Outcome: the plan reviewed, fixed on every finding and approved by Rohan (8 Oct)

> Opened before the first edit and appended to as the work went.

## Log

- 12:51 Read `STATE.md`: `main` at `713ae04`, pushed; CI on it: `test`, `lint`, CodeQL green, `govulncheck` red at its
  install step (known, see `STATE.md` loose ends). Next step is this review (`STATE.md`, "The one thing to do next",
  step 1).
- 12:52 Briefed one fresh Opus reviewer: `design/plan.md` whole, `design/one-pager.md`, `STATE.md`, `CLAUDE.md`, and the
  spec and contract by section where the plan cites them. It reads and reports; it edits nothing.
- 12:55 Rohan: use Haiku for small bookkeeping and audit jobs. `CLAUDE.md`'s models rule now names Haiku for them;
  committed on `main` as `38d01e6`, not pushed yet (it goes out with the plan's fixes).
- 13:03 The reviewer's verdict: READY AFTER FIXES. Five must-fix (part 0's `expect.json` undefined for files with no
  `format:`, for markdown cases and for cases §2.4 does not decide; the schemas' sources miss spec §16 rows 18/23/24/27
  and the "as today" fields; the YAML examples cannot be tested without a reader; line 240 calls govulncheck green; parts
  1, 3 and 5 land before their proof), nine should-fix (among them `BONSAI_HOME` unset so tests would write the real
  `~/.bonsai`, line 144's wrong claim that memory has no format, the Windows check after the push), seven minor, two
  one-pager disagreements, and one question for Rohan (when the hand checks happen). Read-only; about 11 minutes.
- 13:03 Spot-checked three claims against the files: line 240 says govulncheck is "expected green" (it is red); line 144
  says memory "has no format of its own" (spec §10 and §16 row 19 give `bonsai.memory/1`); contract §2.4 does carry the
  "YAML 1.1 or 1.2 library reads the same value" rule. All three hold.
- 13:04 Worktree `~/Servers/Bonsai-plan-fixes` on branch `plan-fixes` from `main` at `38d01e6`. Briefed one Opus builder
  to check each finding against its source and fix the plan and one-pager; the hand-check timing waits on Rohan.
- 13:07 Rohan chose two hand-check sittings: checks b, c and d (about 25 minutes) after part 5, check a (about 20
  minutes) after part 4. Sent to the builder.
- 13:13 The builder's report: eight commits on `plan-fixes` (`e087d9b`..`8c6eb6f`), 168 lines in, 89 out, in
  `design/plan.md` and `design/one-pager.md`. Every finding checked against its source and fixed; none rejected. One
  correction to the review: `|+` is settled by §2.4's "only" (refused), so it became a trick case, not an undecided one.
  Two sittings applied everywhere. What the review missed, now fixed: the plan's models line lacked Haiku; `bonsai.error`
  (spec §16 row 29) was missing from part 0's "Not here"; the log's names for the binary's path and hash are not fixed
  by the format review, so the plan says not to invent them. For part 0's brief: the new private-pattern test flags
  `/home/`, and contract §12's `status --json` example uses `/home/<user>/...`, so part 0's example needs a made-up
  absolute path the pattern does not match. `STATE.md` needs a rewrite (step 1 done, step 4's Windows clone before the
  push, the two sittings, the §19 D note).
- 13:15 The orchestrator read "For Rohan" and the one-pager's diff, and checked the new sums: 139-218 + 30-47 + 6-10 =
  175-275 h; times 1.6, 280-440; part 0 10-16; the skeleton 48-75. Briefed Haiku to tick every finding and minor item
  against the diff and list any private-looking string.
- 13:16 Haiku's audit: all 35 items FOUND, each with a line and a quote. Two lines still say "one 45-minute sitting":
  the orchestrator read both, and both are right (the losing choice in "Choices made", and the note that spec §17 step 6
  still says one sitting). The private-pattern hits are the rules' own wording; no real path, name or address.
  `plan-fixes` clean.
- 13:18 Checked the builder's §19 D claim against the spec: §18 (line 1398) and §19 D both give 60 h and 30-46 h, then
  61 h and 30-47 h. Fast-forwarded `main` to `plan-fixes` (`8c6eb6f`). Rewrote `STATE.md`: the review done, Rohan's
  approval next, part 0's steps with the Windows check before the push, the two sittings, Haiku, the §19 D note.
  Proof for this run: a text change, read by a fresh reviewer and a Haiku audit and the orchestrator; no code, no tests.
- 13:20 Pushed `9931d98`. CI on it: `test`, `lint`, CodeQL green; `govulncheck` red at its install step, as before.
- 13:39 Rohan approved the plan and the one-pager ("option 1"), and said the Homebrew tap token has been removed from
  GitHub. Both files marked approved; the token's line in the plan marked done; `STATE.md` rewritten for part 0.

## Runs

| Run | Model | Start | End | Minutes |
|---|---|---|---|---|
| Reviewer: the plan | Opus | 12:52 | 13:03 | 11 |
| Builder: the plan's fixes | Opus | 13:02 | 13:13 | 11 |
| Audit: findings against the diff | Haiku | 13:14 | 13:16 | 2 |

Times are local (8 Oct).
