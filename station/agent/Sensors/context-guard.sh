#!/usr/bin/env bash
# Context Guard — Tech Lead Agent
# Injects behavioral constraints based on context usage and detects session wrap-up triggers.
# Reads state file written by status-bar sensor (from previous turn).

INPUT=$(cat)

SESSION_ID=$(echo "$INPUT" | python3 -c "import sys,json; print(json.load(sys.stdin).get('session_id',''))" 2>/dev/null)

if [[ -z "$SESSION_ID" ]]; then
  exit 0
fi

STATE_FILE="/tmp/bonsai-awareness-${SESSION_ID}.json"
PROMPT=$(echo "$INPUT" | python3 -c "import sys,json; print(json.load(sys.stdin).get('prompt',''))" 2>/dev/null)

# ── Analysis and injection (Python) ─────────────────────────────────────────

python3 -c "
import json, sys, os

state_file = sys.argv[1]
prompt = sys.argv[2]
docs_path = 'station/'

# ── Load state file ─────────────────────────────────────────────────────────

context_pct = 0
uncommitted = 0
state = None

try:
    if os.path.isfile(state_file):
        with open(state_file) as f:
            state = json.load(f)
        context_pct = state.get('context', {}).get('pct', 0)
        uncommitted = state.get('health', {}).get('uncommitted_files', 0)
except:
    pass

# ── Tiered context injection ────────────────────────────────────────────────

injection = ''

if context_pct >= 85:
    injection = (
        f'CONTEXT CRITICAL [{context_pct:.0f}%]: STOP accepting new work. '
        'Only respond if the user is committing, saving, or wrapping up. '
        'For any other request, respond: \"Context is nearly full. '
        'Please run /clear to start a fresh session.\" '
        'Do not use any file-reading tools.'
    )
elif context_pct >= 70:
    injection = (
        f'CONTEXT ALERT [{context_pct:.0f}%]: Complete ONLY the current task. '
        'Do NOT accept new work. When the current task completes, tell the user: '
        '\"Context is high. I recommend running /clear before starting anything new.\" '
        'Do not read large files.'
    )
elif context_pct >= 50:
    injection = (
        f'CONTEXT WARNING [{context_pct:.0f}%]: Be concise in responses. '
        'Focus on completing the current task. '
        'Do not start new exploratory work unless explicitly asked.'
    )
elif context_pct >= 30:
    injection = (
        f'CONTEXT ADVISORY [{context_pct:.0f}%]: Prefer targeted file reads '
        'over full-file reads. Avoid exploratory browsing.'
    )

# ── Session-done trigger word detection ─────────────────────────────────────

import re

# Normalize: strip apostrophes, lowercase
normalized = re.sub(r\"'\", '', prompt).lower()

# Regex patterns — anchored to end-of-prompt to prevent false positives
# e.g. \"thats all\" triggers, but \"thats all I need to change\" does not
# Handles contractions (thats/that's/that is) via apostrophe stripping
eol = r'[\\s.!?]*$'
trigger_patterns = [
    r'\\b(thats|that is)\\s+all' + eol,
    r'\\b(were|we are)\\s+done' + eol,
    r'\\b(im|i am)\\s+done' + eol,
    r'\\b(lets|let us)\\s+(wrap|wrap up)' + eol,
    r'\\bsession\\s+(done|over)' + eol,
    r'\\bend\\s+session' + eol,
    r'\\b(wrap|wrapping)\\s+up' + eol,
    r'\\bfinish\\s+up' + eol,
    r'\\bcall\\s+it' + eol,
]

triggered = any(re.search(p, normalized) for p in trigger_patterns)

if triggered:
    wrapup = (
        '\\nSESSION WRAP-UP TRIGGERED. '
        'Read and follow agent/Workflows/session-wrapup.md NOW — every step, in order. '
        'Do not skip steps or ask if you should run it. Just do it.'
    )
    injection = (injection + '\\n' + wrapup) if injection else wrapup

# ── Output ──────────────────────────────────────────────────────────────────

if not injection:
    sys.exit(0)

output = {
    'hookSpecificOutput': {
        'hookEventName': 'UserPromptSubmit',
        'additionalContext': injection,
    }
}
print(json.dumps(output))
" "$STATE_FILE" "$PROMPT" 2>/dev/null

exit 0
