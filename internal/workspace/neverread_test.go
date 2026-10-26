package workspace

import (
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/schema"
)

// never_read (BUG-121): a path taken out of it is looser, one added stricter; a bonsai.yaml written before the field
// (none of it) still reads, as an empty list; a record of the checks written before it reads too, as an empty list,
// and equals the same settings read now; a path that cannot be a Read(/<path>) rule is refused by
// the full read (the guard's lean read leaves never_read out).
func TestNeverRead(t *testing.T) {
	with := checksYAML + `never_read: [".env", "secrets/**"]` + "\n"
	was := checksOf(t, with)
	if got := strings.Join(was.NeverRead, " "); got != ".env secrets/**" {
		t.Fatalf("never_read read as %q", got)
	}
	taken := strings.Replace(with, `".env", `, "", 1)
	if got := Loosenings(was, checksOf(t, taken)); len(got) != 1 || got[0] != `never_read no longer holds ".env"` {
		t.Errorf("a never-read path taken out: %q", got)
	}
	if got := Loosenings(checksOf(t, taken), was); len(got) != 0 {
		t.Errorf("a never-read path added reads as looser: %q", got)
	}
	// An old bonsai.yaml, with no never_read: it reads, as none.
	old := checksOf(t, checksYAML)
	if old.NeverRead == nil || len(old.NeverRead) != 0 {
		t.Errorf("an old bonsai.yaml's never_read: %#v", old.NeverRead)
	}
	if got := Loosenings(was, old); len(got) != 2 {
		t.Errorf("never_read taken out whole: %q", got)
	}
	// A record written before never_read: read back, it equals the same settings now.
	raw, err := schema.Encode(old.JSON())
	if err != nil {
		t.Fatal(err)
	}
	rec, err := schema.Decode(raw)
	if err != nil {
		t.Fatal(err)
	}
	var trimmed schema.Object
	for _, f := range rec.(schema.Object) {
		if f.Key != "never_read" {
			trimmed = append(trimmed, f)
		}
	}
	back, why := readChecks(trimmed)
	if why != "" || !sameChecks(back, old) || len(Loosenings(back, old)) != 0 {
		t.Errorf("a record with no never_read: %q, same %v", why, why == "" && sameChecks(back, old))
	}
	for _, bad := range []string{`"a(b)"`, `"~/.env"`, `"/etc/passwd"`} {
		if _, err := ReadConfigFull([]byte(checksYAML + "never_read: [" + bad + "]\n")); err == nil {
			t.Errorf("never_read %s was read", bad)
		}
	}
}
