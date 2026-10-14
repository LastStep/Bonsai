# ci/check.sh: a Bonsai pack's checks, in one place. The pack's CI (.github/workflows/pack.yml) runs it, and Bonsai's
# own CI runs the template's copy on every pack folder in Bonsai's repository, so a pack is checked exactly as Bonsai
# checks its template. Written for sh (Linux, and Git Bash on Windows) and always run as `sh ci/check.sh ...`, so it
# needs no executable mode. It prints what it checks, then "ok" or the reason it failed; it exits 0 when every check
# passed, 1 when one failed, and 2 on a usage error.
#
# Subcommands and their arguments:
#
#   sh ci/check.sh check <bonsai> <folder> <source> <commit> [<path>]
#     1. bonsai check --pack <folder> --json: the pack's declarations and documentation, with no finding.
#     2. README.md's rows: every file in agents/ and skills/ has a row in a table of README.md whose first cell names
#        it in backquotes (| `agents/example.md` | ... |), and every such row names a file there.
#     3. A scratch project in a temporary folder (git init; bonsai init --name ci --source <source> --ref <commit>
#        [--path <path>] --yes --json), with claude off the PATH, so its plugin step reports skipped and no plugin
#        is installed on the machine; its Bonsai home is a temporary folder too, removed at the end.
#     4. bonsai check --json in that project, with no finding (its warnings are printed, never a failure).
#     <bonsai>  the bonsai program to run: a path to a build of the Bonsai commit the pack pins (bonsai.exe on Windows)
#     <folder>  the pack's folder on disk, the one that holds bonsai/pack.yaml: . in the pack's own repository
#     <source>  the git repository the scratch project links the pack from: a local checkout's folder (its commit is
#               copied first into the temporary folder, so the project's bonsai.yaml names no absolute path) or a URL
#     <commit>  the 40-character commit of <source> to link: the commit being checked
#     <path>    the pack's folder inside <source>; left out (or .) for the repository's top
#
#   sh ci/check.sh validate <folder>
#     claude plugin validate --json <folder>, Claude Code's own check of a plugin: it fails on every error and every
#     warning but the missing version, which a Bonsai pack leaves out on purpose (README.md, "Versions and
#     releases"); so not --strict, which fails on that one too. It needs claude (the version the pack's CI pins) and
#     node, which reads validate's JSON, on the PATH, and no Claude login. The fields it reads are those Claude Code
#     2.1.300 prints: success; manifest.errors and manifest.warnings, each with its path (the version's is
#     "version"); and contents, one entry per role, skill or command file with a problem, each with its file, errors
#     and warnings. Its notes and advice are printed, never a failure.
#     <folder>  the pack's folder on disk
#
#   sh ci/check.sh release <tag> <folder>
#     The release check: it fails unless <tag> is "v" and bonsai/pack.yaml's version, quoted there (v0.1.0 for
#     version: "0.1.0"). It reads the name it is given and no git tag, so it runs anywhere with any name.
#     <tag>     the release tag's name, such as v0.1.0 (in CI, the pushed tag's)
#     <folder>  the pack's folder on disk
#
#   sh ci/check.sh version <folder>
#     Prints bonsai/pack.yaml's version alone (0.1.0), as release reads it: the CI's release job names the tag a
#     version needs with it on every push that is not a tag.
#     <folder>  the pack's folder on disk

usage() {
	echo "usage: sh ci/check.sh check <bonsai> <folder> <source> <commit> [<path>]" >&2
	echo "       sh ci/check.sh validate <folder>" >&2
	echo "       sh ci/check.sh release <tag> <folder>" >&2
	echo "       sh ci/check.sh version <folder>" >&2
	echo "(the header of ci/check.sh documents each)" >&2
	exit 2
}

fail() {
	echo "check.sh: FAILED: $*" >&2
	exit 1
}

# abs prints a folder's absolute path in the form the programs it runs read: on Windows (Git Bash) C:/..., else /...
abs() {
	(cd "$1" 2>/dev/null && { pwd -W 2>/dev/null || pwd; })
}

# no_claude prints the PATH without every folder that holds claude, so a bonsai run from here cannot reach it.
no_claude() {
	out=''
	set -f
	old_ifs=$IFS
	IFS=:
	for d in $PATH; do
		[ -n "$d" ] || continue
		if [ -e "$d/claude" ] || [ -e "$d/claude.exe" ] || [ -e "$d/claude.cmd" ]; then
			continue
		fi
		out=${out:+$out:}$d
	done
	IFS=$old_ifs
	set +f
	printf '%s' "$out"
}

# pack_version prints bonsai/pack.yaml's version: the first top-level version: line's quoted value.
pack_version() {
	[ -f "$1/bonsai/pack.yaml" ] || fail "$1 holds no bonsai/pack.yaml: give the pack's folder"
	line=$(grep '^version:' "$1/bonsai/pack.yaml" | head -n 1)
	v=$(printf '%s\n' "$line" | sed -n -e 's/^version:[ ]*"\([^"]*\)".*/\1/p' -e "s/^version:[ ]*'\\([^']*\\)'.*/\\1/p")
	[ -n "$v" ] || fail "bonsai/pack.yaml's version is not a quoted value on its own line: $line"
	printf '%s\n' "$v"
}

# readme_rows checks README.md's rows against the files in agents/ and skills/ (step 2 of check).
readme_rows() {
	[ -f "$1/README.md" ] || fail "$1 holds no README.md"
	# shellcheck disable=SC2016 # the backquotes are the README's, matched as written, not a command to expand
	rows=$(tr -d '\r' <"$1/README.md" | sed -n -E 's/^\| `((agents|skills)\/[^`]*[^`/])` \|.*/\1/p' | sort -u)
	files=$(cd "$1" && find agents skills -type f 2>/dev/null | sort)
	bad=''
	old_ifs=$IFS
	IFS='
'
	for f in $files; do
		printf '%s\n' "$rows" | grep -qxF "$f" || bad="$bad
  $f has no row in README.md"
	done
	for r in $rows; do
		[ -f "$1/$r" ] || bad="$bad
  README.md's row $r names no file"
	done
	IFS=$old_ifs
	[ -z "$bad" ] || fail "README.md's role and skill rows differ from the files:$bad"
}

pack_check() {
	[ $# -eq 4 ] || [ $# -eq 5 ] || usage
	bonsai=$1
	folder=$2
	source=$3
	commit=$4
	path=${5:-.}
	[ -f "$bonsai" ] || fail "$bonsai is not a file: give the path to a bonsai build"
	bonsai_dir=$(abs "$(dirname "$bonsai")") || fail "the folder of $bonsai cannot be read"
	bonsai=$bonsai_dir/$(basename "$bonsai")
	[ -f "$folder/bonsai/pack.yaml" ] || fail "$folder holds no bonsai/pack.yaml: give the pack's folder"
	folder=$(abs "$folder") || fail "$folder cannot be read"
	case $commit in
	*[!0-9a-f]*) fail "the commit $commit is not 40 lower-case hex characters" ;;
	esac
	[ ${#commit} -eq 40 ] || fail "the commit $commit is not 40 lower-case hex characters"

	echo "== bonsai check --pack $folder ($("$bonsai" --version))"
	"$bonsai" check --pack "$folder" --json || fail "bonsai check --pack found the problems above (exit $?)"
	echo "ok: no finding"

	echo "== README.md's rows against agents/ and skills/"
	readme_rows "$folder"
	echo "ok: a row for each file, a file for each row"

	tmp=$(mktemp -d) || fail "mktemp -d could not make a temporary folder"
	trap 'rm -rf "$tmp"' EXIT
	trap 'exit 1' INT TERM
	work=$(abs "$tmp") || fail "the temporary folder $tmp cannot be read"
	# A local source's commit is copied into a bare repository beside the project, on a branch, and linked by its
	# relative path: bonsai check finds an absolute path in a project's bonsai.yaml (contract section 2.6).
	# --update-shallow: CI's checkouts are shallow, and git refuses a branch at a shallow commit without it.
	case $source in
	*://* | git@*) ;;
	*)
		src=$(abs "$source") || fail "the source $source is neither a folder nor a URL"
		git init -q --bare "$work/source.git" || fail "git init --bare failed"
		git -C "$work/source.git" fetch -q --update-shallow --no-tags -- "$src" "+$commit:refs/heads/ci" ||
			fail "the commit $commit could not be fetched from $src"
		got=$(git -C "$work/source.git" rev-parse -q --verify "refs/heads/ci^{commit}")
		[ "$got" = "$commit" ] || fail "the commit $commit could not be copied from $src"
		source=../source.git
		;;
	esac
	git init -q "$work/project" || fail "git init failed"
	PATH=$(no_claude)
	export PATH
	command -v claude >/dev/null 2>&1 && fail "claude is still on the PATH ($(command -v claude))"
	command -v git >/dev/null 2>&1 || fail "git shares a folder with claude on the PATH: install them apart"
	BONSAI_HOME=$work/home
	export BONSAI_HOME
	set -- --name ci --source "$source" --ref "$commit"
	[ "$path" = . ] || set -- "$@" --path "$path"
	echo "== bonsai init $* --yes --json (a scratch project, claude off the PATH)"
	(cd "$work/project" && "$bonsai" init "$@" --yes --json) >"$work/init.json"
	status=$?
	if [ $status -ne 0 ]; then
		cat "$work/init.json"
		fail "bonsai init exited $status"
	fi
	if ! grep -q '"result": "skipped"' "$work/init.json" || grep -Eq '"result": "(installed|waiting|failed)"' "$work/init.json"; then
		cat "$work/init.json"
		fail "init's plugin step did not report skipped: it reached Claude Code"
	fi
	echo "ok: linked, the plugin step skipped"

	echo "== bonsai check --json in the scratch project"
	(cd "$work/project" && "$bonsai" check --json) || fail "bonsai check found the problems above in the scratch project"
	echo "ok: no finding"
}

validate() {
	[ $# -eq 1 ] || usage
	folder=$1
	command -v claude >/dev/null 2>&1 || fail "claude is not on the PATH: install the Claude Code the pack's CI pins"
	command -v node >/dev/null 2>&1 || fail "node is not on the PATH: it reads validate's JSON"
	[ -f "$folder/.claude-plugin/plugin.json" ] || fail "$folder holds no .claude-plugin/plugin.json"
	echo "== claude plugin validate --json $folder (Claude Code $(claude --version))"
	out=$(claude plugin validate --json "$folder")
	status=$?
	printf '%s\n' "$out"
	printf '%s' "$out" | node -e '
		let text = "";
		process.stdin.on("data", (c) => { text += c; });
		process.stdin.on("end", () => {
			let r;
			try { r = JSON.parse(text); } catch (e) { console.error("validate printed no JSON"); process.exit(1); }
			const m = r.manifest;
			if (!m || !Array.isArray(m.errors) || !Array.isArray(m.warnings)) {
				console.error("validate printed no manifest.errors and manifest.warnings: read its output and update ci/check.sh");
				process.exit(1);
			}
			const bad = [];
			for (const e of m.errors) bad.push("error: " + (e.path || "") + ": " + e.message);
			for (const w of m.warnings) if (w.path !== "version") bad.push("warning: " + (w.path || "") + ": " + w.message);
			for (const c of r.contents || []) {
				for (const e of c.errors || []) bad.push("error: " + c.file + ": " + (e.path || "") + ": " + e.message);
				for (const w of c.warnings || []) bad.push("warning: " + c.file + ": " + (w.path || "") + ": " + w.message);
			}
			if (r.success !== true) bad.push("success is not true");
			for (const b of bad) console.error(b);
			process.exit(bad.length ? 1 : 0);
		});
	' || fail "claude plugin validate reported the problems above"
	[ $status -eq 0 ] || fail "claude plugin validate exited $status"
	echo "ok: only the missing version"
}

release() {
	[ $# -eq 2 ] || usage
	tag=$1
	version=$(pack_version "$2") || exit 1
	echo "== release $tag: bonsai/pack.yaml's version is $version"
	[ "$tag" = "v$version" ] || fail "the tag $tag is not v$version: a release tag is v and pack.yaml's version"
	echo "ok: $tag names version $version"
}

[ $# -ge 1 ] || usage
sub=$1
shift
case $sub in
check) pack_check "$@" ;;
validate) validate "$@" ;;
release) release "$@" ;;
version)
	[ $# -eq 1 ] || usage
	pack_version "$1"
	;;
*) usage ;;
esac
