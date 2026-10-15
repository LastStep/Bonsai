package engine

// The form of Bonsai's own hook lines (design/plan-5.md, 5.3.6 note 2, under the director's answer (a) to step 5.3's first
// question, 9 Oct; spec §3 and §7's 8 Oct note): every line calls the installed bonsai at a fixed place no file an
// agent edits can redirect, and never reads the PATH. The guard's line, as the engine writes it:
//
//	b="/c/Program Files/Bonsai/bonsai.exe"; [ -x "$b" ] || b=/usr/local/bin/bonsai; [ -x "$b" ] || { echo "bonsai
//	guard: no bonsai is installed at C:/Program Files/Bonsai/bonsai.exe or /usr/local/bin/bonsai, so this call is
//	refused; next: tell the person: a person installs Bonsai there" >&2; exit 2; }; "$b" hook guard || exit 2
//
// (one line). The stop gate's line is the same with its own words; the start line's, the recorder's and the guard's
// second line's (step 5.5.11, where the lock declares roles: `hook guard --second`) end `[ -x "$b" ] || exit 0; "$b"
// hook <word>`, with no `|| exit 2`, so they print nothing and never block when Bonsai is not there. `[`, `echo` and `exit` are built into sh and Git Bash, so a line starts no process but bonsai.
//
// Two things measured before this form, and how it meets them:
//   - Claude Code reads a Stop hook's exit 2 whose stderr says a file is missing ("No such file or directory") as a
//     hook that failed to start, and lets the session end (step 5.3.4; its hooks reference). So the [ -x ] tests run
//     before bonsai is started, and the shell never prints its own words for a file that is not there; the line's own
//     words for it use none of them (no "not found", "missing", "No such file", "cannot find": a test holds that).
//   - A hook that fails to start, or a line run in PowerShell, does not block the call; with "onFailure": "block"
//     (Claude Code 2.1.295 and later) it does, so the guard's line carries it (GuardOnFailure, settings.go).
//
// Test builds (note 2): a build made with, P standing for github.com/LastStep/Bonsai/internal/engine,
//
//	go build -ldflags "-X P.hookPlace=<a place> -X P.scratchRoot=<a folder>" -o <a scratch folder>/bonsai ./cmd/bonsai
//
// writes that one place alone in its lines (b=<the place>; ...), so a scratch session runs the build it was linked
// by, and refuses init, update, unlink and the ladder in any folder outside its scratch root (Confine), so a stamped
// line never reaches a real project and a test build never makes a proof there. A plain build has both stamps empty (a test holds that). bonsai --version names a stamped
// build's stamps. bonsai's own commands' refusal of an agent's --allow-exec and unlink (step 5.3.6 note 8,
// cmd/bonsai) is skipped by a stamped build: its scratch root confines it.

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// The test build's stamps (this file's comment), set only by the linker's -X: empty in every plain build.
var (
	hookPlace   string // the one place a test build's hook lines run bonsai from, as the hook's shell names it
	scratchRoot string // the folder a test build links, updates, unlinks and climbs the ladder in, and nowhere else
)

// TestBuild gives a test build's stamps: its hook lines' place and its scratch root; both "" in a plain build.
func TestBuild() (place, root string) { return hookPlace, scratchRoot }

// Stamped reports a test build: one stamp or both set.
func Stamped() bool { return hookPlace != "" || scratchRoot != "" }

// hookPlaces are the places this build's hook lines run bonsai from, in order: a test build's stamped place alone,
// else the two installed places.
func hookPlaces() []string {
	if hookPlace != "" {
		return []string{filepath.ToSlash(hookPlace)}
	}
	return []string{PlaceWindows, PlaceLinux}
}

// HookCommand gives the hook line for `bonsai hook <word>` in Bonsai's form (this file's comment), trying the places
// in order. A line that blocks (the guard's and the stop gate's) refuses, exit 2 with its words on stderr, when bonsai
// is at none of them, and turns any other ending of bonsai into exit 2; one that does not (start, record) exits 0 and
// prints nothing then. Each place is written bare when it holds only letters, digits and / . _ + -, else in double
// quotes; a place holds no quote, $, backquote, backslash or control character (validPlace).
func HookCommand(word string, blocks bool, places ...string) string {
	var b strings.Builder
	for i, p := range places {
		if i > 0 {
			b.WriteString(`[ -x "$b" ] || `)
		}
		b.WriteString("b=" + shellPlace(p) + "; ")
	}
	if !blocks {
		b.WriteString(`[ -x "$b" ] || exit 0; "$b" hook ` + word)
		return b.String()
	}
	b.WriteString(`[ -x "$b" ] || { echo "` + notInstalledWords(word, places) + `" >&2; exit 2; }; "$b" hook ` + word + " || exit 2")
	return b.String()
}

// notInstalledWords are a blocking line's words when bonsai is at none of its places: what is refused, and the
// person's step. They never say a file is missing in the words an operating system uses for one, which Claude Code
// reads as a hook that failed to start (this file's comment).
func notInstalledWords(word string, places []string) string {
	shown := make([]string, len(places))
	for i, p := range places {
		shown[i] = shownPlace(p)
	}
	at := strings.Join(shown, " or ")
	if len(shown) > 2 {
		at = strings.Join(shown[:len(shown)-1], ", ") + " or " + shown[len(shown)-1]
	}
	what := "so this call is refused"
	if word == "stop" {
		what = "so the stop gate cannot check this session, which may not end yet (Claude Code ends the turn after its cap of blocks)"
	}
	return "bonsai " + word + ": no bonsai is installed at " + at + ", " + what + "; next: tell the person: a person installs Bonsai there"
}

// shownPlace is a place as a person names it: a Git Bash drive path, /c/..., as C:/....
func shownPlace(p string) string {
	if len(p) >= 3 && p[0] == '/' && p[2] == '/' && isASCIILetter(p[1]) {
		return strings.ToUpper(p[1:2]) + ":" + p[2:]
	}
	return p
}

func isASCIILetter(c byte) bool { return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') }

// shellPlace writes a place for the line: bare, or in double quotes (HookCommand).
func shellPlace(p string) string {
	for i := 0; i < len(p); i++ {
		c := p[i]
		if !isASCIILetter(c) && !(c >= '0' && c <= '9') && !strings.ContainsRune("/._+-", rune(c)) {
			return `"` + p + `"`
		}
	}
	return p
}

// validPlace reports a place a line can name: absolute (/... or a drive, C:/...), printable ASCII, with no quote of
// either kind, $, backquote or backslash, so it reads the same bare or in double quotes and in the line's words.
func validPlace(p string) bool {
	if !strings.HasPrefix(p, "/") && !(len(p) >= 3 && isASCIILetter(p[0]) && p[1] == ':' && p[2] == '/') {
		return false
	}
	for i := 0; i < len(p); i++ {
		if c := p[i]; c < 0x20 || c > 0x7e || strings.ContainsRune("\"'$`\\", rune(c)) {
			return false
		}
	}
	return true
}

// validWord reports a hook word a line can name: lower-case letters, and at most one flag after them, a space, two
// dashes and lower-case letters (the guard's second line, SecondWord: "guard --second").
func validWord(w string) bool {
	word, flag, hasFlag := strings.Cut(w, " --")
	return letters(word) && (!hasFlag || letters(flag))
}

// letters reports a word of lower-case letters, at least one.
func letters(w string) bool {
	if w == "" {
		return false
	}
	for i := 0; i < len(w); i++ {
		if w[i] < 'a' || w[i] > 'z' {
			return false
		}
	}
	return true
}

// parseHookForm reads a line's places and word as HookCommand writes them: "b=<place>; ", then for each further
// place `[ -x "$b" ] || b=<place>; `, then the rest, which names `"$b" hook <word>`. It does not check the rest:
// isBonsaiHook writes the line again from what it read and compares.
func parseHookForm(c string) (places []string, word string, ok bool) {
	rest := c
	for i := 0; ; i++ {
		prefix := "b="
		if i > 0 {
			prefix = `[ -x "$b" ] || b=`
		}
		if !strings.HasPrefix(rest, prefix) {
			if i == 0 {
				return nil, "", false
			}
			break
		}
		rest = rest[len(prefix):]
		var p string
		if strings.HasPrefix(rest, `"`) {
			j := strings.IndexByte(rest[1:], '"')
			if j < 0 {
				return nil, "", false
			}
			p, rest = rest[1:1+j], rest[2+j:]
		} else {
			j := strings.IndexByte(rest, ';')
			if j <= 0 {
				return nil, "", false
			}
			p, rest = rest[:j], rest[j:]
		}
		if !strings.HasPrefix(rest, "; ") || !validPlace(p) || len(places) == 2 {
			return nil, "", false
		}
		rest = rest[2:]
		places = append(places, p)
	}
	const call = `"$b" hook `
	i := strings.Index(rest, call)
	if i < 0 {
		return nil, "", false
	}
	word = strings.TrimSuffix(rest[i+len(call):], " || exit 2")
	if !validWord(word) {
		return nil, "", false
	}
	return places, word, true
}

// stampProblem says what is wrong with a test build's stamps, "" when nothing is (or the build is plain): both must
// be set, the place one a line can name (validPlace), the scratch root an absolute folder.
func stampProblem() string {
	switch {
	case !Stamped():
		return ""
	case hookPlace == "" || scratchRoot == "":
		return "it was stamped with only one of hookPlace and scratchRoot, and a test build needs both"
	case !validPlace(filepath.ToSlash(hookPlace)):
		return "its stamped hookPlace " + ASCII(hookPlace) + " is not an absolute path a hook line can name (no quote, $, backquote or backslash)"
	case !filepath.IsAbs(scratchRoot):
		return "its stamped scratchRoot " + ASCII(scratchRoot) + " is not an absolute folder"
	}
	return ""
}

// Confine refuses a test build's init, update and unlink, and its ladder (bonsai ladder, internal/ladder: a test build
// never makes a proof in a real project, plan-5 5.4.2 note 10), in a folder outside its scratch root (this file's
// comment): each of dirs (the folder it runs in, the checkout's top and its main checkout) must lie inside it, links
// resolved. A plain build is never confined. The refusal is test-build, exit 4; its next step names the installed
// bonsai or a plain build, which the ladder's caller replaces with its own command.
func Confine(dirs ...string) error {
	if !Stamped() {
		return nil
	}
	next := "use the installed bonsai, or a plain build (go build -o <a scratch folder>/bonsai ./cmd/bonsai), for this folder"
	if why := stampProblem(); why != "" {
		return errorf("test-build", ExitState, "build it again with both stamps, or build it plain",
			"this bonsai is a test build, and %s, so it links, updates, unlinks and climbs the ladder nowhere", why)
	}
	root := realFolder(scratchRoot)
	for _, d := range dirs {
		if d == "" {
			continue
		}
		if p := realFolder(d); !insideFolder(root, p) {
			return errorf("test-build", ExitState, next, "this bonsai is a test build: its hook lines run %s, and it links, updates, "+
				"unlinks and climbs the ladder only inside its scratch folder %s, which %s is not in", ASCII(filepath.ToSlash(hookPlace)),
				ASCII(filepath.ToSlash(scratchRoot)), ASCII(filepath.ToSlash(d)))
		}
	}
	return nil
}

// realFolder is a folder's absolute path with its links resolved, as far as it exists.
func realFolder(p string) string {
	abs, err := filepath.Abs(p)
	if err != nil {
		abs = filepath.Clean(p)
	}
	if r, err := filepath.EvalSymlinks(abs); err == nil {
		return r
	}
	return abs
}

// insideFolder reports whether p is root or lies inside it; letter case is ignored on Windows.
func insideFolder(root, p string) bool {
	if runtime.GOOS == "windows" {
		root, p = strings.ToLower(root), strings.ToLower(p)
	}
	if p == root {
		return true
	}
	return strings.HasPrefix(p, strings.TrimSuffix(root, string(os.PathSeparator))+string(os.PathSeparator))
}

// HookPlaceWords say where this build's hook lines run bonsai from, and what they do while it is not there (plan-5
// 5.3.6 note 2; "Risk in the code, 5.3": the fail-closed design, which init's help and closing words name).
func HookPlaceWords() string {
	ps := hookPlaces()
	shown := make([]string, len(ps))
	for i, p := range ps {
		shown[i] = shownPlace(p)
	}
	where := "Bonsai's hook lines run the bonsai installed at " + strings.Join(shown, " (Windows) or ") + " (WSL and Linux), never one on the PATH."
	if hookPlace != "" {
		where = "This test build's hook lines run the bonsai at " + shown[0] + ", never one on the PATH."
	}
	return wrapText(where+" While it is not there, the guard's line refuses every edit and shell command, and the stop "+
		"gate's refuses each session's end until Claude Code ends the turn after its cap of blocks (eight in a row); "+
		"the start and recorder lines do nothing.", 116)
}

// wrapText wraps words at width, each line ending in a line feed.
func wrapText(text string, width int) string {
	var b strings.Builder
	line := ""
	for _, w := range strings.Fields(text) {
		if line != "" && len(line)+1+len(w) > width {
			b.WriteString(line + "\n")
			line = ""
		}
		if line == "" {
			line = w
		} else {
			line += " " + w
		}
	}
	if line != "" {
		b.WriteString(line + "\n")
	}
	return b.String()
}

// ShownLine gives a hook line for a help page, whose lines keep under 120 characters: broken after each "; " that
// is outside quotes and braces, each part on a line of its own after two spaces, and a part still too long broken at
// its spaces, its further lines after four. The settings file holds it as one line, which the help says.
func ShownLine(command string) string {
	var parts []string
	quote, depth, start := false, 0, 0
	for i := 0; i < len(command); i++ {
		switch c := command[i]; {
		case c == '"':
			quote = !quote
		case quote:
		case c == '{':
			depth++
		case c == '}':
			depth--
		case c == ';' && depth == 0 && i+1 < len(command) && command[i+1] == ' ':
			parts = append(parts, command[start:i+1])
			start = i + 2
		}
	}
	parts = append(parts, command[start:])
	var b strings.Builder
	for _, p := range parts {
		if len(p)+2 <= 116 {
			b.WriteString("  " + p + "\n")
			continue
		}
		line := "  "
		for _, w := range strings.Split(p, " ") {
			if strings.TrimSpace(line) != "" && len(line)+1+len(w) > 116 {
				b.WriteString(line + "\n")
				line = "    "
			}
			if strings.TrimSpace(line) == "" {
				line += w
			} else {
				line += " " + w
			}
		}
		b.WriteString(line + "\n")
	}
	return b.String()
}
