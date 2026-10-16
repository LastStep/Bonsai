#!/bin/sh
# install.sh: installs the bonsai beside this script as this machine's Bonsai, on Linux and WSL.
#
# What it does: checks that the bonsai in this script's folder runs (bonsai --version) and takes its SHA-256; copies it
# under a temporary name (.bonsai.new.<pid>) into the target's folder, checks the copy's SHA-256 against the source's,
# and renames it over the target in one step, so a hook starting meanwhile finds the old copy or the new one, never no
# file, and a bonsai already running keeps its old file. At the real place, /usr/local/bin/bonsai, the copy and the
# rename run through sudo (the password once), and the file and its folder are then checked to be root's and writable
# by no one else. Last it writes the installed copy's record, install.json, in the Bonsai home
# (${BONSAI_HOME:-$HOME/.bonsai}): its path, version and sha256, the record bonsai check and bonsai status compare the
# bonsai on the PATH with. It prints the fingerprint, every bonsai on the PATH in order (naming any that comes before
# the installed one) and the installed --version.
#
# When to run it: once per Bonsai release, by the person, as themselves, from the release's folder. Never under sudo:
# it refuses to run as root, since install.json would then land in root's home. No agent runs it at the real place.
#
# Options:
#   --target <file>  install to <file> instead of /usr/local/bin/bonsai, an absolute path in a folder that exists and
#                    that no group or other user can write; a folder you own needs no sudo. Example:
#                      sh install.sh --target "$HOME/scratch/bin/bonsai"
#   --remove         take the installed file out (with sudo at the real place), a temporary copy left by an install
#                    that stopped halfway, and install.json. Example:
#                      sh install.sh --remove
#
# Exit 0 when it is done; otherwise 1, with one plain sentence and its next step.
#
# The way out: sh install.sh --remove; or by hand, sudo rm /usr/local/bin/bonsai and rm ~/.bonsai/install.json.

set -u

place=/usr/local/bin/bonsai

refuse() {
	printf 'install.sh: %s\nnext: %s\n' "$1" "$2" >&2
	exit 1
}

# sha256 prints a file's SHA-256 in lowercase hex.
sha256() {
	if command -v sha256sum >/dev/null 2>&1; then
		sum=$(sha256sum -- "$1" 2>/dev/null) || return 1
	else
		sum=$(shasum -a 256 -- "$1" 2>/dev/null) || return 1
	fi
	printf '%s\n' "${sum%% *}" | tr 'A-F' 'a-f'
}

# mode prints a file's or folder's mode as ls shows it (drwxr-xr-x), owner prints its owner's numeric id.
mode() { ls -ldn -- "$1" 2>/dev/null | cut -c1-10; }
owner() { ls -ldn -- "$1" 2>/dev/null | awk '{print $3}'; }

# others_write says whether a group or other users can write the file or folder.
others_write() {
	m=$(mode "$1")
	[ "$(printf '%s' "$m" | cut -c6)" = w ] || [ "$(printf '%s' "$m" | cut -c9)" = w ]
}

# json escapes a string for a JSON string's inside.
json() { printf '%s' "$1" | sed 's/\\/\\\\/g; s/"/\\"/g'; }

target=$place
remove=no
while [ $# -gt 0 ]; do
	case $1 in
	--target)
		[ $# -ge 2 ] || refuse "--target needs a file after it" "run it again with the file, as: sh install.sh --target /usr/local/bin/bonsai"
		target=$2
		shift 2
		;;
	--remove)
		remove=yes
		shift
		;;
	*)
		refuse "$1 is not an option of install.sh (it takes --target <file> and --remove)" "run: sh install.sh, or read the options at the top of install.sh"
		;;
	esac
done

if [ "$(id -u)" = 0 ]; then
	refuse "install.sh runs as root (under sudo, or as the root user), so install.json would land in root's home, not yours" \
		"run it again as yourself, without sudo: sh install.sh (it asks for sudo's password itself where it needs it)"
fi

case $target in
/*) ;;
*) refuse "the target $target is not an absolute path" "run it again with a path starting with /, as: --target /usr/local/bin/bonsai" ;;
esac
case $target in
*[!\ -~]* | */) refuse "the target must be a file's path in plain ASCII letters" "run it again with another target, as: --target /usr/local/bin/bonsai" ;;
esac
dir=${target%/*}
[ -n "$dir" ] || dir=/
real=no
[ "$target" = "$place" ] && real=yes
sudo=
[ -w "$dir" ] || sudo=sudo
home=${BONSAI_HOME:-$HOME/.bonsai}
record=$home/install.json

if [ $remove = yes ]; then
	gone=
	if [ -e "$target" ] || [ -L "$target" ]; then
		$sudo rm -f -- "$target" || refuse "could not remove $target" "remove it by hand: ${sudo:+sudo }rm $target"
		gone="$gone $target"
	fi
	for f in "$dir"/.bonsai.new.*; do
		[ -e "$f" ] || continue
		$sudo rm -f -- "$f" || refuse "could not remove the leftover copy $f" "remove it by hand: ${sudo:+sudo }rm $f"
		gone="$gone $f"
	done
	if [ -e "$record" ]; then
		rm -f -- "$record" || refuse "could not remove $record" "remove it by hand: rm $record"
		gone="$gone $record"
	fi
	if [ -z "$gone" ]; then
		echo "Nothing to remove: no $target, no leftover copy and no $record."
	else
		echo "Removed:$gone."
	fi
	exit 0
fi

here=$(CDPATH='' cd -- "$(dirname -- "$0")" && pwd -P) || refuse "could not find the folder this script is in" "run it from the release's folder: cd into it, then run: sh install.sh"
src=$here/bonsai
[ -f "$src" ] || refuse "there is no bonsai beside this script ($src)" "run the install.sh that came in the release's folder, beside its bonsai"
out=$("$src" --version 2>/dev/null) || refuse "the bonsai beside this script ($src) does not run (bonsai --version failed)" \
	"download the release for this machine again (Linux, amd64 or arm64) and run its install.sh"
case $out in
"bonsai "*) ;;
*) refuse "the file beside this script ($src) does not answer --version as bonsai does" "download Bonsai's release again and run its install.sh" ;;
esac
want=$(sha256 "$src") || refuse "could not take the SHA-256 of $src" "check that sha256sum runs, then run it again"

[ -d "$dir" ] || refuse "the target's folder $dir does not exist" "make it first (mkdir -p $dir), or choose another target"
if others_write "$dir"; then
	refuse "the target's folder $dir can be written by a group or by other users, so anyone there could replace bonsai" \
		"choose a folder only you or root can write, or take the write right away: ${sudo:+sudo }chmod go-w $dir"
fi
if [ $real = yes ] && [ "$(owner "$dir")" != 0 ]; then
	refuse "the real place's folder $dir is not root's" "a person makes it root's (sudo chown root:root $dir), then run it again"
fi

tmp=$dir/.bonsai.new.$$
if [ -n "$sudo" ]; then
	sudo install -o root -g root -m 0755 "$src" "$tmp" || refuse "the copy into $dir failed (sudo install)" "check sudo works for you (sudo -v), then run it again"
else
	install -m 0755 "$src" "$tmp" || refuse "the copy into $dir failed (install)" "check you can write $dir, then run it again"
fi
got=$(sha256 "$tmp") || got=
if [ "$got" != "$want" ]; then
	$sudo rm -f -- "$tmp"
	refuse "the copy in $dir is not the file that was checked (its SHA-256 differs: the source changed between the check and the copy), so it was removed and nothing was installed" \
		"download Bonsai's release again into a folder only you can write, then run its install.sh"
fi
$sudo mv -f -- "$tmp" "$target" || {
	$sudo rm -f -- "$tmp"
	refuse "the rename over $target failed, so nothing was installed" "check $target is not a folder, then run it again"
}

if [ $real = yes ]; then
	if [ "$(owner "$target")" != 0 ] || others_write "$target" || [ "$(owner "$dir")" != 0 ] || others_write "$dir"; then
		refuse "$target or its folder is not root's alone after the install (owner or write rights)" \
			"a person makes both root's and writable by root only (sudo chown root:root and sudo chmod go-w), then run it again"
	fi
fi

line=$("$target" --version 2>/dev/null | head -n 1) || refuse "the installed $target does not run" "run it again; if it fails again, run: sh install.sh --remove"
version=${line#bonsai }
version=${version%% *}
sum=$(sha256 "$target") || sum=
[ "$sum" = "$want" ] || refuse "the installed $target is not the file that was checked (its SHA-256 differs)" "run it again"

mkdir -p -- "$home" || refuse "could not make the Bonsai home $home" "check you can write $home's folder, then run it again"
part=$record.new.$$
printf '{"path": "%s", "version": "%s", "sha256": "%s"}\n' "$(json "$target")" "$(json "$version")" "$sum" >"$part" &&
	mv -f -- "$part" "$record" || {
	rm -f -- "$part"
	refuse "could not write $record" "check you can write $home, then run it again"
}

echo "Installed $target, bonsai $version."
echo "SHA-256: $sum"
echo "The record: $record"
echo "Every bonsai on the PATH, in order (which -a bonsai):"
first=
oldifs=$IFS
IFS=:
for p in $PATH; do
	[ -n "$p" ] || p=.
	if [ -f "$p/bonsai" ] && [ -x "$p/bonsai" ]; then
		echo "  $p/bonsai"
		[ -n "$first" ] || first=$p/bonsai
	fi
done
IFS=$oldifs
if [ -z "$first" ]; then
	echo "  none: $dir is not on the PATH; add it, or hooks and commands that call bonsai by name will not find it"
elif [ "$first" != "$target" ]; then
	echo "  $first comes first on the PATH, so a command that calls bonsai by name runs it, not $target: take it off the PATH"
fi
echo "$line"
exit 0
