---
id: OPT-0002   # an options write-up; every field and how to fill it: skill workflow:options
title: What should the README's recordings show, and how should they be made?
lane: director
status: decided
author: orchestrator
created: 2026-10-22
decided: 2026-10-22
chosen: B
---

# What should the README's recordings show, and how should they be made?

## Why this is being asked now

The director, 22 Oct: "i want to see what all we can show in the recordings, and how we can do that process, like
showing claude session, terminal etc". Two first attempts are in (T-5711, plan `records/plans/P-T-5711.md`): the
scripted GIF of Bonsai's commands (`docs/demos/first-project.gif` on branch T-5720, 41 s) and three takes of a real
Claude Code session (`~/bonsai-checks/T-5721/takes/3/session.mp4`, 266 s, the best). The final recordings (after
rc.2's fixes, before the rc.2 tag) wait on this choice. Nothing else in T-5711 waits on it: the README, the picture
and the feature document go on.

## What we learnt from the first attempts

- **The scripted terminal recording works** (vhs, a script kept in the repo, rebuilt in about 2.5 minutes). It shows
  exactly what Bonsai prints.
- **A real Claude Code session can be filmed** the same way: the script types the prompts into a real Claude Code.
  Take 3 showed four of the six beats for real: Claude edited a file, tried `bonsai.yaml` and was refused (why, next
  step, rule `person-only`), tried to end and was held by the stop gate (`stop-red`, then `stop-dirty`). It stalled
  on a Claude Code permission question before the green climb. Claude does not take the same path twice: about 25
  minutes a take, and some takes are lost.
- **What shows on screen that should not:** a test build (the only kind an agent may drive through your `bonsai
  update`) prints its own path, which holds your user name (`/home/<you>/...`), and lines a real install never prints.
  The **installed** `bonsai` prints `/usr/local/bin/bonsai`: no name. Claude Code's header shows the plan name
  ("Claude Max") and the folder.
- **Claude Code prints a hook's whole command line** in front of a refusal ("PreToolUse:Edit hook error: [b=...; [ -x
  ...]"). That is Claude Code's own wording, not Bonsai's; it is noisy but true. It would be the same for any user.
- **Size:** README width is about 830 px. A long session at full terminal width makes small text; a 2-minute GIF is
  2-5 MB.

## What could be shown (the menu)

Each is a real moment Bonsai produces today (the feature audits, `docs/features.md` on branch T-5722).

| Moment | Where it happens | Best shown as |
|---|---|---|
| Linking a project: the preview, then `--yes` | terminal | scripted GIF (have it) |
| `check` and `status`: findings each with its next step | terminal | scripted GIF (have it) |
| The guard refusing an edit of a protected file, with why, next step and rule | Claude Code | real session (take 3 has it) |
| The guard refusing a person's command from an agent (`bonsai update --allow-exec`) | Claude Code | real session |
| The guard refusing a delete that names nothing (`rm -rf $DIR`) | Claude Code | real session |
| The stop gate holding Claude until the proof is green | Claude Code | real session (take 3 has it, red only) |
| The ladder: red with the failing test's own command, then green | terminal or Claude Code | either |
| A rung added: refused until the person accepts it in their own terminal | terminal | scripted GIF (have it) |
| An agent asks the person (`bonsai ask`); the person answers (`bonsai answer`) | both | side by side |
| The log: every decision in order (`bonsai logs` through `jq`) | terminal | either |
| The status line in Claude Code's bar (task, ladder, asks) | Claude Code | real session |
| A team: an orchestrator briefs a builder, the builder reports (workflow pack) | Claude Code | too long for the README; the YouTube video |

## What matters here

1. **True to 1.0:** what is on screen is what an installed Bonsai prints, with nothing a user never sees.
2. **Clear at README size,** in under about a minute each, with no private detail in any frame.
3. **Your time:** how much of it each way needs.

## Options

### A. Terminal only

Two or three short scripted GIFs of Bonsai's commands (the one we have, plus the ladder red to green and the log),
and still images of the guard's refusal and the stop gate cut from a real session.

- **Gets us:** small, sharp, rebuilt in minutes whenever the words change; no Claude Code screen to keep current.
- **Costs:** the README never shows Claude being stopped, which is the heart of Bonsai; stills only.
- **Reversible?** Fully; the session can be added later.

### B. The terminal GIF plus one real session, filmed in your terminal on the installed Bonsai

Agents prepare everything: the project, the scripts, the prompts, the frame checks, the cut. After rc.2's fixes are
installed, you run one line per recording in your own terminal (`vhs docs/demos/session.tape`; about 5 minutes
each, a retake if Claude wanders). The session shows the six beats; idle time is sped up and the caption says so.

- **Gets us:** exactly what a user sees with an installed Bonsai; no test-build lines; no user name (the installed
  path); your `bonsai update` really is yours. Claude Code's plan name boxed out.
- **Costs:** about 15-30 minutes of your time at the final pass; a session GIF of 2-5 MB.
- **Reversible?** Fully; the scripts stay in the repo for the YouTube video.

### C. B, side by side

As B, but the frame is split: Claude Code on the left, a live view on the right (`bonsai status --line` and the log's
newest decisions), so the viewer sees the guard's decision land as Claude is refused. Made with tmux inside the same
recording.

- **Gets us:** the "how it works" picture, live; the ask-and-answer moment becomes showable.
- **Costs:** about 2-3 more hours of agent work, untested; at README width each half is narrow and its text small, so
  it suits the YouTube video better than the README.
- **Reversible?** Fully.

### D. Agents film everything on a test build

As B, but agents run it all, with the test build's path and its extra lines boxed out frame by frame.

- **Gets us:** none of your time.
- **Costs:** a build no user runs; boxes over scrolling text are fragile; a frame missed would publish your user name
  in a public repo.
- **Reversible?** Yes, but a published frame is cached once pushed.

## Comparison

| | A | B | C | D |
|---|---|---|---|---|
| True to 1.0 | yes | yes | yes | no (test build) |
| Shows Claude being stopped | stills only | yes | yes, with the live log | yes |
| Clear at README size | yes | yes | small text | yes |
| No private detail | yes | yes (installed path) | yes | only if every box holds |
| Your time | none | 15-30 min | 15-30 min | none |
| Agent time from now | 1-2 h | 2-3 h | 4-6 h | 2-3 h |

## Recommendation

**B.** It is the only way that shows Claude being stopped with exactly what an installed Bonsai prints, and it keeps
your name out of every frame by construction, not by boxes. What would change it: if your time at the final pass is
the scarce thing, A; if the recording is mainly for the YouTube video, C's split screen is worth its cost there
rather than in the README.

## What happens next, either way

- **A:** T-5721 becomes stills; a builder adds two scripted GIFs to T-5720.
- **B:** T-5721's builder fixes the permission stall (allow rules, already drafted) and makes the session's tape run
  on the installed `bonsai` (no test build); one practice take by a runner on rc.1 to prove the script; then your
  final pass after rc.2's fixes are installed.
- **C:** as B, then a builder adds the tmux split and one practice take.
- **D:** T-5721's builder goes on as now with boxes.

## Sources

- T-5721's report `M-2026-10-22-fff85979` (takes 1-3, the path in frame, the fixes since).
- T-5720's report `M-2026-10-22-23b07ccc` (the scripted GIF; the test build's line).
- Take 3's contact sheet `~/bonsai-checks/T-5721/takes/3/st/sheet.png` (the refusal and the stop gate on screen).
- The README research `~/bonsai-checks/T-5711/readme-research.md` (sizes, vhs, GitHub's 10 MB limit).

## The director's choice (22 Oct 2026)

**"lets do option 1. but i also want to make sure that we record clips, and then we can edit and make a good video
using them, which can include showing things side by side wherever it is useful, and any other things for a good
showcase"** ("option 1" is B, the recommended one, first in the list put to them). So: B, filmed as separate clips,
one a moment, and an edited showcase video made from them, side by side where it helps. `records/plans/P-T-5711.md`,
amendment 2.
