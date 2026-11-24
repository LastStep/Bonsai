---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task; how to fill: skill base:task
id: T-6023
title: "Bugs as Bonsai's own documents: one file per bug, a generated register, the bonsai bug command (the studio's ask for bug rows as JSON, 19 Nov)"
status: done
lane: full
done_when:
  - "The director has chosen in records/options/OPT-0003-bug-rows.md (done 19 Nov: a bugs system, BUG- ids only, no import)"
  - "A plan (workflow:plan) from that choice, reviewed by a fresh Opus agent and approved by the director, before any code"
  - "A command gives each row of a project's bugs register as JSON with id, state, title, area, where, found_by, fixed_in and the register's path; proved on a scratch project holding both drifting shapes the studio names, moved by base's bug skill"
depends_on: []
blocked_by: null
created: 2026-11-19
started: 2026-11-19
finished: 2026-11-24
labels:
  bonsai.ladder: []
  bonsai.allows: []
  workflow.model: opus
---

# The bugs register's rows as JSON

The studio, 19 Nov (pasted by the director): `status --json` gives the register only as a document entry
(`{"kind":"bugs","from":"workflow","file":"<path>"}`), no command gives its rows, so a consumer that wants "bug X
filed" or "bug X fixed" parses the markdown itself. It asks for, per row: id, state (open or fixed), title, area (the
`###` heading it sits under), where, found by, fixed in, and the register's path. Nothing else; no decisions table.

Director lane: `design/contract.md` 7.3 says "Rows the studio parses out of a pack document's body (the bugs table)
are not contract", and the formats are final (7 Oct). Making the rows contract reverses that, and adds a field to the
pack format and a command. The options: `records/options/OPT-0003-bug-rows.md`.

The director's choice, 19 Nov (OPT-0003): "we should build this system properly. like having proper commands for
working with bugs"; "we will go with bug prefix only, as we want to keep things uniform"; "i dont think you need to do
any kind of importing for the project, but rather i will tell each project agent myself to do the migration". The lane
moved from director to full on that choice. The plan: `records/plans/P-T-6023.md`.

Approved by the director 19 Nov ("Approve, start building"). Built as T-6024 to T-6028 (the plan's table).
