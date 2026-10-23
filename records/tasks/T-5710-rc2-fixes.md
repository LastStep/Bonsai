---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task
id: T-5710
title: "rc.2's fixes: the list gathered on 20 Oct (each its own task when it starts, T-5712 on)"
status: verify
lane:
done_when:
  - "each item below landed as its own task with its proof, or put after 1.0 on the director's word; then rc.2 (note 5.7.10)"
depends_on: [T-5709]
blocked_by:
created: 2026-10-20
started: 2026-10-23
finished:
labels:
  bonsai.ladder: []
  bonsai.allows: []
---

# rc.2's fixes (T-5710)

Gathered on 20 Oct; rc.2 always follows (note 5.7.10). Each becomes its own task when it starts.

1. **BUG-014** (the trial's proposal A): `bonsai status` reads Bonsai's old v0.1.1-v0.4.3 tags, and from v1.0.0 the 1.0
   tag, as newer base releases (`internal/engine/newer.go:63`). Must land before 1.0. The proposals' options:
   (1) a pack in a subfolder counts only tags starting with its id (recommended), (2) bare v tags never count.
2. **The notes' leftover line** (V2's risk, M-2026-10-20-3434f47c): CHANGELOG.md:48's line about the old product's
   sections ends the published notes with nothing after it.
3. **The changelog line for the trial's feature** (note 5.7.8).
4. **Go:** go1.27.3 or later, with govulncheck able to load it (T-5794 cut, the director's "Stay on 1.27.2").
5. **V1's notes** (M-2026-10-20-2a5d18eb): release.yml:141 and :282; the shell rule's block-style reading.
6. **The trial verifier's risks** (M-2026-10-20-57b77ed9): the stop gate's rule word bypasses the word filter; the
   skill's "the one the log keeps"; `bonsai hook stop <arg>` refuses without a rule line.
7. **CLAUDE.md's stale lines:** the stricter-only list calls an added rung stricter (the code: looser,
   internal/workspace/checks.go:18-24); the named-session line names `claude-here`, which runs in scratch only (the
   trial's finding); "being rebuilt" at line 3 (for 1.0).
8. **The README and demos** (T-5711), talked through with the director first.
9. Check 10's Windows half after rc.1 (note 5.7.5, 8): not yet run.
10. **`bonsai agents` filters** (the director, 20 Oct: "bonsai agents command shows all agents at once, need some
    filter there"): it lists every agent at once; filters to add (by session, task, role, state or age; to plan with
    the director).
11. **`bonsai answer` by option letter** (the director, 20 Oct: "bonsai answer needs to have an options key, like where
    we can answer just A, B etc; currently we use words"): a Decide answered by the option's letter or number, written
    as the option's full text in `choice` (the format's field unchanged); the trial's ask 1 and 2 were answered with
    `--words "B"` because `--choice` needs the option's whole text.
12. **From T-5711 (22 Oct):** BUG-097 (the ladder's leftovers line names other programs: a way to leave them out of a
    printed climb); after fix 11 lands, the README's line "(rc.2 puts that first)" in step 6 rewritten to match; after
    rc.2 is installed, the clip set filmed again where what is on screen changed (`sh docs/demos/clips/film.sh
    ~/bonsai-clips <clip>`, then `sh docs/demos/edit.sh`), clip 10 among them (BUG-098); `~/bonsai-checks/bin/claude-here`
    forces its own Bonsai home and plugin cache, so a dry run through it cannot see the stop gate judge a project
    linked with another home (T-5724; a scratch tool's limit, for T-5793).
13. **Other agents' markers refuse person-only commands** (the director, 23 Oct, C4 and C8 of
    `design/agent-neutral-walkthrough.md`): wherever Claude's child marker refuses `update --yes --allow-exec`,
    `unlink --yes` and looser settings today, any of `CLAUDE_CODE_CHILD_SESSION`, `AI_AGENT`, `PI_CODING_AGENT`,
    `CODEX_THREAD_ID`, `GEMINI_CLI`, `CURSOR_AGENT`, `OPENCODE`, `GOOSE_TERMINAL` refuses (`CLAUDECODE` left out: it
    reaches editor terminals); the refusal names the marker it found; the test helper clears the list. The director's
    terminals print none of them (checked 23 Oct). Before 1.0, because pi runs with the stopgap extension before then.

## The director's words, 23 Oct (the 1.0 session)

- Item 10, `bonsai agents` filters: **after 1.0** ("After 1.0").
- Item 1, BUG-014: **"Subfolder packs: own-name tags"** (a pack with a `path` counts only its `<id>-` tags; a pack at a
  repository's top keeps bare `v` tags).
- Item 5, V1's notes: **"Fix the test only"**: note C (the shell test's flow style) in rc.2; release.yml's notes A and B
  after 1.0.
- Item 4, Go: govulncheck's newest (v1.9.0, 11 Oct) still cannot load go1.27.3; rc.2 stays on go1.27.2 by the
  director's 20 Oct word, checked again before 1.0.
- Item 9: check 10's Windows half runs on rc.2, not rc.1.
- Tasks: T-5725 (1, 5C), T-5726 (6), T-5727 (13), T-5728 (11), T-5729 (12's BUG-097, the dry run's Bonsai-words 11
  and 13); items 2, 3, 7 and 12's README line in a records task after them.
