package nonint

import "github.com/LastStep/Bonsai/internal/generate"

// Result is the structured value every headless mutating core (RunInit,
// RunAdd, and — added by later Plan 41 phases — RunUpdate / RunRemove*)
// returns. Both the CLI JSONL adapter (EmitJSONL → stdout) and the future
// MCP adapter (Plan 42 → structuredContent) consume it. The core itself
// performs no output, so the serialisation seam is the only place a stream
// is touched.
//
// Result is intentionally thin — exactly one field over generate.WriteResult —
// until Plan 42 enriches it for MCP structuredContent. It is HEADLESS-ONLY:
// the cinematic TTY path keeps using updateflow.Result (with its
// ConfigChanged/Cancelled/SyncErr flow-control fields the Yield stage reads).
// Do NOT unify the two.
type Result struct {
	// Write holds the per-file outcomes (created / updated / unchanged /
	// skipped / conflicts). EmitJSONL walks Write.Files in write order.
	Write *generate.WriteResult
	// Warnings are non-fatal anomalies (lock-save failure, invalid
	// discoveries). They are NEVER written to stdout — the CLI adapter
	// prints them to stderr as plain text. This keeps stdout pure JSONL
	// protocol (a hard requirement for the Plan 42 stdio MCP server).
	Warnings []string
}

// Counts delegates to Write.Summary(), returning the five action tallies.
// A nil Write yields all-zero counts so callers (and the all-installed
// zero-summary short-circuit) need no nil guard.
func (r *Result) Counts() (created, updated, unchanged, skipped, conflicts int) {
	if r == nil || r.Write == nil {
		return 0, 0, 0, 0, 0
	}
	return r.Write.Summary()
}
