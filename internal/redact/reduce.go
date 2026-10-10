package redact

// The second job: what a tool call is reduced to before it is written (the log's target and text, contract §8.1),
// the first line of defence, since what is never kept needs no redacting. A shell command is kept as its head only,
// so `TOKEN=... cmd`, `git -c http.extraHeader=...`, `ssh -i key` and `node -e "..."` never reach a record; a path
// inside the checkout is kept relative to it and any other is dropped (contract §2.6: records carry
// workspace-relative paths); a web fetch keeps its host, an MCP call `server.tool`, the Agent tool its subagent
// type, a skill its name, AskUserQuestion its questions. Every one then passes Text and its cap.
//
// The reductions read strings only (no file system), so they give the same on every system: a Windows path is read
// with its backslashes as slashes and, under a drive-letter root, in any case.

import (
	"net/url"
	"path"
	"strings"

	"github.com/LastStep/Bonsai/internal/schema"
)

// The caps of the log's two text fields, in characters (contract §8.1).
const (
	TargetCap = 200
	TextCap   = 300
)

// Target gives the one string a record keeps for a tool call, redacted and capped at TargetCap, or "" (null) when
// the tool keeps none. root is the checkout's absolute folder and cwd the session's, against which a relative path
// is read (root when cwd is empty or not absolute); either may use backslashes.
func Target(tool string, input schema.Object, root, cwd string) string {
	str := func(key string) string {
		if input == nil {
			return ""
		}
		v, _ := input.Get(key)
		s, _ := v.(string)
		return s
	}
	var t string
	switch tool {
	case "Read", "Edit", "Write", "MultiEdit":
		t = InsidePath(str("file_path"), root, cwd)
	case "NotebookEdit":
		p := str("notebook_path")
		if p == "" {
			p = str("file_path")
		}
		t = InsidePath(p, root, cwd)
	case "Grep", "Glob":
		t = InsidePath(str("path"), root, cwd)
	case "Bash":
		t = CommandHead(str("command"), false)
	case "PowerShell":
		t = CommandHead(str("command"), true)
	case "WebFetch":
		t = host(str("url"))
	case "Agent", "Task":
		t = str("subagent_type")
	case "Skill":
		t = str("skill")
	default:
		if rest, ok := strings.CutPrefix(tool, "mcp__"); ok {
			if i := strings.Index(rest[min(1, len(rest)):], "__"); i >= 0 {
				i += min(1, len(rest))
				if server, name := rest[:i], rest[i+2:]; name != "" {
					t = server + "." + name
				}
			}
		}
	}
	if t == "" {
		return ""
	}
	return Capped(t, TargetCap)
}

// Question gives AskUserQuestion's questions joined with " · ", redacted and capped at TextCap: the ask itself. It
// is "" when the input holds no question.
func Question(input schema.Object) string {
	if input == nil {
		return ""
	}
	v, _ := input.Get("questions")
	list, _ := v.([]any)
	var qs []string
	for _, item := range list {
		o, ok := item.(schema.Object)
		if !ok {
			continue
		}
		if q, _ := o.Get("question"); q != nil {
			if s, ok := q.(string); ok && s != "" {
				qs = append(qs, s)
			}
		}
	}
	if len(qs) == 0 {
		return ""
	}
	return Capped(strings.Join(qs, " · "), TextCap)
}

// host is a URL's host name, lower case, or "".
func host(raw string) string {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u.Scheme == "" {
		return ""
	}
	return strings.ToLower(u.Hostname())
}

// InsidePath gives p relative to root with forward slashes ("." for root itself) when it lies inside root, else ""
// (a path outside the checkout is not kept, contract §2.6). A relative p is read against cwd, or root when cwd is
// not absolute; a path starting with ~ is outside. Backslashes read as slashes, and under a drive-letter root
// (C:/...) case does not count, as on Windows.
func InsidePath(p, root, cwd string) string {
	if p == "" || root == "" {
		return ""
	}
	q := slashes(p)
	if strings.HasPrefix(q, "~") {
		return ""
	}
	r := clean(slashes(root))
	if !absolute(r) {
		return ""
	}
	if !absolute(q) {
		base := r
		if c := clean(slashes(cwd)); absolute(c) {
			base = c
		}
		q = base + "/" + q
	}
	q = clean(q)
	fold := drive(r)
	same := func(a, b string) bool {
		if fold {
			return strings.EqualFold(a, b)
		}
		return a == b
	}
	if same(q, r) {
		return "."
	}
	prefix := strings.TrimSuffix(r, "/") + "/"
	if len(q) > len(prefix) && same(q[:len(prefix)], prefix) {
		return q[len(prefix):]
	}
	return ""
}

func slashes(p string) string { return strings.ReplaceAll(p, `\`, "/") }

// drive reports a drive-letter path, C:/...
func drive(p string) bool {
	return len(p) >= 3 && isLetter(rune(p[0])) && p[1] == ':' && p[2] == '/'
}

// absolute reports a path from a root: /..., //server/share/..., or C:/...
func absolute(p string) bool { return strings.HasPrefix(p, "/") || drive(p) }

// clean is path.Clean that keeps a share's leading //, and gives "" for "".
func clean(p string) string {
	if p == "" {
		return ""
	}
	if strings.HasPrefix(p, "//") && !strings.HasPrefix(p, "///") {
		return "/" + path.Clean(p[1:])
	}
	return path.Clean(p)
}

// CommandHead gives a shell command's head, the only part of it a record keeps: past any NAME=value assignments,
// wrappers (sudo, env, time, nohup, exec, command, nice) and steps that only change folder, the program's base name,
// and for git, npm, dotnet, unity, claude, gh, go and bonsai one subcommand word ([a-z][a-z0-9-]*, never a flag,
// path, value or quoted text); in PowerShell a Verb-Noun cmdlet as written. It is "" when nothing is safe to keep.
func CommandHead(command string, powershell bool) string {
	if strings.TrimSpace(command) == "" {
		return ""
	}
	nav := ""
	for _, words := range segments(command, powershell) {
		h := segmentHead(words, powershell)
		if h.skip {
			if h.head != "" && nav == "" {
				nav = h.head
			}
			continue
		}
		return h.head
	}
	return nav
}

// shellWord is one word of a command line: its text with quotes taken off, whether any of it was quoted, and where
// its first quoted character is (-1 for none).
type shellWord struct {
	v      string
	quoted bool
	qAt    int
}

// segments splits a command line into its parts (at line ends, ; | || && & ( )), each a list of words. It is good
// enough to find a program and its first words, and is never used to run anything.
func segments(command string, powershell bool) [][]shellWord {
	var segs [][]shellWord
	var seg []shellWord
	var w *shellWord
	var b strings.Builder
	flush := func() {
		if w != nil {
			w.v = b.String()
			seg = append(seg, *w)
		}
		w = nil
		b.Reset()
	}
	end := func() {
		flush()
		if len(seg) > 0 {
			segs = append(segs, seg)
		}
		seg = nil
	}
	start := func(quoted bool) {
		if w == nil {
			w = &shellWord{qAt: -1}
		}
		if quoted && w.qAt < 0 {
			w.qAt = b.Len()
		}
		if quoted {
			w.quoted = true
		}
	}
	esc := byte('\\')
	if powershell {
		esc = '`'
	}
	s := command
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c == '\'':
			start(true)
			for i++; i < len(s) && s[i] != '\''; i++ {
				b.WriteByte(s[i])
			}
		case c == '"':
			start(true)
			for i++; i < len(s) && s[i] != '"'; i++ {
				if s[i] == esc && i+1 < len(s) && (powershell || strings.IndexByte("\"\\$`", s[i+1]) >= 0) {
					i++
				}
				b.WriteByte(s[i])
			}
		case c == esc && i+1 < len(s):
			start(true)
			i++
			b.WriteByte(s[i])
		case c == '#' && w == nil:
			for i < len(s) && s[i] != '\n' {
				i++
			}
			end()
		case c == '\n' || c == '\r' || c == ';' || c == '(' || c == ')':
			end()
		case c == '|' || c == '&' && (!powershell || i+1 < len(s) && s[i+1] == '&'):
			if i+1 < len(s) && s[i+1] == c {
				i++
			}
			end()
		case c == ' ' || c == '\t':
			flush()
		default:
			start(false)
			b.WriteByte(c)
		}
	}
	end()
	return segs
}

// subcommandPrograms are the programs whose head keeps one subcommand word, each with the flags that take a value
// before it (so the value is not read as the subcommand). This table is the list's one home.
var subcommandPrograms = map[string][]string{
	"git":    {"-C", "-c", "--git-dir", "--work-tree", "--namespace", "--super-prefix", "--config-env"},
	"npm":    {"--prefix", "-C", "-w", "--workspace", "--cache", "--userconfig", "--registry", "--loglevel"},
	"dotnet": {},
	"unity":  {"--format", "--project"},
	"claude": {"--settings", "--model", "--agent", "--add-dir", "--output-format", "--input-format",
		"--permission-mode", "--session-id", "--resume", "-r", "--allowedTools", "--allowed-tools", "--disallowedTools",
		"--disallowed-tools", "--mcp-config", "--append-system-prompt", "--system-prompt", "--max-turns",
		"--fallback-model", "--setting-sources", "--permission-prompt-tool", "--plugin-dir", "--tools"},
	"gh":     {"-R", "--repo", "--hostname"},
	"go":     {"-C"},
	"bonsai": {},
}

var wrappers = map[string]bool{"sudo": true, "env": true, "time": true, "nohup": true, "exec": true, "command": true, "nice": true}

var wrapperValueFlags = map[string]bool{"-u": true, "-g": true, "-C": true, "-n": true}

var navigation = map[string]bool{"cd": true, "pushd": true, "popd": true, "set-location": true, "push-location": true,
	"pop-location": true, "sl": true, "chdir": true}

var programExtensions = []string{".exe", ".cmd", ".bat", ".com", ".ps1", ".sh", ".bash", ".mjs", ".cjs", ".js", ".py"}

// isHeadWord is [a-z][a-z0-9-]*.
func isHeadWord(s string) bool {
	if s == "" || s[0] < 'a' || s[0] > 'z' {
		return false
	}
	for i := 1; i < len(s); i++ {
		c := s[i]
		if !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '-') {
			return false
		}
	}
	return true
}

// isCmdlet is a PowerShell Verb-Noun: [A-Za-z]+-[A-Za-z][A-Za-z0-9]*.
func isCmdlet(s string) bool {
	i := strings.IndexByte(s, '-')
	if i <= 0 || i+1 >= len(s) || !isLetter(rune(s[i+1])) {
		return false
	}
	for k := 0; k < len(s); k++ {
		c := rune(s[k])
		switch {
		case k < i && !isLetter(c), k > i && !isAlnum(c):
			return false
		}
	}
	return true
}

// isAssignment is a NAME=value word whose = is not quoted.
func isAssignment(w shellWord) bool {
	s := w.v
	if s == "" || !(isLetter(rune(s[0])) || s[0] == '_') {
		return false
	}
	for k := 1; k < len(s); k++ {
		c := rune(s[k])
		if c == '=' {
			return w.qAt < 0 || w.qAt >= k+1
		}
		if !isWord(c) {
			return false
		}
	}
	return false
}

// baseName is a program's file name without its folder or a known extension.
func baseName(v string) string {
	if i := strings.LastIndexAny(v, `/\`); i >= 0 {
		v = v[i+1:]
	}
	low := strings.ToLower(v)
	for _, ext := range programExtensions {
		if strings.HasSuffix(low, ext) {
			return v[:len(v)-len(ext)]
		}
	}
	return v
}

type head struct {
	head string
	skip bool
}

// segmentHead reads one part of a command line.
func segmentHead(words []shellWord, powershell bool) head {
	i := 0
	if powershell {
		if len(words) > 0 && strings.HasPrefix(words[0].v, "$") && words[0].qAt != 0 {
			return head{skip: true} // $x = ..., $env:X='...'
		}
		for i < len(words) && !words[i].quoted && (words[i].v == "&" || words[i].v == ".") {
			i++
		}
	} else {
		for i < len(words) && isAssignment(words[i]) {
			i++
		}
		for i < len(words) && !words[i].quoted && wrappers[words[i].v] {
			i++
			for i < len(words) && (isAssignment(words[i]) || !words[i].quoted && strings.HasPrefix(words[i].v, "-")) {
				if !words[i].quoted && wrapperValueFlags[words[i].v] {
					i++
				}
				i++
			}
		}
	}
	if i >= len(words) {
		return head{skip: true}
	}
	prog := words[i]
	if powershell && !prog.quoted && isCmdlet(prog.v) {
		low := strings.ToLower(prog.v)
		if navigation[low] {
			return head{low, true}
		}
		return head{prog.v, false}
	}
	name := strings.ToLower(baseName(prog.v))
	if !isHeadWord(name) {
		return head{}
	}
	if navigation[name] {
		return head{name, true}
	}
	valueFlags, ok := subcommandPrograms[name]
	if !ok {
		return head{name, false}
	}
	for j := i + 1; j < len(words); j++ {
		w := words[j]
		if !w.quoted && strings.HasPrefix(w.v, "-") {
			for _, f := range valueFlags {
				if w.v == f {
					j++
					break
				}
			}
			continue
		}
		if !w.quoted && isHeadWord(w.v) {
			return head{name + " " + w.v, false}
		}
		break
	}
	return head{name, false}
}
