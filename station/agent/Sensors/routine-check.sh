#!/usr/bin/env bash
# Routine Check — Tech Lead Agent
# Parses the routine dashboard and flags overdue routines at session start.

ROOT="${1:-.}"
WORKSPACE="${ROOT}/station/"
DASHBOARD="${WORKSPACE}agent/Core/routines.md"

if [[ ! -f "$DASHBOARD" ]]; then
  exit 0
fi

# Extract dashboard table rows between markers
in_table=false
overdue_count=0
overdue_lines=""

while IFS= read -r line; do
  if [[ "$line" == *"ROUTINE_DASHBOARD_START"* ]]; then
    in_table=true
    continue
  fi
  if [[ "$line" == *"ROUTINE_DASHBOARD_END"* ]]; then
    break
  fi
  if [[ "$in_table" != true ]]; then
    continue
  fi

  # Skip header and separator rows
  if [[ "$line" == "|"*"Routine"* ]] || [[ "$line" == "|"*"---"* ]]; then
    continue
  fi

  # Parse table row: | Routine | Frequency | Last Ran | Next Due | Status |
  if [[ "$line" == "|"* ]]; then
    routine=$(echo "$line" | awk -F'|' '{gsub(/^[ \t]+|[ \t]+$/, "", $2); print $2}')
    frequency=$(echo "$line" | awk -F'|' '{gsub(/^[ \t]+|[ \t]+$/, "", $3); print $3}')
    last_ran=$(echo "$line" | awk -F'|' '{gsub(/^[ \t]+|[ \t]+$/, "", $4); print $4}')

    if [[ -z "$routine" ]]; then
      continue
    fi

    # Extract frequency in days
    freq_days=$(echo "$frequency" | grep -oE '[0-9]+')
    if [[ -z "$freq_days" ]]; then
      continue
    fi

    # Check if overdue
    if [[ "$last_ran" == "_never_" ]]; then
      overdue_count=$((overdue_count + 1))
      overdue_lines="${overdue_lines}  WARNING: OVERDUE ROUTINE: ${routine} (never run, due every ${freq_days} days)\n"
    else
      # Calculate days since last run
      last_epoch=$(date -d "$last_ran" +%s 2>/dev/null)
      if [[ -z "$last_epoch" ]]; then
        # macOS date fallback
        last_epoch=$(date -j -f "%Y-%m-%d" "$last_ran" +%s 2>/dev/null)
      fi
      if [[ -n "$last_epoch" ]]; then
        now_epoch=$(date +%s)
        days_since=$(( (now_epoch - last_epoch) / 86400 ))
        if [[ $days_since -ge $freq_days ]]; then
          overdue_count=$((overdue_count + 1))
          overdue_lines="${overdue_lines}  WARNING: OVERDUE ROUTINE: ${routine} (last ran ${last_ran}, ${days_since} days ago, due every ${freq_days} days)\n"
        fi
      fi
    fi
  fi
done < "$DASHBOARD"

if [[ $overdue_count -gt 0 ]]; then
  echo ""
  echo "=== ROUTINE CHECK ==="
  echo -e "$overdue_lines"
  echo "  ${overdue_count} routine(s) overdue. Read agent/Core/routines.md for procedures. Ask user before running."
fi
