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
	"fmt"
	"strings"

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
// consumed by Yield + the post-harness cleanup in cmd/remove.go. Kept here
// (not in cmd/) so test helpers can stamp synthetic outcomes without back-
// importing cmd.
type Outcome struct {
	// Ran is true once the action closure executed (even on error).
	Ran bool
	// Err is non-nil when the write pipeline failed; routed to a warning by
	// the caller and short-circuits the Yield success render.
	Err error
	// AgentDisplay is the display name of the primary agent affected (agent
	// being removed, or the single target of an item removal). Empty when the
	// "all agents" option fired on item-remove.
	AgentDisplay string
	// ItemDisplay is the display name of the item being removed (item-remove
	// branch only). Empty on agent-remove.
	ItemDisplay string
	// ItemType is the singular item type label ("skill", "workflow", …).
	// Empty on agent-remove.
	ItemType string
	// RemovedCounts reports the per-category totals removed when the flow
	// completes. On agent-remove this is the agent's installed ability count;
	// on item-remove, the number of categories this item crossed (typically 1).
	RemovedCounts AbilityCounts
	// Targets is the number of agents affected — 1 for agent-remove,
	// 1..N for item-remove depending on the picker outcome.
	Targets int
}

// StaticPreview is a minimal non-TTY preview used by cmd/remove.go when
// stdout is not a terminal. Describes the target + prompts the user to
// re-run interactively. Plan 31 Phase E leaves a `--yes` flag for a future
// followup — today non-interactive removal refuses ambiguous confirmation.
type StaticPreview struct {
	// Title appears on the first line — e.g. "Remove Backend?" or
	// "Remove skill coding-standards?".
	Title string
	// Lines are rendered under the title as indented bullets.
	Lines []string
}

// RenderStatic formats a StaticPreview as plain text. Returns the rendered
// string (no ANSI) followed by a refusal message instructing the user to
// re-run from a terminal or pass --yes (which is a future followup — see
// Plan 31 Phase E). Callers should Println the result and return an error
// from the cobra command.
func RenderStatic(p StaticPreview) string {
	var b strings.Builder
	if p.Title != "" {
		b.WriteString(p.Title)
		b.WriteString("\n")
	}
	for _, line := range p.Lines {
		b.WriteString("  • ")
		b.WriteString(line)
		b.WriteString("\n")
	}
	b.WriteString("\n")
	b.WriteString("Non-interactive stdout detected. Re-run from a terminal, or\n")
	b.WriteString("add `--yes` for non-interactive removal (not yet implemented).\n")
	return b.String()
}

// StaticError returns a concise error string suitable for the cobra command
// to surface when the user tried to remove something in a non-TTY context.
// Kept separate from RenderStatic so callers can log the preview and then
// exit with the dedicated error message.
func StaticError() error {
	return fmt.Errorf("add --yes flag for non-interactive removal (not yet implemented)")
}
