#!/bin/sh
# release-tag-check.sh: checks that a pushed tag may be released, before the release workflow builds anything.
#
# Usage: sh .github/scripts/release-tag-check.sh [--answer <file> --own-suite <id>]
#
# What it reads, all set by GitHub Actions in the release workflow's build job:
#   GITHUB_REF_NAME    the tag's name, as v1.0.0 or v1.0.0-rc.1
#   GITHUB_SHA         the commit the tag names
#   GITHUB_REPOSITORY  owner/repo, as LastStep/Bonsai
#   GITHUB_RUN_ID      this run, whose own check suite is left out of the checks below
#   GH_TOKEN           the job's read-only token (contents, checks and actions read), for gh api
#
# It checks, in order, and stops at the first that fails, naming it:
#   1. the name: ^v[0-9]+\.[0-9]+\.[0-9]+(-rc\.[0-9]+)?$ (the workflow's tag patterns, read again here);
#   2. the commit is on main: git merge-base --is-ancestor "$GITHUB_SHA" origin/main (the job's checkout has full
#      history, so origin/main is there);
#   3. the commit's checks, from GET repos/<repo>/commits/<sha>/check-runs (every page; GitHub's default filter,
#      latest, so a job re-run after a failure counts by its newest attempt), leaving out this run's own check suite
#      (its check_suite_id, from GET repos/<repo>/actions/runs/<run id>):
#      - each named job below has at least one check run, and every check run of that name reads completed with
#        success (one still queued or running, or any other conclusion, fails);
#      - every other check run is not failure, cancelled or timed_out (neutral, skipped or still running passes).
#
# The named jobs are CI's (.github/workflows/ci.yml) by their check run names, the matrix job install once per
# system, and CodeQL's (.github/workflows/codeql.yml). A job added to, renamed in or taken out of those files is
# changed here in the same commit.
#
# For its tests only (the release workflow never passes them): --own-suite <id> names the check suite to leave out
# instead of reading this run's, so the query can be tried from a person's machine (GITHUB_RUN_ID is then not read);
# --answer <file>, with --own-suite, reads the check runs from a saved answer of that endpoint (one page, or the pages
# one after another, as gh api --paginate writes them) instead of asking GitHub: no network call, no GH_TOKEN.
#
# Needs: git, gh (logged in through GH_TOKEN unless --answer is given) and jq, all on GitHub's ubuntu-latest runner.
# Exit 0 when the tag may be released; 1 with one line saying what failed; 2 for a wrong call.

set -u

named='test
windows
windows (fault tag)
lint
govulncheck
release-check
packs
install (ubuntu-latest)
install (windows-latest)
Analyze Go'

fail() {
	printf 'release-tag-check.sh: %s\n' "$1" >&2
	exit 1
}

usage() {
	printf 'release-tag-check.sh: %s\nusage: sh .github/scripts/release-tag-check.sh [--answer <file> --own-suite <id>]\n' "$1" >&2
	exit 2
}

answer=
own=
while [ $# -gt 0 ]; do
	case $1 in
	--answer)
		[ $# -ge 2 ] || usage "--answer needs a file"
		answer=$2
		shift 2
		;;
	--own-suite)
		[ $# -ge 2 ] || usage "--own-suite needs a check suite id"
		own=$2
		shift 2
		;;
	*) usage "unknown argument: $1" ;;
	esac
done
if [ -n "$answer" ] && [ -z "$own" ]; then
	usage "--answer needs --own-suite"
fi

tag=${GITHUB_REF_NAME:-}
sha=${GITHUB_SHA:-}
repo=${GITHUB_REPOSITORY:-}
[ -n "$tag" ] || fail "GITHUB_REF_NAME is empty: the tag's name is needed"
[ -n "$sha" ] || fail "GITHUB_SHA is empty: the tag's commit is needed"
[ -n "$repo" ] || fail "GITHUB_REPOSITORY is empty: owner/repo is needed"
command -v jq >/dev/null 2>&1 || fail "jq is not on the PATH"

# 1. The name.
printf '%s\n' "$tag" | grep -Eq '^v[0-9]+\.[0-9]+\.[0-9]+(-rc\.[0-9]+)?$' ||
	fail "tag $tag is not vX.Y.Z or vX.Y.Z-rc.N: no release is made from it"
echo "tag $tag: a version tag"

# 2. On main.
git rev-parse --verify --quiet origin/main >/dev/null || fail "origin/main is not in this checkout: fetch it with full history"
git merge-base --is-ancestor "$sha" origin/main ||
	fail "commit $sha (tag $tag) is not on main: a release is made only from a commit main holds"
echo "commit $sha: on main"

# 3. The checks.
tmp=$(mktemp) || fail "cannot make a temporary file"
trap 'rm -f "$tmp"' EXIT
if [ -n "$answer" ]; then
	[ -r "$answer" ] || fail "cannot read the saved answer $answer"
	cp "$answer" "$tmp" || fail "cannot read the saved answer $answer"
else
	if [ -z "$own" ]; then
		[ -n "${GITHUB_RUN_ID:-}" ] || fail "GITHUB_RUN_ID is empty: this run's own check suite cannot be left out"
		own=$(gh api "repos/$repo/actions/runs/$GITHUB_RUN_ID" --jq .check_suite_id) ||
			fail "cannot read this run (repos/$repo/actions/runs/$GITHUB_RUN_ID)"
	fi
	gh api --paginate "repos/$repo/commits/$sha/check-runs?per_page=100" >"$tmp" ||
		fail "cannot read the checks of $sha (repos/$repo/commits/$sha/check-runs)"
fi
case $own in
'' | *[!0-9]*) fail "this run's check suite id is not a number: '$own'" ;;
esac
echo "this run's own check suite, left out: $own"

# The answer read before it is judged: one or more JSON objects (pages), each with a check_runs list. A pipeline's
# status is its last command's in sh, so every jq call below writes to a file and its own status is tested; a jq
# failure never reaches a sort that would hide it.
[ -s "$tmp" ] || fail "the answer of the check-runs endpoint is empty"
jq empty "$tmp" >/dev/null 2>&1 || fail "the answer of the check-runs endpoint is not JSON"
jq -ne '[inputs] | length > 0 and all(.[]; type == "object" and (.check_runs | type) == "array")' "$tmp" >/dev/null 2>&1 ||
	fail "the answer is not an answer of the check-runs endpoint: a page has no check_runs list (GitHub's message: $(jq -nc '[inputs | .message? // empty] | first // "none"' "$tmp" 2>/dev/null))"

# One line per problem, sorted, so the output is the same for the same answer.
found="$tmp.problems"
trap 'rm -f "$tmp" "$found"' EXIT
jq -nr --arg own "$own" --arg named "$named" '
	($named | split("\n")) as $names
	| [inputs | .check_runs[] | select((.check_suite.id | tostring) != $own)] as $runs
	| (
		($names[] as $n
			| [$runs[] | select(.name == $n)] as $mine
			| if ($mine | length) == 0 then "\($n): no check run on this commit"
			  else ($mine[] | select(.status != "completed" or .conclusion != "success")
				| "\($n): \(.status)\(if .conclusion then ", " + .conclusion else "" end)")
			  end),
		($runs[] | select(.name as $x | $names | index($x) | not)
			| select(.conclusion == "failure" or .conclusion == "cancelled" or .conclusion == "timed_out")
			| "\(.name): \(.conclusion)")
	)
' "$tmp" >"$found" || fail "cannot read the check runs in the answer"
problems=$(LC_ALL=C sort -u "$found") || fail "cannot sort the problems found"
if [ -n "$problems" ]; then
	printf '%s\n' "$problems" | sed 's/^/  /' >&2
	fail "the checks of $sha do not allow a release (above)"
fi
echo "checks of $sha: every named job completed with success, and no other check failed"
