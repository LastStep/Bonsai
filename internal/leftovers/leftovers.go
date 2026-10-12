// Package leftovers is the ladder's read-only end-of-run line (spec §9; plan 5.4.1 note 4): the processes still
// running from a checkout that the runner did not start, each by its pid, its program's name and the head of its
// command (internal/redact's CommandHead, never a whole command line, which may hold a secret). The runner, its
// ancestors and its descendants are left out. It never ends anything and never changes a colour.
//
//	v := leftovers.Check(root) // at a run's end, from the top-level runner only
//	result.Leftovers = v.Record()
//	for _, l := range v.Lines() { fmt.Println("  " + l) }
//
// What ties a process to the checkout:
//   - Linux, from /proc: its working folder, its program, or an argument (a whole argument, or the part after = in
//     --x=<path>) at or below the checkout. A process of another user shows only its arguments.
//   - Windows, from a process snapshot (CreateToolhelp32Snapshot): its image's path, or a whole argument of its
//     command line (NtQueryInformationProcess, which the system answers cheaply for a process this user may query),
//     read as a Windows path (drive letters, \\?\, file:///, MSYS's /c/...), in any case. Windows gives no cheap
//     way to read another process's working folder, so that is not read there.
//   - Elsewhere: not checked, and the value says so.
//
// The value is one of three, as the ladder result's leftovers (formats/schemas/ladder.schema.json):
// {checkout: [{pid, name, head}], ms}, {unknown: <why>, ms}, or "not checked (<platform>)". A run that does not
// check (a --ci run, a nested ladder: the runner's call, spec §9) keeps the zero Value, whose record is null.
//
// Known limits: a process tied to the checkout only by a relative argument and a working folder this package cannot
// read (another user's, or any on Windows) is not seen; one whose program and arguments name nothing in the
// checkout and that works elsewhere is not seen. The line also names processes that are rightly running from the
// checkout but were not started by this run: another session's shell or its tool servers, an editor's helpers. It
// says so: read-only, stop only what you started.
//
// The package also holds the readers rungjob uses to list a rung's processes (proc_linux.go, proc_windows.go), so
// the system's process tables are read in one place.
package leftovers

import (
	"errors"
	"fmt"
	"path"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/LastStep/Bonsai/internal/redact"
	"github.com/LastStep/Bonsai/internal/schema"
)

// Process is one process as the ladder reports it: its pid, its program's name, and the head of its command.
type Process struct {
	PID  int
	Name string // the program's name: Linux the kernel's short name (comm), Windows the image's file name
	Head string // the command's head (redact.CommandHead), "" when nothing in it is safe to keep
}

// Kind says which of the leftovers values a Value is.
type Kind int

// The kinds of Value. The zero Value is Skipped.
const (
	// Skipped: the check was not run (a --ci run, a nested ladder). The result's leftovers is null; no line.
	Skipped Kind = iota
	// Found: the check ran; Checkout holds what it found, perhaps nothing.
	Found
	// Unknown: the check could not read the processes; Why says why.
	Unknown
	// NotChecked: this platform has no reader; Why names the platform.
	NotChecked
)

// Value is what the check found.
type Value struct {
	Kind     Kind
	Checkout []Process // Found: the processes, oldest first
	Why      string    // Unknown: why the check could not tell; NotChecked: the platform
	MS       int64     // how long the check took, in milliseconds (Found, Unknown)
}

// RowsPrinted is how many processes the line prints; the record keeps them all.
const RowsPrinted = 10

// errNotChecked is a snapshot's answer on a system with no reader here.
var errNotChecked = errors.New("not checked")

// evalSymlinks finds a root's real path (a variable so a test can stand in its own).
var evalSymlinks = filepath.EvalSymlinks

// platform names this system for "not checked (<platform>)".
func platform() string { return runtime.GOOS }

// Check runs the check for the checkout at root (an absolute folder), from this process, which counts as the
// runner. It reads the system's process table once and never acts on a process.
func Check(root string) Value {
	start := time.Now()
	rows, style, err := snapshot()
	ms := time.Since(start).Milliseconds()
	switch {
	case errors.Is(err, errNotChecked):
		return Value{Kind: NotChecked, Why: platform()}
	case err != nil:
		return Value{Kind: Unknown, Why: oneLine(err.Error()), MS: ms}
	}
	found := find(rows, roots(root), selfPID(), style)
	return Value{Kind: Found, Checkout: found, MS: time.Since(start).Milliseconds()}
}

// Record gives the value as the ladder result's leftovers holds it: a schema.Object, a string, or nil (Skipped).
// Its field order is fixed, so the bytes written are the same on every run.
func (v Value) Record() any {
	switch v.Kind {
	case Found:
		list := make([]any, 0, len(v.Checkout))
		for _, p := range v.Checkout {
			var head any
			if p.Head != "" {
				head = p.Head
			}
			list = append(list, schema.Object{
				{Key: "pid", Value: int64(p.PID)},
				{Key: "name", Value: strings.ToValidUTF8(p.Name, "?")},
				{Key: "head", Value: head},
			})
		}
		return schema.Object{{Key: "checkout", Value: list}, {Key: "ms", Value: v.MS}}
	case Unknown:
		return schema.Object{{Key: "unknown", Value: strings.ToValidUTF8(v.Why, "?")}, {Key: "ms", Value: v.MS}}
	case NotChecked:
		return "not checked (" + v.Why + ")"
	}
	return nil
}

// Lines gives the human line, ASCII, with no indent of its own: nil when Skipped.
func (v Value) Lines() []string {
	switch v.Kind {
	case Found:
		if len(v.Checkout) == 0 {
			return []string{"leftovers not started by this run: none"}
		}
		out := []string{fmt.Sprintf("leftovers not started by this run: %d (read-only; stop only what you started)",
			len(v.Checkout))}
		for i, p := range v.Checkout {
			if i == RowsPrinted {
				out = append(out, fmt.Sprintf("  ... and %d more (all of them are in the result file)",
					len(v.Checkout)-RowsPrinted))
				break
			}
			head := p.Head
			if head == "" {
				head = "-"
			}
			out = append(out, fmt.Sprintf("  %d  %s  %s", p.PID, ascii(p.Name), ascii(head)))
		}
		return out
	case Unknown:
		return []string{"leftovers: unknown (" + ascii(v.Why) + ")"}
	case NotChecked:
		return []string{"leftovers: not checked (" + v.Why + ")"}
	}
	return nil
}

// Row is one process as a snapshot reads it.
type Row struct {
	PID, PPID int
	Name      string   // the program's name
	Start     uint64   // when it started, in the system's own unit (Linux clock ticks, Windows FILETIME); compared only
	Exe       string   // the program's path, or ""
	Cwd       string   // its working folder, or "" (Linux only)
	Argv      []string // its arguments, or nil when the system did not give them
}

// style says how a snapshot's paths read: POSIX paths in their case, or Windows paths in any case.
type style int

const (
	posixStyle style = iota
	windowsStyle
)

// find is the check itself, on a snapshot: the rows tied to one of the roots (each already in its style's form,
// from roots), less self, its ancestors and its descendants, oldest first. Pure, so its tests run on every system.
func find(rows []Row, rootList []string, self int, st style) []Process {
	byPID := make(map[int]*Row, len(rows))
	for i := range rows {
		byPID[rows[i].PID] = &rows[i]
	}
	// A parent link counts only when the parent is not younger than the child: a pid taken by a later process is
	// no one's parent.
	parentOf := func(r *Row) *Row {
		p := byPID[r.PPID]
		if p == nil || p.PID == r.PID || p.Start > r.Start {
			return nil
		}
		return p
	}
	children := map[int][]int{}
	for i := range rows {
		if p := parentOf(&rows[i]); p != nil {
			children[p.PID] = append(children[p.PID], rows[i].PID)
		}
	}
	out := map[int]bool{}
	if me := byPID[self]; me != nil {
		stack := []int{self}
		for len(stack) > 0 {
			x := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			if out[x] {
				continue
			}
			out[x] = true
			stack = append(stack, children[x]...)
		}
		// The ancestors, each alone: never their other descendants, where an agent's own processes live.
		for a, n := parentOf(me), 0; a != nil && n < 256; a, n = parentOf(a), n+1 {
			out[a.PID] = true
		}
	}
	var found []*Row
	for i := range rows {
		r := &rows[i]
		if out[r.PID] || len(rootList) == 0 {
			continue
		}
		if tied(r, rootList, st) {
			found = append(found, r)
		}
	}
	sort.Slice(found, func(i, j int) bool {
		if found[i].Start != found[j].Start {
			return found[i].Start < found[j].Start
		}
		return found[i].PID < found[j].PID
	})
	list := make([]Process, 0, len(found))
	for _, r := range found {
		list = append(list, Process{PID: r.PID, Name: r.Name, Head: HeadOf(r.Argv)})
	}
	return list
}

// tied reports whether a row's working folder, program or an argument lies at or below one of the roots.
func tied(r *Row, rootList []string, st style) bool {
	cands := []string{r.Exe, r.Cwd}
	for _, a := range r.Argv {
		cands = append(cands, a)
		if strings.HasPrefix(a, "-") {
			if _, v, ok := strings.Cut(a, "="); ok {
				cands = append(cands, v)
			}
		}
	}
	for _, c := range cands {
		p := asPath(c, st)
		if p == "" {
			continue
		}
		for _, root := range rootList {
			if p == root || strings.HasPrefix(p, root+"/") {
				return true
			}
		}
	}
	return false
}

// asPath reads a token as an absolute path in its style's form (forward slashes, cleaned; Windows lower-case), or
// "" when the whole token is not one.
func asPath(token string, st style) string {
	t := strings.TrimSpace(token)
	if t == "" {
		return ""
	}
	if st == posixStyle {
		if strings.HasPrefix(t, "file://") && strings.HasPrefix(t[len("file://"):], "/") {
			t = t[len("file://"):]
		}
		if !strings.HasPrefix(t, "/") {
			return ""
		}
		return path.Clean(t)
	}
	t = strings.ReplaceAll(t, `\`, "/")
	low := strings.ToLower(t)
	switch {
	case strings.HasPrefix(low, "file:///") && driveAt(t, 8):
		t = t[8:]
	case strings.HasPrefix(low, "file://") && driveAt(t, 7):
		t = t[7:]
	case strings.HasPrefix(t, "//?/") && driveAt(t, 4):
		t = t[4:]
	case len(t) >= 2 && t[0] == '/' && isLetter(t[1]) && (len(t) == 2 || t[2] == '/'):
		t = t[1:2] + ":" + t[2:] // MSYS: /c/Users/... is C:/Users/...
		if len(t) == 2 {
			t += "/"
		}
	}
	switch {
	case driveAt(t, 0):
	case strings.HasPrefix(t, "//") && len(t) > 2 && t[2] != '/' && t[2] != '?':
		return "/" + path.Clean(strings.ToLower(t[1:])) // a share: //server/share/...
	default:
		return ""
	}
	return path.Clean(strings.ToLower(t))
}

// driveAt reports a drive-letter path, X:/, starting at i.
func driveAt(t string, i int) bool {
	return len(t) >= i+3 && isLetter(t[i]) && t[i+1] == ':' && t[i+2] == '/'
}

func isLetter(c byte) bool { return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' }

// roots gives the checkout's root in this system's style, and its real path when that differs (a link in the way).
func roots(root string) []string {
	st := nativeStyle()
	var out []string
	add := func(p string) {
		if q := asPath(p, st); q != "" && q != "/" && !(st == windowsStyle && len(q) == 3 && q[1] == ':') {
			for _, have := range out {
				if have == q {
					return
				}
			}
			out = append(out, q)
		}
	}
	add(root)
	if real, err := evalSymlinks(root); err == nil {
		add(real)
	}
	return out
}

// HeadOf gives a process's command head from its arguments: the program's base name and, for the programs
// redact.CommandHead knows, one subcommand word. Never more: "" when nothing is safe to keep.
func HeadOf(argv []string) string {
	if len(argv) == 0 {
		return ""
	}
	var b strings.Builder
	for i, a := range argv {
		if i > 0 {
			b.WriteByte(' ')
		}
		if i == 0 {
			a = strings.ReplaceAll(a, `\`, "/") // a Windows program's path reads as its base name
		}
		b.WriteString(shellQuote(a))
	}
	return redact.CommandHead(b.String(), false)
}

// shellQuote quotes one argument for redact's shell reading when it holds anything that reading splits on or
// takes apart (a space, a quote, a backslash, # ; ( ) | &), so it stays one word; NAME=value stays as it is, so
// an assignment still reads as one.
func shellQuote(a string) string {
	if a == "" {
		return "''"
	}
	if strings.ContainsAny(a, " \t\n\r'\"\\#;()|&") {
		return "'" + strings.ReplaceAll(a, "'", `'\''`) + "'"
	}
	return a
}

// SplitCommandLine splits a Windows command line the way the Microsoft C runtime does (backslashes before a quote,
// "" inside quotes). It reads strings only, so it runs on every system.
func SplitCommandLine(line string) []string {
	var out []string
	var cur strings.Builder
	inQuotes, has := false, false
	for i := 0; i < len(line); {
		c := line[i]
		switch {
		case c == '\\':
			n := 0
			for i < len(line) && line[i] == '\\' {
				n++
				i++
			}
			if i < len(line) && line[i] == '"' {
				cur.WriteString(strings.Repeat(`\`, n/2))
				if n%2 == 1 {
					cur.WriteByte('"')
					i++
				}
			} else {
				cur.WriteString(strings.Repeat(`\`, n))
			}
			has = true
		case c == '"':
			if inQuotes && i+1 < len(line) && line[i+1] == '"' {
				cur.WriteByte('"')
				i += 2
				has = true
				continue
			}
			inQuotes = !inQuotes
			has = true
			i++
		case !inQuotes && (c == ' ' || c == '\t'):
			if has {
				out = append(out, cur.String())
				cur.Reset()
				has = false
			}
			i++
		default:
			cur.WriteByte(c)
			has = true
			i++
		}
	}
	if has {
		out = append(out, cur.String())
	}
	return out
}

// ascii keeps printable ASCII and writes ? for anything else, for the human line (PowerShell 5.1 garbles UTF-8).
func ascii(s string) string {
	var b strings.Builder
	for len(s) > 0 {
		r, n := utf8.DecodeRuneInString(s)
		if r >= 0x20 && r < 0x7f {
			b.WriteRune(r)
		} else {
			b.WriteByte('?')
		}
		s = s[n:]
	}
	return b.String()
}

// oneLine keeps an error's first line.
func oneLine(s string) string {
	s, _, _ = strings.Cut(s, "\n")
	return strings.TrimSpace(strings.ReplaceAll(s, `\`, "/"))
}
