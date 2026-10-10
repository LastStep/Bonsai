package recorder

// The cleaning at a session's end (internal/clean, step 5.2.6b): after the session_end line, and only then, in the
// main checkout's local/, the clean records naming the session's checkout.

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/LastStep/Bonsai/internal/clean"
	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/record"
	"github.com/LastStep/Bonsai/internal/testpack"
)

// oldLadderResult writes a done task and its ladder result, finished 20 days ago.
func oldLadderResult(t *testing.T, root string) string {
	t.Helper()
	write(t, root, "work/tasks/T-0904-x.md", "---\nformat: bonsai.task/1\nid: T-0904\ntitle: x\nstatus: done\nlane: null\n"+
		"done_when: []\ndepends_on: []\nblocked_by: null\ncreated: null\nstarted: null\nfinished: null\nlabels: {}\n---\n")
	at := time.Now().Add(-20 * 24 * time.Hour).UTC()
	task := "T-0904"
	l := &format.Ladder{Task: &task, Workspace: testID, Mode: "local", Started: at.Format(record.AtLayout),
		Finished: at.Format(record.AtLayout), Git: format.LadderGit{SHA: strings.Repeat("cd", 20), Branch: "main"},
		Requested: []int64{0}, Green: true, Rungs: []format.LadderRung{}, Skipped: []format.LadderSkip{}}
	b, err := l.Encode()
	if err != nil {
		t.Fatal(err)
	}
	write(t, root, ".bonsai/local/ladder/T-0904.json", string(b))
	p := filepath.Join(root, ".bonsai", "local", "ladder", "T-0904.json")
	if err := os.Chtimes(p, at, at); err != nil {
		t.Fatal(err)
	}
	return p
}

func endPayload(sess, cwd string) string {
	return `{"session_id":"` + sess + `","hook_event_name":"SessionEnd","reason":"prompt_input_exit","cwd":"` + filepath.ToSlash(cwd) + `"}`
}

func TestSessionEndCleans(t *testing.T) {
	_, root := project(t)
	result := oldLadderResult(t, root)
	sess := "0f1e2d3c-aaaa-4bbb-8ccc-000000000009"
	// Another event cleans nothing.
	Record(opts(root, `{"session_id":"`+sess+`","hook_event_name":"Stop"}`, nil))
	if _, err := os.Stat(result); err != nil {
		t.Fatalf("a Stop cleaned: %v", err)
	}
	Record(opts(root, endPayload(sess, root), nil))
	if _, err := os.Stat(result); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("the old result of a done task is still there: %v", err)
	}
	got := records(t, root, sess)
	if len(got) != 2 || got[1].Event != "session_end" {
		t.Fatalf("the session's records: %+v", got)
	}
	lf, err := record.ReadLog(filepath.Join(root, ".bonsai", "local", "log", "w-"+time.Now().UTC().Format("2006-01-02")+".ndjson"))
	if err != nil {
		t.Fatal(err)
	}
	if len(lf.Records) != 1 || lf.Records[0].Event != "clean" || s(lf.Records[0].Target) != ".bonsai/local/ladder/T-0904.json" ||
		s(lf.Records[0].Reason) != "generated.ladder.keep_days=7" || lf.Records[0].Session != nil || s(lf.Records[0].Checkout) != "proj" {
		t.Errorf("the clean record: %+v", lf.Records)
	}
}

// A session_end line that cannot be written cleans nothing: no clean record could be written either.
func TestNoCleaningWithoutTheEndLine(t *testing.T) {
	_, root := project(t)
	result := oldLadderResult(t, root)
	old := writeLog
	writeLog = func(string, *format.Log) (bool, error) { return false, errors.New("the disk is full") }
	t.Cleanup(func() { writeLog = old })
	Record(opts(root, endPayload("0f1e2d3c-aaaa-4bbb-8ccc-00000000000a", root), nil))
	if _, err := os.Stat(result); err != nil {
		t.Errorf("cleaned without its end line: %v", err)
	}
}

// In a worktree the cleaner is given the main checkout's local/ and the worktree as the records' checkout.
func TestWorktreeCleansMain(t *testing.T) {
	tmp, root := project(t)
	testpack.Git(t, root, "add", "-A")
	testpack.Git(t, root, "commit", "-q", "-m", "x")
	wt := filepath.Join(tmp, "wt")
	testpack.Git(t, root, "worktree", "add", "-q", "-b", "t0905", wt)
	var given *clean.Options
	old := cleanAtEnd
	cleanAtEnd = func(o clean.Options) *clean.Done { given = &o; return &clean.Done{} }
	t.Cleanup(func() { cleanAtEnd = old })
	Record(opts(wt, endPayload("wt-2", wt), nil))
	if given == nil {
		t.Fatal("no cleaning at the session's end")
	}
	main, _ := filepath.EvalSymlinks(root)
	if gm, _ := filepath.EvalSymlinks(given.Local.Main); gm != main || filepath.Base(given.Local.Root) != "wt" || given.Budget != 0 {
		t.Errorf("the cleaner was given %+v", given.Local)
	}
}
