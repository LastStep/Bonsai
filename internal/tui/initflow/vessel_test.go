package initflow

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// newTestVessel constructs a VesselStage with a minimal StageContext suitable
// for unit tests. The project path + version are not read by Vessel during
// its interaction — only the stage index + label matter — so they're stamped
// with benign defaults.
func newTestVessel() *VesselStage {
	return NewVesselStage(StageContext{
		Version:      "test",
		ProjectDir:   "/tmp/vessel-test",
		StationDir:   "station/",
		AgentDisplay: "Tech Lead",
		StartedAt:    time.Now(),
	})
}

// pressTea helper routes a specific tea.KeyType through Update.
func pressTea(s *VesselStage, k tea.KeyType) {
	m, _ := s.Update(tea.KeyMsg{Type: k})
	if vs, ok := m.(*VesselStage); ok {
		*s = *vs
	}
}

// TestVessel_FocusCycling_Tab verifies Tab advances focus through the three
// inputs in order.
func TestVessel_FocusCycling_Tab(t *testing.T) {
	v := newTestVessel()

	if v.focus != vesselIdxName {
		t.Fatalf("initial focus = %d, want %d", v.focus, vesselIdxName)
	}
	pressTea(v, tea.KeyTab)
	if v.focus != vesselIdxDescription {
		t.Fatalf("after Tab focus = %d, want %d", v.focus, vesselIdxDescription)
	}
	pressTea(v, tea.KeyTab)
	if v.focus != vesselIdxStation {
		t.Fatalf("after Tab×2 focus = %d, want %d", v.focus, vesselIdxStation)
	}
	pressTea(v, tea.KeyTab)
	if v.focus != vesselIdxName {
		t.Fatalf("after Tab×3 focus = %d (should wrap to %d)", v.focus, vesselIdxName)
	}
}

// TestVessel_FocusCycling_ShiftTab verifies Shift-Tab from the middle field
// moves focus up, and on the first field is a no-op (propagates to harness).
func TestVessel_FocusCycling_ShiftTab(t *testing.T) {
	v := newTestVessel()

	// Shift-Tab on field 0 should be a no-op — focus stays put.
	pressTea(v, tea.KeyShiftTab)
	if v.focus != vesselIdxName {
		t.Fatalf("Shift-Tab on first field moved focus to %d (expected no-op)", v.focus)
	}
	// Move to description, then Shift-Tab back to name.
	pressTea(v, tea.KeyTab)
	pressTea(v, tea.KeyShiftTab)
	if v.focus != vesselIdxName {
		t.Fatalf("Shift-Tab from Description focus = %d, want %d", v.focus, vesselIdxName)
	}
}

// TestVessel_EnterAdvancesFocus verifies ↵ on a non-last field moves focus
// forward rather than completing the stage.
func TestVessel_EnterAdvancesFocus(t *testing.T) {
	v := newTestVessel()
	v.inputs[vesselIdxName].SetValue("proj")

	pressTea(v, tea.KeyEnter)
	if v.focus != vesselIdxDescription {
		t.Fatalf("Enter on NAME advanced focus to %d, want %d", v.focus, vesselIdxDescription)
	}
	if v.done {
		t.Fatalf("Enter on NAME completed the stage — should only cycle focus")
	}
}

// TestVessel_RequiredEmptyBlocksSubmit verifies submit is blocked when NAME
// is empty and the error flag is surfaced.
func TestVessel_RequiredEmptyBlocksSubmit(t *testing.T) {
	v := newTestVessel()
	// Jump directly to station focus and press Enter — NAME is still empty.
	v.focusAt(vesselIdxStation)
	pressTea(v, tea.KeyEnter)

	if v.done {
		t.Fatalf("submit succeeded despite empty required NAME field")
	}
	if !v.showErrors {
		t.Fatalf("showErrors flag not set after failed submit")
	}
}

// TestVessel_StationSlashRejected verifies STATION value "/" is rejected
// and the stage does not complete.
func TestVessel_StationSlashRejected(t *testing.T) {
	v := newTestVessel()
	v.inputs[vesselIdxName].SetValue("proj")
	v.inputs[vesselIdxStation].SetValue("/")
	v.focusAt(vesselIdxStation)

	pressTea(v, tea.KeyEnter)

	if v.done {
		t.Fatalf("submit succeeded despite STATION = \"/\"")
	}
}

// TestVessel_ResultShape verifies Result() returns the expected
// map[string]string with keys "name", "description", "station".
func TestVessel_ResultShape(t *testing.T) {
	v := newTestVessel()
	v.inputs[vesselIdxName].SetValue("voyager-api")
	v.inputs[vesselIdxDescription].SetValue("Internal voyager service")
	v.inputs[vesselIdxStation].SetValue("workspace")

	res := v.Result()
	m, ok := res.(map[string]string)
	if !ok {
		t.Fatalf("Result() type = %T, want map[string]string", res)
	}
	if m["name"] != "voyager-api" {
		t.Fatalf("name = %q, want %q", m["name"], "voyager-api")
	}
	if m["description"] != "Internal voyager service" {
		t.Fatalf("description = %q, want %q", m["description"], "Internal voyager service")
	}
	// Station gets a trailing slash appended.
	if m["station"] != "workspace/" {
		t.Fatalf("station = %q, want %q", m["station"], "workspace/")
	}
}

// TestVessel_ResultEmptyDescription verifies an empty DESCRIPTION input
// surfaces as an empty string in the result (not a missing key).
func TestVessel_ResultEmptyDescription(t *testing.T) {
	v := newTestVessel()
	v.inputs[vesselIdxName].SetValue("proj")
	// Description left empty.

	res := v.Result().(map[string]string)
	desc, present := res["description"]
	if !present {
		t.Fatalf("description key missing from Result")
	}
	if desc != "" {
		t.Fatalf("description = %q, want empty string", desc)
	}
}

// TestVessel_ResultStationDefault verifies an empty STATION input falls back
// to the default "station/" value.
func TestVessel_ResultStationDefault(t *testing.T) {
	v := newTestVessel()
	v.inputs[vesselIdxName].SetValue("proj")
	// Station left empty.

	res := v.Result().(map[string]string)
	if res["station"] != defaultStationDir {
		t.Fatalf("station default = %q, want %q", res["station"], defaultStationDir)
	}
}

// TestVessel_SubmitOnLastField verifies ↵ on STATION with a valid NAME
// completes the stage (done=true).
func TestVessel_SubmitOnLastField(t *testing.T) {
	v := newTestVessel()
	v.inputs[vesselIdxName].SetValue("proj")
	v.focusAt(vesselIdxStation)

	pressTea(v, tea.KeyEnter)

	if !v.done {
		t.Fatalf("done=false after valid submit; expected stage advance")
	}
}

// TestVessel_ResponsiveInputWidth verifies the input width shrinks on
// narrow terminals and caps at 60 on wide ones. Underline is pinned to
// inputW+4 so a regression here would break focus-rule alignment.
func TestVessel_ResponsiveInputWidth(t *testing.T) {
	cases := []struct {
		termW, wantW int
	}{
		{120, 60}, // ample — cap
		{100, 60}, // still at cap (100-20-4 = 76 → capped to 60)
		{80, 56},  // 80-20-4 = 56
		{70, 46},  // 70-20-4 = 46
		{50, 30},  // floor
	}
	for _, c := range cases {
		v := newTestVessel()
		v.width = c.termW
		v.height = 30
		// Render body to trigger the inputW math (mutates inputs[].Width).
		_ = v.renderBody()
		if v.inputs[vesselIdxName].Width != c.wantW {
			t.Errorf("termW=%d: inputs[0].Width = %d, want %d",
				c.termW, v.inputs[vesselIdxName].Width, c.wantW)
		}
	}
}

// TestVessel_MinSizeFloor verifies <floor terminals route to the
// "please enlarge" panel rather than rendering a broken frame.
func TestVessel_MinSizeFloor(t *testing.T) {
	v := newTestVessel()
	v.width = 60
	v.height = 16
	out := v.View()
	if !contains(out, "please enlarge") {
		t.Errorf("min-size render missing floor panel; got:\n%s", out)
	}
}

// contains is a tiny helper so the test file stays import-light.
func contains(haystack, needle string) bool {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}

// TestVessel_RejectsAbsoluteStation verifies an absolute STATION input is
// rejected by validate(). Defence against accidental writes outside the
// project root (Plan 29 §H).
func TestVessel_RejectsAbsoluteStation(t *testing.T) {
	v := newTestVessel()
	v.inputs[vesselIdxName].SetValue("proj")
	v.inputs[vesselIdxStation].SetValue("/etc/foo")
	if v.validate() {
		t.Fatal("absolute STATION input should fail validate()")
	}
}

// TestVessel_RejectsParentEscapeStation verifies a STATION input with a
// ".." segment is rejected by validate().
func TestVessel_RejectsParentEscapeStation(t *testing.T) {
	v := newTestVessel()
	v.inputs[vesselIdxName].SetValue("proj")
	v.inputs[vesselIdxStation].SetValue("../bar/")
	if v.validate() {
		t.Fatal("parent-escape STATION input should fail validate()")
	}
}

// TestVessel_AcceptsCleanRelative is the positive companion to the
// rejects-* tests above. Verifies that clean relative STATION inputs
// (including ones that Clean reduces to a safe nested path) pass validate.
func TestVessel_AcceptsCleanRelative(t *testing.T) {
	cases := []string{"./foo", "foo/../bar"}
	for _, in := range cases {
		v := newTestVessel()
		v.inputs[vesselIdxName].SetValue("proj")
		v.inputs[vesselIdxStation].SetValue(in)
		if !v.validate() {
			t.Errorf("STATION = %q should pass validate()", in)
		}
	}
}
