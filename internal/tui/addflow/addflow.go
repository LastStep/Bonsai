// Package addflow implements the cinematic 4-stage `bonsai add` flow.
//
// The package mirrors internal/tui/initflow's shape — chromeless stages that
// compose a shared persistent chrome (header + enso rail + footer) around a
// per-stage body — but with a 4-segment rail specific to the add journey:
//
//	選 SELECT  枝 BRANCHES  観 OBSERVE  結 YIELD
//
// Every chrome primitive (RenderHeader, RenderEnsoRail, RenderFooter,
// RenderMinSizeFloor, ClampColumns, Viewport, PanelContentWidth, design
// tokens) is imported from initflow — addflow does not reimplement any of
// them. The rail length adapts because initflow.RenderEnsoRail accepts an
// explicit label slice (Plan 23 Phase 1 refactor).
//
// Plan 23 reference: station/Playbook/Plans/Active/23-uiux-phase2-add.md.
// Plan 27 reference: station/Playbook/Plans/Active/27-add-flow-polish.md —
// shrinks the visible rail from 7 → 4 segments; Ground, Grow, and Conflicts
// stages remain in the splicer but render off-rail (rail index sentinel
// StageIdxOffRail).
package addflow

import (
	"github.com/LastStep/Bonsai/internal/catalog"
	"github.com/LastStep/Bonsai/internal/tui/initflow"
)

// Stage indices in the add-flow rail. Kept as named constants so ctors and
// splicers reference them by name rather than literal integers. Plan 27
// shrinks the rail to four visible stages; Ground, Grow, and Conflicts stages
// are off-rail and use StageIdxOffRail as their rail index sentinel.
const (
	StageIdxSelect   = 0
	StageIdxBranches = 1
	StageIdxObserve  = 2
	StageIdxYield    = 3

	// StageIdxOffRail is the sentinel rail index for stages that render
	// without a visible rail tab (Ground, Grow, Conflicts). The base Stage
	// skips the rail row when its idx is negative.
	StageIdxOffRail = -1
)

// StageLabels holds the four canonical add-flow stage labels in order.
// Matches Plan 27's rail canon — Ground / Grow / Conflicts stages still exist
// as steps in the splicer but do not appear as rail tabs; their rail index is
// StageIdxOffRail and the base Stage.renderFrame skips the rail render when
// the index is negative.
//
//	選 えらぶ    Select    — pick the agent
//	枝 えだ      Branches  — abilities (skills/workflows/protocols/sensors/routines)
//	観 みる      Observe   — one last look before the write
//	結 むすぶ    Yield     — completion card
var StageLabels = []initflow.StageLabel{
	{Kanji: "選", Kana: "えらぶ", English: "SELECT"},
	{Kanji: "枝", Kana: "えだ", English: "BRANCHES"},
	{Kanji: "観", Kana: "みる", English: "OBSERVE"},
	{Kanji: "結", Kana: "むすぶ", English: "YIELD"},
}

// AgentOption is the per-row shape consumed by SelectStage. Deliberately
// minimal — the stage needs a machine name, a display label, a description,
// and a flag indicating whether the agent type is already installed (renders
// as an "(installed)" suffix).
type AgentOption struct {
	Name        string // machine identifier returned verbatim in Result
	DisplayName string // human-readable label shown in the row
	Description string // one-line caption rendered muted after the name
	Installed   bool   // true when cfg.Agents[name] exists
}

// BranchesResult is the advance-payload returned from BranchesStage.Result()
// on Enter. Slices preserve catalog iteration order (alphabetical per
// catalog.loadItems). Required items are always present. Mirrors
// initflow.BranchesResult shape so the action closure can read either path
// with a type switch.
type BranchesResult struct {
	Skills    []string
	Workflows []string
	Protocols []string
	Sensors   []string
	Routines  []string
}

// Total returns the sum of selection counts across the five categories. Used
// by Observe for the CTA's "Graft ~N items" line.
func (r BranchesResult) Total() int {
	return len(r.Skills) + len(r.Workflows) + len(r.Protocols) +
		len(r.Sensors) + len(r.Routines)
}

// Outcome is the cross-stage scratchpad populated by the Grow action and
// consumed by Yield + the post-harness cleanup in cmd/add.go. Kept
// in addflow (not cmd/) so test helpers can stamp synthetic outcomes without
// back-importing cmd.
//
// Field semantics:
//
//   - Ran            — true once the Grow action body executed (even on error).
//   - AgentDef       — resolved catalog AgentDef (nil on unknown-agent error).
//   - Workspace      — normalised workspace path actually written.
//   - NewAgent       — true for the new-agent branch, false for add-items.
//   - TotalSelected  — BranchesResult.Total() captured at write time.
//   - SpinnerErr     — non-nil when the write pipeline failed; routed to a
//     tui.Warning by the caller and short-circuits the Yield render.
type Outcome struct {
	AgentDef      *catalog.AgentDef
	Workspace     string
	NewAgent      bool
	TotalSelected int
	SpinnerErr    error
	Ran           bool
}
