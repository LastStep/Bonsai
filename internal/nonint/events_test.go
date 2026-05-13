package nonint

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

// TestEmitFile_JSONLShape emits one file event and asserts the line round-trips
// through json.Unmarshal with the expected key/value set. Per-event keys honour
// `omitempty` so an empty Source is dropped from the wire payload.
func TestEmitFile_JSONLShape(t *testing.T) {
	var buf bytes.Buffer
	if err := EmitFile(&buf, "station/CLAUDE.md", "created", "tech-lead"); err != nil {
		t.Fatalf("EmitFile: %v", err)
	}
	got := buf.String()
	if !strings.HasSuffix(got, "\n") {
		t.Fatalf("event must end with newline; got %q", got)
	}
	if strings.Count(got, "\n") != 1 {
		t.Fatalf("event must be exactly one line; got %d newlines in %q", strings.Count(got, "\n"), got)
	}
	var parsed map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(got)), &parsed); err != nil {
		t.Fatalf("unmarshal: %v (raw: %q)", err, got)
	}
	if parsed["event"] != "file" {
		t.Errorf("event field: want file, got %v", parsed["event"])
	}
	if parsed["path"] != "station/CLAUDE.md" {
		t.Errorf("path field: want station/CLAUDE.md, got %v", parsed["path"])
	}
	if parsed["action"] != "created" {
		t.Errorf("action field: want created, got %v", parsed["action"])
	}
	if parsed["source"] != "tech-lead" {
		t.Errorf("source field: want tech-lead, got %v", parsed["source"])
	}
}

// TestEmitFile_OmitemptyDropsEmptyFields verifies the per-file event drops
// zero-value strings (source, etc.) so a sparse run doesn't bloat the JSONL
// stream.
func TestEmitFile_OmitemptyDropsEmptyFields(t *testing.T) {
	var buf bytes.Buffer
	if err := EmitFile(&buf, "a.md", "skipped", ""); err != nil {
		t.Fatalf("EmitFile: %v", err)
	}
	got := buf.String()
	if strings.Contains(got, "source") {
		t.Errorf("empty source field must be omitted; got %q", got)
	}
}

// TestEmitSummary_AlwaysEmitsAllCountFields enforces the contract: even when
// every count is zero, the summary line MUST carry created/updated/unchanged/
// skipped/conflicts keys so downstream consumers (Bonsai-Eval rung-3
// telemetry) can parse the shape without special-casing missing keys.
func TestEmitSummary_AlwaysEmitsAllCountFields(t *testing.T) {
	var buf bytes.Buffer
	if err := EmitSummary(&buf, 0, 0, 0, 0, 0); err != nil {
		t.Fatalf("EmitSummary: %v", err)
	}
	got := buf.String()
	for _, key := range []string{`"created":0`, `"updated":0`, `"unchanged":0`, `"skipped":0`, `"conflicts":0`} {
		if !strings.Contains(got, key) {
			t.Errorf("summary missing %q; got %q", key, got)
		}
	}
	if !strings.Contains(got, `"event":"summary"`) {
		t.Errorf("summary line missing event key; got %q", got)
	}
}

// TestEmitSummary_NonZeroCounts round-trips a non-trivial count distribution
// through json.Unmarshal and asserts every field matches what the caller
// emitted.
func TestEmitSummary_NonZeroCounts(t *testing.T) {
	var buf bytes.Buffer
	if err := EmitSummary(&buf, 5, 2, 7, 1, 3); err != nil {
		t.Fatalf("EmitSummary: %v", err)
	}
	var parsed struct {
		Event     string `json:"event"`
		Created   int    `json:"created"`
		Updated   int    `json:"updated"`
		Unchanged int    `json:"unchanged"`
		Skipped   int    `json:"skipped"`
		Conflicts int    `json:"conflicts"`
	}
	if err := json.Unmarshal(bytes.TrimSpace(buf.Bytes()), &parsed); err != nil {
		t.Fatalf("unmarshal: %v (raw: %q)", err, buf.String())
	}
	if parsed.Event != "summary" {
		t.Errorf("event: want summary, got %s", parsed.Event)
	}
	if parsed.Created != 5 || parsed.Updated != 2 || parsed.Unchanged != 7 || parsed.Skipped != 1 || parsed.Conflicts != 3 {
		t.Errorf("count mismatch: got %+v", parsed)
	}
}

// TestEmit_NoStyling asserts the writer receives plain JSON bytes only — no
// ANSI escape sequences, no LipGloss styling. This is a guardrail against
// future drift; the nonint package must never import tui/styles.
func TestEmit_NoStyling(t *testing.T) {
	var buf bytes.Buffer
	_ = EmitFile(&buf, "x", "created", "src")
	_ = EmitSummary(&buf, 1, 0, 0, 0, 0)
	out := buf.String()
	// ANSI escapes start with \x1b — a quick allowlist check is sufficient
	// here because the JSONL output is straight stdlib encoding/json.
	if strings.ContainsRune(out, '\x1b') {
		t.Errorf("output contains ANSI escape; got %q", out)
	}
}

// TestEmitWarning_Shape verifies the warning channel produces the documented
// event=warning + message envelope.
func TestEmitWarning_Shape(t *testing.T) {
	var buf bytes.Buffer
	if err := EmitWarning(&buf, "could not save lock file"); err != nil {
		t.Fatalf("EmitWarning: %v", err)
	}
	var parsed map[string]any
	if err := json.Unmarshal(bytes.TrimSpace(buf.Bytes()), &parsed); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if parsed["event"] != "warning" {
		t.Errorf("event: want warning, got %v", parsed["event"])
	}
	if parsed["message"] != "could not save lock file" {
		t.Errorf("message: got %v", parsed["message"])
	}
}
