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

// gluedWords is size bytes of secret words with sep between them.
func gluedWords(sep string, size int) string {
	words := []string{"passphrase", "api-key", "pwd", "privatekey", "token", "accesskey", "secret"}
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

// mixedText is size bytes of the kind of text the redactor meets: a shell session, an env file, a service's config,
// a stack trace, a request dump, a CSV export, a package lock's hashes, and long runs of secret words.
func mixedText(size int) string {
	m := newMaker(21)
	block := strings.Join([]string{
		"$ make deploy ENV=staging",
		"reading .env.staging",
		"DATABASE_URL=postgres://svc_orders:" + m.secret(11) + "@db-3.internal:5432/orders",
		"SMTP_PASS=" + m.secret(15) + "  # rotated each quarter",
		"services:", "  queue:", "    image: queue:4.1", "    signing_secret:", "      " + m.secret(19),
		"panic: runtime error: index out of range [5] with length 5",
		"goroutine 17 [running]:", "main.(*Ledger).Post(0xc000112000)", "\t/src/ledger/post.go:88 +0x1d4",
		"POST /v2/charges HTTP/1.1", "Host: pay.example.net", "X-Api-Key: " + m.secret(26), "Content-Length: 64",
		`id,owner,"access_key",note`, `17,ops,"` + m.secret(16) + `","moved, then tagged 4f9c2e81a7d03b56"`,
		"integrity sha512-" + strings.Repeat("Zm9yIHRoZSBsb2NrIGZpbGU", 25),
		"ssh-keygen -t ed25519 --passphrase '" + m.secret(13) + "' -f ./deploy_key && echo done",
		"words " + gluedWords("-", 2500), "words " + gluedWords("_", 2500), "words " + gluedWords("", 2500),
		"scheme " + strings.Repeat("v1.", 1200) + "://", "words " + gluedWords(".", 2500),
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
		{"secret words with _ between [26 KB]", gluedWords("_", 26*kb)},
		{"secret words with - between [26 KB]", gluedWords("-", 26*kb)},
		{"secret words with . between [26 KB]", gluedWords(".", 26*kb)},
		{"secret words with nothing between [26 KB]", gluedWords("", 26*kb)},
		{"a shell session, configs and logs [1 MB]", mixedText(1024 * kb)},
		{"one line of names behind names [1 MB]", repeatTo(`use Bearer secret= 'k1' as before; store: passwd= m2, -password access_key= n3; Proxy-Authorization: Token pwd= p4 extraHeader= BEARER zyxw98765`+"\n", 1024*kb)},
		{"apikey: and pwd= with nothing between [1 MB]", repeatTo("apikey:pwd=", 1024*kb)},
		{"an open quote after every name [1 MB]", repeatTo(`passwd: 'q `, 1024*kb)},
		{"a digest header on every line [1 MB]", repeatTo("PROXY-AUTHORIZATION:\tdigest ", 1024*kb)},
		{"header names with nothing between [1 MB]", repeatTo("Upstream-AUTHORIZATION=", 1024*kb)},
		{"a key after every bearer [1 MB]", repeatTo("send bearer apiKey:q ", 1024*kb)},
		{"a name ending every line [1 MB]", repeatTo("password:\n", 1024*kb)},
		{"a quote after every separator [1 MB]", repeatTo(`x' token='`, 1024*kb)},
		{"dashes [1 MB]", strings.Repeat("-", 1024*kb)},
		{"spaces and then a quote [1 MB]", strings.Repeat(" ", 1024*kb) + `"`},
		{"a scheme's characters before :// [1 MB]", repeatTo("v1.", 1024*kb) + "://"},
	}
	// Runs that every match used to rescan to their end: a flag's characters after the long s or the Kelvin sign (each
	// `-` read the rest of the run), JSON web token starts that never become one, and the exact-length Google key.
	for _, unit := range []string{"ſ-", "K-", "eyJ-", "AIza-"} {
		for _, size := range []int{256 * kb, 1024 * kb} {
			cases = append(cases, struct {
				name string
				text string
			}{fmt.Sprintf("%q repeated [%d KB]", unit, size/kb), repeatTo(unit, size)})
		}
	}
	// Every flag in one such run ends where the run does, and each used to skip the whitespace after that end again.
	for _, unit := range []string{"ſ-", "K-"} {
		cases = append(cases, struct {
			name string
			text string
		}{fmt.Sprintf("%q flags sharing one end, then whitespace [1 MB]", unit),
			repeatTo(unit, 512*kb) + "password" + strings.Repeat(" \t", 256*kb) + "x"})
	}
	// Shapes glued into one run, so that every shape leaves a part of the run for the others (the leftovers of each are
	// read once), and key ids each followed by a word character, the run judged whole once.
	for _, unit := range []string{"-ghp_" + strings.Repeat("Qx7", 8), "-AKIA" + strings.Repeat("Q7", 8) + "zz",
		"-AKIA" + strings.Repeat("QZ", 8) + "a", "+AIza" + strings.Repeat("Qx7Wk", 8)} {
		cases = append(cases, struct {
			name string
			text string
		}{fmt.Sprintf("%q repeated [1 MB]", unit), repeatTo(unit, 1024*kb)})
	}
	Text("first call: passwd=Tq83vz")
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
