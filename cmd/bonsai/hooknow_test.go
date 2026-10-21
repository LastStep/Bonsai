package main

// bonsai hook now (plan P-T-5795 step 5, piece E): the line it prints in three zones and across a day boundary, the
// JSON a subagent's start gets, exit 0 on any input, and its hook line run through the hook shell (sh on Linux, Git
// Bash on Windows): with bonsai there, at neither place, and with a bonsai that answers exit 2 (a build older than the
// word), where the line still exits 0 and the prompt runs.

import (
	"bytes"
	"encoding/json"
	"errors"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
	_ "time/tzdata" // tests only: Go on Windows ignores TZ, so each test injects its zone

	"github.com/LastStep/Bonsai/internal/engine"
)

func mustZone(t *testing.T, name string) *time.Location {
	t.Helper()
	loc, err := time.LoadLocation(name)
	if err != nil {
		t.Fatal(err)
	}
	return loc
}

func TestNowLine(t *testing.T) {
	at := func(s string) time.Time {
		v, err := time.Parse(time.RFC3339, s)
		if err != nil {
			t.Fatal(err)
		}
		return v
	}
	cases := []struct {
		t    time.Time
		zone string
		want string
	}{
		{at("2026-10-21T08:19:04Z"), "Asia/Kolkata", "Now: 2026-10-21 08:19 UTC (local 13:49 +05:30)"},
		{at("2026-10-21T23:50:00Z"), "Asia/Kolkata", "Now: 2026-10-21 23:50 UTC (local 2026-10-22 05:20 +05:30)"},
		{at("2026-10-21T08:19:59Z"), "UTC", "Now: 2026-10-21 08:19 UTC (local 08:19 +00:00)"},
		{at("2026-07-01T02:00:00Z"), "America/New_York", "Now: 2026-07-01 02:00 UTC (local 2026-06-30 22:00 -04:00)"},
		{at("2026-01-15T15:30:00Z"), "America/New_York", "Now: 2026-01-15 15:30 UTC (local 10:30 -05:00)"},
	}
	for _, c := range cases {
		if got := nowLine(c.t, mustZone(t, c.zone)); got != c.want {
			t.Errorf("%s in %s: %q, want %q", c.t, c.zone, got, c.want)
		}
	}
}

func TestHookNowOutput(t *testing.T) {
	when := time.Date(2026, 10, 21, 8, 19, 4, 0, time.UTC)
	loc := mustZone(t, "Asia/Kolkata")
	line := "Now: 2026-10-21 08:19 UTC (local 13:49 +05:30)"
	run := func(stdin string, terminal bool) string {
		t.Helper()
		var out bytes.Buffer
		if code := hookNow(strings.NewReader(stdin), terminal, &out, when, loc); code != 0 {
			t.Fatalf("exit %d for %q", code, stdin)
		}
		return out.String()
	}
	// A prompt, and anything that is not a subagent's start: the plain line.
	for _, in := range []string{
		`{"hook_event_name":"UserPromptSubmit","session_id":"s","cwd":"/x","prompt":"what time is it?"}`,
		`{"hook_event_name":"SessionStart","session_id":"s"}`,
		`{}`, `not json`, ``, `{"hook_event_name":"SubagentStart"} trailing`,
	} {
		if got := run(in, false); got != line+"\n" {
			t.Errorf("%q: %q", in, got)
		}
	}
	if got := run("", true); got != line+"\n" {
		t.Errorf("a terminal: %q", got)
	}
	// A subagent's start: one JSON object alone on stdout, its context the line.
	got := run(`{"hook_event_name":"SubagentStart","session_id":"s","cwd":"/x","agent_id":"a1","agent_type":"Explore"}`, false)
	if want := `{"hookSpecificOutput":{"hookEventName":"SubagentStart","additionalContext":"` + line + `"}}` + "\n"; got != want {
		t.Errorf("SubagentStart: %q, want %q", got, want)
	}
	var doc map[string]map[string]string
	if err := json.Unmarshal([]byte(got), &doc); err != nil || doc["hookSpecificOutput"]["additionalContext"] != line {
		t.Errorf("SubagentStart's JSON does not read back: %v %v", err, doc)
	}
}

// runNowLine runs a hook line through the hook shell with stdin, giving its exit code, stdout and stderr.
func runNowLine(t *testing.T, line, stdin string) (int, string, string) {
	t.Helper()
	cmd := exec.Command(hookShell(t), "-c", line)
	cmd.Env = hookEnv("PATH="+hookPath(t), "BONSAI_HOME="+t.TempDir())
	cmd.Stdin = strings.NewReader(stdin)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Run()
	code := 0
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		code = ee.ExitCode()
	} else if err != nil {
		t.Fatalf("running the hook line: %v", err)
	}
	return code, stdout.String(), stderr.String()
}

func TestTheNowLineThroughTheShell(t *testing.T) {
	normal, _ := bonsaiBuilds(t)
	exe := filepath.Base(normal)
	tmp := t.TempDir()
	first := filepath.Join(tmp, "place one", exe)
	second := filepath.Join(tmp, "place-two", exe)
	places := []string{shellPath(first), shellPath(second)}
	prompt := `{"hook_event_name":"UserPromptSubmit","session_id":"s","cwd":"/x","prompt":"hi"}`
	sub := `{"hook_event_name":"SubagentStart","session_id":"s","cwd":"/x","agent_id":"a1","agent_type":"Explore"}`
	shape := regexp.MustCompile(`^Now: \d{4}-\d{2}-\d{2} \d{2}:\d{2} UTC \(local (\d{4}-\d{2}-\d{2} )?\d{2}:\d{2} [+-]\d{2}:\d{2}\)$`)

	t.Run("neither place", func(t *testing.T) {
		for _, in := range []string{prompt, sub} {
			if code, out, errOut := runNowLine(t, engine.OpenHookCommand("now", places...), in); code != 0 || out != "" || errOut != "" {
				t.Fatalf("exit %d, stdout %q, stderr %q; want 0 and nothing", code, out, errOut)
			}
		}
	})
	t.Run("a bonsai without the word", func(t *testing.T) {
		// The real build asked for a hook word it does not have answers exit 2, as a build older than hook now does;
		// the line turns it into exit 0, so the prompt is never blocked.
		copyFile(t, normal, second)
		line := engine.OpenHookCommand("nosuchword", places...)
		code, out, errOut := runNowLine(t, line, prompt)
		if code != 0 || out != "" || !strings.Contains(errOut, `is not a hook`) {
			t.Fatalf("exit %d, stdout %q, stderr %q; want 0, nothing on stdout, the refusal on stderr", code, out, errOut)
		}
		// The same word on the line form that does not end || exit 0: exit 2, which would block the prompt.
		if code, _, _ := runNowLine(t, engine.HookCommand("nosuchword", false, places...), prompt); code != 2 {
			t.Fatalf("the plain form: exit %d, want 2 (the case the open form is for)", code)
		}
	})
	t.Run("bonsai there", func(t *testing.T) {
		copyFile(t, normal, first)
		line := engine.OpenHookCommand("now", places...)
		before := time.Now().UTC().Truncate(time.Minute)
		code, out, errOut := runNowLine(t, line, prompt)
		after := time.Now().UTC()
		if code != 0 || errOut != "" || !strings.HasSuffix(out, "\n") || strings.Count(out, "\n") != 1 || !shape.MatchString(strings.TrimSuffix(out, "\n")) {
			t.Fatalf("a prompt: exit %d, stdout %q, stderr %q", code, out, errOut)
		}
		utc, err := time.Parse("2006-01-02 15:04", out[len("Now: "):len("Now: 2006-01-02 15:04")])
		if err != nil || utc.Before(before) || utc.After(after) {
			t.Errorf("the line's UTC time %s is not between %s and %s (%v)", utc, before, after, err)
		}
		code, out, errOut = runNowLine(t, line, sub)
		var doc struct {
			Hook struct {
				Event   string `json:"hookEventName"`
				Context string `json:"additionalContext"`
			} `json:"hookSpecificOutput"`
		}
		if code != 0 || errOut != "" || json.Unmarshal([]byte(out), &doc) != nil || doc.Hook.Event != "SubagentStart" || !shape.MatchString(doc.Hook.Context) {
			t.Fatalf("a subagent's start: exit %d, stdout %q, stderr %q", code, out, errOut)
		}
	})
}
