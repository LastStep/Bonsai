# 02-check-status: bonsai check and bonsai status on a linked project: every finding in one sentence, with its next step.
# Crop: the whole frame. Length: about 20 s.
#
# Run by film.sh (sh docs/demos/clips/film.sh <out folder> 02-check-status), which makes the project, the scratch Bonsai home and
# the environment first (CLIPS, CLIP_BIN, CLIP_PROJECT, CLIP_REF) and runs this tape from the project's folder. To run
# it alone, see README.md ("Filming one clip again"). Every command shown is run; nothing but the typed lines is scripted.
Output "02-check-status.mp4"
Require jq

Set Shell "bash"
Set FontFamily "DejaVu Sans Mono"
Set FontSize 20
Set Width 1600
Set Height 900
Set Padding 30
Set Margin 0
Set Theme "Catppuccin Mocha"
Set TypingSpeed 35ms
Set PlaybackSpeed 1

Hide
Type "export PS1='$ '; cd $CLIP_PROJECT; clear"
Enter
Show

Type "# is this project in order?"
Enter
Type "bonsai check"
Enter
Sleep 5s
Type "clear"
Enter
Type "bonsai status"
Enter
Sleep 6s
