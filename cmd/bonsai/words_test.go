package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// T-5729: update leads with what it recorded when it recorded settings for the checks looser than the ones this
// machine accepted, not with "nothing to change"; an update with nothing at all to do keeps its old lead.
func TestUpdateLeadsWithWhatItRecorded(t *testing.T) {
	c := newCLI(t)
	c.link()
	c.commit("link")
	linked := c.read("bonsai.yaml")
	lists := `".claude/**", "bonsai.yaml", ".bonsai/lock.json"`
	if code, out, errOut := c.run("", "update", "--yes"); code != 0 || !strings.Contains(flat(out), "nothing to change") {
		t.Fatalf("an update with nothing to do: exit %d\n%s%s", code, out, errOut)
	}
	c.write("bonsai.yaml", strings.Replace(linked, "protected: ["+lists+"]", `protected: ["bonsai.yaml", ".bonsai/lock.json"]`, 1))
	code, out, errOut := c.run("", "update", "--yes")
	first := strings.Join(strings.SplitN(out, "\n", 3)[:2], " ")
	if code != 0 || strings.Contains(first, "nothing to change") || !strings.Contains(first, "recorded") {
		t.Errorf("update that recorded looser settings, lead line %q (exit %d)\n%s%s", first, code, out, errOut)
	}
}

// T-5729: no word Bonsai prints names a step of its own plan ("step 5.6", "set 8", "plan-5 5.4.5"): a person who
// adopts Bonsai cannot know it. Every string in the program's source (comments are not read) is held to it.
func TestNoPlanStepNamesInPrintedWords(t *testing.T) {
	bad := regexp.MustCompile(`[Ss]tep [0-9]|plan-5|plan part|\b[Ss]et [0-9]|formats set|(^|[^0-9.])5\.[0-7](\.[0-9]+)?([^0-9.]|$)`)
	// A citation of the contract or the spec ("contract section 5.1") is another matter, left for a later fix.
	cited := regexp.MustCompile(`section [0-9.]+`)
	fset := token.NewFileSet()
	for _, top := range []string{"../../cmd", "../../internal"} {
		err := filepath.WalkDir(top, func(path string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return err
			}
			f, err := parser.ParseFile(fset, path, nil, 0)
			if err != nil {
				return err
			}
			ast.Inspect(f, func(n ast.Node) bool {
				lit, ok := n.(*ast.BasicLit)
				if !ok || lit.Kind != token.STRING {
					return true
				}
				if s, err := strconv.Unquote(lit.Value); err == nil && bad.MatchString(cited.ReplaceAllString(s, "")) {
					t.Errorf("%s: %q", fset.Position(lit.Pos()), s)
				}
				return true
			})
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
}
