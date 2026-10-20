#!/bin/sh
# release-notes.sh: writes a release's notes, its version's section of CHANGELOG.md, to a file.
#
# Usage: sh .github/scripts/release-notes.sh <tag> <out>
#   <tag>  vX.Y.Z or vX.Y.Z-rc.N, as: v1.0.0, v1.0.0-rc.1 (a pre-release takes its version's section)
#   <out>  the file to write, always outside the checkout (in a job, "$RUNNER_TEMP/notes.md"): an untracked file in
#          the checkout would make the tree dirty, and the programs GoReleaser builds would say so
#          (vcs.modified=true).
#
# It finds the first heading in CHANGELOG.md that starts "## [X.Y.Z]" and writes the lines after it, up to the next
# line starting "## ", with blank lines at the start and the end dropped. The changelog read is the one at the top of
# this script's repository; BONSAI_CHANGELOG names another file (its own tests use that).
#
# Who runs it: CI's release-check job, on every push, for the changelog's newest section (so a broken changelog shows
# on the push that broke it); the release workflow, whose publish job gives <out> to gh release create --notes-file.
#
# Exit 0 when the notes are written; 1 when the tag is not a version tag, the section is missing or holds nothing,
# or a file cannot be read or written, with one sentence naming the file.

set -u

fail() {
	printf 'release-notes.sh: %s\n' "$1" >&2
	exit 1
}

[ $# -eq 2 ] || fail "it takes two arguments, the tag and the file to write, as: sh $0 v1.0.0 \"\$RUNNER_TEMP/notes.md\""
tag=$1
out=$2

version=$(printf '%s\n' "$tag" | sed -n 's/^v\([0-9][0-9]*\.[0-9][0-9]*\.[0-9][0-9]*\)\(-rc\.[0-9][0-9]*\)\{0,1\}$/\1/p')
[ -n "$version" ] || fail "$tag is not a version tag (vX.Y.Z or vX.Y.Z-rc.N)"

changelog=${BONSAI_CHANGELOG:-$(dirname -- "$0")/../../CHANGELOG.md}
[ -f "$changelog" ] || fail "$changelog does not exist"

# The section: from the heading's next line up to the next "## " heading, blank lines dropped at both ends.
notes=$(awk -v head="## [$version]" '
	found && /^## / { exit }
	found { lines[n++] = $0 }
	!found && index($0, head) == 1 { found = 1 }
	END {
		if (!found) exit 2
		first = 0; while (first < n && lines[first] ~ /^[ \t]*$/) first++
		last = n - 1; while (last >= first && lines[last] ~ /^[ \t]*$/) last--
		for (i = first; i <= last; i++) print lines[i]
	}' "$changelog")
case $? in
0) ;;
2) fail "$changelog has no section headed ## [$version]" ;;
*) fail "could not read $changelog" ;;
esac
[ -n "$notes" ] || fail "$changelog's section ## [$version] is empty"

printf '%s\n' "$notes" >"$out" || fail "could not write $out"
