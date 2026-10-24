# The director's trial on rc.2 (LinkSink): did it succeed for Bonsai?

For the director (T-5733). Nothing here asks anything of you. All times UTC. The evidence: two auditors (opus),
reports `M-2026-10-24-d85969b6` (Bonsai's machinery) and `M-2026-10-24-c3d8453a` (the packs' steering and your side),
`bonsai message --show <id>`; "L<n>" is line n of LinkSink's `.bonsai/local/log/s-19462946-...ndjson`.

## In short

**Bonsai's machinery: a success. The packs' steering: partial.** On a fresh project of yours (`~/Apps/LinkSink`,
linked 24 Oct 06:36, session 06:39-09:05, about 2 h 26 min), Bonsai linked cleanly, guarded every call, kept a whole
log and carried every brief and report, with no crash and nothing that blocks 1.0. Where it fell short is the
workflow pack: it did not make the orchestrator keep a run report, STATE or task moves while it planned, research
and decided, and nothing in Bonsai noticed. The orchestrator caught up only at the end (09:03-09:05: STATE, the task
moves, one run report, the first commit `f8f03dc`). Five fixes go to rc.3.

## The criteria

| | What | Verdict | Evidence |
|---|---|---|---|
| A | The link (README's first project path) | pass | packs the same in bonsai.yaml, the lock, the settings and Claude Code's plugin records; lock sha256s match; 156 deny rules; 7 workflow roles loaded; `bonsai status`: 0 changed, 0 missing |
| B | The guard | pass: no wrong allow; 2 of 3 refusals over-strict | 632 decisions; nothing written to `.claude/**`, bonsai.yaml, the lock, `.bonsai/local` or the tables; builders wrote only in scratch; the `rm $var` refusal right (L1382); a researcher's and the verifier's scratch writes refused wrongly after `cd $VAR` (L344, L810; BUG-108) |
| C | The log | pass | 2,191 lines all parse with all 33 fields; 605 calls each with start, guard and end; every typed subagent paired; no home path, email, token or key found; the sessions table lagged only because `--write` had not run |
| D | Messages | pass | 7 threads, 16 messages, every brief answered; all 16 shown by `bonsai message --show` |
| E | Documents | pass | `bonsai check`: no findings, only the tables' lag |
| F | The packs steered the agents | partial | roles and briefs right (7 agents, each by role with its brief; helpers wrote nothing); director lane honoured for the stack (options, your choice, ADR); but no run report for 2 h 25 min, STATE empty until 09:04, T-0001 at `todo` while worked on, so 5 of 7 agent runs' hours read "none" |
| G | Your side | smooth | no permission prompt (auto mode); three rounds of questions (07:23, 07:50, 09:02); by hand: editing bonsai.yaml to add workflow, `bonsai update --allow-exec --yes`, `/plugin` twice (the marketplace rename, BUG-107) |

Hooks were quick: session start 37 ms, prompt 5-11 ms, the stop gate 10-53 ms, a subagent's start 21-82 ms.

## What goes where

**rc.3 (the trial's fixes).** BUG-103: a run report for any task's work, planning and decisions included, not only
code. BUG-104: orchestrate names STATE and a home for the orchestrator's notes. BUG-105: `running` when any agent
starts on a task, research and decisions included. BUG-106: `bonsai check` warns on a repo with no commit or an
uncommitted bonsai.yaml. The words of BUG-107 (say a reload is needed after a pack is added) and BUG-108 (a refusal's
next step that fits a write to scratch). BUG-103 to BUG-105 are the workflow pack's text; BUG-106 to BUG-108 Bonsai's.

**After 1.0.** BUG-107's cleanup of the old marketplace; BUG-108's reading of `cd $VAR`; BUG-109 (a message pointer
into Claude Code's scratchpad refused); BUG-110 (no model in subagent rows); BUG-111 (log files holding only a
session's end); the known BUG-034 and BUG-039 seen again.

**A question for later, not a bug.** You approved the one-pager and the plan through Claude Code's question box, and
the orchestrator applied it by editing the files: Bonsai keeps no record of your answer there, so `check` cannot tell
your move from an agent's. `bonsai ask` exists for that; whether the pack should require it is a design choice.

**The agent's, nothing in code.** The one-pager and the plan approved together, not in order; the orchestrator wrote
the options itself, not a researcher; the spike's mockup code committed into `design/reference/` (your approved plan
asks for it); reporting, verification-ladder and reward-hacking-guards never opened.

## Not covered

The session after 09:05 (`a60e4bad`, T-0003 running), which starts the build: the ladder, a builder in a worktree and
the stop gate have not met real work yet in LinkSink. Not checked: LinkSink's files byte for byte against a fresh
init; the exact commands you typed by hand (read from file times and the agent's words).
