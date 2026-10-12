package asks

// The ladder's own asks (contract §9.1, §9.2; design/plan-5.md, 5.4.4 notes 2 and 3): bonsai ladder files one Bless ask
// when a green, clean climb in the main checkout, on the base branch at its HEAD, measures a ratchet above its floor,
// and resolves it when that offer has gone. Only the ladder files Bless: File refuses the type to agents, and Resolve
// refuses a ladder: key to them. Both go through Write, the same function as every ask, so each record has its ask log
// record.
//
//   - FileBless files the ask: source ladder, type Bless, key ladder:bless-<task>, the task, the title, why and then
//     (each by text.go's rules: one line where it must be, a hidden character refused, redacted, its limit) and data,
//     the offer (contract §9.1: {task, head, result_sha, rises}). Filing an open key again appends a new file record,
//     so a later climb's numbers replace the earlier ones; filing a closed key opens it again, as an agent's does.
//   - ResolveLadder resolves the ladder's open ask: a resolve record, every field of the ask null, as an agent's
//     withdrawal is (today's resolve record). An unknown, answered or resolved key is ask-not-open (exit 4), nothing
//     written.

import (
	"fmt"
	"strings"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/record"
	"github.com/LastStep/Bonsai/internal/redact"
	"github.com/LastStep/Bonsai/internal/schema"
)

// BlessPrefix is how the key of every Bless ask begins: ladder:bless-<task> (contract §9.1).
const BlessPrefix = SourceLadder + ":bless-"

// BlessKey is the key of a task's Bless ask, ladder:bless-<task>; "" when the task's id cannot make a key
// (bonsai.ask/1's key: a letter or digit, then letters, digits, dot, dash or underscore, at most 60 after the colon),
// or the redactor would change it (a key is logged and forwarded whole).
func BlessKey(task string) string {
	key := BlessPrefix + task
	if task == "" || !keyPattern().MatchString(key) || redact.Text(key) != key {
		return ""
	}
	return key
}

// Bless is the Bless ask as bonsai ladder files it: the task it offers to bless, the words a person reads, and the
// offer itself (contract §9.1's data: {task, head, result_sha, rises: [{name, was, now, new_tests}]}).
type Bless struct {
	Task  string
	Title string // one line, at most TitleMax characters once redacted
	Why   string // at most WhyMax characters once redacted; may hold lines
	Then  string // one line, at most ThenMax characters once redacted; "" for none
	Data  schema.Object
}

// The ladder's free-text fields: the same rules as an agent's, named as the ladder's ask names them.
var (
	blessTitle = field{name: "the Bless ask's title", max: TitleMax}
	blessWhy   = field{name: "the Bless ask's why", max: WhyMax, lines: true}
	blessThen  = field{name: "the Bless ask's then", max: ThenMax}
)

// FileBless files the ladder's Bless ask for b.Task (this file's comment): its record and its log record (Write). A task
// whose id cannot make a key, a field that breaks its rule, or no data is refused (bad-value or missing-value, exit 2),
// nothing written. A record written whose log record was not (partly-written, exit 3) gives its Done with the error:
// the record stands.
func FileBless(p Place, b Bless) (*Done, *Error) {
	key := BlessKey(b.Task)
	if key == "" {
		return nil, badValue(fmt.Sprintf("the task %+q cannot make a Bless ask's key (%s<task>: a letter or digit, then letters, digits, dot, "+
			"dash or underscore, at most 60 after the colon)", redact.Text(b.Task), BlessPrefix), "give the task an id such as T-0901")
	}
	if b.Data == nil {
		return nil, &Error{Code: "missing-value", Exit: ExitInput, What: "a Bless ask carries its offer in data, and none was given",
			Next: "give the offer: {task, head, result_sha, rises}"}
	}
	title, e := required(blessTitle, b.Title)
	if e != nil {
		return nil, e
	}
	why, e := required(blessWhy, b.Why)
	if e != nil {
		return nil, e
	}
	then, e := optional(blessThen, b.Then)
	if e != nil {
		return nil, e
	}
	typ, task := LadderType, b.Task
	rec := &format.Ask{ID: record.NewID(), At: p.at(), Op: OpFile, Key: key, Workspace: p.Workspace, Source: SourceLadder,
		Session: ptr(p.session()), Type: &typ, Task: &task, Title: &title, Why: &why, ThenText: ptr(then), Options: []string{},
		Data: b.Data}
	doc, e := Write(p, rec)
	if doc == nil {
		return nil, e
	}
	return &Done{Entry: &Entry{Key: key, State: Open, Filed: rec, Docs: [2]schema.Object{doc, nil}}, Written: doc}, e
}

// ResolveLadder resolves the ladder's own open ask, key ladder:<part>, once its offer has gone (this file's comment): a
// resolve record and its log record. An agent's key is bad-value (exit 2: bonsai ask --resolve withdraws it); an
// unknown, answered or resolved one is ask-not-open (exit 4); nothing written either way. A partly-written resolve (exit
// 3) gives its Done with the error, as FileBless does.
func ResolveLadder(p Place, key string) (*Done, *Error) {
	source, e := checkKey(key)
	if e != nil {
		return nil, e
	}
	if source != SourceLadder {
		return nil, badValue("the ladder resolves its own asks; "+key+" is an agent's", "withdraw it with: bonsai ask --resolve "+key)
	}
	return resolve(p, key, source)
}

// IsBless reports whether key is a Bless ask's key as BlessKey makes it (ladder:bless-<task>), and gives its task. A
// record written by hand with any other key under the prefix (a path, say) is not one.
func IsBless(key string) (string, bool) {
	task, ok := strings.CutPrefix(key, BlessPrefix)
	return task, ok && BlessKey(task) == key
}
