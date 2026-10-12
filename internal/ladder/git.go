package ladder

// The git reads a climb makes (note 6, note 13): HEAD, its branch and the tree's status at the start and the end, and
// the merge base. Each is one git process in the checkout climbed, with git's own variables that would point it at
// another repository taken out of its environment. A read that fails gives "" (or dirty), never a guess that the tree
// is clean.

import (
	"bytes"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/LastStep/Bonsai/internal/format"
)

// gitDrop reports the variables a git read leaves out, so it reads the checkout it is pointed at.
func gitDrop(k string) bool {
	switch k {
	case "GIT_DIR", "GIT_WORK_TREE", "GIT_INDEX_FILE", "GIT_OBJECT_DIRECTORY", "GIT_ALTERNATE_OBJECT_DIRECTORIES",
		"GIT_COMMON_DIR", "GIT_NAMESPACE", "GIT_TERMINAL_PROMPT", "GIT_OPTIONAL_LOCKS":
		return true
	}
	return false
}

// git runs git in dir and gives its standard output; ok is false when it fails.
func git(dir string, args ...string) (string, bool) {
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	var env []string
	for _, kv := range os.Environ() {
		k := kv
		if i := strings.IndexByte(kv, '='); i >= 0 {
			k = kv[:i]
		}
		if runtime.GOOS == "windows" {
			k = strings.ToUpper(k)
		}
		if !gitDrop(k) {
			env = append(env, kv)
		}
	}
	cmd.Env = append(env, "GIT_TERMINAL_PROMPT=0", "GIT_OPTIONAL_LOCKS=0")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", false
	}
	return out.String(), true
}

// gitState reads HEAD (the full lower-case SHA) and its branch (HEAD when detached, as git rev-parse --abbrev-ref
// says); dirty is the caller's (kinds.go's treeDirty).
func gitState(dir string) format.LadderGit {
	g := format.LadderGit{SHA: head(dir)}
	if b, ok := git(dir, "rev-parse", "--abbrev-ref", "HEAD"); ok {
		g.Branch = strings.TrimSpace(b)
	}
	return g
}

// head is HEAD's commit, lower-case, "" when it cannot be read (no commit yet).
func head(dir string) string {
	s, ok := git(dir, "rev-parse", "--verify", "-q", "HEAD")
	if !ok {
		return ""
	}
	return strings.ToLower(strings.TrimSpace(s))
}

// dirtyNow is whether the tree is dirty now: git status --porcelain printed anything (untracked files that are not
// ignored included); a status that cannot be read counts as dirty.
func dirtyNow(dir string) bool {
	s, ok := git(dir, "status", "--porcelain")
	return !ok || strings.TrimSpace(s) != ""
}

// mergeBase is git merge-base of base and HEAD in dir, "" when git finds none.
func mergeBase(dir, base string) string {
	s, ok := git(dir, "merge-base", "refs/heads/"+base, "HEAD")
	if !ok {
		return ""
	}
	return strings.ToLower(strings.TrimSpace(s))
}
