package main

// bonsai log append (step 5.2.5, contract section 8.4): one event record in today's day file, session and agent
// null; each label checked against the definitions in force, a locked pack's and those attached on this machine; a
// label no definition has, or a value of the wrong kind, exits 2 (label-not-defined); --target and --text redacted
// and capped; --json against bonsai.logs/1. Every secret here is made up for the test.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/LastStep/Bonsai/internal/asks"
	"github.com/LastStep/Bonsai/internal/record"
	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// opsLabels are label definitions attached on this machine, a made-up outside program's.
const opsLabels = `format: bonsai.labels/1
namespace: ops
version: 1
labels:
  - name: ops.event
    kind: choice
    values: [deploy, applied]
    items: null
    pattern: null
    max: null
    kinds: [task]
    set_by: outside
    grants: false
    description: "What the outside program did."
  - name: ops.ref
    kind: text
    values: []
    items: null
    pattern: "^[0-9a-f]{7,40}$"
    max: 60
    kinds: [task]
    set_by: outside
    grants: false
    description: "A commit."
  - name: ops.note
    kind: text
    values: []
    items: null
    pattern: null
    max: 40
    kinds: [task]
    set_by: outside
    grants: false
    description: "A short note."
  - name: ops.cost
    kind: number
    values: []
    items: null
    pattern: null
    max: null
    kinds: [task]
    set_by: outside
    grants: false
    description: "What it cost."
  - name: ops.counts
    kind: list
    values: []
    items: number
    pattern: null
    max: 3
    kinds: [task]
    set_by: outside
    grants: false
    description: "Some counts."
`

func TestLogAppend(t *testing.T) {
	c := newCLI(t)
	t.Setenv(asks.SessionEnv, askSession) // an outside event has no session, even when run in one
	defer func(f func() time.Time) { logAppendNow = f }(logAppendNow)
	logAppendNow = func() time.Time { return time.Date(2026, 10, 8, 16, 30, 0, 0, time.UTC) }
	c.declaring() // the base fixture: bonsai.allows (a list of text) and bonsai.branch (text) from a locked pack
	c.commit("link")
	home := os.Getenv("BONSAI_HOME")
	dir, err := workspace.MachineDir(home, c.root)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, workspace.MachineLabelsDir), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, workspace.MachineLabelsDir, "ops.yaml"), []byte(opsLabels), 0o644); err != nil {
		t.Fatal(err)
	}
	before := c.localFiles()

	// Refused: exit 2, its word, nothing written.
	for _, r := range []struct {
		args []string
		word string
		in   string
	}{
		{nil, "missing-value", "log append needs at least one --label"},
		{[]string{"--label", "ops.event"}, "bad-value", `--label "ops.event" is not name=value`},
		{[]string{"--label", "ops.event=deploy", "--label", "ops.event=applied"}, "bad-value", "is given twice"},
		{[]string{"--label", "ops.nothing=1"}, "label-not-defined", `no label definition in force has "ops.nothing" (in force: bonsai.allows, bonsai.branch, ops.event`},
		{[]string{"--label", "Ops.Event=deploy"}, "label-not-defined", "no label definition in force has"},
		{[]string{"--label", "ops.event=Deploy"}, "label-not-defined", `value "Deploy" is not one of deploy, applied`},
		{[]string{"--label", "ops.ref=not-a-commit"}, "label-not-defined", "not matching ^[0-9a-f]{7,40}$"},
		{[]string{"--label", "ops.note=" + strings.Repeat("n", 41)}, "label-not-defined", "longer than 40 characters"},
		{[]string{"--label", "ops.event=password=MadeUpLogSecret0"}, "label-not-defined", `value "password=[redacted]" is not one of`},
		{[]string{"--label", "ops.cost=1e3"}, "label-not-defined", "not a number"},
		{[]string{"--label", "ops.cost=007"}, "label-not-defined", "not a number"},
		{[]string{"--label", "ops.counts=1,two"}, "label-not-defined", `item "two" is not a number`},
		{[]string{"--label", "ops.counts=1,2,3,4"}, "label-not-defined", "a list of 4 items, more than 3"},
		{[]string{"--label", "bonsai.allows=a,,b"}, "label-not-defined", "an empty item"},
		{[]string{"--label", "bonsai.branch=main\nnext"}, "label-not-defined", "not one line"},
		{[]string{"--label", "ops.event=deploy", "now"}, "bad-flag", `log append takes no "now"`},
	} {
		args := append([]string{"log", "append"}, r.args...)
		code, _, errOut := c.run("", args...)
		if code != 2 || !strings.Contains(errOut, r.in) || !strings.Contains(errOut, "\nnext: ") {
			t.Errorf("%v: %d %q, want 2 and %q", r.args, code, errOut, r.in)
		}
		wantRefusal(t, logAppendWord, args, 2, r.word)
	}
	if !sameFiles(before, c.localFiles()) {
		t.Errorf("a refusal wrote something")
	}

	// Written: every kind, a pack's label and the machine's; the target and text redacted.
	code, out, errOut := c.run("", "log", "append", "--label", "ops.event=deploy", "--label", "ops.ref=c6ba392", "--label", "ops.cost=-12.50",
		"--label", "ops.counts=1,2", "--label", "bonsai.allows=docs/**,assets/*.png", "--label", "ops.note=hello token=MadeUpLog0",
		"--target", "T-0901 password=MadeUpLogSecret1", "--text", "deployed with api_key=MadeUpLogSecret2 "+strings.Repeat("x", 400))
	if code != 0 || out != "written: .bonsai/local/log/w-2026-10-08.ndjson (an event record, 6 labels)\n" {
		t.Fatalf("log append: %d %q %q", code, out, errOut)
	}
	code, out, _ = c.run("", "log", "append", "--label", "ops.event=applied", "--json")
	doc := fits(t, out, "logs")
	written, _ := doc.Get("written")
	if f, _ := doc.Get("files"); code != 0 || f != nil || written == nil {
		t.Fatalf("--json: %d\n%s", code, out)
	}
	lf, err := record.ReadLog(filepath.Join(c.root, ".bonsai", "local", "log", "w-2026-10-08.ndjson"))
	if err != nil || len(lf.Records) != 2 || lf.Skipped != 0 {
		t.Fatalf("the day file: %v %+v", err, lf)
	}
	r := lf.Records[0]
	if r.Event != "event" || r.Session != nil || r.Agent != nil || r.At != "2026-10-08T16:30:00.000Z" ||
		text0(r.Target) != "T-0901 password=[redacted]" || !strings.HasPrefix(text0(r.Text), "deployed with api_key=[redacted] xxx") ||
		len([]rune(text0(r.Text))) != 300 {
		t.Errorf("the record: %+v", r)
	}
	if got := schema.Show(r.Labels); got != `{"ops.event":"deploy","ops.ref":"c6ba392","ops.cost":-12.50,"ops.counts":[1,2],`+
		`"bonsai.allows":["docs/**","assets/*.png"],"ops.note":"hello token=[redacted]"}` {
		t.Errorf("the labels: %s", got)
	}
	if lf.Records[1].ID != written.(schema.Object).String("id") {
		t.Errorf("--json's written is not the record written: %s", schema.Show(written))
	}
	for f, b := range c.localFiles() {
		if strings.Contains(b, "MadeUpLog") {
			t.Errorf("a made-up secret is in %s", f)
		}
		if before[f] != b && f != ".bonsai/local/log/w-2026-10-08.ndjson" {
			t.Errorf("%s changed", f)
		}
	}
}

// With no definition in force at all, the refusal says so; and a project not linked is refused with exit 4.
func TestLogAppendNothingInForce(t *testing.T) {
	c := newCLI(t)
	code, _, errOut := c.run("", "log", "append", "--label", "ops.event=deploy")
	if code != 4 || !strings.Contains(errOut, "next: run bonsai init") {
		t.Errorf("not linked: %d %q", code, errOut)
	}
	c.link()
	code, _, errOut = c.run("", "log", "append", "--label", "ops.event=deploy")
	if code != 2 || !strings.Contains(errOut, `no label definition in force has "ops.event" (none is in force)`) {
		t.Errorf("none in force: %d %q", code, errOut)
	}
	code, _, errOut = c.run("", "log")
	if code != 2 || !strings.Contains(errOut, "bonsai log needs a log action") {
		t.Errorf("no action: %d %q", code, errOut)
	}
}
