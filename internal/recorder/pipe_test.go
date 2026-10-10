//go:build unix

package recorder

// bonsai hook start whose stdout's reader is gone before it prints: the write fails and it exits 0, its record
// written, instead of dying by SIGPIPE (signal 13). The hook runs in a child process (this test binary again), since
// the signal is a process's. Windows has no SIGPIPE (a write to a closed pipe is an error there, which the hook passes
// over already), so this runs on Linux and macOS only.

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"testing"
)

const pipeChild = "BONSAI_TEST_PIPE_PROJECT"

func TestStartSurvivesAClosedPipe(t *testing.T) {
	if root := os.Getenv(pipeChild); root != "" {
		sess := "0f1e2d3c-aaaa-4bbb-8ccc-00000000000b"
		o := opts(root, `{"session_id":"`+sess+`","hook_event_name":"SessionStart","source":"startup","cwd":"`+root+`"}`, nil)
		os.Exit(Start(o, os.Stdout))
	}
	_, root := project(t)
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	_ = r.Close() // the reader is gone before the hook prints
	cmd := exec.Command(os.Args[0], "-test.run=^TestStartSurvivesAClosedPipe$")
	cmd.Env = append(os.Environ(), pipeChild+"="+root)
	cmd.Stdout = w
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err = cmd.Run()
	_ = w.Close()
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		if ws, ok := ee.Sys().(syscall.WaitStatus); ok && ws.Signaled() {
			t.Fatalf("hook start died by signal %v, stderr %q", ws.Signal(), stderr.String())
		}
	}
	if err != nil {
		t.Fatalf("hook start: %v, stderr %q", err, stderr.String())
	}
	got := records(t, root, "0f1e2d3c-aaaa-4bbb-8ccc-00000000000b")
	if len(got) != 1 || got[0].Event != "session_start" || !strings.HasPrefix(got[0].At, "20") {
		t.Errorf("the record: %+v", got)
	}
}
