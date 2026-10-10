package redact

import (
	"strings"
	"testing"
	"unicode/utf8"
)

// plants are lines that hold a made-up secret no context may keep: each starts its own line, so a name before it
// cannot hide it, and whatever text surrounds it, its secret goes.
var plants = []string{
	"passwd=%s",
	"--secret %s",
	"Authorization: Basic %s",
	"send as bearer %s",
	`"authKey": "%s"`,
	"-c http.extraHeader=%s",
	"ghp_%s",
}

const plantedSecret = "Qx7Rv2Lm9Tz4Wk8Pn3Hc6Jd5"

// FuzzRedact: Text never panics; its output is a fixed point, also cut at any length; Find's spans give its output;
// and a secret planted on a line of its own, in whatever text, never survives.
func FuzzRedact(f *testing.F) {
	for _, s := range []string{
		"", "passwd=Hq7v", "a bearer password: \"hunter2\"", "store:\n  secret: Hq7v\n", "Proxy-Authorization: Token \"Hq 7v\"",
		"git -c http.extraheader= {\"authKey\": \"Hq7v\"}", "--secret \"Hq {\"passwd\": \"7v\"}", "pa\u017f\u017fphrase\u00a0= Hq7v",
		"my-secret-authorization= 'passwd'= Hq7v", "-----BEGIN OPENSSH PRIVATE KEY-----\nb3Bl", "secret: \"Hq passwd=\"7 v\" w",
		"AKIAIOSFODNN7EXAMPLE9QPASSWORD: Zt5", "ftp+ssh://deploy:Zk4q@files.example.org", "vault_token:\n\tQ8m2Lk\n",
		"proxy-authorization= Token := y7Gh",
	} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		out := Text(s)
		if again := Text(out); again != out {
			t.Fatalf("not a fixed point: %q gave %q, then %q", s, out, again)
		}
		step := 1
		if len(out) > 400 {
			step = len(out)/400 + 1
		}
		for k := 1; k < len(out); k += step {
			for k < len(out) && !utf8.RuneStart(out[k]) {
				k++
			}
			if k >= len(out) {
				break
			}
			if c := out[:k]; Text(c) != c {
				t.Fatalf("not a fixed point cut at %d: %q gave %q; the cut %q gave %q", k, s, out, c, Text(c))
			}
		}
		var b strings.Builder
		at := 0
		for _, sp := range Find(s) {
			if sp.Start < at || sp.End <= sp.Start || sp.End > len(s) {
				t.Fatalf("bad spans for %q: %+v", s, Find(s))
			}
			b.WriteString(s[at:sp.Start])
			b.WriteString(Marker)
			at = sp.End
		}
		b.WriteString(s[at:])
		if b.String() != out {
			t.Fatalf("Find's spans give %q, Text gives %q", b.String(), out)
		}
		if survives(s, plantedSecret) != "" {
			return
		}
		k := len(s) / 2
		for k > 0 && !utf8.RuneStart(s[k]) {
			k--
		}
		plant := strings.Replace(plants[len(s)%len(plants)], "%s", plantedSecret, 1)
		if strings.HasPrefix(plant, "ghp_") {
			plant += "abcdefghijkl"
		}
		planted := s[:k] + "\n" + plant + "\n" + s[k:]
		if got := Text(planted); survives(got, plantedSecret) != "" {
			t.Fatalf("the planted secret survived in %q: %q", planted, got)
		}
	})
}
