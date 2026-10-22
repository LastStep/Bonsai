# view.sh: the live view in the right-hand pane of the Claude Code clips: Bonsai's status line, then the newest
# decisions the log holds (guard, stop gate, ladder, asks), redrawn every second from the project's own log.
# Needs CLIP_BIN and CLIP_PROJECT (film.sh sets them).
cd "$CLIP_PROJECT" || exit 1
W=$(tput cols)
while :; do
  out=$(
    echo " bonsai status --line"
    echo
    "$CLIP_BIN" status --line | cut -c1-"$W" | sed 's/^/ /'
    echo
    echo " the log, newest decisions"
    echo
    cat .bonsai/local/log/s-*.ndjson 2>/dev/null | jq -rs '
      [ .[] | select(.decision != null or .event == "ladder" or .event == "ask" or .event == "stop") ]
      | sort_by(.at) | .[-16:] | .[]
      | " \(.at[11:19]) \(.event) \(.decision // "") \(.tool // "") \(.rule // "")"' 2>/dev/null | cut -c1-"$W"
  )
  printf '\033[H%s\033[J' "$out"
  sleep 1
done
