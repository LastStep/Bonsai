#!/usr/bin/env bash
# Status Bar — Tech Lead Agent
# Warnings-only transcript line. Silent on zero warnings.
# Metrics live in the personal statusLine at ~/.claude/bonsai-statusline.sh.

INPUT=$(cat)

STOP_ACTIVE=$(echo "$INPUT" | python3 -c "import sys,json; print(json.load(sys.stdin).get('stop_hook_active', False))" 2>/dev/null)
if [[ "$STOP_ACTIVE" == "True" ]]; then
  exit 0
fi

ROOT="${1:-.}"
WORKSPACE="${ROOT}/station/"

ROOT="$ROOT" WORKSPACE="$WORKSPACE" python3 <<'PY'
import json, os, re, subprocess, sys, time
from datetime import datetime

root = os.environ.get('ROOT', '.')
workspace = os.environ.get('WORKSPACE', os.path.join(root, 'station/'))

warnings = []

# Uncommitted files
try:
    r = subprocess.run(['git', '-C', root, 'status', '--porcelain'],
                       capture_output=True, text=True, timeout=2)
    if r.returncode == 0:
        n = sum(1 for l in r.stdout.splitlines() if l.strip())
        if n > 0:
            warnings.append(f'{n} uncommitted')
except Exception:
    pass

# Memory staleness — days since last commit touching agent/Core/memory.md
memory_rel = 'agent/Core/memory.md'
memory_abs = os.path.join(workspace, memory_rel)
if os.path.isfile(memory_abs):
    try:
        r = subprocess.run(
            ['git', '-C', workspace, 'log', '-1', '--format=%ct', '--', memory_rel],
            capture_output=True, text=True, timeout=2,
        )
        if r.returncode == 0 and r.stdout.strip():
            last = int(r.stdout.strip())
            days = int((time.time() - last) / 86400)
            if days >= 2:
                warnings.append(f'memory stale {days}d')
    except Exception:
        pass

# Overdue routines — parse dashboard
dashboard = os.path.join(workspace, 'agent', 'Core', 'routines.md')
overdue = 0
if os.path.isfile(dashboard):
    try:
        in_table = False
        with open(dashboard) as df:
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
                if not dline.startswith('|'):
                    continue
                cols = [c.strip() for c in dline.split('|')]
                if len(cols) < 5:
                    continue
                freq_match = re.search(r'(\d+)', cols[2])
                if not freq_match:
                    continue
                fd = int(freq_match.group(1))
                last_ran = cols[3]
                if last_ran == '_never_':
                    overdue += 1
                else:
                    try:
                        lr = datetime.strptime(last_ran, '%Y-%m-%d')
                        if (datetime.now() - lr).days >= fd:
                            overdue += 1
                    except Exception:
                        pass
    except Exception:
        pass

if overdue:
    warnings.append(f'{overdue} overdue routine' + ('s' if overdue != 1 else ''))

if not warnings:
    sys.exit(0)

output = '  ! ' + ' · '.join(warnings)
print(json.dumps({'systemMessage': output}))
PY

exit 0
