#!/usr/bin/env bash
# Scope Guard — File Edits
# Blocks Tech Lead Agent from editing files outside station/
# Exit 2 = block the tool call.

input=$(cat)
file_path=$(echo "$input" | python3 -c "import sys,json; print(json.load(sys.stdin).get('tool_input',{}).get('file_path',''))" 2>/dev/null)

if [[ -z "$file_path" ]]; then
  exit 0
fi

ROOT="${1:-.}"

# Resolve to absolute path for comparison
abs_file=$(realpath -m "$file_path" 2>/dev/null || echo "$file_path")
abs_workspace=$(realpath -m "${ROOT}/station/" 2>/dev/null || echo "${ROOT}/station/")
abs_root=$(realpath -m "${ROOT}" 2>/dev/null || echo "${ROOT}")

# Allow edits anywhere within the repo root
if [[ "$abs_file" == "${abs_root}"* ]]; then
  # Block writes to .env files
  basename_file=$(basename "$file_path")
  if [[ "$basename_file" == .env* ]]; then
    echo "BLOCKED: Tech Lead Agent cannot modify .env files."
    exit 2
  fi
  exit 0
fi

# Block everything outside the repo
echo "BLOCKED: Tech Lead Agent cannot modify files outside the project root. File: $file_path"
exit 2
