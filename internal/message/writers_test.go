package message

// Two writers at once (design/plan-5.md, note 5.5.10, 7): two processes, the test binary started again as each
// (TestMain), write briefs to one recipient's folder at the same moment, each many times over; every message is a
// file of its own, none overwritten, every one reads back, and the log holds one message record for each. And the
// collisions a random id could meet: an id another message has is never used again, and a move onto a name that is
// taken leaves both files as they were.

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/LastStep/Bonsai/internal/workspace"
)

// writerChildEnv, when set, makes the test binary one of TestTwoWritersAtOnce's processes. Its value is
// "<start file>|<main>|<child>|<messages>": the child waits for the start file, writes its briefs, and prints their
// ids, one a line.
const writerChildEnv = "BONSAI_TEST_MESSAGE_WRITER"

func TestMain(m *testing.M) {
	if spec := os.Getenv(writerChildEnv); spec != "" {
		os.Exit(writerChild(spec))
	}
	os.Exit(m.Run())
}

func writerChild(spec string) int {
	parts := strings.Split(spec, "|")
	if len(parts) != 4 {
		fmt.Fprintln(os.Stderr, "bad spec", spec)
		return 3
	}
	start, main := parts[0], parts[1]
	child, _ := strconv.Atoi(parts[2])
	count, _ := strconv.Atoi(parts[3])
	deadline := time.Now().Add(30 * time.Second)
	for {
		if _, err := os.Stat(start); err == nil {
			break
		}
		if time.Now().After(deadline) {
			fmt.Fprintln(os.Stderr, "the start file never came")
			return 3
		}
		time.Sleep(time.Millisecond)
	}
	p := Place{Workspace: testWorkspace, Local: workspace.Local{Root: main, Main: main, Branch: "main"},
		Getenv: func(string) string { return "" }}
	for i := 0; i < count; i++ {
		w, e := Write(p, Writing{From: "orchestrator", To: "builder@T-0901", Kind: "brief", Task: "T-0901",
			Body: fmt.Sprintf("Brief %d of writer %d.", i, child), WorkDir: main}, DefaultRules())
		if e != nil {
			fmt.Fprintln(os.Stderr, e)
			return 1
		}
		fmt.Println(w.Message.ID)
	}
	return 0
}

func TestTwoWritersAtOnce(t *testing.T) {
	const writers, each = 2, 25
	main := t.TempDir()
	t.Setenv(workspace.HomeEnv, t.TempDir())
	start := filepath.Join(t.TempDir(), "start")
	cmds := make([]*exec.Cmd, writers)
	outs := make([]*bytes.Buffer, writers)
	errs := make([]*bytes.Buffer, writers)
	for i := range cmds {
		cmd := exec.Command(os.Args[0], "-test.run=^$")
		cmd.Env = append(os.Environ(), fmt.Sprintf("%s=%s|%s|%d|%d", writerChildEnv, start, main, i, each))
		outs[i], errs[i] = &bytes.Buffer{}, &bytes.Buffer{}
		cmd.Stdout, cmd.Stderr = outs[i], errs[i]
		if err := cmd.Start(); err != nil {
			t.Fatal(err)
		}
		cmds[i] = cmd
	}
	if err := os.WriteFile(start, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	ids := map[string]bool{}
	for i, cmd := range cmds {
		if err := cmd.Wait(); err != nil {
			t.Fatalf("writer %d: %v: %s", i, err, errs[i])
		}
		for _, id := range strings.Fields(outs[i].String()) {
			if ids[id] {
				t.Errorf("the id %s was printed twice", id)
			}
			ids[id] = true
		}
	}
	if len(ids) != writers*each {
		t.Fatalf("%d ids, want %d", len(ids), writers*each)
	}
	files, err := os.ReadDir(filepath.Join(main, ".bonsai", "local", "messages", "builder-T-0901"))
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != writers*each {
		t.Errorf("%d files in the recipient's folder, want %d (no staged file left)", len(files), writers*each)
	}
	st, err := Read(main)
	if err != nil {
		t.Fatal(err)
	}
	if len(st.Messages) != writers*each || len(st.Malformed) != 0 {
		t.Fatalf("read back %d, malformed %+v", len(st.Messages), st.Malformed)
	}
	bodies := map[string]bool{}
	for _, e := range st.Messages {
		if !ids[e.Msg.ID] || bodies[e.Msg.Body] {
			t.Errorf("%s: not printed, or its body twice: %q", e.Msg.ID, e.Msg.Body)
		}
		bodies[e.Msg.Body] = true
	}
	f := &fixture{t: t, main: main}
	recs := f.logRecords()
	if len(recs) != writers*each {
		t.Fatalf("%d message records, want %d", len(recs), writers*each)
	}
	for _, r := range recs {
		if !ids[str(r.Target)] {
			t.Errorf("a record of no message: %s", str(r.Target))
		}
		delete(ids, str(r.Target))
	}
	if len(ids) != 0 {
		t.Errorf("messages with no record: %v", ids)
	}
}

// An id another message has, in any folder, is never used again: the writer draws a new one, and after three in a
// row it writes nothing. A move onto a taken name leaves both files as they were and no staged file behind.
func TestIDCollisions(t *testing.T) {
	f := newFixture(t)
	defer func(r func([]byte) (int, error)) { random = r }(random)
	random = func(b []byte) (int, error) { copy(b, []byte{0xde, 0xad, 0xbe, 0xef}); return len(b), nil }
	first := f.write(Writing{From: "orchestrator", To: "builder@T-0901", Kind: "brief"}, DefaultRules())
	if !strings.HasSuffix(first.Message.ID, "-deadbeef") {
		t.Fatalf("the id: %s", first.Message.ID)
	}
	// Each later draw the same: to another folder, and on another second, the id's date is the same day's.
	f.refused(Writing{From: "orchestrator", To: "verifier", Kind: "brief"}, DefaultRules(), "write-failed")
	draws := 0
	random = func(b []byte) (int, error) {
		draws++
		if draws < 3 {
			copy(b, []byte{0xde, 0xad, 0xbe, 0xef})
		} else {
			copy(b, []byte{0x00, 0x00, 0x00, 0x01})
		}
		return len(b), nil
	}
	if w := f.write(Writing{From: "orchestrator", To: "verifier", Kind: "brief"}, DefaultRules()); !strings.HasSuffix(w.Message.ID, "-00000001") || draws != 3 {
		t.Errorf("after two taken ids: %s, %d draws", w.Message.ID, draws)
	}
	// A move onto a name a file has (another writer's, between its read and its move).
	dir := filepath.Join(f.main, ".bonsai", "local", "messages", "builder-T-0901")
	before, _ := os.ReadFile(filepath.Join(dir, first.Message.ID+".md"))
	if err := stage(dir, first.Message.ID, []byte("another writer's bytes\n")); err != errTaken {
		t.Fatalf("staged over a taken name: %v", err)
	}
	after, _ := os.ReadFile(filepath.Join(dir, first.Message.ID+".md"))
	entries, _ := os.ReadDir(dir)
	if string(after) != string(before) || len(entries) != 1 {
		t.Errorf("the taken file changed, or a staged file is left: %d files", len(entries))
	}
}
