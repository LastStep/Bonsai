package updateflow

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/LastStep/Bonsai/internal/generate"
	"github.com/LastStep/Bonsai/internal/tui/hints"
	"github.com/LastStep/Bonsai/internal/tui/initflow"
)

// newYieldStage builds a YieldStage with the given inputs and a 120x40
// viewport, ready to render.
func newYieldStage(inputs YieldInputs) *YieldStage {
	s := NewYieldStage(initflow.StageContext{StartedAt: time.Now()}, inputs)
	s.SetSize(120, 40)
	return s
}

// TestYield_UpToDatePanelWhenNoDiscoveries — zero changes + zero
// conflicts + no config mutation → the "Up to date" panel renders with
// the legacy string contract ("Workspace is in sync with the catalog.").
func TestYield_UpToDatePanelWhenNoDiscoveries(t *testing.T) {
	s := newYieldStage(YieldInputs{
		WriteResult:   &generate.WriteResult{},
		ConfigChanged: false,
	})
	view := s.View()
	if !strings.Contains(view, "Workspace is in sync with the catalog.") {
		t.Fatalf("up-to-date panel missing legacy string; got:\n%s", view)
	}
	if !strings.Contains(view, "UP TO DATE") {
		t.Fatalf("up-to-date panel should show UP TO DATE heading; got:\n%s", view)
	}
}

// TestYield_SyncedPanelWithCounts — at least one WriteResult change
// routes to the synced variant with SUMMARY counts.
func TestYield_SyncedPanelWithCounts(t *testing.T) {
	wr := &generate.WriteResult{}
	wr.Add(generate.FileResult{RelPath: "station/foo.md", Action: generate.ActionCreated, Source: "generated:test"})
	wr.Add(generate.FileResult{RelPath: "station/bar.md", Action: generate.ActionUpdated, Source: "generated:test"})

	s := newYieldStage(YieldInputs{
		WriteResult:   wr,
		ConfigChanged: false,
	})
	view := s.View()
	if !strings.Contains(view, "SYNCED") {
		t.Fatalf("synced panel should show SYNCED; got:\n%s", view)
	}
	if !strings.Contains(view, "SUMMARY") {
		t.Fatalf("synced panel should show SUMMARY header; got:\n%s", view)
	}
	if !strings.Contains(view, "CREATED") || !strings.Contains(view, "UPDATED") {
		t.Fatalf("synced panel should show CREATED/UPDATED rows; got:\n%s", view)
	}
}

// TestSync_ErrorAggregatesSurface — a non-nil SyncErr routes to the
// error variant and renders the error message.
func TestSync_ErrorAggregatesSurface(t *testing.T) {
	s := newYieldStage(YieldInputs{
		WriteResult: &generate.WriteResult{},
		SyncErr:     errors.New("boom: pipeline failed"),
	})
	view := s.View()
	if !strings.Contains(view, "SYNC ERROR") {
		t.Fatalf("error variant should show SYNC ERROR; got:\n%s", view)
	}
	if !strings.Contains(view, "boom: pipeline failed") {
		t.Fatalf("error variant should surface the error text; got:\n%s", view)
	}
}

// TestYield_HintBlockRendersWhenPresent — a non-zero hint block shows
// the NEXT STEPS section inside the Yield body.
func TestYield_HintBlockRendersWhenPresent(t *testing.T) {
	s := newYieldStage(YieldInputs{
		WriteResult: &generate.WriteResult{},
		HintBlock: hints.Block{
			NextCLI: []string{"bonsai list"},
		},
	})
	view := s.View()
	if !strings.Contains(view, "NEXT STEPS") {
		t.Fatalf("yield should render hint block when present; got:\n%s", view)
	}
}

// TestYield_EmptyHintBlockRendersNothingExtra — an IsZero hint block
// must not produce a NEXT STEPS header (safe for callers lacking a
// hints source).
func TestYield_EmptyHintBlockRendersNothingExtra(t *testing.T) {
	s := newYieldStage(YieldInputs{
		WriteResult: &generate.WriteResult{},
		HintBlock:   hints.Block{},
	})
	view := s.View()
	if strings.Contains(view, "NEXT STEPS") {
		t.Fatalf("zero block should NOT render NEXT STEPS; got:\n%s", view)
	}
}
