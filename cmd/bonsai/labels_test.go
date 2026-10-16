package main

// bonsai labels attach and detach (plan-5 5.6.2; contract sections 5.1-5.3): the file checked, its namespace never a
// pack's or bonsai, a re-attach additive only, the refusal in an agent's session, and what an attach then reaches
// (status --json's labels, check's values). Every test sets or clears CLAUDE_CODE_CHILD_SESSION itself (newCLI
// clears it; asAgent sets it), so the suite behaves the same inside an agent's session and in CI.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/testpack"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// studioLabel is one label definition of the studio's namespace.
func studioLabel(name, description string) string {
	return "  - name: studio." + name + "\n    kind: choice\n    values: [small, large]\n    items: null\n    pattern: null\n" +
		"    max: null\n    kinds: [task]\n    set_by: agent\n    grants: false\n    description: \"" + description + "\"\n"
}

// studioLabels is a bonsai.labels/1 file in the namespace ns.
func studioLabels(ns string, version string, labels ...string) string {
	return "format: bonsai.labels/1\nnamespace: " + ns + "\nversion: " + version + "\nlabels:\n" + strings.Join(labels, "")
}

// attachDoc runs labels attach or detach with --json and holds the output to bonsai.attach/1.
func (c *cli) attachDoc(exit int, args ...string) schema.Object {
	c.t.Helper()
	code, out, errOut := c.run("", append(args, "--json")...)
	if code != exit {
		c.t.Fatalf("%v: exit %d, want %d\n%s%s", args, code, exit, out, errOut)
	}
	return fits(c.t, out, "attach")
}

// attachedFile is the machine folder's file for a namespace.
func (c *cli) attachedFile(ns string) string {
	c.t.Helper()
	p, err := workspace.LabelsFile(os.Getenv("BONSAI_HOME"), c.root, ns)
	if err != nil {
		c.t.Fatal(err)
	}
	return p
}

// source writes a file outside the project and gives its place.
func (c *cli) source(name, content string) string {
	c.t.Helper()
	p := filepath.Join(c.tmp, name)
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		c.t.Fatal(err)
	}
	return p
}

func resultOf(d schema.Object) string {
	r, _ := d.Get("result")
	s, _ := r.(string)
	return s
}

func codeOf(d schema.Object) string { return errorIn(d, "attach").String("code") }

func TestLabelsAttachAndDetach(t *testing.T) {
	c := newCLI(t)
	c.declaring()
	c.commit("link")
	v1 := studioLabels("studio", "1", studioLabel("size", "How big."), studioLabel("team", "Who owns it."))
	v1file := c.source("v1.yaml", strings.ReplaceAll(v1, "\n", "\r\n")) // a Windows checkout's file: CRLF

	// A first attach: attached, the copy has LF endings, and status lists the namespace from the machine.
	d := c.attachDoc(0, "labels", "attach", v1file)
	if resultOf(d) != "attached" || d.String("namespace") != "studio" || d.String("command") != "attach" {
		t.Errorf("first attach: %s", schema.Show(d))
	}
	if v, _ := d.Get("version"); schema.Show(v) != "1" {
		t.Errorf("version %v", v)
	}
	if l, _ := d.Get("labels"); schema.Show(l) != `["studio.size", "studio.team"]` && schema.Show(l) != `["studio.size","studio.team"]` {
		t.Errorf("labels %s", schema.Show(l))
	}
	file := c.attachedFile("studio")
	if f, _ := d.Get("file"); f != filepath.ToSlash(file) {
		t.Errorf("file %v, want %s", f, filepath.ToSlash(file))
	}
	if raw, err := os.ReadFile(file); err != nil || string(raw) != v1 {
		t.Errorf("the copy: %v\n%q", err, raw)
	}
	_, out, _ := c.run("", "status", "--json")
	if !strings.Contains(flat(out), `"namespace": "studio"`) || !strings.Contains(flat(out), `"from": "machine"`) {
		t.Errorf("status --json does not list the namespace from the machine:\n%s", out)
	}

	// What check then holds a task to: a value its definition does not take is a finding; one it takes is not.
	c.task("T-0901", "todo", "light", "\n  studio.size: huge")
	_, out, _ = c.run("", "check", "--json")
	if codes := codesIn(fits(t, out, "check"), "findings"); !strings.Contains(codes, "label") {
		t.Errorf("check does not hold studio.size to its definition: %s", codes)
	}
	c.task("T-0901", "todo", "light", "\n  studio.size: small")
	_, out, _ = c.run("", "check", "--json")
	if strings.Contains(codesIn(fits(t, out, "check"), "findings"), "label") {
		t.Errorf("check refuses a value the definition takes:\n%s", out)
	}

	// The same file again: unchanged, nothing written (the copy keeps its bytes and its time).
	before, _ := os.Stat(file)
	if d := c.attachDoc(0, "labels", "attach", v1file); resultOf(d) != "unchanged" {
		t.Errorf("same again: %s", schema.Show(d))
	}
	if after, _ := os.Stat(file); !after.ModTime().Equal(before.ModTime()) {
		t.Errorf("an unchanged attach wrote the file")
	}

	// A newer version that only adds: updated.
	v2 := studioLabels("studio", "2", studioLabel("size", "How big."), studioLabel("team", "Who owns it."), studioLabel("area", "Where."))
	if d := c.attachDoc(0, "labels", "attach", c.source("v2.yaml", v2)); resultOf(d) != "updated" {
		t.Errorf("v2: %s", schema.Show(d))
	}
	if raw, _ := os.ReadFile(file); string(raw) != v2 {
		t.Errorf("the copy is not v2")
	}

	// Anything else is refused label-redefined (exit 4, the person's step named), and the copy stays as it was.
	for name, content := range map[string]string{
		"a definition changed":     studioLabels("studio", "3", studioLabel("size", "Another sentence."), studioLabel("team", "Who owns it."), studioLabel("area", "Where.")),
		"a definition removed":     studioLabels("studio", "3", studioLabel("size", "How big."), studioLabel("area", "Where.")),
		"other content, same":      studioLabels("studio", "2", studioLabel("size", "How big."), studioLabel("team", "Who owns it."), studioLabel("area", "Where."), studioLabel("more", "More.")),
		"an older version":         studioLabels("studio", "1", studioLabel("size", "How big."), studioLabel("team", "Who owns it.")),
		"a comment only, same one": "# a note\n" + v2,
	} {
		code, out, errOut := c.run("", "labels", "attach", c.source("x.yaml", content), "--json")
		d := fits(t, out, "attach")
		if code != 4 || codeOf(d) != "label-redefined" || whoOf(errorIn(d, "attach")) != "person" ||
			!strings.Contains(flat(out), "labels detach studio") || resultOf(d) != "" {
			t.Errorf("%s: exit %d\n%s%s", name, code, out, errOut)
		}
		if raw, _ := os.ReadFile(file); string(raw) != v2 {
			t.Errorf("%s: the copy changed", name)
		}
	}
	// Plain text: the refusal block names its word and its next step.
	if code, _, errOut := c.run("", "labels", "attach", c.source("x.yaml", v1)); code != 4 || !strings.Contains(errOut, "[label-redefined]") || !hasNext(errOut) {
		t.Errorf("text refusal: %d\n%s", code, errOut)
	}

	// Detach: the file goes; again, nothing; the old version can then be attached.
	if d := c.attachDoc(0, "labels", "detach", "studio"); resultOf(d) != "detached" || d.String("command") != "detach" || d.String("namespace") != "studio" {
		t.Errorf("detach: %s", schema.Show(d))
	} else if v, _ := d.Get("version"); schema.Show(v) != "2" {
		t.Errorf("detach's version %v", v)
	}
	if _, err := os.Stat(file); err == nil {
		t.Errorf("detach left the file")
	}
	if d := c.attachDoc(0, "labels", "detach", "studio"); resultOf(d) != "nothing" {
		t.Errorf("detach again: %s", schema.Show(d))
	}
	if d := c.attachDoc(0, "labels", "attach", c.source("x.yaml", v1)); resultOf(d) != "attached" {
		t.Errorf("attach after detach: %s", schema.Show(d))
	}

	// Human text, style C.
	code, out, errOut := c.run("", "labels", "detach", "studio")
	if code != 0 || errOut != "" || !strings.HasPrefix(out, "bonsai labels detach  detached studio\n") || !strings.Contains(out, filepath.ToSlash(file)) {
		t.Errorf("detach text: %d\n%s%s", code, out, errOut)
	}
	code, out, _ = c.run("", "labels", "attach", v1file)
	if code != 0 || !strings.HasPrefix(out, "bonsai labels attach  attached studio, version 1\n") || !strings.Contains(out, "2 labels") ||
		!strings.Contains(out, "bonsai status --json lists them") {
		t.Errorf("attach text: %d\n%s", code, out)
	}
}

func TestLabelsAttachRefusals(t *testing.T) {
	c := newCLI(t)
	good := c.source("good.yaml", studioLabels("studio", "1", studioLabel("size", "How big.")))

	// Not linked: attach is refused, nothing written; detach needs only a git checkout.
	if d := c.attachDoc(4, "labels", "attach", good); codeOf(d) != "not-linked" {
		t.Errorf("not linked: %s", schema.Show(d))
	}
	if d := c.attachDoc(0, "labels", "detach", "studio"); resultOf(d) != "nothing" {
		t.Errorf("detach, not linked: %s", schema.Show(d))
	}
	c.declaring()
	c.commit("link")

	// The namespace: bonsai, and a locked pack's.
	src, shas := testpack.Fixture(t, c.tmp, "outside-pack", map[string]string{
		".claude-plugin/plugin.json": `{"name": "outside-pack", "description": "A fixture pack."}` + "\n",
		"bonsai/pack.yaml":           "format: bonsai.pack/1\nid: outside-pack\nversion: \"0.1.0\"\nfiles: []\nhooks: []\ndeny: []\n",
		"bonsai/labels.yaml":         studioLabels("outside-pack", "1", strings.ReplaceAll(studioLabel("size", "Its own."), "studio.", "outside-pack.")),
	})
	c.addPack("outside-pack", src, shas[0])
	c.commit("a second pack")
	for _, r := range []struct{ ns, in string }{{"bonsai", "Bonsai's own"}, {"outside-pack", "the locked pack"}} {
		file := c.source("taken.yaml", studioLabels(r.ns, "1", strings.ReplaceAll(studioLabel("other", "Mine."), "studio.", r.ns+".")))
		d := c.attachDoc(4, "labels", "attach", file)
		if codeOf(d) != "namespace-taken" || whoOf(errorIn(d, "attach")) != "person" || d.String("namespace") != r.ns ||
			!strings.Contains(flat(schema.Show(d)), r.in) || resultOf(d) != "" {
			t.Errorf("%s: %s", r.ns, schema.Show(d))
		}
		if _, err := os.Stat(c.attachedFile(r.ns)); err == nil {
			t.Errorf("%s: a file was written", r.ns)
		}
	}

	// The file: a field that does not read is named (bad-value, exit 2); a file that is not there is read-failed.
	bad := strings.Replace(studioLabels("studio", "1", studioLabel("size", "How big.")), "kind: choice", "kind: sparkly", 1)
	d := c.attachDoc(2, "labels", "attach", c.source("bad.yaml", bad))
	if codeOf(d) != "bad-value" || !strings.Contains(flat(schema.Show(d)), "kind") {
		t.Errorf("bad kind: %s", schema.Show(d))
	}
	if d := c.attachDoc(2, "labels", "attach", c.source("bad2.yaml", "not: labels\n")); codeOf(d) != "bad-value" {
		t.Errorf("not labels: %s", schema.Show(d))
	}
	if d := c.attachDoc(3, "labels", "attach", filepath.Join(c.tmp, "nowhere.yaml")); codeOf(d) != "read-failed" {
		t.Errorf("missing file: %s", schema.Show(d))
	}
	if d := c.attachDoc(2, "labels", "attach"); codeOf(d) != "missing-value" {
		t.Errorf("no file: %s", schema.Show(d))
	}
	if d := c.attachDoc(2, "labels", "detach"); codeOf(d) != "missing-value" {
		t.Errorf("no namespace: %s", schema.Show(d))
	}
	if d := c.attachDoc(2, "labels", "detach", "Not A Namespace"); codeOf(d) != "bad-value" {
		t.Errorf("bad namespace: %s", schema.Show(d))
	}
	if code, _, _ := c.run("", "labels", "attach", good, good); code != 2 {
		t.Errorf("two files: exit %d", code)
	}
	if _, err := os.Stat(c.attachedFile("studio")); err == nil {
		t.Errorf("a refusal wrote the file")
	}

	// A copy that does not read blocks an attach (it cannot be compared), and detach clears it.
	if err := os.MkdirAll(filepath.Dir(c.attachedFile("studio")), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(c.attachedFile("studio"), []byte("junk: [\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if d := c.attachDoc(4, "labels", "attach", good); codeOf(d) != "label-redefined" {
		t.Errorf("over a broken copy: %s", schema.Show(d))
	}
	if d := c.attachDoc(0, "labels", "detach", "studio"); resultOf(d) != "detached" {
		t.Errorf("detach of a broken copy: %s", schema.Show(d))
	}
	if d := c.attachDoc(0, "labels", "attach", good); resultOf(d) != "attached" {
		t.Errorf("attach after: %s", schema.Show(d))
	}

	// A project taken out (no bonsai.yaml) can still be unregistered; outside git, detach is not-a-checkout.
	if err := os.Remove(filepath.Join(c.root, "bonsai.yaml")); err != nil {
		t.Fatal(err)
	}
	if d := c.attachDoc(0, "labels", "detach", "studio"); resultOf(d) != "detached" {
		t.Errorf("detach, bonsai.yaml gone: %s", schema.Show(d))
	}
	t.Chdir(t.TempDir())
	if d := c.attachDoc(4, "labels", "detach", "studio"); codeOf(d) != "not-a-checkout" {
		t.Errorf("outside git: %s", schema.Show(d))
	}
}

// In an agent's session both refuse themselves before anything is read or written, a --json given or not.
func TestLabelsRefuseInAnAgentsSession(t *testing.T) {
	c := newCLI(t)
	c.declaring()
	c.commit("link")
	good := c.source("good.yaml", studioLabels("studio", "1", studioLabel("size", "How big.")))
	if d := c.attachDoc(0, "labels", "attach", good); resultOf(d) != "attached" {
		t.Fatalf("attach: %s", schema.Show(d))
	}
	file := c.attachedFile("studio")
	asAgent(c)
	for _, args := range [][]string{{"labels", "attach", good}, {"labels", "detach", "studio"}, {"labels", "attach"}, {"labels", "detach"}} {
		code, out, errOut := c.run("", args...)
		if code != 4 || !strings.Contains(errOut, "[person-command]") || !hasNext(errOut) || out != "" {
			t.Errorf("%v: %d\n%s%s", args, code, out, errOut)
		}
		d := c.attachDoc(4, args...)
		if codeOf(d) != "person-command" || whoOf(errorIn(d, "attach")) != "person" {
			t.Errorf("%v --json: %s", args, schema.Show(d))
		}
	}
	if _, err := os.Stat(file); err != nil {
		t.Errorf("a refused detach removed the file: %v", err)
	}
	more := c.source("more.yaml", studioLabels("other", "1", strings.ReplaceAll(studioLabel("size", "How big."), "studio.", "other.")))
	_, _, _ = c.run("", "labels", "attach", more)
	if _, err := os.Stat(c.attachedFile("other")); err == nil {
		t.Errorf("a refused attach wrote a file")
	}
	asPerson(c)
	if d := c.attachDoc(0, "labels", "detach", "studio"); resultOf(d) != "detached" {
		t.Errorf("as a person: %s", schema.Show(d))
	}
}
