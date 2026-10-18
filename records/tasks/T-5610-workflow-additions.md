---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task
id: T-5610
title: "workflow's additions: workflow:correct, workflow:changelog and four short lines (5.6.10)"
status: done
lane:
done_when:
  - "workflow:correct in workflow (a mistake seen twice fixed at the highest level that works; a new check proved to fail on the real past mistake); its first run is at 5.6's end"
  - "workflow:changelog, a skill only: a range by commits, tags or dates; four forms (text in the Keep a Changelog layout, visual with before/after images, short, plain words for the person); sources in order (tasks done in the range, their commits, run reports and decisions, ladder results); every line traced to a task or commit in the range; no person's name, home path or secret, images checked; the range written in the file; written where the person names, else CHANGELOG.md, continuing one that exists"
  - "The four short lines: a how-sure grade on claims (workflow:reporting), a missing result is a gap not a pass (workflow:lenses), a parallel-work checkpoint (workflow:plan), restate the goal first (workflow:orchestrate); the README's lists"
  - "workflow's CI green; workflow:changelog tried by a builder in scratch on 5.5's range (text form) and on T-5607's spike images (visual form), read by the orchestrator; workflow:correct tried on 5.5's fix rounds, its proposals read"
depends_on: [T-5607]
blocked_by:
created: 2026-10-16
started: 2026-10-16
finished: 2026-10-16
labels:
  bonsai.ladder: []
  bonsai.allows: []
  bonsai.wants: []
---

# workflow's additions (5.6.10)

Piece 5.6.10 of step 5.6: its row in `design/plan-5.md`'s "Added 16 Oct, before the build".

- **Owns:** in `LastStep/bonsai-workflow` (worked in `~/bonsai-checks/bonsai-workflow`), `skills/correct/`,
  `skills/changelog/`, `skills/reporting/`, `skills/lenses/`, `skills/plan/`, `skills/orchestrate/`, the README; scratch
  under `~/bonsai-checks/T-5610/`. Nothing in Bonsai's repo.
- **Grants:** none here. **Proof:** the lines of `done_when`. The visual try uses T-5607's spike images, since 5.6.8's
  before/after captures were skipped on the director's word.
