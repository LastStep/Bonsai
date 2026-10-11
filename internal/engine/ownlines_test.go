package engine

// Bonsai's own lines across builds. Its hook lines (step 5.2.4; plan-5 5.2.4 note 2; step 5.3.4 note 6): a project
// linked by a build that wrote the guard's line only gets the stop gate's, start's and the recorder's lines at its
// next update only with --allow-exec and --yes, and a project linked by a build before the stop gate (step 5.3.4) gets
// its line the same way; check warns (own-hooks) until then. Its own deny rules (step 5.3.5): written at a first link and taken by a
// project linked before at its next update on --yes alone, since a deny rule runs no code, each previewed with its
// sentence, the file's key order kept; never_edit's rules move to the form that anchors at the project's top folder.
// Taking a pack out keeps Bonsai's own lines, which are not the pack's, and unlink takes every one of them out.

import (
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/testpack"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// linkedBefore rewrites a project this build linked as a build before step 5.2.4 left it: in .claude/settings.json
// Bonsai's own hook lines are the guard's only, and the lock's fingerprint is of the lines that build wrote.
func linkedBefore(t *testing.T, root string) {
	t.Helper()
	linkedWithout(t, root, func(l Line) bool { return l.Command != GuardCommand })
}

// linkedWithout rewrites a project this build linked as a build without some of Bonsai's own hook lines left it: the
// lines drop names are taken out of .claude/settings.json, and the lock's fingerprint is of the lines left, as that
// build wrote it.
func linkedWithout(t *testing.T, root string, drop func(Line) bool) {
	t.Helper()
	sd, err := readSettings(root)
	if err != nil {
		t.Fatal(err)
	}
	lock, err := workspace.LoadLock(root)
	if err != nil {
		t.Fatal(err)
	}
	// Every line of the file is Bonsai's here (a fresh link, no line of the project's own): the file's lines, a
	// marketplace in Bonsai's own order (the fingerprint's), are the reference claim finds them by.
	disk := diskLines(sd.root)
	ref := make([]Line, len(disk))
	for i, l := range disk {
		if l.Kind == "marketplace" {
			l.Value = bonsaiOrdered(l.Value)
		}
		ref[i] = l
	}
	claimed, same := claim(disk, ref, false, lock.Files[SettingsFile].SHA256)
	if !same {
		t.Fatal("the settings file does not hold the lines the lock records")
	}
	var old, dropped []Line
	want := 0
	for _, l := range ownHooks {
		if drop(l) {
			want++
		}
	}
	for _, l := range claimed {
		if l.Kind == "hook" && isBonsaiHook(l.Command) && drop(l) {
			dropped = append(dropped, l)
			continue
		}
		old = append(old, l)
	}
	if len(dropped) != want || want == 0 {
		t.Fatalf("%d own lines to take out, want %d", len(dropped), want)
	}
	b, err := schema.EncodeUTF8(applyLines(sd.root, dropped, nil))
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, root, SettingsFile, string(b))
	raw := read(t, root, workspace.LockFile)
	writeFile(t, root, workspace.LockFile, strings.Replace(raw, lock.Files[SettingsFile].SHA256, linesHash(old), 1))
}

func TestOwnLinesOfAProjectLinkedBefore(t *testing.T) {
	e := setup(t)
	side, sc := testpack.SidePack(t, e.tmp)
	root := testpack.Project(t, e.tmp, "before")
	writeFile(t, root, "bonsai.yaml", twoPacks(e, side, sc[0]))
	e.apply(t, root, Request{Command: "init", Init: &InitValues{}, AllowExec: true})
	linkedBefore(t, root)
	before := read(t, root, SettingsFile)
	if strings.Contains(before, StopCommand) || strings.Contains(before, StartCommand) || strings.Contains(before, RecordCommand) ||
		!strings.Contains(before, GuardCommand) {
		t.Fatalf("not as a build before 5.2.4 left it:\n%s", before)
	}

	// check: a warning naming the twelve lines, its next step the consent; no finding but the two local sources'.
	r, err := Check(root, e.home)
	if err != nil {
		t.Fatal(err)
	}
	var warned *Finding
	for i, w := range r.Warnings {
		if w.Code == "own-hooks" {
			warned = &r.Warnings[i]
		}
	}
	if warned == nil || warned.Next != "run: bonsai update --allow-exec --yes" || warned.Who != "person" || warned.File != SettingsFile ||
		!strings.Contains(warned.Message, "12 to add (Stop: bonsai hook stop || exit 2; SessionStart: bonsai hook start; "+
			"UserPromptSubmit: bonsai hook record [async];") {
		t.Fatalf("the own-hooks warning: %+v", r.Warnings)
	}
	for _, f := range r.Findings {
		if f.Code == "changed" {
			t.Errorf("a changed finding: %+v", f)
		}
	}

	// update --yes is refused whole: each new line runs code (rule 1), named under Runs code; nothing is written.
	p := e.plan(t, root, Request{})
	if got := items(p); !p.NeedsExec() || len(p.RunsCode) != len(ownHooks)-1 ||
		!strings.HasPrefix(got, "hook add Stop: bonsai hook stop || exit 2 (bonsai); hook add SessionStart: bonsai hook start (bonsai); ") {
		t.Fatalf("runs code: %s", got)
	}
	if err := Apply(p); err == nil {
		t.Fatal("update --yes wrote the new lines without --allow-exec")
	}
	if read(t, root, SettingsFile) != before {
		t.Error("a refused update wrote the settings file")
	}

	// Taking a pack out in the same run: still refused without --allow-exec, since the run adds code.
	cfg := read(t, root, "bonsai.yaml")
	writeFile(t, root, "bonsai.yaml", cfg[:strings.Index(cfg, "  - id: side-pack")])
	if p := e.plan(t, root, Request{}); len(p.Removed) != 1 || !p.NeedsExec() {
		t.Fatalf("take-out with the new lines: removed %+v, runs code %v", p.Removed, p.RunsCode)
	}
	e.apply(t, root, Request{AllowExec: true})
	after := read(t, root, SettingsFile)
	for _, l := range ownHooks {
		if !strings.Contains(after, `"`+l.Command+`"`) {
			t.Errorf("%s is not written", l.Text())
		}
	}
	if strings.Contains(after, "side") || !strings.Contains(after, "echo demo hook A") {
		t.Errorf("the side pack's lines stay, or the demo pack's went:\n%s", after)
	}
	if r, _ := checkLocal(t, root, e.home); hasWarning(r, "own-hooks") || len(r.Findings) != 0 {
		t.Errorf("check after the update: %+v %+v", r.Findings, r.Warnings)
	}

	// With every own line in place, taking the last pack's sibling out again changes none of them, and unlink takes
	// every one of them out.
	writeFile(t, root, "bonsai.yaml", twoPacks(e, side, sc[0]))
	e.apply(t, root, Request{AllowExec: true})
	writeFile(t, root, "bonsai.yaml", cfg[:strings.Index(cfg, "  - id: side-pack")])
	p = e.plan(t, root, Request{})
	for _, c := range p.Settings {
		if c.Own || strings.Contains(c.Line, ": bonsai hook ") {
			t.Errorf("taking a pack out touches Bonsai's own line %+v", c)
		}
	}
	if p.NeedsExec() {
		t.Errorf("taking a pack out runs code: %s", items(p))
	}
	e.apply(t, root, Request{})
	for _, l := range ownHooks {
		if !strings.Contains(read(t, root, SettingsFile), `"`+l.Command+`"`) {
			t.Errorf("taking a pack out took %s", l.Text())
		}
	}
	for _, l := range ownDeny {
		if !strings.Contains(read(t, root, SettingsFile), `"`+l.Rule+`"`) {
			t.Errorf("taking a pack out took %s", l.Rule)
		}
	}
	u := unlinkPlan(t, e, root)
	removed := map[string]bool{}
	for _, c := range u.Settings {
		removed[c.Line] = c.Change == "remove"
	}
	for _, l := range append(append([]Line{}, ownHooks...), ownDeny...) {
		if !removed[l.Text()] {
			t.Errorf("unlink leaves %s", l.Text())
		}
	}
	if err := Apply(u); err != nil {
		t.Fatal(err)
	}
	if s, err := readSettings(root); err != nil || (s.exists && (strings.Contains(string(s.raw), "bonsai hook") || strings.Contains(string(s.raw), ".bonsai"))) {
		t.Errorf("after unlink: %v", err)
	}
}

// The stop gate's line (step 5.3.4 note 6): bonsai hook stop || exit 2 on Stop, timeout 10, no matcher, with its
// sentence. A first link writes it on --yes alone, among Bonsai's own lines; a project linked by a build before it
// (every other own line in place) takes it at its next update only with --allow-exec and --yes, the update refused
// whole and nothing written without, and check warns until then; unlink takes it out.
func TestStopLineOfAProjectLinkedBefore(t *testing.T) {
	var stop *Line
	for i, l := range ownHooks {
		if l.Event == StopEvent && l.Command == StopCommand {
			stop = &ownHooks[i]
		}
	}
	if stop == nil || stop.Matcher != "" || stop.Timeout != StopTimeout || stop.Async || !stop.Own || stop.Origin != "bonsai" ||
		!strings.HasSuffix(stop.Why, ".") || ascii(stop.Why) != stop.Why || StopTimeout != "10" || StopCommand != "bonsai hook stop || exit 2" {
		t.Fatalf("the stop gate's line %+v", stop)
	}

	e := setup(t)
	quiet, qc := testpack.QuietPack(t, e.tmp)
	root := testpack.Project(t, e.tmp, "stop")
	p := e.linkTo(t, root, "stop", quiet, qc[0], false)
	if p.NeedsExec() || !strings.Contains(p.Preview(false), "  add     hook        Stop: "+StopCommand+"\n          "+stop.Why+"\n") {
		t.Fatalf("a first link, the stop line on --yes alone:\n%s", p.Preview(false))
	}
	if err := Apply(p); err != nil {
		t.Fatal(err)
	}
	hooks, _ := settingsDocOf(t, root).Get("hooks")
	groups, _ := hooks.(schema.Object).Get(StopEvent)
	// One group with no matcher, the stop gate's line first, the recorder's async line on Stop beside it.
	if got := schema.Show(groups); !strings.HasPrefix(got, `[{"hooks":[{"type":"command","command":"`+StopCommand+`","timeout":10},`) ||
		strings.Contains(got, `"matcher"`) {
		t.Fatalf("the Stop groups as written: %s", got)
	}

	linkedWithout(t, root, func(l Line) bool { return l.Command == StopCommand })
	before := read(t, root, SettingsFile)
	if strings.Contains(before, StopCommand) || !strings.Contains(before, GuardCommand) || !strings.Contains(before, StartCommand) {
		t.Fatalf("not as a build before 5.3.4 left it:\n%s", before)
	}
	r, err := Check(root, e.home)
	if err != nil {
		t.Fatal(err)
	}
	warned := false
	for _, w := range r.Warnings {
		if w.Code == "own-hooks" {
			warned = w.Message == "Bonsai's own hook lines in "+SettingsFile+" are not this build's: 1 to add (Stop: "+StopCommand+
				"); until then this project's sessions run without them (a new hook line runs code, so it needs a person's --allow-exec)" &&
				w.Next == "run: bonsai update --allow-exec --yes" && w.Who == "person"
		}
	}
	if !warned {
		t.Errorf("the own-hooks warning: %+v", r.Warnings)
	}

	// update --yes: refused whole, the stop line named under Runs code, nothing written.
	p = e.plan(t, root, Request{})
	if got := items(p); !p.NeedsExec() || got != "hook add Stop: "+StopCommand+" (bonsai)" || len(p.OwnHooks) != 0 {
		t.Fatalf("runs code: %s; own at a first link %+v", got, p.OwnHooks)
	}
	if err := Apply(p); err == nil {
		t.Fatal("update --yes wrote the stop line without --allow-exec")
	}
	if read(t, root, SettingsFile) != before {
		t.Error("a refused update wrote the settings file")
	}

	// update --allow-exec --yes writes it, and check is quiet.
	e.apply(t, root, Request{AllowExec: true})
	if !strings.Contains(read(t, root, SettingsFile), `"`+StopCommand+`"`) {
		t.Errorf("the stop line is not written:\n%s", read(t, root, SettingsFile))
	}
	if r, _ := checkLocal(t, root, e.home); hasWarning(r, "own-hooks") || len(r.Findings) != 0 {
		t.Errorf("check after the update: %+v %+v", r.Findings, r.Warnings)
	}

	// unlink takes it out.
	u := unlinkPlan(t, e, root)
	removed := false
	for _, c := range u.Settings {
		if c.Line == "Stop: "+StopCommand {
			removed = c.Change == "remove" && !c.RunsCode
		}
	}
	if !removed {
		t.Errorf("unlink does not remove the stop line: %+v", u.Settings)
	}
	if err := Apply(u); err != nil {
		t.Fatal(err)
	}
	if s, err := readSettings(root); err != nil || (s.exists && strings.Contains(string(s.raw), "bonsai hook stop")) {
		t.Errorf("after unlink the stop line stays (%v)", err)
	}
}

func hasWarning(r *CheckResult, code string) bool {
	if r == nil {
		return false
	}
	for _, w := range r.Warnings {
		if w.Code == code {
			return true
		}
	}
	return false
}

// Bonsai's own deny rules (step 5.3.5): one over .bonsai/local/ at any depth, one for each table check --write
// rebuilds (from format.GeneratedKinds), one over Bonsai's home; each valid as Claude Code reads a rule, each with
// its sentence. A first link to a pack with no hook line writes them on --yes alone, Bonsai's own deny rules first,
// then never_edit's (in the form anchored at the project's top folder), each named in the preview with its sentence
// and none under runs code; check is quiet after.
func TestOwnDenyRules(t *testing.T) {
	// The tables are the files check --write writes, and nothing else: the one home is format.GeneratedKinds.
	if got := strings.Join(tableFiles(), " "); got != workspace.SessionsTableFile+" "+workspace.TasksTableFile {
		t.Errorf("tableFiles %q", got)
	}
	var rules []string
	for _, l := range ownDeny {
		rules = append(rules, l.Rule)
		if l.Kind != "deny" || l.Origin != "bonsai" || !l.Own || !strings.HasSuffix(l.Why, ".") || ascii(l.Why) != l.Why {
			t.Errorf("own deny line %+v", l)
		}
		if why := ruleProblem(l.Rule); why != "" {
			t.Errorf("%s: %s", l.Rule, why)
		}
	}
	want := "Edit(//**/.bonsai/local/**) Edit(//**/.bonsai/sessions.md) Edit(//**/.bonsai/tasks.md) Edit(~/.bonsai/**)"
	if got := strings.Join(rules, " "); got != want {
		t.Fatalf("own deny rules\n  %s\nwant\n  %s", got, want)
	}
	if got := neverEditRule("work/ledger.json"); got != "Edit(/work/ledger.json)" {
		t.Errorf("never_edit's rule %s", got)
	}

	e := setup(t)
	quiet, qc := testpack.QuietPack(t, e.tmp)
	root := testpack.Project(t, e.tmp, "deny")
	p := e.apply(t, root, Request{Command: "init", Init: &InitValues{Name: "deny", Source: quiet, Ref: qc[0], NeverEdit: []string{"ledger.json"}}})
	if p.NeedsExec() || len(p.RunsCode) != 0 {
		t.Fatalf("a first link with no pack code runs code: %s", items(p))
	}
	deny, _ := getList(settingsDocOf(t, root), "permissions", "deny")
	if got := schema.Show(deny); got != `["`+strings.Join(rules, `","`)+`","Edit(/ledger.json)"]` {
		t.Errorf("deny list %s", got)
	}
	preview := p.Preview(false)
	changes := p.Changes("applied", nil)
	for _, l := range ownDeny {
		if !strings.Contains(preview, "  add     deny        "+l.Rule+"\n          "+l.Why+"\n") {
			t.Errorf("the preview lacks %s with its sentence:\n%s", l.Rule, preview)
		}
		found := false
		for _, c := range changes.Settings {
			if c.Line == l.Rule {
				found = c.Change == "add" && c.Kind == "deny" && c.Why == l.Why && !c.RunsCode
			}
		}
		if !found {
			t.Errorf("--json does not add %s with its sentence: %+v", l.Rule, changes.Settings)
		}
	}
	if r, _ := checkLocal(t, root, e.home); len(r.Findings) != 0 || len(r.Warnings) != 0 {
		t.Errorf("check after the link: %+v %+v", r.Findings, r.Warnings)
	}
}

// linkedBefore535 rewrites a project this build linked, its settings file ownSettings at first, as a build before
// step 5.3.5 left it: no deny rule of Bonsai's own, never_edit's rules as Edit(<path>), and the lock's fingerprint of
// those lines.
func linkedBefore535(t *testing.T, root string) {
	t.Helper()
	sd, err := readSettings(root)
	if err != nil {
		t.Fatal(err)
	}
	lock, err := workspace.LoadLock(root)
	if err != nil {
		t.Fatal(err)
	}
	own := map[string]bool{}
	for _, l := range ownDeny {
		own[l.Rule] = true
	}
	cfg, err := workspace.LoadConfig(root)
	if err != nil {
		t.Fatal(err)
	}
	old := map[string]string{}
	for _, p := range cfg.NeverEdit {
		old[neverEditRule(p)] = "Edit(" + p + ")"
	}
	// Every line of the file is Bonsai's but ownSettings' own deny rule and hook line.
	disk := diskLines(sd.root)
	var ref []Line
	for _, l := range disk {
		if (l.Kind == "hook" && l.Command == "npm run lint-staged") || (l.Kind == "deny" && l.Rule == "Read(secrets/**)") {
			continue
		}
		if l.Kind == "marketplace" {
			l.Value = bonsaiOrdered(l.Value)
		}
		ref = append(ref, l)
	}
	claimed, same := claim(disk, ref, false, lock.Files[SettingsFile].SHA256)
	if !same {
		t.Fatal("the settings file does not hold the lines the lock records")
	}
	var lines []Line
	for _, l := range claimed {
		if l.Kind == "deny" && own[l.Rule] {
			continue
		}
		if o, ok := old[l.Rule]; ok {
			l.Rule = o
		}
		lines = append(lines, l)
	}
	perms, _ := getObject(sd.root, "permissions")
	list, _ := perms.Get("deny")
	var kept []any
	for _, r := range list.([]any) {
		s := r.(string)
		if own[s] {
			continue
		}
		if o, ok := old[s]; ok {
			s = o
		}
		kept = append(kept, s)
	}
	doc := putObject(sd.root, "permissions", set(perms, "deny", kept), false)
	b, err := schema.EncodeUTF8(doc)
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, root, SettingsFile, string(b))
	raw := read(t, root, workspace.LockFile)
	writeFile(t, root, workspace.LockFile, strings.Replace(raw, lock.Files[SettingsFile].SHA256, linesHash(lines), 1))
}

// A project linked by a build before step 5.3.5 is as its lock says (check: no finding, no warning); its next update
// adds Bonsai's own deny rules and gives never_edit's rule its new form, each named with its sentence, on --yes alone
// (no --allow-exec: a deny rule runs no code); every other key, value and rule of the file stays where it was; a
// second update has nothing to change; unlink takes every one of them out.
func TestOwnDenyOfAProjectLinkedBefore(t *testing.T) {
	e := setup(t)
	root := testpack.Project(t, e.tmp, "before535")
	writeFile(t, root, SettingsFile, ownSettings)
	e.link(t, root, e.pack.A, "ledger.json")
	linkedBefore535(t, root)
	before := read(t, root, SettingsFile)
	if strings.Contains(before, ".bonsai/") || !strings.Contains(before, `"Edit(ledger.json)"`) {
		t.Fatalf("not as a build before 5.3.5 left it:\n%s", before)
	}
	if r, _ := checkLocal(t, root, e.home); len(r.Findings) != 0 || len(r.Warnings) != 0 {
		t.Errorf("check of the project linked before: %+v %+v", r.Findings, r.Warnings)
	}

	p := e.plan(t, root, Request{})
	var got []string
	for _, c := range p.Settings {
		got = append(got, c.Change+" "+c.Kind+" "+c.Line+" (was "+c.Was+"): "+c.Why)
	}
	var want []string
	for _, l := range ownDeny {
		want = append(want, "add deny "+l.Rule+" (was ): "+l.Why)
	}
	want = append(want, "change deny Edit(/ledger.json) (was Edit(ledger.json)): Agents can never edit ledger.json: it is in never_edit in bonsai.yaml.")
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Errorf("settings lines\n%s\nwant\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
	if p.NeedsExec() {
		t.Errorf("the deny rules run code: %s", items(p))
	}
	e.apply(t, root, Request{})
	after := read(t, root, SettingsFile)

	// Only the deny list changed: the file without it is the same, byte for byte; the project's own rule and the
	// pack's keep their places, and Bonsai's are added after them.
	strip := func(text string) string {
		v, err := schema.Decode([]byte(text))
		if err != nil {
			t.Fatal(err)
		}
		doc := v.(schema.Object)
		perms, _ := getObject(doc, "permissions")
		b, _ := schema.EncodeUTF8(putObject(doc, "permissions", without(perms, "deny"), false))
		return string(b)
	}
	if strip(before) != strip(after) {
		t.Errorf("the update changed more than the deny list:\n%s\n%s", before, after)
	}
	deny, _ := getList(settingsDocOf(t, root), "permissions", "deny")
	wantDeny := `["Read(secrets/**)","Edit(demo/never.txt)","` + strings.Join(ruleList(ownDeny), `","`) + `","Edit(/ledger.json)"]`
	if got := schema.Show(deny); got != wantDeny {
		t.Errorf("deny list\n  %s\nwant\n  %s", got, wantDeny)
	}
	if r, _ := checkLocal(t, root, e.home); len(r.Findings) != 0 || len(r.Warnings) != 0 {
		t.Errorf("check after the update: %+v %+v", r.Findings, r.Warnings)
	}
	if p := e.plan(t, root, Request{}); !p.Nothing() || len(p.Settings) != 0 {
		t.Errorf("a second update:\n%s", p.Preview(false))
	}
	u := unlinkPlan(t, e, root)
	if err := Apply(u); err != nil {
		t.Fatal(err)
	}
	if got := read(t, root, SettingsFile); strings.Contains(got, "Edit(") || !strings.Contains(got, "Read(secrets/**)") {
		t.Errorf("after unlink:\n%s", got)
	}
}

func ruleList(ls []Line) []string {
	var out []string
	for _, l := range ls {
		out = append(out, l.Rule)
	}
	return out
}

// disableAllHooks: false is written at a link and kept by an update that changes other lines; check finds true in the
// project's settings file (a changed line of Bonsai's too, which --adopt takes back) and in the local settings file,
// which no update touches.
func TestDisableAllHooksHeld(t *testing.T) {
	e := setup(t)
	root := testpack.Project(t, e.tmp, "hooks")
	e.link(t, root, e.pack.A, "ledger.json")
	off := func() any { v, _ := settingsDocOf(t, root).Get("disableAllHooks"); return v }
	if off() != false {
		t.Fatalf("disableAllHooks is %v", off())
	}
	cfg := read(t, root, "bonsai.yaml")
	writeFile(t, root, "bonsai.yaml", strings.Replace(cfg, `never_edit: ["ledger.json"]`, `never_edit: ["ledger.json", "other.json"]`, 1))
	if p := e.apply(t, root, Request{}); len(p.Settings) != 1 || off() != false {
		t.Errorf("an update of another line: %d lines, disableAllHooks %v", len(p.Settings), off())
	}

	hooksOff := func(file string) *Finding {
		r, _ := checkLocal(t, root, e.home)
		for i, f := range r.Findings {
			if f.Code == "hooks-off" && f.File == file {
				return &r.Findings[i]
			}
		}
		return nil
	}
	writeFile(t, root, SettingsFile, strings.Replace(read(t, root, SettingsFile), `"disableAllHooks": false`, `"disableAllHooks": true`, 1))
	if f := hooksOff(SettingsFile); f == nil || f.Who != "person" || !strings.Contains(f.Next, "--adopt") {
		t.Errorf("check with disableAllHooks true in %s: %+v", SettingsFile, f)
	}
	e.apply(t, root, Request{Adopt: []string{SettingsFile}})
	if off() != false || hooksOff(SettingsFile) != nil {
		t.Errorf("--adopt left disableAllHooks %v", off())
	}
	writeFile(t, root, LocalSettingsFile, `{"disableAllHooks": true}`+"\n")
	if f := hooksOff(LocalSettingsFile); f == nil || f.Who != "person" {
		t.Errorf("check with disableAllHooks true in %s: %+v", LocalSettingsFile, f)
	}
	if p := e.plan(t, root, Request{}); !p.Nothing() || read(t, root, LocalSettingsFile) != `{"disableAllHooks": true}`+"\n" {
		t.Errorf("an update with the local file's true:\n%s", p.Preview(false))
	}
}
