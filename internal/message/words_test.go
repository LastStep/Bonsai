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

// No word waits on a step that has landed (5.5.11's follow-up): the guard's records are read since step 5.5.11, so
// neither this package's code nor format.MessageFlags' text says a flag or the recipient waits "until" it.
func TestNoWordWaitsOnTheGuardsRecords(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".go") || strings.HasSuffix(e.Name(), "_test.go") {
			continue
		}
		raw, err := os.ReadFile(e.Name())
		if err != nil {
			t.Fatal(err)
		}
		for _, stale := range []string{"until step 5.5.11", "from step 5.5.11, never raised", "until 5.5.11"} {
			if strings.Contains(string(raw), stale) {
				t.Errorf("%s still says %q", e.Name(), stale)
			}
		}
	}
	for _, f := range format.MessageFlags {
		if strings.Contains(f.Means, "until") || strings.Contains(f.Means, "5.5.11") {
			t.Errorf("format.MessageFlags' %s: %s", f.Word, f.Means)
		}
	}
}
