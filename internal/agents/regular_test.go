package agents

// VG's item F (15 Oct), its reproducer a test: a FIFO, a device, a folder or a link planted where Bonsai keeps an
// agent's files is never opened by any reader here (the guard's, the stop gate's, bonsai agents', bonsai messages' and
// status's), each answering at once, skipped and named, never waiting on it. VG's item I's other half: MoveAside keeps
// what was there under a name of its own.

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/message"
)

// within runs f, failing the test when it has not returned in a few seconds: a reader waiting on a planted file.
func within(t *testing.T, what string, f func()) {
	t.Helper()
	done := make(chan struct{})
	go func() {
		defer close(done)
		f()
	}()
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatalf("%s did not return: a reader waits on a planted file", what)
	}
}

// plants are the things planted where a regular file belongs, each by a function that makes it at a path (false
// when this system cannot make it).
var plants = []struct {
	name string
	make func(t *testing.T, path string) bool
}{
	{"a folder", func(t *testing.T, p string) bool { return os.MkdirAll(p, 0o755) == nil }},
	{"a FIFO", func(t *testing.T, p string) bool { return mkfifo(p) == nil }},
	{"a link to a device", func(t *testing.T, p string) bool {
		return runtime.GOOS != "windows" && os.Symlink("/dev/zero", p) == nil
	}},
	{"a link to a FIFO", func(t *testing.T, p string) bool {
		return mkfifo(p+"-fifo") == nil && os.Symlink(filepath.Base(p)+"-fifo", p) == nil
	}},
}

func TestReadersOpenRegularFilesOnly(t *testing.T) {
	for _, pl := range plants {
		t.Run(pl.name, func(t *testing.T) {
			s := store(t)
			if _, err := s.Create(header()); err != nil {
				t.Fatal(err)
			}
			dir := s.SessionDir(sess)
			calls := filepath.Join(dir, agent+".calls")
			if !pl.make(t, calls) {
				t.Skip("this system cannot make " + pl.name + " (Windows: no FIFO or device in its file system, and a link needs a right a test does not have)")
			}
			// The calls file planted: every reader answers at once, and no write waits on it.
			within(t, "the readers of a planted calls file", func() {
				if _, err := s.Calls(sess, agent); !errors.Is(err, ErrNotRegular) {
					t.Errorf("Calls: %v", err)
				}
				if _, _, ok := s.LastCall(sess, agent); ok {
					t.Error("LastCall read a planted file")
				}
				if _, err := s.Read(sess, agent); err == nil {
					t.Error("Read took a state whose calls file is planted")
				}
				if _, err := s.Recorded(sess, agent); err == nil {
					t.Error("Recorded read a planted calls file")
				}
				if _, err := s.Call(sess, agent, t0, 0); !errors.Is(err, ErrNotRegular) {
					t.Errorf("Call: %v", err)
				}
				v, err := s.View(&message.Store{})
				if err != nil || len(v.Skipped) != 1 || !strings.Contains(v.Skipped[0].Reason, "its calls file is not a regular file") {
					t.Errorf("View: %+v %v", v, err)
				}
				s.LastLine(sess)
			})
			// The state planted: the same.
			state := filepath.Join(dir, "a2.state")
			if !pl.make(t, state) {
				t.Fatal("the plant worked once")
			}
			within(t, "the readers of a planted state", func() {
				if _, err := s.Read(sess, "a2"); err == nil || errors.Is(err, ErrNoState) {
					t.Errorf("Read: %v", err)
				}
				if err := s.Append(sess, "a2", Event{E: EventSpent, At: StampMilli(t0)}); err == nil {
					t.Error("Append wrote into a planted state")
				}
				v, err := s.View(&message.Store{})
				named := false
				for _, sk := range v.Skipped {
					named = named || strings.HasSuffix(sk.Path, "/a2.state") && strings.Contains(sk.Reason, "not a regular file, so it is not opened")
				}
				if err != nil || !named {
					t.Errorf("View: %+v %v", v, err)
				}
			})
			// A tie file, a claim and the spawn calls planted: none is read, none waits.
			tie := filepath.Join(s.Dir(), TiesFolder, "M-2026-10-15-0000000a")
			claim := filepath.Join(dir, ClaimsFolder, "toolu_01X")
			spawns := filepath.Join(dir, spawnsFile)
			for _, p := range []string{tie, claim, spawns} {
				if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
					t.Fatal(err)
				}
				if !pl.make(t, p) {
					t.Fatal("the plant worked once")
				}
			}
			within(t, "the readers of planted ties, claims and spawn calls", func() {
				if _, _, ok := s.TiedTo("M-2026-10-15-0000000a"); ok {
					t.Error("TiedTo read a planted tie")
				}
				if who := s.TiedBy("M-2026-10-15-0000000a"); who != "" {
					t.Errorf("TiedBy %q", who)
				}
				if who := s.ClaimHolder(sess, "toolu_01X"); who != "" {
					t.Errorf("ClaimHolder %q", who)
				}
				if _, err := s.Spawns(sess); !errors.Is(err, ErrNotRegular) {
					t.Errorf("Spawns: %v", err)
				}
				if err := s.AddSpawn(sess, Spawn{Use: "toolu_01Y", Tool: "Agent"}); !errors.Is(err, ErrNotRegular) {
					t.Errorf("AddSpawn: %v", err)
				}
				th := &message.Thread{Entries: []message.Entry{{Msg: &format.Message{ID: "M-2026-10-15-0000000a", Kind: format.KindBrief}}}}
				th.State = message.Open
				s.Thread(th, 5, t0)
				if th.Recipient != nil {
					t.Errorf("a thread read a planted tie: %+v", th.Recipient)
				}
				_, _ = s.Ties()
				_, _ = s.SessionFiles(sess)
			})
		})
	}
}

// A state that cannot be read is moved aside with its calls file, kept, under <agent>-aside-<time>; bonsai agents
// names it in skipped; a state that reads is never moved.
func TestMoveAside(t *testing.T) {
	s := store(t)
	if _, err := s.Create(header()); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Call(sess, agent, t0, 0); err != nil {
		t.Fatal(err)
	}
	if aside, err := s.MoveAside(sess, agent, t0); aside != "" || err != nil {
		t.Fatalf("a state that reads was moved: %q %v", aside, err)
	}
	path := filepath.Join(s.SessionDir(sess), agent+".state")
	if err := os.WriteFile(path, []byte("not a header\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	aside, err := s.MoveAside(sess, agent, t0)
	want := "agents/" + sess + "/" + agent + "-aside-20261015T030000000Z.state"
	if err != nil || aside != want {
		t.Fatalf("moved to %q %v, want %q", aside, err, want)
	}
	if _, err := s.Read(sess, agent); !errors.Is(err, ErrNoState) {
		t.Errorf("the state is still there: %v", err)
	}
	for _, f := range []string{agent + "-aside-20261015T030000000Z.state", agent + "-aside-20261015T030000000Z.calls"} {
		if _, err := os.Lstat(filepath.Join(s.SessionDir(sess), f)); err != nil {
			t.Errorf("%s not kept: %v", f, err)
		}
	}
	v, err := s.View(&message.Store{})
	if err != nil || len(v.Skipped) != 1 || !strings.Contains(v.Skipped[0].Reason, "the guard moved it aside") {
		t.Errorf("View: %+v %v", v, err)
	}
	// An id too long to take the mark is cut, so the name stays one a state file takes.
	long := strings.Repeat("a", 120)
	h := header()
	h.Agent = long
	if _, err := s.Create(h); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(s.SessionDir(sess), long+".state"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if aside, err := s.MoveAside(sess, long, t0); err != nil || len(filepath.Base(aside)) != 128+len(".state") {
		t.Errorf("a long id moved to %q %v", aside, err)
	}
}
