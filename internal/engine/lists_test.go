package engine

import (
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"
)

func words(ws []string) string { sort.Strings(ws); return strings.Join(ws, ",") }

// Every result the plugin step's code gives is in PluginResults, and every word there is given somewhere.
func TestPluginResultsAreTheCode(t *testing.T) {
	src, err := os.ReadFile("plugins.go")
	if err != nil {
		t.Fatal(err)
	}
	re := regexp.MustCompile(`\bResult\b[^=\n]*= "([a-z-]+)"`)
	seen := map[string]bool{}
	for _, m := range re.FindAllStringSubmatch(string(src), -1) {
		seen[m[1]] = true
	}
	var got, want []string
	for w := range seen {
		got = append(got, w)
	}
	for _, w := range PluginResults {
		want = append(want, w.Word)
		if w.Means == "" {
			t.Errorf("plugin result %s has no meaning", w.Word)
		}
	}
	if words(got) != words(want) {
		t.Errorf("plugins.go gives the results %s, engine.PluginResults holds %s", words(got), words(want))
	}
}

// Every state JudgeClaude gives is in ClaudeStates, and the other way round.
func TestClaudeStatesAreTheCode(t *testing.T) {
	src, err := os.ReadFile("claude.go")
	if err != nil {
		t.Fatal(err)
	}
	re := regexp.MustCompile(`\bState\b[^=\n]*= "([a-z]+)"`)
	seen := map[string]bool{}
	for _, m := range re.FindAllStringSubmatch(string(src), -1) {
		seen[m[1]] = true
	}
	var got, want []string
	for w := range seen {
		got = append(got, w)
	}
	for _, w := range ClaudeStates {
		want = append(want, w.Word)
	}
	if words(got) != words(want) {
		t.Errorf("claude.go gives the states %s, engine.ClaudeStates holds %s", words(got), words(want))
	}
}

// The orders are copies: changing one does not change the code's.
func TestClaudeOrdersAreCopies(t *testing.T) {
	o := ClaudeOrders()
	if len(o) != 3 || len(o[0].Keys) != len(claudeKeyOrder) {
		t.Fatalf("ClaudeOrders: %v", o)
	}
	o[0].Keys[0] = "changed"
	if claudeKeyOrder[0] == "changed" {
		t.Errorf("ClaudeOrders hands out the code's own slice")
	}
	if len(CheckLaterWords()) != len(checkLater) {
		t.Errorf("CheckLaterWords differs from checkLater")
	}
}
