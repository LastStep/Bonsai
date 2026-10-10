package redact

import (
	"strings"
	"testing"
	"time"
)

// No input makes the redactor slow: every rule reads a run once. The limits are the plan's (no string over 100 ms,
// and past 100 KB no more than 1 ms a KB), far above what these take (a few milliseconds for 26 KB, under 300 ms
// for 1 MB) and far below what a rule that rescans a run would take (the studio once took 97 s on the first input).

// gluedWords is size bytes of secret words joined by sep.
func gluedWords(sep string, size int) string {
	words := []string{"password", "token", "apikey", "secret", "credentials", "passwd"}
	var b strings.Builder
	for i := 0; b.Len() < size; i++ {
		b.WriteString(words[i%len(words)])
		b.WriteString(sep)
	}
	return b.String()[:size]
}

func repeatTo(unit string, size int) string {
	return strings.Repeat(unit, size/len(unit)+1)[:size]
}

// mixedText is size bytes of the kind of text the redactor meets: prose, YAML and JSON with secrets in them, a URL
// with credentials, headers, flags, a table, base64 and long runs of glued words.
func mixedText(size int) string {
	m := newMaker(21)
	block := strings.Join([]string{
		"# Notes for the next session", "",
		"Two drafts, or three? I have 2 ready. Merged at 3f1c2a9e8b7d6c5f4e3d2c1b0a9f8e7d6c5b4a3f.",
		"vault:", "  password: hunter2", "  api_key: " + m.secret(16), "roles:", "  deploy-token: read", "token:", "  hunter2",
		`{"apiKey": "` + m.secret(20) + `", "user": "pat", "url": "https://builder:` + m.secret(10) + `@example.com/r.git"}`,
		`curl -H "Authorization: Bearer ` + m.secret(24) + `" https://api.example.com/v2?api_key=` + m.secret(12) + "&x=1",
		"| id | what | where | who |", "|---|---|---|---|", "| 7 | a table row | docs/notes.md | builder |",
		`tool --password "` + m.secret(8) + ` -u root; git -c http.extraHeader="AUTHORIZATION: basic aHVudGVyMg==" push`,
		"blob " + strings.Repeat("QmFzZTY0IGJsb2Igb2YgdGV4dA", 30),
		"run " + gluedWords("_", 3000), "run " + gluedWords("-", 3000), "run " + gluedWords(".", 3000) + " " + strings.Repeat("a+", 1500) + "://",
		"",
	}, "\n")
	return repeatTo(block, size)
}

func TestLinearTime(t *testing.T) {
	const kb = 1024
	cases := []struct {
		name string
		text string
	}{
		{"26 KB of secret words glued by _", gluedWords("_", 26*kb)},
		{"26 KB of secret words glued by -", gluedWords("-", 26*kb)},
		{"26 KB of secret words glued by .", gluedWords(".", 26*kb)},
		{"26 KB of secret words glued", gluedWords("", 26*kb)},
		{"1 MB of mixed text", mixedText(1024 * kb)},
		{"1 MB of names after names", repeatTo(`a bearer password: "x" then vault: token: y, --token api_key: z, Authorization: Bearer token: w, extraheader= bearer abcdefgh123`+"\n", 1024*kb)},
		{"1 MB of token=password= glued", repeatTo("token=password=", 1024*kb)},
		{"1 MB of quotes left open", repeatTo(`token: "x `, 1024*kb)},
		{"1 MB of Authorization: Bearer", repeatTo("Authorization: Bearer ", 1024*kb)},
		{"1 MB of authorization=", repeatTo("authorization=", 1024*kb)},
		{"1 MB of a bearer Token=x", repeatTo("a bearer Token=x ", 1024*kb)},
		{"1 MB of names at line ends", repeatTo("password:\n", 1024*kb)},
		{"1 MB of quotes after a separator", repeatTo(`x' token='`, 1024*kb)},
		{"1 MB of dashes", strings.Repeat("-", 1024*kb)},
		{"1 MB of whitespace then a quote", strings.Repeat(" ", 1024*kb) + `"`},
		{"1 MB of scheme characters", repeatTo("a+", 1024*kb) + "://"},
	}
	Text("warm: password=hunter2")
	for _, c := range cases {
		best := time.Duration(1 << 62)
		for try := 0; try < 3; try++ {
			start := time.Now()
			Text(c.text)
			if d := time.Since(start); d < best {
				best = d
			}
		}
		limit := 100 * time.Millisecond
		if len(c.text) > 100*kb {
			limit = time.Duration(len(c.text)/kb) * time.Millisecond
		}
		t.Logf("%s: %v (limit %v)", c.name, best, limit)
		if best > limit {
			t.Errorf("%s took %v, over its limit %v", c.name, best, limit)
		}
	}
}
