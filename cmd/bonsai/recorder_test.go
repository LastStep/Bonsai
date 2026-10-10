package main

// bonsai hook start and bonsai hook record from the command line (step 5.2.4). The recorder's own tests, payload by
// payload, are internal/recorder's.

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// The hook words from the command line: help, and hook record printing nothing and exiting 0 whatever it is handed.
func TestHookStartAndRecordWords(t *testing.T) {
	for _, c := range []struct {
		args []string
		in   string
	}{
		{[]string{"hook", "start", "--help"}, "  bonsai hook start\non SessionStart"},
		{[]string{"hook", "record", "--help"}, "  bonsai hook record\non UserPromptSubmit"},
		{[]string{"hook", "--help"}, "bonsai hook record"},
	} {
		code, out, errOut := runArgs(c.args...)
		if code != 0 || !strings.Contains(out, c.in) || errOut != "" {
			t.Errorf("%v: exit %d, %q, %q", c.args, code, out, errOut)
		}
	}
	if code, _, errOut := runArgs("hook", "stop"); code != 2 || !strings.Contains(errOut, "bonsai hook stop is not built yet") {
		t.Errorf("hook stop: exit %d, %q", code, errOut)
	}
}

// hook record, the built binary, on every kind of input: exit 0 and no byte on stdout or stderr, whether the payload
// is empty, broken, huge or unknown, CLAUDE_PROJECT_DIR is unset, the folder holds no bonsai.yaml, bonsai.yaml does not
// read, or the log folder cannot be written; and a good payload writes its record.
func TestHookRecordNeverInterferes(t *testing.T) {
	normal, _ := bonsaiBuilds(t)
	tmp := t.TempDir()
	home := filepath.Join(tmp, "home")
	proj := filepath.Join(tmp, "proj")
	if err := os.MkdirAll(proj, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(proj, "bonsai.yaml"), []byte(linkedYAML), 0o644); err != nil {
		t.Fatal(err)
	}
	plain := filepath.Join(tmp, "plain")
	if err := os.MkdirAll(filepath.Join(plain, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	good := `{"session_id":"s-1","hook_event_name":"Stop","cwd":"` + filepath.ToSlash(proj) + `"}`
	huge := `{"session_id":"s-1","hook_event_name":"PostToolUse","tool_name":"Read","tool_input":{"file_path":"a"},` +
		`"tool_response":"` + strings.Repeat("x", 3<<20) + `"}`
	run := func(name, dir, payload string, env ...string) {
		t.Helper()
		cmd := exec.Command(normal, "hook", "record")
		cmd.Env = append(os.Environ(), "BONSAI_HOME="+home, "CLAUDE_CODE_BRIDGE_SESSION_ID=")
		cmd.Env = append(withoutEnv(cmd.Env, "CLAUDE_PROJECT_DIR"), env...)
		if dir != "" {
			cmd.Env = append(cmd.Env, "CLAUDE_PROJECT_DIR="+dir)
		}
		cmd.Stdin = strings.NewReader(payload)
		var stdout, stderr bytes.Buffer
		cmd.Stdout, cmd.Stderr = &stdout, &stderr
		err := cmd.Run()
		if err != nil || stdout.Len() != 0 || stderr.Len() != 0 {
			t.Errorf("%s: %v, stdout %q, stderr %q", name, err, stdout.String(), stderr.String())
		}
	}
	logDir := filepath.Join(proj, ".bonsai", "local", "log")
	count := func() int {
		raw, _ := os.ReadFile(filepath.Join(logDir, "s-s-1.ndjson"))
		return bytes.Count(raw, []byte("\n"))
	}
	run("empty", proj, "")
	run("broken", proj, `{"session_id":`)
	run("not an object", proj, `["Stop"]`)
	run("unknown event", proj, `{"session_id":"s-1","hook_event_name":"PreCompact"}`)
	run("no session", proj, `{"hook_event_name":"Stop"}`)
	run("a session id that cannot name a file", proj, `{"session_id":"../x","hook_event_name":"Stop"}`)
	run("over the cap", proj, `{"session_id":"s-1","hook_event_name":"Stop","x":"`+strings.Repeat("y", 64<<20)+`"}`)
	if n := count(); n != 0 {
		t.Fatalf("%d records from payloads that do not read", n)
	}
	run("good", proj, good)
	run("huge, read whole", proj, huge)
	run("no CLAUDE_PROJECT_DIR: the payload's cwd", "", good)
	if n := count(); n != 3 {
		t.Fatalf("%d records, want 3", n)
	}
	run("no bonsai.yaml", plain, good)
	if _, err := os.Stat(filepath.Join(plain, ".bonsai")); err == nil {
		t.Error("a folder with no bonsai.yaml got a .bonsai/")
	}
	// A log folder that cannot be written: a file stands where the folder goes.
	if err := os.RemoveAll(logDir); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(logDir, []byte("in the way"), 0o644); err != nil {
		t.Fatal(err)
	}
	run("a log folder that cannot be written", proj, good)
	if err := os.Remove(logDir); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(proj, "bonsai.yaml"), []byte("format: bonsai.workspace/1\nid: [\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	run("a bonsai.yaml that does not read", proj, good)
	if _, err := os.Stat(logDir); err == nil {
		t.Error("a record was written under a bonsai.yaml that does not read")
	}
}

// withoutEnv drops a variable from an environment list.
func withoutEnv(env []string, name string) []string {
	var out []string
	for _, e := range env {
		if !strings.HasPrefix(e, name+"=") {
			out = append(out, e)
		}
	}
	return out
}
