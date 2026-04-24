// Package updateflow implements the cinematic 5-stage `bonsai update` flow.
//
// Rail:
//
//	探 DISCOVER  択 SELECT  同 SYNC  衝 CONFLICT  結 YIELD
//
// Stages are either on-rail (DISCOVER / SELECT / SYNC / YIELD) or off-rail
// (CONFLICT — chromeless, spliced lazily when the Sync step surfaces a
// conflict list). Every chrome primitive is imported from initflow — the
// package never reimplements header/footer/rail. Plan 31 Phase F.
package updateflow

import (
	"github.com/LastStep/Bonsai/internal/catalog"
	"github.com/LastStep/Bonsai/internal/config"
	"github.com/LastStep/Bonsai/internal/generate"
	"github.com/LastStep/Bonsai/internal/tui/initflow"
)

// Stage indices in the update-flow rail. Kept as named constants so
// splicer logic references them by name rather than magic integers. The
// Conflict stage renders off-rail (StageIdxOffRail) but still lives as a
// real step in the harness list when wr.HasConflicts() is true.
const (
	StageIdxDiscover = 0
	StageIdxSelect   = 1
	StageIdxSync     = 2
	StageIdxYield    = 3

	// StageIdxOffRail matches addflow's sentinel — off-rail stages suppress
	// the rail row in renderFrame (negative rail index).
	StageIdxOffRail = -1
)

// StageLabels holds the four on-rail update-flow labels in order. The
// Conflict stage has its own off-rail label rendered inside its body
// (see conflicts.go).
var StageLabels = []initflow.StageLabel{
	{Kanji: "探", Kana: "さがす", English: "DISCOVER"},
	{Kanji: "択", Kana: "えらぶ", English: "SELECT"},
	{Kanji: "同", Kana: "どう", English: "SYNC"},
	{Kanji: "結", Kana: "むすぶ", English: "YIELD"},
}

// AgentDiscoveries bundles the scan result for a single installed agent.
// Valid files are user-promotable (clean frontmatter); Invalid files are
// surfaced as warnings inside the Discover stage panel.
type AgentDiscoveries struct {
	AgentName  string                    // machine name (cfg.Agents key)
	AgentLabel string                    // display name (falls back to AgentName)
	Installed  *config.InstalledAgent    // pointer — mutated downstream when user accepts
	Valid      []generate.DiscoveredFile // user-selectable rows
	Invalid    []generate.DiscoveredFile // surfaced as warnings, not user-selectable
}

// Result is the outcome payload returned from Run. Caller applies any
// post-flow persistence (cfg.Save / lock.Save) — the flow itself writes
// through generate.* pathways but leaves final config serialization to
// the command layer so `cmd/update.go` keeps full control.
type Result struct {
	// ConfigChanged is true when at least one user-accepted discovery
	// mutated the project config (installed.Skills/Workflows/... grew).
	// Triggers a cfg.Save post-run.
	ConfigChanged bool

	// WriteResult is the filesystem snapshot populated by the Sync
	// action's generator calls. Post-run consumers read it for conflict
	// dispatch and the YIELD success panel's counts.
	WriteResult *generate.WriteResult

	// SyncErr, when non-nil, is the aggregated error surfaced by Sync.
	// The YIELD stage renders an error panel instead of the normal
	// success card.
	SyncErr error

	// Cancelled is true when the user Ctrl-C'd. Caller skips persistence.
	Cancelled bool
}

// FlowInputs carries the shared dependencies runUpdate passes into the
// flow. Kept as a struct so the splicer closures can capture by pointer
// once rather than threading individual args across every stage ctor.
type FlowInputs struct {
	Cwd     string
	Version string
	Cfg     *config.ProjectConfig
	Cat     *catalog.Catalog
	Lock    *config.LockFile
}
