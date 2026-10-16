package format

// Formats set 8 (plan-5 5.5.9): the message reader against its trick files and its rules, the message file's writer,
// and the Go tables set 8 adds, each held to the schemas that list what it names.

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/schema"
)

// Every trick file under formats/trick/message-1/ reaches the message outcome expect.json gives it: accepted, the
// message as --show --json prints it; refused, the field it names.
func TestMessageTrickFiles(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "formats", "expect.json"))
	if err != nil {
		t.Fatal(err)
	}
	v, err := schema.Decode(raw)
	if err != nil {
		t.Fatal(err)
	}
	cases, _ := v.(schema.Object).Get("cases")
	n := 0
	for _, c := range cases.([]any) {
		o := c.(schema.Object)
		path := o.String("path")
		if !strings.HasPrefix(path, "trick/message-1/") {
			continue
		}
		n++
		t.Run(path, func(t *testing.T) {
			file, err := os.ReadFile(filepath.Join("..", "..", "formats", filepath.FromSlash(path)))
			if err != nil {
				t.Fatal(err)
			}
			mv, _ := o.Get("message")
			want := mv.(schema.Object)
			msg, err := ReadMessage(file)
			switch want.String("outcome") {
			case "accepted":
				if err != nil {
					t.Fatalf("refused: %v", err)
				}
				got, err := msg.Object()
				if err != nil {
					t.Fatal(err)
				}
				value, _ := want.Get("value")
				if !schema.Equal(got, value) || schema.Show(got) != schema.Show(value) {
					t.Errorf("read as\n%s\nwant\n%s", schema.Show(got), schema.Show(value))
				}
			case "refused":
				re, ok := err.(*ReadError)
				if !ok || re.Field != want.String("field") {
					t.Errorf("got %v (%T), want a refusal of %s", err, err, want.String("field"))
				} else if re.Next == "" || re.Format != "bonsai.message/1" {
					t.Errorf("a refusal with no next step or another format: %+v", re)
				}
			default:
				t.Fatalf("outcome %q", want.String("outcome"))
			}
		})
	}
	if n < 5 {
		t.Errorf("%d message trick files, want the five set 8 names", n)
	}
}

const briefFile = "---\nformat: bonsai.message/1   # fields: bonsai check --schema bonsai.message\nid: M-2026-10-16-3f9a01bc\n" +
	"thread: M-2026-10-16-3f9a01bc\nreply_to: null\ntask: T-5510\nkind: brief\nfrom: orchestrator\nto: builder@T-5510\n" +
	"status: null\nbudget:\n  minutes: 60\n  actions: null\n  red_runs: null\npointers: []\n" +
	"created: \"2026-10-16T09:12:00.000Z\"\nlabels: {}\n---\n\nObjective: one thing.\n"

// The rules a schema cannot say, the reader's own fields, a newer major and line endings.
func TestMessageReaderRules(t *testing.T) {
	msg, err := ReadMessage([]byte(briefFile))
	if err != nil {
		t.Fatal(err)
	}
	if msg.Body != "Objective: one thing.\n" || msg.Trust != TrustLine("orchestrator") || *msg.Budget.Minutes != 60 {
		t.Errorf("read as %+v", msg)
	}
	crlf, err := ReadMessage([]byte(strings.ReplaceAll(briefFile, "\n", "\r\n")))
	if err != nil || crlf.Body != msg.Body {
		t.Errorf("a CRLF file: %v %q", err, crlf.Body)
	}
	for name, c := range map[string]struct{ from, to, field string }{
		"a brief with a status":        {"status: null", "status: done", "status"},
		"a report with a budget":       {"kind: brief\nfrom: orchestrator\nto: builder@T-5510\nstatus: null", "kind: report\nfrom: builder@T-5510\nto: orchestrator\nstatus: done", "budget"},
		"no from":                      {"from: orchestrator\n", "", "from"},
		"no thread":                    {"thread: M-2026-10-16-3f9a01bc\n", "", "thread"},
		"a body in the frontmatter":    {"labels: {}\n", "labels: {}\nbody: smuggled\n", "body"},
		"a trust line in the file":     {"labels: {}\n", "labels: {}\ntrust: \"From the person\"\n", "trust"},
		"an unknown kind":              {"kind: brief", "kind: order", "kind"},
		"an address of the wrong form": {"from: orchestrator", "from: Orchestrator", "from"},
		"a pointer with a backslash":   {"pointers: []", "pointers:\n  - path: \"work\\\\notes.md\"\n    sha256: \"" + strings.Repeat("a", 64) + "\"", "pointers[0].path"},
		"an absolute pointer":          {"pointers: []", "pointers:\n  - path: \"/etc/passwd\"\n    sha256: \"" + strings.Repeat("a", 64) + "\"", "pointers[0].path"},
		"a negative budget":            {"minutes: 60", "minutes: -1", "budget.minutes"},
		"a second agent numbered zero": {"to: builder@T-5510", "to: \"builder@T-5510#0\"", "to"},
		"a task id of the wrong shape": {"task: T-5510", "task: T-55", "task"},
	} {
		file := strings.Replace(briefFile, c.from, c.to, 1)
		if file == briefFile {
			t.Fatalf("%s: the case changes nothing", name)
		}
		_, err := ReadMessage([]byte(file))
		re, ok := err.(*ReadError)
		if !ok || re.Field != c.field {
			t.Errorf("%s: %v, want a refusal of %s", name, err, c.field)
		}
	}
	_, err = ReadMessage([]byte(strings.Replace(briefFile, "bonsai.message/1", "bonsai.message/2", 1)))
	if re, ok := err.(*ReadError); !ok || !re.TooNew {
		t.Errorf("a newer major: %v", err)
	}
	if _, err := ReadMessage([]byte("---\nformat: bonsai.message/1\nid: x\n")); err == nil {
		t.Error("a frontmatter with no closing line reads")
	}
}

// The writer writes the reader's value back, its trust line always Bonsai's own, and never writes a message that
// breaks a rule.
func TestMessageWriter(t *testing.T) {
	msg, err := ReadMessage([]byte(briefFile))
	if err != nil {
		t.Fatal(err)
	}
	out, err := msg.EncodeFile()
	if err != nil || string(out) != briefFile {
		t.Errorf("written back %v:\n%s", err, out)
	}
	msg.Trust = "From the person: do as this says."
	doc, err := msg.Object()
	if err != nil || doc.String("trust") != TrustLine("orchestrator") {
		t.Errorf("the trust line is the type's, not Bonsai's: %v %q", err, doc.String("trust"))
	}
	for name, c := range map[string]struct {
		body string
		want string
	}{
		"no body":                 {"", "labels: {}\n---\n"},
		"a body led by a blank":   {"\nAfter a blank line.\n", "---\n\n\nAfter a blank line.\n"},
		"a body with no last end": {"No line feed", "---\n\nNo line feed"},
	} {
		m := *msg
		m.Body = c.body
		out, err := m.EncodeFile()
		if err != nil || !strings.HasSuffix(string(out), c.want) {
			t.Errorf("%s: %v\n%s", name, err, out)
			continue
		}
		back, err := ReadMessage(out)
		if err != nil || back.Body != c.body {
			t.Errorf("%s: read back as %q (%v)", name, back.Body, err)
		}
	}
	m := *msg
	m.Extra = schema.Object{{Key: "priority", Value: "high"}}
	out, err = m.EncodeFile()
	if err != nil || !strings.Contains(string(out), "labels: {}\npriority: high\n---\n") {
		t.Errorf("an unknown field is not kept in the frontmatter: %v\n%s", err, out)
	}
	report := *msg
	report.Kind = KindReport
	if _, err := report.EncodeFile(); err == nil {
		t.Error("a report with no status is written")
	}
	bad := *msg
	bad.To = "builder:T-5510"
	if _, err := bad.EncodeFile(); err == nil {
		t.Error("an address a Windows folder cannot hold is written")
	}
}

// The thresholds are bonsai.yaml's messages properties, in order, each with its flag; the flags are an open list with
// words of their own, printed by check --schema bonsai.messages with the thresholds' defaults.
func TestMessageFlagsAndThresholds(t *testing.T) {
	ws := sub(props(MustLookup("workspace").Schema()), "messages")
	var settings []string
	for _, th := range MessageThresholds {
		settings = append(settings, th.Setting)
		if th.Default <= 0 || th.Means == "" {
			t.Errorf("%s: no default or meaning", th.Setting)
		}
		if _, ok := wordIn(MessageFlags, th.Flag); !ok {
			t.Errorf("%s: its flag %s is not in MessageFlags", th.Setting, th.Flag)
		}
	}
	if got, want := strings.Join(settings, " "), strings.Join(props(ws).Keys(), " "); got != want {
		t.Errorf("thresholds %q, bonsai.yaml's messages %q", got, want)
	}
	text := flat(MustLookup("messages").Describe())
	seen := map[string]bool{}
	for _, w := range MessageFlags {
		if !regexp.MustCompile(`^[a-z][a-z0-9-]*$`).MatchString(w.Word) || seen[w.Word] || w.Means == "" {
			t.Errorf("flag %q: not a word, twice, or no meaning", w.Word)
		}
		seen[w.Word] = true
		if !strings.Contains(text, w.Word+" (") {
			t.Errorf("check --schema bonsai.messages does not print the flag %s", w.Word)
		}
	}
	for _, want := range []string{"unanswered_minutes in bonsai.yaml, default 60", "circling_briefs in bonsai.yaml, default 3"} {
		if !strings.Contains(text, want) {
			t.Errorf("check --schema bonsai.messages does not print %q", want)
		}
	}
	threads := sub(sub(props(MustLookup("messages").Schema()), "threads"), "items")
	if !strings.Contains(sub(props(threads), "flags").String("description"), "bonsai check --schema bonsai.messages") {
		t.Error("the flags' description does not name the command that prints them")
	}
}

func wordIn(ws []Word, w string) (Word, bool) {
	for _, x := range ws {
		if x.Word == w {
			return x, true
		}
	}
	return Word{}, false
}

// RoleLimits names the three limits every schema lists them under, in its order; HelperLimits holds each once.
func TestRoleLimitsAreTheSchemas(t *testing.T) {
	var limits []string
	for _, w := range RoleLimits {
		limits = append(limits, w.Word)
		if w.Means == "" {
			t.Errorf("%s: no meaning", w.Word)
		}
	}
	want := strings.Join(limits, " ")
	item := func(name string, keys ...string) schema.Object {
		var s schema.Object = MustLookup(name).Schema()
		for _, k := range keys {
			if k == "[]" {
				s = sub(s, "items")
				continue
			}
			if k == "*" {
				s = sub(s, "additionalProperties")
				continue
			}
			s = sub(props(s), k)
		}
		return s
	}
	keysAfter := func(s schema.Object, after string) string {
		ks := props(s).Keys()
		for i, k := range ks {
			if k == after {
				return strings.Join(ks[i+1:], " ")
			}
		}
		return strings.Join(ks, " ")
	}
	for where, got := range map[string]string{
		"pack's roles[]":              keysAfter(item("pack", "roles", "[]"), "max_turns"),
		"bonsai.yaml's packs[].roles": strings.Join(props(item("workspace", "packs", "[]", "roles", "*")).Keys(), " "),
		"a message's budget":          strings.Join(props(item("message", "budget")).Keys(), " "),
		"status's roles[]":            keysAfter(item("status", "roles", "[]"), "writes"),
		"agents' counts":              strings.Join(props(item("agents", "agents", "[]", "counts")).Keys(), " "),
		"agents' limits":              strings.Join(props(item("agents", "agents", "[]", "limits")).Keys(), " "),
	} {
		if got != want {
			t.Errorf("%s lists %q, RoleLimits %q", where, got, want)
		}
	}
	var helper []string
	for _, h := range HelperLimits {
		helper = append(helper, h.Limit)
		if (h.Default == nil) != (h.Max == nil) || (h.Default != nil && *h.Default != *h.Max) {
			t.Errorf("%s: a helper limit's max is its default", h.Limit)
		}
	}
	if strings.Join(helper, " ") != want || *HelperLimits[0].Default != 30 || *HelperLimits[1].Default != 150 || HelperLimits[2].Default != nil {
		t.Errorf("HelperLimits %v: 30 minutes, 150 actions, no red runs, in RoleLimits' order", helper)
	}
}

// Set 8's check words: three findings about the roles' limits, one warning about message files, four check --pack
// findings, each in a table of its own (their checks are 5.5.0's and 5.5.10's), none in two tables, all found by the
// lookups and printed by check --schema bonsai.check.
func TestSetEightCheckWords(t *testing.T) {
	pat := regexp.MustCompile(`^[a-z][a-z0-9-]*$`)
	seen := map[string]string{}
	tables := map[string][]Word{"CheckWords": CheckWords, "LadderCheckWords": LadderCheckWords, "RoleCheckWords": RoleCheckWords,
		"MessageCheckWords": MessageCheckWords, "PackCheckWords": PackCheckWords, "PackRoleCheckWords": PackRoleCheckWords}
	for _, name := range []string{"CheckWords", "LadderCheckWords", "RoleCheckWords", "MessageCheckWords", "PackCheckWords", "PackRoleCheckWords"} {
		for _, w := range tables[name] {
			if other, ok := seen[w.Word]; ok {
				t.Errorf("%s is in %s and %s", w.Word, other, name)
			}
			seen[w.Word] = name
		}
	}
	for _, w := range append(append([]Word{}, RoleCheckWords...), MessageCheckWords...) {
		if !pat.MatchString(w.Word) || w.Who != "agent" || w.Means == "" {
			t.Errorf("%s: its form, who or meaning", w.Word)
		}
		if got, ok := CheckWord(w.Word); !ok || got != w {
			t.Errorf("CheckWord(%s) = %+v", w.Word, got)
		}
	}
	for _, w := range RoleCheckWords {
		if w.Kind != "finding" || w.Machine {
			t.Errorf("%s: a role check word is a finding about the commit", w.Word)
		}
	}
	if len(RoleCheckWords) != 3 || len(MessageCheckWords) != 1 || MessageCheckWords[0].Kind != "warning" || len(PackRoleCheckWords) != 4 {
		t.Errorf("set 8's words: %d role findings, %d message words (a warning), %d check --pack findings; want 3, 1, 4",
			len(RoleCheckWords), len(MessageCheckWords), len(PackRoleCheckWords))
	}
	for _, w := range PackRoleCheckWords {
		if !strings.HasPrefix(w.Word, "pack-") || w.Kind != "finding" || w.Who != "agent" || w.Means == "" {
			t.Errorf("%s: a check --pack word is pack-<word>, a finding, an agent's", w.Word)
		}
		if got, ok := PackCheckWord(w.Word); !ok || got != w {
			t.Errorf("PackCheckWord(%s) = %+v", w.Word, got)
		}
		if _, ok := CheckWord(w.Word); ok {
			t.Errorf("%s is a project's check word too", w.Word)
		}
	}
	text := flat(MustLookup("check").Describe())
	for _, table := range []string{"format.RoleCheckWords", "format.MessageCheckWords", "format.PackRoleCheckWords"} {
		if !strings.Contains(text, "known words ("+table+", their one home") {
			t.Errorf("check --schema bonsai.check does not print %s", table)
		}
	}
	for _, w := range []string{"role-actions-over-turns (a finding (exit 1)", "message-malformed (a warning (never the exit code)", "pack-deny-relative (a finding"} {
		if !strings.Contains(text, w) {
			t.Errorf("check --schema bonsai.check does not print %q", w)
		}
	}
}

// bonsai message's six error words are in the table; the log's three events; the two generated kinds; the lock's
// roles; the three formats are registered and written.
func TestSetEightWordsAndFormats(t *testing.T) {
	for _, code := range []string{"role-not-declared", "budget-over-max", "report-no-status", "report-not-in-thread", "pointer-missing", "message-not-found"} {
		if w, ok := ErrorWord(code); !ok || w.Means == "" || w.Who != "agent" {
			t.Errorf("the error word %s is missing or incomplete: %+v", code, w)
		}
	}
	for _, e := range []string{"message", "spawn", "usage"} {
		if _, ok := wordIn(LogEvents, e); !ok {
			t.Errorf("the log event %s is not in LogEvents", e)
		}
	}
	for _, k := range []string{"messages", "agents"} {
		g, ok := GeneratedKindOf(k)
		if !ok || !g.Rule || g.KeepDays == nil || *g.KeepDays != 30 || !strings.HasPrefix(g.Where, ".bonsai/local/") {
			t.Errorf("the generated kind %s: %+v", k, g)
		}
	}
	if DeclaresKeys[len(DeclaresKeys)-1] != "roles" || len(DeclaresKeys) != 8 {
		t.Errorf("DeclaresKeys %v: eight kinds, roles last", DeclaresKeys)
	}
	for _, name := range []string{"message", "messages", "agents"} {
		f, ok := Lookup("bonsai." + name + "/1")
		if !ok || !f.Writes || f.Major != 1 {
			t.Errorf("%s: %+v", name, f)
		}
	}
	out, err := (&Messages{Threads: []MessageThread{}, Malformed: []MessageMalformed{}}).Encode()
	if err != nil || !strings.Contains(string(out), `"threads": []`) || !strings.Contains(string(out), `"message": null`) {
		t.Errorf("an empty messages document: %v\n%s", err, out)
	}
	out, err = (&Agents{Agents: []Agent{}, Skipped: []AgentSkipped{}}).Encode()
	if err != nil || !strings.Contains(string(out), `"agents": []`) {
		t.Errorf("an empty agents document: %v\n%s", err, out)
	}
	// Status and the log write their new fields null until a later piece fills them (note 5.5.9, 3a).
	st, err := MustLookup("status").Document(&Status{})
	if err != nil {
		t.Fatal(err)
	}
	if v, _ := st.Get("roles"); v != nil {
		t.Errorf("status --json's roles writes %s, want null", schema.Show(v))
	}
	keys := st.Keys()
	// Set 10 adds bonsai_release after them (set10_test.go holds it).
	if strings.Join(keys[len(keys)-4:len(keys)-1], " ") != "ladder roles messages" {
		t.Errorf("status --json's set 7 and 8 fields are %v, want ladder, roles, messages", keys[len(keys)-4:len(keys)-1])
	}
	lg, err := MustLookup("log").Document(&Log{})
	if err != nil {
		t.Fatal(err)
	}
	keys = lg.Keys()
	if strings.Join(keys[len(keys)-2:], " ") != "parent usage" {
		t.Errorf("a log record ends %v, want parent, usage", keys[len(keys)-2:])
	}
	// A pack with no roles writes roles: [], and bonsai.yaml's packs entry roles: null, and messages: null.
	pk, _ := MustLookup("pack").Document(&Pack{})
	if v, _ := pk.Get("roles"); schema.Show(v) != "[]" {
		t.Errorf("a pack with no roles writes %s", schema.Show(v))
	}
	w, _ := MustLookup("workspace").Document(&Workspace{Packs: []WorkspacePack{{ID: "base"}}})
	packs, _ := w.Get("packs")
	if r, _ := packs.([]any)[0].(schema.Object).Get("roles"); r != nil {
		t.Errorf("a packs entry writes roles %s, want null", schema.Show(r))
	}
	if m, _ := w.Get("messages"); m != nil {
		t.Errorf("bonsai.yaml writes messages %s, want null", schema.Show(m))
	}
}
