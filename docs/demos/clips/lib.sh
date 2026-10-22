# lib.sh: the steps film.sh uses to make a clip's project off camera. Sourced by film.sh, never run alone. POSIX sh.
# Needs CLIP_BIN (the bonsai to show), CLIP_REF (the pack's commit), CLIPS (this folder), and BONSAI_HOME already
# exported by film.sh for the clip. Each function takes the project's folder.

# mk_base DIR RUNG: a git project linked to Bonsai's base pack with task T-0001 running and its run report, and a
# word counter whose tests fail (two). RUNG=1 puts the tests in the ladder as rung 1 (on the floor); RUNG=0 leaves the
# ladder empty. Committed.
mk_base() {
  d=$1; rung=$2
  today=$(date -u +%Y-%m-%d)
  rm -rf "$d"; mkdir -p "$d"; cd "$d" || return 1
  git init -q -b main
  git config user.name demo; git config user.email demo@example.invalid
  "$CLIP_BIN" init --name demo --source https://github.com/LastStep/Bonsai --path packs/base --ref "$CLIP_REF" --yes || return 1
  cat > words.py <<'PY'
def count_words(text):
    """How many words are in text."""
    return len(text.split(" "))
PY
  cat > test_words.py <<'PY'
import unittest

from words import count_words


class CountWords(unittest.TestCase):
    def test_plain(self):
        self.assertEqual(count_words("one two three"), 3)

    def test_extra_spaces(self):
        self.assertEqual(count_words("one  two   three"), 3)

    def test_empty(self):
        self.assertEqual(count_words(""), 0)


if __name__ == "__main__":
    unittest.main()
PY
  printf '%s\n' '__pycache__/' '.claude/settings.local.json' > .gitignore
  if [ "$rung" = 1 ]; then
    python3 -I - <<'PY'
p = "bonsai.yaml"
s = open(p).read()
s = s.replace("ladder_floor: []", "ladder_floor: [1]").replace("\nladder: []\n", "\nladder:\n  - rung: 1\n    name: tests\n    kind: command\n    command: python3 -m unittest -q\n    required: true\n")
open(p, "w").write(s)
PY
  fi
  mkdir -p work/tasks work/runs
  printf '%s\n' '---' 'format: bonsai.task/1' 'id: T-0001' 'title: "Count words across extra spaces"' 'status: running' 'lane:' 'done_when:' '  - "the tests pass and the ladder is green"' 'depends_on: []' 'blocked_by:' "created: $today" "started: $today" 'finished:' 'labels: {}' '---' '' '# Count words across extra spaces' '' 'count_words miscounts text with more than one space between words.' > work/tasks/T-0001-count-words.md
  printf '%s\n' '---' 'format: bonsai.run/1' "id: R-$today-T-0001" 'task: T-0001' 'role: builder' 'model: sonnet' "started: $today 08:00" 'finished: null' 'outcome: running' 'commits: []' 'labels: {}' '---' '' '# Run: count words across extra spaces' '' '> Opened at the start of the session and added to as it goes.' '' '## What is different now' '' 'Nothing yet.' > "work/runs/R-$today-T-0001-count-words.md"
  git add -A && git commit -qm "link to Bonsai, a word counter and its tests"
}

# mk_update DIR: the person's step: records the settings (and installs the packs' plugin); committed.
mk_update() {
  cd "$1" || return 1
  "$CLIP_BIN" update --yes || return 1
  git add -A; git diff --cached --quiet || git commit -qm "bonsai update"
}

# mk_local DIR: the clip's own Claude Code rules in .claude/settings.local.json (out of git): the commands the story
# needs run without Claude Code's own permission question (Bonsai's guard still judges every call), and Bonsai's status
# line sits in Claude Code's bar. Narrow rules only: no blanket Bash allow.
mk_local() {
  cd "$1" || return 1
  cat > .claude/settings.local.json <<JSON
{
  "statusLine": {"type": "command", "command": "$CLIP_BIN status --line"},
  "permissions": {
    "defaultMode": "acceptEdits",
    "allow": [
      "Bash(python3 -m unittest:*)",
      "Bash(*bonsai ladder*)",
      "Bash(*bonsai status*)",
      "Bash(*bonsai ask*)",
      "Bash(*bonsai update*)",
      "Bash(git status:*)",
      "Bash(git diff:*)",
      "Bash(git add:*)",
      "Bash(git commit:*)",
      "Bash(git log:*)",
      "Bash(tail:*)",
      "Bash(head:*)",
      "Bash(cat:*)",
      "Bash(ls:*)",
      "Bash(jq:*)",
      "Bash(grep:*)"
    ]
  }
}
JSON
}

# mk_red DIR: a red climb is recorded, so the stop gate has a result to read.
mk_red() {
  cd "$1" || return 1
  "$CLIP_BIN" ladder --task T-0001 >/dev/null 2>&1 || true
}

# mk_prep DIR: Claude Code is started twice in the project off camera (prep.tape), as a person's first start would, so
# the next mk_update installs the packs' plugin.
mk_prep() {
  ( cd "$1" && vhs -o "$1/../prep.mp4" "$CLIPS/prep.tape" ) >/dev/null 2>&1
}

# mk_claude DIR: the project a Claude Code clip starts in: linked, the packs' plugin installed, rules, a red climb.
mk_claude() {
  d=$1
  mk_base "$d" 1 || return 1
  mk_prep "$d"
  mk_update "$d" || return 1
  mk_local "$d" || return 1
  mk_red "$d"
  cd "$d" && git status --short
}
