package redact

import (
	"fmt"
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

// mixedText is size bytes of the kind of text the redactor meets: notes, YAML and JSON holding secrets, a mirror URL
// with credentials, a header, flags, a table, an attachment's base64 and long runs of glued words.
func mixedText(size int) string {
	m := newMaker(21)
	block := strings.Join([]string{
		"## Release checklist", "",
		"Tag the build after review; commit 9be04c7d21aa3f58e6b1c0d4f7a29e83b5c6d1f0 is the last green one.",
		"rollout:", "  region: eu-west", "  access_key: " + m.secret(18), "  passwd: Zq81", "  readers: [ops, qa]",
		"secret:", "\t" + m.secret(10),
		`[{"name": "cache", "client_secret": "` + m.secret(22) + `"}, {"name": "queue", "port": 5672}]`,
		"mirror https://mirror:" + m.secret(9) + "@pkgs.example.org/simple and retry",
		`wget --header "Proxy-Authorization: Basic ` + m.secret(20) + `" -q https://dl.example.org/x.tgz?access_token=` + m.secret(14),
		"| step | owner | state |", "| -- | -- | -- |", "| sign | release | done |",
		"tool sync --api-key '" + m.secret(12) + "' --dry-run; git -c http.extraheader=\"Bearer " + m.secret(16) + "\" pull",
		"attachment " + strings.Repeat("U29tZSBhdHRhY2hlZCBieXRlcw", 30),
		"glued " + gluedWords("_", 3000), "glued " + gluedWords("-", 3000),
		"glued " + gluedWords(".", 3000) + " " + strings.Repeat("a+", 1500) + "://",
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
		{"1 MB of names after names", repeatTo(`use Bearer secret= 'k1' as before; store: passwd= m2, -password access_key= n3; Proxy-Authorization: Token pwd= p4 extraHeader= BEARER zyxw98765`+"\n", 1024*kb)},
		{"1 MB of token=password= glued", repeatTo("token=password=", 1024*kb)},
		{"1 MB of quotes left open", repeatTo(`token: "x `, 1024*kb)},
		{"1 MB of PROXY-AUTHORIZATION: digest", repeatTo("PROXY-AUTHORIZATION:\tdigest ", 1024*kb)},
		{"1 MB of x-authorization=", repeatTo("x-authorization=", 1024*kb)},
		{"1 MB of a bearer Token=x", repeatTo("a bearer Token=x ", 1024*kb)},
		{"1 MB of names at line ends", repeatTo("password:\n", 1024*kb)},
		{"1 MB of quotes after a separator", repeatTo(`x' token='`, 1024*kb)},
		{"1 MB of dashes", strings.Repeat("-", 1024*kb)},
		{"1 MB of whitespace then a quote", strings.Repeat(" ", 1024*kb) + `"`},
		{"1 MB of scheme characters", repeatTo("a+", 1024*kb) + "://"},
	}
	// Runs that every match used to rescan to their end: a flag's characters after the long s or the Kelvin sign (each
	// `-` read the rest of the run), JSON web token starts that never become one, and the exact-length Google key.
	for _, unit := range []string{"\u017f-", "\u212a-", "eyJ-", "AIza-"} {
		for _, size := range []int{256 * kb, 1024 * kb} {
			cases = append(cases, struct {
				name string
				text string
			}{fmt.Sprintf("%d KB of %q", size/kb, unit), repeatTo(unit, size)})
		}
	}
	// Every flag in one such run ends where the run does, and each used to skip the whitespace after that end again.
	for _, unit := range []string{"\u017f-", "\u212a-"} {
		cases = append(cases, struct {
			name string
			text string
		}{fmt.Sprintf("1 MB of %q flags sharing one end, then whitespace", unit),
			repeatTo(unit, 512*kb) + "password" + strings.Repeat(" \t", 256*kb) + "x"})
	}
	Text("warm: password=hunter2")
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
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
			t.Logf("%v (limit %v)", best, limit)
			if best > limit {
				t.Errorf("took %v, over its limit %v", best, limit)
			}
		})
	}
}
