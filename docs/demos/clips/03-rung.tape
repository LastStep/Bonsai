# 03-rung: a rung is added to the ladder: refused until the person accepts it in their own terminal, then green.
# Crop: the whole frame. Raw length: about 31 s.
#
# Run by film.sh (sh docs/demos/clips/film.sh <out folder> 03-rung), which makes the project, the scratch Bonsai home and
# the environment first (CLIPS, CLIP_BIN, CLIP_PROJECT, CLIP_REF) and runs this tape from the project's folder. To run
# it alone, see README.md ("Filming one clip again"). Every command shown is run; nothing but the typed lines is scripted.
Output "03-rung.mp4"
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
Type `python3 -c "p='bonsai.yaml';s=open(p).read();s=s.replace('ladder_floor: []','ladder_floor: [1]').replace('\nladder: []\n','\nladder:\n  - rung: 1\n    name: whitespace\n    kind: command\n    command: git diff --check HEAD\n    required: true\n');open(p,'w').write(s)"; clear`
Enter
Sleep 1s
Show

Type "# a rung that checks the work was added to bonsai.yaml"
Enter
Type "git diff bonsai.yaml | grep '^[+-] '"
Enter
Sleep 3s
Type "git commit -qam 'add a rung'"
Enter
Sleep 1s

# The climb is refused, with its next step.
Type "clear"
Enter
Type "bonsai ladder --task T-0001"
Enter
Sleep 6s

# The person's step, then the climb is green.
Type "clear"
Enter
Type "# a rung loosens the checks, so the person records it"
Enter
Type "bonsai update --yes"
Enter
Sleep 5s
Type "clear"
Enter
Type "bonsai ladder --task T-0001"
Enter
Sleep 6s
