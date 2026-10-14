package hookin

import (
	"os"
	"path/filepath"
	"testing"
)

// The starting call is read from Claude Code's meta file beside the subagent's transcript, as measured on 2.1.300; a
// file that is not there, does not read, names no call, or an id that could leave the folder, gives none.
func TestSpawnCall(t *testing.T) {
	dir := t.TempDir()
	session, agent := "393fe3e7-4666-4e2a-80e7-62cfe58f6a23", "ab8125621420b9524"
	transcript := filepath.ToSlash(filepath.Join(dir, session+".jsonl"))
	meta := filepath.Join(dir, session, "subagents", "agent-"+agent+".meta.json")
	if err := os.MkdirAll(filepath.Dir(meta), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, _, err := SpawnCall(transcript, session, agent); err == nil {
		t.Error("no meta file: read")
	}
	// The measured file (fixtures/wsl/transcripts' meta), as Claude Code writes it.
	write := func(body string) {
		if err := os.WriteFile(meta, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write(`{"agentType":"helper","description":"Run echo and reply ok","toolUseId":"toolu_016Z7tzdB6g9U9Z1bWHtsAah",` +
		`"spawnDepth":1,"requestShape":"foreground","requestNonInteractive":true}`)
	use, typ, err := SpawnCall(transcript, session, agent)
	if err != nil || use != "toolu_016Z7tzdB6g9U9Z1bWHtsAah" || typ != "helper" {
		t.Errorf("got %q %q %v", use, typ, err)
	}
	for _, body := range []string{`{"agentType":"workflow-subagent","workflowPhase":"Run"}`, `{"toolUseId":"a/b"}`, `not json`} {
		write(body)
		if use, _, err := SpawnCall(transcript, session, agent); err == nil {
			t.Errorf("%s: read %q", body, use)
		}
	}
	for _, c := range [][2]string{{"../x", agent}, {session, "../../etc"}, {"", agent}} {
		if _, _, err := SpawnCall(transcript, c[0], c[1]); err == nil {
			t.Errorf("%v: read", c)
		}
	}
}
