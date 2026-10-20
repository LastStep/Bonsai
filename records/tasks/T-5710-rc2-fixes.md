---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task
id: T-5710
title: "rc.2's fixes: the list gathered on 20 Oct (each its own task when it starts, T-5712 on)"
status: todo
lane:
done_when:
  - "each item below landed as its own task with its proof, or put after 1.0 on the director's word; then rc.2 (note 5.7.10)"
depends_on: [T-5709]
blocked_by:
created: 2026-10-20
started:
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
