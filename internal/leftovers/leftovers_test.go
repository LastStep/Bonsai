package leftovers

import (
	"slices"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/schema"
)

// pids gives the found processes' pids, in their order.
func pids(ps []Process) []int {
	out := make([]int, 0, len(ps))
	for _, p := range ps {
		out = append(out, p.PID)
	}
	return out
}

// Linux's rows (POSIX paths, in their case): what ties a process to the checkout, and what leaves it out.
func TestFindPosix(t *testing.T) {
	root := []string{"/work/repo"}
	rows := []Row{
		{PID: 1, PPID: 0, Name: "init", Start: 1, Exe: "/sbin/init", Folder: "/", Argv: []string{"/sbin/init"}},
		{PID: 10, PPID: 1, Name: "claude", Start: 10, Folder: "/work/repo", Argv: []string{"claude"}},    // an ancestor
		{PID: 11, PPID: 10, Name: "bash", Start: 11, Folder: "/work/repo", Argv: []string{"bash", "-c"}}, // an ancestor
		{PID: 12, PPID: 11, Name: "bonsai", Start: 12, Folder: "/work/repo", Argv: []string{"bonsai", "ladder"}},
		{PID: 13, PPID: 12, Name: "git", Start: 13, Folder: "/work/repo", Argv: []string{"git", "status"}},   // a descendant
		{PID: 14, PPID: 13, Name: "sh", Start: 14, Folder: "/work/repo", Argv: []string{"sh"}},               // its child
		{PID: 20, PPID: 10, Name: "node", Start: 20, Folder: "/work/repo", Argv: []string{"node", "srv.js"}}, // an ancestor's other child
		{PID: 21, PPID: 1, Name: "vite", Start: 21, Folder: "/home", Exe: "/work/repo/node_modules/.bin/vite"},
		{PID: 22, PPID: 1, Name: "node", Start: 22, Folder: "/home", Argv: []string{"node", "/work/repo/server.js"}},
		{PID: 23, PPID: 1, Name: "x", Start: 23, Argv: []string{"x", "--config=/work/repo/x.json"}},
		{PID: 24, PPID: 1, Name: "x", Start: 24, Argv: []string{"x", "file:///work/repo/a"}},
		{PID: 25, PPID: 1, Name: "x", Start: 25, Argv: []string{"x", "/work/repo2/a"}},   // a sibling folder
		{PID: 26, PPID: 1, Name: "x", Start: 26, Argv: []string{"x", "work/repo/a"}},     // relative: not a path
		{PID: 27, PPID: 1, Name: "x", Start: 27, Argv: []string{"x", "/work/repo/../b"}}, // cleaned: outside
		{PID: 28, PPID: 1, Name: "x", Start: 28, Argv: []string{"x", "/work/repo"}},      // the root itself
		{PID: 29, PPID: 1, Name: "x", Start: 29, Folder: "/work/Repo"},                   // POSIX keeps case
		{PID: 30, PPID: 12, Name: "y", Start: 5, Folder: "/work/repo"},                   // older than its "parent": a taken pid
		{PID: 31, PPID: 1, Name: "go", Start: 31, Folder: "/work/repo/cmd", Exe: "/tmp/go-build1/exe/x", Argv: []string{"/tmp/go-build1/exe/x"}},
	}
	got := find(rows, root, 12, posixStyle)
	if want := []int{30, 20, 21, 22, 23, 24, 28, 31}; !slices.Equal(pids(got), want) {
		t.Errorf("found %v, want %v (oldest first)", pids(got), want)
	}
	if got[1].Head != "node" || got[0].Name != "y" {
		t.Errorf("rows: %+v", got[:2])
	}
	if f := find(rows, root, 999, posixStyle); len(f) != 13 {
		t.Errorf("an unknown runner leaves out nothing: %v", pids(f))
	}
	if f := find(rows, nil, 12, posixStyle); len(f) != 0 {
		t.Errorf("no root finds nothing: %v", pids(f))
	}
}

// Windows' rows (Windows paths, in any case): images, command-line paths in each form, no working folder.
func TestFindWindows(t *testing.T) {
	root := []string{asPath(`C:\Work\Repo`, windowsStyle)}
	if root[0] != "c:/work/repo" {
		t.Fatalf("root %q", root[0])
	}
	rows := []Row{
		{PID: 4, PPID: 0, Name: "System", Start: 1},
		{PID: 100, PPID: 4, Name: "explorer.exe", Start: 2, Exe: `C:\Windows\explorer.exe`},
		{PID: 200, PPID: 100, Name: "bonsai.exe", Start: 3, Exe: `C:\Program Files\Bonsai\bonsai.exe`,
			Argv: []string{`C:\Program Files\Bonsai\bonsai.exe`, "ladder"}},
		{PID: 201, PPID: 200, Name: "cmd.exe", Start: 4, Exe: `C:\Windows\System32\cmd.exe`,
			Argv: []string{`C:\Windows\System32\cmd.exe`, "/c", `C:\work\repo\x.cmd`}}, // the runner's descendant
		{PID: 300, PPID: 4, Name: "node.exe", Start: 5, Exe: `C:\Work\Repo\node_modules\node.exe`},
		{PID: 301, PPID: 4, Name: "node.exe", Start: 6, Exe: `C:\Program Files\nodejs\node.exe`,
			Argv: []string{"node", `c:\WORK\repo\server.js`}},
		{PID: 302, PPID: 4, Name: "bash.exe", Start: 7, Argv: []string{"bash", "/c/work/repo/run.sh"}},
		{PID: 303, PPID: 4, Name: "x.exe", Start: 8, Argv: []string{"x", `\\?\C:\Work\Repo\a`}},
		{PID: 304, PPID: 4, Name: "x.exe", Start: 9, Argv: []string{"x", "file:///C:/Work/Repo/a"}},
		{PID: 305, PPID: 4, Name: "x.exe", Start: 10, Argv: []string{"x", `--out=C:/work/repo/o`}},
		{PID: 306, PPID: 4, Name: "x.exe", Start: 11, Argv: []string{"x", `C:\work\repo2\a`}},
		{PID: 307, PPID: 4, Name: "x.exe", Start: 12, Argv: []string{"x", `repo\a`}},
		{PID: 308, PPID: 4, Name: "x.exe", Start: 13}, // nothing readable
	}
	got := find(rows, root, 200, windowsStyle)
	if want := []int{300, 301, 302, 303, 304, 305}; !slices.Equal(pids(got), want) {
		t.Errorf("found %v, want %v", pids(got), want)
	}
	if got[1].Head != "node" {
		t.Errorf("head %q", got[1].Head)
	}
}

func TestAsPath(t *testing.T) {
	for _, c := range []struct {
		in   string
		st   style
		want string
	}{
		{"/a/b/../c", posixStyle, "/a/c"},
		{"file:///a/b", posixStyle, "/a/b"},
		{"a/b", posixStyle, ""},
		{"", posixStyle, ""},
		{`C:\A\B`, windowsStyle, "c:/a/b"},
		{"/c/Users/x", windowsStyle, "c:/users/x"},
		{"/c", windowsStyle, "c:"},
		{`\\server\Share\x`, windowsStyle, "//server/share/x"},
		{`\\?\D:\x`, windowsStyle, "d:/x"},
		{"file:///D:/x", windowsStyle, "d:/x"},
		{"/usr/bin", windowsStyle, ""},
		{"x", windowsStyle, ""},
	} {
		if got := asPath(c.in, c.st); got != c.want {
			t.Errorf("asPath(%q, %d) = %q, want %q", c.in, c.st, got, c.want)
		}
	}
}

func TestHeadOf(t *testing.T) {
	for _, c := range []struct {
		argv []string
		want string
	}{
		{[]string{"/usr/bin/git", "-C", "/a b", "status", "--porcelain"}, "git status"},
		{[]string{`C:\Program Files\Go\bin\go.exe`, "test", "./..."}, "go test"},
		{[]string{"node", "-e", "fetch('x?token=SECRET')"}, "node"},
		{[]string{"/bin/sh", "-c", "TOKEN=SECRET curl https://x"}, "sh"},
		{[]string{"env", "A=SECRET", "npm", "run", "dev"}, "npm run"},
		{[]string{"/tmp/x/rungjob.test", "sleep"}, ""},
		{[]string{"it's"}, ""},
		{[]string{"npm exec @scope/some-server"}, "npm exec"},
		{[]string{"/opt/My App/bin/tool", "x"}, "tool"},
		{nil, ""},
	} {
		if got := HeadOf(c.argv); got != c.want {
			t.Errorf("HeadOf(%q) = %q, want %q", c.argv, got, c.want)
		}
	}
}

func TestSplitCommandLine(t *testing.T) {
	for _, c := range []struct {
		in   string
		want []string
	}{
		{`a b  c`, []string{"a", "b", "c"}},
		{`"C:\Program Files\x.exe" -a "b c"`, []string{`C:\Program Files\x.exe`, "-a", "b c"}},
		{`a\\\"b "c\\" d`, []string{`a\"b`, `c\`, "d"}},
		{`"a""b" ""`, []string{`a"b`, ""}},
		{`a\b\c`, []string{`a\b\c`}},
		{``, nil},
	} {
		if got := SplitCommandLine(c.in); !slices.Equal(got, c.want) {
			t.Errorf("SplitCommandLine(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// The value as the result records it, byte-stable, and the line as a person reads it.
func TestRecordAndLines(t *testing.T) {
	var many []Process
	for i := 1; i <= 12; i++ {
		many = append(many, Process{PID: 100 + i, Name: "sl\u00e9ep", Head: ""})
	}
	many[0].Head = "go test"
	for _, c := range []struct {
		v     Value
		rec   string
		lines []string
	}{
		{Value{}, "null\n", nil},
		{Value{Kind: Found, MS: 7}, `{"checkout":[],"ms":7}` + "\n", []string{"leftovers not started by this run: none"}},
		{Value{Kind: Found, Checkout: many[:2], MS: 3},
			`{"checkout":[{"pid":101,"name":"sl\u00e9ep","head":"go test"},{"pid":102,"name":"sl\u00e9ep","head":null}],"ms":3}` + "\n",
			[]string{"leftovers not started by this run: 2 (pids 101, 102; read-only, stop only what you started; " +
				"their names and commands are in the result file and in --json)"}},
		{Value{Kind: Unknown, Why: "no /proc", MS: 1}, `{"unknown":"no /proc","ms":1}` + "\n",
			[]string{"leftovers: unknown (no /proc)"}},
		{Value{Kind: NotChecked, Why: "darwin"}, `"not checked (darwin)"` + "\n",
			[]string{"leftovers: not checked (darwin)"}},
	} {
		b, err := schema.EncodeLine(c.v.Record())
		if err != nil {
			t.Fatal(err)
		}
		if string(b) != c.rec {
			t.Errorf("record %s, want %s", b, c.rec)
		}
		if got := c.v.Lines(); !slices.Equal(got, c.lines) {
			t.Errorf("lines %q, want %q", got, c.lines)
		}
	}
	lines := Value{Kind: Found, Checkout: many}.Lines()
	if len(lines) != 1 || !strings.Contains(lines[0], "12 (pids 101, 102, 103, 104, 105, 106, 107, 108, 109, 110 and 2 more;") {
		t.Errorf("twelve: %q", lines)
	}
	for _, l := range lines {
		if strings.IndexFunc(l, func(r rune) bool { return r > 0x7e }) >= 0 {
			t.Errorf("not ASCII: %q", l)
		}
	}
}

// BUG-097: the line a climb prints names no program: a count and pids only, so a published climb does not name
// what else the person runs. The names and command heads stay in the record (the result file and --json).
func TestLinesNameNoProgram(t *testing.T) {
	v := Value{Kind: Found, Checkout: []Process{
		{PID: 4121, Name: "chrome", Head: "chrome --user-data-dir"},
		{PID: 4130, Name: "chrome_crashpad", Head: ""},
	}}
	got := strings.Join(v.Lines(), "\n")
	for _, bad := range []string{"chrome", "user-data-dir", "crashpad"} {
		if strings.Contains(got, bad) {
			t.Errorf("the line names %q:\n%s", bad, got)
		}
	}
	for _, want := range []string{": 2 ", "4121", "4130", "result file", "--json"} {
		if !strings.Contains(got, want) {
			t.Errorf("the line lacks %q:\n%s", want, got)
		}
	}
	b, err := schema.EncodeLine(v.Record())
	if err != nil || !strings.Contains(string(b), `"name":"chrome"`) || !strings.Contains(string(b), `"head":"chrome --user-data-dir"`) {
		t.Errorf("the record lost the names: %s (%v)", b, err)
	}
	if n := len(v.Lines()); n != 1 {
		t.Errorf("%d lines, want 1", n)
	}
}
