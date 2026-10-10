package status

import (
	"testing"

	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// Every kind needsOf gives is in NeedKinds, the three this build writes among them.
func TestNeedKindsHoldWhatNeedsOfGives(t *testing.T) {
	lock := &workspace.Lock{Packs: []workspace.LockedPack{{ID: "a"}, {ID: "b"}}}
	known := map[string]bool{}
	for _, k := range NeedKinds {
		known[k.Word] = true
		if k.Means == "" {
			t.Errorf("need kind %s has no meaning", k.Word)
		}
	}
	got := map[string]bool{}
	for _, installed := range []map[string]bool{nil, {"a": true, "b": false}} {
		for _, n := range needsOf(lock, installed) {
			kind := n.(schema.Object).String("kind")
			got[kind] = true
			if !known[kind] {
				t.Errorf("needsOf gives the kind %q, which NeedKinds lacks", kind)
			}
		}
	}
	for _, k := range []string{"pack", "plugin", "tool"} {
		if !got[k] {
			t.Errorf("needsOf never gave the kind %s", k)
		}
	}
}
