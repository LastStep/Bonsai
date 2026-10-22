# 10-ask-answer: Claude Code asks the person a question with bonsai ask; the person answers in their own shell, below the live view.
# Crops: Claude alone = the left 64%; the right side (the view above, the person's shell below) = the right 36%; both =
# the whole frame. Length: about 60 s of screen. The bottom right pane (Alt+d) is the person's terminal: it holds no agent
# session, so bonsai takes its answer as the person's.
#
# Run by film.sh (sh docs/demos/clips/film.sh <out folder> 10-ask-answer), which makes the project (linked to Bonsai's base pack,
# the packs' plugin installed, a red climb recorded, Claude Code's allow rules in .claude/settings.local.json) and the
# environment (CLIPS, CLIP_BIN, CLIP_PROJECT) first, and runs this tape from the project's folder. Claude Code is
# filmed inside tmux (layout.sh): Claude Code on the left, the live view on the right (view.sh: bonsai status --line
# and the newest decisions of the log). Only the typed lines are scripted; Claude's replies are its own, so a take can
# differ, and film.sh can be run for this clip alone again (README.md).
# Claude Code's header (its first rows, top left) shows the plan name: the edit boxes it out.
Output "10-ask-answer.mp4"
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
Type "Before you touch anything, ask me which of two things to do first, with bonsai ask: type Decide, key which-first, title Which first, options Fix words.py and Add a lint rung. Then run bonsai ask --status agent:which-first as a command of its own, with nothing else in it, and run it again after a pause until it shows an answer. Tell me the answer."
Sleep 1s
Enter
# the person, in their own pane (Alt+d), answers once the ask is filed
Sleep 25s
Alt+d
Sleep 1s
Type "# Claude asked me something: here is my answer"
Enter
Type "until bonsai ask --status agent:which-first >/dev/null 2>&1; do sleep 1; done; bonsai ask --status agent:which-first"
Enter
Sleep 6s
Type `bonsai answer agent:which-first --choice "Fix words.py" --words "The bug first."`
Enter
Sleep 4s
Alt+a
Sleep 45s
