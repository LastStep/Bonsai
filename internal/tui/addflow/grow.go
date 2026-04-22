package addflow

import (
	"github.com/LastStep/Bonsai/internal/tui/initflow"
)

// NewGrowStage constructs the Grow stage (育 GROW) at rail position 4.
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
	g.SetRailIndex(StageIdxGrow)
	g.SetLabel(StageLabels[StageIdxGrow])
	g.SetBodyTitle(StageLabels[StageIdxGrow].Kanji, "GROWING")
	return g
}
