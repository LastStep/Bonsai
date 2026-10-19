---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task
id: T-5614
title: "5.6's end Windows run, the director's half (runbook steps 4, 5, 10) and check 8's WSL deny-rule try"
status: done
lane:
done_when:
  - "one PowerShell batch for the director under %USERPROFILE%\\bonsai-checks\\end-5.6\\person\\: step 4 (the warm-up and login check), step 5 (the cross-side deny rule refuses Write and Edit in a native Windows Claude Code session; the control with --add-dir and no rule passes), step 10 (the guard refuses Write and a Bash redirect through a junction out of the project; a dangling junction is bad-path); every setup made and dry-run by agents, each step one script the director runs, each printing PASS or FAIL against the runbook's expectation and saving its log"
  - "check 8's WSL half: the same cross-side deny rule tried in a scripted WSL Claude Code session, isolated (the decoy refused by the rule, the control passing), user settings file hashes recorded before and after"
  - "the director's results read and recorded in the run report; a failure reported with its log, not fixed here; nothing in the repository changed"
depends_on: [T-5603, T-5699]
blocked_by:
created: 2026-10-19
started: 2026-10-19
finished: 2026-10-19
labels:
  bonsai.ladder: []
  bonsai.allows: []
  bonsai.wants: []
---

# 5.6's end Windows run, the director's half (T-5614)

The director, 19 Oct: "we will finish up 5.6 in this session"; step 7 moved to CI (T-5613) and 7e to 5.7, since this
computer has no Windows Sandbox. Prepared by a builder (scratch repositories and stamped builds); the director runs
the three scripts. Kept `todo` while T-5613 holds `.github/workflows/ci.yml` (a granting task runs with no other
`running`); its preparation writes only scratch folders.

- **Owns:** nothing in the repository; `~/bonsai-checks/end-5.6/` and `%USERPROFILE%\bonsai-checks\end-5.6\person\`.
- **Grants:** none.
