package guard

// VG's third check of 5.5.11 (15 Oct), each reproducer a test: S2, past a budget only this system's installed place
// (s2plain.py on Linux, replay-w4.ps1 on Windows, both with a plain build, whose places are the installed ones, as a
// go test build's are).

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

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
