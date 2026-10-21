package engine

// The hook lines' form (plan-5 5.3.6 note 2): what isBonsaiHook claims and isOldBonsaiHook takes, a test build's
// stamps and its confinement, and the plan's changes of Bonsai's own guard and stop lines (note 8).

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/testpack"
)

// stampAs makes this test a test build stamped with place and root, its own lines made again; restored after it.
func stampAs(t *testing.T, place, root string) {
	t.Helper()
	oldPlace, oldRoot := hookPlace, scratchRoot
	remake := func() {
		GuardCommand = HookCommand("guard", true, hookPlaces()...)
		SecondCommand = HookCommand(SecondWord, false, hookPlaces()...)
		StopCommand = HookCommand("stop", true, hookPlaces()...)
		StartCommand = HookCommand("start", false, hookPlaces()...)
		RecordCommand = HookCommand("record", false, hookPlaces()...)
		NowCommand = OpenHookCommand("now", hookPlaces()...)
		ownHooks, ownRolesHooks = buildOwnHooks(false), buildOwnHooks(true)
	}
	hookPlace, scratchRoot = place, root
	remake()
	t.Cleanup(func() { hookPlace, scratchRoot = oldPlace, oldRoot; remake() })
}

func TestTheFormIsClaimedExactly(t *testing.T) {
	installed := []string{PlaceWindows, PlaceLinux}
	claimed := map[string]bool{
		GuardCommand: true, StopCommand: true, StartCommand: true, RecordCommand: true, SecondCommand: true, NowCommand: true,
		OpenHookCommand("now", "/home/a b/bonsai-checks/x/bonsai"):           true, // the open form, a test build's one place
		OpenHookCommand("ladder", installed...):                              true, // the open form, any word
		strings.Replace(NowCommand, " || exit 0", " || exit 1", 1):           false,
		strings.Replace(NowCommand, " || exit 0", " || exit 0 || exit 0", 1): false,
		strings.Replace(NowCommand, "[ -x \"$b\" ] || exit 0; ", "", 1):      false,
		NowCommand + "; rm -rf x":                                            false,
		HookCommand("guard --second", true, installed...):                    true, // a flag after the word: still Bonsai's form
		HookCommand("guard --Second", false, installed...):                   false,
		HookCommand("guard --second --x", false, installed...):               false,
		HookCommand("guard -- second", false, installed...):                  false,
		HookCommand("guard --", false, installed...):                         false,
		HookCommand("guard", true, "/home/a b/bonsai-checks/x/bonsai"):       true, // a test build's one place, with a space
		HookCommand("guard", true, "C:/Users/a/bonsai-checks/bonsai.exe"):    true,
		HookCommand("guard", true, "/c/one/bonsai", "/c/two/bonsai"):         true,
		HookCommand("ladder", true, installed...):                            true, // a word this build does not write: Bonsai's still
		"bonsai hook guard || exit 2":                                        true, // the form before step 5.3.6
		"bonsai hook record":                                                 true,
		// Not Bonsai's: the form with anything changed.
		strings.Replace(GuardCommand, "a person installs Bonsai there", "anyone may", 1):                                     false,
		strings.Replace(GuardCommand, `b="/c/Program Files/Bonsai/bonsai.exe"`, `b='/c/Program Files/Bonsai/bonsai.exe'`, 1): false,
		strings.Replace(GuardCommand, `b=/usr/local/bin/bonsai`, `b="/usr/local/bin/bonsai"`, 1):                             false,
		GuardCommand + "; rm -rf x":                                       false,
		"echo hi; " + GuardCommand:                                        false,
		strings.Replace(GuardCommand, " || exit 2", "", 1):                false,
		strings.Replace(StartCommand, "exit 0", "exit 2", 1):              false,
		HookCommand("guard", true, "/a/bonsai", "/b/bonsai", "/c/bonsai"): false, // three places
		HookCommand("guard", true, "relative/bonsai"):                     false,
		HookCommand("guard", true, "/a/$HOME/bonsai"):                     false,
		HookCommand("Guard", true, installed...):                          false,
		`b=/usr/local/bin/bonsai; "$b" hook guard || exit 2`:              false,
		"/usr/local/bin/bonsai hook guard":                                false,
		"":                                                                false,
	}
	for cmd, want := range claimed {
		if got := isBonsaiHook(cmd); got != want {
			t.Errorf("isBonsaiHook(%q) = %v", cmd, got)
		}
		if isBonsaiHook(cmd) && isOldBonsaiHook(cmd) {
			t.Errorf("Bonsai's own line %q is taken for an old one", cmd)
		}
	}
	for cmd, want := range map[string]bool{
		"/usr/local/bin/bonsai hook guard":            true, // a 0.4.3 line naming the installed place by its full path
		`"/c/Program Files/Bonsai/bonsai.exe" hook x`: true,
		`"C:\Program Files\Bonsai\bonsai.exe" hook x`: true,
		GuardCommand: false,
		HookCommand("guard", true, "/opt/x/bonsai"): false,
	} {
		if got := isOldBonsaiHook(cmd); got != want {
			t.Errorf("isOldBonsaiHook(%q) = %v", cmd, got)
		}
	}
	// The words for a bonsai that is not there never read as a hook that failed to start.
	for _, w := range []string{"guard", "stop", "start", "record"} {
		low := strings.ToLower(HookCommand(w, true, installed...))
		for _, bad := range []string{"no such file", "not found", "missing", "cannot find", "enoent", "does not exist", "failed to start"} {
			if strings.Contains(low, bad) {
				t.Errorf("the %s line says %q", w, bad)
			}
		}
	}
	// The matcher is the table's: without the roles' tools where the lock declares none, with them where it does.
	var names, all []string
	for _, tool := range GuardTools {
		all = append(all, tool.Name)
		if !tool.Roles {
			names = append(names, tool.Name)
		}
		if tool.Why == "" {
			t.Errorf("%s has no reason", tool.Name)
		}
	}
	if GuardMatcher != strings.Join(names, "|") || !strings.HasSuffix(GuardMatcher, "|Monitor") {
		t.Errorf("the guard's matcher %q", GuardMatcher)
	}
	if GuardRolesMatcher != strings.Join(all, "|") || GuardRolesMatcher != GuardMatcher+"|Agent|Task|SendMessage|Workflow" {
		t.Errorf("the guard's matcher where roles are declared %q", GuardRolesMatcher)
	}
	if ownHooks[0].OnFailure != "block" || ownHooks[0].Matcher != GuardMatcher {
		t.Errorf("the guard's line %+v", ownHooks[0])
	}
	for _, l := range ownHooks[1:] {
		if l.OnFailure != "" {
			t.Errorf("onFailure on %s, which it does not apply to or need", l.Text())
		}
	}
	// Where the lock declares roles: the guard's line widened, its form, timeout and onFailure unchanged, and the second
	// line after it, in the never-blocking form, on every other tool; the rest as without roles.
	r := ownRolesHooks
	if len(r) != len(ownHooks)+1 || r[0].Command != ownHooks[0].Command || r[0].Timeout != ownHooks[0].Timeout ||
		r[0].OnFailure != "block" || r[0].Matcher != GuardRolesMatcher {
		t.Fatalf("the roles' guard line %+v", r[0])
	}
	if r[1].Event != "PreToolUse" || r[1].Matcher != SecondMatcher || r[1].Command != SecondCommand || r[1].OnFailure != "" ||
		r[1].Async || r[1].Timeout != "10" || strings.Contains(r[1].Command, "exit 2") ||
		!strings.HasSuffix(r[1].Command, `[ -x "$b" ] || exit 0; "$b" hook guard --second`) {
		t.Errorf("the second line %+v", r[1])
	}
	for i, l := range ownHooks[1:] {
		if r[i+2].canon() != l.canon() {
			t.Errorf("line %d differs where roles are declared: %s", i+2, r[i+2].Text())
		}
	}
	if want := "^(?!Edit$)(?!Write$)(?!MultiEdit$)(?!NotebookEdit$)(?!Bash$)(?!PowerShell$)(?!Monitor$)(?!Agent$)(?!Task$)" +
		"(?!SendMessage$)(?!Workflow$).*"; SecondMatcher != want {
		t.Errorf("the second line's matcher %q", SecondMatcher)
	}
}

// The second line's matcher, read as Claude Code reads a matcher that is not a list of names (a JavaScript regular
// expression; Go's regexp has no lookahead, so each name is tried by hand): every tool the first line names is left
// out, every other one met, an MCP tool and a later tool among them.
func TestTheSecondLineMeetsEveryOtherTool(t *testing.T) {
	first := map[string]bool{}
	for _, tool := range GuardTools {
		first[tool.Name] = true
	}
	meets := func(tool string) bool {
		// ^(?!A$)(?!B$)...(?!Z$).* : a name is met unless it is exactly one of the names.
		rest := strings.TrimSuffix(strings.TrimPrefix(SecondMatcher, "^"), ".*")
		for _, part := range strings.Split(strings.TrimSuffix(strings.TrimPrefix(rest, "(?!"), "$)"), "$)(?!") {
			if part == tool {
				return false
			}
		}
		return true
	}
	for _, tool := range []string{"Read", "Grep", "Glob", "WebFetch", "Skill", "ToolSearch", "AskUserQuestion",
		"mcp__claude_ai_Notion__notion-search", "SubagentHandback", "Edits", "BashOutput", "AgentX", "ATask"} {
		if !meets(tool) {
			t.Errorf("the second line does not meet %s", tool)
		}
	}
	for tool := range first {
		if meets(tool) {
			t.Errorf("the second line meets %s, which the first judges", tool)
		}
	}
}

// A plain build's stamps are empty; a stamped one writes its place alone, and is confined to its scratch root.
func TestAStampedBuildIsConfined(t *testing.T) {
	if Stamped() || hookPlace != "" || scratchRoot != "" {
		t.Fatalf("this test binary is stamped: %q %q", hookPlace, scratchRoot)
	}
	e := setup(t)
	root := filepath.Join(e.tmp, "scratch")
	place := "/srv/scratch place/bonsai"
	stampAs(t, place, root)
	if GuardCommand != HookCommand("guard", true, place) || !strings.HasPrefix(GuardCommand, `b="/srv/scratch place/bonsai"; [ -x "$b" ] || { echo`) {
		t.Fatalf("a stamped build's line: %s", GuardCommand)
	}
	inside := testpack.Project(t, root, "inside")
	outside := testpack.Project(t, e.tmp, "outside")
	if p := e.plan(t, inside, Request{Command: "init", Init: e.values(e.pack.A), AllowExec: true}); !strings.Contains(flat(p.Preview(false)), flat(GuardCommand)) {
		t.Errorf("inside its scratch root, no line naming its place:\n%s", p.Preview(false))
	}
	for _, build := range []func() error{
		func() error {
			_, err := Build(Request{Command: "init", Dir: outside, Home: e.home, Init: e.values(e.pack.A)})
			return err
		},
		func() error { _, err := BuildUnlink(outside, e.home); return err },
	} {
		var ee *Error
		if err := build(); !errors.As(err, &ee) || ee.Code != "test-build" || ee.Exit != ExitState || !strings.Contains(ee.What, "only inside its scratch folder") {
			t.Errorf("outside its scratch root: %v", err)
		}
	}
	// A link into the scratch root from outside it, and a worktree of an outside main inside it: refused.
	if err := os.Symlink(outside, filepath.Join(root, "link")); err == nil {
		if _, err := Build(Request{Command: "init", Dir: filepath.Join(root, "link"), Home: e.home, Init: e.values(e.pack.A)}); err == nil {
			t.Error("a link inside the scratch root to a folder outside it was let through")
		}
	} // a link needs a privilege on Windows: the case runs where one can be made
	testpack.Git(t, outside, "commit", "-q", "--allow-empty", "-m", "x")
	testpack.Git(t, outside, "worktree", "add", "-q", filepath.Join(root, "wt"))
	if _, err := Build(Request{Command: "update", Dir: filepath.Join(root, "wt"), Home: e.home}); err == nil || !strings.Contains(err.Error(), "test build") {
		t.Errorf("a worktree inside, its main outside: %v", err)
	}
	// One stamp alone, or a place a line cannot name: refused everywhere.
	for _, s := range [][2]string{{place, ""}, {"", root}, {"relative/bonsai", root}, {`/a/"b"/bonsai`, root}, {"/a/bonsai", "relative"}} {
		hookPlace, scratchRoot = s[0], s[1]
		if err := Confine(inside); err == nil {
			t.Errorf("stamps %q: confined nothing", s)
		}
	}
}

// A scratch root reached through a link confines a project folder not made yet to the root it names (CI's Windows
// runner reaches its temporary folder by an 8.3 short name, which resolves the same way); a link inside the root to a
// folder outside it, its target not made yet, stays outside.
func TestConfineLinkedRoot(t *testing.T) {
	real, outside := t.TempDir(), t.TempDir()
	link := filepath.Join(t.TempDir(), "link")
	if err := os.Symlink(real, link); err != nil {
		t.Skipf("no link can be made here (on Windows it needs a privilege): %v", err)
	}
	stampAs(t, "/srv/scratch/bonsai", link)
	for _, d := range []string{filepath.Join(link, "new"), filepath.Join(link, "a", "new"), filepath.Join(real, "new"), link} {
		if err := Confine(d); err != nil {
			t.Errorf("%s, inside the linked root: %v", d, err)
		}
	}
	away := filepath.Join(real, "away")
	if err := os.Symlink(filepath.Join(outside, "not-made"), away); err != nil {
		t.Fatal(err)
	}
	for _, d := range []string{away, filepath.Join(away, "new"), filepath.Join(link, "..", "elsewhere")} {
		var ee *Error
		if err := Confine(d); !errors.As(err, &ee) || ee.Code != "test-build" {
			t.Errorf("%s, outside the root: %v, want test-build", d, err)
		}
	}
}

// The plan's changes of Bonsai's own guard or stop line (OwnGuardChanges): a change or a removal of either, never an
// addition, never the recorder's lines on the same events.
func TestOwnGuardChanges(t *testing.T) {
	p := &Plan{Settings: []SettingsChange{
		{Change: "add", Kind: "hook", Own: true, Line: ownHooks[0].Text(), event: GuardEvent},
		{Change: "change", Kind: "hook", Own: true, Line: ownHooks[2].Text(), Was: "SessionStart: bonsai hook start", event: StartEvent},
		{Change: "remove", Kind: "hook", Own: true, Line: "PreToolUse: bonsai hook record [async]", event: GuardEvent},
		{Change: "change", Kind: "hook", Own: true, Line: "Stop: " + RecordCommand + " [async]", Was: "Stop: bonsai hook record [async]", event: StopEvent},
		{Change: "remove", Kind: "hook", Own: false, Line: "Stop: bonsai hook stop || exit 2", event: StopEvent},
	}}
	if got := p.OwnGuardChanges(); len(got) != 0 {
		t.Errorf("none of these is a change of the guard's or stop gate's line: %+v", got)
	}
	p.Settings = append(p.Settings,
		SettingsChange{Change: "remove", Kind: "hook", Own: true, Line: "PreToolUse (Edit|Write): bonsai hook guard || exit 2", event: GuardEvent},
		SettingsChange{Change: "change", Kind: "hook", Own: true, Line: ownHooks[1].Text(), Was: "Stop: bonsai hook stop || exit 2", event: StopEvent})
	if got := p.OwnGuardChanges(); len(got) != 2 {
		t.Errorf("the guard's line taken out and the stop gate's changed: %+v", got)
	}
}
