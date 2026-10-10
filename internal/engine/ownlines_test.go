package engine

// Bonsai's own hook lines across builds (step 5.2.4; plan-5 5.2.4 note 2): a project linked by a build that wrote the
// guard's line only gets the start and recorder lines at its next update only with --allow-exec and --yes, and check
// warns (own-hooks) until then; taking a pack out keeps Bonsai's own lines, which are not the pack's, and unlink takes
// every one of them out.

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
	var old, drop []Line
	for _, l := range claimed {
		if l.Kind == "hook" && isBonsaiHook(l.Command) && l.Command != GuardCommand {
			drop = append(drop, l)
			continue
		}
		old = append(old, l)
	}
	if len(drop) != len(ownHooks)-1 {
		t.Fatalf("%d own lines to take out, want %d", len(drop), len(ownHooks)-1)
	}
	b, err := schema.EncodeUTF8(applyLines(sd.root, drop, nil))
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
	if strings.Contains(before, StartCommand) || strings.Contains(before, RecordCommand) || !strings.Contains(before, GuardCommand) {
		t.Fatalf("not as a build before 5.2.4 left it:\n%s", before)
	}

	// check: a warning naming the eleven lines, its next step the consent; no finding but the two local sources'.
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
		!strings.Contains(warned.Message, "11 to add (SessionStart: bonsai hook start; UserPromptSubmit: bonsai hook record [async];") {
		t.Fatalf("the own-hooks warning: %+v", r.Warnings)
	}
	for _, f := range r.Findings {
		if f.Code == "changed" {
			t.Errorf("a changed finding: %+v", f)
		}
	}

	// update --yes is refused whole: each new line runs code (rule 1), named under Runs code; nothing is written.
	p := e.plan(t, root, Request{})
	if got := items(p); !p.NeedsExec() || len(p.RunsCode) != len(ownHooks)-1 || !strings.HasPrefix(got, "hook add SessionStart: bonsai hook start (bonsai); ") {
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
	u := unlinkPlan(t, e, root)
	removed := map[string]bool{}
	for _, c := range u.Settings {
		removed[c.Line] = c.Change == "remove"
	}
	for _, l := range ownHooks {
		if !removed[l.Text()] {
			t.Errorf("unlink leaves %s", l.Text())
		}
	}
	if err := Apply(u); err != nil {
		t.Fatal(err)
	}
	if s, err := readSettings(root); err != nil || (s.exists && strings.Contains(string(s.raw), "bonsai hook")) {
		t.Errorf("after unlink: %v", err)
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
