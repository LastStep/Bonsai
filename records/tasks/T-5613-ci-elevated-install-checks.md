---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task
id: T-5613
title: "5.6 end run step 7a-7d moved into CI's Windows install job (5.6.6's elevated installer cases)"
status: done
lane:
done_when:
  - "CI's install job on windows-latest, through the forced elevated child, checks the runbook's 7a-7d (~/bonsai-checks/end-5.6/runbook.md): 7a no new .cs, .dll or .tmp in %TEMP% after an install, the machine Path still REG_EXPAND_SZ with its %...% entries kept, a fresh process finds bonsai by the machine Path, no BONSAI_INSTALL_NOTICE left; 7b a stub Microsoft.PowerShell.Utility module ahead on PSModulePath writes no marker; 7c a pre-made install folder at the real place granting Users Modify refused with nothing installed, then with only write-attributes the same, the script's exit code read from the code and asserted exactly; 7d a folder given as -Target refused, then install and -Remove leave no file, no Path entry and no record"
  - "each new check shown failing at least once on a run against a deliberately broken copy of install.ps1 (a negative control inside the job itself, since no branch is pushed before landing; or natively by the orchestrator where no elevation is needed), its log quoted in the report; install.ps1 itself unchanged unless a check finds a real fault, which is then reported before any fix"
  - "CI green on the landed commit, every job; a green climb; WSL check 10"
depends_on: [T-5606]
blocked_by:
created: 2026-10-19
started: 2026-10-19
finished: 2026-10-19
labels:
  bonsai.ladder: []
  bonsai.allows: [".github/workflows/ci.yml"]
  bonsai.wants: []
  bonsai.branch: T-5613
---

# Step 7a-7d of 5.6's end Windows run, in CI (T-5613)

The director, 19 Oct: this computer runs Windows 11 Home, which has no Windows Sandbox, so the elevated installer cases
(runbook step 7) move to CI's throwaway windows-latest machine, already an administrator; "yes, go ahead". Step 7e
(the prompt from an unelevated session, declined) needs a person and joins the director's real install at 5.7.

- **Owns:** `.github/workflows/ci.yml`'s install job (its Windows step, or a new Windows step beside it); a helper
  script under `install/` only if the job's step grows past reading; `install/install.ps1` only for a real fault a new
  check finds, reported first.
- **Grants:** `.github/workflows/ci.yml` (protected, not person-only): a stricter change, checks added, none removed.
