package main

// bonsai check --schema <format> (plan-5 5.1.4a): every format, for a person and with --json; an unknown name
// exits 2 listing every name.

import (
	"bytes"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/formats"
	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/schema"
)

func TestCheckSchema(t *testing.T) {
	if len(format.Names()) != len(formats.Names) {
		t.Fatalf("check --schema knows %d formats, the set holds %d", len(format.Names()), len(formats.Names))
	}
	for i, name := range format.Names() {
		f := format.All[i]
		t.Run(f.Name, func(t *testing.T) {
			for _, given := range []string{name, f.Name, f.Versioned()} {
				var stdout, stderr bytes.Buffer
				if code := run([]string{"check", "--schema", given}, &stdout, &stderr); code != 0 || stderr.Len() > 0 {
					t.Fatalf("check --schema %s: exit %d, %s", given, code, stderr.String())
				}
				asciiOnly(t, "check --schema "+given, stdout.String())
				if !strings.HasPrefix(stdout.String(), f.Versioned()) || !strings.Contains(stdout.String(), "\nFields:\n  format\n") &&
					f.Name != "error" {
					t.Errorf("check --schema %s printed:\n%s", given, stdout.String())
				}
			}
			var stdout, stderr bytes.Buffer
			if code := run([]string{"check", "--json", "--schema=" + name}, &stdout, &stderr); code != 0 {
				t.Fatalf("check --schema --json: exit %d, %s", code, stderr.String())
			}
			asciiOnly(t, "check --schema --json", stdout.String())
			got, err := schema.Decode(stdout.Bytes())
			if err != nil {
				t.Fatal(err)
			}
			raw, _ := formats.Schema(f.Name)
			want, _ := schema.Decode(raw)
			if !schema.Equal(got, want) {
				t.Errorf("--json does not print formats/schemas/%s.schema.json", f.Name)
			}
		})
	}
	for _, c := range []struct {
		args []string
		says string
	}{
		{[]string{"check", "--schema", "bonsai.nope"}, `"bonsai.nope" is not one of Bonsai's formats`},
		{[]string{"check", "--schema", "task/2"}, `"task/2" is not one of Bonsai's formats`},
		{[]string{"check", "--schema"}, "check --schema needs a format's name"},
		{[]string{"check", "--schema", "task", "--schema", "run"}, "check takes --schema once"},
	} {
		var stdout, stderr bytes.Buffer
		if code := run(c.args, &stdout, &stderr); code != 2 || stdout.Len() > 0 {
			t.Errorf("%v: exit %d, stdout %q", c.args, code, stdout.String())
		}
		asciiOnly(t, "the refusal", stderr.String())
		if !strings.Contains(stderr.String(), c.says) || !strings.Contains(stderr.String(), "\nnext: ") {
			t.Errorf("%v: %s", c.args, stderr.String())
		}
		if c.says != "check takes --schema once" && !strings.Contains(stderr.String(), strings.Join(format.Names(), ", ")) {
			t.Errorf("%v: the refusal does not list every name: %s", c.args, stderr.String())
		}
		// With --json, the refusal is check's document, its error object naming the same (step 5.1.4b).
		stdout.Reset()
		stderr.Reset()
		if code := run(append(c.args, "--json"), &stdout, &stderr); code != 2 || stderr.Len() > 0 {
			t.Errorf("%v --json: exit %d, stderr %q", c.args, code, stderr.String())
		}
		e := errorIn(fits(t, stdout.String(), "check"), "check")
		n, _ := e.Get("next")
		if !strings.Contains(e.String("message"), c.says) ||
			(c.says != "check takes --schema once" && !strings.Contains(n.(schema.Object).String("do"), strings.Join(format.Names(), ", "))) {
			t.Errorf("%v --json: %s", c.args, stdout.String())
		}
	}
	var stdout, stderr bytes.Buffer
	run([]string{"check", "--help"}, &stdout, &stderr)
	help := stdout.String()
	for _, want := range []string{"--schema <format>", "Example: bonsai check --schema bonsai.task", "a format Bonsai does not know"} {
		if !strings.Contains(strings.Join(strings.Fields(help), " "), strings.Join(strings.Fields(want), " ")) {
			t.Errorf("check --help does not say %q", want)
		}
	}
	for _, line := range strings.Split(help, "\n") {
		if len(line) > 120 {
			t.Errorf("check --help has a line of %d characters", len(line))
		}
	}
	for _, name := range format.Names() {
		if !strings.Contains(strings.Join(strings.Fields(help), " "), name) {
			t.Errorf("check --help does not name %s", name)
		}
	}
}
