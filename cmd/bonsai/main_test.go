package main

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

// The stub answers --version alone, and refuses anything else with exit 2, ASCII lines on stderr, and a next step.
func TestRun(t *testing.T) {
	cases := []struct {
		name     string
		args     []string
		code     int
		stdout   string
		inStderr string
	}{
		{"version", []string{"--version"}, 0, "bonsai dev\n", ""},
		{"no command", nil, 2, "", "bonsai: no command given:"},
		{"unknown command", []string{"init"}, 2, "", `bonsai: "init" is not a command yet:`},
		{"version with more", []string{"--version", "--json"}, 2, "", "bonsai: --version takes no other argument:"},
		{"non-ASCII argument", []string{"b\u00f6nsai"}, 2, "", `bonsai: "b\u00f6nsai" is not a command yet:`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := run(c.args, &stdout, &stderr)
			if code != c.code {
				t.Errorf("exit code %d, want %d", code, c.code)
			}
			if stdout.String() != c.stdout {
				t.Errorf("stdout %q, want %q", stdout.String(), c.stdout)
			}
			if c.inStderr == "" {
				if stderr.Len() != 0 {
					t.Errorf("stderr %q, want nothing", stderr.String())
				}
				return
			}
			out := stderr.String()
			if !strings.HasPrefix(out, c.inStderr) {
				t.Errorf("stderr %q, want it to start with %q", out, c.inStderr)
			}
			if !strings.Contains(out, "\nnext: run `bonsai --version`") {
				t.Errorf("stderr %q names no next step", out)
			}
			for i := 0; i < len(out); i++ {
				if out[i] > 0x7e || (out[i] < 0x20 && out[i] != '\n') {
					t.Fatalf("stderr byte %d is %#x, not printable ASCII: %q", i, out[i], out)
				}
			}
		})
	}
}

// A --version whose answer cannot be written exits 3 (runtime), not 0.
func TestRunVersionWriteFails(t *testing.T) {
	var stderr bytes.Buffer
	if code := run([]string{"--version"}, failingWriter{}, &stderr); code != 3 {
		t.Errorf("exit code %d, want 3", code)
	}
}

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) { return 0, errors.New("closed") }
