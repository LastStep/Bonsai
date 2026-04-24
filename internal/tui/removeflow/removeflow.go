// Package removeflow implements the cinematic 4-stage `bonsai remove` flow.
//
// The package mirrors internal/tui/addflow's shape — chromeless stages that
// compose the shared persistent chrome (header + enso rail + footer) from
// initflow around a per-stage body — but with a 4-segment rail tuned for the
// remove journey:
//
//	択 SELECT  観 OBSERVE  確 CONFIRM  結 YIELD
//
// The Conflicts stage is spliced in chromelessly (rail index StageIdxOffRail)
// when the post-generate write-pipeline produces user-modified files, so the
// visible rail does not churn between Confirm and Yield.
//
// Two entry shapes:
//
//   - Agent removal (`bonsai remove <agent>`) — Select is skipped; Observe
//     previews the installed agent's ability tree, Confirm gates the write,
//     Conflicts reconciles per-file picks, Yield shows the summary card.
//   - Item removal (`bonsai remove skill foo`) — Select fires iff multiple
//     agents have the item installed; Observe, Confirm, Conflicts, Yield
//     share the same shape as agent-remove.
//
// Plan 31 reference: station/Playbook/Plans/Active/31-v03-release-readiness.md.
// Phase E — port `bonsai remove` from the raw harness + tui.FatalPanel path to
// a dedicated flow package, matching the Plan 22/23/27/28/29/30 cinematic
// rollout across init/add/list/catalog/guide.
package removeflow

import (
	"github.com/LastStep/Bonsai/internal/tui/initflow"
)

// Stage indices in the remove-flow rail. Conflicts renders off-rail so the
// visible 4-segment rail stays stable when the conflicts picker splices in.
const (
	StageIdxSelect  = 0
	StageIdxObserve = 1
	StageIdxConfirm = 2
	StageIdxYield   = 3

	// StageIdxOffRail is the sentinel rail index for Conflicts — the base
	// Stage skips the rail row when its idx is negative.
	StageIdxOffRail = -1
)

// StageLabels holds the four canonical remove-flow stage labels in order.
// Ordering matches the flow's state machine: pick the target (when ambiguous),
// preview what will be removed, confirm the destructive action, report.
//
//	択 えらぶ    Select    — pick the agent (iff item-remove with multiple matches)
//	観 みる      Observe   — preview what will be removed
//	確 かくにん  Confirm   — explicit yes/no gate before the write
//	結 むすぶ    Yield     — completion card
var StageLabels = []initflow.StageLabel{
	{Kanji: "択", Kana: "えらぶ", English: "SELECT"},
	{Kanji: "観", Kana: "みる", English: "OBSERVE"},
	{Kanji: "確", Kana: "かくにん", English: "CONFIRM"},
	{Kanji: "結", Kana: "むすぶ", English: "YIELD"},
}

// AgentOption is the per-row shape consumed by SelectStage on the item-remove
// branch. Rows are one agent that has the named ability installed; selection
// returns the machine name (or "_all_" for the aggregate row).
type AgentOption struct {
	Name        string // machine identifier, or "_all_" for the aggregate row
	DisplayName string // human-readable label shown in the row
	Workspace   string // workspace path, rendered muted after the name
	All         bool   // true for the aggregate "All agents" row
}

// AbilityCounts captures how many installed abilities each category carries
// for the agent being removed (agent-remove branch) or the target(s) of an
// item removal. Used by Observe's preview panel and Yield's summary.
type AbilityCounts struct {
	Skills    int
	Workflows int
	Protocols int
	Sensors   int
	Routines  int
}

// Total returns the sum across categories.
func (c AbilityCounts) Total() int {
	return c.Skills + c.Workflows + c.Protocols + c.Sensors + c.Routines
}

// Outcome is the cross-stage scratchpad populated by the action closure and
// consumed by the post-harness cleanup in cmd/remove.go. Kept here (not in
// cmd/) so test helpers can stamp synthetic outcomes without back-importing
// cmd.
type Outcome struct {
	// Ran is true once the action closure executed (even on error).
	Ran bool
	// Err is non-nil when the write pipeline failed; routed to a warning by
	// the caller and short-circuits the Yield success render.
	Err error
}
