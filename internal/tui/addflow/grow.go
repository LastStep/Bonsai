package addflow

import (
	"github.com/LastStep/Bonsai/internal/tui/initflow"
)

// growLabel is the kanji/kana/English triple shown in the Grow stage's body
// title. Plan 27 shrunk the rail to four visible stages, so Grow no longer
// owns a rail tab — its rail index is StageIdxOffRail and the rail row is
// suppressed. Retained here (rather than pulled from StageLabels) so the
// body title + spinner still read "育 GROWING" unchanged.
var growLabel = initflow.StageLabel{Kanji: "育", Kana: "そだつ", English: "GROW"}

// NewGrowStage constructs the Grow stage (育 GROW). Renders off-rail —
// the visible rail stays anchored on OBSERVE while Grow runs so there is
// no rail churn between Observe and the spinner.
//
// Thin wrapper around initflow.NewGenerateStage — installs the add-flow
// rail labels on the embedded Stage and overrides the body title kanji
// from 生 (init's Generate) to 育 (Grow) so the visual language matches
// the bonsai-raising metaphor of the add flow.
//
// action is the same GenerateAction contract as the init flow — a
// one-shot func() error that runs the write pipeline and returns nil on
// success. The underlying min-hold + animation behaviour is unchanged.
func NewGrowStage(ctx initflow.StageContext, action initflow.GenerateAction) *initflow.GenerateStage {
	g := initflow.NewGenerateStage(ctx, action)
	g.SetRailLabels(StageLabels)
	g.SetRailIndex(StageIdxOffRail)
	g.SetLabel(growLabel)
	g.SetBodyTitle(growLabel.Kanji, "GROWING")
	g.SetRailHidden(true)
	// Plan 27 PR2 §C7 — Grow renders chromeless. No header, no enso rail, no
	// footer. The bonsai spinner sits centred in the AltScreen with an
	// inline key-hint row beneath it.
	g.SetBodyOnly(true)
	return g
}
