---
format: bonsai.run/1   # fields: bonsai check --schema bonsai.run; how to fill: skill base:run
id: R-2026-10-24-T-5736
task: T-5736
role: builder
model: opus
started: 2026-10-24 10:06
finished: null
outcome: running
commits: []
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
