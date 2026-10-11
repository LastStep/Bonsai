package guard

import (
	"strings"
	"testing"
)

// FuzzShellDecide: the delete check never panics and answers the same twice; a refusal says why and what next on
// one ASCII line each, naming its rule; a line it cannot split that holds a delete word is never allowed; and a
// delete that names nothing, put before any text, is refused whatever the text holds.
func FuzzShellDecide(f *testing.F) {
	for _, row := range delTable {
		f.Add(row.line, row.tool == tP)
	}
	for _, s := range []string{
		"", " ", ";", "&&", "|", "\n\n", "'", `"`, "`", "$(", "${", "$((", "((", "{", "}", ")", "(", "<<", "<<<", "@'",
		"@\"\n", "<#", "#>", "`\n", "\\\n", "$'\\x", "{a,b", "{,}", "a{b,c}d{e,f}", "!(x", "%", "%%", "^", "--%",
		"cat <<EOF\nrm -rf junk/*", "bash <<'X'\nrm -rf junk/*\nX\nrm -rf build", "cmd /c \"", "pwsh -enc !!!",
		"rm -rf –r x", "Remove-Item “junk/*” -r", "x=(a b) rm -rf junk/*", "case x in a) rm -rf junk/*;; esac",
		"$((1<<2)) rm -rf junk/*", "echo ${x//'/'/y}; rm -rf .", "f() ( rm -rf junk/* )", "rm -rf \x00junk/*",
		"rm -rf \xff\xfe", strings.Repeat("$(", 8) + "rm" + strings.Repeat(")", 8), strings.Repeat("{a,b}", 12),
	} {
		f.Add(s, false)
		f.Add(s, true)
	}
	s := delFixture(f)
	f.Fuzz(func(t *testing.T, line string, ps bool) {
		tool, d, prefix := tB, shBash, "rm -rf junk/*; "
		if ps {
			tool, d, prefix = tP, shPower, "Remove-Item -Recurse junk/*; "
		}
		a := shellDecide(tool, line, s.own, s)
		if b := shellDecide(tool, line, s.own, s); a != b {
			t.Fatalf("%q: two answers: %+v and %+v", line, a, b)
		}
		if !a.Allow {
			for _, text := range []string{a.Why, a.Next} {
				if text == "" || oneLine(text) != text {
					t.Fatalf("%q: a refusal's reason and next step are one ASCII line each: %q / %q", line, a.Why, a.Next)
				}
			}
			if !strings.Contains(a.Why, a.Rule) {
				t.Fatalf("%q: the reason does not name the rule %s: %s", line, a.Rule, a.Why)
			}
		}
		if _, why := shSplit(line, d, 0); why != "" && delHoldsWord(line) && a.Allow {
			t.Fatalf("%q cannot be split (%s), holds a delete word, and was allowed", line, why)
		}
		if p := shellDecide(tool, prefix+line, s.own, s); p.Allow {
			t.Fatalf("%q after an unnamed delete was allowed: %+v", line, p)
		}
	})
}
