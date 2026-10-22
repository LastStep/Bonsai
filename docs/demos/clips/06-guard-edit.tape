# 06-guard-edit: Claude Code is asked to edit bonsai.yaml with its Edit tool, and the guard refuses (why, next step, rule).
# Crops: Claude alone = the left 64% of the frame; the view alone = the right 36%; both = the whole frame. Length: about 40 s of screen.
#
# Run by film.sh (sh docs/demos/clips/film.sh <out folder> 06-guard-edit), which makes the project (linked to Bonsai's base pack,
# the packs' plugin installed, a red climb recorded, Claude Code's allow rules in .claude/settings.local.json) and the
# environment (CLIPS, CLIP_BIN, CLIP_PROJECT) first, and runs this tape from the project's folder. Claude Code is
# filmed inside tmux (layout.sh): Claude Code on the left, the live view on the right (view.sh: bonsai status --line
# and the newest decisions of the log). Only the typed lines are scripted; Claude's replies are its own, so a take can
# differ, and film.sh can be run for this clip alone again (README.md).
# Claude Code's header (its first rows, top left) shows the plan name: the edit boxes it out.
Output "06-guard-edit.mp4"
Require tmux
Require jq

Set Shell "bash"
Set FontFamily "DejaVu Sans Mono"
Set FontSize 16
Set Width 1600
Set Height 900
Set Padding 20
Set Margin 0
Set Theme "Catppuccin Mocha"
Set TypingSpeed 12ms
Set PlaybackSpeed 1

Hide
Type "export PS1='$ '; . $CLIPS/layout.sh"
Enter
Sleep 3s
Type "clear"
Enter
Show

Type "claude --model sonnet"
Enter
Sleep 12s
Type "Read bonsai.yaml, then use the Edit tool on it, as it is, to change ladder_floor: [1] to ladder_floor: [] . Then tell me in one sentence what happened."
Sleep 1s
Enter
Sleep 75s
