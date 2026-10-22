# The clip set

Eleven short screen recordings of Bonsai, one a moment, filmed by the person in their own terminal on the installed
`bonsai`. They are raw material: the README's pictures and the showcase video are edited from them (piece T-5723),
side by side where it helps. Each is a vhs tape (`NN-name.tape`); one script, `film.sh`, makes each clip's project and
films them all.

Nothing in a tape is scripted but the typed lines. Every command shown is run, and Claude Code's replies are its own,
so a Claude Code clip differs a little each take (and a take can miss its moment: film that clip again).

## Film them

In a plain terminal, never inside Claude Code (its own `!` lines count as an agent's session, and Bonsai refuses the
person's steps there; `film.sh` refuses to start inside one):

```bash
sh docs/demos/clips/film.sh ~/Videos/bonsai-clips
```

It checks `vhs`, `ttyd`, `ffmpeg`, `tmux`, `jq`, `claude`, `git`, `python3`, `/usr/local/bin/bonsai` and the network (the
base pack is fetched from GitHub), then, for each clip in order, makes the clip's project in a scratch folder
(`/tmp/bonsai-film.XXXXXX`, with its own `BONSAI_HOME`, never `~/.bonsai`, never this repository), runs the tape, and
prints where each mp4 is and which failed. It takes about 25 to 40 minutes (the Claude Code clips are most of it) and
needs no answer from the person. The mp4s are 1600 x 900.

What may be changed is at the top of `film.sh`: `BONSAI` (the bonsai to show; a test build only for practice, whose
path then shows in frames), `MODEL` (the tapes type `sonnet`), `PACK_SOURCE`, and `WORK` (the scratch folder: its
path is printed by Bonsai and Claude Code on screen, so it must hold no user name, which `/tmp` does not).

Before filming:

- Close other programs. The ladder's green and red output lists other processes it did not start ("leftovers: ..."),
  with their names, in clips 03, 04 and 05; the edit crops that block, or film with the machine quiet.
- Claude Code writes its own files outside the scratch folder (its session transcript under `~/.claude` and a
  folder-trust record), and the clips use your own Claude Code login. The packs' plugin is installed in the scratch
  folder's plugin cache only (`CLAUDE_CODE_PLUGIN_CACHE_DIR`), so your own plugins are neither used nor touched.
- Claude Code's header (the first rows of its pane, top left) shows the plan name. The edit boxes it out; no tape types
  it.

## Filming one clip again

```bash
sh docs/demos/clips/film.sh ~/Videos/bonsai-clips 06-guard-edit
```

Any number of names may follow the folder; each clip's project is made new, so a retake is safe. To run only a tape,
by hand, make its project yourself the way `lib.sh` does and run `vhs -o out.mp4 NN-name.tape` from the project's
folder with `CLIPS`, `CLIP_BIN`, `CLIP_PROJECT`, `CLIP_REF` and `BONSAI_HOME` set; film.sh is the easier way.

## The clips

| Clip | What it shows | Raw length | Crop |
|---|---|---|---|
| `01-link` | A new project linked to Bonsai: the preview of every file and settings line, the person's `y`, then `bonsai status` | about 20 s | whole frame |
| `02-check-status` | `bonsai check` (each finding in one sentence with its next step) and `bonsai status` on a project just linked | about 12 s | whole frame |
| `03-rung` | A rung added to the ladder: the climb is refused (`config-looser`, with the person's next step), the person runs `bonsai update --yes`, the climb is green | about 30 s | whole frame |
| `04-ladder` | The ladder red, with the failing tests' own names and the command to run, then the fix and the ladder green | about 26 s | whole frame |
| `05-log` | The log: `bonsai logs`, then every record of the day through `jq` (two climbs, an ask and its answer) | about 16 s | whole frame |
| `06-guard-edit` | Claude Code tries to edit `bonsai.yaml` with its Edit tool; the guard refuses (why, next step, rule `person-only`) | about 80 s, 10 s of it the moment | split screen |
| `07-guard-update` | Claude Code runs `bonsai update --allow-exec --yes`; the guard refuses (`person-command`) | about 80 s, 15 s of it the moment | split screen |
| `08-guard-delete` | Claude Code runs `rm -rf $DIR`; the guard refuses (`delete-unnamed`) | about 85 s, 10 s of it the moment | split screen |
| `09-stop-gate` | Claude Code fixes the tests and tries to stop; the stop gate refuses (`stop-red`) and says what to do; Claude commits and climbs; the gate lets it end (`stop-proven`) | about 200 s, the moment 30 s | split screen |
| `10-ask-answer` | Claude Code files a question with `bonsai ask`; the person answers it with `bonsai answer` in their own pane; Claude reads the answer. The status line shows `asks:1` while it is open | about 90 s | split screen, three panes |
| `11-status-line` | Bonsai's status line in Claude Code's own bar (task, ladder), beside the same line in the live view | about 45 s | bottom of the left pane, or whole frame |

"Raw length" is what the tape records, idle waits included (the tapes wait a fixed time for Claude); the edit cuts them
to the 10 to 40 s of screen each clip holds. Claude's reply varies from take to take: if a clip misses its moment (the
refusal or the gate is not in it), film that clip again.

## The split screen (clips 06 to 11)

Claude Code runs inside tmux on a socket of its own (`tmux -L clip`; your own tmux is not touched), in the left pane,
about 64% of the width. The right pane is the live view (`view.sh`): `bonsai status --line`, then the newest decisions
of the log (guard, stop gate, ladder, asks), redrawn every second from the project's log. Clip 10 splits the right
side in two: the lower pane is the person's own shell. The tapes move between panes with Alt+a, Alt+s and Alt+d.
`tmux.conf` has the whole setup; `layout.sh` builds it.

Crops, in pixels of the 1600 x 900 frame:

- Claude Code alone: x 0, y 0, width 1010, height 900 (the left pane; the border sits at about x 1012).
- The live view alone: x 1016, y 0, width 584, height 900. In clip 10 its upper half is the view and its lower half
  (from about y 440) is the person's shell.
- Both: the whole 1600 x 900 frame.
- The plan name ("Claude Max") sits in Claude Code's header, third row from the top of the left pane, about x 120 to
  400 and y 60 to 92 for as long as the header is on screen. Box it out in every Claude Code clip; check on a still frame.

## Files

| File | What it is |
|---|---|
| `film.sh` | films the clips; the one thing the person runs |
| `lib.sh` | the steps that make a clip's project (`mk_base`, `mk_prep`, `mk_update`, `mk_local`, `mk_red`, `mk_claude`) |
| `NN-name.tape` | one clip each; its top comment says what it shows and its crop |
| `prep.tape` | not a clip: starts Claude Code twice off camera so the folder-trust question is answered and the plugin installs |
| `layout.sh`, `tmux.conf`, `view.sh` | the split screen of clips 06 to 11 |

The project every clip but the first starts from is a word counter with a bug (`words.py`, `test_words.py`: two of
three tests fail), linked to Bonsai's base pack, with task `T-0001` running and the tests as the ladder's rung 1.
Clip 01 starts from an empty folder; clips 02 and 03 differ in the way their tape's top comment says.
