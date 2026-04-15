#!/usr/bin/env bash
# Post-Agent Review — Tech Lead Agent
# Outputs review checklist after a dispatched agent completes.
# NOTE: PostToolUse output may not be visible to the agent in current Claude Code versions.

echo "=== AGENT COMPLETED ==="
echo ""
echo "Before proceeding, complete the review workflow:"
echo ""
echo "1. REVIEW output against the plan — every step followed, nothing improvised"
echo "2. CHECK security — verify against project security standards"
echo "3. VERIFY — confirm verification steps from the plan passed"

echo "4. LOG — write results to station/Logs/ if significant work was done"
echo "5. STATUS — update station/Playbook/Status.md if the task status changed"
echo "6. REPORT — process any pending reports in station/Reports/Pending/"

