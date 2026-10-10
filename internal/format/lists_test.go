package format

import (
	"strings"
	"testing"
)

// The ask schema's description of type names the known types; AskTypes, their one home in code, holds the same words
// in the same order.
func TestAskTypesAreTheSchemas(t *testing.T) {
	desc := sub(props(MustLookup("ask").Schema()), "type").String("description")
	var words []string
	for _, w := range AskTypes {
		words = append(words, w.Word)
		if !strings.Contains(desc, w.Word) {
			t.Errorf("ask type %s is in format.AskTypes but not in the ask schema's description of type", w.Word)
		}
		if w.Means == "" {
			t.Errorf("ask type %s has no meaning", w.Word)
		}
	}
	// The schema names them in this order: Answer, Decide, Look or Play from agents, Bless from the ladder.
	at := -1
	for _, w := range words {
		i := strings.Index(desc, w)
		if i <= at {
			t.Errorf("ask type %s is out of the schema's order", w)
		}
		at = i
	}
}

// ExitCodes: codes 0 to 5 for every word but hook, in order, then hook's two; each with its words.
func TestExitCodesTable(t *testing.T) {
	for i, e := range ExitCodes {
		if e.Short == "" || e.Means == "" {
			t.Errorf("exit code row %d has no words", i)
		}
		switch {
		case i < 6 && (e.Code != i || e.Applies != ExitEveryWord):
			t.Errorf("row %d: code %d for %q, want %d for every word but hook", i, e.Code, e.Applies, i)
		case i >= 6 && e.Applies != ExitHook:
			t.Errorf("row %d: %q, want hook's", i, e.Applies)
		}
	}
	if len(ExitCodes) != 8 {
		t.Errorf("%d exit codes, want 6 and hook's 2", len(ExitCodes))
	}
}
