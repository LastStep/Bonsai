---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task; how to fill: skill base:task
id: T-5725
title: "BUG-014: a subfolder pack counts only its own-name tags; V1's note C"
status: done
lane: null
done_when:
  - "bonsai status no longer lists a bare v tag (v0.4.3, v1.0.0) as a newer release of a pack kept in a subfolder (path set), shown by a test failing on the old code"
  - "a pack at a repository's top (path null, as workflow) still reads bare v tags as its releases, by test"
  - "the release-workflow shell test refuses a flow-style job default such as defaults: {run: {shell: sh}}, by a test failing on the old code"
  - "the climb at the branch's HEAD green, CI on the branch green"
depends_on: []
blocked_by: null
created: 2026-10-23
started: 2026-10-23
finished: 2026-10-23
labels:
  bonsai.branch: T-5725
  bonsai.ladder: []
  bonsai.allows: []
  bonsai.wants: []
---

# BUG-014: a subfolder pack counts only its own-name tags; V1's note C

## Context

T-5710 items 1 and 5. `internal/engine/newer.go` `NewerTags`: when a pack's ref is not a release tag it counts tags
with no prefix, so Bonsai's own tags read as newer `base` releases. The director, 23 Oct: "Subfolder packs:
own-name tags" (a pack with a `path` counts only `<id>-`/`<id>/` tags; a top-level pack keeps bare `v` tags). V1's
note C (`M-2026-10-20-2a5d18eb`): `internal/selfcheck/workflows_test.go` `releaseShellProblems` reads block style only.

## Out of scope

release.yml itself (the director, 23 Oct: its notes A and B after 1.0).
