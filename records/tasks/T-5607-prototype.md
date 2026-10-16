---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task
id: T-5607
title: "workflow:prototype and its first run, the CLI look in five styles (5.6.7)"
status: done
lane:
done_when:
  - "workflow:prototype in LastStep/bonsai-workflow (when a question is a run's to answer; a builder writes the spike, a runner may run it; the spike in ~/bonsai-checks/<task>/spike/, never the repo; the variants behind one switch; the output is evidence for a decision, never code that lands; the result in the options or the artifact), with the observe-don't-ask paragraph in workflow:orchestrate and workflow:lanes and the README's skill list; workflow's CI green"
  - "The CLI spike, briefed from the skill's steps: a throwaway Go module in ~/bonsai-checks/T-5607/spike/ rendering the real status --json and init --json preview in five styles (A standard library, ASCII, no colour; B A with colour; C B with golang.org/x/term's width; D lipgloss; E B with Unicode marks), with init --explain's detail view"
  - "Each style's module count (go list -m all), binary growth and start time (spec 3's bound: under 5 ms), from the built binaries"
  - "The outputs shown to the director as images in the artifact, every image from a real run of the spike on real data; the director's answers recorded"
depends_on: [T-5690]
blocked_by:
created: 2026-10-16
started: 2026-10-16
finished: 2026-10-16
labels:
  bonsai.ladder: []
  bonsai.allows: []
  bonsai.wants: []
---

# workflow:prototype and its first run, the CLI look in five styles (5.6.7)

Piece 5.6.7 of step 5.6: its row in `design/plan-5.md`'s "Added 16 Oct, before the build" (the director's decisions,
16 Oct; `T-5690`).

- **Owns:** in `LastStep/bonsai-workflow` (worked in `~/bonsai-checks/bonsai-workflow`), `skills/prototype/SKILL.md`, the
  paragraph in `skills/orchestrate/` and `skills/lanes/`, the README's skill list; the scratch folder
  `~/bonsai-checks/T-5607/spike/`; the artifact. Nothing in Bonsai's repo.
- **Grants:** none in Bonsai's repo, so `bonsai.allows` is empty and it has no branch here.
- **Proof:** the lines of `done_when`; `workflow`'s CI; the ref move to it is `T-5691`'s, on the director's yes.
