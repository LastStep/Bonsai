package record

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// Self reads the binary its executable names, path in forward slashes and its SHA-256; what cannot be found is "".
func TestSelf(t *testing.T) {
	dir := t.TempDir()
	bin := filepath.Join(dir, "bin", "bonsai")
	if err := os.MkdirAll(filepath.Dir(bin), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(bin, []byte("a made-up binary"), 0o644); err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256([]byte("a made-up binary"))
	at := func(p string, err error) func() (string, error) { return func() (string, error) { return p, err } }
	if p, s := Self(at(bin, nil), true); p != filepath.ToSlash(bin) || s != hex.EncodeToString(sum[:]) {
		t.Errorf("Self = %q %q", p, s)
	}
	if p, s := Self(at(bin, nil), false); p != filepath.ToSlash(bin) || s != "" {
		t.Errorf("Self without the hash = %q %q", p, s)
	}
	if p, s := Self(at(filepath.Join(dir, "gone"), nil), true); p != filepath.ToSlash(filepath.Join(dir, "gone")) || s != "" {
		t.Errorf("a binary that cannot be read: %q %q", p, s)
	}
	if p, s := Self(at("", errors.New("none")), true); p != "" || s != "" {
		t.Errorf("no binary: %q %q", p, s)
	}
	if p, s := Self(nil, true); p == "" || len(s) != 64 {
		t.Errorf("this test binary: %q %q", p, s)
	}
}

// FromHome writes a path under the home from ~/, case-blind on Windows, and any other as it is.
func TestFromHome(t *testing.T) {
	cases := []struct{ path, home, want string }{
		{"/home/someone/bin/bonsai", "/home/someone", "~/bin/bonsai"},
		{"/home/someone/bin/bonsai", "/home/someone/", "~/bin/bonsai"},
		{"/usr/local/bin/bonsai", "/home/someone", "/usr/local/bin/bonsai"},
		{"/home/someone-else/bonsai", "/home/someone", "/home/someone-else/bonsai"},
		{"/home/someone", "/home/someone", "/home/someone"},
		{"/usr/local/bin/bonsai", "", "/usr/local/bin/bonsai"},
		{"", "/home/someone", ""},
	}
	if runtime.GOOS == "windows" {
		cases = append(cases, struct{ path, home, want string }{`C:\Users\Someone\b\bonsai.exe`, `c:\users\someone`, "~/b/bonsai.exe"})
	}
	for _, c := range cases {
		if got := FromHome(filepath.FromSlash(c.path), filepath.FromSlash(c.home)); got != c.want {
			t.Errorf("FromHome(%q, %q) = %q, want %q", c.path, c.home, got, c.want)
		}
	}
}
