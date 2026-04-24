package updateflow

import (
	"github.com/LastStep/Bonsai/internal/tui/initflow"
)

// syncLabel is the body title for the Sync stage — "同 SYNC" (動 can
// also read "DO", 同 was chosen for the "synchronise" connotation).
var syncLabel = initflow.StageLabel{Kanji: "同", Kana: "どう", English: "SYNC"}

// NewSyncStage wraps initflow.GenerateStage with the update-flow rail
// labels and body title. Renders on-rail at StageIdxSync (rail position
// 2) so the user sees the active SYNC checkpoint while the generator
// pipeline runs.
//
// action follows the standard GenerateAction contract — a one-shot
// func() error that performs the write pipeline. The update-flow caller
// passes a closure that applies user-selected discoveries, runs the
// AgentWorkspace / PathScopedRules / WorkflowSkills / SettingsJSON /
// WriteCatalogSnapshot pipeline, and returns errors.Join of any failures.
func NewSyncStage(ctx initflow.StageContext, action initflow.GenerateAction) *initflow.GenerateStage {
	g := initflow.NewGenerateStage(ctx, action)
	g.SetRailLabels(StageLabels)
	g.SetRailIndex(StageIdxSync)
	g.SetLabel(syncLabel)
	g.SetBodyTitle(syncLabel.Kanji, "SYNCING")
	return g
}
