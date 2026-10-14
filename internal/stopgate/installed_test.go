package stopgate

import (
	"os"
	"runtime"
	"strings"
	"testing"
)

// installed is the bonsai a next step names, written out here on purpose (the hook line's place, as a shell names it):
// /usr/local/bin/bonsai on Linux, C:\Program Files\Bonsai\bonsai.exe, quoted for its space, on Windows. Every test
// that reads a next step compares it to this, so each side reads its own.
var installed = func() string {
	if runtime.GOOS == "windows" {
		return `"C:\Program Files\Bonsai\bonsai.exe"`
	}
	return "/usr/local/bin/bonsai"
}()

func TestLadderCommandNamesTheInstalledBonsai(t *testing.T) {
	if got, want := ladderCommand("T-0901"), installed+" ladder --task T-0901"; got != want {
		t.Errorf("ladderCommand = %q, want %q", got, want)
	}
}

// No next step is written by hand: the gate's source never says a bare "bonsai ladder", only ladderCommand does.
func TestNoNextStepNamesABareBonsai(t *testing.T) {
	src, err := os.ReadFile("stopgate.go")
	if err != nil {
		t.Fatal(err)
	}
	for i, line := range strings.Split(string(src), "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "//") {
			continue
		}
		if strings.Contains(line, "bonsai ladder") {
			t.Errorf("stopgate.go:%d writes a bare \"bonsai ladder\": use ladderCommand: %s", i+1, line)
		}
	}
}
