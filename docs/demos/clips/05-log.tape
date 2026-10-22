# 05-log: the log, every decision Bonsai made in order, read through jq (the climbs are made off camera first).
# Crop: the whole frame. Raw length: about 19 s.
#
# Run by film.sh (sh docs/demos/clips/film.sh <out folder> 05-log), which makes the project, the scratch Bonsai home and
# the environment first (CLIPS, CLIP_BIN, CLIP_PROJECT, CLIP_REF) and runs this tape from the project's folder. To run
# it alone, see README.md ("Filming one clip again"). Every command shown is run; nothing but the typed lines is scripted.
Output "05-log.mp4"
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
Type `bonsai ladder --task T-0001 >/dev/null; sed -i 's/split(" ")/split()/' words.py; git commit -qam fix; bonsai ladder --task T-0001 >/dev/null; bonsai ask --type Answer --task T-0001 --key ship --title 'Ship it?' --why 'the tests are green' >/dev/null; bonsai answer agent:ship --words yes >/dev/null; clear`
Enter
Sleep 20s
Show

Type "# everything Bonsai records is in the log, one JSON line each"
Enter
Type "bonsai logs"
Enter
Sleep 4s
Type `bonsai logs --day $(date -u +%F) | jq -r '[.at[11:19], .event, (.task // "-"), (.kind // "-"), (if .ok == null then "-" else (.ok | tostring) end)] | @tsv'`
Enter
Sleep 6s
