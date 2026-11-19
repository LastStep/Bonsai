package engine

// The bug files and the bugs page (set 14; plan P-T-6023, steps 3 and 4): the marker rule (an unmoved, hand-kept
// register comes out of check --write byte for byte the same and is never a finding; a marked page is rebuilt), check's
// rules on the bug files, and a pack that still declares the old bugs kind (the workflow pack's v1.0.2) linking and
// checking with this Bonsai.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/testpack"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// bugFile is a bug file's text: its id, status and closed day ("" for null), and was ("" for null).
func bugFile(id, status, closed, was string) string {
	null := func(s string) string {
		if s == "" {
			return ""
		}
		return " " + s
	}
	return "---\nformat: bonsai.bug/1   # fields: bonsai check --schema bonsai.bug\nid: " + id + "\ntitle: " + id +
		" goes wrong | sometimes\nstatus: " + status + "\narea: the guard\nwhere: internal/guard/x.go:1\nfound_by: verifier@T-0001\n" +
		"filed_by: builder@T-0001\nfixed_by:\nfixed_in:\ntask:\nwas:" + null(was) + "\nwhy:\ncreated: 2026-11-01\nclosed:" + null(closed) +
		"\nlabels: {}\n---\n\nHow to repeat it.\n"
}

// A register as the workflow pack's template wrote it, by hand, with no generated marker.
const handRegister = "---\ntitle: Known bugs   # the bugs register; every field and how to fill it: skill workflow:bugs\n" +
	"updated: 2026-11-06\n---\n\n# Known bugs\n\n## Open\n\n| Id | What goes wrong | Where | Found by |\n|---|---|---|---|\n" +
	"| **BUG-008** | A thing | a.go:1 | STATE |\n"

func codes(fs []Finding) string {
	var out []string
	for _, f := range fs {
		out = append(out, f.Code+" "+f.File)
	}
	return strings.Join(out, "; ")
}

func TestUnmovedRegisterIsLeftByteForByte(t *testing.T) {
	e := setup(t)
	root := e.linked(t, "unmoved")
	page := "work/bugs.md"
	if cfg, err := workspace.LoadConfigFull(root); err != nil || cfg.Full.Documents.BugsFile() != page || cfg.Full.Documents.BugFolder() != "work/bugs" {
		t.Fatalf("init's bug places: %v", err)
	}
	// With neither a page nor a bug file, check --write writes no page.
	if res, err := WriteTables(root); err != nil || len(res.Left) != 0 || strings.Contains(strings.Join(res.Written, " "), page) {
		t.Fatalf("no page, no bug files: %v %+v", err, res)
	}
	if _, err := os.Stat(filepath.Join(root, page)); !os.IsNotExist(err) {
		t.Fatalf("a page was written with no bug files: %v", err)
	}
	// A hand-kept register, in CRLF too: left byte for byte, a warning naming base's bug skill, never a finding, even
	// with bug files beside it.
	writeFile(t, root, "work/bugs/BUG-145.md", bugFile("BUG-145", "open", "", ""))
	for _, reg := range []string{handRegister, strings.ReplaceAll(handRegister, "\n", "\r\n")} {
		writeFile(t, root, page, reg)
		res, werr := WriteTables(root)
		if werr != nil || strings.Join(res.Left, " ") != page || strings.Contains(strings.Join(res.Written, " "), page) {
			t.Fatalf("write: %v %+v", werr, res)
		}
		if got := read(t, root, page); got != reg {
			t.Fatalf("the register changed:\n%q\nwant\n%q", got, reg)
		}
		r, err := checkLocal(t, root, e.home)
		if err != nil || len(r.Findings) != 0 || codes(r.Warnings) != "bugs-not-moved "+page || !strings.Contains(r.Warnings[0].Next, "base:bug") {
			t.Fatalf("an unmoved register: %v findings %s warnings %s", err, codes(r.Findings), codes(r.Warnings))
		}
	}
	if !strings.Contains(strings.Join(mustWrite(t, root).Notes(), "\n"), page+" is left as it is") {
		t.Error("check --write's notes do not say the register was left")
	}
}

func mustWrite(t *testing.T, root string) *TablesResult {
	t.Helper()
	res, err := WriteTables(root)
	if err != nil {
		t.Fatal(err)
	}
	return res
}

func TestMarkedPageIsRebuilt(t *testing.T) {
	e := setup(t)
	root := e.linked(t, "marked")
	page := "work/bugs.md"
	writeFile(t, root, "work/bugs/BUG-145.md", bugFile("BUG-145", "open", "", ""))
	writeFile(t, root, "work/bugs/BUG-012-old.md", bugFile("BUG-012", "fixed", "", "TB-012"))
	writeFile(t, root, "work/bugs/BUG-146.md", bugFile("BUG-146", "wontfix", "2026-11-02", ""))
	writeFile(t, root, "work/bugs/BUG-147.md", bugFile("BUG-147", "watched", "", ""))
	writeFile(t, root, "work/bugs/BUG-099.md", bugFile("BUG-099", "open", "", ""))
	writeFile(t, root, "work/bugs/README.md", "Not a bug file: its name is no bug id.\n")
	// A missing page while there are bug files: the warning tables, and check --write writes it.
	r, err := checkLocal(t, root, e.home)
	if err != nil || len(r.Findings) != 0 || codes(r.Warnings) != "tables "+page {
		t.Fatalf("a missing page: %v findings %s warnings %s", err, codes(r.Findings), codes(r.Warnings))
	}
	res := mustWrite(t, root)
	if !strings.Contains(strings.Join(res.Written, " "), page) {
		t.Fatalf("the page was not written: %+v", res)
	}
	got := read(t, root, page)
	b, err := format.ReadBugs([]byte(got))
	if err != nil {
		t.Fatalf("the page does not read back: %v\n%s", err, got)
	}
	var order []string
	for _, row := range b.Bugs {
		order = append(order, row.ID+" "+row.Status)
	}
	if strings.Join(order, ", ") != "BUG-145 open, BUG-099 open, BUG-147 watched, BUG-012 fixed, BUG-146 wontfix" {
		t.Errorf("the page's rows: %v", order)
	}
	if !strings.Contains(got, "| BUG-145 | BUG-145 goes wrong \\| sometimes | the guard | internal/guard/x.go:1 | verifier@T-0001 | | | | |\n") {
		t.Errorf("a row:\n%s", got)
	}
	if r, err = checkLocal(t, root, e.home); err != nil || len(r.Findings) != 0 || len(r.Warnings) != 0 {
		t.Fatalf("after the write: %v findings %s warnings %s", err, codes(r.Findings), codes(r.Warnings))
	}
	// The same bytes again, whoever builds them: nothing rewritten.
	if res := mustWrite(t, root); !strings.Contains(strings.Join(res.Same, " "), page) {
		t.Errorf("a second write: %+v", res)
	}
	// A bug moves: the marked page is stale (a warning), then rebuilt; a hand edit of a marked page is undone.
	writeFile(t, root, "work/bugs/BUG-099.md", bugFile("BUG-099", "fixed", "2026-11-03", ""))
	if r, _ = checkLocal(t, root, e.home); len(r.Findings) != 0 || codes(r.Warnings) != "tables "+page {
		t.Fatalf("a stale page: findings %s warnings %s", codes(r.Findings), codes(r.Warnings))
	}
	writeFile(t, root, page, got+"\nA line added in an edit.\n")
	mustWrite(t, root)
	again := read(t, root, page)
	if strings.Contains(again, "added in an edit") || !strings.Contains(again, "| BUG-099 | BUG-099 goes wrong \\| sometimes | the guard | internal/guard/x.go:1 | verifier@T-0001 | | | 2026-11-03 | |") {
		t.Errorf("the marked page was not rebuilt:\n%s", again)
	}
	// A marked page with a newer major is left: this Bonsai does not write it, and the document finding names it.
	newer := strings.Replace(again, "format: bonsai.bugs/1", "format: bonsai.bugs/2", 1)
	writeFile(t, root, page, newer)
	if res := mustWrite(t, root); strings.Join(res.Left, " ") != page || read(t, root, page) != newer {
		t.Errorf("a newer page: %+v", res)
	}
	if r, _ = checkLocal(t, root, e.home); codes(r.Findings) != "document "+page {
		t.Errorf("a newer page's findings: %s", codes(r.Findings))
	}
}

func TestBugFileFindings(t *testing.T) {
	e := setup(t)
	root := e.linked(t, "bugfiles")
	writeFile(t, root, "work/bugs/BUG-145.md", bugFile("BUG-145", "open", "", ""))
	writeFile(t, root, "work/bugs/BUG-145-again.md", bugFile("BUG-145", "open", "", ""))
	writeFile(t, root, "work/bugs/BUG-150.md", bugFile("BUG-151", "open", "", ""))
	writeFile(t, root, "work/bugs/BUG-152.md", bugFile("BUG-152", "fixed", "", ""))
	writeFile(t, root, "work/bugs/BUG-153.md", bugFile("BUG-153", "wontfix", "", "TB-153"))
	writeFile(t, root, "work/bugs/BUG-154.md", strings.Replace(bugFile("BUG-154", "open", "", ""), "status: open", "status: closed", 1))
	writeFile(t, root, "work/bugs/BUG-155.md", bugFile("BUG-155", "open", "2099-01-01", ""))
	writeFile(t, root, "work/bugs/BUG-156.md", strings.Replace(bugFile("BUG-156", "open", "", ""), "where: internal/guard/x.go:1", `where: "/srv/example/a.go:1"`, 1))
	r, err := checkLocal(t, root, e.home)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		"document work/bugs/BUG-154.md",     // a status the format does not have
		"future-stamp work/bugs/BUG-155.md", // a closed day ahead of the UTC today
		"absolute-path work/bugs/BUG-156.md", // a field holding an absolute path (read from the one parse, as every kind's)
		"bug-id work/bugs/BUG-150.md",       // named for BUG-150, holding BUG-151
		"bug-closed work/bugs/BUG-152.md",   // fixed with no closed day and no was (BUG-153 has was: no finding)
		"bug-id work/bugs/BUG-145.md",       // two files hold BUG-145
	}
	if got := codes(r.Findings); got != strings.Join(want, "; ") {
		t.Errorf("findings:\n%s\nwant\n%s", strings.ReplaceAll(got, "; ", "\n"), strings.Join(want, "\n"))
	}
	for _, f := range r.Findings {
		if f.Next == "" || f.Who != "agent" {
			t.Errorf("%s: next %q who %q", f.Code, f.Next, f.Who)
		}
	}
}

// oldWorkflowPackYAML is a pack declaring the bugs kind exactly as the workflow pack's v1.0.2 does (one file, no id,
// no statuses), beside a folder kind, as DeclaringPack's pack.yaml otherwise is.
var oldWorkflowPackYAML = strings.Replace(testpack.BasePackYAML, "protected: [", oldBugsKind+"protected: [", 1)

// oldBugsKind is the workflow pack v1.0.2's bugs kind, its lines as that pack.yaml writes them.
const oldBugsKind = `  - kind: bugs                        # the project's register of known bugs, one file: skill workflow:bugs
    path: null                        # one file, not a folder
    file: work/bugs.md                # its default file
    id: null                          # the file holds no id of its own; each bug's id is a row in it
    statuses: []                      # no statuses
    person: []                        # no moves
    agent: []                         # no moves
    stamp: {}                         # no dates set by a move
    task_field: null                  # the register belongs to no one task
`

// A project on a pack that still declares bugs (the workflow pack's v1.0.2) links and checks with this Bonsai: the
// pack's kind is set aside, Bonsai's bugs is in force once, and check warns kind-set-aside, never a finding; its hand
// register stays as it is. check --pack of that pack warns too, with no finding.
func TestOldBugsKindStillLinks(t *testing.T) {
	e := setup(t)
	src, shas := testpack.Fixture(t, e.tmp, "oldflow", map[string]string{
		".claude-plugin/plugin.json":    "{\"name\": \"base\", \"description\": \"A fixture pack.\"}\n",
		"bonsai/pack.yaml":              oldWorkflowPackYAML,
		"bonsai/labels.yaml":            testpack.BaseLabelsYAML,
		"bonsai/lanes.yaml":             testpack.BaseLanesYAML,
		"bonsai/block.md":               "<!-- the fixture's block docs -->\nThe base fixture is linked.\n",
		"bonsai/files/session-start.md": "# Session start\n\nRead STATE first.\n",
	})
	root := testpack.Project(t, e.tmp, "oldflow-project")
	writeFile(t, root, "work/bugs.md", handRegister)
	e.apply(t, root, Request{Command: "init", Init: &InitValues{Name: "demo", Source: src, Ref: shas[0]}})
	cfg, err := workspace.LoadConfigFull(root)
	if err != nil {
		t.Fatal(err)
	}
	if got := schema.Show(cfg.Full.Documents.Extra); got != `{"bug":"work/bugs","bugs":"work/bugs.md","plan":"work/plans","notes":"work/notes.md"}` {
		t.Errorf("documents' further keys: %s", got)
	}
	lock, err := workspace.LoadLock(root)
	if err != nil {
		t.Fatal(err)
	}
	kinds, err := workspace.DocKinds(cfg.Full, lock)
	if err != nil {
		t.Fatal(err)
	}
	var bugs []string
	for _, k := range kinds {
		if k.Kind == "bugs" || k.Kind == "bug" {
			bugs = append(bugs, k.Kind+" "+k.From+" "+k.Path+k.File)
		}
	}
	if strings.Join(bugs, ", ") != "bug bonsai work/bugs, bugs bonsai work/bugs.md" {
		t.Errorf("the bug kinds in force: %v", bugs)
	}
	r, err := checkLocal(t, root, e.home)
	if err != nil || len(r.Findings) != 0 || codes(r.Warnings) != "kind-set-aside "+workspace.LockFile+"; bugs-not-moved work/bugs.md" {
		t.Fatalf("check: %v findings %s warnings %s", err, codes(r.Findings), codes(r.Warnings))
	}
	if !strings.Contains(r.Warnings[0].Next, "bonsai pack set base@") {
		t.Errorf("kind-set-aside's next step: %s", r.Warnings[0].Next)
	}
	if got := read(t, root, "work/bugs.md"); got != handRegister {
		t.Errorf("the link changed the register:\n%s", got)
	}
	// An update on the same pack: nothing refused, nothing changed.
	e.apply(t, root, Request{Command: "update"})
	if r, err = checkLocal(t, root, e.home); err != nil || len(r.Findings) != 0 {
		t.Fatalf("check after update: %v findings %s", err, codes(r.Findings))
	}
	// A documented pack declaring the same kind, under check --pack: the warning, no finding.
	dir := filepath.Join(t.TempDir(), "pack")
	files := testpack.DocumentedPack()
	if !strings.Contains(files["bonsai/pack.yaml"], "\nprotected: [") {
		t.Fatal("the documented pack's pack.yaml has no protected: line to put the kind before")
	}
	files["bonsai/pack.yaml"] = strings.Replace(files["bonsai/pack.yaml"], "\nprotected: [", "\n"+oldBugsKind+"protected: [", 1)
	testpack.WriteFolder(t, dir, files)
	pr, perr := CheckPack(dir)
	if perr != nil {
		t.Fatal(perr)
	}
	if len(pr.Findings) != 0 || codes(pr.Warnings) != "kind-set-aside bonsai/pack.yaml" {
		t.Errorf("check --pack: findings %s warnings %s", codes(pr.Findings), codes(pr.Warnings))
	}
}
