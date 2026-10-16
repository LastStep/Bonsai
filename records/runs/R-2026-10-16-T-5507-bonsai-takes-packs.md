---
format: bonsai.run/1   # fields: bonsai check --schema bonsai.run
id: R-2026-10-16-T-5507
task: T-5507
role: orchestrator
model: opus
started: 2026-10-16 05:51
finished:
outcome:
commits: []
labels: {}
---

# Run: T-5507, Bonsai takes base and workflow (5.5.7)

- Who: the orchestrator; a Sonnet runner (the pre-release on a scratch project); a Haiku counter (the log's calls per
  run); the director's sitting (install, `bonsai update --allow-exec --yes`, the yes to `workflow`); a Sonnet auditor
  (CLAUDE.md's lines); the end verifier
- Why: `design/plan-5.md` note 5.5.7 (from 5612)

> Opened before the first edit and appended to as the work goes.

## Log

- 04:57 **The Haiku counter** (04:56-04:57; `~/bonsai-checks/T-5507/count/runs.tsv`, summary
  `~/bonsai-checks/T-5507/limits-notes.md`): 24,232 records, 13 Oct 07:13Z to 15 Oct 23:27Z; 96 subagent runs with calls,
  all typed `general-purpose`; most calls in one run 321 in 230 minutes, longest run 416 minutes (262 calls), fastest 177
  calls in 33 minutes (5.4 a minute), median 1.8 a minute. 2,167 subagent ids carry only a stop record (for the
  recorder's look before 1.0).
- 05:51 **`T-5505` and `T-5506` set `blocked`** (each waits on the director: the Windows S1 step; the walls' Windows
  half), so `T-5507`, whose `bonsai.allows` is not empty, runs alone. `T-5507` `todo` to `running`.
- 05:51 **The new pre-release**, built after every code fix landed (`T-5520` at `bf7cc3d1`, the last): the installed
  one first copied to `~/bonsai-checks/prerelease/previous/bonsai` (sha256
  `356f787c3d55234920a20a437ef1e667e65e032b93d682994fa202a2afeef644`, `bonsai dev (commit a88436a8ff4f)`), its install
  lines in `previous/INSTALL.md`. A clean clone of `main` at `bf7cc3d124f1e1d7304a302c01e587a52c5dc0ca` in
  `~/bonsai-checks/prerelease/src-bf7cc3d`, `CGO_ENABLED=0 go build -trimpath -buildvcs=true`: `go version -m` shows
  go1.27.2, `CGO_ENABLED=0`, `vcs.revision=bf7cc3d124f1e1d7304a302c01e587a52c5dc0ca`, `vcs.modified=false`, no tag.
  **Its fingerprint: `70516eb8dd169515975b0eaf99fca023882a6fb8959c6d1a2a2a807e1c3a2478`**
  (`~/bonsai-checks/prerelease/bonsai`). A Sonnet runner tries it on a scratch project next.
- 05:55 **The pre-release tried on a scratch project** (Sonnet, 05:52-05:55; `~/bonsai-checks/T-5507/try/`): sha256 matched; `bonsai dev (commit bf7cc3d124f1)`. Linked to both packs (`init --yes`, both in `bonsai.yaml` first; a stand-in `claude` on the PATH); `check` no finding (two `plugin-trust` warnings: no real session); `status` shows the `bonsai` and `workflow` label namespaces, the three lanes and the six roles' limits (builder 180 min / 600 actions / 4 red runs, verifier 120/400, researcher and auditor 30/150, runner 150/250, the orchestrator none); a climb green (guard and `go vet` rungs), its log record naming the build and its sha256; a message round trip (brief, `--show` with the trust line, report, the thread `answered`); the guard on payloads: an auditor's Agent call refused `call-refused`, a verifier's Edit refused `role-writes-nothing`. As designed, and read so: adding `workflow` by `update` in an agent's session refused (it widens Bonsai's guard line: the person's `--allow-exec --yes`); a rung added after the link refused `config-looser` until a person's `update` (K1); a local pack source by absolute path is `absolute-path` (sources are URLs; the runner mapped two URLs to local clones with a scratch git config); payloads with no backing Agent call held as helpers.
- 05:56 **CLAUDE.md's audit** (Sonnet, 05:52-05:55; `~/bonsai-checks/T-5507/claude-md-audit.md`, 96 rows): 70 kept, 21 now in a pack, 5 dropped; 13 rows where the pack says less than `CLAUDE.md` (the orchestrator's place, a brief's contents, `cut` moved by the orchestrator, the labels' use here, every plan reviewed, the verifier list, the model lines, the run rows), which stay in the shrunk file. The shrink itself waits for the link.
- 09:28 **The director installed the pre-release** (their word: "done"). Read on this side: `/usr/local/bin/bonsai` sha256 `70516eb8dd169515975b0eaf99fca023882a6fb8959c6d1a2a2a807e1c3a2478`, `-rwxr-xr-x root root`, installed 09:28; `which -a bonsai` lists it alone; `bonsai dev (commit bf7cc3d124f1)`.
- 09:30 **The director's yes to `workflow`** (asked with the counts: add with builder minutes 240 by default and 480 at most, recommended; the pack's limits; 360; or not yet): **"option 1"**, and asked whether limits can be set by the command line: answered that they are lines in `bonsai.yaml` under the pack's entry (no `set` command; `check` and `status` after an edit; a budget change is not a loosening); a command to set them noted for 5.6-5.7.
- 09:33 **The director on the Windows login check** (asked: a script that prints only the size, recommended; they check it themselves; or skip it): **"Skip the check"**. From now no agent sizes or touches the Windows credentials file; the warm-up session stays; a session that reports no login stops the run. Written in tonight's shared brief rules (`~/bonsai-checks/briefs/common-2026-10-16.md`).

## Runs

| Run | Model | Start | End | Minutes |
|---|---|---|---|---|
| Counter: calls per run | Haiku | 04:56 | 04:57 | 1 |
| Runner: the pre-release on a scratch project | Sonnet | 05:52 | 05:55 | 3 |
| Auditor: CLAUDE.md's lines | Sonnet | 05:52 | 05:55 | 3 |
