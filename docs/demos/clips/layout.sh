# layout.sh: sourced by a Claude Code clip's tape, in its hidden part. Builds the split screen on a tmux socket of its
# own and attaches to it, so what follows on the tape is typed into the left pane. Left: Claude Code's pane (pane 0,
# Alt+1), about 64% wide. Right: the live view (pane 1, Alt+2), view.sh. With CLIP_PERSON=1 the right side is split in
# two and its lower pane (Alt+3) is the person's own shell. Needs CLIPS and CLIP_PROJECT (film.sh sets them), and the
# environment (PATH, BONSAI_HOME, PS1) it hands every pane.
tmux -L clip kill-server 2>/dev/null
tmux -L clip -f "$CLIPS/tmux.conf" new-session -d -s clip -x "$(tput cols)" -y "$(tput lines)" -c "$CLIP_PROJECT"
tmux -L clip split-window -h -l 36% -t clip:0.0 -c "$CLIP_PROJECT" "sh '$CLIPS/view.sh'"
if [ "${CLIP_PERSON:-}" = 1 ]; then
  tmux -L clip split-window -v -l 50% -t clip:0.1 -c "$CLIP_PROJECT"
fi
tmux -L clip select-pane -t clip:0.0
exec tmux -L clip attach -t clip
