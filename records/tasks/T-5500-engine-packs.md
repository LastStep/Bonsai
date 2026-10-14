---
format: bonsai.task/1   # fields: bonsai check --schema bonsai.task
id: T-5500
title: "The engine's part for packs: protocols paths, roles and limits in force, status's roles (5.5.0)"
status: verify
lane:
done_when:
  - "A pack file's path may start with <protocols>/, written into the project's documents.protocols folder and imported by the block; Go tests for a non-default and a moved protocols folder and each refusal"
  - "A lock's roles from a fixture pack, and each limit's default and max in force from the pack, from the project and from both; a raised default or max not looser to the settings record"
  - "check --pack finds a declared role with no file, a role file not declared, and a maxTurns other than the declared max_turns; status --json's roles against its schema"
  - "init writes STATE in a t.TempDir() project and never overwrites it; go generate changes nothing, and a changed template in a temporary copy fails the test; P0's privacy read of STATE's template before it lands"
  - "Its last commit (.gitattributes) made while T-5500 is the only task running"
  - "Check 10 on both sides; the installed bonsai's ladder green at the branch's HEAD; landed by the landing rule; CI green; VM (roles and limits) and V1 (<protocols>/) pass"
depends_on: [T-5509]
blocked_by:
created: 2026-10-14
started: 2026-10-15
finished:
labels:
  bonsai.branch: "t5500"
  bonsai.ladder: []
  bonsai.allows: [".gitattributes"]
  bonsai.wants: []
---

# The engine's part for packs: protocols paths, roles and limits in force, status's roles (5.5.0)

Piece 5.5.0 of step 5.5 (approved 14 Oct): its row in `design/plan-5.md`'s "### Step 5.5" "The pieces and their order", and its note in "Notes per piece".

- **Owns:** `internal/engine/` (the `<protocols>/` start, `init`'s STATE and its `roles: null` line, the generators, the roles in `declares`, the `.bonsai/.gitignore` comment); `internal/workspace/` (`pack.go`, `paths.go`, a `packs` entry's `roles`, the `messages` setting, the limits in force); `check`'s role findings; `check --pack`'s rules; `roles` in `internal/status/`; `internal/reference/generatedfiles.go` and `internal/reference/gen/`; `docs/reference/lists.md`; `docs/reference/generated-files.md` (removed); base's three generated skill folders; a declared role in `DocumentedPack`; its last commit: `.gitattributes`.
- **Grants:** `.gitattributes` (protected, not person-only), for its last commit alone, made while it is the only task `running`.
- **Proof:** the lines of `done_when`; the installed `bonsai`'s ladder green at the branch's HEAD; the landing rule; CI. P0; VM; V1.
