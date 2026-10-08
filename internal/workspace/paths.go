package workspace

// Project-relative paths, as Bonsai stores and prints them (CLAUDE.md's Windows rules: forward slashes everywhere),
// and small text helpers that keep every message ASCII.

import (
	"errors"
	"strconv"
	"strings"
	"unicode/utf8"
)

// windowsDevices are names Windows keeps for devices in every folder, with or without an extension, in lower case:
// the console's own CONIN$ and CONOUT$, and COM and LPT with a superscript 1, 2 or 3 among them.
var windowsDevices = map[string]bool{"con": true, "prn": true, "aux": true, "nul": true, "conin$": true,
	"conout$": true, "com1": true, "com2": true, "com3": true, "com4": true, "com5": true, "com6": true, "com7": true,
	"com8": true, "com9": true, "lpt1": true, "lpt2": true, "lpt3": true, "lpt4": true, "lpt5": true, "lpt6": true,
	"lpt7": true, "lpt8": true, "lpt9": true,
	"com\u00b9": true, "com\u00b2": true, "com\u00b3": true, "lpt\u00b9": true, "lpt\u00b2": true, "lpt\u00b3": true}

// CheckRelPath checks a path Bonsai writes or records inside a project (a pack's file, a lock entry): relative,
// forward slashes, inside the project, and a path that Linux and Windows both hold as the same file. It refuses an
// empty path; a leading /; a backslash; a drive or stream colon; an empty, . or .. segment (so no path leaves the
// project); a .git segment or its NTFS short name GIT~1 (git's own folder, where a file would run as a hook; git
// refuses both); a control character; any of < > " | ? * (Windows cannot hold them); a segment ending in a dot or a
// space (Windows drops them); and a Windows device name (CON, NUL, CONIN$, COM1, COM with a superscript digit and
// the rest), with or without an extension, spaces before the extension included (con .txt).
func CheckRelPath(p string) error {
	switch {
	case p == "":
		return errors.New("an empty path")
	case !utf8.ValidString(p):
		return pathError(p, "is not valid UTF-8")
	case strings.HasPrefix(p, "/"):
		return pathError(p, "starts with /: it must be relative to the project")
	case strings.Contains(p, `\`):
		return pathError(p, `holds a backslash: write forward slashes`)
	case strings.Contains(p, ":"):
		return pathError(p, "holds a colon (a drive or a Windows stream)")
	case strings.ContainsAny(p, `<>"|?*`):
		return pathError(p, `holds one of < > " | ? *, which Windows cannot hold`)
	}
	for _, r := range p {
		if r < 0x20 || r == 0x7f {
			return pathError(p, "holds a control character")
		}
	}
	for _, seg := range strings.Split(p, "/") {
		lower := strings.ToLower(seg)
		base := lower
		if i := strings.IndexByte(base, '.'); i >= 0 {
			base = base[:i]
		}
		base = strings.TrimRight(base, " ") // Windows drops the spaces before an extension too: "con .txt" is CON
		switch {
		case seg == "":
			return pathError(p, "has an empty segment (// or a trailing /)")
		case seg == "." || seg == "..":
			return pathError(p, "has a . or .. segment: it must stay inside the project")
		case lower == ".git" || lower == "git~1":
			return pathError(p, "is inside a .git folder (or GIT~1, its NTFS short name)")
		case strings.HasSuffix(seg, ".") || strings.HasSuffix(seg, " "):
			return pathError(p, "has a segment ending in a dot or a space, which Windows drops")
		case windowsDevices[base]:
			return pathError(p, "names a Windows device ("+strings.ToUpper(base)+")")
		}
	}
	return nil
}

func pathError(p, why string) error {
	return errors.New("the path " + strconv.QuoteToASCII(p) + " " + why)
}

func itoa(n int) string { return strconv.Itoa(n) }

// oneLine gives an error's text on one ASCII line.
func oneLine(err error) string {
	return asciiOnly(strings.ReplaceAll(err.Error(), "\n", " "))
}

// asciiOnly keeps printable ASCII and writes every other character as a Go escape (backslash, u, four hex digits),
// so a message prints unbroken in PowerShell 5.1 (spec §3: human output is ASCII).
func asciiOnly(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= 0x20 && r < 0x7f {
			b.WriteRune(r)
			continue
		}
		q := strconv.QuoteRuneToASCII(r)
		b.WriteString(q[1 : len(q)-1])
	}
	return b.String()
}
