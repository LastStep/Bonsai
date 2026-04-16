---
tags: [workflow, routines, digest]
description: Synthesize all pending routine reports into a consolidated digest — extract actionable items, present interactive Q&A for decisions, route results to plans or backlog.
---

# Workflow: Routine Digest

---

## When to Use

Manually, after routines have run and produced reports in `Reports/Pending/`. This is the "read all reports so the human doesn't have to" workflow.

---

## Procedure

### Phase 1 — Scan & Extract

1. **List reports.** Read every file in `Reports/Pending/` that has `routine:` in its frontmatter. These are routine reports. Ignore non-routine reports (plan completion reports, session transcripts, etc.).

2. **Extract from each report:**
   - **Flagged items** — the "Items Flagged for User Review" section. These are the primary actionable findings.
   - **Findings table** — the "Findings Summary" table. Note severity levels.
   - **Errors** — the "Errors & Warnings" section. Any errors are auto-promoted to high priority.
   - **Future notes** — the "Notes for Next Run" section. These contain forward-looking observations.

3. **Record which reports were processed** — you'll need this for the archive step.

### Phase 2 — Deduplicate & Categorize

4. **Deduplicate.** Cross-reference findings across reports. The same issue often surfaces in multiple routines (e.g., "upgrade Go toolchain" in both dependency-audit and vulnerability-scan). Merge duplicates into a single finding, noting which routines surfaced it.

5. **Categorize every unique finding into exactly one bucket:**

   | Bucket | Criteria | Examples |
   |--------|----------|---------|
   | **Action Required** | User must decide or do something. Has a concrete fix or decision. | Upgrade dependency, fix stale doc, resolve config gap |
   | **Warning** | Not broken yet, but will be if ignored. Time-sensitive or risk-bearing. | Security vuln with no exploit but growing exposure, routine that does nothing |
   | **Informational** | Good to know, no action needed. Confirms health or documents baseline. | "All links valid", "no secrets found", "first run — baseline established" |
   | **Future Possibility** | Ideas, suggestions, or improvements surfaced by routine analysis. Not urgent. | "Consider removing unused routine", "install semgrep for better coverage" |

6. **Within Action Required, assign rough effort:**
   - **Quick fix** — under 5 minutes, no plan needed (edit a file, update a number, add a gitignore entry)
   - **Small task** — under 30 minutes, backlog-worthy but no plan needed
   - **Plan-worthy** — needs a plan, possibly multi-step, possibly cross-cutting

### Phase 3 — Present to User

7. **Show the consolidated digest.** Format:

   ```
   ## Routine Digest — YYYY-MM-DD

   **Reports processed:** N (list routine names)
   **Findings:** N total — X action required, Y warnings, Z informational, W future

   ### Action Required (X items)

   For each item:
   > **[severity] Finding title**
   > Source: routine-name (+ routine-name if deduplicated)
   > Detail: one-line explanation
   > Effort: quick fix | small task | plan-worthy

   ### Warnings (Y items)
   ...

   ### Informational (Z items)
   ...

   ### Future Possibilities (W items)
   ...
   ```

### Phase 4 — Interactive Q&A

8. **For each Action Required item, ask the user what to do.** Use `AskUserQuestion` with these choices:

   - **Quick fixes:** Ask as a batch — "These N items are quick fixes. Do all now, pick which ones, or defer?"
     - "Do all now" — execute the fixes inline during this workflow
     - "Let me pick" — present each, user selects which to do now
     - "Add all to backlog" — batch-add to Backlog.md
     - "Skip" — acknowledge and move on

   - **Small tasks:** Ask individually or in related clusters — "What should we do with this?"
     - "Add to backlog as P1" 
     - "Add to backlog as P2"
     - "Skip — not relevant"

   - **Plan-worthy items:** Ask individually — "This looks like it needs a plan. How to handle?"
     - "Write a plan report" — will be included in the pending report output
     - "Add to backlog as P1"
     - "Add to backlog as P2"
     - "Skip — not relevant"

9. **For each Warning, ask:** "Acknowledge or act?"
   - "Acknowledge" — noted, no action
   - "Add to backlog" — capture for future action
   - "Act now" — treat as Action Required

10. **Future Possibilities are presented for awareness only.** Ask once at the end: "Any of these worth capturing in the backlog?" Let the user multi-select.

### Phase 5 — Execute Decisions

11. **Quick fixes the user approved:** Execute them now. Make the edits, confirm each one.

12. **Backlog items:** Add to `Playbook/Backlog.md` under the appropriate priority section. Use the standard item format:
    ```
    - **[category] Short description** — Context. *(added YYYY-MM-DD, source: routine-digest)*
    ```
    Choose the category (`bug`, `debt`, `security`, `improvement`, `feature`) based on the finding type. Group related items if they naturally cluster.

13. **Plan report:** If any items were marked "write a plan report", create a single consolidated report at `Reports/Pending/YYYY-MM-DD-routine-digest.md` containing:
    - All plan-worthy items with full context from the source reports
    - User decisions captured during Q&A
    - Suggested plan scope and grouping
    - Cross-references to source routine reports

    Use this format:
    ```markdown
    ---
    tags: [report, digest]
    from: tech-lead
    to: Tech Lead
    date: YYYY-MM-DD
    status: needs-plan
    ---

    # Routine Digest Report — YYYY-MM-DD

    ## Items Requiring Plans

    ### Item title
    - **Source:** routine-name(s)
    - **Severity:** medium/high
    - **Detail:** full context from source reports
    - **User decision:** what the user chose and any notes
    - **Suggested scope:** what a plan should cover

    ## Backlog Items Added
    - List of items added to backlog with priorities

    ## Quick Fixes Applied
    - List of fixes made during this digest

    ## Acknowledged Warnings
    - List of warnings the user acknowledged

    ## Source Reports
    - List of all reports processed with dates
    ```

    **Only write this report if there are plan-worthy items.** If everything was either quick-fixed or added to backlog, skip the report entirely.

### Phase 6 — Archive

14. **Move processed reports.** Move all routine reports that were processed from `Reports/Pending/` to `Reports/Archive/`. Create `Reports/Archive/` if it doesn't exist.

    Do NOT move non-routine reports (plan completion reports, session transcripts, etc.) — those have their own lifecycle.

15. **Update the routine log.** Append a single entry to `Logs/RoutineLog.md`:
    ```
    ### YYYY-MM-DD — Routine Digest
    - **Outcome:** success
    - **Reports processed:** N (list names)
    - **Quick fixes applied:** N
    - **Backlog items added:** N
    - **Plan report written:** yes/no
    - **Warnings acknowledged:** N
    ```

---

## Rules

- **Never auto-decide for the user.** Every actionable item gets a question. The whole point is informed human decision-making.
- **Deduplicate aggressively.** If two routines found the same thing, present it once with both sources cited.
- **Preserve context.** When adding to backlog, include enough detail that a future session can act on it without re-reading the source report.
- **Don't inflate.** If all reports are clean (no actionable items, no warnings), say so and skip the Q&A. A clean digest is a one-line message: "All N routine reports clean — no action needed. Archiving."
- **Effort estimates are guidance, not gates.** If the user wants to plan a "quick fix" or inline a "plan-worthy" item, follow their lead.
- **Cluster related backlog items.** If three findings all point to "upgrade Go toolchain", that's one backlog item, not three.
