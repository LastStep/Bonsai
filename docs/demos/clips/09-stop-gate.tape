# 09-stop-gate: Claude Code fixes the failing tests and tries to stop; the stop gate holds it until the ladder is green and the work is committed.
# Crops: Claude alone = the left 64%; the view alone = the right 36%; both = the whole frame. Length: the gate holds it
# for a minute or three; the edit cuts the idle time. The session is started on the task (BONSAI_TASK), as a person does
# to name the work, which is what lets the gate hold it.
#
# Run by film.sh (sh docs/demos/clips/film.sh <out folder> 09-stop-gate), which makes the project (linked to Bonsai's base pack,
# the packs' plugin installed, a red climb recorded, Claude Code's allow rules in .claude/settings.local.json) and the
# environment (CLIPS, CLIP_BIN, CLIP_PROJECT) first, and runs this tape from the project's folder. Claude Code is
# filmed inside tmux (layout.sh): Claude Code on the left, the live view on the right (view.sh: bonsai status --line
# and the newest decisions of the log). Only the typed lines are scripted; Claude's replies are its own, so a take can
# differ, and film.sh can be run for this clip alone again (README.md).
# Claude Code's header (its first rows, top left) shows the plan name: the edit boxes it out.
Output "09-stop-gate.mp4"
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
Type "Two tests in test_words.py fail: count_words miscounts text with more than one space between words. Fix words.py, then tell me when you are done."
Sleep 1s
Enter
Sleep 200s
