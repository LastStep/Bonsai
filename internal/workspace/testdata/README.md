# Test data for internal/workspace

`test-pack-a/pack.yaml` is a byte-for-byte copy of `bonsai/pack.yaml` from Bonsai's test pack:

- repository: https://github.com/LastStep/bonsai-test-pack (public)
- commit: `506205354b7589f82f849820987aad17dba3309d` (commit A)
- file: `bonsai/pack.yaml`, git blob `e679d2d5396a4cbbd69f4898ed22dab3b744f90c`

It is copied so the tests need no network (plan part 2: "part 2's reader reads its `pack.yaml`").
`TestReadPackTestPackA` checks the copy's git blob hash, so an edited copy fails the test; to take a newer commit,
copy the file again with `git show <commit>:bonsai/pack.yaml` and change the commit and blob here and in the test.
`.gitattributes` (`*.yaml text eol=lf`) keeps its bytes on a Windows checkout.
