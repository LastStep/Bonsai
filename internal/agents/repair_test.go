package agents

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

// VG round 2's M1: an agent's calls file made read-only, and its state made unreadable, are set back to the owner's
// read and write (on Windows, the read-only attribute cleared), and the count goes on where it was; Repair changes
// nothing that needs nothing, and reads no link.
func TestRepairSetsBackTheOwnersAccess(t *testing.T) {
	s := store(t)
	if _, err := s.Create(header()); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		if _, err := s.Call(sess, agent, t0, 0); err != nil {
			t.Fatal(err)
		}
	}
	if got := s.Repair(sess, agent); len(got) != 0 {
		t.Errorf("nothing needed setting back, yet: %v", got)
	}
	calls, state := s.callsPath(sess, agent), s.statePath(sess, agent)
	if err := os.Chmod(calls, 0o444); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(calls, 0o644) })
	if os.Geteuid() != 0 {
		if _, err := s.Call(sess, agent, t0, 0); err == nil {
			t.Fatal("a read-only calls file took a line")
		}
	}
	got := s.Repair(sess, agent)
	if len(got) != 1 || got[0] != "agents/"+sess+"/"+agent+".calls" {
		t.Errorf("set back %v", got)
	}
	if n, err := s.Call(sess, agent, t0, 0); err != nil || n != 3 {
		t.Errorf("after: %d %v", n, err)
	}
	if runtime.GOOS == "windows" {
		return // a mode never stops a read or a folder there: the read-only attribute is the one os.Chmod sets
	}
	if err := os.Chmod(state, 0); err != nil {
		t.Fatal(err)
	}
	folder := s.SessionDir(sess)
	if err := os.Chmod(folder, 0o500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(folder, 0o755); _ = os.Chmod(state, 0o644) })
	got = s.Repair(sess, agent)
	if len(got) != 2 || got[0] != "agents/"+sess || got[1] != "agents/"+sess+"/"+agent+".state" {
		t.Errorf("set back %v", got)
	}
	if st, err := s.Read(sess, agent); err != nil || st.Header.Agent != agent {
		t.Errorf("after: %v", err)
	}
	if fi, _ := os.Stat(folder); fi.Mode().Perm() != 0o700 {
		t.Errorf("the folder's mode %v: only the owner's bits are added", fi.Mode().Perm())
	}
	// A link in a file's place is not followed (it is moved aside, MoveAside).
	other := filepath.Join(t.TempDir(), "other")
	if err := os.WriteFile(other, []byte("x\n"), 0o400); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(calls); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(other, calls); err != nil {
		t.Skip("no symbolic link here: " + err.Error())
	}
	_ = s.Repair(sess, agent)
	if fi, _ := os.Stat(other); fi.Mode().Perm() != 0o400 {
		t.Errorf("the link was followed: %v", fi.Mode().Perm())
	}
}

// Asides counts an agent's states moved aside, its id cut as MoveAside cuts it, and no other agent's.
func TestAsides(t *testing.T) {
	s := store(t)
	long := ""
	for len(long) < 120 {
		long += "a1b2c3d4"
	}
	for _, a := range []string{agent, long, "other"} {
		h := header()
		h.Agent = a
		if _, err := s.Create(h); err != nil {
			t.Fatal(err)
		}
	}
	for i, a := range []string{agent, agent, long, "other"} {
		if err := os.WriteFile(s.statePath(sess, a), []byte("broken\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := s.MoveAside(sess, a, t0.Add(time.Duration(i)*time.Second)); err != nil {
			t.Fatal(err)
		}
	}
	if n := s.Asides(sess, agent); n != 2 {
		t.Errorf("%d asides of %s", n, agent)
	}
	if n := s.Asides(sess, long); n != 1 {
		t.Errorf("%d asides of the long id", n)
	}
	if n := s.Asides(sess, "nobody"); n != 0 {
		t.Errorf("%d asides of nobody", n)
	}
}
