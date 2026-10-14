package message

// Every refusal's word in this package's code is one of format.ErrorWords, their one home: each Code of an Error
// literal, read from the non-test files, as cmd/bonsai's TestErrorWordsInTheCode reads its own packages'.

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

func TestErrorWordsInTheCode(t *testing.T) {
	pkgs, err := parser.ParseDir(token.NewFileSet(), ".", func(fi os.FileInfo) bool { return !strings.HasSuffix(fi.Name(), "_test.go") }, 0)
	if err != nil {
		t.Fatal(err)
	}
	found := map[string]bool{}
	for _, pkg := range pkgs {
		for name, f := range pkg.Files {
			ast.Inspect(f, func(node ast.Node) bool {
				lit, ok := node.(*ast.CompositeLit)
				if !ok {
					return true
				}
				if id, ok := lit.Type.(*ast.Ident); !ok || id.Name != "Error" {
					return true
				}
				for _, el := range lit.Elts {
					kv, ok := el.(*ast.KeyValueExpr)
					if k, isIdent := kv.Key.(*ast.Ident); !ok || !isIdent || k.Name != "Code" {
						continue
					}
					if bl, ok := kv.Value.(*ast.BasicLit); ok && bl.Kind == token.STRING {
						word, _ := strconv.Unquote(bl.Value)
						found[word] = true
						if _, known := format.ErrorWord(word); !known {
							t.Errorf("%s: the word %q is not in format.ErrorWords", name, word)
						}
					}
				}
				return true
			})
		}
	}
	// badValue, badFlag and unexpected give the other three.
	for _, w := range []string{"missing-value", "report-no-status", "role-not-declared", "budget-over-max", "pointer-missing",
		"message-not-found", "report-not-in-thread", "read-failed", "write-failed", "partly-written"} {
		if !found[w] {
			t.Errorf("the code names no %s", w)
		}
	}
}
