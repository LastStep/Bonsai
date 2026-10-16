package main

// check's stranded warning (plan-5 5.6.1, 5; contract section 3, "Moved checkouts"), its case in TestCheckTable's walk,
// and a checkout moved for real: its machine folder written (settings and an attached labels file), the folder renamed,
// the checkout recorded again there; then check warns, status's text names it, and the person's lines, run as printed
// through the system's own shell (a POSIX shell, or PowerShell on Windows) with the variable cleared, bring the
// settings over and clear the old folder, so the warning is gone. The Bonsai home's path holds a space and a quote.

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/LastStep/Bonsai/internal/engine"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// strandFolder records a checkout of this workspace at a path that is then removed, with what its machine folder holds
// (settings.json's content, attached namespaces), and gives the folder.
func strandFolder(c *cli, name, settings string, namespaces ...string) string {
	c.t.Helper()
	home, err := workspace.Home()
	if err != nil {
		c.t.Fatal(err)
	}
	cfg, err := workspace.LoadConfig(c.root)
	if err != nil {
		c.t.Fatal(err)
	}
	old := filepath.Join(c.tmp, name)
	if err := os.MkdirAll(old, 0o755); err != nil {
		c.t.Fatal(err)
	}
	if _, err := workspace.RecordCheckout(home, old, cfg.ID, time.Now()); err != nil {
		c.t.Fatal(err)
	}
	dir, err := workspace.MachineDir(home, old)
	if err != nil {
		c.t.Fatal(err)
	}
	if settings != "" {
		writeFile(c.t, filepath.Join(dir, workspace.MachineSettingsFile), settings)
	}
	for _, ns := range namespaces {
		writeFile(c.t, filepath.Join(dir, "labels", ns+".yaml"), "format: bonsai.labels/1\nnamespace: "+ns+"\nversion: 1\nlabels: []\n")
	}
	if err := os.RemoveAll(old); err != nil {
		c.t.Fatal(err)
	}
	return dir
}

func writeFile(t *testing.T, p, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func init() {
	checkCases["stranded"] = func(c *cli) {
		c.link()
		strandFolder(c, "old", `{"status_writes": "command", "status_command": "studio task move"}`, "studio")
	}
	// Not stranded: a folder holding only its record; a recorded path that holds a bonsai.yaml with another id (that
	// checkout's own folder); and a folder whose name is not a machine folder's gets no removal line.
	checkMore["stranded"] = func(t *testing.T, c *cli) {
		_, out, _ := c.run("", "check", "--json")
		if strings.Count(out, `"code": "stranded"`) != 1 || !strings.Contains(out, "status_command studio task move") ||
			!strings.Contains(out, "the attached labels studio") || !strings.Contains(out, "this checkout has status_writes agents and no attached labels") ||
			!strings.Contains(out, `bonsai settings set status_writes=command \"status_command=studio task move\"`) ||
			!strings.Contains(out, "bonsai labels attach ") {
			t.Errorf("the warning:\n%s", out)
		}
		strandFolder(c, "bare", "")
		other := filepath.Join(c.tmp, "other")
		dir := strandFolder(c, "other", `{"status_writes": "command", "status_command": "x"}`)
		writeFile(t, filepath.Join(other, workspace.ConfigFile), "id: ws-0000000000000000000000000z\n")
		_, out, _ = c.run("", "check", "--json")
		if strings.Count(out, `"code": "stranded"`) != 1 || strings.Contains(out, filepath.ToSlash(dir)) {
			t.Errorf("a bare folder, or a path holding another id's bonsai.yaml, gave stranded:\n%s", out)
		}
		if err := os.RemoveAll(other); err != nil {
			t.Fatal(err)
		}
		odd := filepath.Join(filepath.Dir(dir), "r-not-a-hex-name")
		if err := os.Rename(dir, odd); err != nil {
			t.Fatal(err)
		}
		_, out, _ = c.run("", "check", "--json")
		if strings.Count(out, `"code": "stranded"`) != 2 || strings.Contains(out, "r-not-a-hex-name'") ||
			!strings.Contains(out, filepath.ToSlash(odd)+", is stranded") {
			t.Errorf("an odd folder's name: no removal line wanted:\n%s", out)
		}
		// status's text names it; status --json carries no warning.
		_, st, _ := c.run("", "status")
		if !strings.Contains(st, "stranded") || !strings.Contains(st, "bonsai check") {
			t.Errorf("status's text:\n%s", st)
		}
		_, sj, _ := c.run("", "status", "--json")
		if strings.Contains(sj, "is stranded") || strings.Contains(sj, "stranded under") {
			t.Errorf("status --json carries the warning:\n%s", sj)
		}
	}
}

// TestStrandedMoved moves a linked checkout for real and runs the person's lines as printed.
func TestStrandedMoved(t *testing.T) {
	c := newCLI(t)
	home := filepath.Join(c.tmp, "a home's o'name")
	t.Setenv("BONSAI_HOME", home)
	c.link()
	asPerson(c)
	if code, out, errOut := c.run("", "settings", "set", "status_writes=command", "status_command=studio task move"); code != exitOK {
		t.Fatalf("settings set: %d\n%s%s", code, out, errOut)
	}
	oldDir, _ := workspace.MachineDir(home, c.root)
	writeFile(t, filepath.Join(oldDir, "labels", "studio.yaml"), "format: bonsai.labels/1\nnamespace: studio\nversion: 1\nlabels: []\n")
	moved := filepath.Join(c.tmp, "moved")
	t.Chdir(c.tmp)
	if err := os.Rename(c.root, moved); err != nil {
		t.Fatal(err)
	}
	c.root = moved
	t.Chdir(moved)
	if code, out, errOut := c.run("", "update", "--yes"); code != exitOK {
		t.Fatalf("update in the moved checkout: %d\n%s%s", code, out, errOut)
	}
	sf := engine.Stranded(home, moved, configID(t, moved), runtime.GOOS)
	if len(sf) != 1 || sf[0].Command != "studio task move" || strings.Join(sf[0].Namespaces, ",") != "studio" || len(sf[0].Lines) != 3 {
		t.Fatalf("stranded: %+v", sf)
	}
	// Each path is quoted for its shell: a ' doubled for PowerShell, written '\'' for a POSIX shell.
	gone := filepath.ToSlash(oldDir)
	for goos, want := range map[string]string{
		"linux":   "rm -r '" + strings.ReplaceAll(gone, "'", `'\''`) + "'",
		"windows": "Remove-Item -Recurse -LiteralPath '" + strings.ReplaceAll(gone, "'", "''") + "'",
	} {
		lines := engine.Stranded(home, moved, configID(t, moved), goos)[0].Lines
		if lines[0] != `bonsai settings set status_writes=command "status_command=studio task move"` || lines[2] != want ||
			!strings.HasPrefix(lines[1], "bonsai labels attach '") {
			t.Errorf("%s's lines: %q", goos, lines)
		}
	}
	_, out, _ := c.run("", "check")
	if !strings.Contains(out, "stranded") || !strings.Contains(flat(out), "status_command studio task move") {
		t.Errorf("check's text:\n%s", out)
	}
	for _, line := range sf[0].Lines {
		if strings.HasPrefix(line, "bonsai ") {
			args := throughShell(t, line)
			if later := subLater(args[0], args[1]); later != "" {
				t.Logf("bonsai %s %s is not built in this build (%s): the line is not run", args[0], args[1], later)
				continue
			}
			var stdout, stderr bytes.Buffer
			if code := run(args, &stdout, &stderr); code != exitOK {
				t.Fatalf("%v: exit %d\n%s%s", args, code, stdout.String(), stderr.String())
			}
			continue
		}
		runShell := exec.Command("sh", "-c", line)
		if runtime.GOOS == "windows" {
			runShell = exec.Command("powershell.exe", "-NoProfile", "-NonInteractive", "-Command", line)
		}
		if out, err := runShell.CombinedOutput(); err != nil {
			t.Fatalf("%s: %v\n%s", line, err, out)
		}
	}
	if _, err := os.Stat(oldDir); err == nil {
		t.Errorf("the old folder is still there: %s", oldDir)
	}
	if s, err := workspace.LoadMachineSettings(home, moved); err != nil || s.StatusCommand != "studio task move" {
		t.Errorf("the settings were not brought over: %+v %v", s, err)
	}
	if _, out, _ := c.run("", "check", "--json"); strings.Contains(out, `"code": "stranded"`) {
		t.Errorf("the warning is still there:\n%s", out)
	}
}

func configID(t *testing.T, root string) string {
	t.Helper()
	cfg, err := workspace.LoadConfig(root)
	if err != nil {
		t.Fatal(err)
	}
	return cfg.ID
}

// throughShell gives the words the system's own shell reads a bonsai line as: a POSIX shell's, or PowerShell's on
// Windows, with bonsai a function printing its arguments.
func throughShell(t *testing.T, line string) []string {
	t.Helper()
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("powershell.exe", "-NoProfile", "-NonInteractive", "-Command",
			"function bonsai { [Console]::Out.Write('bonsai' + [char]0); foreach ($a in $args) { [Console]::Out.Write($a + [char]0) } }; "+line)
	} else {
		cmd = exec.Command("sh", "-c", `bonsai() { printf 'bonsai\0'; for a in "$@"; do printf '%s\0' "$a"; done; }; `+line)
	}
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("%s: %v", line, err)
	}
	args := strings.Split(strings.TrimSuffix(string(out), "\x00"), "\x00")
	if len(args) < 3 || args[0] != "bonsai" {
		t.Fatalf("%s: read as %q", line, args)
	}
	return args[1:]
}

// subLater is a sub-word's Later, "" when it is built (or not a sub-word).
func subLater(word, sub string) string {
	if w := words[word]; w != nil {
		for _, s := range w.Subs {
			if s.Name == word+" "+sub {
				return s.Later
			}
		}
	}
	return ""
}
