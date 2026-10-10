package reader

import "testing"

// A key line's comment, as the reader skips it: after each kind of value, and never inside quotes or a flow
// sequence's quoted item, nor a # that follows no space.
func TestKeyLineComment(t *testing.T) {
	for text, want := range map[string]bool{
		"id: test-pack                # the pack's id": true,
		"id: test-pack": false,
		"id: test#pack": false,
		"needs:                       # what it needs": true,
		"needs:":                                            false,
		"  - path: a/b.md   # a file":                       true,
		"  - path: a/b.md":                                  false,
		"- - name: x # nested items":                        true,
		`command: "echo # not a comment"`:                   false,
		`command: "echo # not a comment" # but this is`:     true,
		`command: 'it''s # inside'`:                         false,
		`command: 'it''s' # after`:                          true,
		`why: "a \" quote" # after an escape`:               true,
		`runs: ["a #b", c]`:                                 false,
		`runs: ["a #b", c]  # the files`:                    true,
		"runs: [] # none":                                   true,
		"stamp: {} # none":                                  true,
		"description: |  # a block":                         true,
		"description: >-":                                   false,
		"  claude_code: \"2.1.0\"               # oldest\r": true,
		"# a comment line is no key line":                   false,
		"just text":                                         false,
	} {
		if got := KeyLineComment(text); got != want {
			t.Errorf("KeyLineComment(%q) = %v, want %v", text, got, want)
		}
	}
}
