---
format: bonsai.run/1   # fields: bonsai check --schema bonsai.run; how to fill: skill base:run
id: R-2026-11-06-T-6022
task: T-6022
role: orchestrator
model: opus
started: 2026-11-06 12:34
finished: 2026-11-06 14:58
outcome: merged
commits: []
labels: {}
---

# Run: the scratch folders' audit and clean-up

> Times UTC (`bonsai now`). The director present.

## Log

- **12:34** The director: "do an audit of bonsai checks folder and clean up all the old and unnessary stuff. first
  show me a table with overview of how things are there", then "use a subagent to do it". T-6021 moved to `verify`
  (its CI on `85aae391` running; its grant closed), T-6022 `running`. A first look by the orchestrator: WSL's
  `~/bonsai-checks` 8.4 GB, 100 folders, 26 loose files; the biggest T-5721 1.2 GB, T-6001 1.0 GB, T-5741 0.66 GB.
- **12:36** Brief `M-2026-11-06-f20f2759` to auditor@T-6022 (**sonnet**, read only): every top-level entry of both scratch folders, size, last change, what made it, still used by what, keep/remove/ask. Not started: the session holds no workflow role (the plugins at the new marketplace installed at its start); the director asked to reload.
- **12:37** The director ran `/reload-plugins` (13 agents) and said "go": auditor@T-6022 started on its brief.
- **12:44** auditor@T-6022 (sonnet) report `M-2026-11-06-7e101c4f`, done, read only; its tables in the session's
  scratchpad (`audit-tables.md`: 126 WSL rows, 55 Windows rows, 68 Windows git states). WSL 8.4 GB: keep 25 (782 MB),
  ask 7 (1251 MB), remove 94 (6628 MB); Windows 2.9 GB: keep 3 (85 MB), ask 1 (32 MB), remove 51 (2801 MB). No unpushed
  work; `bonsai-workflow` clean and in sync; a decoy-named `.credentials.json` in T-5506 not opened; eleven remove
  rows cited by finished run reports; three STATE paths gone or elsewhere. The orchestrator's calls on top: the done
  tasks' verifier copies (T-6016-verify, T-6020-verify) to remove; T-6001 and T-5791 remove; bin keeps claude-here
  only; home, T-5797, T-5795 and the morning note kept; Windows T-6007-win remove. Shown to the director, with the
  choice of keeping the eleven cited folders.
- **14:53** The director: "yes as recommended". Nothing running under either scratch folder (`ps`). Removed in WSL,
  by written-out paths: every remove row, plus T-6016-verify, T-6020-verify, T-6001, T-5791,
  `prerelease/src-a88436a` and `src-bf7cc3d`, and `bin` but `claude-here`. The guard refused `bin/bonsai`
  (delete-protected: a file named as Bonsai's program); it stays for a person. WSL 8.4 GB to 420 MB, 28 entries.
- **14:55** Windows: a first try with relative names after `cd` was refused (the guard read `packs` as the repo's
  `packs/base/bonsai/pack.yaml`; it does not follow the `cd`); done again by full paths: every remove row, T-6007-win,
  T-6020-verify, and `bin` but `bonsai.exe` (left for a person, as in WSL). 2.9 GB to 66 MB: `scripts`, `T-6020`, `bin`.
- **14:57** STATE: the housekeeping line names what stays; the end verifier's gone report path and T-5791's line
  rewritten; the Now lines updated. T-6022 to `done`.

## Processes left running

None: the removals were foreground `rm` runs; checked with `ps -eo pid,etime,cmd`.

## What is left

The director's, if they want: `rm ~/bonsai-checks/bin/bonsai` in WSL and
`Remove-Item $env:USERPROFILE\bonsai-checks\bin\bonsai.exe` on Windows (old builds the guard keeps from agents).
