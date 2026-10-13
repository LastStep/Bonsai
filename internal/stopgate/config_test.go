package stopgate

// CommittedConfig, read directly: the committed bonsai.yaml at the main checkout's HEAD when the file there is the
// same (line ends read as LF), and a refusal naming why when it is not: changed, gone, or HEAD holding none.

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCommittedConfig(t *testing.T) {
	refused := func(t *testing.T, root, why string) {
		t.Helper()
		_, err := CommittedConfig(root, time.Minute)
		var u *UncommittedError
		if !errors.As(err, &u) || !strings.Contains(u.Why, why) {
			t.Errorf("err %v, want an UncommittedError saying %q", err, why)
		}
	}
	t.Run("the same file", func(t *testing.T) {
		root := project(t)
		cfg, err := CommittedConfig(root, time.Minute)
		if err != nil || cfg.ID != testID || len(cfg.Full.LadderFloor) != 2 {
			t.Fatalf("%+v %v", cfg, err)
		}
	})
	t.Run("read from the commit, not the file", func(t *testing.T) {
		root := project(t)
		committed := strings.Replace(testYAML, "ladder_floor: [0, 1]", "ladder_floor: [0, 1, 2]", 1)
		write(t, root, ".git/refs/heads/main", commitYAML(t, filepath.Join(root, ".git"), committed)+"\n")
		refused(t, root, "differs from the one committed at its HEAD")
		write(t, root, "bonsai.yaml", strings.ReplaceAll(committed, "\n", "\r\n"))
		cfg, err := CommittedConfig(root, time.Minute)
		if err != nil || len(cfg.Full.LadderFloor) != 3 {
			t.Fatalf("%+v %v", cfg, err)
		}
	})
	t.Run("changed", func(t *testing.T) {
		root := project(t)
		write(t, root, "bonsai.yaml", strings.Replace(testYAML, "ladder_floor: [0, 1]", "ladder_floor: [0]", 1))
		refused(t, root, "the main checkout's bonsai.yaml differs from the one committed at its HEAD")
	})
	t.Run("gone", func(t *testing.T) {
		root := project(t)
		remove(t, root, "bonsai.yaml")
		refused(t, root, "the main checkout's bonsai.yaml is gone, though its HEAD holds one")
	})
	t.Run("no commit", func(t *testing.T) {
		root := project(t)
		remove(t, root, ".git/refs/heads/main")
		refused(t, root, "the main checkout's HEAD names no commit yet")
	})
	t.Run("not in HEAD's commit", func(t *testing.T) {
		root := project(t)
		write(t, root, ".git/refs/heads/main", emptyCommit(t, filepath.Join(root, ".git"))+"\n")
		refused(t, root, "the main checkout's HEAD holds no bonsai.yaml (it was never committed there)")
	})
	t.Run("git cannot read the checkout", func(t *testing.T) {
		root := project(t)
		write(t, root, ".git/HEAD", "not a ref\n")
		_, err := CommittedConfig(root, time.Minute)
		var u *UncommittedError
		if err == nil || errors.As(err, &u) || !strings.Contains(err.Error(), "bonsai.yaml cannot be read at the main checkout's HEAD") {
			t.Errorf("err %v", err)
		}
	})
}
