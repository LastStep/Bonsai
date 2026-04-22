#!/usr/bin/env bash
# Bonsai personal statusLine renderer.
# Reads Claude Code statusLine stdin JSON, prints one-line bar.
# Env toggles: BONSAI_STATUSLINE_HIDE=cost,5h,7d · NO_COLOR=1

set +e
INPUT=$(cat)

if [ -n "$BONSAI_STATUSLINE_DEBUG" ]; then
  {
    echo "=== $(date -Iseconds) pwd=$PWD ==="
    echo "$INPUT"
  } >> /tmp/bonsai-statusline-debug.log 2>/dev/null
fi

CAVEMAN_FLAG="${CLAUDE_CONFIG_DIR:-$HOME/.claude}/.caveman-active"
CAVEMAN=""
if [ -f "$CAVEMAN_FLAG" ] && [ ! -L "$CAVEMAN_FLAG" ]; then
  MODE=$(head -c 64 "$CAVEMAN_FLAG" 2>/dev/null | tr -d '\n\r' | tr '[:upper:]' '[:lower:]' | tr -cd 'a-z0-9-')
  case "$MODE" in
    off|lite|full|ultra|wenyan-lite|wenyan|wenyan-full|wenyan-ultra|commit|review|compress)
      CAVEMAN="cave:$MODE"
      ;;
  esac
fi

STATUSLINE_JSON="$INPUT" CAVEMAN_BADGE="$CAVEMAN" python3 <<'PY'
import json, os, re, subprocess, sys

try:
    data = json.loads(os.environ.get('STATUSLINE_JSON', '') or '{}')
except Exception:
    data = {}

caveman = os.environ.get('CAVEMAN_BADGE', '')
hide = {x.strip() for x in os.environ.get('BONSAI_STATUSLINE_HIDE', '').split(',') if x.strip()}
no_color = bool(os.environ.get('NO_COLOR'))

def c(code, text):
    if no_color or not text:
        return text
    return f'\x1b[38;5;{code}m{text}\x1b[0m'

def dim(text):
    return c(244, text)

def tier(pct):
    if pct is None:
        return 250
    if pct < 50:
        return 108  # sage
    if pct < 70:
        return 180  # sand
    return 174      # rose

parts = []
if caveman:
    parts.append(caveman)

# Workspace / agent tag — walk up looking for .bonsai.yaml
cwd = (data.get('workspace') or {}).get('current_dir') or data.get('cwd') or os.getcwd()
tag = None
p = cwd
for _ in range(12):
    cfg = os.path.join(p, '.bonsai.yaml')
    if os.path.isfile(cfg):
        try:
            with open(cfg) as f:
                txt = f.read()
            m = re.search(r'^agents:\s*\n((?:\s+[^\n]*\n?)+)', txt, re.MULTILINE)
            if m:
                m2 = re.search(r'^\s{4}([a-z0-9-]+):', m.group(1), re.MULTILINE)
                if m2:
                    tag = m2.group(1)
        except Exception:
            pass
        break
    parent = os.path.dirname(p)
    if parent == p:
        break
    p = parent

if not tag:
    tag = os.path.basename(cwd) or '?'
parts.append(c(108, tag))

# Model
m = data.get('model') or {}
model_name = m.get('display_name') or m.get('id') or '?'
parts.append(c(109, model_name))

# Git branch + dirty count
try:
    br = subprocess.run(['git', '-C', cwd, 'symbolic-ref', '--short', 'HEAD'],
                       capture_output=True, text=True, timeout=2)
    if br.returncode == 0:
        branch = br.stdout.strip()
        st = subprocess.run(['git', '-C', cwd, 'status', '--porcelain'],
                           capture_output=True, text=True, timeout=2)
        dirty = 0
        if st.returncode == 0:
            dirty = sum(1 for l in st.stdout.splitlines() if l.strip())
        seg = c(250, branch)
        if dirty:
            seg += c(179, f'*{dirty}')
        parts.append(seg)
except Exception:
    pass

# Context %
ctx = (data.get('context_window') or {}).get('used_percentage')
if isinstance(ctx, (int, float)):
    parts.append(c(tier(ctx), f'ctx {ctx:.0f}%'))

# Rate limits (Pro/Max only — silent if absent)
rl = data.get('rate_limits') or {}
fh = (rl.get('five_hour') or {}).get('used_percentage')
if isinstance(fh, (int, float)) and '5h' not in hide:
    parts.append(c(tier(fh), f'5h {fh:.0f}%'))
sd = (rl.get('seven_day') or {}).get('used_percentage')
if isinstance(sd, (int, float)) and '7d' not in hide:
    parts.append(c(tier(sd), f'7d {sd:.0f}%'))

# Elapsed
cost = data.get('cost') or {}
dur_ms = cost.get('total_duration_ms')
if isinstance(dur_ms, (int, float)) and dur_ms > 0:
    s = int(dur_ms / 1000)
    h, rem = divmod(s, 3600)
    mn = rem // 60
    elapsed = f'{h}h{mn:02d}m' if h else f'{mn}m'
    parts.append(dim(elapsed))

# Cost (show if non-zero)
usd = cost.get('total_cost_usd')
if isinstance(usd, (int, float)) and usd > 0 and 'cost' not in hide:
    parts.append(dim(f'${usd:.2f}'))

sep = c(240, '  ·  ')
sys.stdout.write(sep.join(parts))
PY
