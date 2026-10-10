package asks

// Filing, withdrawing, answering and reading one ask (contract §9; design/plan-5.md, 5.2.5 notes 2-4 and 8).
//
// What filing checks, each refused with exit 2 and the rule named, before anything is written: the four agent types
// only (Bless is the ladder's; a type a pack defines is refused until a pack can declare one); at most four options,
// each different, and only on a Decide; --verdict only on a Look, pass or fail; Look and Play need --task; --task
// and --doc not both, each an id matching a declared kind's id pattern, never a path; then every free-text field by
// text.go's rules (one line where it must be; a hidden character refused; redacted; its limit); then the record's
// 8,192 bytes (write.go).
//
// The key (contract §9.1): agent:<--key> ([A-Za-z0-9][A-Za-z0-9._-]{0,59}, and a key the redactor would change is
// refused: a key is forwarded and logged whole), else agent:h- and 12 hex of the SHA-256 over the type, the doc or
// task id (or "answers"), and the stored, redacted title, joined by NUL, so a key can be checked from its record
// alone. Filing an open key again appends a new file record (the newest wording stands); filing a closed key opens it
// again, as today.
//
// Answering: the same answer again (the same key and by) writes nothing and exits 0, checked first, so a retry is
// never read as ask-not-open; any other answer on a key that is not open is ask-not-open (exit 4), saying whether it
// is unknown, answered or resolved: the first answer stands. An answer from the session that asked is refused
// (answer-own-session, exit 4), comparing CLAUDE_CODE_SESSION_ID with the ask's session (contract §9.3). A Decide's
// --choice must be one of its options; a Look's --verdict is pass or fail. Withdrawing (resolve) an ask that is not
// open is ask-not-open too (exit 4: format.ErrorWords' meaning of the word), saying whether it is unknown, answered
// or resolved; today's command took a second resolve as done.

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/record"
	"github.com/LastStep/Bonsai/internal/redact"
	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// keyPart is what an agent's --key may be, and keyPattern a whole key (bonsai.ask/1's key: <source>:<part>).
var (
	keyPart    = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,59}$`)
	keyPattern = regexp.MustCompile(`^(agent|ladder):[A-Za-z0-9][A-Za-z0-9._-]{0,59}$`)
)

// Done is what a command did: the key's entry after it, and the record it wrote, or why it wrote none.
type Done struct {
	Entry   *Entry
	Written schema.Object // the record as written (bonsai.ask/1), nil when none was
	Nothing string        // why nothing was written; "" when a record was
}

// Filing is what bonsai ask files, its flags as given ("" or none for a flag not given).
type Filing struct {
	Type, Title, Why, Then string
	Task, Doc              string
	Options                []string
	Verdict                string
	Key                    string
}

// File files an ask: every check of filing, then its record and its log record (Write). kinds are the document kinds
// in force (workspace.DocKinds): --task matches the task kind's id pattern, --doc any kind's.
func File(p Place, f Filing, kinds []workspace.DocKind) (*Done, *Error) {
	types := AgentTypes()
	switch {
	case f.Type == "":
		return nil, &Error{Code: "missing-value", Exit: ExitInput, What: "ask needs --type: " + strings.Join(types, ", "),
			Next: "give --type " + strings.Join(types, ", --type ")}
	case f.Type == LadderType:
		return nil, badValue("--type "+LadderType+" is the ladder's alone: bonsai ladder files it when a ratchet count rises (contract section 9.2)",
			"file an agent's ask: --type "+strings.Join(types, ", --type "))
	case !contains(types, f.Type):
		return nil, badValue(fmt.Sprintf("--type %+q is not a type an agent files (%s); a type a pack defines is refused until a pack can declare one",
			redact.Text(f.Type), strings.Join(types, ", ")), "give --type "+strings.Join(types, ", --type ")+", as written")
	case len(f.Options) > 0 && f.Type != "Decide":
		return nil, badFlag("--option goes only on a Decide, not on "+f.Type, "leave --option out, or file a Decide")
	case len(f.Options) > OptionsMax:
		return nil, badValue(fmt.Sprintf("a Decide takes at most %d --option, not %d", OptionsMax, len(f.Options)),
			fmt.Sprintf("give at most %d options", OptionsMax))
	case f.Verdict != "" && f.Type != "Look":
		return nil, badFlag("--verdict goes only on a Look, not on "+f.Type, "leave --verdict out, or file a Look")
	case f.Verdict != "" && f.Verdict != "pass" && f.Verdict != "fail":
		return nil, badValue(fmt.Sprintf("--verdict is pass or fail, not %+q", redact.Text(f.Verdict)), "give --verdict pass or --verdict fail")
	case (f.Type == "Look" || f.Type == "Play") && f.Task == "":
		return nil, &Error{Code: "missing-value", Exit: ExitInput, What: "a " + f.Type + " needs --task: the task it is about",
			Next: "give --task <task id>"}
	case f.Task != "" && f.Doc != "":
		return nil, badFlag("ask takes --task or --doc, not both", "give the one the ask is about")
	}
	if f.Task != "" {
		if e := checkID("--task", f.Task, kinds, "task"); e != nil {
			return nil, e
		}
	}
	if f.Doc != "" {
		if e := checkID("--doc", f.Doc, kinds, ""); e != nil {
			return nil, e
		}
	}
	title, e := required(titleField, f.Title)
	if e != nil {
		return nil, e
	}
	why, e := required(whyField, f.Why)
	if e != nil {
		return nil, e
	}
	then, e := optional(thenField, f.Then)
	if e != nil {
		return nil, e
	}
	options := []string{}
	for _, o := range f.Options {
		if strings.TrimSpace(o) == "" {
			return nil, badValue("an --option is empty", "give each --option its text")
		}
		c, e := clean(optionField, o)
		if e != nil {
			return nil, e
		}
		if contains(options, c) {
			return nil, badValue("two --option are the same once redacted: each must differ", "give each option different words")
		}
		options = append(options, c)
	}
	key, e := fileKey(f.Key, f.Type, f.Task, f.Doc, title)
	if e != nil {
		return nil, e
	}
	typ := f.Type
	rec := &format.Ask{ID: record.NewID(), At: p.at(), Op: OpFile, Key: key, Workspace: p.Workspace, Source: SourceAgent,
		Session: ptr(p.session()), Type: &typ, Task: ptr(f.Task), Doc: ptr(f.Doc), Title: &title, Why: &why,
		ThenText: ptr(then), Options: options, Verdict: ptr(f.Verdict)}
	doc, e := Write(p, rec)
	if e != nil {
		return nil, e
	}
	return &Done{Entry: &Entry{Key: key, State: Open, Filed: rec, Docs: [2]schema.Object{doc, nil}}, Written: doc}, nil
}

// fileKey is a filed ask's key: the agent's --key, or the hash of type, target and stored title (contract §9.1).
func fileKey(given, typ, task, doc, title string) (string, *Error) {
	if given != "" {
		if !keyPart.MatchString(given) {
			return "", badValue(fmt.Sprintf("--key %+q is not a key: a letter or digit, then letters, digits, dot, dash or underscore, "+
				"at most 60 in all", redact.Text(given)), "give --key a name such as colour-choice, or leave it out for a key made from the ask")
		}
		if redact.Text(given) != given {
			return "", badValue("--key looks like a secret to the redactor, and a key is logged and forwarded whole",
				"give --key a plain name, or leave it out for a key made from the ask")
		}
		return SourceAgent + ":" + given, nil
	}
	target := doc
	if target == "" {
		target = task
	}
	if target == "" {
		target = "answers"
	}
	sum := sha256.Sum256([]byte(typ + "\x00" + target + "\x00" + title))
	return SourceAgent + ":h-" + hex.EncodeToString(sum[:])[:12], nil
}

// checkID holds a --task or --doc id to the document kinds: never a path, at most IDMax characters, matching the
// task kind's id pattern (kind "task") or any kind's (kind "").
func checkID(flag, id string, kinds []workspace.DocKind, kind string) *Error {
	if strings.ContainsAny(id, `/\`) || strings.HasSuffix(id, ".md") || len([]rune(id)) > IDMax {
		return badValue(fmt.Sprintf("%s %+q is a path or too long: an ask names a document by its id, never its path", flag, redact.Text(id)),
			"give "+flag+" the document's id, such as T-0901")
	}
	var names []string
	for _, k := range kinds {
		if k.ID == "" || (kind != "" && k.Kind != kind) {
			continue
		}
		names = append(names, k.Kind+" ("+k.ID+")")
		re, err := regexp.Compile(k.ID)
		if err != nil {
			continue
		}
		if loc := re.FindStringIndex(id); loc != nil && loc[0] == 0 && loc[1] == len(id) {
			return nil
		}
	}
	what := "matches no declared kind's id pattern"
	if kind != "" {
		what = "is not a " + kind + "'s id"
	}
	return badValue(fmt.Sprintf("%s %+q %s: %s", flag, redact.Text(id), what, strings.Join(names, ", ")),
		"give "+flag+" an id of the document the ask is about (bonsai status --json lists the kinds and their id patterns)")
}

// checkKey holds a key given on the command line to bonsai.ask/1's shape, and gives its source.
func checkKey(key string) (string, *Error) {
	if !keyPattern.MatchString(key) {
		return "", badValue(fmt.Sprintf("%+q is not an ask's key (agent:<part> or ladder:<part>)", redact.Text(key)),
			"run: bonsai asks --all, to list the keys")
	}
	source, _, _ := strings.Cut(key, ":")
	return source, nil
}

// lookup reads one key's entry; an unknown key is ask-not-open (exit 4).
func lookup(p Place, key string) (*Entry, *Error) {
	b, err := Read(p.Local.Main, key)
	if err != nil {
		return nil, &Error{Code: "read-failed", Exit: ExitRuntime, What: ".bonsai/local/asks/ cannot be read: " + oneLine(err),
			Next: "check that .bonsai/local/asks/ can be read, then run the command again"}
	}
	e := b.Get(key)
	if e == nil {
		return nil, &Error{Code: "ask-not-open", Exit: ExitState, What: "no ask has the key " + key + ": nothing was written",
			Next: "run: bonsai asks --all, to list the keys"}
	}
	return e, nil
}

// Status gives one key's entry (bonsai ask --status): its latest file record and, once closed, the record that
// closed it. An unknown key is ask-not-open (exit 4).
func Status(p Place, key string) (*Entry, *Error) {
	if _, e := checkKey(key); e != nil {
		return nil, e
	}
	return lookup(p, key)
}

// List gives every open ask, or with all every key, newest first, and how many lines did not read.
func List(p Place, all bool) ([]*Entry, int, *Error) {
	b, err := Read(p.Local.Main, "")
	if err != nil {
		return nil, 0, &Error{Code: "read-failed", Exit: ExitRuntime, What: ".bonsai/local/asks/ cannot be read: " + oneLine(err),
			Next: "check that .bonsai/local/asks/ can be read, then run the command again"}
	}
	return b.Entries(all), b.Unreadable, nil
}

// Resolve withdraws an agent's open ask (bonsai ask --resolve): a resolve record and its log record. An unknown,
// answered or resolved one is ask-not-open (exit 4), nothing written.
func Resolve(p Place, key string) (*Done, *Error) {
	source, e := checkKey(key)
	if e != nil {
		return nil, e
	}
	if source != SourceAgent {
		return nil, badValue("ask --resolve withdraws an agent's ask; "+key+" is the ladder's (step 5.4)", "give an agent's key, agent:<part>")
	}
	entry, e := lookup(p, key)
	if e != nil {
		return nil, e
	}
	switch entry.State {
	case Resolved:
		return nil, &Error{Code: "ask-not-open", Exit: ExitState,
			What: fmt.Sprintf("%s is resolved already (at %s): nothing was written", key, entry.Closed.At),
			Next: "nothing to withdraw: run bonsai asks to see the open asks"}
	case Answered:
		return nil, &Error{Code: "ask-not-open", Exit: ExitState,
			What: fmt.Sprintf("%s is answered already (by %s, at %s): nothing was written", key, ascii(answeredBy(entry.Closed)), entry.Closed.At),
			Next: "read the answer: bonsai ask --status " + key}
	}
	rec := closing(p, OpResolve, key, source)
	doc, e := Write(p, rec)
	if e != nil {
		return nil, e
	}
	entry.State, entry.Closed, entry.Docs[1] = Resolved, rec, doc
	return &Done{Entry: entry, Written: doc}, nil
}

// Answering is what bonsai answer gives, its flags as given ("" for a flag not given).
type Answering struct {
	Choice, Verdict, Words, By, Via string
}

// Answer answers an open ask (bonsai answer): an answer record and its log record. It grants nothing.
func Answer(p Place, key string, a Answering) (*Done, *Error) {
	source, e := checkKey(key)
	if e != nil {
		return nil, e
	}
	by, e := optional(byField, a.By)
	if e != nil {
		return nil, e
	}
	if by == "" {
		by = Terminal
	}
	via, e := optional(viaField, a.Via)
	if e != nil {
		return nil, e
	}
	choice, e := optional(choiceField, a.Choice)
	if e != nil {
		return nil, e
	}
	words, e := optional(wordsField, a.Words)
	if e != nil {
		return nil, e
	}
	switch {
	case a.Verdict != "" && a.Verdict != "pass" && a.Verdict != "fail":
		return nil, badValue(fmt.Sprintf("--verdict is pass or fail, not %+q", redact.Text(a.Verdict)), "give --verdict pass or --verdict fail")
	case choice == "" && a.Verdict == "" && words == "":
		return nil, &Error{Code: "missing-value", Exit: ExitInput, What: "an answer needs --choice, --verdict or --words",
			Next: "give the answer: --choice for a Decide, --verdict for a Look, --words for any"}
	}
	entry, e := lookup(p, key)
	if e != nil {
		return nil, e
	}
	if entry.State == Answered && answeredBy(entry.Closed) == by {
		return &Done{Entry: entry, Nothing: fmt.Sprintf("%s was answered already by %s, at %s: nothing was written", key, ascii(by),
			entry.Closed.At)}, nil
	}
	switch entry.State {
	case Answered:
		return nil, &Error{Code: "ask-not-open", Exit: ExitState,
			What: fmt.Sprintf("%s is answered already (by %s, at %s): the first answer stands, and nothing was written", key,
				ascii(answeredBy(entry.Closed)), entry.Closed.At),
			Next: "read the answer: bonsai ask --status " + key}
	case Resolved:
		return nil, &Error{Code: "ask-not-open", Exit: ExitState,
			What: fmt.Sprintf("%s was withdrawn (resolved at %s): nothing was written", key, entry.Closed.At),
			Next: "nothing to answer: run bonsai asks to see the open asks"}
	}
	if s := entry.Filed.Session; s != nil && p.Session != "" && *s == p.Session {
		return nil, &Error{Code: "answer-own-session", Exit: ExitState,
			What: "the session that filed " + key + " cannot answer it: a person, or a session that did not ask, answers " +
				"(contract section 9.3); nothing was written",
			Next: "ask the person to answer it in their own terminal: bonsai answer " + key}
	}
	typ := val(entry.Filed.Type)
	switch {
	case choice != "" && typ != "Decide":
		return nil, badFlag("--choice answers a Decide; "+key+" is "+article(typ), "leave --choice out; give --words")
	case choice != "" && !contains(entry.Filed.Options, choice):
		return nil, badValue(fmt.Sprintf("--choice %+q is not one of the ask's options: %s", choice, quoteAll(entry.Filed.Options)),
			"give --choice one of the options as written (bonsai ask --status "+key+" lists them), or answer with --words")
	case a.Verdict != "" && typ != "Look":
		return nil, badFlag("--verdict answers a Look; "+key+" is "+article(typ), "leave --verdict out; give --words")
	}
	rec := closing(p, OpAnswer, key, source)
	rec.Answer = &format.AskAnswer{By: by, Via: ptr(via), Choice: ptr(choice), Verdict: ptr(a.Verdict), Words: ptr(words)}
	doc, e := Write(p, rec)
	if e != nil {
		return nil, e
	}
	entry.State, entry.Closed, entry.Docs[1] = Answered, rec, doc
	return &Done{Entry: entry, Written: doc}, nil
}

// closing lays out a resolve or answer record: its common fields, every field of the ask itself null.
func closing(p Place, op, key, source string) *format.Ask {
	return &format.Ask{ID: record.NewID(), At: p.at(), Op: op, Key: key, Workspace: p.Workspace, Source: source,
		Session: ptr(p.session()), Options: []string{}}
}

func contains(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}

// article names an ask's type with its article: "a Decide", "an Answer".
func article(typ string) string {
	switch {
	case typ == "":
		return "of no type"
	case strings.ContainsRune("AEIOU", rune(typ[0])):
		return "an " + ascii(typ)
	}
	return "a " + ascii(typ)
}

func quoteAll(list []string) string {
	if len(list) == 0 {
		return "it has none"
	}
	q := make([]string, len(list))
	for i, s := range list {
		q[i] = fmt.Sprintf("%+q", s)
	}
	return strings.Join(q, ", ")
}

// answeredBy is who an answer record names, "" when it holds no answer (a record written by hand).
func answeredBy(rec *format.Ask) string {
	if rec.Answer == nil {
		return ""
	}
	return rec.Answer.By
}
