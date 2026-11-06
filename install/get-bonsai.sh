#!/bin/sh
# get-bonsai.sh: installs a Bonsai release on this machine (Linux and WSL) in one command, with every check the
# install page's lines make (docs/install.md), in the same order, stopping at the first that fails.
#
# What it does: checks this machine first (Linux; run as yourself, not under sudo; curl, sha256sum, tar and gh on the
# PATH; gh 2.97 or newer, logged in); takes the newest release (gh api repos/LastStep/Bonsai/releases/latest) or the
# version given; in a new temporary folder of its own (mktemp -d, its path printed and left for you to read), downloads
# the release's archive for this machine (bonsai_<version>_linux_<amd64|arm64>.tar.gz) and its checksums.txt; checks
# the archive's SHA-256 against the one line checksums.txt holds for it; checks where it was built (gh attestation
# verify: Bonsai's own release workflow, from that version's tag, on GitHub's machines); only then unpacks it and runs
# the release's own install.sh, which asks for your password once and puts bonsai at /usr/local/bin/bonsai. It has no
# option that skips a check.
#
# Usage, as yourself (fetch it, read it, run it):
#   curl -fsSLO https://raw.githubusercontent.com/LastStep/Bonsai/main/install/get-bonsai.sh
#   sh get-bonsai.sh            # the newest release
#   sh get-bonsai.sh 1.0.3      # one release, by its version (v1.0.3 too; a pre-release as 1.0.0-rc.1)
#
# Needs: Linux or WSL, sh, curl, sha256sum, tar, and GitHub's command line gh 2.97 or newer, logged in (gh auth login).
#
# Exit 0 when the release's install.sh is done; otherwise 1, with one plain sentence and its next step, and nothing
# installed unless the failure came from install.sh itself (it says what it did).
#
# The way back: run it again with the release you had, as: sh get-bonsai.sh 1.0.2; the way out: the release's own
# install.sh --remove (docs/install.md, "The way out").
set -eu

repo=LastStep/Bonsai
workflow=$repo/.github/workflows/release.yml

refuse() {
	printf 'get-bonsai.sh: %s\nnext: %s\n' "$1" "$2" >&2
	exit 1
}

# A version given is read before anything else: one argument at most, X.Y.Z or X.Y.Z-rc.N, 1.0.0 or newer.
if [ $# -gt 1 ]; then
	refuse "it takes one version at most, and was given $#" "run it again with one version, as: sh get-bonsai.sh 1.0.3, or with none for the newest"
fi
given=${1-}
if [ $# -eq 1 ]; then
	case $given in
	'' | *[!0-9a-z.v-]*) refuse "the version given is not a release version such as 1.0.3 (a version holds only digits, dots and -rc.N)" "run it again with a version from https://github.com/$repo/releases, or with none for the newest" ;;
	esac
fi

# version prints its argument without a leading v when it is a version this script installs, else nothing.
version() {
	v=${1#v}
	printf '%s\n' "$v" | awk '/^[0-9]+\.[0-9]+\.[0-9]+(-rc\.[0-9]+)?$/ && !/^0\./ { print; exit }'
}

if [ $# -eq 1 ]; then
	ver=$(version "$given")
	if [ -z "$ver" ]; then
		case ${given#v} in
		0.*) refuse "'$given' is the old product (0.x), which this script does not install" "read its own install notes at https://github.com/$repo/tree/v0.4.3, or run it with no version for the newest release" ;;
		esac
		refuse "'$given' is not a release version such as 1.0.3 (or 1.0.0-rc.1)" "run it again with a version from https://github.com/$repo/releases, or with none for the newest"
	fi
fi

system=$(uname -s 2>/dev/null) || system=
if [ "$system" != Linux ]; then
	refuse "this machine runs ${system:-a system uname does not name}, not Linux: get-bonsai.sh is for Linux and WSL" \
		"on Windows, use get-bonsai.ps1 (docs/install.md); elsewhere there is no installer"
fi

if [ "$(id -u)" = 0 ]; then
	refuse "it runs as root (under sudo, or as the root user); the release's install.sh refuses that, and asks for the password itself" \
		"run it again as yourself, without sudo: sh get-bonsai.sh"
fi

machine=$(uname -m 2>/dev/null) || machine=
case $machine in
x86_64 | amd64) arch=amd64 ;;
aarch64 | arm64) arch=arm64 ;;
*) refuse "there is no Bonsai build for this machine (${machine:-uname -m names none}); builds are amd64 and arm64" "build it from source (go build ./cmd/bonsai) and copy it by hand, or use an amd64 or arm64 machine" ;;
esac

for tool in curl sha256sum tar gh; do
	if ! command -v "$tool" >/dev/null 2>&1; then
		case $tool in
		gh) refuse "gh, GitHub's command line, is not on the PATH; the provenance check needs it" "install gh 2.97 or newer (https://cli.github.com), run gh auth login, then run this again" ;;
		*) refuse "$tool is not on the PATH" "install $tool with your system's package manager, then run this again" ;;
		esac
	fi
done

# gh 2.97 or newer: --signer-workflow came in 2.51, --source-ref in 2.68, and 2.97 fixed GHSA-mm27-mwq9-fr5g.
ghv=$(gh --version 2>/dev/null | awk 'NR == 1 && $1 == "gh" && $2 == "version" { print $3 }') || ghv=
if ! printf '%s\n' "$ghv" | awk -F. '/^[0-9]+\.[0-9]+(\.[0-9]+)?/ { ok = ($1 > 2 || ($1 == 2 && $2 + 0 >= 97)) } END { exit !ok }'; then
	refuse "gh is ${ghv:-of a version gh --version does not name}, older than 2.97, whose provenance check a lookalike workflow name could pass" \
		"update gh to 2.97 or newer (https://cli.github.com), then run this again"
fi
if ! gh auth status >/dev/null 2>&1; then
	refuse "gh is not logged in to GitHub; the provenance check needs it" "run gh auth login, then run this again"
fi

if [ $# -eq 0 ]; then
	latest=$(gh api "repos/$repo/releases/latest" --jq .tag_name 2>/dev/null) || latest=
	ver=$(version "$latest")
	if [ -z "$ver" ]; then
		refuse "could not read the newest release's version from GitHub (gh api gave '${latest}')" \
			"check gh can reach GitHub (gh api repos/$repo/releases/latest), or run it with a version, as: sh get-bonsai.sh 1.0.3"
	fi
fi
tag=v$ver
archive=bonsai_${ver}_linux_${arch}.tar.gz
base=https://github.com/$repo/releases/download/$tag

dir=$(mktemp -d) || refuse "could not make a temporary folder" "check your temporary folder (\$TMPDIR, else /tmp) can be written, then run this again"
cd "$dir"
echo "Bonsai $ver for linux_$arch from $repo, in $dir"

echo "Downloading $archive and checksums.txt:"
curl -fsSLO "$base/$archive" || refuse "downloading $base/$archive failed; nothing was installed" \
	"check the version is on https://github.com/$repo/releases and has a linux_$arch archive, then run this again"
curl -fsSLO "$base/checksums.txt" || refuse "downloading $base/checksums.txt failed; nothing was installed" \
	"check your connection to github.com, then run this again"

echo "Checking the archive's SHA-256:"
awk -v name="$archive" '$2 == name || $2 == "*" name { n++; line = $0 } END { if (n == 1) print line; exit n != 1 }' \
	checksums.txt >archive.sha256 ||
	refuse "checksums.txt does not hold exactly one line for $archive; nothing was installed" "report it at https://github.com/$repo/issues; do not install this archive"
sha256sum -c archive.sha256 || refuse "$archive's SHA-256 is not the one checksums.txt holds; nothing was installed" \
	"run this again for a new download; if it fails again, report it at https://github.com/$repo/issues"

echo "Checking where it was built:"
gh attestation verify "$archive" -R "$repo" --signer-workflow "$workflow" --source-ref "refs/tags/$tag" \
	--deny-self-hosted-runners || refuse "gh attestation verify did not pass for $archive; nothing was installed" \
	"read gh's message above; if gh can reach GitHub and it fails again, report it at https://github.com/$repo/issues"
echo "$archive: built by $workflow from refs/tags/$tag"

echo "Every check passed: unpacking $archive and running its install.sh"
tar -xzf "$archive" || refuse "unpacking $archive failed; nothing was installed" "check $dir has room, then run this again"
[ -f install.sh ] || refuse "$archive holds no install.sh; nothing was installed" "report it at https://github.com/$repo/issues"
sh ./install.sh || refuse "the release's install.sh stopped (its message above says what it did)" \
	"follow its next step, or read docs/install.md, \"When a line fails\""
