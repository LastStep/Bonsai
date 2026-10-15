package recorder

// The usage record (contract §8.2's note of 14 Oct; design/plan-5.md, note 5.5.12, 4; Rohan, 14 Oct: "Yes, raw
// counts"): at an agent's end, a subagent's SubagentStop and a session's own SessionEnd, after that end's own record,
// hook record reads the transcript whose path the payload gives (SubagentStop's agent_transcript_path, the subagent's
// own file; SessionEnd's transcript_path) and writes one usage record for each model in it: model the model, usage
// its totals over the whole transcript at that end (input, output, cache_read, cache_write), summed as Claude Code saved
// them, nothing else: no cost, no estimate, never a difference from an earlier record. A resumed agent ends again, so
// it gets newer records, each whole; bonsai agents shows the newest and sums nothing. One end's records share one at,
// and are written in the order of each model's last reply, so the newest of them names the model of the agent's last
// reply.
//
// The layout read is the one measured on Claude Code 2.1.300 (WSL, 15 Oct; step 5.5.11's run report): one JSON object
// a line; a reply's token counts in message.usage (input_tokens, output_tokens, cache_read_input_tokens as cache_read,
// cache_creation_input_tokens as cache_write), on lines whose type is assistant, with message.id and message.model. One
// reply is written as several lines, one per content block, each repeating its message.id and the usage as it stood
// when that block was written (a reply's two tool calls, measured: 16 output tokens on its first line, 153 on its
// last), so each reply is counted once, by its message.id, with its last line's counts. A session's file holds its main
// agent's replies only (a subagent's are in the subagent's own file), so a session's records never count its
// subagents'; the file's last line, cost-state, is not read. Only the counts, the model and the reply's id are read,
// never any text.
//
// At a subagent's end the transcript is read once it has settled: Claude Code writes a subagent's last lines a moment
// after its SubagentStop hook starts (measured 15 Oct on WSL, a real session of a lead and a helper: each one's final
// reply was missing from its file when its records were written at once, and in the file after the session), so the
// recorder first waits until the file has not grown for settleQuiet, at most settleMax (settle). The hook is async, so
// the agent's end never waits for it. Known limit (VG, 15 Oct): the wait is a heuristic, read on the file's size alone.
// A final reply that reaches the file after settleMax, or after a pause longer than settleQuiet, is missed without a
// word: the records are written from the replies before it, their kind the model's, not unknown, and its tokens are
// not counted (a resumed agent's later end writes whole records again). The payload's last_assistant_message, the
// final reply's text, could tell, but the payload reader (internal/hookin) keeps no such field, so the file is not
// checked against it. A session's transcript was whole at its SessionEnd in the same session (its
// records equal its file's replies, and with its subagents' files its own closing totals), and is read at once: that
// line is synchronous, and the person's exit waits for it.
//
// The transcript is read as it streams, line by line: a line costs memory for itself alone, up to transcriptLineCap
// (8 MiB); a longer line is read through without being held (typeScanner: its top-level type, and nothing else), and
// passed over when that type is another than assistant; otherwise the end is unknown. A line that does not name
// "assistant" holds no reply and is not decoded.
//
// Where the counts cannot be read, the end gets one record whose kind is unknown and whose reason says why (usage and
// model null), and nothing fails or blocks: the payload names no transcript, or no absolute path; the file is not there,
// is not a regular file, or cannot be read; it ends part-way through a line (cut short); a line naming "assistant" is
// not JSON, or a reply lacks a field of the measured layout (a reply's id, model or one of its four counts, each a
// whole number from 0); it holds no reply at all; it names more than maxModels models; a total passes the largest
// whole number a record holds; or the reading runs past usageBudget. The reasons never name the file's path (its
// folder holds the user's name: contract §2.6).

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/hookin"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// usageBudget is how long reading a transcript at an agent's end may take (a variable for the tests). SessionEnd's
// line has 5 s and the hook's own Budget is 4 s: its session_end line takes milliseconds, the reading at most this,
// and the cleaning after it clean.SessionEndBudget and one step. At a SubagentStop (async, no cleaning) the wait for
// the transcript to settle comes first, at most settleMax: together within the Budget.
var usageBudget = 2 * time.Second

// settleQuiet, settleMax and settlePoll are how a subagent's transcript is waited for (settle): read once its size has
// not changed for settleQuiet, looked at every settlePoll, waited for at most settleMax (variables for the tests).
var (
	settleQuiet = 300 * time.Millisecond
	settleMax   = 1500 * time.Millisecond
	settlePoll  = 25 * time.Millisecond
)

// transcriptLineCap caps the memory one transcript line may take, in bytes (a variable for the tests); maxModels caps
// the models one end's records may name.
var transcriptLineCap = 8 << 20

const maxModels = 16

// usageUnknown is the usage record's kind when the transcript could not be read.
const usageUnknown = "unknown"

// modelUsage is one model's totals over a transcript.
type modelUsage struct {
	Model                                string
	Input, Output, CacheRead, CacheWrite int64
	last                                 int // the line of its last reply, for the records' order
}

// usageFrom is the transcript an end's usage records read: the subagent's own at SubagentStop, the session's at
// SessionEnd; ok false for any other event, or a SubagentStop whose agent id cannot be recorded.
func usageFrom(p *hookin.Payload) (path string, ok bool) {
	switch p.Event {
	case "SubagentStop":
		return p.AgentTranscriptPath, hookin.ID(p.AgentID) != ""
	case "SessionEnd":
		return p.TranscriptPath, true
	}
	return "", false
}

// writeUsage writes an end's usage records after its own record (this file's comment): one for each model, or one
// unknown with its reason. A record that cannot be written ends the writing; nothing is returned, since nothing may
// fail or block an agent's end.
func writeUsage(o Options, p *hookin.Payload, cfg *workspace.Config, local workspace.Local, path string) {
	if p.Event == "SubagentStop" {
		settleWait(path)
	}
	models, reason := readUsage(path, o.Now().Add(usageBudget), o.Now)
	at := o.Now()
	if reason != "" {
		l := base(o, p, cfg, local, "usage", at)
		l.Kind, l.Reason = ptr(usageUnknown), ptr(reason)
		_, _ = writeLog(local.Main, l)
		return
	}
	for _, m := range models {
		l := base(o, p, cfg, local, "usage", at)
		l.Model = text(m.Model)
		l.Usage = &format.LogUsage{Input: m.Input, Output: m.Output, CacheRead: m.CacheRead, CacheWrite: m.CacheWrite}
		if _, err := writeLog(local.Main, l); err != nil {
			return
		}
	}
}

// settleWait is settle (a variable for the tests).
var settleWait = settle

// settle waits until a transcript's size has not changed for settleQuiet, at most settleMax, by the wall clock (this
// file's comment). A file that is not there counts as one size, so one that appears is waited for too; a path that
// is none, or not absolute, is not waited for (readUsage says why it is not read).
func settle(path string) {
	if path == "" || !filepath.IsAbs(path) {
		return
	}
	size := func() int64 {
		fi, err := os.Stat(path)
		if err != nil {
			return -1
		}
		return fi.Size()
	}
	start := time.Now()
	last, quiet := size(), start
	for time.Since(start) < settleMax {
		time.Sleep(settlePoll)
		if n := size(); n != last {
			last, quiet = n, time.Now()
			continue
		}
		if time.Since(quiet) >= settleQuiet {
			return
		}
	}
}

// reply is one reply's counts as its last line holds them.
type reply struct {
	model                                string
	input, output, cacheRead, cacheWrite int64
	line                                 int
}

// transcriptLine is what a line naming "assistant" is decoded into: its type, and a reply's id, model and counts.
type transcriptLine struct {
	Type    *string `json:"type"`
	Message *struct {
		ID    *string `json:"id"`
		Model *string `json:"model"`
		Usage *struct {
			Input      *int64 `json:"input_tokens"`
			Output     *int64 `json:"output_tokens"`
			CacheRead  *int64 `json:"cache_read_input_tokens"`
			CacheWrite *int64 `json:"cache_creation_input_tokens"`
		} `json:"usage"`
	} `json:"message"`
}

// readUsage reads a transcript's token counts (this file's comment): each model's totals, in the order of each model's
// last reply, or why they cannot be read. now is the clock the deadline is read by.
func readUsage(path string, deadline time.Time, now func() time.Time) ([]modelUsage, string) {
	switch {
	case path == "":
		return nil, "the hook input names no transcript"
	case !filepath.IsAbs(path):
		return nil, "the hook input's transcript path is not an absolute path"
	}
	fi, err := os.Stat(path)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return nil, "the transcript is not there"
	case err != nil:
		return nil, "the transcript cannot be read (" + fileFault(err) + ")"
	case !fi.Mode().IsRegular():
		return nil, "the transcript is not a regular file"
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, "the transcript cannot be read (" + fileFault(err) + ")"
	}
	defer func() { _ = f.Close() }()
	replies := map[string]reply{}
	lr := lineReader{r: bufio.NewReaderSize(f, 64<<10), max: transcriptLineCap}
	for n := 1; ; n++ {
		if now().After(deadline) {
			return nil, fmt.Sprintf("reading the transcript ran past its budget of %s", usageBudget)
		}
		line, over, whole, err := lr.next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, "the transcript cannot be read (" + fileFault(err) + ")"
		}
		if !whole {
			return nil, fmt.Sprintf("the transcript ends part-way through line %d (cut short)", n)
		}
		if over {
			switch t := lr.scan; {
			case t.found && t.value != "" && t.value != "assistant":
				continue // no reply: passed over, never held
			case t.value == "assistant":
				return nil, fmt.Sprintf("line %d of the transcript is a reply over %d MiB, more than Bonsai holds of a line", n,
					transcriptLineCap>>20)
			}
			return nil, fmt.Sprintf("line %d of the transcript is over %d MiB and names no type, so its layout is not the one "+
				"Bonsai reads", n, transcriptLineCap>>20)
		}
		if !bytes.Contains(line, []byte(`"assistant"`)) {
			continue
		}
		var tl transcriptLine
		err = json.Unmarshal(line, &tl)
		var syntax *json.SyntaxError
		var typeErr *json.UnmarshalTypeError
		switch {
		case errors.As(err, &syntax) || (err != nil && tl.Type == nil):
			return nil, fmt.Sprintf("line %d of the transcript is not JSON in the layout Bonsai reads", n)
		case errors.As(err, &typeErr) && typeErr.Field == "type":
			return nil, fmt.Sprintf("line %d of the transcript has a type that is not text, so its layout is not the one Bonsai reads", n)
		case tl.Type == nil:
			return nil, fmt.Sprintf("line %d of the transcript has no type, so its layout is not the one Bonsai reads", n)
		case *tl.Type != "assistant":
			continue
		case err != nil:
			return nil, fmt.Sprintf("line %d of the transcript is a reply whose fields are not of the layout Bonsai reads", n)
		}
		r, why := replyOf(&tl, n)
		if why != "" {
			return nil, fmt.Sprintf("line %d of the transcript is a reply with no %s, so its layout is not the one Bonsai reads", n, why)
		}
		replies[*tl.Message.ID] = r
	}
	if len(replies) == 0 {
		return nil, "the transcript holds no reply in the layout Bonsai reads"
	}
	return totals(replies)
}

// replyOf reads one reply's model and counts, or names the field of the measured layout it lacks.
func replyOf(tl *transcriptLine, n int) (reply, string) {
	m := tl.Message
	switch {
	case m == nil:
		return reply{}, "message"
	case m.ID == nil || *m.ID == "":
		return reply{}, "message.id"
	case m.Model == nil || *m.Model == "":
		return reply{}, "message.model"
	case m.Usage == nil:
		return reply{}, "message.usage"
	}
	u := m.Usage
	for _, c := range []struct {
		name string
		v    *int64
	}{{"input_tokens", u.Input}, {"output_tokens", u.Output}, {"cache_read_input_tokens", u.CacheRead},
		{"cache_creation_input_tokens", u.CacheWrite}} {
		if c.v == nil || *c.v < 0 {
			return reply{}, "message.usage." + c.name + " as a whole number from 0"
		}
	}
	return reply{model: *m.Model, input: *u.Input, output: *u.Output, cacheRead: *u.CacheRead, cacheWrite: *u.CacheWrite, line: n}, ""
}

// totals sums each model's replies, each counted once, and orders the models by their last reply, oldest first.
func totals(replies map[string]reply) ([]modelUsage, string) {
	byModel := map[string]*modelUsage{}
	ids := make([]string, 0, len(replies))
	for id := range replies {
		ids = append(ids, id)
	}
	sort.Strings(ids) // a sum is the same in any order; sorted so the walk itself is never in map order
	for _, id := range ids {
		r := replies[id]
		m := byModel[r.model]
		if m == nil {
			if len(byModel) == maxModels {
				return nil, fmt.Sprintf("the transcript names more than %d models", maxModels)
			}
			m = &modelUsage{Model: r.model}
			byModel[r.model] = m
		}
		for _, add := range []struct {
			to *int64
			n  int64
		}{{&m.Input, r.input}, {&m.Output, r.output}, {&m.CacheRead, r.cacheRead}, {&m.CacheWrite, r.cacheWrite}} {
			if *add.to > math.MaxInt64-add.n {
				return nil, "a total in the transcript passes the largest whole number a record holds"
			}
			*add.to += add.n
		}
		m.last = max(m.last, r.line)
	}
	out := make([]modelUsage, 0, len(byModel))
	for _, m := range byModel {
		out = append(out, *m)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].last < out[j].last }) // distinct: one line holds one reply
	return out, ""
}

// fileFault names a file error without its path (the transcript's folder holds the user's name).
func fileFault(err error) string {
	switch {
	case errors.Is(err, fs.ErrPermission):
		return "permission denied"
	case errors.Is(err, fs.ErrNotExist):
		return "not there"
	}
	return "an error of the file system"
}

// lineReader reads a file's lines, each held whole up to max bytes; a longer one is read through a typeScanner, held
// nowhere.
type lineReader struct {
	r    *bufio.Reader
	buf  []byte
	max  int
	scan typeScanner
}

// next gives the next line without its line feed (or carriage return and line feed), and whole false when the file
// ended inside it. A line longer than max gives over true and no text: its top-level type is in lr.scan then. io.EOF
// once no line is left.
func (lr *lineReader) next() (line []byte, over, whole bool, err error) {
	lr.buf = lr.buf[:0]
	for {
		frag, err := lr.r.ReadSlice('\n')
		switch {
		case over:
			lr.scan.feed(frag)
		case len(lr.buf)+len(frag) > lr.max:
			over = true
			lr.scan = typeScanner{}
			lr.scan.feed(lr.buf)
			lr.scan.feed(frag)
			lr.buf = lr.buf[:0]
		default:
			lr.buf = append(lr.buf, frag...)
		}
		switch {
		case err == nil && over:
			return nil, true, true, nil
		case err == nil:
			return bytes.TrimSuffix(bytes.TrimSuffix(lr.buf, []byte("\n")), []byte("\r")), false, true, nil
		case errors.Is(err, bufio.ErrBufferFull):
			continue
		case errors.Is(err, io.EOF):
			if len(lr.buf) == 0 && !over {
				return nil, false, false, io.EOF
			}
			return lr.buf, over, false, nil
		default:
			return nil, false, false, err
		}
	}
}

// typeScanner reads the value of a JSON object's top-level "type" key from a stream of bytes, holding none of them but
// the key or value it is reading (at most typeCap bytes; a longer one is no key or type it looks for). It reads only
// as far as that value.
type typeScanner struct {
	depth            int
	inString, escape bool
	wantKey          bool // at the top level, the next string is a key
	isKey, isType    bool // the string being read is a top-level key, or the value of the top-level "type"
	cur              []byte
	long             bool
	lastKey          string
	found            bool   // the top-level "type"'s value has been read
	value            string // that value, when it was a string of at most typeCap bytes; "" otherwise
}

// typeCap caps the key or type value a typeScanner holds.
const typeCap = 32

func (s *typeScanner) feed(b []byte) {
	for _, c := range b {
		if s.found {
			return
		}
		if s.inString {
			switch {
			case s.escape:
				s.escape = false
				s.add(c)
			case c == '\\':
				s.escape = true
				s.add(c)
			case c == '"':
				s.inString = false
				s.endString()
			default:
				s.add(c)
			}
			continue
		}
		atType := s.depth == 1 && !s.wantKey && s.lastKey == "type"
		switch c {
		case '"':
			s.inString, s.cur, s.long = true, s.cur[:0], false
			s.isKey, s.isType = s.depth == 1 && s.wantKey, atType
		case '{', '[':
			if atType {
				s.found = true // a type that is no string
				return
			}
			s.depth++
			s.wantKey = s.depth == 1 && c == '{'
		case '}', ']':
			s.depth--
		case ',':
			if s.depth == 1 {
				s.wantKey, s.lastKey = true, ""
			}
		case ':':
			if s.depth == 1 {
				s.wantKey = false
			}
		case ' ', '\t', '\r', '\n':
		default:
			if atType {
				s.found = true // a number, true, false or null
				return
			}
		}
	}
}

func (s *typeScanner) add(c byte) {
	if len(s.cur) < typeCap {
		s.cur = append(s.cur, c)
	} else {
		s.long = true
	}
}

func (s *typeScanner) endString() {
	switch {
	case s.isKey:
		s.lastKey = ""
		if !s.long {
			s.lastKey = string(s.cur)
		}
	case s.isType:
		s.found = true
		if !s.long {
			s.value = string(s.cur)
		}
	}
}
