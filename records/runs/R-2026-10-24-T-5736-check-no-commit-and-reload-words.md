---
format: bonsai.run/1   # fields: bonsai check --schema bonsai.run; how to fill: skill base:run
id: R-2026-10-24-T-5736
task: T-5736
role: builder
model: opus
started: 2026-10-24 10:06
finished: 2026-10-24 10:50
outcome: verify
commits: ["2484fbc6", "54724c86", "fdd9caa3", "d6cd1749", "eb070060"]
labels: {}
---

# Run: bonsai check warns on a project with no commit or bonsai.yaml uncommitted; init and update say a running session reloads

> Opened at the start of the session and added to as it goes. Not written at the end.

Brief `M-2026-10-24-3898b760` (orchestrator), builder on Opus 5.5 (`claude-opus-5-5`), worktree
`~/Servers/Bonsai-T-5736`, branch `T-5736` from `main` `41c797bf`.

## Log

- 10:06 Read the brief, the task, BUG-106/107. Worktree made.
- 10:10 Read the code. The shared HEAD reading is `stopgate.CommittedConfig` (the ladder's and the gate's). `engine`
  cannot import `stopgate` (stopgate imports engine), so check gets it the way it gets Claude Code's answers: cmd/bonsai
  runs `engine.CompareCommitted(r, read)` after `CheckWith`, `read` wrapping `stopgate.CommittedConfig`. To tell "no
  commit / never committed" (commit what init wrote) from "differs" (commit or revert), `stopgate.UncommittedError`
  gains a field saying which (the only change in stopgate). rung0 and status never get the warning (rung0 judges a
  commit; status lists findings only).
- 10:10 BUG-107: Claude Code 2.1.308 to 2.1.310 (the versions on this machine) carry `/reload-plugins` ("reload-plugins
  (or relaunch) to load what qualifies", read from the installed binary's strings), so the words name it and a fresh
  start. They go in each plugin result's `message` (bonsai.changes/1's `plugins[].message`; no format change): on an
  install when the plan's marketplace name moved (a first link, a pack added or changed, a rename), and on the waiting
  results and check's plugin-trust warning ("once it is installed").
- 10:20 Tests first, shown failing on the old code (only `ReloadWords` and `PluginConsent.Fresh` declared, unwired, so
  they compile): `TestPluginReloadWords` failed ("a first link's consent is not fresh"); `TestCheckUncommitted` failed
  at all four cases (no commit, check's text, an edit, not in HEAD: no warning, `"" "" ""`); `TestReloadWordsInOutput`
  failed (init --yes and update --yes --json carried no words).
- 10:30 Built. Four existing cmd tests asserted a clean `check` on a fresh link never committed (TestUpdateCommand,
  TestPluginStep, TestUpdateTakesAPackOut, TestChangesFirstLink): that project now warns by design. Each now commits
  the project just before that check (`c.commit`), its assertion unchanged. TestCheckTable gained the word's case
  (`c.link()`), as the table requires of every word.
- 10:40 Docs: the reference regenerated (`go generate ./...`: lists.md's new row; check's help names the reading),
  features.md's check row (49 words), README step 3 (commit what init wrote before a climb) and step 4's note (the
  reload), operating-bonsai's warning lines. Commits `2484fbc6`, `54724c86`, `fdd9caa3`, `d6cd1749`, rebased onto main
  `251abb28` on the orchestrator's follow-up `M-2026-10-24-d60a9ae0`.
- 10:45 Climb 1 at `eb070060` (waited for T-5737's climb holding the lock): green, sha256 `00c48b26...8c77`.
- 10:50 This report finished and committed; climb 2 at the new HEAD (below).

## What is different now

`bonsai check` on a freshly linked project that has not been committed now warns `config-uncommitted`, saying a climb
and the stop gate refuse until a person commits what `bonsai init` wrote; on a project whose `bonsai.yaml` was edited
and not committed it warns the same, naming the commit-or-revert step. It stays a warning (the exit code is the
findings' alone). When `init` or `update` installs a pack's plugin under a new marketplace name (a first link, a pack
added or moved, a rename), each installed plugin's line says that a Claude Code session already open in the checkout
loads the pack's skills only after `/reload-plugins` in it or a fresh start; a waiting plugin and check's
`plugin-trust` warning say it for when it is installed. An update that changes nothing says nothing of it.

## Deviations from the plan

- Four existing cmd tests asserted a clean `check` on a fresh link that was never committed; each now commits the
  project right before that check, the assertion itself unchanged (TestUpdateCommand, TestPluginStep,
  TestUpdateTakesAPackOut, TestChangesFirstLink). The orchestrator should read that diff (`fdd9caa3`).
- The reading is shared by handing `stopgate.CommittedConfig`'s result to engine from cmd/bonsai (engine cannot import
  stopgate: stopgate imports engine). The only change in stopgate is a `Never` field on `UncommittedError`.
- The warning is computed in `bonsai check` only, not in rung 0 (which judges a commit) nor status (findings only).

## Ladder

Climb 1, `/usr/local/bin/bonsai ladder --task T-5736 --json`, at `eb070060c3d150318fb1e283361719dc36798ed2`, mode
local, dirty false, green true, result sha256 `00c48b269d5ef43dd0e3f6b14b07022f5e44f5ed6060245be2e39d97abcd8c77`:

| Rung | Check | Result | Evidence |
|---|---|---|---|
| 0 | guard | green | 16 checked, 4.8 s |
| 1 | vet | green | 1.3 s |
| 2 | schemas | green | 0.4 s |
| 3 | tests | green | 5074 seen, 9 skipped (OS-only), go_tests 4756 -> 5065, 4 new tests, 61.9 s |
| 4 | tests with the fault switch | green | 2342 seen, 3 skipped, go_tests_tagged 2120 -> 2339, 3 new, 58.4 s |
| 5 | git integrity | green | 5 test files changed, nothing hidden |

The final climb at the report's commit is in the report message to the orchestrator.

## What did not work

- My first wording of check's help text ran two lines past the help width (TestWordTables, TestCheckSchema red twice);
  rewrapped. One line of that paragraph is now short (it ends at "Each"), ragged but within width.
- An extra `claude plugin list` before each install would have told "newly installed" exactly, but changes the call
  sequence existing tests pin; the plan's marketplace name moving is used instead. So an install that completes a
  plugin earlier left `waiting` under the same name prints no reload words (its waiting line already said them).
- Not run: Windows (`go test` natively); CI on the branch is the orchestrator's push.

## Evidence for a human

In a scratch folder: `git init`, `bonsai init ... --yes`, then `bonsai check`: a `Warnings` block with
`config-uncommitted`; commit, and it goes.

## Files changed

- internal/engine/plugins.go, plan.go: `ReloadWords`, `PluginConsent.Fresh`, the words on installed and waiting results
  and on plugin-trust.
- internal/stopgate/config.go: `UncommittedError.Never`.
- internal/engine/checkcommitted.go, internal/format/check.go, cmd/bonsai/check.go: the warning, its word, check's help.
- Tests: internal/engine/reload_test.go, cmd/bonsai/uncommitted_test.go (new); checktable_test.go (the case);
  engine_test.go, outputs_test.go (a commit before a clean check).
- docs/reference/lists.md (regenerated), docs/features.md, README.md, packs/base/skills/operating-bonsai/SKILL.md.

## Processes left running

None: every build and test ran in the foreground; checked with `ps -eo pid,etime,cmd` (no go, bonsai or test process of
this session).

## What it cost

About 45 minutes wall-clock; tokens not known.

## Left for next time

- BUG-107's cleanup (the old marketplace left registered) is after 1.0, as the register says.
