package guard

// VG's third check of 5.5.11 (15 Oct), each reproducer a test: S2, past a budget only this system's installed place
// (s2plain.py on Linux, replay-w4.ps1 on Windows, both with a plain build, whose places are the installed ones, as a
// go test build's are); a calls file rewritten under its count holding the window spent (rewrite.py).

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/LastStep/Bonsai/internal/agents"
	"github.com/LastStep/Bonsai/internal/engine"
)

// S2: the forms each system's shells name its own installed place by, and the other system's place refused, read for
// both systems on every system (namesPlace is the word reader installedBonsai runs with this system's place).
func TestS2OnlyThisSystemsPlaceIsTheInstalledBonsai(t *testing.T) {
	type form struct {
		word      string
		bash, ps1 bool // read as the place in Bash (Git Bash on Windows), in PowerShell
	}
	linux := []form{{"/usr/local/bin/bonsai", true, true},
		// Windows' place is a path relative to the shell's folder on Linux, which any agent may make.
		{"C:/Program Files/Bonsai/bonsai.exe", false, false}, {"/c/Program Files/Bonsai/bonsai.exe", false, false},
		{`C:\Program Files\Bonsai\bonsai.exe`, false, false}, {"c:/program files/bonsai/BONSAI", false, false},
		// On Linux a backslash is a file name's letter, and the letter case and .exe count.
		{`\usr\local\bin\bonsai`, false, false}, {"/usr/local/bin/Bonsai", false, false}, {"/usr/local/bin/bonsai.exe", false, false},
		{"/usr/local/bin//bonsai", false, false}, {"bonsai", false, false}, {"./bonsai", false, false}}
	windows := []form{{"/c/Program Files/Bonsai/bonsai.exe", true, false}, {"C:/Program Files/Bonsai/bonsai.exe", true, true},
		{`C:\Program Files\Bonsai\bonsai.exe`, true, true}, {`c:\program files\bonsai\BONSAI`, true, true},
		{"C:/PROGRAM FILES/BONSAI/BONSAI.EXE", true, true}, {`C:/Program Files\Bonsai/bonsai`, true, true},
		// Linux's place is a folder at the drive's root on Windows, which an ordinary user may make (C:\usr\local\bin);
		// so is /c/... in PowerShell, and \c\... in either shell (C:\c\...).
		{"/usr/local/bin/bonsai", false, false}, {`\usr\local\bin\bonsai`, false, false}, {`\c\Program Files\Bonsai\bonsai.exe`, false, false},
		// Git Bash's mount names the drive in lower case.
		{"/C/Program Files/Bonsai/bonsai.exe", false, false}, {"D:/Program Files/Bonsai/bonsai.exe", false, false},
		{"C:Program Files/Bonsai/bonsai.exe", false, false}, {"C:/Program Files/Bonsai/bonsai.com", false, false},
		{"bonsai.exe", false, false}, {"Program Files/Bonsai/bonsai.exe", false, false}}
	for _, sys := range []struct {
		name, place string
		windows     bool
		forms       []form
	}{{"linux", engine.PlaceLinux, false, linux}, {"windows", engine.PlaceWindows, true, windows}} {
		for _, f := range sys.forms {
			if got := namesPlace(f.word, sys.place, sys.windows, shBash); got != f.bash {
				t.Errorf("%s, Bash: %q read as the place %v, want %v", sys.name, f.word, got, f.bash)
			}
			if got := namesPlace(f.word, sys.place, sys.windows, shPower); got != f.ps1 {
				t.Errorf("%s, PowerShell: %q read as the place %v, want %v", sys.name, f.word, got, f.ps1)
			}
		}
	}
	// A test build's stamped place, in Git Bash's form as the build's line writes it on Windows: its own forms there.
	stamp := "/c/Users/x/bonsai-checks/p/bonsai.exe"
	for word, want := range map[string]bool{"/c/Users/x/bonsai-checks/p/bonsai.exe": true, `C:\Users\x\bonsai-checks\p\bonsai`: true,
		"/c/Program Files/Bonsai/bonsai.exe": false, "/usr/local/bin/bonsai": false} {
		if got := namesPlace(word, stamp, true, shBash); got != want {
			t.Errorf("a stamped place, %q: %v, want %v", word, got, want)
		}
	}
	// This system's: thisPlace, which hookPlace shows.
	place, onWindows := thisPlace()
	if want := runtime.GOOS == "windows"; onWindows != want || (want && place != engine.PlaceWindows) || (!want && place != engine.PlaceLinux) {
		t.Errorf("this system's place: %q (windows %v)", place, onWindows)
	}
}

// S2, s2plain.py: on Linux, with a plain build's places, a builder plants a stand-in at ./C:/Program Files/Bonsai/
// bonsai.exe within its budget; past it, that relative path (in each form s2plain.py tried) is neither its report nor
// a brief's read, and the installed bonsai at /usr/local/bin/bonsai is.
func TestS2PlainLinuxTakesOnlyLinuxPlace(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Linux's case: on Windows C:/Program Files/Bonsai/bonsai.exe is the installed place itself (TestS2PlainWindowsTakesOnlyWindowsPlace)")
	}
	dir := briefProject(t)
	stand := filepath.Join(dir, "C:", "Program Files", "Bonsai", "bonsai.exe")
	writeFile(t, stand, "#!/bin/sh\necho stand-in ran\n")
	if err := os.Chmod(stand, 0o755); err != nil {
		t.Fatal(err)
	}
	spentTiny(t, dir, "s2")
	report := " message --kind report --from tiny@T-0001 --to orchestrator --thread " + b1 + " --status blocked --body 'anything at all'"
	for _, prog := range []string{`"C:/Program Files/Bonsai/bonsai.exe"`, `'c:/program files/bonsai/BONSAI'`, `C:/Program\ Files/Bonsai/bonsai.exe`,
		`'/c/Program Files/Bonsai/bonsai.exe'`, `'C:\Program Files\Bonsai\bonsai.exe'`, `'\usr\local\bin\bonsai'`, "bonsai"} {
		s := wantAnswer(t, dir, tinyFrom(dir, "s2", "Bash", bash(prog+report)), 2, RuleBudgetSpent)
		if !strings.Contains(s, "only your report") || !strings.Contains(s, "/usr/local/bin/bonsai message --kind report --from tiny@T-0001") {
			t.Errorf("the refusal of %s:\n%s", prog, s)
		}
		wantAnswer(t, dir, tinyFrom(dir, "s2", "PowerShell", bash("& "+prog+report)), 2, RuleBudgetSpent)
	}
	wantAnswer(t, dir, tinyFrom(dir, "s2", "Bash", bash(`"C:/Program Files/Bonsai/bonsai.exe" message --show `+b1)), 2, RuleBudgetSpent)
	wantAnswer(t, dir, tinyFrom(dir, "s2", "Bash", bash("/usr/local/bin/bonsai message --show "+b1)), 0, RuleBudgetExempt)
	wantAnswer(t, dir, tinyFrom(dir, "s2", "PowerShell", bash("& /usr/local/bin/bonsai"+report)), 0, RuleBudgetExempt)
	wantAnswer(t, dir, tinyFrom(dir, "s2", "Bash", bash("/usr/local/bin/bonsai"+report)), 0, RuleBudgetExempt)
}

// S2, replay-w4.ps1: on Windows, with a plain build's places, past a budget Linux's place is neither the report nor a
// brief's read in Git Bash or PowerShell (in PowerShell it is C:\usr\local\bin\bonsai), nor is /c/... in PowerShell,
// nor a copy of bonsai.exe anywhere else; the installed place in its Windows forms is, in each shell.
func TestS2PlainWindowsTakesOnlyWindowsPlace(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows' case: on Linux /usr/local/bin/bonsai is the installed place itself (TestS2PlainLinuxTakesOnlyLinuxPlace)")
	}
	dir := briefProject(t)
	spentTiny(t, dir, "w4")
	report := " message --kind report --from tiny@T-0001 --to orchestrator --thread " + b1 + " --status blocked --body 'x'"
	other := filepath.ToSlash(filepath.Join(t.TempDir(), "other", "bonsai.exe"))
	for _, prog := range []string{"bonsai", "bonsai.exe", "./bonsai.exe", other, `env "C:/Program Files/Bonsai/bonsai.exe"`,
		"/usr/local/bin/bonsai", "'/usr/local/bin/bonsai'", `'\usr\local\bin\bonsai'`, `"/C/Program Files/Bonsai/bonsai.exe"`} {
		s := wantAnswer(t, dir, tinyFrom(dir, "w4", "Bash", bash(prog+report)), 2, RuleBudgetSpent)
		if !strings.Contains(s, `"C:/Program Files/Bonsai/bonsai.exe" message --kind report --from tiny@T-0001`) {
			t.Errorf("the refusal of %s:\n%s", prog, s)
		}
	}
	for _, prog := range []string{"& /usr/local/bin/bonsai", "& '/usr/local/bin/bonsai'", "/usr/local/bin/bonsai", `& '\usr\local\bin\bonsai'`,
		`& '/c/Program Files/Bonsai/bonsai.exe'`, `& "\c\Program Files\Bonsai\bonsai.exe"`, "& bonsai", "& '" + other + "'"} {
		s := wantAnswer(t, dir, tinyFrom(dir, "w4", "PowerShell", bash(prog+report)), 2, RuleBudgetSpent)
		if !strings.Contains(s, `(in PowerShell: & "C:/Program Files/Bonsai/bonsai.exe" message --kind report`) {
			t.Errorf("the refusal of %s:\n%s", prog, s)
		}
	}
	wantAnswer(t, dir, tinyFrom(dir, "w4", "PowerShell", bash("& /usr/local/bin/bonsai message --show "+b1)), 2, RuleBudgetSpent)
	for _, prog := range []string{`"C:/Program Files/Bonsai/bonsai.exe"`, `"/c/Program Files/Bonsai/bonsai.exe"`, `'C:\Program Files\Bonsai\bonsai.exe'`,
		`"c:/program files/bonsai/BONSAI"`} {
		wantAnswer(t, dir, tinyFrom(dir, "w4", "Bash", bash(prog+" message --show "+b1)), 0, RuleBudgetExempt)
		wantAnswer(t, dir, tinyFrom(dir, "w4", "Bash", bash(prog+report)), 0, RuleBudgetExempt)
	}
	for _, prog := range []string{`& "C:/Program Files/Bonsai/bonsai.exe"`, `& 'C:\Program Files\Bonsai\bonsai.exe'`, `& 'c:/program files/bonsai/bonsai'`} {
		wantAnswer(t, dir, tinyFrom(dir, "w4", "PowerShell", bash(prog+report)), 0, RuleBudgetExempt)
	}
}

// The should-fix, rewrite.py: a write word outside the guard's tables (rsync of an empty file over the agent's own
// calls file) is let through, and the file is rewritten under the count; the agent's next call finds fewer call lines
// than were counted, so its window is held spent (as a state moved aside a second time is), its report, its brief's
// read and the hand-back still let through, and a newer brief from its caller opens a fresh window.
func TestRewrittenCallsHoldTheWindowSpent(t *testing.T) {
	dir := briefProject(t)
	at := func(sec int, payload string, code int, rule string) string {
		t.Helper()
		got, s := runAt(t, dir, payload, testNow.Add(time.Duration(sec)*time.Second), false)
		if got != code || rule != "" && lastRule(t, dir, payload) != rule {
			t.Fatalf("at %ds: exit %d, rule %s, want %d %s: %s", sec, got, lastRule(t, dir, payload), code, rule, s)
		}
		return s
	}
	calls := ".bonsai/local/agents/" + sessA + "/r1.calls"
	file := filepath.Join(dir, filepath.FromSlash(calls))
	at(0, start(dir, "toolu_01RW1", "Your brief: "+b1+"."), 0, RuleCallAllowed)
	at(0, tinyFrom(dir, "r1", "Bash", bash("echo 1")), 0, RuleShell)
	at(1, tinyFrom(dir, "r1", "Bash", bash("rsync /tmp/vg3-empty "+calls)), 0, "")
	writeFile(t, file, "") // what that rsync does
	s := at(30, tinyFrom(dir, "r1", "Bash", bash("echo after0")), 2, RuleBudgetSpent)
	for _, want := range []string{"budget held spent (this agent's calls file holds fewer calls than the guard had counted",
		"report what you have", hookPlace(false) + " message --kind report --from tiny@T-0001 --to orchestrator --thread " + b1} {
		if !strings.Contains(s, want) {
			t.Errorf("the refusal has no %q:\n%s", want, s)
		}
	}
	for i := 1; i <= 4; i++ {
		if s := at(30+i, tinyFrom(dir, "r1", "Bash", bash("echo after")), 2, RuleBudgetSpent); !strings.Contains(s, "only your report") {
			t.Errorf("a later refusal: %s", s)
		}
	}
	found := false
	for _, r := range sessionRecords(t, dir, tinyFrom(dir, "r1", "Read", in("file_path", "x"))) {
		found = found || r.String("rule") == RuleStateUnreadable && strings.Contains(r.String("text"), "holds no call line before this call's")
	}
	if !found {
		t.Error("no state-unreadable record of the rewritten calls file")
	}
	report := hookPlace(false) + " message --kind report --from tiny@T-0001 --to orchestrator --thread " + b1 + " --status blocked --body 'x'"
	at(40, tinyFrom(dir, "r1", "Bash", bash(report)), 0, RuleBudgetExempt)
	at(41, tinyFrom(dir, "r1", "Read", readOf(dir, ".bonsai/local/messages/tiny-T-0001/"+b1+".md")), 0, RuleBudgetExempt)
	at(42, tinyFrom(dir, "r1", "SubagentHandback", in("message", "x")), 0, RuleBudgetExempt)
	at(43, hookCall{session: sessA, tool: "SendMessage", input: in("to", "r1", "message", "Read "+b2)}.payload(dir), 0, RuleCallAllowed)
	at(44, tinyFrom(dir, "r1", "Bash", bash("echo rebriefed")), 0, RuleShell)

	// A window opened by a later tie: the file put back to an earlier copy of itself, shorter than at the window's
	// start, holds it spent at once (here 3 seconds into the agent's life).
	dir = briefProject(t)
	at(0, start(dir, "toolu_01RW2", "Your brief: "+b1+"."), 0, RuleCallAllowed)
	at(0, tinyFrom(dir, "r2", "Bash", bash("echo 1")), 0, RuleShell)
	copied := filepath.Join(dir, ".bonsai", "local", "agents", sessA, "r2.calls")
	early, err := os.ReadFile(copied)
	if err != nil {
		t.Fatal(err)
	}
	at(1, tinyFrom(dir, "r2", "Bash", bash("echo 2")), 0, RuleShell)
	at(2, hookCall{session: sessA, tool: "SendMessage", input: in("to", "r2", "message", "Read "+b2)}.payload(dir), 0, RuleCallAllowed)
	at(2, tinyFrom(dir, "r2", "Bash", bash("echo 3")), 0, RuleShell)
	if w := stateOf(t, dir, sessA, "r2").Window(); w.Tie == nil || w.Tie.Brief != b2 || w.Calls != 2 {
		t.Fatalf("the re-briefed window: %+v", w)
	}
	writeFile(t, copied, string(early)) // one call line: fewer than the two before the window
	s = at(3, tinyFrom(dir, "r2", "Bash", bash("echo 4")), 2, RuleBudgetSpent)
	if !strings.Contains(s, "budget held spent (this agent's calls file holds fewer calls") {
		t.Errorf("%s", s)
	}
	found = false
	for _, r := range sessionRecords(t, dir, tinyFrom(dir, "r2", "Read", in("file_path", "x"))) {
		found = found || r.String("rule") == RuleStateUnreadable && strings.Contains(r.String("text"), "holds 1 call lines before this call's, fewer than the 2")
	}
	if !found {
		t.Error("no state-unreadable record of the calls file put back")
	}
}

// The other side of the backstop: an agent's first calls side by side, one of which finds its state made by another
// whose call line is still to come (its calls file empty), are counted as before, within the first call's guard run;
// and an agent whose calls file was never rewritten is never held spent, however long it runs.
func TestAnEmptyCallsFileBesideTheFirstCallIsCounted(t *testing.T) {
	dir := briefProject(t)
	runAtOK := func(sec int, payload string) {
		t.Helper()
		if code, s := runAt(t, dir, payload, testNow.Add(time.Duration(sec)*time.Second), false); code != 0 {
			t.Fatalf("at %ds: exit %d: %s", sec, code, s)
		}
	}
	runAtOK(0, start(dir, "toolu_01SB1", "Your brief: "+b5+"."))
	runAtOK(0, hookCall{session: sessA, agentID: "q1", agentType: "workflow:tiny", tool: "Bash", input: bash("echo first")}.payload(dir))
	// The first call's line not yet written, as a call side by side sees it 3 seconds on.
	if err := os.Remove(filepath.Join(dir, ".bonsai", "local", "agents", sessA, "q1.calls")); err != nil {
		t.Fatal(err)
	}
	runAtOK(3, hookCall{session: sessA, agentID: "q1", agentType: "workflow:tiny", tool: "Bash", input: bash("echo beside")}.payload(dir))
	for i := 1; i <= 5; i++ {
		runAtOK(3+i*60, hookCall{session: sessA, agentID: "q1", agentType: "workflow:tiny", tool: "Bash", input: bash("echo later")}.payload(dir))
	}
	if st := stateOf(t, dir, sessA, "q1"); st.Spent() != nil {
		t.Errorf("held spent: %+v", st.Spent())
	}
	if n, _ := (agents.Store{Main: dir}).Calls(sessA, "q1"); n != 6 {
		t.Errorf("%d calls counted, want 6", n)
	}
}
