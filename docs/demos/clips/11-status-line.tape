# 11-status-line: Claude Code's own bar shows Bonsai's status line (the task, its ladder, asks), live, beside the same line in the view.
# Crops: the bottom bar of Claude's pane = the left 64%, the last three rows; the whole frame for the view too. Length:
# about 30 s of screen. The status line is the person's own setting (bonsai writes none), set in this project's
# .claude/settings.local.json to: bonsai status --line
#
# Run by film.sh (sh docs/demos/clips/film.sh <out folder> 11-status-line), which makes the project (linked to Bonsai's base pack,
# the packs' plugin installed, a red climb recorded, Claude Code's allow rules in .claude/settings.local.json) and the
# environment (CLIPS, CLIP_BIN, CLIP_PROJECT) first, and runs this tape from the project's folder. Claude Code is
# filmed inside tmux (layout.sh): Claude Code on the left, the live view on the right (view.sh: bonsai status --line
# and the newest decisions of the log). Only the typed lines are scripted; Claude's replies are its own, so a take can
# differ, and film.sh can be run for this clip alone again (README.md).
# Claude Code's header (its first rows, top left) shows the plan name: the edit boxes it out.
Output "11-status-line.mp4"
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

Type "BONSAI_TASK=T-0001 claude --model sonnet"
Enter
Sleep 12s
Type "Tell me in one sentence which task is running, from bonsai status --active."
Sleep 1s
Enter
Sleep 40s
