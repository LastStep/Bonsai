# 01-link: a new project is linked to Bonsai: the preview of what it will write, the person's yes, then the status.
# Crop: the whole frame. Raw length: about 21 s (the pack's download is cut here with Hide).
#
# Run by film.sh (sh docs/demos/clips/film.sh <out folder> 01-link), which makes the project, the scratch Bonsai home and
# the environment first (CLIPS, CLIP_BIN, CLIP_PROJECT, CLIP_REF) and runs this tape from the project's folder. To run
# it alone, see README.md ("Filming one clip again"). Every command shown is run; nothing but the typed lines is scripted.
Output "01-link.mp4"
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

Type "# a new project, linked to Bonsai's own pack"
Enter
Type "git init -q -b main"
Enter
Sleep 500ms
# Without --yes, init previews every file and settings line, then asks.
# film.sh writes the address and the commit out in a copy of this tape, so the frame shows what to type.
Type "bonsai init --name demo --source $PACK_SOURCE --path packs/base --ref $CLIP_REF"
# long enough to read the address and the commit before Enter
Sleep 3s
Enter
# the pack is fetched from GitHub first, which takes a while: that wait is not shown
Hide
Sleep 40s
Show
Sleep 6s
Type "y"
Enter
Sleep 5s
Type "clear"
Enter
Type "bonsai status"
Enter
Sleep 5s
