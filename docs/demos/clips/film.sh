#!/bin/sh
# film.sh: films the clip set, one mp4 a moment, on the installed bonsai. The person runs it in a plain terminal,
# never inside Claude Code (its own `!` lines count as an agent's session, and Bonsai refuses the person's steps there):
#
#   sh docs/demos/clips/film.sh <out folder> [clip ...]
#
# With no clip names it films all eleven in order; with names (the tapes' names without .tape, e.g. 04-ladder) only
# those. Every clip's project, Bonsai home and plugin folder are made new under a scratch folder (WORK, below), never
# the real ~/.bonsai or this repository. The scratch folder is /tmp/bonsai-film.XXXXXX by default, because Bonsai and
# Claude Code print the project's path on screen and that path must not hold a user name. Claude Code itself still
# writes its own files under ~/.claude (its session transcript and its folder-trust record); no plugin is touched,
# the plugin folder is the scratch one.
#
# WHAT MAY BE CHANGED (these three lines; nothing else is meant to change):
BONSAI=${BONSAI:-/usr/local/bin/bonsai}   # the bonsai to show. A test build for practice: the frames then show its path
MODEL=${MODEL:-sonnet}                    # Claude Code's model in the Claude clips (the tapes type it)
PACK_SOURCE=${PACK_SOURCE:-https://github.com/LastStep/Bonsai}   # where `bonsai init` fetches the base pack
WORK=${WORK:-}                            # the scratch folder, shown on screen: keep its path free of a user name

CLIPS=$(cd "$(dirname "$0")" && pwd)
ALL="01-link 02-check-status 03-rung 04-ladder 05-log 06-guard-edit 07-guard-update 08-guard-delete 09-stop-gate 10-ask-answer 11-status-line"

die() { echo "film.sh: $1" >&2; echo "next: $2" >&2; exit 2; }

[ $# -ge 1 ] || die "no out folder given" "sh docs/demos/clips/film.sh <out folder> [clip ...]"
mkdir -p "$1" || die "cannot make $1" "name a folder you can write"
OUT=$(cd "$1" && pwd); shift
CLIP_LIST=${*:-$ALL}

# Not inside Claude Code (FILM_PRACTICE=1 is for a practice run by an agent's runner on a test build, nothing else).
if [ -n "${CLAUDECODE:-}${CLAUDE_CODE_SESSION_ID:-}${BONSAI_TASK:-}" ] && [ "${FILM_PRACTICE:-}" != 1 ]; then
  die "this shell is inside a Claude Code session" "run it in a plain terminal"
fi
# What a plain terminal does not have (a runner's practice shell has all of it; CLAUDE_CODE_CHILD_SESSION alone makes
# bonsai believe there is no terminal to ask at).
for v in $(env | sed -n 's/^\(CLAUDE_CODE_[A-Z_]*\)=.*/\1/p'); do unset "$v"; done
unset CLAUDECODE CLAUDE_PID CLAUDE_EFFORT BONSAI_TASK BONSAI_ROLE BONSAI_HOME

for t in vhs ttyd ffmpeg tmux jq claude git python3; do
  command -v "$t" >/dev/null 2>&1 || die "$t is not installed or not on the PATH" "install $t, then run again"
done
[ -x "$BONSAI" ] || die "$BONSAI is not an executable bonsai" "install bonsai at /usr/local/bin/bonsai, or set BONSAI=<path> (practice)"
for c in $CLIP_LIST; do
  [ -f "$CLIPS/$c.tape" ] || die "no tape named $c" "the names are: $ALL"
done
CLIP_REF=$(git ls-remote "$PACK_SOURCE" refs/heads/main | cut -f1)
[ -n "$CLIP_REF" ] || die "cannot reach $PACK_SOURCE (the pack is fetched from it)" "check the network, then run again"

# shellcheck disable=SC1091
. "$CLIPS/lib.sh"
[ -n "$WORK" ] || WORK=$(mktemp -d /tmp/bonsai-film.XXXXXX)
mkdir -p "$WORK/plugin-cache" "$OUT/logs"
export CLIPS CLIP_REF
export CLIP_BIN="$BONSAI"
export CLAUDE_CODE_PLUGIN_CACHE_DIR="$WORK/plugin-cache"
export CLAUDE_CODE_FORCE_SESSION_PERSISTENCE=1
export GIT_AUTHOR_NAME=demo GIT_AUTHOR_EMAIL=demo@example.invalid GIT_COMMITTER_NAME=demo GIT_COMMITTER_EMAIL=demo@example.invalid
export PATH="$(dirname "$BONSAI"):$PATH"
export PS1='$ '

ok=""; failed=""
for c in $CLIP_LIST; do
  echo "== $c"
  root="$WORK/$c"
  rm -rf "$root"; mkdir -p "$root"
  export CLIP_PROJECT="$root/project" BONSAI_HOME="$root/home" CLIP_PERSON=""
  log="$OUT/logs/$c.log"
  : > "$log"
  case "$c" in
    01-link)          mkdir -p "$CLIP_PROJECT" ;;
    02-check-status)  mk_base "$CLIP_PROJECT" 1 ;;
    03-rung)          mk_base "$CLIP_PROJECT" 0 && mk_prep "$CLIP_PROJECT" && mk_update "$CLIP_PROJECT" ;;
    04-ladder|05-log) mk_base "$CLIP_PROJECT" 1 && mk_update "$CLIP_PROJECT" ;;
    10-ask-answer)    CLIP_PERSON=1; export CLIP_PERSON; mk_claude "$CLIP_PROJECT" ;;
    *)                mk_claude "$CLIP_PROJECT" ;;
  esac >> "$log" 2>&1
  if [ $? -ne 0 ]; then echo "   setup failed (see $log)"; failed="$failed $c"; continue; fi
  ( cd "$CLIP_PROJECT" && vhs -o "$OUT/$c.mp4" "$CLIPS/$c.tape" ) >> "$log" 2>&1
  rc=$?
  tmux -L clip kill-server >/dev/null 2>&1   # the Claude clips' split screen
  if [ $rc -eq 0 ] && [ -s "$OUT/$c.mp4" ]; then
    secs=$(ffprobe -v error -show_entries format=duration -of csv=p=0 "$OUT/$c.mp4" | cut -d. -f1)
    echo "   $OUT/$c.mp4 (${secs}s)"; ok="$ok $c"
  else
    echo "   vhs failed (see $log)"; failed="$failed $c"
  fi
done
echo
echo "scratch: $WORK (safe to delete)"
echo "made:${ok:- none}"
echo "failed:${failed:- none}"
[ -z "$failed" ]
