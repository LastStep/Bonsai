package recorder

// The usage records (note 5.5.12, 4): at a subagent's SubagentStop and a session's SessionEnd, the token counts of the
// transcript the payload names, one record for each model, summed once per reply; or one unknown record with its
// reason, the hook exiting 0 in time.
//
// testdata/claude-2.1.300/ holds the transcripts measured on Claude Code 2.1.300 (WSL, 15 Oct; step 5.5.11's run
// report), cut down: every text replaced, only the ids, the times, the types and the counts kept (their keys re-written
// in another order than Claude Code's). session.jsonl is a session's own file (its main agent's two replies, four
// lines, and its closing cost-state line, whose modelUsage holds the session's totals with its subagent's);
// subagent.jsonl its one subagent's (two replies, three lines). Two trick files: other-layout.jsonl, session.jsonl
// with each reply's input_tokens renamed inputTokens; cut-short.jsonl, subagent.jsonl cut half-way through its last
// line.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/LastStep/Bonsai/internal/format"
)

// fixture is a measured transcript's absolute path, as a payload names it (forward slashes).
func fixture(t *testing.T, name string) string {
	t.Helper()
	p, err := filepath.Abs(filepath.Join("testdata", "claude-2.1.300", name))
	if err != nil {
		t.Fatal(err)
	}
	return filepath.ToSlash(p)
}

// jsonText writes a string as JSON.
func jsonText(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

// subagentStop is a subagent's SubagentStop payload naming its transcript; sessionEnd its session's SessionEnd naming
// the session's.
func subagentStop(root, sess, agent, transcript string) string {
	return `{"session_id":"` + sess + `","transcript_path":"/nowhere/session.jsonl","cwd":` + jsonText(filepath.ToSlash(root)) +
		`,"hook_event_name":"SubagentStop","agent_id":"` + agent + `","agent_type":"workflow:helper","stop_hook_active":false,` +
		`"agent_transcript_path":` + jsonText(transcript) + `,"last_assistant_message":"done"}`
}

func sessionEnd(root, sess, transcript string) string {
	return `{"session_id":"` + sess + `","transcript_path":` + jsonText(transcript) + `,"cwd":` + jsonText(filepath.ToSlash(root)) +
		`,"hook_event_name":"SessionEnd","reason":"prompt_input_exit"}`
}

// usageOf is a usage record as one line: its end, agent, model, counts, kind and reason.
func usageOf(l *format.Log) string {
	counts := "<null>"
	if l.Usage != nil {
		counts = fmt.Sprintf("%d/%d/%d/%d", l.Usage.Input, l.Usage.Output, l.Usage.CacheRead, l.Usage.CacheWrite)
	}
	return fmt.Sprintf("%s %s %s %s %s %s", s(l.AgentEvent), s(l.SubagentID), s(l.Model), counts, s(l.Kind), s(l.Reason))
}

// usageRecords gives a session's usage records, each held to bonsai.log/1.
func usageRecords(t *testing.T, root, sess string) []*format.Log {
	t.Helper()
	var out []*format.Log
	for _, l := range records(t, root, sess) {
		if l.Event == "usage" {
			out = append(out, l)
		}
	}
	return out
}

// The measured transcripts: each total equal to the counts in the file, each reply counted once (its three lines and
// two lines read as one), the subagent's from its own file and the session's from its own, and the two together equal
// to the session file's own closing totals (cost-state's modelUsage, which counts both): input, output, cache read and
// cache write.
func TestUsageFromTheMeasuredTranscripts(t *testing.T) {
	_, root := project(t)
	sess := "393fe3e7-4666-4e2a-80e7-62cfe58f6a23"
	if code := Record(opts(root, subagentStop(root, sess, "ab8125621420b9524", fixture(t, "subagent.jsonl")), nil)); code != 0 {
		t.Fatalf("exit %d", code)
	}
	if code := Record(opts(root, sessionEnd(root, sess, fixture(t, "session.jsonl")), nil)); code != 0 {
		t.Fatalf("exit %d", code)
	}
	got := usageRecords(t, root, sess)
	want := []string{
		"SubagentStop ab8125621420b9524 claude-opus-5-5 4/186/6637/9363 <null> <null>",
		"SessionEnd <null> claude-opus-5-5 4/383/38638/14439 <null> <null>",
	}
	if len(got) != len(want) {
		t.Fatalf("%d usage records", len(got))
	}
	for i, l := range got {
		if g := usageOf(l); g != want[i] {
			t.Errorf("record %d: %q, want %q", i, g, want[i])
		}
		if s(l.SubagentType) != map[bool]string{true: "workflow:helper", false: "<null>"}[i == 0] || s(l.Agent) != "claude-code" {
			t.Errorf("record %d's agent: %+v", i, l)
		}
	}
	// The session's own totals, as Claude Code closed its file with them.
	raw, err := os.ReadFile(fixture(t, "session.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	lines := bytes.Split(bytes.TrimSpace(raw), []byte("\n"))
	var cost struct {
		Type       string `json:"type"`
		ModelUsage map[string]struct {
			Input      int64 `json:"inputTokens"`
			Output     int64 `json:"outputTokens"`
			CacheRead  int64 `json:"cacheReadInputTokens"`
			CacheWrite int64 `json:"cacheCreationInputTokens"`
		} `json:"modelUsage"`
	}
	if err := json.Unmarshal(lines[len(lines)-1], &cost); err != nil || cost.Type != "cost-state" || len(cost.ModelUsage) != 1 {
		t.Fatalf("the fixture's last line: %v %+v", err, cost)
	}
	m := cost.ModelUsage["claude-opus-5-5"]
	a, b := got[0].Usage, got[1].Usage
	if a.Input+b.Input != m.Input || a.Output+b.Output != m.Output || a.CacheRead+b.CacheRead != m.CacheRead ||
		a.CacheWrite+b.CacheWrite != m.CacheWrite {
		t.Errorf("the two records do not sum to the session's own totals %+v", m)
	}
	// One end's records share one at; a resumed agent's second end writes newer records, each whole (never a
	// difference from the first).
	if code := Record(opts(root, subagentStop(root, sess, "ab8125621420b9524", fixture(t, "subagent.jsonl")), nil)); code != 0 {
		t.Fatalf("exit %d", code)
	}
	got = usageRecords(t, root, sess)
	if len(got) != 3 || usageOf(got[2]) != want[0] || got[2].At < got[0].At {
		t.Errorf("the second end's record: %+v", got)
	}
}

// The trick files: a transcript in another layout, and one cut short mid-line, each give one unknown record with its
// reason, and the hook exits 0 at once.
func TestUsageUnknownLayouts(t *testing.T) {
	_, root := project(t)
	sess := "0f1e2d3c-aaaa-4bbb-8ccc-0000000000a1"
	start := time.Now()
	Record(opts(root, subagentStop(root, sess, "a1", fixture(t, "other-layout.jsonl")), nil))
	Record(opts(root, sessionEnd(root, sess, fixture(t, "cut-short.jsonl")), nil))
	if took := time.Since(start); took > Budget {
		t.Errorf("the two ends took %s", took)
	}
	got := usageRecords(t, root, sess)
	want := []string{
		"SubagentStop a1 <null> <null> unknown line 24 of the transcript is a reply with no message.usage.input_tokens as a whole " +
			"number from 0, so its layout is not the one Bonsai reads",
		"SessionEnd <null> <null> <null> unknown the transcript ends part-way through line 23 (cut short)",
	}
	if len(got) != 2 {
		t.Fatalf("%d usage records", len(got))
	}
	for i, l := range got {
		if g := usageOf(l); g != want[i] {
			t.Errorf("record %d:\n got %q\nwant %q", i, g, want[i])
		}
	}
}

// A made-up transcript's lines, one per string, each with its line feed.
func transcript(t *testing.T, lines ...string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "t.jsonl")
	if err := os.WriteFile(p, []byte(strings.Join(lines, "\n")+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

// replyLine is one line of a reply in the measured layout, made up.
func replyLine(id, model string, in, out, read, write int64) string {
	return fmt.Sprintf(`{"parentUuid":"p","isSidechain":false,"type":"assistant","message":{"id":%q,"model":%q,"role":"assistant",`+
		`"usage":{"input_tokens":%d,"cache_creation_input_tokens":%d,"cache_read_input_tokens":%d,"output_tokens":%d},`+
		`"content":[{"type":"text","text":"made up"}]},"uuid":"u"}`, id, model, in, write, read, out)
}

const userLine = `{"parentUuid":null,"type":"user","message":{"role":"user","content":"tell the assistant: \"assistant\""},"uuid":"u0"}`

// Every way the counts are read, or are unknown, read on made-up transcripts.
func TestReadUsage(t *testing.T) {
	far := time.Now().Add(time.Hour)
	read := func(path string) string {
		models, reason := readUsage(path, far, time.Now)
		if reason != "" {
			return "unknown: " + reason
		}
		var out []string
		for _, m := range models {
			out = append(out, fmt.Sprintf("%s %d/%d/%d/%d", m.Model, m.Input, m.Output, m.CacheRead, m.CacheWrite))
		}
		return strings.Join(out, "; ")
	}
	dir := t.TempDir()
	many := []string{}
	for i := 0; i <= maxModels; i++ {
		many = append(many, replyLine(fmt.Sprintf("msg_%d", i), fmt.Sprintf("model-%02d", i), 1, 1, 1, 1))
	}
	cases := []struct{ name, path, want string }{
		{"no transcript named", "", "unknown: the hook input names no transcript"},
		{"a relative path", "t.jsonl", "unknown: the hook input's transcript path is not an absolute path"},
		{"not there", filepath.Join(dir, "gone.jsonl"), "unknown: the transcript is not there"},
		{"a folder", dir, "unknown: the transcript is not a regular file"},
		{"no reply", transcript(t, userLine, `{"type":"system","content":"x"}`),
			"unknown: the transcript holds no reply in the layout Bonsai reads"},
		{"an empty file", transcript(t), "unknown: the transcript holds no reply in the layout Bonsai reads"},
		// Two models; each reply's lines counted once, by its last line; a user line naming the word passed over; the
		// models in the order of their last reply.
		{"two models", transcript(t,
			userLine,
			replyLine("msg_a", "claude-b", 1, 10, 100, 1000),
			replyLine("msg_a", "claude-b", 1, 12, 100, 1000),
			replyLine("msg_b", "claude-a", 2, 20, 200, 2000),
			replyLine("msg_c", "claude-b", 3, 30, 300, 3000),
			replyLine("msg_c", "claude-b", 3, 30, 300, 3000),
			replyLine("msg_d", "claude-a", 4, 40, 400, 4000)),
			"claude-b 4/42/400/4000; claude-a 6/60/600/6000"},
		{"CRLF lines", transcript(t, replyLine("msg_a", "m", 1, 2, 3, 4)+"\r", userLine+"\r"), "m 1/2/3/4"},
		{"a line naming assistant that is no JSON", transcript(t, replyLine("msg_a", "m", 1, 2, 3, 4), `{"type":"assistant",`),
			"unknown: line 2 of the transcript is not JSON in the layout Bonsai reads"},
		{"a broken line that names no reply", transcript(t, `{"type":"user"`, replyLine("msg_a", "m", 1, 2, 3, 4)), "m 1/2/3/4"},
		{"a line with no type", transcript(t, `{"role":"assistant"}`), "unknown: line 1 of the transcript has no type, so its " +
			"layout is not the one Bonsai reads"},
		{"a type that is no string", transcript(t, `{"type":7,"role":"assistant"}`),
			"unknown: line 1 of the transcript has a type that is not text, so its layout is not the one Bonsai reads"},
		{"a reply with no usage", transcript(t, `{"type":"assistant","message":{"id":"m1","model":"m"}}`),
			"unknown: line 1 of the transcript is a reply with no message.usage, so its layout is not the one Bonsai reads"},
		{"a reply with no id", transcript(t, `{"type":"assistant","message":{"model":"m","usage":{}}}`),
			"unknown: line 1 of the transcript is a reply with no message.id, so its layout is not the one Bonsai reads"},
		{"a count as text", transcript(t, strings.Replace(replyLine("msg_a", "m", 1, 2, 3, 4), `"output_tokens":2`, `"output_tokens":"2"`, 1)),
			"unknown: line 1 of the transcript is a reply whose fields are not of the layout Bonsai reads"},
		{"a count below 0", transcript(t, replyLine("msg_a", "m", 1, -2, 3, 4)),
			"unknown: line 1 of the transcript is a reply with no message.usage.output_tokens as a whole number from 0, so its " +
				"layout is not the one Bonsai reads"},
		{"a total past the largest whole number", transcript(t, replyLine("msg_a", "m", 1, 2, 3, 1<<62), replyLine("msg_b", "m", 1, 2, 3, 1<<62),
			replyLine("msg_c", "m", 1, 2, 3, 1<<62)),
			"unknown: a total in the transcript passes the largest whole number a record holds"},
		{"too many models", transcript(t, many...), fmt.Sprintf("unknown: the transcript names more than %d models", maxModels)},
	}
	for _, c := range cases {
		if got := read(c.path); got != c.want {
			t.Errorf("%s:\n got %q\nwant %q", c.name, got, c.want)
		}
	}
	// Past its budget: unknown, and nothing more read.
	if _, reason := readUsage(transcript(t, replyLine("msg_a", "m", 1, 2, 3, 4)), time.Now().Add(-time.Second), time.Now); reason !=
		"reading the transcript ran past its budget of 2s" {
		t.Errorf("past the budget: %q", reason)
	}
}

// A line longer than the cap is never held: read through for its top-level type, passed over when that is another
// than assistant (wherever the type stands in the line, a nested one never read for it), and unknown otherwise.
func TestUsageLongLines(t *testing.T) {
	old := transcriptLineCap
	transcriptLineCap = 300
	t.Cleanup(func() { transcriptLineCap = old })
	long := strings.Repeat("x", 1000)
	reply := replyLine("msg_a", "m", 1, 2, 3, 4)
	far := time.Now().Add(time.Hour)
	cases := []struct {
		name, line, want string
	}{
		{"a long user line, its type last", `{"message":{"type":"assistant","content":"` + long + `"},"parentUuid":"p","type":"user"}`, ""},
		{"a long user line, its type first", `{"type":"progress","data":{"text":"` + long + `\"type\":\"assistant\""}}`, ""},
		{"a long reply", `{"type":"assistant","message":{"id":"x","content":"` + long + `"}}`,
			"line 1 of the transcript is a reply over 0 MiB, more than Bonsai holds of a line"},
		{"a long line with no type", `{"data":"` + long + `"}`,
			"line 1 of the transcript is over 0 MiB and names no type, so its layout is not the one Bonsai reads"},
		{"a long line whose type is no string", `{"type":["assistant"],"data":"` + long + `"}`,
			"line 1 of the transcript is over 0 MiB and names no type, so its layout is not the one Bonsai reads"},
	}
	for _, c := range cases {
		_, reason := readUsage(transcript(t, c.line, reply), far, time.Now)
		if reason != c.want {
			t.Errorf("%s:\n got %q\nwant %q", c.name, reason, c.want)
		}
	}
}

// The type scanner reads a line fed in pieces of any size alike.
func TestTypeScannerInPieces(t *testing.T) {
	cases := []struct{ line, want string }{
		{`{"type":"user","message":{"type":"assistant"}}`, "user"},
		{`{"message":{"type":"assistant","x":[1,{"type":"a"}]},"type":"user"}`, "user"},
		{`{"a":"\"type\":\"assistant\"","type" : "assistant"}`, "assistant"},
		{`{"typ":"x","types":"y","type":"system"}`, "system"},
		{`{"type":null}`, ""},
		{`["type","assistant"]`, ""},
	}
	for _, c := range cases {
		line, want := c.line, c.want
		for size := 1; size <= len(line); size++ {
			var sc typeScanner
			for i := 0; i < len(line); i += size {
				sc.feed([]byte(line[i:min(len(line), i+size)]))
			}
			if sc.value != want {
				t.Errorf("%s in pieces of %d: %q, want %q", line, size, sc.value, want)
				break
			}
		}
	}
}

// The cost at an agent's end: a transcript of about 20 MB (5,000 replies of three lines and their tool results)
// read well within the budget.
func TestUsageCost(t *testing.T) {
	if testing.Short() {
		t.Skip("writes a 20 MB transcript")
	}
	p := filepath.Join(t.TempDir(), "big.jsonl")
	var b strings.Builder
	result := `{"parentUuid":"p","type":"user","message":{"role":"user","content":[{"type":"tool_result","content":"` +
		strings.Repeat("y", 3000) + `"}]},"uuid":"u"}`
	for i := 0; i < 5000; i++ {
		id := fmt.Sprintf("msg_%06d", i)
		for j := 0; j < 3; j++ {
			b.WriteString(replyLine(id, "claude-made-up", 1, 2, 3, 4))
			b.WriteByte('\n')
		}
		b.WriteString(result)
		b.WriteByte('\n')
	}
	if err := os.WriteFile(p, []byte(b.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	start := time.Now()
	models, reason := readUsage(p, time.Now().Add(usageBudget), time.Now)
	took := time.Since(start)
	if reason != "" || len(models) != 1 || models[0].Input != 5000 || models[0].CacheWrite != 20000 {
		t.Fatalf("%q %+v", reason, models)
	}
	t.Logf("a transcript of %d bytes read in %s", b.Len(), took)
	if took > usageBudget/2 {
		t.Errorf("a %d-byte transcript took %s", b.Len(), took)
	}
}

// A subagent's last reply written a moment after its SubagentStop (as measured on WSL, 15 Oct) is counted: the
// transcript is read once it has settled. One that never grows is read after settleQuiet.
func TestUsageWaitsForTheLastReply(t *testing.T) {
	_, root := project(t)
	sess := "0f1e2d3c-aaaa-4bbb-8ccc-0000000000a2"
	path := transcript(t, userLine, replyLine("msg_a", "m", 1, 10, 100, 1000))
	go func() {
		time.Sleep(100 * time.Millisecond)
		f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0)
		if err != nil {
			return
		}
		_, _ = f.WriteString(replyLine("msg_b", "m", 2, 20, 200, 2000) + "\n")
		_ = f.Close()
	}()
	start := time.Now()
	Record(opts(root, subagentStop(root, sess, "a2", filepath.ToSlash(path)), nil))
	took := time.Since(start)
	got := usageRecords(t, root, sess)
	if len(got) != 1 || usageOf(got[0]) != "SubagentStop a2 m 3/30/300/3000 <null> <null>" {
		t.Fatalf("the records: %+v", got)
	}
	if took < 100*time.Millisecond+settleQuiet || took > Budget {
		t.Errorf("the end took %s", took)
	}
	// A session's end is read at once.
	start = time.Now()
	Record(opts(root, sessionEnd(root, sess, filepath.ToSlash(path)), nil))
	if took := time.Since(start); took >= settleQuiet {
		t.Errorf("a session's end took %s", took)
	}
}
