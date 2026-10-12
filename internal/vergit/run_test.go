package vergit

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The exported runner runs git as Judge does: the caller's GIT_* variables cannot point it at another repository or
// add settings, the user's line-end settings are kept, extra -c settings may lead the arguments, and ExitCode reads a
// failed call's exit.
func TestRunner(t *testing.T) {
	home := isolate(t)
	root, base := repo(t)
	other, _ := repo(t)
	git(t, other, "commit", "-q", "--allow-empty", "-m", "other")
	global := filepath.Join(home, "gitconfig")
	must(t, os.WriteFile(global, []byte("[core]\n\tautocrlf = input\n"), 0o644))
	t.Setenv("GIT_DIR", filepath.Join(other, ".git"))
	t.Setenv("GIT_WORK_TREE", other)
	t.Setenv("GIT_CONFIG_PARAMETERS", "'run.test'='steered'")

	r, note, err := NewRunner(context.Background(), "", root)
	if err != nil || note != "" {
		t.Fatalf("NewRunner: %v, note %q", err, note)
	}
	out, err := r.Run("rev-parse", "HEAD~1")
	if err != nil || strings.TrimSpace(string(out)) != base {
		t.Errorf("rev-parse HEAD~1 = %q, %v; want %s (the checkout's own, not GIT_DIR's)", out, err, base)
	}
	if out, err := r.Run("config", "core.autocrlf"); err != nil || strings.TrimSpace(string(out)) != "input" {
		t.Errorf("core.autocrlf = %q, %v; want the user's input", out, err)
	}
	if _, err := r.Run("config", "run.test"); ExitCode(err) != 1 {
		t.Errorf("GIT_CONFIG_PARAMETERS reached git: %v", err)
	}
	if out, err := r.Run("-c", "run.test=mine", "config", "run.test"); err != nil || strings.TrimSpace(string(out)) != "mine" {
		t.Errorf("a leading -c setting: %q, %v", out, err)
	}
	_, err = r.Run("cat-file", "-e", strings.Repeat("0", 40)+"^{commit}")
	if c := ExitCode(err); c != 128 && c != 1 {
		t.Errorf("cat-file of a missing commit: exit %d (%v)", c, err)
	}
	if ExitCode(errors.New("not git's")) != -1 || ExitCode(nil) != -1 {
		t.Errorf("ExitCode of an error that is not a Run's is not -1")
	}

	ctx, cancel := context.WithCancel(context.Background())
	r, _, err = NewRunner(ctx, "", root)
	if err != nil {
		t.Fatal(err)
	}
	cancel()
	if _, err := r.Run("status"); err == nil || ExitCode(err) != -1 {
		t.Errorf("a call after its context ended: %v, exit %d", err, ExitCode(err))
	}
	if _, _, err := NewRunner(context.Background(), filepath.Join(t.TempDir(), "no-git"), root); err == nil {
		t.Errorf("a git program that is not there gave a runner")
	}
}
