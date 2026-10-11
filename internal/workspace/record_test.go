package workspace

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// This machine's record of a checkout: written with the checkout's path and its id since a time; written again only
// when the id changes, the new id at the end and the first kept; a record Bonsai does not read is not written over
// (it would lose its first claim on the id), its error naming a person's step; every record in the home listed by its
// machine folder.
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
	if wrote, err := RecordCheckout(home, main, "ws-cccccccccccccccccccccccccc", at); wrote || err == nil ||
		!strings.Contains(err.Error(), "so it is not written over") || !strings.Contains(err.Error(), "next: a person checks") {
		t.Errorf("a record Bonsai does not read was written over: %v %v", wrote, err)
	}
	if raw, _ := os.ReadFile(filepath.Join(dir, MachineRecordFile)); string(raw) != "not json" {
		t.Errorf("the damaged record changed: %q", raw)
	}
	if MachineRecords(t.TempDir()) != nil {
		t.Error("an empty home lists records")
	}
}

// putRecord writes a record by hand into the home's machine folder key.
func putRecord(t *testing.T, home, key, body string) {
	t.Helper()
	dir := filepath.Join(home, "workspaces", key)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, MachineRecordFile), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// FirstClaim: among the records holding an id now, the earliest claim (the earliest since of the id in each) is first;
// a record that gave the id up no longer claims it; two claims in the same second name no first, and say both; a
// record that cannot be read names none, with an error; no claim names none.
func TestFirstClaim(t *testing.T) {
	const id = "ws-aaaaaaaaaaaaaaaaaaaaaaaaaa"
	rec := func(path string, ids ...string) string {
		var parts []string
		for i := 0; i+1 < len(ids); i += 2 {
			parts = append(parts, `{"id": "`+ids[i]+`", "since": "`+ids[i+1]+`"}`)
		}
		return `{"path": "` + path + `", "ids": [` + strings.Join(parts, ", ") + `]}`
	}
	home := t.TempDir()
	if first, ties, err := FirstClaim(home, id); first != nil || ties != nil || err != nil {
		t.Fatalf("an empty home: %+v %v %v", first, ties, err)
	}
	putRecord(t, home, "r-0000000000000001", rec("/p/main", id, "2026-10-01T09:00:00Z"))
	putRecord(t, home, "r-0000000000000002", rec("/p/later", id, "2026-10-02T09:00:00Z"))
	// Held the id earliest, then gave it up (init --new-id): no longer a claim.
	putRecord(t, home, "r-0000000000000003", rec("/p/copy", id, "2026-09-01T09:00:00Z", "ws-bbbbbbbbbbbbbbbbbbbbbbbbbb", "2026-09-02T09:00:00Z"))
	// Holds the id now, first claimed earlier, held another between: its claim is its earliest of the id.
	putRecord(t, home, "r-0000000000000004", rec("/p/back", id, "2026-10-03T09:00:00Z", "ws-cccccccccccccccccccccccccc", "2026-10-04T09:00:00Z", id, "2026-10-05T09:00:00Z"))
	// Another id altogether, and a machine folder with no record.
	putRecord(t, home, "r-0000000000000005", rec("/p/other", "ws-dddddddddddddddddddddddddd", "2026-08-01T09:00:00Z"))
	if err := os.MkdirAll(filepath.Join(home, "workspaces", "r-0000000000000006"), 0o755); err != nil {
		t.Fatal(err)
	}
	first, ties, err := FirstClaim(home, id)
	if err != nil || ties != nil || first == nil || first.Key != "r-0000000000000001" || first.Record.Path != "/p/main" ||
		!first.Since.Equal(time.Date(2026, 10, 1, 9, 0, 0, 0, time.UTC)) {
		t.Fatalf("first claim: %+v %+v %v", first, ties, err)
	}
	// A record's claim, read alone.
	m, _ := readMachineRecord(filepath.Join(home, "workspaces", "r-0000000000000004", MachineRecordFile))
	if at, ok := m.ClaimOf(id); !ok || !at.Equal(time.Date(2026, 10, 3, 9, 0, 0, 0, time.UTC)) {
		t.Errorf("a claim held, given up and held again: %v %v", at, ok)
	}
	if _, ok := m.ClaimOf("ws-cccccccccccccccccccccccccc"); ok {
		t.Error("an id given up still claimed")
	}
	// The same second as the first: a tie names no first, and both.
	putRecord(t, home, "r-0000000000000007", rec("/p/decoy", id, "2026-10-01T09:00:00Z"))
	first, ties, err = FirstClaim(home, id)
	if err != nil || first != nil || len(ties) != 2 || ties[0].Key != "r-0000000000000001" || ties[1].Key != "r-0000000000000007" {
		t.Fatalf("a tie: %+v %+v %v", first, ties, err)
	}
	// A record that cannot be read: none is named, whatever the others say.
	putRecord(t, home, "r-0000000000000008", "{")
	if first, ties, err = FirstClaim(home, id); first != nil || ties != nil || err == nil || !strings.Contains(err.Error(), "r-0000000000000008") {
		t.Fatalf("a record that cannot be read: %+v %+v %v", first, ties, err)
	}
}
