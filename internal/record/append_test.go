package record

// Tests of the append path: eight processes appending at once (the test binary started again as each, through
// TestMain), a file held busy and then let go, the busy retry's limits, what Append refuses, and the .gitignore every
// writer restores.

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// appendChildEnv, when set, makes the test binary one of TestAppendsFromEightProcesses' processes. Its value is
// "<start file>|<main>|<child>|<records>": the child waits for the start file, appends its records to one session's
// log in main, and prints how many of its appends made the file.
const appendChildEnv = "BONSAI_TEST_APPEND_CHILD"

func TestMain(m *testing.M) {
	if spec := os.Getenv(appendChildEnv); spec != "" {
		os.Exit(appendChild(spec))
	}
	os.Exit(m.Run())
}

const (
	raceWorkspace = "ws-aaaaaaaaaaaaaaaaaaaaaaaaaa"
	raceSession   = "6d1e2f3a-4b5c-4d6e-8f7a-9b0c1d2e3f4a"
)

func appendChild(spec string) int {
	parts := strings.Split(spec, "|")
	if len(parts) != 4 {
		fmt.Fprintln(os.Stderr, "bad spec", spec)
		return 3
	}
	start, main := parts[0], parts[1]
	child, _ := strconv.Atoi(parts[2])
	n, _ := strconv.Atoi(parts[3])
	deadline := time.Now().Add(30 * time.Second)
	for {
		if _, err := os.Stat(start); err == nil {
			break
		}
		if time.Now().After(deadline) {
			fmt.Fprintln(os.Stderr, "the start file never came")
			return 3
		}
		time.Sleep(time.Millisecond)
	}
	made := 0
	for i := 0; i < n; i++ {
		l := New("event", Common{Workspace: raceWorkspace, Local: workspace.Local{Root: main, Main: main, Branch: "main"},
			Session: raceSession, Agent: "claude-code", Getenv: func(string) string { return "" }})
		target := fmt.Sprintf("child-%d/record-%d", child, i)
		l.Target = &target
		// Lines from about 400 to about 2,000 bytes: a character outside ASCII is written as six bytes (\uXXXX).
		text := strings.Repeat("é", (i*37+child*11)%260+40)
		l.Text = &text
		m, err := WriteLog(main, l)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		if m {
			made++
		}
	}
	fmt.Print(made)
	return 0
}

// Eight processes, started at once, each append 250 records to one session's file: on Linux and on Windows every one
// of the 2,000 lines is whole (it reads as a record, ends in its line feed, and fits the log's 2,048 bytes), every id
// is new, each process's records are all there in its own order, and exactly one append made the file.
func TestAppendsFromEightProcesses(t *testing.T) {
	const children, each = 8, 250
	main := t.TempDir()
	t.Setenv(workspace.HomeEnv, t.TempDir())
	start := filepath.Join(t.TempDir(), "start")
	cmds := make([]*exec.Cmd, children)
	outs := make([]*bytes.Buffer, children)
	errs := make([]*bytes.Buffer, children)
	for i := range cmds {
		cmd := exec.Command(os.Args[0], "-test.run=^$")
		cmd.Env = append(os.Environ(), fmt.Sprintf("%s=%s|%s|%d|%d", appendChildEnv, start, main, i, each))
		outs[i], errs[i] = &bytes.Buffer{}, &bytes.Buffer{}
		cmd.Stdout, cmd.Stderr = outs[i], errs[i]
		if err := cmd.Start(); err != nil {
			t.Fatal(err)
		}
		cmds[i] = cmd
	}
	began := time.Now()
	if err := os.WriteFile(start, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	made := 0
	for i, cmd := range cmds {
		if err := cmd.Wait(); err != nil {
			t.Fatalf("process %d: %v: %s", i, err, errs[i])
		}
		m, err := strconv.Atoi(outs[i].String())
		if err != nil {
			t.Fatalf("process %d printed %q", i, outs[i])
		}
		made += m
	}
	t.Logf("%d processes wrote %d records in %v", children, children*each, time.Since(began).Round(time.Millisecond))
	if made != 1 {
		t.Errorf("%d appends made the file, want exactly 1", made)
	}
	path := filepath.Join(main, ".bonsai", "local", "log", "s-"+raceSession+".ndjson")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	lines := bytes.SplitAfter(raw, []byte("\n"))
	if last := lines[len(lines)-1]; len(last) != 0 {
		t.Errorf("the file does not end in a line feed: %q", last)
	}
	lines = lines[:len(lines)-1]
	if len(lines) != children*each {
		t.Fatalf("%d lines, want %d", len(lines), children*each)
	}
	longest := 0
	for i, line := range lines {
		if len(line) > format.MustLookup("log").MaxLine() {
			t.Errorf("line %d is %d bytes", i+1, len(line))
		}
		longest = max(longest, len(line))
	}
	t.Logf("the longest line: %d bytes", longest)
	f, err := ReadLog(path)
	if err != nil {
		t.Fatal(err)
	}
	if f.Skipped != 0 || len(f.Records) != children*each {
		t.Fatalf("%d records read, %d lines skipped", len(f.Records), f.Skipped)
	}
	ids := map[string]bool{}
	next := make([]int, children)
	for _, r := range f.Records {
		if ids[r.ID] {
			t.Errorf("the id %s twice", r.ID)
		}
		ids[r.ID] = true
		var child, i int
		if _, err := fmt.Sscanf(*r.Target, "child-%d/record-%d", &child, &i); err != nil || child < 0 || child >= children {
			t.Fatalf("a record's target %q", *r.Target)
		}
		if i != next[child] {
			t.Errorf("process %d: record %d came where %d was due", child, i, next[child])
		}
		next[child] = i + 1
	}
	for c, n := range next {
		if n != each {
			t.Errorf("process %d: %d records, want %d", c, n, each)
		}
	}
	if b, err := os.ReadFile(filepath.Join(main, ".bonsai", ".gitignore")); err != nil || string(b) != workspace.GitignoreText {
		t.Errorf("the .gitignore: %q, %v", b, err)
	}
}

// A file another process holds open: on Windows, with no write sharing (holdFile), the append's open is refused as
// busy until the file is let go, and the append waits for it; Linux and macOS have no such hold, so there the append
// succeeds at once. The file is let go at the first busy refusal, or when the append returns where none comes, so
// nothing here can hang, as in internal/workspace's TestRenameRetryOnWindows.
func TestAppendWaitsOutAHeldFile(t *testing.T) {
	main := t.TempDir()
	rel := "log/w-2026-10-10.ndjson"
	path := filepath.Join(main, ".bonsai", "local", filepath.FromSlash(rel))
	if _, err := Append(main, rel, []byte("{\"n\":1}\n"), 2048); err != nil {
		t.Fatal(err)
	}
	defer func(o func(string, int, os.FileMode) (*os.File, error)) { openFile = o }(openFile)
	busyTries := 0
	refused := make(chan struct{})
	openFile = func(name string, flag int, perm os.FileMode) (*os.File, error) {
		f, err := os.OpenFile(name, flag, perm)
		if err != nil && busy(err) {
			busyTries++
			if busyTries == 1 {
				close(refused)
			}
		}
		return f, err
	}
	release := holdFile(t, path)
	done := make(chan struct{})
	released := make(chan struct{})
	go func() {
		select {
		case <-refused:
		case <-done:
		}
		release()
		close(released)
	}()
	made, err := Append(main, rel, []byte("{\"n\":2}\n"), 2048)
	close(done)
	<-released
	if err != nil || made {
		t.Fatalf("the append did not wait out the held file: made %v, %v", made, err)
	}
	if b, _ := os.ReadFile(path); string(b) != "{\"n\":1}\n{\"n\":2}\n" {
		t.Errorf("the file holds %q", b)
	}
	t.Logf("opens refused as busy: %d", busyTries)
	if runtime.GOOS == "windows" && busyTries == 0 {
		t.Errorf("Windows let the append open a file held with no write sharing: the busy path was not met")
	}
}

// The busy retry waits out a busy open, gives up on any other error at once, and stops after openWait. The busy
// error is stood in, so this runs on every system; TestAppendWaitsOutAHeldFile meets a real one on Windows.
func TestOpenRetry(t *testing.T) {
	defer func(o func(string, int, os.FileMode) (*os.File, error), b func(error) bool, w time.Duration) {
		openFile, busy, openWait = o, b, w
	}(openFile, busy, openWait)
	errBusy := errors.New("busy")
	busy = func(err error) bool { return err == errBusy }
	path := filepath.Join(t.TempDir(), "f")
	calls := 0
	openFile = func(name string, flag int, perm os.FileMode) (*os.File, error) {
		calls++
		if calls < 4 {
			return nil, errBusy
		}
		return os.OpenFile(name, flag, perm)
	}
	f, err := openRetry(path, os.O_WRONLY|os.O_CREATE)
	if err != nil || calls != 4 {
		t.Errorf("busy three times: %v after %d calls", err, calls)
	} else {
		_ = f.Close()
	}
	other := errors.New("other")
	calls = 0
	openFile = func(string, int, os.FileMode) (*os.File, error) { calls++; return nil, other }
	if _, err := openRetry(path, os.O_WRONLY); err != other || calls != 1 {
		t.Errorf("another error: %v after %d calls", err, calls)
	}
	openWait = 30 * time.Millisecond
	openFile = func(string, int, os.FileMode) (*os.File, error) { return nil, errBusy }
	began := time.Now()
	if _, err := openRetry(path, os.O_WRONLY); err != errBusy || time.Since(began) > 2*time.Second {
		t.Errorf("busy forever: %v after %v", err, time.Since(began))
	}
	// Through Append, a file busy for longer than the wait is an error, and nothing is written.
	main := t.TempDir()
	if _, err := Append(main, "log/x.ndjson", []byte("{}\n"), 2048); !errors.Is(err, errBusy) {
		t.Errorf("Append on a file busy for good: %v", err)
	}
	if _, err := os.Stat(filepath.Join(main, ".bonsai", "local", "log", "x.ndjson")); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("a file was made: %v", err)
	}
}

// Append writes only one whole record line, within its cap, inside .bonsai/local/; it says when it made the file.
func TestAppend(t *testing.T) {
	main := t.TempDir()
	line := []byte("{\"a\":1}\n")
	for _, c := range []struct {
		name, rel string
		line      []byte
		max       int
	}{
		{"no line feed", "log/a.ndjson", []byte("{}"), 100},
		{"two lines", "log/a.ndjson", []byte("{}\n{}\n"), 100},
		{"a line feed alone", "log/a.ndjson", []byte("\n"), 100},
		{"a line feed inside", "log/a.ndjson", []byte("{\n}\n"), 100},
		{"over the cap", "log/a.ndjson", line, len(line) - 1},
		{"no cap", "log/a.ndjson", line, 0},
		{"outside local/", "../x.ndjson", line, 100},
		{"an absolute path", "/x.ndjson", line, 100},
		{"a backslash", `log\a.ndjson`, line, 100},
		{"empty", "", line, 100},
	} {
		if _, err := Append(main, c.rel, c.line, c.max); err == nil {
			t.Errorf("%s: written", c.name)
		}
	}
	if _, err := os.Stat(filepath.Join(main, ".bonsai")); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("a refused line made .bonsai: %v", err)
	}
	if made, err := Append(main, "asks/2026-10-10.ndjson", line, len(line)); err != nil || !made {
		t.Errorf("the first line, at the cap: made %v, %v", made, err)
	}
	if made, err := Append(main, "asks/2026-10-10.ndjson", line, len(line)); err != nil || made {
		t.Errorf("the second line: made %v, %v", made, err)
	}
	if b, _ := os.ReadFile(filepath.Join(main, ".bonsai", "local", "asks", "2026-10-10.ndjson")); !bytes.Equal(b, append(line, line...)) {
		t.Errorf("the file holds %q", b)
	}
}

// Every writer restores a missing .bonsai/.gitignore before its line (spec §6), leaves a changed one alone (check
// finds it), and still writes its line when the .gitignore cannot be written, saying so.
func TestAppendRestoresTheGitignore(t *testing.T) {
	main := t.TempDir()
	gi := filepath.Join(main, ".bonsai", ".gitignore")
	l := New("event", Common{Workspace: raceWorkspace, Local: workspace.Local{Root: main, Main: main}})
	if _, err := WriteLog(main, l); err != nil {
		t.Fatal(err)
	}
	if b, err := os.ReadFile(gi); err != nil || string(b) != workspace.GitignoreText {
		t.Fatalf("after the first record: %q, %v", b, err)
	}
	if err := os.Remove(gi); err != nil {
		t.Fatal(err)
	}
	if _, err := WriteLog(main, New("event", Common{Workspace: raceWorkspace})); err != nil {
		t.Fatal(err)
	}
	if b, err := os.ReadFile(gi); err != nil || string(b) != workspace.GitignoreText {
		t.Errorf("deleted, then a record: %q, %v", b, err)
	}
	if err := os.WriteFile(gi, []byte("mine\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := WriteLog(main, New("event", Common{Workspace: raceWorkspace})); err != nil {
		t.Fatal(err)
	}
	if b, _ := os.ReadFile(gi); string(b) != "mine\n" {
		t.Errorf("a changed .gitignore was written over: %q", b)
	}
	// A .gitignore that cannot be written: the line is still written, and the error names the .gitignore.
	defer func(e func(string) (bool, error)) { ensureGitignore = e }(ensureGitignore)
	ensureGitignore = func(string) (bool, error) { return false, errors.New("the disk is full") }
	other := t.TempDir()
	made, err := Append(other, "log/w-2026-10-10.ndjson", []byte("{}\n"), 2048)
	if !made || err == nil || !strings.Contains(err.Error(), ".bonsai/.gitignore cannot be restored: the disk is full") {
		t.Errorf("a .gitignore that cannot be written: made %v, %v", made, err)
	}
	if b, _ := os.ReadFile(filepath.Join(other, ".bonsai", "local", "log", "w-2026-10-10.ndjson")); string(b) != "{}\n" {
		t.Errorf("the line was not written: %q", b)
	}
}
