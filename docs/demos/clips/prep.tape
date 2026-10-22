# prep.tape: not a clip. Starts Claude Code twice in the clip's project, off camera, so the folder-trust question is
# answered (the tape picks "Yes, I trust this folder") and Claude Code registers the project's plugin marketplace;
# film.sh then runs `bonsai update --yes`, which installs the packs' plugin. Its video is thrown away.
Output prep.mp4
Set Shell "bash"
Set FontSize 15
Set Width 1180
Set Height 700
Set TypingSpeed 15ms
Hide
Type "export PS1='$ '; cd $CLIP_PROJECT; clear"
Enter
Show
Type "claude --model sonnet"
Enter
Sleep 12s
Down
Enter
Sleep 20s
Type "/exit"
Sleep 1s
Enter
Sleep 4s
Type "claude --model sonnet"
Enter
Sleep 20s
Type "/exit"
Sleep 1s
Enter
Sleep 4s
