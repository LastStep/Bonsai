package guard

// No user name in a record's text (contract §2.6; the record's own rule for bonsai_path): a refusal names a path
// relative to the session's project or from ~/ (shown), a shell call's refusal holds no part of its command line, and
// the record writer writes the user's home folder, and any other user folder's name a reason quotes, as ~ (userless).
// stderr may keep more.

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/workspace"
)

func TestUserless(t *testing.T) {
	saved := userHome
	defer func() { userHome = saved }()
	userHome = func() (string, error) { return filepath.FromSlash("/home/alice"), nil }
	for in, want := range map[string]string{
		`Write of "/home/alice/bin/bonsai" refused`:           `Write of "~/bin/bonsai" refused`,
		`open /home/bob/x/settings.json: permission denied`:   `open /home/~/x/settings.json: permission denied`,
		`Write of "C:\\Users\\Carol\\AppData\\x" refused`:     `Write of "C:\\Users\\~\\AppData\\x" refused`,
		`at C:/Users/carol/proj and /mnt/c/Users/dan/.bonsai`: `at C:/Users/~/proj and /mnt/c/Users/~/.bonsai`,
		`/Users/erin/Library`:                                 `/Users/~/Library`,
		`no path here`:                                        `no path here`,
		`a ~/x path`:                                          `a ~/x path`,
	} {
		if got := userless(in); got != want {
			t.Errorf("userless(%q) = %q, want %q", in, got, want)
		}
	}
}

// Every reason that quotes a path outside the project names it from ~/ under the user's home folder, in the refusal
// and in its record: Bonsai's home, a stand-in, the dangling or /proc path (bad-path), another project's folder.
func TestRefusalsNameNoUserFolder(t *testing.T) {
	user := t.TempDir()
	saved := userHome
	defer func() { userHome = saved }()
	userHome = func() (string, error) { return user, nil }
	t.Setenv("HOME", user)
	t.Setenv("USERPROFILE", user)
	dir := withTask(t, "running", "**")
	home, _ := workspace.Home()
	for _, p := range []string{filepath.Join(user, "bin", "bonsai"), filepath.Join(user, ".bonsai", "salt"), filepath.Join(home, "settings.json")} {
		in := toolPayload("r1", dir, "Write", schema.Object{{Key: "file_path", Value: p}, {Key: "content", Value: "x"}})
		code, stderr := answer(t, dir, in)
		if code != 2 {
			t.Fatalf("%s: exit %d", p, code)
		}
		recs := records(t, dir, "s-r1.ndjson")
		text := recs[len(recs)-1].String("text")
		if strings.Contains(text, filepath.ToSlash(user)) || strings.Contains(stderr, filepath.ToSlash(user)) {
			t.Errorf("%s: the user's home folder in the reason: %q / %q", p, text, stderr)
		}
	}
	if !strings.Contains(records(t, dir, "s-r1.ndjson")[0].String("text"), `"~/bin/bonsai"`) {
		t.Errorf("the stand-in's reason: %v", records(t, dir, "s-r1.ndjson")[0])
	}
	// A shell call's refusal names what it found, never the line; a bonsai call's, its word.
	wantRule(t, dir, bashIn(dir, "Bash", "unset CLAUDE_CODE_CHILD_SESSION; cp "+filepath.ToSlash(user)+"/secret.txt x"), RulePersonCommand)
	recs := records(t, dir, "s-r1.ndjson")
	if text := recs[len(recs)-1].String("text"); strings.Contains(text, "secret") || strings.Contains(text, "cp ") {
		t.Errorf("the line in the record: %q", text)
	}
	d := judgeIn(dir).ownCommand(dir, []string{"bonsai", "labels", "attach", filepath.Join(user, "acme.yaml")})
	if d.Rule != RuleHomeCommand || strings.Contains(d.Why, "acme") || !strings.HasPrefix(d.Why, `"bonsai labels" refused`) &&
		!strings.HasPrefix(d.Why, "bonsai labels refused") {
		t.Errorf("a bonsai call's reason: %+v", d)
	}
}
