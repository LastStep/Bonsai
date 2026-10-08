# Bonsai, rebuilt as one small program that gives every project the same structure, guards and proof

- Status: draft, waiting for Rohan's approval with the plan (`design/plan.md`)
- Written: 8 Oct 2026, adapted from the studio's one-pager and moved here on Rohan's word
- Design: `design/bonsai-spec.md` (the spec), `design/contract.md` (the formats), `design/format-review.md` (Rohan's
  confirmed review of every format)

**One sentence:** a project gets its tasks, run reports, guards, recorder and proof ladder from one Go program and two
public packs, the same on WSL and on Windows, instead of from Node scripts copied into each repo.

## Why now

The studio's (Trinetra's) guards, ladder and recorder are Node files that live inside its repo, with a frozen copy in
its first game, Mimas. Every new project would need its own copy. Rohan chose on 5 Oct to redesign Bonsai as the thing
that holds that structure; on 7 Oct he confirmed every format (they are final) and answered the spec's four open
questions (spec section 19). On 8 Oct he moved Bonsai's design and records into this repo, made the formats' schemas
and trick files Bonsai's first job, and kept the spec's order inside Bonsai ("lets do this the proper way. no
shortcuts"): the formats first, then the walking skeleton, a cheap proof that the hard parts work, then his gate.

## What it is, in plain words

Bonsai gives a project formats (task, run report, log line, question to Rohan), packs (roles, lanes and protocols as
Claude Code plugins), guards that stop mistakes, a recorder that keeps a clean log, and the ladder that proves work is
done. It never moves a task, applies an approval or starts an agent: the studio does those (Rohan, 7 Oct). It is one
binary with nothing else to install, on Linux and Windows. Spec sections 1-3.

## What correct means

- Every command runs without a person: no prompt without a terminal, `--json` everywhere, every refusal names the next
  step (spec section 3).
- Writing and reading are byte-stable and path-stable: forward slashes, line-ending-blind hashes, no absolute path in a
  committed file, and the rules learnt from Windows hold from day one.
- `init` and `update` change nothing when nothing changed, show what they would write before writing it, and never
  overwrite a file the project edited (the lock, spec section 6).
- The guard fails closed: a missing binary, a crash or an over-time run blocks. A hook starts in under 5 ms on Linux
  (spec section 3); on Windows Git Bash's own start dominates (65 ms measured on 7 Oct, spec section 15).
- Every reader, Bonsai's and the studio's, reaches the same outcome on every trick file in `formats/` (contract section
  2.4).
- Every template and pack file documents itself (contract section 2.8).

## The gate

The walking skeleton builds the risky parts on scratch copies and reports measured numbers: hours per part against
estimate, Windows-only failures and their cost, hook start-up on both sides, the fail-closed results, in which session
kinds a pack's roles load, plugin install time, two projects at two commits, binary size. Rohan then picks path (a) the
full 1.0, (b) the smaller cut, or (d) 1.0 with Bonsai's own screens (spec section 15). If he pauses, `main` keeps what
was built so far and the old product stays at tag `v0.4.3`.

## Done when

- [ ] The formats are in `formats/`: a JSON Schema for each of the contract's ten formats and the trick files with
      their expected outcomes, read rule by rule by a fresh verifier; STATE names the commit (plan
      part 0).
- [ ] The skeleton's twelve checks pass (spec section 14; check 7's studio half stays open until the studio links,
      step 7) and its gate report is in Rohan's hands.
- [ ] Rohan picks a path at the gate, on the measured numbers.
- [ ] Path (a), if chosen: Bonsai 1.0 as spec section 14's parts 5.1-5.7, then the studio links to it (step 7) and
      Mimas last (step 8).

## Not in this

The registry, the bridge, forwarding, the studio's dashboard (the Desk), notifications, deploys and machine keys (all
the studio's). Dispatch, status moves, approvals and grants (the studio's). Outside services (studio connectors). A
managed settings file and a Claude Code version pin (dropped, question B). Routines, a dispatch guard and the
assistant's runtime (parked until after 1.0). Agents other than Claude Code. The full list is spec section 2.

## Open questions for Rohan

None of its own. Where the trick files live was answered on 8 Oct: in this repo's `formats/`, the master; the studio
adopts the formats when it links (step 7).

## Cost

The formats 6-10 AI hours (stop line 13 h). The skeleton 30-47 AI hours, stop line 61 h, plus about 45 minutes of
Rohan's hand checks in two sittings (about 25 minutes after part 5, about 20 after part 4; Rohan, 8 Oct). After the
gate, path (a) is 139-218 h; Bonsai 1.0 with the formats and the skeleton 175-275 h (280-440 on the studio's record of
estimates growing 1.6 times). Proof until spec step 5.4: Go tests and `go vet` in WSL and natively on
Windows, Bonsai's CI on Linux and Windows, and a fresh verifier; from 5.4, Bonsai's own ladder.
