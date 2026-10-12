package engine

// The ladder's check words in the code (format.LadderCheckWords, plan-5 5.4.2 note 11), as TestCheckWordsInTheCode
// holds format.CheckWords: every word of the table is added by the check code, as a literal, and none is in both
// tables. Each word's case, a project where check finds it, is cmd/bonsai's TestLadderCheckTable.

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/format"
)

func TestLadderCheckWordsInTheCode(t *testing.T) {
	pkgs, err := parser.ParseDir(token.NewFileSet(), ".", func(fi os.FileInfo) bool { return !strings.HasSuffix(fi.Name(), "_test.go") }, 0)
	if err != nil {
		t.Fatal(err)
	}
	used := map[string]bool{}
	for _, pkg := range pkgs {
		for _, f := range pkg.Files {
			ast.Inspect(f, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok || len(call.Args) == 0 {
					return true
				}
				if sel, ok := call.Fun.(*ast.SelectorExpr); !ok || sel.Sel.Name != "add" {
					return true
				}
				if lit, ok := call.Args[0].(*ast.BasicLit); ok && lit.Kind == token.STRING {
					word, _ := strconv.Unquote(lit.Value)
					used[word] = true
				}
				return true
			})
		}
	}
	for _, w := range format.LadderCheckWords {
		if !used[w.Word] {
			t.Errorf("format.LadderCheckWords has %s, which no code adds", w.Word)
		}
		if w.Kind != "finding" && w.Kind != "warning" || w.Who != "agent" && w.Who != "person" {
			t.Errorf("%s: kind %q, who %q", w.Word, w.Kind, w.Who)
		}
		for _, c := range format.CheckWords {
			if c.Word == w.Word {
				t.Errorf("%s is in both CheckWords and LadderCheckWords", w.Word)
			}
		}
	}
}
