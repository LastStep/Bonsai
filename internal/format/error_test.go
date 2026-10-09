package format

import (
	"regexp"
	"testing"

	"github.com/LastStep/Bonsai/internal/schema"
)

// The error words' table (plan-5 5.1.4b), the one home of the error object's codes: each word fits the schema's
// pattern and is there once, says what it means in one ASCII line, and names who usually takes the next step, as
// next.who's closed list has it.
func TestErrorWords(t *testing.T) {
	if len(ErrorWords) == 0 {
		t.Fatal("no error words")
	}
	props, _ := MustLookup("error").Schema().Get("properties")
	code, _ := props.(schema.Object).Get("code")
	pat := regexp.MustCompile(code.(schema.Object).String("pattern"))
	seen := map[string]bool{}
	for _, w := range ErrorWords {
		if !pat.MatchString(w.Word) || seen[w.Word] {
			t.Errorf("%q: not the schema's pattern, or there twice", w.Word)
		}
		seen[w.Word] = true
		if w.Who != "agent" && w.Who != "person" {
			t.Errorf("%s: who %q, want agent or person", w.Word, w.Who)
		}
		if w.Means == "" {
			t.Errorf("%s: no meaning", w.Word)
		}
		for i := 0; i < len(w.Means); i++ {
			if w.Means[i] < 0x20 || w.Means[i] > 0x7e {
				t.Errorf("%s: its meaning is not one line of ASCII", w.Word)
				break
			}
		}
		if got, ok := ErrorWord(w.Word); !ok || got != w {
			t.Errorf("ErrorWord(%s) = %+v", w.Word, got)
		}
	}
	if _, ok := ErrorWord("no-such-word"); ok {
		t.Error("ErrorWord finds a word not in the table")
	}
}
