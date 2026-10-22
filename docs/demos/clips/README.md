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

CLIPS_TABLE

## The split screen (clips 06 to 11)

Claude Code runs inside tmux on a socket of its own (`tmux -L clip`; your own tmux is not touched), in the left pane,
about 64% of the width. The right pane is the live view (`view.sh`): `bonsai status --line`, then the newest decisions
of the log (guard, stop gate, ladder, asks), redrawn every second from the project's log. Clip 10 splits the right
side in two: the lower pane is the person's own shell. The tapes move between panes with Alt+a, Alt+s and Alt+d.
`tmux.conf` has the whole setup; `layout.sh` builds it.

Crops, in pixels of the 1600 x 900 frame:

CROPS

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
