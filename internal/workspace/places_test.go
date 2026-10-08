package workspace

// Tests of the home, the machine key, the LF hash, paths, atomic writes and finding a checkout through git.

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"syscall"
	"testing"
	"time"
)

func jsonNumber(s string) json.Number { return json.Number(s) }

func TestHome(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv(HomeEnv, tmp)
	if h, err := Home(); err != nil || h != tmp {
		t.Errorf("Home() = %q, %v; want %q", h, err, tmp)
	}
	wd, _ := os.Getwd()
	t.Setenv(HomeEnv, "rel-home")
	if h, err := Home(); err != nil || h != filepath.Join(wd, "rel-home") {
		t.Errorf("a relative BONSAI_HOME gave %q, %v", h, err)
	}
	t.Setenv(HomeEnv, "")
	user := t.TempDir()
	t.Setenv("HOME", user)
	t.Setenv("USERPROFILE", user)
	if h, err := Home(); err != nil || h != filepath.Join(user, ".bonsai") {
		t.Errorf("with no BONSAI_HOME, Home() = %q, %v; want %q", h, err, filepath.Join(user, ".bonsai"))
	}
	if _, err := os.Stat(filepath.Join(user, ".bonsai")); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("Home made the folder")
	}
}

func TestMachineKey(t *testing.T) {
	key := regexp.MustCompile(`^r-[0-9a-f]{16}$`)
	dir := t.TempDir()
	k1, err := MachineKey(dir)
	if err != nil || !key.MatchString(k1) {
		t.Fatalf("MachineKey = %q, %v", k1, err)
	}
	real, _ := filepath.EvalSymlinks(dir)
	if k2, _ := MachineKey(real + string(filepath.Separator) + "."); k2 != k1 {
		t.Errorf("the same folder gave %s and %s", k1, k2)
	}
	sum := sha256.Sum256([]byte(filepath.ToSlash(real)))
	if runtime.GOOS == "windows" {
		sum = sha256.Sum256([]byte(strings.ToLower(filepath.ToSlash(real))))
	}
	if k1 != "r-"+hex.EncodeToString(sum[:])[:16] {
		t.Errorf("the key is not the first 16 hex of the real path's SHA-256")
	}
	if a, b := machineKey(`D:\Work\Proj`, "windows"), machineKey("d:/work/proj", "windows"); a != b {
		t.Errorf("on Windows, letter case and separators change the key: %s %s", a, b)
	}
	if a, b := machineKey("/srv/Proj", "linux"), machineKey("/srv/proj", "linux"); a == b {
		t.Errorf("on Linux, letter case does not change the key")
	}
	if _, err := MachineKey(filepath.Join(dir, "missing")); err == nil {
		t.Errorf("a missing folder gave a key")
	}
}

func TestHashLF(t *testing.T) {
	if HashLF(nil) != "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855" {
		t.Errorf("the hash of nothing is wrong")
	}
	if HashLF([]byte("a\r\nb\r\n")) != HashLF([]byte("a\nb\n")) {
		t.Errorf("CRLF and LF hash differently")
	}
	if HashLF([]byte("a\rb")) == HashLF([]byte("a\nb")) {
		t.Errorf("a lone CR is made LF")
	}
	if HashLF([]byte("a\n")) == HashLF([]byte("a")) {
		t.Errorf("a final line ending is lost")
	}
}

func TestCheckRelPath(t *testing.T) {
	for _, p := range []string{"a", "a/b.md", ".claude/settings.json", "test-pack/guide.md", "caf\xc3\xa9/x", "a.b/c-d_e",
		".github/workflows/ci.yml", "con-x/y", "x/aux_y.md", "git~2", "x/git~1x", "conin", "com10", "com\xc2\xb9x",
		"connect.txt", "x/.gitignore", "github"} {
		if err := CheckRelPath(p); err != nil {
			t.Errorf("%q refused: %v", p, err)
		}
	}
	for _, p := range []string{"", "/a", `a\b`, "C:/x", "a:b", "a//b", "a/", "./a", "a/../b", "..", ".git/x",
		"x/.GIT/y", "a<b", "a>b", "a?b", "a*", "a|b", `a"b`, "a.", "a ", "a /b", "con", "NUL.txt", "x/COM1.md",
		"lpt9", "a\x01b", "a\x7f", "\xff", "GIT~1/hooks/x", "x/git~1", "CONIN$", "conout$.log", "COM\xc2\xb9",
		"x/com\xc2\xb2.txt", "lpt\xc2\xb3", "con .txt", "x/nul  .md", "aux ", "PRN"} {
		if err := CheckRelPath(p); err == nil {
			t.Errorf("%q passed", p)
		}
	}
}

func TestWriteFileAtomic(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "made", "here")
	path := filepath.Join(dir, "f.json")
	if err := WriteFileAtomic(path, []byte("one")); err != nil {
		t.Fatal(err)
	}
	if err := WriteFileAtomic(path, []byte("two")); err != nil {
		t.Fatal(err)
	}
	if b, _ := os.ReadFile(path); string(b) != "two" {
		t.Errorf("read %q", b)
	}
	// A failed rename leaves the old file and no temporary one.
	defer func(r func(string, string) error) { rename = r }(rename)
	rename = func(string, string) error { return errors.New("no") }
	if err := WriteFileAtomic(path, []byte("three")); err == nil {
		t.Fatal("a failed rename was not reported")
	}
	entries, _ := os.ReadDir(dir)
	if b, _ := os.ReadFile(path); string(b) != "two" || len(entries) != 1 {
		t.Errorf("after a failed write: %q and %d files", b, len(entries))
	}
}

// renameRetry waits out a busy target, gives up on any other error at once, and stops after renameWait. The busy
// error is stood in, so this runs on every system; TestRenameRetryOnWindows meets a real one.
func TestRenameRetry(t *testing.T) {
	defer func(r func(string, string) error, b func(error) bool, w time.Duration) {
		rename, busy, renameWait = r, b, w
	}(rename, busy, renameWait)
	errBusy := errors.New("busy")
	busy = func(err error) bool { return err == errBusy }
	calls := 0
	rename = func(string, string) error {
		calls++
		if calls < 4 {
			return errBusy
		}
		return nil
	}
	if err := renameRetry("a", "b"); err != nil || calls != 4 {
		t.Errorf("busy three times: %v after %d calls", err, calls)
	}
	calls = 0
	other := errors.New("other")
	rename = func(string, string) error { calls++; return other }
	if err := renameRetry("a", "b"); err != other || calls != 1 {
		t.Errorf("another error: %v after %d calls", err, calls)
	}
	renameWait = 30 * time.Millisecond
	rename = func(string, string) error { return errBusy }
	start := time.Now()
	if err := renameRetry("a", "b"); err != errBusy || time.Since(start) > 2*time.Second {
		t.Errorf("busy forever: %v after %v", err, time.Since(start))
	}
}

// A target another process holds open: on Windows a rename over it fails busy until it is closed, and the write
// waits for it. Linux and macOS replace an open file, so there the write succeeds at once.
func TestRenameRetryOnWindows(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "held.json")
	if err := os.WriteFile(path, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	defer func(r func(string, string) error) { rename = r }(rename)
	busyTries := 0
	rename = func(from, to string) error {
		err := os.Rename(from, to)
		if err != nil && isBusy(err) {
			busyTries++
		}
		return err
	}
	held, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		time.Sleep(150 * time.Millisecond)
		_ = held.Close()
	}()
	if err := WriteFileAtomic(path, []byte("new")); err != nil {
		t.Fatalf("the write did not wait out the held file: %v", err)
	}
	if b, _ := os.ReadFile(path); string(b) != "new" {
		t.Errorf("read %q", b)
	}
	t.Logf("renames refused as busy while the file was held: %d", busyTries)
	if runtime.GOOS == "windows" && busyTries == 0 {
		t.Errorf("Windows let a rename replace an open file: the busy path was not met")
	}
	if runtime.GOOS == "windows" && !isBusy(&os.LinkError{Err: syscall.Errno(32)}) {
		t.Errorf("a sharing violation is not busy")
	}
}

// git runs git in dir with no user or system configuration, failing the test on an error.
func git(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir, "-c", "user.name=bonsai-test", "-c", "user.email=bonsai-test",
		"-c", "commit.gpgsign=false", "-c", "init.defaultBranch=main"}, args...)...)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

// isolateGit keeps a test's git runs off the machine's configuration and stops git looking above tmp.
func isolateGit(t *testing.T, tmp string) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not on the PATH: finding a checkout needs it")
	}
	empty := filepath.Join(t.TempDir(), "gitconfig")
	if err := os.WriteFile(empty, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GIT_CONFIG_GLOBAL", empty)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	t.Setenv("GIT_CEILING_DIRECTORIES", filepath.Dir(tmp))
}

func TestFind(t *testing.T) {
	tmp := t.TempDir()
	isolateGit(t, tmp)
	if _, err := Find(tmp); err == nil || !strings.Contains(err.Error(), "is not inside a git checkout") ||
		!strings.Contains(err.Error(), "next: ") {
		t.Errorf("outside git: %v", err)
	}
	main := filepath.Join(tmp, "main")
	if err := os.MkdirAll(filepath.Join(main, "sub", "deeper"), 0o755); err != nil {
		t.Fatal(err)
	}
	git(t, main, "init", "-q")
	realMain, _ := filepath.EvalSymlinks(main)
	for _, from := range []string{main, filepath.Join(main, "sub", "deeper")} {
		c, err := Find(from)
		if err != nil {
			t.Fatal(err)
		}
		if c.Root != realMain || c.Main != realMain {
			t.Errorf("from %s: %+v, want both %s", from, c, realMain)
		}
	}
	if _, err := Find(filepath.Join(main, ".git")); err == nil {
		t.Errorf("inside .git: no error")
	}
	// A worktree: its own root, and the main checkout it was added from.
	if err := os.WriteFile(filepath.Join(main, "f.txt"), []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	git(t, main, "add", "f.txt")
	git(t, main, "commit", "-q", "-m", "one")
	wt := filepath.Join(tmp, "main-wt")
	git(t, main, "worktree", "add", "-q", wt)
	c, err := Find(wt)
	if err != nil {
		t.Fatal(err)
	}
	realWT, _ := filepath.EvalSymlinks(wt)
	if c.Root != realWT || c.Main != realMain {
		t.Errorf("in a worktree: %+v, want root %s and main %s", c, realWT, realMain)
	}
	if strings.Contains(filepath.ToSlash(c.Root), `\`) {
		t.Errorf("a stored path holds a backslash")
	}
}

func TestFindWithoutGit(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	_, err := Find(t.TempDir())
	if err == nil || !strings.Contains(err.Error(), "git: is not on the PATH") || !strings.Contains(err.Error(), "next: install git") {
		t.Errorf("no git: %v", err)
	}
}
