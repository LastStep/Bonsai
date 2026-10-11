package guard

// A checkout of another project inside one (V1's finding on 5.3.1, adopted for 5.3.2): the nearest verified checkout
// sets aside only the enclosing projects of its own workspace id (Claude Code's worktrees inside their main checkout,
// TestAWorktreeInsideTheMainCheckout); a project of another id linked in a folder of this one, or another project's
// worktree placed there, leaves this one's protection standing. Grants still come only from the nearest verified one.

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/LastStep/Bonsai/internal/workspace"
)

func TestANestedProjectOfAnotherIdKeepsTheEnclosingLists(t *testing.T) {
	outerYAML := strings.Replace(grantYAML, `"docs/plan.md"]`, `"docs/plan.md", "inner/secret.txt", "inner/assets-of-outer/**"]`, 1)
	outer := grantProject(t, outerYAML, map[string]string{"inner/secret.txt": "s\n"})
	home, _ := workspace.Home()
	inner := filepath.Join(outer, "inner")
	const innerID = "ws-iiiiiiiiiiiiiiiiiiiiiiiiii"
	writeFile(t, filepath.Join(inner, "bonsai.yaml"), strings.Replace(grantYAML, testID, innerID, 1))
	writeFile(t, filepath.Join(inner, ".git", "HEAD"), "ref: refs/heads/main\n")
	if _, err := workspace.RecordCheckout(home, inner, innerID, time.Now()); err != nil {
		t.Fatal(err)
	}
	j := judgeIn(inner)
	if !j.verified(j.project(inner)).Verified {
		t.Fatal("the inner checkout is not verified")
	}
	// The enclosing project's own entries still guard inside the inner checkout; its free paths stay free.
	if s := wantRule(t, inner, editIn(inner, "secret.txt", "s", "t"), RuleProtected); !strings.Contains(s, `"inner/secret.txt"`) ||
		!strings.Contains(s, `(in the project "`+filepath.Base(outer)+`")`) {
		t.Errorf("the enclosing project's list: %s", s)
	}
	wantRule(t, inner, writeOf(inner, "free.txt", "x\n"), RuleNotProtected)
	// Grants come from the nearest verified checkout, the inner one.
	writeFile(t, filepath.Join(inner, "work", "tasks", "T-0001-the-task.md"), taskText("T-0001", "running", "light", "secret.txt"))
	wantRule(t, inner, editIn(inner, "secret.txt", "s", "t"), RuleGranted)
}
