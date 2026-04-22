#!/usr/bin/env bash
# Context Guard — Tech Lead Agent
# Injects behavioral constraints based on context usage and detects session wrap-up triggers.
# Self-sufficient: reads transcript from stdin, computes ctx% internally.

INPUT=$(cat)
ROOT="${1:-.}"

STATUSLINE_JSON="$INPUT" ROOT="$ROOT" python3 <<'PY'
import json, os, re, sys

try:
    data = json.loads(os.environ.get('STATUSLINE_JSON', '') or '{}')
except Exception:
    sys.exit(0)

session_id = data.get('session_id', '')
if not session_id:
    sys.exit(0)

transcript = data.get('transcript_path', '')
prompt = data.get('prompt', '') or ''
root = os.environ.get('ROOT', '.')
docs_path = os.path.join(root, 'station/')

# ── Compute context % from transcript ───────────────────────────────────────
context_pct = 0.0
total_chars = 0
if transcript and os.path.isfile(transcript):
    try:
        with open(transcript) as f:
            for line in f:
                line = line.strip()
                if not line:
                    continue
                try:
                    entry = json.loads(line)
                except Exception:
                    continue
                etype = entry.get('type', '')
                msg = entry.get('message', {}) or {}
                content = msg.get('content', '')
                if isinstance(content, str):
                    total_chars += len(content)
                elif isinstance(content, list):
                    for block in content:
                        if not isinstance(block, dict):
                            continue
                        bt = block.get('type', '')
                        if bt == 'text':
                            total_chars += len(block.get('text', ''))
                        elif bt == 'tool_use':
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
                if etype == 'system':
                    c = entry.get('content', '')
                    if isinstance(c, str):
                        total_chars += len(c)
        est_tokens = int(total_chars / 3.5)
        context_pct = min((est_tokens / 200000) * 100, 100)
    except Exception:
        pass

# ── Tiered context injection ────────────────────────────────────────────────
injection = ''
if context_pct >= 85:
    injection = (
        f'CONTEXT CRITICAL [{context_pct:.0f}%]: STOP accepting new work. '
        'Only respond if the user is committing, saving, or wrapping up. '
        'For any other request, respond: "Context is nearly full. '
        'Please run /clear to start a fresh session." '
        'Do not use any file-reading tools.'
    )
elif context_pct >= 70:
    injection = (
        f'CONTEXT ALERT [{context_pct:.0f}%]: Complete ONLY the current task. '
        'Do NOT accept new work. When the current task completes, tell the user: '
        '"Context is high. I recommend running /clear before starting anything new." '
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

# ── Trigger detection on user prompt ────────────────────────────────────────
normalized = re.sub(r"'", '', prompt).lower()
eol = r'[\s.!?]*$'

trigger_patterns = [
    r'\b(thats|that is)\s+all' + eol,
    r'\b(were|we are)\s+done' + eol,
    r'\b(im|i am)\s+done' + eol,
    r'\b(lets|let us)\s+(wrap|wrap up)' + eol,
    r'\bsession\s+(done|over)' + eol,
    r'\bend\s+session' + eol,
    r'\b(wrap|wrapping)\s+up' + eol,
    r'\bfinish\s+up' + eol,
    r'\bcall\s+it' + eol,
]
triggered = any(re.search(p, normalized) for p in trigger_patterns)
if triggered:
    checklist = (
        '\nSESSION WRAP-UP REQUESTED. Complete this checklist before ending:\n'
        '1. Run git status - if uncommitted changes exist, ask user about committing\n'
        '2. Update agent/Core/memory.md with current work state\n'
        f'3. Review {docs_path}Playbook/Backlog.md - add any items discovered this session\n'
        f'4. Update {docs_path}Playbook/Status.md if any task status changed\n'
        f'5. Write session notes to {docs_path}Logs/ if significant work was done'
    )
    injection = (injection + '\n' + checklist) if injection else checklist

verify_patterns = [
    r'\bverify\s+(everything|it\s+all)\b',
    r'\bcheck\s+(your|the)\s+work\b',
    r'\bcheck\s+if\s+you\s+missed\b',
    r'\breview\s+(your|the)\s+changes\b',
    r'\breview\s+before\s+(commit|push|ship)\b',
    r'\bdoes\s+everything\s+look\s+(right|good)\b',
]
if any(re.search(p, normalized) for p in verify_patterns):
    checklist = (
        '\nVERIFICATION REQUESTED. Before proceeding:\n'
        '1. Re-read your own changes — check for bugs, edge cases, regressions\n'
        '2. Verify all tests pass (if applicable)\n'
        '3. Check for stale references in documentation\n'
        '4. Confirm no security issues introduced'
    )
    injection = (injection + '\n' + checklist) if injection else checklist

plan_patterns = [
    r'\b(lets|let\s+us)\s+plan\b',
    r'\bplan\s+(this|the|a)\b',
    r'\bcreate\s+a\s+plan\b',
    r'\bdesign\s+(this|the|a)\b',
    r'\barchitect\s+(this|the|a)\b',
]
if not triggered and any(re.search(p, normalized) for p in plan_patterns):
    reminder = (
        '\nPLANNING DETECTED. Before drafting a plan:\n'
        f'1. Load planning workflow: {os.path.join(root, "")}agent/Workflows/planning.md\n'
        f'2. Load planning-template skill: {os.path.join(root, "")}agent/Skills/planning-template.md\n'
        '3. Follow the Tier rules and Verification requirements'
    )
    injection = (injection + '\n' + reminder) if injection else reminder

if not injection:
    sys.exit(0)

output = {
    'hookSpecificOutput': {
        'hookEventName': 'UserPromptSubmit',
        'additionalContext': injection,
    }
}
print(json.dumps(output))
PY

exit 0
