// Package nonint drives the mutating Bonsai commands (`init`, `add`, and —
// added by later Plan 41 phases — `update`, `remove`) without TUI prompts.
// Inputs are typed options (often loaded from a YAML file with the same shape
// as .bonsai.yaml); each core returns a structured *Result and validation
// errors surface through plain Go errors so the cobra entry point can choose
// the exit code. The CLI adapter serialises the Result to JSON Lines on
// stdout (via EmitJSONL) and prints Result.Warnings to stderr.
//
// The package is intentionally free of TUI imports — every byte EmitJSONL
// writes is JSON, every diagnostic is an error string or a stderr warning.
// That keeps the headless code path safe to drive from a Python subprocess
// in Bonsai-Eval rung-3 (Plan 38) without ANSI escape codes leaking into the
// test transcript, and makes the cores ready for a thin MCP wrapper (Plan 42).
package nonint

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/LastStep/Bonsai/internal/generate"
)

// fileEvent is the JSONL shape for the "file" event. Per-file payloads stay
// tight via `omitempty` because zero-valued strings are uninformative noise
// in the JSONL stream.
type fileEvent struct {
	Event  string `json:"event"`
	Path   string `json:"path,omitempty"`
	Action string `json:"action,omitempty"` // created | updated | unchanged | skipped | conflict
	Source string `json:"source,omitempty"`
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

// EmitJSONL serialises a *Result to JSON Lines on w: one `{"event":"file",...}`
// line per file outcome (in write order), then a single terminal
// `{"event":"summary",...}` line. It emits ONLY file + summary events —
// r.Warnings are never written here. Warnings live in the Result and the CLI
// adapter prints them to stderr, keeping w (stdout) pure protocol. The empty
// Write case (all-installed short-circuit) renders as a lone zero-count
// summary line, byte-identical to the old EmitSummary(w,0,0,0,0,0).
//
// A nil r or nil r.Write is treated as a zero-count run so callers need no
// guard. This is the single serialisation seam the CLI uses today and the
// MCP adapter (Plan 42) will mirror.
func EmitJSONL(w io.Writer, r *Result) error {
	var wr *generate.WriteResult
	if r != nil {
		wr = r.Write
	}
	var created, updated, unchanged, skipped, conflicts int
	if wr != nil {
		for _, f := range wr.Files {
			action := actionString(f.Action)
			if action == "" {
				continue
			}
			if err := EmitFile(w, f.RelPath, action, f.Source); err != nil {
				return fmt.Errorf("nonint: emit file event: %w", err)
			}
		}
		created, updated, unchanged, skipped, conflicts = wr.Summary()
	}
	if err := EmitSummary(w, created, updated, unchanged, skipped, conflicts); err != nil {
		return fmt.Errorf("nonint: emit summary event: %w", err)
	}
	return nil
}

// actionString maps a generate.FileAction to its JSONL `action` value. Maps
// 1:1 with the FileAction enum — see Plan 39 §A.4.
func actionString(a generate.FileAction) string {
	switch a {
	case generate.ActionCreated:
		return "created"
	case generate.ActionUpdated, generate.ActionForced:
		// ForcedAction shouldn't occur on the non-interactive path (we never
		// call ForceSelected/ForceConflicts), but bucket it with updated for
		// completeness in case a future change introduces it.
		return "updated"
	case generate.ActionUnchanged:
		return "unchanged"
	case generate.ActionSkipped:
		return "skipped"
	case generate.ActionConflict:
		return "conflict"
	default:
		return ""
	}
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
