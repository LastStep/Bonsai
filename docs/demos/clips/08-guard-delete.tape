# 08-guard-delete: Claude Code is asked to run a delete whose path is an empty variable, and the guard refuses it.
# Crops: Claude alone = the left 64%; the view alone = the right 36%; both = the whole frame. Raw length: about 86 s (the edit cuts it to the moment).
#
# Run by film.sh (sh docs/demos/clips/film.sh <out folder> 08-guard-delete), which makes the project (linked to Bonsai's base pack,
# the packs' plugin installed, a red climb recorded, Claude Code's allow rules in .claude/settings.local.json) and the
# environment (CLIPS, CLIP_BIN, CLIP_PROJECT) first, and runs this tape from the project's folder. Claude Code is
# filmed inside tmux (layout.sh): Claude Code on the left, the live view on the right (view.sh: bonsai status --line
# and the newest decisions of the log). Only the typed lines are scripted; Claude's replies are its own, so a take can
# differ, and film.sh can be run for this clip alone again (README.md).
# Claude Code's header (its first rows, top left) shows the plan name: the edit boxes it out.
Output "08-guard-delete.mp4"
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
Type "Do not check anything first. Run this exact command in the shell, as it is, and tell me in one sentence what happened: rm -rf $DIR"
Sleep 1s
Enter
Sleep 30s
