package format

import (
	"regexp"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/schema"
)

// check's words (step 5.1.6), the one home of a finding's and a warning's code: each fits the schema's pattern and
// is there once, is a finding or a warning, names who usually takes its next step, and says what it means in one
// ASCII line; check --schema bonsai.check prints them, each with its kind.
func TestCheckWords(t *testing.T) {
	props, _ := MustLookup("check").Schema().Get("properties")
	findings, _ := props.(schema.Object).Get("findings")
	items, _ := findings.(schema.Object).Get("items")
	ip, _ := items.(schema.Object).Get("properties")
	code, _ := ip.(schema.Object).Get("code")
	pat := regexp.MustCompile(code.(schema.Object).String("pattern"))
	seen := map[string]bool{}
	kinds := map[string]int{}
	for _, w := range CheckWords {
		if !pat.MatchString(w.Word) || seen[w.Word] {
			t.Errorf("%q: not the schema's pattern, or there twice", w.Word)
		}
		seen[w.Word] = true
		kinds[w.Kind]++
		if w.Who != "agent" && w.Who != "person" {
			t.Errorf("%s: who %q", w.Word, w.Who)
		}
		for i := 0; i < len(w.Means); i++ {
			if w.Means[i] < 0x20 || w.Means[i] > 0x7e {
				t.Errorf("%s: its meaning is not one line of ASCII", w.Word)
				break
			}
		}
		if got, ok := CheckWord(w.Word); !ok || got != w {
			t.Errorf("CheckWord(%s) = %+v", w.Word, got)
		}
	}
	if kinds["finding"] == 0 || kinds["warning"] == 0 || kinds["finding"]+kinds["warning"] != len(CheckWords) {
		t.Errorf("kinds %v", kinds)
	}
	// check --pack's words (step 5.1.9) share the open list: the same rules, every one a finding, none in both tables.
	for _, w := range PackCheckWords {
		if !pat.MatchString(w.Word) || seen[w.Word] {
			t.Errorf("%q: not the schema's pattern, or there twice (in either table)", w.Word)
		}
		seen[w.Word] = true
		if w.Kind != "finding" || w.Who != "agent" {
			t.Errorf("%s: kind %q, who %q: check --pack's words are findings whose next step is an agent's", w.Word, w.Kind, w.Who)
		}
		for i := 0; i < len(w.Means); i++ {
			if w.Means[i] < 0x20 || w.Means[i] > 0x7e {
				t.Errorf("%s: its meaning is not one line of ASCII", w.Word)
				break
			}
		}
		if got, ok := PackCheckWord(w.Word); !ok || got != w {
			t.Errorf("PackCheckWord(%s) = %+v", w.Word, got)
		}
		if _, ok := CheckWord(w.Word); ok {
			t.Errorf("%s is in CheckWords too", w.Word)
		}
	}
	text := flat(MustLookup("check").Describe())
	if !strings.Contains(text, "changed (a finding (exit 1): a file the lock lists was edited") ||
		!strings.Contains(text, "same-id (a warning (never the exit code): another checkout") ||
		!strings.Contains(text, "known words (format.PackCheckWords, their one home") ||
		!strings.Contains(text, "pack-why (a finding (exit 1): a deny rule in bonsai/pack.yaml has no why") {
		t.Errorf("check --schema bonsai.check does not print the words with their kinds:\n%s", text)
	}
}
