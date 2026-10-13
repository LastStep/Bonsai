package stopgate

// CommittedConfig, read directly: the committed bonsai.yaml at the main checkout's HEAD when the file there is the
// same (line ends read as LF), and a refusal naming why when it is not: changed, gone, or HEAD holding none.

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/LastStep/Bonsai/internal/workspace"
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

// Accepted, read directly: the committed settings the same as, or stricter than, the ones this machine's record accepted
// pass; looser ones are refused, each loosening named (cut at eight, the rest counted), with the person's step; a record
// with none, or none at all, fails closed; and it writes nothing.
func TestAccepted(t *testing.T) {
	root := project(t)
	home := os.Getenv(workspace.HomeEnv)
	read := func(yaml string) *workspace.Config {
		t.Helper()
		c, err := workspace.ReadConfigFull([]byte(yaml))
		if err != nil {
			t.Fatal(err)
		}
		return c
	}
	if le := Accepted(home, root, read(testYAML)); le != nil {
		t.Errorf("the same settings: %v", le)
	}
	if le := Accepted(home, root, read(strings.Replace(testYAML, "ladder_floor: [0, 1]", "ladder_floor: [0, 1, 2]", 1))); le != nil {
		t.Errorf("stricter settings: %v", le)
	}
	dir, _ := workspace.MachineDir(home, root)
	before, _ := os.ReadFile(filepath.Join(dir, workspace.MachineRecordFile))
	le := Accepted(home, root, read(strings.Replace(testYAML, "ladder_floor: [0, 1]", "ladder_floor: []", 1)))
	if le == nil || le.Why != "the main checkout's bonsai.yaml, as committed at its HEAD, sets the checks looser than the settings this "+
		"machine last accepted for the workspace: ladder_floor no longer holds rung 0; ladder_floor no longer holds rung 1" ||
		le.Next != LooserNext || le.Error() != le.Why {
		t.Errorf("looser: %+v", le)
	}
	if after, _ := os.ReadFile(filepath.Join(dir, workspace.MachineRecordFile)); string(after) != string(before) {
		t.Error("Accepted wrote the record")
	}
	marks := strings.Replace(testYAML, "ci_marked_tests: []", `ci_marked_tests: ["A", "B", "C", "D", "E", "F", "G", "H", "I", "J"]`, 1)
	if le := Accepted(home, root, read(marks)); le == nil || !strings.HasSuffix(le.Why, `ci_marked_tests gained "H"; and 2 more`) {
		t.Errorf("ten loosenings: %+v", le)
	}
	if le := Accepted(home, root, nil); le == nil || !strings.Contains(le.Why, "was not read in full") {
		t.Errorf("no settings read: %+v", le)
	}
	if le := Accepted(t.TempDir(), root, read(testYAML)); le == nil || le.Why != "this machine has no record of the main checkout" {
		t.Errorf("no record: %+v", le)
	}
}
