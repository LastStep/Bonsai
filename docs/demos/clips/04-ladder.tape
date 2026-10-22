# 04-ladder: the ladder red, with the failing test's own command and name, then the fix and the ladder green.
# Crop: the whole frame. Raw length: about 27 s.
#
# Run by film.sh (sh docs/demos/clips/film.sh <out folder> 04-ladder), which makes the project, the scratch Bonsai home and
# the environment first (CLIPS, CLIP_BIN, CLIP_PROJECT, CLIP_REF) and runs this tape from the project's folder. To run
# it alone, see README.md ("Filming one clip again"). Every command shown is run; nothing but the typed lines is scripted.
Output "04-ladder.mp4"
Require jq

Set Shell "bash"
Set FontFamily "DejaVu Sans Mono"
Set FontSize 22
Set Width 1100
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

Type "# the ladder runs the tests; two of them fail"
Enter
Type "bonsai ladder --task T-0001"
Enter
Sleep 8s
Type "clear"
Enter
Type "# count_words splits on one space; split() splits on any run of spaces"
Enter
Type `sed -i 's/split(" ")/split()/' words.py`
Enter
Type "git commit -qam 'count words across extra spaces'"
Enter
Sleep 1s
Type "bonsai ladder --task T-0001"
Enter
Sleep 8s
