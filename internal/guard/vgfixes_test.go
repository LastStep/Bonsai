package guard

// VG's items D, E, F and J (15 Oct), each reproducer a test: past a budget only the installed bonsai counts as the
// report or a brief's read (D); a climb with no --task counts against BONSAI_TASK (E); mkfifo and mknod are writes into
// Bonsai's files (F); the literal write forms VG listed, and Claude Code's native binary by its path (J).

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// spentTiny starts a tiny with the brief b1 (3 actions) and spends it: its fourth call is refused.
func spentTiny(t *testing.T, dir, id string) {
	t.Helper()
	wantAnswer(t, dir, start(dir, "toolu_01SPENT"+id, "Your brief: "+b1+"."), 0, RuleCallAllowed)
	for i := 0; i < 3; i++ {
		wantAnswer(t, dir, tinyFrom(dir, id, "Read", readOf(dir, "free.txt")), 0, RuleOtherTool)
	}
	wantAnswer(t, dir, tinyFrom(dir, id, "Read", readOf(dir, "free.txt")), 2, RuleBudgetSpent)
}

// D, VG's reproducer: past a budget, a program named bonsai at another place (a stand-in on the agent's own path, one
// in the folder it runs in) is neither its report nor a brief's read; the installed bonsai at its place, in each form
// a shell names it, and a bare bonsai are.
func TestDOnlyTheInstalledBonsaiPastABudget(t *testing.T) {
	dir := briefProject(t)
	spentTiny(t, dir, "d1")
	report := " message --kind report --from tiny@T-0001 --to orchestrator --thread " + b1 + " --status blocked --body 'left'"
	for _, prog := range []string{filepath.ToSlash(t.TempDir()) + "/fakebin/bonsai", "./bonsai", "../bonsai", "bin/bonsai",
		"/usr/bin/bonsai", "/usr/local/bin/bonsai.sh", "/tmp/Program Files/Bonsai/bonsai.exe", "Bonsai"} {
		wantAnswer(t, dir, tinyFrom(dir, "d1", "Bash", bash(prog+report)), 2, RuleBudgetSpent)
		wantAnswer(t, dir, tinyFrom(dir, "d1", "Bash", bash(prog+" message --show "+b1)), 2, RuleBudgetSpent)
	}
	for _, prog := range []string{"bonsai", "/usr/local/bin/bonsai", `"/c/Program Files/Bonsai/bonsai.exe"`,
		`"C:/Program Files/Bonsai/bonsai.exe"`, `"C:\Program Files\Bonsai\bonsai.exe"`, `"c:/program files/bonsai/BONSAI"`} {
		wantAnswer(t, dir, tinyFrom(dir, "d1", "Bash", bash(prog+" message --show "+b1)), 0, RuleBudgetExempt)
		wantAnswer(t, dir, tinyFrom(dir, "d1", "Bash", bash(prog+report)), 0, RuleBudgetExempt)
	}
	// The forms the word reader holds to the place, one by one.
	for word, want := range map[string]bool{"bonsai": true, "bonsai.exe": true, "BONSAI.EXE": true, "/usr/local/bin/bonsai": true,
		"/usr/local/bin/Bonsai": false, "/usr/local/bin//bonsai": false, "/c/Program Files/Bonsai/bonsai.exe": true,
		`C:\Program Files\Bonsai\bonsai`: true, "D:/Program Files/Bonsai/bonsai.exe": false, "bonsai.com": false} {
		if got := installedBonsai(shWord{text: word}); got != want {
			t.Errorf("%q: installed %v, want %v", word, got, want)
		}
	}
	if installedBonsai(shWord{text: "bonsai", expand: true}) {
		t.Error("a word a variable stands for read as the installed bonsai")
	}
}

// ladderResult writes the ladder's own result for a task, red or green, started at a time.
func ladderResult(t *testing.T, dir, task string, green bool, started time.Time) {
	t.Helper()
	stamp := started.UTC().Format("2006-01-02T15:04:05.000Z")
	writeFile(t, filepath.Join(dir, ".bonsai", "local", "ladder", task+".json"), fmt.Sprintf(`{"format":"bonsai.ladder/1",`+
		`"task":"%s","workspace":"%s","mode":"local","started":"%s","finished":"%s","git":{"sha":"%s","branch":"main",`+
		`"dirty":false},"requested":[0],"green":%v,"rungs":[],"skipped":[],"leftovers":null,"proof":null,"labels":{}}`,
		task, rolesWS, stamp, stamp, strings.Repeat("c", 40), green))
}

// E, VG's reproducer: a named session (BONSAI_ROLE fragile, one red run; BONSAI_TASK T-0001) climbs with no --task
// behind || true, and the climbs count against BONSAI_TASK's red runs, as the ladder itself reads it; an assignment of
// BONSAI_TASK on the line, and a --task the guard cannot read, are read as the ladder would.
func TestEAClimbWithNoTaskCounts(t *testing.T) {
	dir := briefProject(t)
	t0 := time.Date(2026, 10, 15, 3, 0, 0, 0, time.UTC)
	at := func(s int) time.Time { return t0.Add(time.Duration(s) * time.Second) }
	named := func(cmd string) string {
		return hookCall{session: sessB, tool: "Bash", input: bash(cmd)}.payload(dir)
	}
	step := func(s int, cmd string, code int) string {
		t.Helper()
		got, out := runAt(t, dir, named(cmd), at(s), false, RoleEnv, "fragile", "BONSAI_TASK", "T-0001")
		if got != code {
			t.Fatalf("at %ds, %s: exit %d: %s", s, cmd, got, out)
		}
		return out
	}
	step(1, "bonsai ladder || true", 0)
	ladderResult(t, dir, "T-0001", false, at(2))
	step(3, "git status", 0)
	if st := stateOf(t, dir, sessB, sessB); st.RedRuns() != 1 {
		t.Fatalf("the climb with no --task was not counted: red runs %d, open %v", st.RedRuns(), st.OpenClimbs())
	}
	step(4, "bonsai ladder --json; true", 0)
	ladderResult(t, dir, "T-0001", false, at(5))
	if s := step(6, "git status", 2); !strings.Contains(s, "budget spent (red runs: 2 of 1)") {
		t.Errorf("%s", s)
	}
	// The task as the ladder reads it: an assignment on the line, export's, and a --task held in a variable (the
	// session's own then): each noted, its result read at a later call.
	for i, c := range []struct{ cmd, task string }{{"BONSAI_TASK=T-0007 bonsai ladder", "T-0007"},
		{"export BONSAI_TASK=T-0008; bonsai ladder", "T-0008"}, {"T=T-0009; bonsai ladder --task $T", "T-0001"}} {
		session := fmt.Sprintf("6e3cffd5-bd72-4f4a-a089-220fd2fdbf%02d", i)
		p := hookCall{session: session, tool: "Bash", input: bash(c.cmd)}.payload(dir)
		if code, s := runAt(t, dir, p, at(10), false, RoleEnv, "fragile", "BONSAI_TASK", "T-0001"); code != 0 {
			t.Fatalf("%s: %s", c.cmd, s)
		}
		if open := stateOf(t, dir, session, session).OpenClimbs(); len(open) != 1 || open[0].Task != c.task {
			t.Errorf("%s: noted %+v, want %s", c.cmd, open, c.task)
		}
	}
}

// F, VG's reproducer: mkfifo and mknod are writes into Bonsai's files, refused to every agent (a FIFO under
// .bonsai/local/agents/ hung bonsai status and messages, and stalled the guard), and writes a role that changes
// nothing may not make; the person's own session keeps today's guard.
func TestFFIFOAndNodeWrites(t *testing.T) {
	dir := rolesProject(t, nil)
	started(t, dir, sessA, "workflow:lead")
	for i, line := range []string{"mkfifo .bonsai/local/agents/" + sessA + "/x.state", "mkfifo -m 600 .bonsai/local/messages/a/b.md",
		"mknod .bonsai/local/agents/" + sessA + "/y.calls p", "mknod -m 644 .bonsai/lock.json c 1 3",
		"mkfifo --mode=600 a .bonsai/local/agents/ties/M-2026-10-15-0000000a"} {
		for _, typ := range []string{"workflow:lead", "workflow:helper", "Explore"} {
			wantAnswer(t, dir, shellAs(dir, fmt.Sprint("f", typ[len(typ)-1:], i), typ, line, false), 2, RuleBonsaiFiles)
		}
	}
	wantAnswer(t, dir, shellAs(dir, "fh", "workflow:helper", "mkfifo pipe", false), 2, RuleRoleWritesNothing)
	wantAnswer(t, dir, shellAs(dir, "fl", "workflow:lead", "mkfifo "+filepath.ToSlash(t.TempDir())+"/pipe", false), 0, "")
	wantAnswer(t, dir, shellAs(dir, "", "", "mkfifo .bonsai/local/agents/x", false), 0, "")
}

// J, VG's literal forms: each a write into the checkout a role that changes nothing may not make, and into Bonsai's
// own files no agent may; Claude Code's native binary by its path starts Claude Code.
func TestJTheListedForms(t *testing.T) {
	dir := rolesProject(t, nil)
	out := filepath.ToSlash(t.TempDir())
	for i, line := range []string{"git stage free.txt", "git symbolic-ref HEAD refs/heads/x",
		"git symbolic-ref --short HEAD refs/heads/x", "git fast-import < /dev/null", "git diff --output=out.txt",
		"git diff --output out.txt HEAD", "git log -p --output=log.txt", "git archive -o a.tar HEAD", "git archive --output=a.tar HEAD",
		"git archive -oa.tar HEAD", "git format-patch -o patches -1"} {
		wantAnswer(t, dir, shellAs(dir, fmt.Sprint("j", i), "workflow:helper", line, false), 2, RuleRoleWritesNothing)
	}
	for i, line := range []string{"Clear-Content free.txt", "clc -Path free.txt", "Set-Item -Path free.txt -Value x", "si free.txt x",
		"Clear-Item free.txt", "Get-Process | Export-Csv -Path p.csv", "Get-Process | epcsv p.csv", "Start-Transcript -Path t.txt",
		"Start-Transcript t.txt", "Invoke-WebRequest https://example.invalid -OutFile o.txt", "iwr https://example.invalid -OutFile o.txt",
		"Invoke-RestMethod -Uri https://example.invalid -OutFile o.json"} {
		wantAnswer(t, dir, shellAs(dir, fmt.Sprint("p", i), "workflow:helper", line, true), 2, RuleRoleWritesNothing)
	}
	started(t, dir, sessA, "workflow:lead")
	for _, c := range []struct {
		line string
		ps   bool
	}{{"git diff --output=.bonsai/local/agents/x", false}, {"git archive -o .bonsai/lock.json HEAD", false},
		{"Clear-Content .bonsai/local/agents/x.calls", true}, {"Set-Item .bonsai/lock.json x", true},
		{"iwr https://example.invalid -OutFile .bonsai/local/messages/a/M-2026-10-15-0000000a.md", true}} {
		wantAnswer(t, dir, shellAs(dir, "jl", "workflow:lead", c.line, c.ps), 2, RuleBonsaiFiles)
	}
	for i, line := range []string{"git symbolic-ref --short HEAD", "git symbolic-ref -q HEAD", "git diff --output=" + out + "/d.txt",
		"git diff --stat", "git log --oneline"} {
		wantAnswer(t, dir, shellAs(dir, fmt.Sprint("r", i), "workflow:helper", line, false), 0, "")
	}
	for _, line := range []string{"~/.local/share/claude/versions/2.1.300 -p x", "/home/x/.local/share/claude/versions/2.1.300 --agent lead",
		`"C:\Users\x\.local\share\claude\versions\2.1.300.exe" -p x`, "cp ~/.local/share/claude/versions/2.1.300 c && echo; ~/.local/share/claude/versions/2.1.296"} {
		wantAnswer(t, dir, shellAs(dir, "jl", "workflow:lead", line, false), 2, RuleCallRefused)
	}
}
