// Package nonint drives `bonsai init` and `bonsai add` without TUI prompts.
// Inputs are loaded from a YAML file (same shape as .bonsai.yaml), filesystem
// outcomes are emitted as JSON Lines on the caller-supplied writer, and
// validation errors surface through plain Go errors so the cobra entry point
// can choose the exit code.
//
// The package is intentionally free of TUI imports — every byte written here
// is JSON, every diagnostic is an error string. That keeps the headless code
// path safe to drive from a Python subprocess in Bonsai-Eval rung-3 (Plan 38)
// without ANSI escape codes leaking into the test transcript.
package nonint

import (
	"encoding/json"
	"io"
)

// fileEvent is the JSONL shape for "file" and "warning" events. Per-file
// payloads stay tight via `omitempty` because zero-valued strings are
// uninformative noise in the JSONL stream.
type fileEvent struct {
	Event   string `json:"event"`
	Path    string `json:"path,omitempty"`
	Action  string `json:"action,omitempty"` // created | updated | unchanged | skipped | conflict
	Source  string `json:"source,omitempty"`
	Message string `json:"message,omitempty"`
}

// summaryEvent is the JSONL shape for the terminal summary line. Count fields
// are deliberately NOT `omitempty` so an all-zero run still emits every key —
// downstream consumers (Bonsai-Eval rung-3 telemetry) expect a stable shape so
// they can parse `created/updated/.../conflicts` unconditionally without
// special-casing missing keys.
type summaryEvent struct {
	Event     string `json:"event"`
	Created   int    `json:"created"`
	Updated   int    `json:"updated"`
	Unchanged int    `json:"unchanged"`
	Skipped   int    `json:"skipped"`
	Conflicts int    `json:"conflicts"`
}

// Event is the public shape used by callers (cmd/init.go, cmd/add.go) to feed
// events into Emit. The constructors below collapse it to the right private
// shape on serialisation so the all-zero-counts contract holds.
//
// Action and Path carry the per-file payload; Created/Updated/.../Conflicts
// carry the summary counts; Message is the warning channel. Callers should
// use the EmitFile / EmitSummary / EmitWarning helpers instead of populating
// fields by hand.
type Event struct {
	Event     string
	Path      string
	Action    string
	Source    string
	Message   string
	Created   int
	Updated   int
	Unchanged int
	Skipped   int
	Conflicts int
}

// EmitFile writes one `{"event":"file",...}` line. Returns the underlying
// writer error so the caller can decide whether to bail out (e.g. broken
// pipe to a parent process that died).
func EmitFile(w io.Writer, path, action, source string) error {
	return emitJSON(w, fileEvent{
		Event:  "file",
		Path:   path,
		Action: action,
		Source: source,
	})
}

// EmitSummary writes the terminal summary line. All five count fields are
// always present in the serialised JSON even when zero, by design.
func EmitSummary(w io.Writer, created, updated, unchanged, skipped, conflicts int) error {
	return emitJSON(w, summaryEvent{
		Event:     "summary",
		Created:   created,
		Updated:   updated,
		Unchanged: unchanged,
		Skipped:   skipped,
		Conflicts: conflicts,
	})
}

// EmitWarning writes a `{"event":"warning","message":"..."}` line. Used for
// non-fatal anomalies — currently only the lock-save-failure path uses this
// (and it routes to stderr, never stdout — see runner.go).
func EmitWarning(w io.Writer, message string) error {
	return emitJSON(w, fileEvent{
		Event:   "warning",
		Message: message,
	})
}

// emitJSON marshals v to a single JSON line + newline. json.Marshal is used
// rather than an Encoder so the JSON line is byte-identical regardless of
// writer buffering — the Encoder adds its own trailing newline but our
// callers want explicit control over framing.
func emitJSON(w io.Writer, v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	data = append(data, '\n')
	_, err = w.Write(data)
	return err
}
