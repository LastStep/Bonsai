package workspace

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Every key's values: each good one kept as the file stores it, each bad one refused naming the values the key takes.
func TestCheckSetting(t *testing.T) {
	home := filepath.Join(t.TempDir(), "home")
	for _, c := range []struct {
		key, value, goos string
		want             any
		bad              string
	}{
		{"status_writes", "agents", "linux", "agents", ""},
		{"status_writes", "command", "linux", "command", ""},
		{"status_writes", "Command", "linux", nil, "agents or command"},
		{"status_writes", "", "linux", nil, "agents or command"},
		{"status_command", "studio task move", "linux", "studio task move", ""},
		{"status_command", "a-b_c.d/e:f 9", "linux", "a-b_c.d/e:f 9", ""},
		{"status_command", strings.Repeat("a", 100), "linux", strings.Repeat("a", 100), ""},
		{"status_command", strings.Repeat("a", 101), "linux", nil, "1 to 100 characters"},
		{"status_command", "move; rm -rf x", "linux", nil, "outside its rule"},
		{"status_command", "move $(id)", "linux", nil, "outside its rule"},
		{"status_command", "move `id`", "linux", nil, "outside its rule"},
		{"status_command", "move\nx", "linux", nil, "outside its rule"},
		{"status_command", "move 'x'", "linux", nil, "outside its rule"},
		{"status_command", "   ", "linux", nil, "outside its rule"},
		{"status_command", "", "linux", nil, "outside its rule"},
		{"cache_keep_days", "1", "linux", int64(1), ""},
		{"cache_keep_days", "3650", "linux", int64(3650), ""},
		{"cache_keep_days", "none", "linux", nil, ""},
		{"cache_keep_days", "0", "linux", nil, "from 1 to 3650"},
		{"cache_keep_days", "3651", "linux", nil, "from 1 to 3650"},
		{"cache_keep_days", "-5", "linux", nil, "from 1 to 3650"},
		{"cache_keep_days", "+5", "linux", nil, "from 1 to 3650"},
		{"cache_keep_days", "1.5", "linux", nil, "from 1 to 3650"},
		{"cache_keep_days", "30d", "linux", nil, "from 1 to 3650"},
		{"personal_from", "//wsl.localhost/Ubuntu/home/u/.bonsai/personal", "windows", "//wsl.localhost/Ubuntu/home/u/.bonsai/personal", ""},
		{"personal_from", `C:\data\personal`, "windows", "C:/data/personal", ""},
		{"personal_from", "none", "windows", nil, ""},
		{"personal_from", "//wsl.localhost/Ubuntu/home/u/.bonsai/personal", "linux", nil, "Windows only"},
		{"personal_from", "relative/personal", "windows", nil, "not an absolute Windows path"},
		{"personal_from", "C:/x\ty", "windows", nil, "control character"},
		{"nothing", "x", "linux", nil, "not a key"},
	} {
		got, why := CheckSetting(c.key, c.value, home, c.goos)
		if c.bad != "" {
			if why == "" || !strings.Contains(why, c.bad) {
				t.Errorf("%s=%q on %s: want refused (%s), got %v %q", c.key, c.value, c.goos, c.bad, got, why)
			}
			continue
		}
		if why != "" || got != c.want {
			t.Errorf("%s=%q on %s: got %#v %q, want %#v", c.key, c.value, c.goos, got, why, c.want)
		}
	}
	// personal_from never names the home's own personal/ or a folder in it, whatever its case or slashes.
	winHome := `C:\Users\someone\.bonsai`
	for _, p := range []string{`C:\Users\someone\.bonsai\personal`, "c:/users/someone/.bonsai/Personal/", `C:\Users\someone\.bonsai\personal\notes`} {
		if _, why := personalFromCheck(p, winHome); !strings.Contains(why, "its target") {
			t.Errorf("personal_from %s with the home %s: %q", p, winHome, why)
		}
	}
	if _, why := personalFromCheck(`C:\Users\someone\.bonsai\personal-old`, winHome); why != "" {
		t.Errorf("a sibling of personal/: %q", why)
	}
}

// personalFromCheck is personalFromProblem with a Windows home given as typed (filepath.Join would use this system's
// separator).
func personalFromCheck(p, home string) (string, string) {
	return p, personalFromProblem(p, home)
}

// openAt opens a scope's settings on a given operating system.
func openAt(t *testing.T, scope, home, main, goos string) *SettingsFile {
	t.Helper()
	f, err := openSettings(scope, home, main, goos)
	if err != nil {
		t.Fatal(err)
	}
	return f
}

// settings set: all or nothing, its rules between the two workspace keys, byte-stable files with unknown keys kept,
// and nothing written when every value is already in force.
func TestSettingsSet(t *testing.T) {
	home, main := t.TempDir(), t.TempDir()
	f := openAt(t, ScopeWorkspace, home, main, "linux")
	in, err := f.InForce()
	if err != nil || len(in) != 2 || in[0].Value != "agents" || in[0].From != FromDefault || in[1].Value != nil {
		t.Fatalf("no file: %+v %v", in, err)
	}
	refused := func(pairs ...string) {
		t.Helper()
		before, _ := os.ReadFile(f.Path)
		if _, err := f.Set(pairs); err == nil {
			t.Errorf("%v: not refused", pairs)
		} else if _, ok := err.(*SettingError); !ok {
			t.Errorf("%v: %T %v", pairs, err, err)
		}
		after, _ := os.ReadFile(f.Path)
		if string(before) != string(after) {
			t.Errorf("%v: refused, but the file changed:\n%s", pairs, after)
		}
	}
	refused()
	refused("status_writes")
	refused("=agents")
	refused("status_writes=command")                                        // no command given or set
	refused("status_command=studio task move")                              // with agents
	refused("status_writes=command", "status_command=studio; rm")           // the second bad: nothing written
	refused("status_writes=command", "status_command=a", "nothing=1")       // an unknown key: nothing written
	refused("cache_keep_days=30")                                           // the home's key without --machine
	refused("status_writes=agents", "status_writes=command")                // twice
	refused("status_writes=command", "status_command=a", "status_writes=x") // twice, and bad
	if _, err := os.Stat(f.Path); err == nil {
		t.Fatalf("a refused set wrote %s", f.Path)
	}
	changed, err := f.Set([]string{"status_writes=command", "status_command=studio task move"})
	if err != nil || strings.Join(changed, ",") != "status_writes,status_command" {
		t.Fatalf("set: %v %v", changed, err)
	}
	want := "{\n  \"status_writes\": \"command\",\n  \"status_command\": \"studio task move\"\n}\n"
	if raw, _ := os.ReadFile(f.Path); string(raw) != want {
		t.Errorf("the file:\n%s\nwant:\n%s", raw, want)
	}
	if s, err := LoadMachineSettings(home, main); err != nil || s.StatusWrites != "command" || s.StatusCommand != "studio task move" {
		t.Errorf("the guard's reader: %+v %v", s, err)
	}
	// The same values again: nothing written (the file's time is left alone).
	stamp := mustStat(t, f.Path)
	if changed, err := openAt(t, ScopeWorkspace, home, main, "linux").Set([]string{"status_writes=command"}); err != nil || changed == nil || len(changed) != 0 {
		t.Errorf("again: %v %v", changed, err)
	}
	if mustStat(t, f.Path) != stamp {
		t.Errorf("a set of values already in force wrote the file")
	}
	// A key this Bonsai does not know is kept as read, after the known keys, in its place; agents drops the command.
	machineFile(t, home, main, MachineSettingsFile, "{\"later\": [1, \"x\"], \"status_command\": \"a b\", \"status_writes\": \"command\"}")
	f = openAt(t, ScopeWorkspace, home, main, "linux")
	if changed, err := f.Set([]string{"status_writes=agents"}); err != nil || strings.Join(changed, ",") != "status_writes,status_command" {
		t.Fatalf("agents: %v %v", changed, err)
	}
	want = "{\n  \"status_writes\": \"agents\",\n  \"later\": [\n    1,\n    \"x\"\n  ]\n}\n"
	if raw, _ := os.ReadFile(f.Path); string(raw) != want {
		t.Errorf("the file:\n%s\nwant:\n%s", raw, want)
	}
	// The home's own: cache_keep_days written as a number, none takes it out; personal_from refused off Windows.
	m := openAt(t, ScopeMachine, home, "", "linux")
	if changed, err := m.Set([]string{"cache_keep_days=90"}); err != nil || strings.Join(changed, ",") != "cache_keep_days" {
		t.Fatalf("cache_keep_days: %v %v", changed, err)
	}
	if raw, _ := os.ReadFile(filepath.Join(home, HomeSettingsFile)); string(raw) != "{\n  \"cache_keep_days\": 90\n}\n" {
		t.Errorf("the home's file:\n%s", raw)
	}
	if hs, err := LoadHomeSettings(home); err != nil || hs.CacheKeepDays != 90 {
		t.Errorf("LoadHomeSettings: %+v %v", hs, err)
	}
	m = openAt(t, ScopeMachine, home, "", "linux")
	if _, err := m.Set([]string{"personal_from=//wsl.localhost/x/home/u/.bonsai/personal"}); err == nil {
		t.Errorf("personal_from off Windows: not refused")
	}
	if changed, err := m.Set([]string{"cache_keep_days=none"}); err != nil || strings.Join(changed, ",") != "cache_keep_days" {
		t.Fatalf("none: %v %v", changed, err)
	}
	if raw, _ := os.ReadFile(filepath.Join(home, HomeSettingsFile)); string(raw) != "{}\n" {
		t.Errorf("the home's file after none:\n%s", raw)
	}
	// On Windows the home's scope has personal_from too, in the table's order.
	w := openAt(t, ScopeMachine, home, "", "windows")
	if changed, err := w.Set([]string{"personal_from=//wsl.localhost/x/home/u/.bonsai/personal", "cache_keep_days=7"}); err != nil ||
		strings.Join(changed, ",") != "cache_keep_days,personal_from" {
		t.Fatalf("windows: %v %v", changed, err)
	}
	if raw, _ := os.ReadFile(filepath.Join(home, HomeSettingsFile)); string(raw) != "{\n  \"cache_keep_days\": 7,\n  \"personal_from\": \"//wsl.localhost/x/home/u/.bonsai/personal\"\n}\n" {
		t.Errorf("the home's file on Windows:\n%s", raw)
	}
}

// fileStamp is a file's time and size, to see that a call left it alone.
type fileStamp struct{ ns, size int64 }

func mustStat(t *testing.T, p string) fileStamp {
	t.Helper()
	fi, err := os.Stat(p)
	if err != nil {
		t.Fatal(err)
	}
	return fileStamp{fi.ModTime().UnixNano(), fi.Size()}
}

// The reader applies the writer's rules: a value outside its key's rule, written straight into the file, is refused
// naming the file and the key, by settings show's reader and by LoadMachineSettings, which the guard reads.
func TestSettingsReaderRules(t *testing.T) {
	home, main := t.TempDir(), t.TempDir()
	for _, c := range []struct{ content, key string }{
		{`{"status_writes": "command", "status_command": "move; rm -rf ~"}`, "status_command"},
		{`{"status_writes": "command", "status_command": "move $(id)"}`, "status_command"},
		{"{\"status_writes\": \"command\", \"status_command\": \"move `id`\"}", "status_command"},
		{`{"status_writes": "command", "status_command": "move\nnext"}`, "status_command"},
		{`{"status_writes": "command", "status_command": "` + strings.Repeat("a", 101) + `"}`, "status_command"},
		{`{"status_writes": "command", "status_command": 5}`, "status_command"},
		{`{"status_writes": "sometimes"}`, "status_writes"},
	} {
		p := machineFile(t, home, main, MachineSettingsFile, c.content)
		show := filepath.ToSlash(p)
		if _, err := LoadMachineSettings(home, main); err == nil || !strings.Contains(err.Error(), show) || !strings.Contains(err.Error(), c.key) {
			t.Errorf("LoadMachineSettings %s: %v", c.content, err)
		}
		f, err := OpenSettings(ScopeWorkspace, home, main)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := f.InForce(); err == nil || !strings.Contains(err.Error(), show) || !strings.Contains(err.Error(), c.key) {
			t.Errorf("InForce %s: %v", c.content, err)
		}
	}
	for _, c := range []string{`{"cache_keep_days": 0}`, `{"cache_keep_days": 1.5}`, `{"cache_keep_days": "30"}`, `{"cache_keep_days": 4000}`} {
		if err := os.WriteFile(filepath.Join(home, HomeSettingsFile), []byte(c), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := LoadHomeSettings(home); err == nil || !strings.Contains(err.Error(), "cache_keep_days") ||
			!strings.Contains(err.Error(), filepath.ToSlash(filepath.Join(home, HomeSettingsFile))) {
			t.Errorf("LoadHomeSettings %s: %v", c, err)
		}
	}
	if err := os.WriteFile(filepath.Join(home, HomeSettingsFile), []byte(`{"cache_keep_days": null}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if hs, err := LoadHomeSettings(home); err != nil || hs.CacheKeepDays != 0 {
		t.Errorf("null: %+v %v", hs, err)
	}
}
