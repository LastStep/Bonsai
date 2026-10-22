#!/bin/sh
# docs/demos/edit.sh: cuts the clip set (docs/demos/clips/, filmed by film.sh) into the README's two GIFs and the
# showcase video, with ffmpeg alone. The story, its captions and where each clip is cut are in showcase.txt beside
# this script; the picture on the showcase's right is how-it-works-dark.svg beside it (or PICTURE=<file>).
#
#   sh docs/demos/edit.sh <clips folder> <out folder>
#
# Reads <clips folder>/NN-name.mp4 (the eleven clips as film.sh names them: clips 01 to 05 are 1100 x 900, 06 to 11 are 1600 x 900;
# each whole-frame scene is cropped and scaled by its own clip's size, read with ffprobe) and writes into <out folder>:
#   first-project.gif   the README's top picture: link, check, a rung refused, the person's update, green (30-45 s)
#   session.gif         the README's "A real session": the guard refusing, the stop gate holding, then green (40-60 s)
#   showcase.mp4        2-3 minutes, 1920 x 1080: a title card, the story in the order a person meets Bonsai, an end card
# and <out folder>/work/ (the analysis of each clip, every cut scene, the stills; kept so a cut can be checked).
#
# How a clip is cut. Nothing is scripted but the typed lines, so a clip holds idle time (Claude thinking, a sleep
# before the next line) whose length differs from take to take. The script finds the frames where something appears
# (a frame-to-frame change above SCENE_T; a cursor blink or Claude Code's spinner is below it, a line of output above),
# keeps KEEP_A seconds after each such change (time to read it) and KEEP_B seconds before the next one (the typing of
# the next line), and plays what lies between at a speed that makes it FF_LEN seconds long, with a "sped up" tag on
# those frames and a note in the caption bar. The clip's tail after its last change is cut to TAIL seconds. The plan
# name in Claude Code's header is covered by a box the colour of the pane, tracked by the header's mascot, in every
# frame the header is on screen (at rest or scrolling off). Each output is looked at as stills in work/stills/.
#
# What to change for a new take: the times in showcase.txt (each scene's in and out points and the captions' times
# are in the clip's own seconds): read them off work/stills/<clip>.png (one frame every 2 s, numbered) and
# work/<clip>.big (the seconds at which something appeared). The rest is automatic.
#
# Settings, each an environment variable:
#   PICTURE   the How it works picture (an SVG or PNG); default how-it-works-dark.svg beside this script. If it is
#             missing, the showcase's right column holds the live view alone.
#   FONT, FONT_BOLD, MONO   the caption fonts; default DejaVu Sans, DejaVu Sans Bold and DejaVu Sans Mono (fc-list).
#   SCENE_T   the change threshold (default 0.0006).  KEEP_A, KEEP_B, FF_LEN, TAIL   seconds (default 3, 5.5, 1, 3.5).
#   ONLY      one of first-project, session, showcase: build that output alone (the analysis is shared and cached).
#   STILLS=0  skip the stills.
# The paths given must hold no ':' , ',' or quote (they are written into ffmpeg filter strings).
# Needs: ffmpeg and ffprobe (4.4 or newer, with libfreetype and, for the SVG, librsvg), awk, od. Linux; about three
# minutes for all three outputs on the practice clips.
#
# Modes besides the cut:
#   sh docs/demos/edit.sh --stills <video> <png> [seconds]   one frame every <seconds> (default 2), numbered, tiled
#   sh docs/demos/edit.sh --events <clip.mp4>                 the seconds at which something appears, one per line
set -eu

HERE=$(cd "$(dirname "$0")" && pwd)
FONT=${FONT:-/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf}
FONT_BOLD=${FONT_BOLD:-/usr/share/fonts/truetype/dejavu/DejaVuSans-Bold.ttf}
MONO=${MONO:-/usr/share/fonts/truetype/dejavu/DejaVuSansMono.ttf}
PICTURE=${PICTURE:-$HERE/how-it-works-dark.svg}
SCENE_T=${SCENE_T:-0.0006}
KEEP_A=${KEEP_A:-3}
KEEP_B=${KEEP_B:-5.5}
FF_LEN=${FF_LEN:-1}
TAIL=${TAIL:-3.5}
ONLY=${ONLY:-}
STILLS=${STILLS:-1}
GIF_W=830; GIF_C=32; GIF_FPS=10   # a GIF's width, colours and frames a second: showcase.txt's gif line sets them
gifset() { GIF_W=$1; GIF_C=$2; GIF_FPS=$3; GBAR=$((GIF_W*44/830)); GFS=$((GIF_W*19/830)); }
gifset $GIF_W $GIF_C $GIF_FPS
STORY=$HERE/showcase.txt
FPS=25

# The look: Catppuccin Mocha, the clips' own theme.
BG=0x11111b        # the page behind the clips (crust)
BAR=0x181825       # the caption bar (mantle)
PANE=0x1e1e2e      # Claude Code's pane (base): the colour of the box over the plan name
TEXT=0xcdd6f4
MUTED=0x9399b2
ACCENT=0xf9e2af    # the lit step in the picture, the sped-up tag
GREEN=0xa6e3a1
RED=0xf38ba8
BLUE=0x89b4fa

say() { printf '%s\n' "$*"; }
die() { printf 'edit.sh: %s\n' "$*" >&2; exit 1; }
ff() { ffmpeg -hide_banner -loglevel error -nostdin -y "$@"; }

need() {
  for t in ffmpeg ffprobe awk od; do command -v "$t" >/dev/null 2>&1 || die "$t is not installed"; done
  for f in "$FONT" "$FONT_BOLD" "$MONO"; do [ -f "$f" ] || die "font not found: $f (set FONT, FONT_BOLD or MONO; fc-list shows what is installed)"; done
  [ -f "$STORY" ] || die "showcase.txt not found beside this script"
}

duration() { ffprobe -v error -show_entries format=duration -of csv=p=0 "$1"; }

# ---- the modes besides the cut -------------------------------------------------------------------------------------

# stills <video> <png> [every s]: one frame every few seconds, each stamped with its own second, tiled six wide.
stills() {
  v=$1; png=$2; every=${3:-2}
  d=$(duration "$v")
  n=$(awk -v d="$d" -v e="$every" 'BEGIN{print int(d/e)+1}')
  rows=$(( (n + 5) / 6 ))
  ff -i "$v" -vf "select='isnan(prev_selected_t)+gte(t-prev_selected_t,$every-0.02)',drawtext=fontfile=$MONO:text='%{eif\:t\:d}s':fontsize=40:fontcolor=$ACCENT:box=1:boxcolor=black@0.6:boxborderw=8:x=12:y=12,scale=480:-2,tile=6x${rows}:padding=4:color=$BG" -frames:v 1 "$png"
}

# events <clip>: the seconds at which something appears on screen (a change above SCENE_T between two frames).
events() {
  ffmpeg -hide_banner -loglevel error -nostdin -i "$1" -vf "select='gt(scene,$SCENE_T)',metadata=print:key=lavfi.scene_score:file=-" -f null - \
    | awk '/pts_time/{split($0,a,"pts_time:"); printf "%.2f\n", a[2]+0}'
}

case ${1:-} in
  --stills) need; stills "$2" "$3" "${4:-2}"; say "$3"; exit 0 ;;
  --events) need; events "$2"; exit 0 ;;
esac
[ $# -eq 2 ] || { sed -n '2,13p' "$0" | sed 's/^# \{0,1\}//'; exit 2; }
need
CLIPS=$(cd "$1" && pwd) || die "no such folder: $1"
mkdir -p "$2" && OUT=$(cd "$2" && pwd)
W=$OUT/work
mkdir -p "$W" "$W/stills"
NAMES="01-link 02-check-status 03-rung 04-ladder 05-log 06-guard-edit 07-guard-update 08-guard-delete 09-stop-gate 10-ask-answer 11-status-line"
for n in $NAMES; do [ -f "$CLIPS/$n.mp4" ] || die "clip missing: $CLIPS/$n.mp4 (film.sh names them NN-name.mp4)"; done

# ---- the analysis, once per clip, cached in work/ -----------------------------------------------------------------

# analyze <name>: work/<name>.big (the seconds at which something appears) and work/<name>.box (the plan-name box:
# "from to y h" while Claude Code's header is on screen, tracked by the mascot's pink in the column x 30-100, in
# ten bands of 10 px from the top; a band is pink above V 148; at rest the mascot's last band is 7 (its feet end at
# y 79) and the plan name's row is y 65-80, so the box is y 60-85; once the header scrolls up, the mascot's feet
# (which stay whole after its top is cut) say by how much: the box moves up by 10 px a band lost and grows to 35 px
# tall, since a band is only that exact; the first version followed the mascot's top and missed the plan name when
# the top was cut off). One decode per clip.
analyze() {
  n=$1; src=$CLIPS/$n.mp4
  [ -f "$W/$n.big" ] && [ -f "$W/$n.box" ] && [ -f "$W/$n.len" ] && return 0
  duration "$src" > "$W/$n.len"
  ffmpeg -hide_banner -loglevel error -nostdin -i "$src" -filter_complex \
    "[0:v]split[a][b];[a]select='gt(scene,$SCENE_T)',metadata=print:key=lavfi.scene_score:file=$W/$n.scene[o1];[b]crop=70:100:30:0,scale=1:10:flags=area[o2]" \
    -map '[o1]' -f null - -map '[o2]' -pix_fmt yuv444p -f rawvideo "$W/$n.bands"
  awk '/pts_time/{split($0,a,"pts_time:"); printf "%.2f\n", a[2]+0}' "$W/$n.scene" > "$W/$n.big"
  rm -f "$W/$n.scene"
  # the bands: 30 bytes a frame (10 Y, 10 U, 10 V); the topmost pink band gives the mascot's top.
  od -An -v -tu1 -w30 "$W/$n.bands" | awk -v fps="$FPS" '
    { top=-1; bot=-1; for (i=21; i<=30; i++) if ($i+0 > 148) { if (top < 0) top=i-21; bot=i-21 }
      if (top < 0) { y=-1; h=0 }
      else {
        off=(7-bot)*10; if (off < 0) off=0
        if (off == 0) { y=60; h=25 } else { y=55-off; h=35; if (y < 0) { h=h+y; y=0 }; if (h <= 0) { y=-1; h=0 } }
      }
      t=NR-1
      if (NR == 1) { cy=y; ch=h; from=0; next }
      if (y != cy || h != ch) { if (cy >= 0) printf "%.2f %.2f %d %d\n", from/fps, (t+0.5)/fps, cy, ch; cy=y; ch=h; from=t }
    }
    END { if (cy >= 0) printf "%.2f %.2f %d %d\n", from/fps, (t+1)/fps, cy, ch }' > "$W/$n.box"
  rm -f "$W/$n.bands"
}

# plan <name> <in> <out|auto>: work/<name>-<in>-<out>.plan, the segments "start end k" in the clip's seconds (k = 1,
# or the speed of a fast-forwarded stretch) and work/...map, each segment's start in the cut's seconds.
plan() {
  n=$1; in=$2; out=$3; len=$(cat "$W/$n.len")
  if [ "$out" = auto ]; then
    out=$(awk -v ip="$in" -v len="$len" -v tail="$TAIL" '$1>ip{last=$1} END{o=(last?last:ip)+tail; if (o>len) o=len; printf "%.2f", o}' "$W/$n.big")
  fi
  awk -v ip="$in" -v len="$len" '$1>len{exit} END{if (ip>len) exit 1}' "$W/$n.big" || die "$n: in point $in is past the clip's end ($len s)"
  tag="$n-$in-$out"
  awk -v ip="$in" -v out="$out" -v A="$KEEP_A" -v B="$KEEP_B" -v F="$FF_LEN" -v tail="$TAIL" '
    BEGIN { m=0; ev[m++]=ip+0 }
    { t=$1+0; if (t > ip+0.2 && t < out-0.2) ev[m++]=t }
    END {
      ev[m++]=out+0; cs=ip+0
      for (i=0; i<m-1; i++) {
        if (i == m-2) {                       # the tail: after the last change, TAIL seconds at most
          if (i == 0) e=out; else { e=ev[i]+tail; if (e > out) e=out }
          printf "%.2f %.2f 1\n", cs, e
          break
        }
        g=ev[i+1]-ev[i]
        if (g > A+B+1.5) {                    # a long wait: keep A after the change and B before the next, fast-forward between
          e1=ev[i]+A; s2=ev[i+1]-B; k=int((s2-e1)/F+0.5); if (k < 2) k=2
          printf "%.2f %.2f 1\n", cs, e1; printf "%.2f %.2f %d\n", e1, s2, k; cs=s2
        }
      }
    }' "$W/$n.big" > "$W/$tag.plan"
  awk '{ printf "%s %s %s %.3f\n", $1, $2, $3, o; o += ($2-$1)/$3 }' "$W/$tag.plan" > "$W/$tag.map"
  printf '%s\n' "$tag"
}

# condense <name> <tag> <tagx> <tagy>: work/<tag>.mp4, the clip cut by its plan at the clip's own size, the plan
# name boxed, the sped-up tag drawn (its right edge at <tagx>, its top at <tagy>) on the fast-forwarded frames.
condense() {
  n=$1; tag=$2; tagx=$3; tagy=$4; src=$CLIPS/$n.mp4
  [ -f "$W/$tag.mp4" ] && [ "$W/$tag.mp4" -nt "$W/$n.box" ] && [ "$W/$tag.mp4" -nt "$W/$tag.plan" ] && [ "$W/$tag.mp4" -nt "$0" ] && return 0
  f="[0:v]"
  while read -r a b y h; do
    f="${f}drawbox=x=120:y=$y:w=280:h=$h:color=$PANE:t=fill:enable='between(t,$a,$b)',"
  done < "$W/$n.box"
  sel=""; ffd=""
  while read -r s e k; do
    if [ "$k" = 1 ]; then sel="$sel+between(t,$s,$e)"
    else
      sel="$sel+between(t,$s,$e)*not(mod(n,$k))"
      ffd="${ffd}drawtext=fontfile=$FONT_BOLD:text='>> sped up':fontsize=26:fontcolor=$ACCENT:box=1:boxcolor=$BG@0.75:boxborderw=10:x=$tagx-tw:y=$tagy:enable='between(t,$s,$e)',"
    fi
  done < "$W/$tag.plan"
  sel=${sel#+}
  f="${f}${ffd}select='$sel',setpts=N/($FPS*TB)[out]"
  ff -i "$src" -filter_complex "$f" -map '[out]' -r $FPS -c:v libx264 -preset veryfast -crf 16 -pix_fmt yuv420p "$W/$tag.mp4"
}

# ---- the picture, lit ------------------------------------------------------------------------------------------------

# The How it works picture at 820 x 470, and a copy with one step outlined, for each step the story names. The
# rectangles are the SVG's own (its viewBox is 820 x 470).
picture() {
  [ -f "$PICTURE" ] || { say "no picture at $PICTURE: the showcase's right column holds the live view alone"; return 0; }
  [ -f "$W/pic-none.png" ] && return 0
  ff -i "$PICTURE" -vf "scale=820:470" "$W/pic-none.png"
  for step in "agent 20 40 180 400" "guard 260 70 150 80" "allowed 470 76 330 36" "refused 470 126 330 96" \
              "log 260 250 540 56" "ladder 260 330 150 100" "green 470 330 330 36" "red 470 380 330 60"; do
    set -- $step
    ff -i "$W/pic-none.png" -vf "drawbox=x=$(($2-7)):y=$(($3-7)):w=$(($4+14)):h=$(($5+14)):color=$ACCENT:t=3" "$W/pic-$1.png"
  done
}

# ---- the scenes ------------------------------------------------------------------------------------------------------

# A caption's time, from the clip's seconds to the cut's seconds (through the plan's map), then the scene's speed.
outtime() { # <map> <clip s> <speed>
  awk -v c="$2" -v g="$3" '{ s=$1; e=$2; k=$3; o=$4; if (c >= s && c <= e) { printf "%.2f", (o+(c-s)/k)/g; exit } last=(o+(e-s)/k)/g } END{ if (c > e) printf "%.2f", last }' "$1"
}

# tf <text>: the text in a new file, its path in TF (drawtext reads the file, so nothing needs escaping). No
# subshell: the counter must survive the call.
NT=0
tf() { NT=$((NT+1)); TF=$W/t$NT.txt; printf '%s' "$1" > "$TF"; }

# card <out> <seconds> <kind> <a> <b> <c>: a title card (kind title: a name, two lines under it), a chapter card
# (kind chapter: a number and a title) or the end card (kind end: a name and a line under it); 1920 x 1080.
card() {
  o=$1; d=$2; kind=$3; a=$4; b=$5; c=$6
  tf "$a"; ta=$TF; tf "$b"; tb=$TF; tf "$c"; tc=$TF
  case $kind in
    title)   t="drawtext=fontfile=$FONT_BOLD:textfile=$ta:fontsize=150:fontcolor=$TEXT:x=(w-tw)/2:y=330,drawtext=fontfile=$FONT:textfile=$tb:fontsize=46:fontcolor=$TEXT:x=(w-tw)/2:y=560,drawtext=fontfile=$FONT:textfile=$tc:fontsize=36:fontcolor=$MUTED:x=(w-tw)/2:y=640" ;;
    chapter) t="drawtext=fontfile=$MONO:textfile=$ta:fontsize=44:fontcolor=$ACCENT:x=(w-tw)/2:y=430,drawtext=fontfile=$FONT_BOLD:textfile=$tb:fontsize=80:fontcolor=$TEXT:x=(w-tw)/2:y=500" ;;
    end)     t="drawtext=fontfile=$FONT_BOLD:textfile=$ta:fontsize=110:fontcolor=$TEXT:x=(w-tw)/2:y=380,drawtext=fontfile=$MONO:textfile=$tb:fontsize=50:fontcolor=$BLUE:x=(w-tw)/2:y=560,drawtext=fontfile=$FONT:textfile=$tc:fontsize=34:fontcolor=$MUTED:x=(w-tw)/2:y=660" ;;
  esac
  fo=$(awk -v d="$d" 'BEGIN{printf "%.2f", d-0.4}')
  ff -f lavfi -i "color=c=$BG:s=1920x1080:r=$FPS:d=$d" -vf "$t,fade=t=in:st=0:d=0.4,fade=t=out:st=$fo:d=0.4" -c:v libx264 -preset veryfast -crf 18 -pix_fmt yuv420p "$o"
}

# scene <out> <profile> <name> <layout> <in> <out|auto> <speed> <chapter> <caps file>: one cut clip, composed. Profile
# show is 1920 x 1080 with a caption bar of 110 px; profile gif is 830 wide with a bar of 44 px. Layout whole: the
# whole frame of the clip (its own width and height: 1100 x 900 or 1600 x 900), centred in the showcase; claude: Claude Code's pane (the left 1010 px) and, in profile show, the picture with its lit
# step over the live view (the right pane's top 380 px) in a right column. The caps file: "from to lit|text" lines in
# the clip's seconds; lit names the picture's step (- for none).
scene() {
  o=$1; prof=$2; n=$3; layout=$4; in=$5; outp=$6; g=$7; chapter=$8; caps=$9
  analyze "$n"
  tag=$(plan "$n" "$in" "$outp")
  cw=$(ffprobe -v error -select_streams v:0 -show_entries stream=width -of csv=p=0 "$CLIPS/$n.mp4" 2>/dev/null || true)
  ch=$(ffprobe -v error -select_streams v:0 -show_entries stream=height -of csv=p=0 "$CLIPS/$n.mp4" 2>/dev/null || true)
  [ -n "$cw" ] && [ -n "$ch" ] || die "cannot read the size of clip $n" "check that the clip is a video"
  tagx=$((cw-20)); tagy=$((ch-70)); [ "$layout" = claude ] && { tagx=990; tagy=22; }   # a whole frame: bottom right
  condense "$n" "$tag" "$tagx" "$tagy"
  map=$W/$tag.map
  D=$(awk -v g="$g" '{d=$4+($2-$1)/$3} END{printf "%.2f", d/g}' "$map")
  fast=$(awk '$3>1{f=1} END{print f+0}' "$map")
  # the captions (and the lit step, in profile show) as drawtext/overlay filters, each enabled for its cut-time range
  capf=""; litf=""; i=0
  while IFS='|' read -r times text; do
    [ -n "$times" ] || continue
    set -- $times; a=$(outtime "$map" "$1" "$g"); b=$(outtime "$map" "$2" "$g"); lit=$3
    [ "$2" = end ] && b=$D
    tf "$text"
    if [ "$prof" = show ]; then
      capf="${capf}drawtext=fontfile=$FONT:textfile=$TF:fontsize=36:fontcolor=$TEXT:x=60:y=970+55-th/2:enable='between(t,$a,$b)',"
      if [ "$lit" != - ] && [ -f "$W/pic-$lit.png" ]; then
        i=$((i+1)); litf="$litf -i $W/pic-$lit.png"; LITS="$LITS $i:$a:$b"
      fi
    else
      capf="${capf}drawtext=fontfile=$FONT:textfile=$TF:fontsize=$GFS:fontcolor=$TEXT:x=14:y=h-$((GBAR/2))-th/2:enable='between(t,$a,$b)',"
    fi
  done < "$caps"
  fo=$(awk -v d="$D" 'BEGIN{printf "%.2f", d-0.3}')
  if [ "$prof" = show ]; then
    note=""
    [ "$fast" = 1 ] && note="drawtext=fontfile=$FONT:text='idle time sped up':fontsize=24:fontcolor=$MUTED:x=w-tw-60:y=1038,"
    tf "$chapter"; chap="drawtext=fontfile=$FONT:textfile=$TF:fontsize=24:fontcolor=$MUTED:x=w-tw-60:y=984,"
    case $layout in
      whole)
        f="color=c=$BG:s=1920x1080:r=$FPS[bg];[0:v]setpts=PTS/$g,crop=$cw:$ch:0:0[a];[bg][a]overlay=(W-w)/2:35:shortest=1[c]" ;;
      claude)
        f="color=c=$BG:s=1920x1080:r=$FPS[bg];[0:v]setpts=PTS/$g,split[v1][v2];[v1]crop=1010:900:0:0[a];[v2]crop=584:380:1016:0[b];[bg][a]overlay=30:35[c1]"
        if [ -f "$W/pic-none.png" ]; then
          f="$f;[c1][1:v]overlay=1060:35[p0]"; k=0
          for l in ${LITS:-}; do idx=${l%%:*}; r=${l#*:}; a=${r%%:*}; b=${r#*:}; f="$f;[p$k][$((idx+1)):v]overlay=1060:35:enable='between(t,$a,$b)'[p$((k+1))]"; k=$((k+1)); done
          f="$f;[p$k][b]overlay=1178:540:shortest=1[c]"
        else
          f="$f;[c1][b]overlay=1178:35:shortest=1[c]"
        fi ;;
      *) die "unknown layout $layout" ;;
    esac
    f="$f;[c]drawbox=x=0:y=970:w=1920:h=110:color=$BAR:t=fill,${capf}${chap}${note}fade=t=in:st=0:d=0.3,fade=t=out:st=$fo:d=0.3,fps=$FPS[out]"
    # shellcheck disable=SC2086
    ff -i "$W/$tag.mp4" $( [ -f "$W/pic-none.png" ] && [ "$layout" = claude ] && printf '%s' "-i $W/pic-none.png $litf" ) -filter_complex "$f" -map '[out]' -t "$D" -c:v libx264 -preset veryfast -crf 18 -pix_fmt yuv420p "$o"
  else
    case $layout in
      whole)  f="[0:v]setpts=PTS/$g,crop=$cw:$ch:0:0,scale=$GIF_W:-2[a]" ;;
      claude) f="[0:v]setpts=PTS/$g,crop=1010:900:0:0,scale=$GIF_W:-2[a]" ;;
    esac
    note=""
    [ "$fast" = 1 ] && note="drawtext=fontfile=$FONT:text='idle time sped up':fontsize=$((GFS*4/5)):fontcolor=$MUTED:x=w-tw-14:y=h-$((GBAR/2))-th/2,"
    f="$f;[a]pad=$GIF_W:ih+$GBAR:0:0:$BAR,${capf}${note}fps=$FPS[out]"
    ff -i "$W/$tag.mp4" -filter_complex "$f" -map '[out]' -t "$D" -c:v libx264 -preset veryfast -crf 14 -pix_fmt yuv420p "$o"
  fi
  LITS=""
}

# ---- the story: showcase.txt --------------------------------------------------------------------------------------
# Sections [showcase], [first-project], [session]. Lines (fields after the first are split on |):
#   card <s> | <name> | <line> | <line>         chapter <s> | <number> | <title>        end <s> | <name> | <url> | <line>
#   scene <clip> <layout> <in> <out|auto> <speed> | <chapter label>
#   cap <from> <to|end> <lit|-> | <caption>     (the scene's captions, in the clip's seconds)

# build <section> <profile> <out list file>: every card and scene of a section into work/<section>-NN.mp4, listed.
build() {
  sec=$1; prof=$2; list=$3
  : > "$list"; nseg=0; cur=""; curchap=""; capsf=""
  flush() {
    if [ -n "$cur" ]; then
      nseg=$((nseg+1)); o=$W/$sec-$(printf '%02d' $nseg).mp4
      set -- $cur
      say "  scene $1 ($3 to $4, x$5)"
      scene "$o" "$prof" "$1" "$2" "$3" "$4" "$5" "$curchap" "$capsf" >/dev/null
      printf "file '%s'\n" "$o" >> "$list"
      cur=""
    fi
  }
  insec=0
  while IFS= read -r line || [ -n "$line" ]; do
    case $line in
      '#'*|'') continue ;;
      "[$sec]") insec=1; continue ;;
      '['*) [ $insec = 1 ] && break; continue ;;
    esac
    [ $insec = 1 ] || continue
    kind=${line%% *}; rest=${line#* }
    head=$(printf '%s' "$rest" | awk -F'|' '{gsub(/^ +| +$/,"",$1); print $1}')
    f2=$(printf '%s' "$rest" | awk -F'|' '{gsub(/^ +| +$/,"",$2); print $2}')
    f3=$(printf '%s' "$rest" | awk -F'|' '{gsub(/^ +| +$/,"",$3); print $3}')
    f4=$(printf '%s' "$rest" | awk -F'|' '{gsub(/^ +| +$/,"",$4); print $4}')
    case $kind in
      card|chapter|end)
        flush; nseg=$((nseg+1)); o=$W/$sec-$(printf '%02d' $nseg).mp4
        say "  $kind $head s"
        card "$o" "$head" "$([ "$kind" = card ] && echo title || echo "$kind")" "$f2" "$f3" "$f4"
        printf "file '%s'\n" "$o" >> "$list" ;;
      scene)
        flush; set -- $head
        [ $# -eq 5 ] || die "showcase.txt: scene needs <clip> <layout> <in> <out> <speed>: $line"
        cur="$1 $2 $3 $4 $5"; curchap=$f2; capsf=$W/caps-$sec-$nseg.txt; : > "$capsf" ;;
      cap)
        [ -n "$cur" ] || die "showcase.txt: a cap line with no scene before it: $line"
        printf '%s|%s\n' "$head" "$f2" >> "$capsf" ;;
      gif)
        set -- $head; [ $# -eq 3 ] || die "showcase.txt: gif needs <width> <colours> <fps>: $line"
        gifset "$1" "$2" "$3" ;;
      *) die "showcase.txt: unknown line: $line" ;;
    esac
  done < "$STORY"
  flush
}

# gif <list> <out.gif>: the section's scenes joined, then one palette for the whole GIF, no dither (the terminal's
# colours are flat, and dither patterns cost most of a GIF's size), only the changed rectangle stored each frame.
gif() {
  src=${2%.gif}.mp4
  ff -f concat -safe 0 -i "$1" -c copy "$W/$(basename "$src")"
  ff -i "$W/$(basename "$src")" -filter_complex "fps=$GIF_FPS,split[a][b];[a]palettegen=max_colors=$GIF_C:stats_mode=diff[p];[b][p]paletteuse=dither=none:diff_mode=rectangle" "$2"
}

LITS=""
say "edit.sh: clips $CLIPS, out $OUT"
say "analysing the clips (once; cached in work/)"
for n in $NAMES; do analyze "$n"; done
picture

if [ -z "$ONLY" ] || [ "$ONLY" = first-project ]; then
  say "first-project.gif"
  build first-project gif "$W/first-project.list"
  gif "$W/first-project.list" "$OUT/first-project.gif"
fi
if [ -z "$ONLY" ] || [ "$ONLY" = session ]; then
  say "session.gif"
  build session gif "$W/session.list"
  gif "$W/session.list" "$OUT/session.gif"
fi
if [ -z "$ONLY" ] || [ "$ONLY" = showcase ]; then
  say "showcase.mp4"
  build showcase show "$W/showcase.list"
  ff -f concat -safe 0 -i "$W/showcase.list" -c copy -movflags +faststart "$OUT/showcase.mp4"
fi

say "the outputs:"
for o in first-project.gif session.gif showcase.mp4; do
  [ -f "$OUT/$o" ] || continue
  d=$(duration "$OUT/$o" 2>/dev/null || echo ?)
  s=$(wc -c < "$OUT/$o")
  say "  $OUT/$o  $(awk -v s="$s" 'BEGIN{printf "%.2f MB", s/1048576}')  $(awk -v d="$d" 'BEGIN{printf "%.0f s", d}')"
  [ "$STILLS" = 1 ] && stills "$OUT/$o" "$W/stills/${o%.*}.png" 3 >/dev/null && say "    stills: $W/stills/${o%.*}.png"
done
exit 0
