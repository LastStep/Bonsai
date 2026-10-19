---
format: bonsai.run/1   # fields: bonsai check --schema bonsai.run
id: R-2026-10-19-T-5613
task: T-5613
role: orchestrator
model: opus
started: 2026-10-19 10:07
finished: 2026-10-19 11:34
outcome: merged
commits: ["959c6a70", "7a05c12d"]
labels: {}
---

# Run: T-5613, Step 7a-7d of 5.6's end Windows run, in CI

- Who: the orchestrator, with the builder it briefs
- Why: the director, 19 Oct: 5.6 finishes this session; Windows 11 Home has no Sandbox, so step 7a-7d goes to CI and 7e to 5.7

> Opened before the first edit and appended to as the work goes.

## Runs

| Run | Model | Start | End | Minutes |
|---|---|---|---|---|
| Builder M-2026-10-19-708a71fa (audit) | opus | 2026-10-19 10:08 | 2026-10-19 10:12 | 4 |
| Builder, go M-2026-10-19-a569f0de | opus | 2026-10-19 10:13 | 2026-10-19 10:23 | 10 |
| Verifier M-2026-10-19-940c97af | opus | 2026-10-19 10:24 | 2026-10-19 10:32 | 8 |
| Builder, fix B1 M-2026-10-19-a1352b01 | opus | 2026-10-19 10:32 | 2026-10-19 10:36 | 4 |

## Log

- 10:07 T-5613 written; main at `d0340731`, CI green.
- 10:08 Worktree `~/Servers/Bonsai-T-5613` from `d0340731`. Builder (opus: CI and the elevated installer) briefed `M-2026-10-19-708a71fa`, audit first. T-5613 `running`, the one granting task.
- 10:12 Audit `M-2026-10-19-5a136b4e` (opus 10:08-10:12): the job's Windows step holds no negative
  control; its %...% check is likely empty on the runner (no such entry in its Path). 7c from the code: the child exits
  18 (install.ps1:180, :188), the parent's Stop-Install exits 1 (:302-303, :70): both reports right at different layers.
  7d's folder refusal exits 1 (:255-256). Plan: a new Windows step, each check run against install.ps1 and a broken copy
  (one exact-once replace each). Risk: first run on main (no branch pushed), fixed forward if red. Go sent: a new step;
  CreateEnvironmentBlock plus where.exe; the controls quoted from main's CI log after landing; a narrow %TEMP% watcher.
- 10:23 Report `M-2026-10-19-fda68260` (opus): branch `T-5613` at `959c6a70` (one commit on `d0340731`; ci.yml +301 only;
  install.ps1 unchanged). Climb read in the main checkout: local, green, git.sha `959c6a70f7e3ce64eb6a86c279b19c4e7d6bb84d`
  = HEAD, dirty false, rungs 0-5 green (guard 3.1 s, vet 4.1 s, schemas 0.7 s, tests 62.1 s 4764 seen, tagged 63.4 s
  4770 seen, git 0.04 s), sha256 `414b8f5f11cf4d038f9bcc9b9619c8cc4e5d289ab178cb7cafd1969bfcf08cc9`. Eleven broken
  copies, each by an exact-once replace; a finally block restores the runner. Not run anywhere yet: four assumptions
  about the runner (csc's temp folder, the stub module's autoload, the event queue, CreateEnvironmentBlock's expansion).
- 10:24 The orchestrator: both PowerShell blocks of the branch's ci.yml parse in Windows PowerShell 5.1 with 0 errors
  (`%USERPROFILE%\bonsai-checks\end-5.6\T-5613-parse.ps1`, natively). T-5613 to `verify`; a fresh verifier (opus; CI
  and the elevated installer are risky work) briefed `M-2026-10-19-940c97af`, the task, the diff and done_when only.
- 10:26 Note: this report's times were first written ahead of the clock and corrected from the messages' file times.
- 10:32 Verifier `M-2026-10-19-88913182` (opus, 10:24-10:32): **FAIL on B1**: the -Remove "no file left" check
  (ci.yml:596) has no broken copy of its own (record-kept, path-kept trip other checks). Held: the other nine broken
  copies change what their checks observe; all eleven replaced texts occur once (install.ps1 is LF everywhere); 7c's 1
  and 18 asserted (ci.yml:564, :571); a broken copy that passes fails the job (:422-427), so a wrong runner assumption
  shows red, never a quiet green; the finally block (:643-647) restores the runner; 301 lines added, none removed. Not
  blocking: the user PATH and -Remove exit code in Test-Remove, and Test-FolderTarget's exit 1 and record checks, sit
  under properties proved on the same broken copies. The verifier's own climb rewrote the result (sha256 now
  `3c16b752...4343ed`); the landing quotes the next climb. Fix sent to the builder (B1 only). T-5613 back to `running`;
  T-5614 to `blocked` (waiting on the director's three scripts) so the grant stands alone.
- 10:36 Report `M-2026-10-19-817b73bb` (opus): `7a05c12d`, one line on `959c6a70`: the broken copy `file-kept`
  (`{ Remove-Busy $f }` to `{ }`, once in install.ps1 at :168) through Test-Remove at the scratch target, caught by
  "remove: left" (ci.yml:597); the step's header comment says so. My read of the diff: exactly the verifier's B1 fix,
  nothing else; no second verifier round for this one line (B1's own prescription), said here.
- 10:36 **Landed.** Climb read in the main checkout: local, green, git.sha
  `7a05c12d785940a469a0115d4a8e8c343e55f14f` = HEAD, dirty false, rungs 0-5 green (guard 3.2 s, vet 0.8 s, schemas
  0.4 s, tests 58.9 s 4764 seen / floor 4756, tagged 59.2 s 4770 seen / floor 4762, git 0.04 s), sha256
  `4b570bb46f21003e3f0f0fa87e510c93653969e781448db00110eeb7d7ea1d98`, its log record naming `/usr/local/bin/bonsai`.
  Diff inside Owns (ci.yml only, +303). No local Windows half (the director's 18 Oct waiver for 5.6). `main`
  fast-forwarded `d0340731..7a05c12d` and pushed. T-5613 to `verify` until CI's install job is read: 12 "broken
  copy ... caught" lines expected.
- 11:31 **CI green on `7a05c12d`, every job** (run 39202835809: test, lint, govulncheck, windows, install ubuntu and
  windows-latest, packs, Analyze Go). The new step's log (`~/bonsai-checks/end-5.6/T-5613-ci.log`): **12 of 12 broken
  copies caught** (compiles: `RES84C7.tmp` made in %TEMP%; path-sz: String not ExpandString; path-expanded: the other
  entries changed; path-unset: where.exe in a new logon's environment found nothing; notice-left: machine '1';
  child-cmdlet: the stub loaded; mask-no-wa and rights-pass: exit 0, not 1; folder-taken; record-kept; path-kept;
  file-kept: `scratch\bonsai.exe` left), and **6 good cases passed** (an install; the stub not loaded; Users Modify and
  write-attributes folders refused with 1 and 18; a folder target refused; install and -Remove, scratch and real).
  install (windows-latest) took 1 min 54 s; the windows test job 53 min 8 s. T-5613 `done`. The four runner
  assumptions held. The runbook's step 7a-7d closed by CI; 7e to 5.7.
