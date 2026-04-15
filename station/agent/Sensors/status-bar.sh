#!/usr/bin/env bash
# Status Bar — Tech Lead Agent
# Persistent status line after every response: context %, turns, tools, session health.
# Writes state file for context-guard to read on next turn.

INPUT=$(cat)

# Guard: if stop_hook_active, another Stop hook already blocked this turn — bail out
STOP_ACTIVE=$(echo "$INPUT" | python3 -c "import sys,json; print(json.load(sys.stdin).get('stop_hook_active', False))" 2>/dev/null)
if [[ "$STOP_ACTIVE" == "True" ]]; then
  exit 0
fi

TRANSCRIPT=$(echo "$INPUT" | python3 -c "import sys,json; print(json.load(sys.stdin).get('transcript_path',''))" 2>/dev/null)
SESSION_ID=$(echo "$INPUT" | python3 -c "import sys,json; print(json.load(sys.stdin).get('session_id',''))" 2>/dev/null)

if [[ -z "$TRANSCRIPT" || ! -f "$TRANSCRIPT" ]]; then
  exit 0
fi

ROOT="${1:-.}"
WORKSPACE="${ROOT}/station/"
STATE_FILE="/tmp/bonsai-awareness-${SESSION_ID}.json"

# ── Main analysis (Python) ──────────────────────────────────────────────────

python3 -c "
import json, os, sys, subprocess, time
from datetime import datetime

transcript_path = sys.argv[1]
state_file = sys.argv[2]
workspace = sys.argv[3]

file_size = os.path.getsize(transcript_path)

# ── Parse transcript ────────────────────────────────────────────────────────

user_turns = 0
assistant_turns = 0
tool_calls = 0
total_chars = 0

with open(transcript_path) as f:
    for line in f:
        line = line.strip()
        if not line:
            continue
        try:
            entry = json.loads(line)
        except:
            continue

        entry_type = entry.get('type', '')
        msg = entry.get('message', {})
        content = msg.get('content', '')

        if entry_type == 'user':
            user_turns += 1
            if isinstance(content, str):
                total_chars += len(content)
            elif isinstance(content, list):
                for block in content:
                    if isinstance(block, dict):
                        total_chars += len(block.get('text', ''))
                    elif isinstance(block, str):
                        total_chars += len(block)

        elif entry_type == 'assistant':
            assistant_turns += 1
            if isinstance(content, str):
                total_chars += len(content)
            elif isinstance(content, list):
                for block in content:
                    if isinstance(block, dict):
                        bt = block.get('type', '')
                        if bt == 'text':
                            total_chars += len(block.get('text', ''))
                        elif bt == 'tool_use':
                            tool_calls += 1
                            inp = block.get('input', {})
                            total_chars += len(json.dumps(inp)) if isinstance(inp, dict) else len(str(inp))
                        elif bt == 'tool_result':
                            rc = block.get('content', '')
                            if isinstance(rc, str):
                                total_chars += len(rc)
                            elif isinstance(rc, list):
                                for rb in rc:
                                    if isinstance(rb, dict):
                                        total_chars += len(rb.get('text', ''))

        elif entry_type == 'system':
            c = entry.get('content', '')
            if isinstance(c, str):
                total_chars += len(c)

est_tokens = int(total_chars / 3.5)
context_pct = min((est_tokens / 200000) * 100, 100)

# ── Health checks ───────────────────────────────────────────────────────────

# Git: uncommitted file count (with timeout)
uncommitted = -1
try:
    result = subprocess.run(
        ['git', 'status', '--porcelain'],
        capture_output=True, text=True, timeout=2, cwd=os.getcwd()
    )
    if result.returncode == 0:
        lines = [l for l in result.stdout.strip().split('\n') if l.strip()]
        uncommitted = len(lines)
except:
    pass

# Memory staleness
memory_stale_days = -1
memory_path = os.path.join(workspace, 'agent', 'Core', 'memory.md')
try:
    mtime = os.path.getmtime(memory_path)
    memory_stale_days = int((time.time() - mtime) / 86400)
except:
    pass

# Overdue routines
overdue_routines = 0
dashboard_path = os.path.join(workspace, 'agent', 'Core', 'routines.md')
try:
    if os.path.isfile(dashboard_path):
        in_table = False
        with open(dashboard_path) as df:
            for dline in df:
                dline = dline.strip()
                if 'ROUTINE_DASHBOARD_START' in dline:
                    in_table = True
                    continue
                if 'ROUTINE_DASHBOARD_END' in dline:
                    break
                if not in_table:
                    continue
                if 'Routine' in dline or '---' in dline:
                    continue
                if dline.startswith('|'):
                    cols = [c.strip() for c in dline.split('|')]
                    # cols: ['', routine, frequency, last_ran, next_due, status, '']
                    if len(cols) < 5:
                        continue
                    freq_str = cols[2]
                    last_ran = cols[3]
                    import re
                    freq_match = re.search(r'(\d+)', freq_str)
                    if not freq_match:
                        continue
                    freq_days = int(freq_match.group(1))
                    if last_ran == '_never_':
                        overdue_routines += 1
                    else:
                        try:
                            last_date = datetime.strptime(last_ran, '%Y-%m-%d')
                            days_since = (datetime.now() - last_date).days
                            if days_since >= freq_days:
                                overdue_routines += 1
                        except:
                            pass
except:
    pass

# ── Write state file ────────────────────────────────────────────────────────

state = {
    'timestamp': datetime.now().isoformat(),
    'context': {
        'est_tokens': est_tokens,
        'pct': round(context_pct, 1),
        'size_kb': int(file_size / 1024),
        'turns': user_turns,
        'tool_calls': tool_calls,
    },
    'health': {
        'uncommitted_files': uncommitted,
        'memory_stale_days': memory_stale_days,
        'overdue_routines': overdue_routines,
    },
}

try:
    with open(state_file, 'w') as sf:
        json.dump(state, sf)
except:
    pass

# ── Format status line ──────────────────────────────────────────────────────

line1 = f'~{est_tokens:,}tok ({context_pct:.0f}%) | {user_turns} turns | {tool_calls} tools | {file_size/1024:.0f}KB'

warnings = []
if uncommitted > 0:
    warnings.append(f'{uncommitted} uncommitted')
if memory_stale_days >= 2:
    warnings.append(f'memory stale {memory_stale_days}d')
if overdue_routines > 0:
    warnings.append(f'{overdue_routines} overdue routine{\"s\" if overdue_routines != 1 else \"\"}')

output = line1
if warnings:
    output += '\n  ! ' + ' | '.join(warnings)

print(json.dumps({'systemMessage': output}))
" "$TRANSCRIPT" "$STATE_FILE" "$WORKSPACE" 2>/dev/null

exit 0
