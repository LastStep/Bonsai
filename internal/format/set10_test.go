package format

// Set 10 (plan-5 5.6.0): the three outputs' formats, status's bonsai_release written null until step 5.6.5, and the
// two error words bonsai labels needs. The examples are read and written back by their writers in format_test.go
// (TestEveryExampleReads, TestWritersWriteTheExamples).

import (
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/schema"
)

func TestSetTenWordsAndFormats(t *testing.T) {
	for _, code := range []string{"namespace-taken", "label-redefined"} {
		if w, ok := ErrorWord(code); !ok || w.Means == "" || w.Who != "person" {
			t.Errorf("the error word %s is missing or incomplete: %+v", code, w)
		}
	}
	for _, name := range []string{"settings", "attach", "line"} {
		f, ok := Lookup("bonsai." + name + "/1")
		if !ok || !f.Writes || f.Major != 1 || f.Shape != JSONDoc {
			t.Errorf("%s: %+v", name, f)
		}
	}
	// A refusal's documents: every list [], every other field null but the closed words the command knows.
	out, err := (&Settings{Scope: "workspace"}).Encode()
	if err != nil || !strings.Contains(string(out), `"settings": []`) || !strings.Contains(string(out), `"changed": []`) ||
		!strings.Contains(string(out), `"file": null`) {
		t.Errorf("an empty settings document: %v\n%s", err, out)
	}
	out, err = (&Attach{Command: "detach"}).Encode()
	if err != nil || !strings.Contains(string(out), `"labels": []`) || !strings.Contains(string(out), `"result": null`) {
		t.Errorf("an empty attach document: %v\n%s", err, out)
	}
	out, err = (&Line{}).Encode()
	if err != nil || !strings.Contains(string(out), `"workspace": null`) || !strings.Contains(string(out), `"line": ""`) {
		t.Errorf("an unlinked folder's line document: %v\n%s", err, out)
	}
	// The closed lists are closed: a word outside one does not fit, so the writer refuses it.
	for name, doc := range map[string]encoder10{
		"settings scope":  &Settings{Scope: "home"},
		"settings from":   &Settings{Scope: "machine", Settings: []SettingValue{{Key: "cache_keep_days", Value: int64(30), From: "env"}}},
		"settings value":  &Settings{Scope: "machine", Settings: []SettingValue{{Key: "cache_keep_days", Value: true, From: "file"}}},
		"attach command":  &Attach{Command: "remove"},
		"attach result":   &Attach{Command: "attach", Result: ptr("replaced")},
		"line ladder":     &Line{Ladder: ptr("yellow")},
		"settings key":    &Settings{Scope: "workspace", Settings: []SettingValue{{Key: "Status-Writes", From: "file"}}},
		"attach version0": &Attach{Command: "attach", Version: func() *int64 { n := int64(0); return &n }()},
	} {
		if _, err := doc.Encode(); err == nil {
			t.Errorf("%s: a word outside its list was written", name)
		}
	}
	// The example's status, its release state outside the closed list, does not fit; inside it, it does.
	f := MustLookup("status")
	for state, fits := range map[string]bool{"older": false, "newer": true, "not-a-release": true} {
		v, err := readers["status"](example(t, "status.json"))
		if err != nil {
			t.Fatal(err)
		}
		v.(*Status).BonsaiRelease.State = state
		doc, err := f.Document(v)
		if err != nil {
			t.Fatal(err)
		}
		if err := f.Check(doc); (err == nil) != fits {
			t.Errorf("bonsai_release.state %s: fits %v, want %v (%v)", state, err == nil, fits, err)
		}
	}
	// Status writes bonsai_release null until step 5.6.5 fills it (note 5.6.0, 3), at its end.
	st, err := MustLookup("status").Document(&Status{})
	if err != nil {
		t.Fatal(err)
	}
	keys := st.Keys()
	if keys[len(keys)-1] != "bonsai_release" {
		t.Errorf("status --json ends %v, want bonsai_release last", keys[len(keys)-3:])
	}
	if v, _ := st.Get("bonsai_release"); v != nil {
		t.Errorf("status --json's bonsai_release writes %s, want null until step 5.6.5", schema.Show(v))
	}
}

// encoder10 is a document with a writer.
type encoder10 interface{ Encode() ([]byte, error) }

func ptr(s string) *string { return &s }
