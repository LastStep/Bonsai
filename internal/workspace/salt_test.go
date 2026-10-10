package workspace

// Tests of the home's salt: made at first need, its shape, its mode, the loser of a race reading the winner's, and
// eight processes racing to make it (the test binary started again as each, through TestMain).

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// saltChildEnv, when set, makes the test binary one of TestSaltRace's processes: it waits for the file the variable
// names, then prints Salt()'s answer and exits.
const saltChildEnv = "BONSAI_TEST_SALT_CHILD"

func TestMain(m *testing.M) {
	if start := os.Getenv(saltChildEnv); start != "" {
		os.Exit(saltChild(start))
	}
	os.Exit(m.Run())
}

func saltChild(start string) int {
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
	s, err := Salt()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	fmt.Print(s)
	return 0
}

// isSalt reports whether s has the salt's shape: 64 lower-case hex characters.
func isSalt(s string) bool { return len(s) == 64 && strings.Trim(s, "0123456789abcdef") == "" }

// The first call makes the salt, whole, with no line ending; later calls read the same one; nothing else is left in
// the home. Mode 0600 on Linux and macOS (Windows has no such mode; the home's permissions stand).
func TestSalt(t *testing.T) {
	home := filepath.Join(t.TempDir(), "home")
	t.Setenv(HomeEnv, home)
	s, err := Salt()
	if err != nil || !isSalt(s) {
		t.Fatalf("Salt() = %q, %v", s, err)
	}
	raw, err := os.ReadFile(filepath.Join(home, SaltFile))
	if err != nil || string(raw) != s {
		t.Errorf("the file holds %q, %v", raw, err)
	}
	if again, err := Salt(); err != nil || again != s {
		t.Errorf("a second call gave %q, %v", again, err)
	}
	if entries, _ := os.ReadDir(home); len(entries) != 1 {
		t.Errorf("the home holds %d entries, want only the salt", len(entries))
	}
	if runtime.GOOS != "windows" {
		if fi, err := os.Stat(filepath.Join(home, SaltFile)); err != nil || fi.Mode().Perm() != 0o600 {
			t.Errorf("the salt's mode is %v, %v; want 0600", fi.Mode().Perm(), err)
		}
	}
	other := t.TempDir()
	t.Setenv(HomeEnv, other)
	if s2, err := Salt(); err != nil || s2 == s {
		t.Errorf("another home gave %q, %v: the same salt", s2, err)
	}
}

// Of two first writers, the one whose link finds the salt already in place reads the winner's and says nothing.
func TestSaltLoserReadsTheWinners(t *testing.T) {
	home := t.TempDir()
	t.Setenv(HomeEnv, home)
	winner := strings.Repeat("ab", 32)
	defer func(l func(string, string) error) { link = l }(link)
	link = func(from, to string) error {
		if err := os.WriteFile(to, []byte(winner), 0o600); err != nil {
			return err
		}
		return os.Link(from, to)
	}
	if s, err := Salt(); err != nil || s != winner {
		t.Errorf("the loser read %q, %v; want the winner's", s, err)
	}
	if entries, _ := os.ReadDir(home); len(entries) != 1 {
		t.Errorf("the loser left its temporary file: %d entries", len(entries))
	}
	// A link that fails for another reason is an error naming the home, and nothing is left behind.
	link = func(string, string) error { return errors.New("no links here") }
	t.Setenv(HomeEnv, t.TempDir())
	if _, err := Salt(); err == nil || !strings.Contains(err.Error(), "cannot be made") || !strings.Contains(err.Error(), "next: ") {
		t.Errorf("a failed link: %v", err)
	}
}

// A salt file of the wrong shape is refused, never used or written over; a final line ending is let through. A home
// that cannot be made is an error with the next step.
func TestSaltRefusesABadFile(t *testing.T) {
	home := t.TempDir()
	t.Setenv(HomeEnv, home)
	path := filepath.Join(home, SaltFile)
	good := strings.Repeat("0f", 32)
	for _, text := range []string{good + "\n", good + "\r\n"} {
		if err := os.WriteFile(path, []byte(text), 0o600); err != nil {
			t.Fatal(err)
		}
		if s, err := Salt(); err != nil || s != good {
			t.Errorf("%q: %q, %v", text, s, err)
		}
	}
	for _, text := range []string{"", "short", strings.ToUpper(good), good + "00", good[:63] + "g", good + "\n\n"} {
		if err := os.WriteFile(path, []byte(text), 0o600); err != nil {
			t.Fatal(err)
		}
		_, err := Salt()
		if err == nil || !strings.Contains(err.Error(), "is not 64 lower-case hex characters") || !strings.Contains(err.Error(), "next: delete") {
			t.Errorf("%q: %v", text, err)
		}
		if b, _ := os.ReadFile(path); string(b) != text {
			t.Errorf("%q was written over: %q", text, b)
		}
	}
	file := filepath.Join(t.TempDir(), "a-file")
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv(HomeEnv, file)
	if _, err := Salt(); err == nil || !strings.Contains(err.Error(), "next: ") || errors.Is(err, fs.ErrNotExist) {
		t.Errorf("a home that is a file: %v", err)
	}
}

// Eight processes, started at once with an empty home, all make the salt: one salt, which every one of them reads,
// and nothing left beside it.
func TestSaltRace(t *testing.T) {
	home := t.TempDir()
	start := filepath.Join(t.TempDir(), "start")
	const n = 8
	cmds := make([]*exec.Cmd, n)
	outs := make([]*strings.Builder, n)
	for i := range cmds {
		cmd := exec.Command(os.Args[0], "-test.run=^$")
		cmd.Env = append(os.Environ(), saltChildEnv+"="+start, HomeEnv+"="+home)
		outs[i] = &strings.Builder{}
		cmd.Stdout, cmd.Stderr = outs[i], outs[i]
		if err := cmd.Start(); err != nil {
			t.Fatal(err)
		}
		cmds[i] = cmd
	}
	if err := os.WriteFile(start, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	for i, cmd := range cmds {
		if err := cmd.Wait(); err != nil {
			t.Errorf("process %d: %v: %s", i, err, outs[i])
		}
	}
	raw, err := os.ReadFile(filepath.Join(home, SaltFile))
	if err != nil || !isSalt(string(raw)) {
		t.Fatalf("the salt: %q, %v", raw, err)
	}
	for i, out := range outs {
		if out.String() != string(raw) {
			t.Errorf("process %d read %q, the file holds %q", i, out.String(), raw)
		}
	}
	if entries, _ := os.ReadDir(home); len(entries) != 1 {
		names := []string{}
		for _, e := range entries {
			names = append(names, e.Name())
		}
		t.Errorf("the home holds %v, want only the salt", names)
	}
}
