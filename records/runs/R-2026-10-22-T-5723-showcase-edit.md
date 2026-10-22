---
format: bonsai.run/1   # fields: bonsai check --schema bonsai.run; how to fill: skill base:run
id: R-2026-10-22-T-5723
task: T-5723
role: builder
model: fable
started: 2026-10-22 09:57
finished: 2026-10-22 10:31
outcome: merged
commits: ["c0f3ea04", "8ac57e1e"]
labels: {}
---

# Run: README and demos, piece F: the showcase edit (T-5723)

> Opened at the start of the session and added to as it goes. Kept in the main checkout, uncommitted, for the
> orchestrator's records commit: the brief says the branch commits nothing under `records/`.

Brief `M-2026-10-22-94ca258b` (the orchestrator). Builder: Fable (the visual design). Worktree
`~/Servers/Bonsai-T-5723`, branch `T-5723` from T-5721's tip `ee88a00c`. Read: the common brief of 22 Oct, CLAUDE.md,
the task, the plan (amendments 1 and 2), OPT-0002, the clip set's README and its eleven tapes, the README on T-5718,
the dark How it works picture on T-5719.

## What is different now

A script, `sh docs/demos/edit.sh <clips folder> <out folder>`, turns the eleven clips film.sh records into the
README's two pictures and a showcase video, with ffmpeg alone: `first-project.gif` (link, check, a rung refused, the
person's update, green), `session.gif` (the guard refusing, the stop gate holding, then green) and `showcase.mp4`
(2 min 58 s at 1920 x 1080: a title card, five chapters in the order a person meets Bonsai, each moment with a
one-line caption, the Claude Code moments beside the How it works picture with the step lit and the live log under
it, idle time sped up and labelled, an end card with the repository's address). The story, the captions and each
clip's in and out points are in `docs/demos/showcase.txt`, so a new take is re-cut by changing times in a text file,
not the script. The plan name in Claude Code's header is covered in every frame it is on screen, by a box the colour
of the pane that follows the header when it scrolls. The practice cut from the practice clips is at
`~/bonsai-checks/T-5723/practice-cut/` (scratch only: its frames hold a user name in the test build's path).

## Deviations from the plan

None from the brief. Three choices worth knowing:
- The first-project GIF is 830 px wide with 32 colours and no dither: the only setting under 1 MB (36 s came to
  0.77 MB; 1000 px wide was 1.2 MB, 1100 px 1.25 MB even at 8 frames a second). At that width the clips' 160-column
  terminal (font 20 px at 1600 wide) is small: a line of output is readable only with effort, and the long wrapped
  lines are not. See "What did not work".
- The session GIF is 1000 px wide with 64 colours (3.08 MB, 42 s): Claude Code's pane (the left 1010 px of the
  frame) at nearly 1:1, which reads well.
- The run report is in the main checkout, uncommitted, not on the branch (the brief: nothing under `records/`).

## Ladder

`/usr/local/bin/bonsai ladder --task T-5723 --json` from the worktree, twice: at `c0f3ea04` (green, sha256
`f7567efc07a863299afed33356bd34a72abb709885666df0930364e254383c56`) and at the HEAD, `8ac57e1e` (green, sha256
`4016fc0a3e36f7aa63a7709e370f1e03c5ae7416c9c9df7eb71b5ee6c53997d9`, `mode: local`, `git.dirty: false`, started
10:28:15 UTC, finished 10:30:35 UTC; `.bonsai/local/ladder/T-5723.json` in the main checkout). The HEAD's climb:

| Rung | Check | Result | Evidence |
|---|---|---|---|
| 0 | guard | green | 4598 ms; every change inside the task's grants (two files under `docs/demos/`) |
| 1 | vet | green | 798 ms |
| 2 | schemas | green | 444 ms |
| 3 | tests | green | 62.6 s; 4883 seen, 9 skipped with reasons (Windows, dist); ratchet `go_tests` 4756 -> 4874 |
| 4 | tests with the fault switch | green | 57.0 s; 2176 seen, 3 skipped; ratchet `go_tests_tagged` 2120 -> 2173 |
| 5 | git integrity | green | 39 ms; stash 0, assume-unchanged 0, skip-worktree 0 |

No code changed; the counts are the floors' plus what landed since the floors were set.

## Log (UTC)

- 09:57 start. Practice clips: eleven mp4s, 1600 x 900, 25 fps, h264, lengths 19, 12, 31, 27, 19, 78, 82, 86,
  205, 88, 46 s. ffmpeg 4.4.2 with librsvg, freetype and fontconfig. Fonts: DejaVu Sans and Sans Mono installed.
- 10:01-10:10 measured (ran): the frame-to-frame scene score (ffmpeg's `select` metric): the cursor blink 0.00038, a
  typed character 0.00005-0.00015, Claude Code's spinner 0.00006-0.00010, a new line of output 0.0007-0.05. So
  "something appeared" is separable at 0.0006, and typing is not separable from the spinner: the cut keeps 3 s
  after each change and 5.5 s before the next (the longest typed line, clip 05's jq, is 5.3 s), and fast-forwards
  only the middle into 1 s with a tag; a clip's tail after its last change is cut to 3.5 s. The pane's background
  under the plan name reads 0x1d1c2d in the encoded frames. Claude Code's mascot (x 25-110, y 40-90) reads V 153-169
  in a 70 px column against 127 when absent: the header is tracked by it in ten 10 px bands. librsvg renders the
  picture with its fonts (checked by eye). Event times of 09 and 10 read off tiled stills.
- 10:12 `docs/demos/edit.sh` and `docs/demos/showcase.txt` written. Analysis of all eleven clips: 7 s (one decode per
  clip, both measures in one pass).
- 10:14-10:20 four defects found by looking at the stills and fixed: awk refused `in` as a variable name; the scene
  counter was clobbered by a `read` in another function (three scenes wrote one file); every caption read one text
  file (a counter inside a command substitution: all captions said "green"); the condensed intermediate was cached
  by name, so a fix to the box did not reach the output (now rebuilt when the box, the plan or the script is newer).
- 10:21 the box over the plan name: first at y 52-92, which clipped the "Claude Code" title row (seen in a 2x crop);
  now y 60-85 at rest, the title and the path rows intact, the plan name blanked. 09's header leaves in one jump at
  34.0 s (bands all 127 at 34.1 s); the box ends at 33.98 s.
- 10:22 the showcase was 229 s: the status-line scene cut from 25 to 9 s, speeds raised; now 178 s.
- 10:24 first climb green at `c0f3ea04`. The stills' time labels were off by up to 1.5 s (an `fps` filter's
  output time, not the frame's): the stamp is now the frame's own second, which matters for retiming captions.
  The sped-up tag moved to the bottom right on whole-frame clips (it covered the live view's status line).
- 10:28 the session GIF slowed from 1.3x to 1.15x (37 s to 42 s, the brief's 40-60); second climb green at `8ac57e1e`.
- 10:31 processes checked, report written.

## What did not work

- **The first-project GIF is not comfortably readable at 830 px.** The clips are filmed at 1600 x 900 with a 20 px
  font: 160 columns, scaled to 830 px, give 5 px per character. Short lines can be read; the long wrapped lines of the
  init preview and the refusal cannot. No cut fixes that: it is the tape's width. Proposed, not done (T-5721's
  files): film clips 01-05 at `Set Width 1100`, `Set FontSize 22` (about 95 columns), which reads at 830 px; the
  edit script needs no change (it scales whatever width it gets; the plan-name box and the pane crops apply only to
  the Claude Code clips, which keep their size).
- The GIF's size budget forced 32 colours at 830 px (0.77 MB for 36 s); a 1000 px version is 1.2 MB. The session
  GIF has room (3.08 MB of 5).
- The caption and in/out times in `showcase.txt` are the practice take's. A new take of a Claude Code clip (Claude's
  own replies) needs them read off the stills again; the idle cuts, the box and the tail are automatic. The 09 clip
  needs an explicit out point (52 s here) because the practice take ends with the tmux server gone at 59 s.
- In the practice frames the hook lines print the test build's path, with a user name: expected (OPT-0002), the
  reason the practice cut stays in scratch. The director's clips on the installed `bonsai` print
  `/usr/local/bin/bonsai`.
- Not run: the light picture (the script takes the dark one; the clips are dark); a Windows run (the script is
  POSIX sh for Linux's ffmpeg; the director films on WSL).

## Evidence for a human

The practice cut (scratch; frames hold the user name, never committed):
- `~/bonsai-checks/T-5723/practice-cut/first-project.gif`  0.77 MB (803,412 bytes), 36 s, 830 x 510
- `~/bonsai-checks/T-5723/practice-cut/session.gif`  3.08 MB (3,228,530 bytes), 42 s, 1000 x 944
- `~/bonsai-checks/T-5723/practice-cut/showcase.mp4`  4.79 MB (5,022,237 bytes), 178 s, 1920 x 1080, 25 fps
- The stills I looked at (one frame every 3 s of each output, every 2 s of a scene): `.../practice-cut/work/stills/`
  (`first-project.png`, `session.png`, `showcase.png`, `sc-06.png`, `sc-10.png`, `sc-12.png`, `sc-13.png`,
  `hdr-box.png` the header at 2x with the box, `raw09-scroll.png` the header leaving, `tag13.png` the sped-up tag,
  `card-title.png`, `card-chapter.png`, `card-end.png`, `sc-06-frame.png` and `sc-10-frame.png` full frames of
  the split layout). Every scene is also kept as `work/showcase-NN.mp4`.
- To rebuild: `PICTURE=<the dark SVG> sh docs/demos/edit.sh ~/bonsai-checks/T-5721/practice <out>` (about 70 s;
  once T-5719 lands, the picture is found beside the script).

## Files changed

- `docs/demos/edit.sh` (new, 22 KB): the cut. Analysis (change times, the header's position), the plan of each
  clip's segments, the condensed intermediate, the lit picture, the cards, the scenes in two profiles (1920 x 1080
  and a GIF width), the GIFs' palette, the stills. Two modes besides the cut: `--stills <video> <png> [s]` and
  `--events <clip>`.
- `docs/demos/showcase.txt` (new, 6 KB): the story: cards, scenes with in/out/speed, captions with the lit step, and
  each GIF's width, colours and frame rate.

## Processes left running

None. Checked with `ps -eo pid,etime,cmd | grep -E "ffmpeg|ffprobe|edit.sh|vhs|ttyd|tmux"` at 10:30 UTC: only the
machine's own `ttyd` on port 7681 (up 4 h 37 min, not mine, left alone as the plan says).

## What it cost

34 minutes wall clock (09:57 to 10:31 UTC), one builder (Fable); two climbs of about 2.5 minutes each. Spend not
recorded.

## Left for next time

- The director looks at the practice cut (the three files above) and says where the video lives (the repo, a
  release, YouTube: the plan's open question) and what to change in the story.
- The final cut from the director's clips after rc.2's fixes: retime `showcase.txt` from the stills (a Sonnet pass,
  about 15 minutes), run the script, look at the stills, check every frame holds no user name (none by construction on
  the installed `bonsai`).
- Proposed task (T-5721's files): clips 01-05 at a narrower width for a readable README GIF (above).
- The README (T-5718) names `docs/demos/session.md` as how the session recording was made; with clips, that page
  would describe film.sh and edit.sh. Not mine; noted for the orchestrator.

## Verifier

> Added by the verifier, in a fresh context. Pass or fail, with reasons.
