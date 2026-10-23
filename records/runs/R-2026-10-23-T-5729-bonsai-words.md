---
format: bonsai.run/1   # fields: bonsai check --schema bonsai.run; how to fill: skill base:run
id: R-2026-10-23-T-5729
task: T-5729
role: builder
model: sonnet
started: 2026-10-23 07:39
finished: 2026-10-23 08:00
outcome: verify
commits: []
labels: {}
---

# Run: T-5729, Bonsai's own words (update lead line, plan-step names, ladder header, BUG-097)

## Log

- Audit M-2026-10-23-4667513b; go M-2026-10-23-7563af3b (option A for BUG-097; Owns widened to internal/engine,
  internal/format printed words, internal/guard/rules.go one string, internal/reference, docs/reference; formats/ untouched;
  contract/spec citations left for a bug after 1.0).
- Deviation first: none from the go. Two generated files outside the brief's list changed because the code generator
  writes them from the code: packs/base/skills/generated-files/SKILL.md and packs/base/skills/workspace/SKILL.md
  (the printed words only; `go generate` in internal/reference, with docs/reference/lists.md).
- (a) cmd/bonsai/engine.go: update's lead is "recorded the settings for the checks, looser than the ones accepted
  before" when it accepted looser settings; "nothing to change" otherwise. Test TestUpdateLeadsWithWhatItRecorded
  (cmd/bonsai/words_test.go) failed first on "nothing to change; demo-pack at ...", then passed.
- (c) internal/ladder: the header names the checkout relative to the folder the climb ran in (shownRoot); the result
  line was already relative. TestShownRoot failed first (undefined), then passed.
- BUG-097, option A: internal/leftovers Lines() prints one line, the count and up to 10 pids, no name or command;
  Record() (result file, --json) keeps both. TestLinesNameNoProgram failed first ("chrome", "user-data-dir" in the
  line). The two existing expectations in TestRecordAndLines pinned the old line and moved with the changed spec.
- (b) TestNoPlanStepNamesInPrintedWords parses every non-test .go file under cmd and internal and fails on a string
  literal holding "step N", "plan-5", "set N", "formats set" or a "5.x(.y)" number (contract/spec "section N"
  citations are skipped). It failed first on 45 strings, then passed after the strings were reworded. Some code
  comments with the same citations were edited by the same substitution; no logic changed.
- Left for a bug after 1.0: about 130 printed lines cite "contract section N", "spec section/§N" or "design/..."
  (engine/checkdocs.go, checkpack.go, format/check.go, error.go, ask.go, reference.go, cmd/bonsai/*).
- Evidence: full `go test ./...` and `go vet ./...` green before the climb. Climb at 97f93f8a (code identical to this
  branch's tip but this report): green, 6 rungs, rung 3 tests seen 4887, rung 4 seen 2179 (fault switch), 9 and 3 skipped
  as on main (Windows-only); the climb for the final HEAD is named in the report message.
- Model: sonnet (builder). Processes left running: none started by me remain (ps checked).
