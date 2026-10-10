package redact

import (
	"fmt"
	"strings"
	"testing"
)

// Every rule that takes a whole value by its shape, each with a made-up secret: none survives, and the output is a
// fixed point at every cut.
func TestShapesAreTakenOut(t *testing.T) {
	m := newMaker(1)
	b64 := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	urlsafe := alnum + "_-"
	pemBody := m.from(b64, 64)
	jwt := "eyJ" + m.from(urlsafe, 20) + "." + m.from(urlsafe, 30) + "." + m.from(urlsafe, 40)
	rows := []struct {
		name, text string
		secrets    []string
	}{
		{"private-key with its END line", "the key follows:\n-----BEGIN RSA PRIVATE KEY-----\n" + pemBody + "\n-----END RSA PRIVATE KEY-----\nthat was all", []string{pemBody}},
		{"private-key with no END line", "-----BEGIN OPENSSH PRIVATE KEY-----\n" + pemBody, []string{pemBody}},
		{"webhook-url on discord.com", "notify via https://discord.com/api/webhooks/42/" + m.secret(30) + " tonight", nil},
		{"webhook-url on a discordapp.com subdomain", "https://media.discordapp.com/api/webhooks/7/" + m.secret(20), nil},
		{"webhook-url on Slack", "alerts go to https://hooks.slack.com/services/" + m.secret(9) + "/" + m.secret(24), nil},
		{"url-credentials", "pip install --index-url https://ci:" + m.secret(12) + "@pypi.example.com/simple pkg", nil},
		{"anthropic-key", "the console showed sk-ant-api03-" + m.secret(40) + " once", nil},
		{"openai-key", "client(sk-proj-" + m.secret(30) + ")", nil},
		{"stripe-key, live", "charge with sk_live_" + m.secret(24), nil},
		{"stripe-key, restricted test", "rk_test_" + m.secret(24) + " in staging", nil},
		{"github-token, fine-grained", "github_pat_" + m.secret(40), nil},
		{"github-token, classic", "remote set-url with ghp_" + m.secret(36), nil},
		{"github-token, from an app", "ghs_" + m.secret(36) + " lasts an hour", nil},
		{"slack-token", "bot xoxb-2718281828-" + m.secret(24), nil},
		{"npm-token", "npm_" + m.secret(36) + " was in the CI log", nil},
		{"aws-key-id", "iam user " + awsKeyID + ", rotated", []string{awsKeyID}},
		{"google-api-key", "map tiles use " + "AIza" + m.from(urlsafe, 35), nil},
		{"google-oauth-token", "refresh gave ya29." + m.secret(40), nil},
		{"jwt", "session cookie " + jwt, []string{jwt}},
		{"random-run", "checksum-like " + m.secret(44) + " follows", nil},
	}
	for _, r := range rows {
		t.Run(r.name, func(t *testing.T) {
			secrets := r.secrets
			if secrets == nil {
				secrets = randomParts(r.text)
			}
			checkGone(t, r.text, secrets...)
		})
	}
}

// randomParts picks the made-up secrets out of a row: its words of 9 or more characters with a digit and a capital.
func randomParts(text string) []string {
	var out []string
	for _, w := range wordsOf(text) {
		if len(w) >= 9 && strings.ContainsAny(w, "0123456789") && strings.ContainsAny(w, "ABCDEFGHIJKLMNOPQRSTUVWXYZ") {
			out = append(out, w)
		}
	}
	return out
}

// Every kind of name, and the value after it: the value goes, the name and the words around it stay.
func TestNamesAndTheirValues(t *testing.T) {
	rows := []struct{ in, want string }{
		// keys
		{"set passwd=Tq83vz before the run", "set passwd=[redacted] before the run"},
		{"smtp.password: Tq83vz (rotated)", "smtp.password: [redacted] (rotated)"},
		{`{"clientSecret": "Tq83vz", "region": "eu"}`, `{"clientSecret": [redacted], "region": "eu"}`},
		{"x-access-key: Tq83vz", "x-access-key: [redacted]"},
		{"q=1&pwd=Tq83vz&page=2", "q=1&pwd=[redacted]&page=2"},
		{"passphrase = 'three short words' kept", "passphrase = [redacted] kept"},
		{"mail.credentials=Tq83vz;port=25", "mail.credentials=[redacted];port=25"},
		{"(private-key: Tq83vz)", "(private-key: [redacted])"},
		{"AWS_SECRET_ACCESS_KEY=" + awsSecretKey + " aws s3 ls", "AWS_SECRET_ACCESS_KEY=[redacted] aws s3 ls"},
		{"RELEASE_TOKEN=Tq83vz ./ship.sh", "RELEASE_TOKEN=[redacted] ./ship.sh"},
		{"--admin-password=Tq83vz", "--admin-password=[redacted]"},
		// flags
		{"ctl --password Tq83vz --quiet", "ctl --password [redacted] --quiet"},
		{"ctl -secret Tq83vz", "ctl -secret [redacted]"},
		{"auth --client-secret Tq83vz --scope read", "auth --client-secret [redacted] --scope read"},
		{"y --apikey 'with a space' z", "y --apikey [redacted] z"},
		{"y --auth-token\tTq83vz", "y --auth-token\t[redacted]"},
		{"ctl --passwd --quiet", "ctl --passwd --quiet"},
		// Authorization headers
		{"hdr authorization:bearer Tq83vzWx (expired)", "hdr authorization:bearer [redacted] (expired)"},
		{"X-Forwarded-Authorization:   basic dXNlcjpwYXNz", "X-Forwarded-Authorization:   basic [redacted]"},
		{"Proxy-Authorization: Digest Tq83vz", "Proxy-Authorization: Digest [redacted]"},
		{"PROXY-AUTHORIZATION=Tq83vz x", "PROXY-AUTHORIZATION=[redacted] x"},
		{`http GET "Authorization: token Tq83vz" -v`, `http GET "Authorization: token [redacted]" -v`},
		// Bearer
		{"pass it as a bearer Tq83vzWxYz please", "pass it as a bearer [redacted] please"},
		{"a bearer brief", "a bearer brief"},
		// extraheader
		{`GIT_TRACE=1 git -c http.extraheader="PRIVATE-TOKEN: Tq83vz" clone`, `GIT_TRACE=1 git -c http.extraheader=[redacted] clone`},
		{"run: http.extraHeader=Tq83vz && ls-remote", "run: http.extraHeader=[redacted] && ls-remote"},
	}
	for _, r := range rows {
		got := Text(r.in)
		if got != r.want {
			t.Errorf("Text(%q)\n got %q\nwant %q", r.in, got, r.want)
		}
		fixedAtEveryCut(t, r.in, got)
	}
}

// A quoted value ends at its closing quote on its line and takes what is glued after it; a quote never closed on
// its line runs to the line's end; a quote straight after another key's separator opens that key's value.
func TestQuotedValues(t *testing.T) {
	rows := []struct{ in, want string }{
		{`secret: "two words" then`, `secret: [redacted] then`},
		{`secret: "a b"glued then`, `secret: [redacted] then`},
		{`passwd: 'left open to the end`, `passwd: [redacted]`},
		{"passwd: 'left open\nbelow it", "passwd: [redacted]\nbelow it"},
		{"secret: \"closes below\nhere\" y", "secret: [redacted]\nhere\" y"},
		{`the PASSWD: "u v w, then ACCESS_KEY="r s" done`, `the PASSWD: [redacted]`},
		{`{"pwd": "c; d", "next": 2}`, `{"pwd": [redacted], "next": 2}`},
		{`--secret "q"r s`, `--secret [redacted] s`},
	}
	for _, r := range rows {
		got := Text(r.in)
		if got != r.want {
			t.Errorf("Text(%q)\n got %q\nwant %q", r.in, got, r.want)
		}
		fixedAtEveryCut(t, r.in, got)
	}
}

// A value on a later line than its name is taken, unless that line starts with a name on its own line, or is the
// text's last line and one word.
func TestValuesBelowTheirName(t *testing.T) {
	rows := []struct{ in, want string }{
		{"passwd:\n      Tq83vz\n", "passwd:\n      [redacted]\n"},
		{"secret:\n\n  \"Tq83vz y\"\ndone: 1", "secret:\n\n  [redacted]\ndone: 1"},
		{"reply with BEARER\n  Tq83vzWxYz\n", "reply with BEARER\n  [redacted]\n"},
		{"store:\n  secret: Tq83vz", "store:\n  secret: [redacted]"},
		{"store:\n  'passwd': 'Tq83vz'\n", "store:\n  'passwd': [redacted]\n"},
		{"upstream authorization:\r\n-secret Tq83vz end", "upstream authorization:\r\n-secret [redacted] end"},
		{"deploy_secret:\nauthorization: Token Tq83vz", "deploy_secret:\nauthorization: Token [redacted]"},
		// The last line, one word: it may be a name the recorder's cut left, so it is kept.
		{"store:\n  secr", "store:\n  secr"},
		{"passwd:\n  Tq83vz", "passwd:\n  Tq83vz"},
		// A flag takes nothing from the next line.
		{"run --token\nnotes: kept", "run --token\nnotes: kept"},
		// A word that only ends in a secret word is not a flag.
		{"read the deploy-secret\nnote: text", "read the deploy-secret\nnote: text"},
	}
	for _, r := range rows {
		got := Text(r.in)
		if got != r.want {
			t.Errorf("Text(%q)\n got %q\nwant %q", r.in, got, r.want)
		}
		fixedAtEveryCut(t, r.in, got)
	}
}

// What the redactor keeps unchanged.
func TestKeeps(t *testing.T) {
	m := newMaker(2)
	rows := []string{
		"rebased onto 0d9e8f7a6b5c4d3e2f1a0b9c8d7e6f5a4b3c2d1e and e5f6a7b",
		"trace 4b7e2a91-0c3d-4f5e-8a6b-1c2d3e4f5a6b",
		"TRACE 4B7E2A91-0C3D-4F5E-8A6B-1C2D3E4F5A6B",
		"uuid: 9a8b7c6d5e4f30211203f4e5d6c7b8a9",
		"digest " + "BADC0FFEE0DDF00D0123456789abcdefABCDEF99",
		"toolu_01" + m.secret(24),
		"srvtoolu_01" + m.secret(30),
		"spec/Billing.Specs/InvoiceTotalsRoundHalfEvenWhenCurrencyHasNoCents2Spec.cs",
		"notes/2026-10-11-what-the-cleaner-keeps.md",
		`D:/work/shop-front/cmd/server/main.go`,
		"Which of the four plans ships first? Two are ready.",
		"we rotate each secret monthly; tokens expire after an hour",
		"the nightly-token job, an ops-secret folder, our admin-password rule",
		"build --secret\nregion: eu-west",
		"db_passwd=[redacted] and x-api-key: [redacted]",
		"Proxy-Authorization: Digest [redacted]",
		"",
	}
	for _, r := range rows {
		if got := Text(r); got != r {
			t.Errorf("Text(%q) changed it to %q", r, got)
		}
	}
}

// A cut tail of the marker, or after an Authorization header of a scheme word, with only whitespace after it, is
// kept; anywhere else it is a value.
func TestCutTails(t *testing.T) {
	for _, kept := range []string{"passwd=[r", "a secret: [redact", "--apikey [re", "remote.extraheader = [re",
		"proxy-authorization: Dige", "AUTHORIZATION:\tBasic ", "authorization= 'Bea", "x-authorization=\tToken [r", "y= token= [  "} {
		if got := Text(kept); got != kept {
			t.Errorf("Text(%q) = %q, want it kept", kept, got)
		}
	}
	for in, want := range map[string]string{
		"passwd=[reda next":             "passwd=[redacted] next",
		"--apikey [re y":                "--apikey [redacted] y",
		"x-proxy-authorization= Basics": "x-proxy-authorization= [redacted]",
		"secret=basic":                  "secret=[redacted]",
	} {
		if got := Text(in); got != want {
			t.Errorf("Text(%q) = %q, want %q", in, got, want)
		}
	}
}

// A keyword matches in any case under simple case folding: the long s counts as s, the Kelvin sign as k.
func TestCaseFolding(t *testing.T) {
	rows := []struct{ in, want string }{
		{"DB_PASSWD=Tq83vz", "DB_PASSWD=[redacted]"},
		{"SeCrEt: Tq83vz", "SeCrEt: [redacted]"},
		{"pa\u017f\u017fphrase=Tq83vz", "pa\u017f\u017fphrase=[redacted]"},
		{"access_\u212aey: Tq83vz", "access_\u212aey: [redacted]"},
		{"user_to\u212aen=Tq83vz", "user_to\u212aen=[redacted]"},
		{"--\u017fecret Tq83vz", "--\u017fecret [redacted]"},
		{"BeArEr Tq83vzWxYz", "BeArEr [redacted]"},
		{"PROXY-AUTHORIZATION: DIGEST Tq83vz", "PROXY-AUTHORIZATION: DIGEST [redacted]"},
	}
	for _, r := range rows {
		got := Text(r.in)
		if got != r.want {
			t.Errorf("Text(%q)\n got %q\nwant %q", r.in, got, r.want)
		}
		fixedAtEveryCut(t, r.in, got)
	}
}

// Any character JavaScript's \s matches may stand between a name and its separator.
func TestEveryWhitespaceBeforeTheSeparator(t *testing.T) {
	spaces := []rune{'\t', '\v', '\f', ' ', 0xA0, 0x1680, 0x2000, 0x2005, 0x200A, 0x2028, 0x2029, 0x202F, 0x205F, 0x3000, 0xFEFF}
	m := newMaker(3)
	for _, sp := range spaces {
		for _, sep := range []string{":", "="} {
			for _, name := range []string{"passwd", "x-access-key", "\"secret\"", "authorization", "extraheader"} {
				if name == "extraheader" && sep == ":" {
					continue
				}
				secret := m.secret(14)
				in := fmt.Sprintf("with %s%c%s %s today", name, sp, sep, secret)
				checkGone(t, in, secret)
			}
		}
	}
}

// Find gives the spans Text replaces, with their rules' kinds.
func TestFind(t *testing.T) {
	in := "set secret=Tq83vz, then " + awsKeyID + ", then Authorization: Basic Tq83vzWx"
	spans := Find(in)
	var b strings.Builder
	at := 0
	var kinds []string
	for _, sp := range spans {
		if sp.Start < at || sp.End <= sp.Start {
			t.Fatalf("spans out of order or touching: %+v", spans)
		}
		b.WriteString(in[at:sp.Start])
		b.WriteString(Marker)
		at = sp.End
		kinds = append(kinds, string(sp.Kind))
	}
	b.WriteString(in[at:])
	if got := b.String(); got != Text(in) {
		t.Errorf("Find's spans give %q, Text gives %q", got, Text(in))
	}
	if want := "secret-named-key aws-key-id authorization-header"; strings.Join(kinds, " ") != want {
		t.Errorf("kinds %q, want %q", strings.Join(kinds, " "), want)
	}
	if Find("nothing to see") != nil || Find("") != nil {
		t.Error("Find found something in clean text")
	}
	known := map[Kind]bool{}
	for _, k := range Kinds {
		known[k.Kind] = true
	}
	for _, r := range shapeRules {
		if !known[r.kind] {
			t.Errorf("kind %q is not in Kinds", r.kind)
		}
	}
	for _, k := range nameKinds {
		if !known[k] {
			t.Errorf("kind %q is not in Kinds", k)
		}
	}
}

// Text of a cut keeps a cut tail; Capped cuts after redacting, by characters.
func TestCapped(t *testing.T) {
	in := "y passwd=Tq83vz " + strings.Repeat("é", 10)
	got := Capped(in, 13)
	if got != "y passwd=[red" {
		t.Errorf("Capped = %q", got)
	}
	if Text(got) != got {
		t.Errorf("a capped field changed on a second pass: %q", Text(got))
	}
	if c := Cut("ééé", 2); c != "éé" {
		t.Errorf("Cut by characters gave %q", c)
	}
	if c := Cut("abc", 0); c != "" {
		t.Errorf("Cut to 0 gave %q", c)
	}
}

// Text that is not valid UTF-8 keeps its bytes, and its secrets still go.
func TestInvalidUTF8(t *testing.T) {
	in := "b\xfe\xff secret=Tq83vz \xe2"
	got := Text(in)
	if got != "b\xfe\xff secret=[redacted] \xe2" {
		t.Errorf("Text(%q) = %q", in, got)
	}
}
