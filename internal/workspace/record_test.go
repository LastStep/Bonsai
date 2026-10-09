package workspace

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// This machine's record of a checkout: written with the checkout's path and its id since a time; written again only
// when the id changes, the new id at the end and the first kept; a record Bonsai does not read is written again;
// every record in the home listed by its machine folder.
func TestMachineRecord(t *testing.T) {
	home, main := t.TempDir(), t.TempDir()
	if rec, err := LoadMachineRecord(home, main); rec != nil || err != nil {
		t.Fatalf("no record yet: %+v %v", rec, err)
	}
	at := time.Date(2026, 10, 9, 12, 0, 0, 0, time.UTC)
	if wrote, err := RecordCheckout(home, main, "ws-aaaaaaaaaaaaaaaaaaaaaaaaaa", at); !wrote || err != nil {
		t.Fatalf("first: %v %v", wrote, err)
	}
	if wrote, err := RecordCheckout(home, main, "ws-aaaaaaaaaaaaaaaaaaaaaaaaaa", at.Add(time.Hour)); wrote || err != nil {
		t.Errorf("the same id again wrote: %v %v", wrote, err)
	}
	if wrote, _ := RecordCheckout(home, main, "ws-bbbbbbbbbbbbbbbbbbbbbbbbbb", at.Add(time.Hour)); !wrote {
		t.Error("a new id was not recorded")
	}
	rec, err := LoadMachineRecord(home, main)
	real, _ := filepath.EvalSymlinks(main)
	if err != nil || rec.Path != filepath.ToSlash(real) || len(rec.IDs) != 2 || rec.IDs[0].Since != "2026-10-09T12:00:00Z" ||
		rec.Current() != "ws-bbbbbbbbbbbbbbbbbbbbbbbbbb" || rec.IDs[1].Since != "2026-10-09T13:00:00Z" {
		t.Fatalf("record %+v %v", rec, err)
	}
	dir, _ := MachineDir(home, main)
	raw, _ := os.ReadFile(filepath.Join(dir, MachineRecordFile))
	if strings.Contains(string(raw), `\`) || !strings.HasSuffix(string(raw), "}\n") || strings.Contains(string(raw), "\r") {
		t.Errorf("not forward slashes and LF:\n%s", raw)
	}
	keyed := MachineRecords(home)
	if len(keyed) != 1 || keyed[0].Key != filepath.Base(dir) || keyed[0].Record.Current() != rec.Current() {
		t.Errorf("records %+v", keyed)
	}
	if err := os.WriteFile(filepath.Join(dir, MachineRecordFile), []byte("not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadMachineRecord(home, main); err == nil {
		t.Error("a record Bonsai does not read was read")
	}
	if wrote, err := RecordCheckout(home, main, "ws-cccccccccccccccccccccccccc", at); !wrote || err != nil {
		t.Errorf("a record Bonsai does not read is written again: %v %v", wrote, err)
	}
	if MachineRecords(t.TempDir()) != nil {
		t.Error("an empty home lists records")
	}
}
