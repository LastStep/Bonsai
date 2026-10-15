package hookin

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// The starting call, the type, the parent and the depth are read from Claude Code's meta file beside the subagent's
// transcript, as measured on 2.1.300; a file that is not there is fs.ErrNotExist; one that does not read, or names an
// id that is not one, an error of its own; an id that could leave the folder finds no file.
func TestSpawnInfo(t *testing.T) {
	dir := t.TempDir()
	session, agent := "393fe3e7-4666-4e2a-80e7-62cfe58f6a23", "ab8125621420b9524"
	transcript := filepath.ToSlash(filepath.Join(dir, session+".jsonl"))
	meta := filepath.Join(dir, session, "subagents", "agent-"+agent+".meta.json")
	if err := os.MkdirAll(filepath.Dir(meta), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := SpawnInfo(transcript, session, agent); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("no meta file: %v", err)
	}
	write := func(body string) {
		t.Helper()
		if err := os.WriteFile(meta, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	for _, c := range []struct {
		body string
		want SpawnMeta
	}{
		// The measured files: an Agent call's child (fixtures/wsl/transcripts' meta); a skill's fork from the session;
		// a subagent's own subagent.
		{`{"agentType":"helper","description":"Run echo and reply ok","toolUseId":"toolu_016Z7tzdB6g9U9Z1bWHtsAah",` +
			`"spawnDepth":1,"requestShape":"foreground","requestNonInteractive":true}`,
			SpawnMeta{Call: "toolu_016Z7tzdB6g9U9Z1bWHtsAah", Type: "helper", Depth: 1}},
		{`{"agentType":"builder","spawnDepth":1,"requestShape":"foreground","requestNonInteractive":true}`,
			SpawnMeta{Type: "builder", Depth: 1}},
		{`{"agentType":"auditor","description":"auditor","toolUseId":"toolu_011iTngvTvmAQfg95Qbt34wu",` +
			`"parentAgentId":"af006bfa747ee8f84","spawnDepth":2,"requestShape":"foreground","requestNonInteractive":true}`,
			SpawnMeta{Call: "toolu_011iTngvTvmAQfg95Qbt34wu", Type: "auditor", Parent: "af006bfa747ee8f84", Depth: 2}},
		{`{"agentType":"workflow-subagent","workflowPhase":"Run"}`, SpawnMeta{Type: "workflow-subagent"}},
	} {
		write(c.body)
		if got, err := SpawnInfo(transcript, session, agent); err != nil || got != c.want {
			t.Errorf("%s: got %+v %v", c.body, got, err)
		}
	}
	for _, body := range []string{`{"toolUseId":"a/b"}`, `{"agentType":"x","parentAgentId":"../x"}`, `not json`} {
		write(body)
		if got, err := SpawnInfo(transcript, session, agent); err == nil || errors.Is(err, fs.ErrNotExist) {
			t.Errorf("%s: read %+v %v", body, got, err)
		}
	}
	for _, c := range [][2]string{{"../x", agent}, {session, "../../etc"}, {"", agent}} {
		if _, err := SpawnInfo(transcript, c[0], c[1]); !errors.Is(err, fs.ErrNotExist) {
			t.Errorf("%v: %v", c, err)
		}
	}
}

// VG round 2's can-wait item: a FIFO at a meta file's path is named, never opened, so it stalls nothing (it stalled
// that agent's first call until the hook's timeout); a folder there is named too.
func TestSpawnInfoReadsRegularFilesOnly(t *testing.T) {
	dir := t.TempDir()
	session, agent := "393fe3e7-4666-4e2a-80e7-62cfe58f6a23", "ab8125621420b9524"
	transcript := filepath.ToSlash(filepath.Join(dir, session+".jsonl"))
	meta := filepath.Join(dir, session, "subagents", "agent-"+agent+".meta.json")
	if err := os.MkdirAll(meta, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := SpawnInfo(transcript, session, agent); err == nil || errors.Is(err, fs.ErrNotExist) {
		t.Errorf("a folder at the meta path: %v", err)
	}
	if err := os.Remove(meta); err != nil {
		t.Fatal(err)
	}
	if err := mkfifo(meta); err != nil {
		t.Skip("no FIFO on this system (Windows' file system holds none): " + err.Error())
	}
	done := make(chan error, 1)
	go func() {
		_, err := SpawnInfo(transcript, session, agent)
		done <- err
	}()
	select {
	case err := <-done:
		if err == nil || errors.Is(err, fs.ErrNotExist) {
			t.Errorf("a FIFO at the meta path: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("a FIFO at the meta path stalled the read")
	}
}
