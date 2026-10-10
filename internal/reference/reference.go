// Package reference builds the reference page of lists, docs/reference/lists.md (plan-5 5.1.10; spec section 6:
// "Every list has one home"): every list Bonsai owns, its values, whether it is closed or open, where it is defined,
// and the command that prints it in a project. The page is written by `go generate ./...` from the code's own tables
// and the formats set's embedded schemas, so it holds no list of its own: Page is the one function that builds it, the
// generator (gen) writes its bytes, and TestPageIsCurrent rebuilds it and fails on any difference from the committed
// file, naming `go generate ./...` as the fix.
//
// The page is byte-stable: LF line ends, ASCII only, forward slashes, no map-order iteration; where the code's order is
// meaningful (a schema's, a table's, Claude Code's measured key orders) it is kept, and nothing is sorted that has an
// order of its own. A list that a schema holds as an enum is found by walking every schema, so a new enum in a new
// schema appears on the page without a change here (the generator's list of copies it skips is below). A list the
// code holds in a table is added here by one entry. Standard library only.
//
//go:generate go run ./gen ../../docs/reference/lists.md
package reference

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/LastStep/Bonsai/internal/engine"
	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/status"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// Path is where the page lives, from the repository's root.
const Path = "docs/reference/lists.md"

// Page builds the page.
func Page() []byte {
	p := &page{}
	p.header()
	p.section("Exit codes", "closed", "format.ExitCodes (internal/format/exit.go)", "`bonsai --help --json` (exit_codes, and each word's exits)",
		"Spec section 3. Every word's own table says which of these it returns and what each means for it.")
	p.exitCodes()
	p.section("Error words", "open", "format.ErrorWords (internal/format/error.go)",
		"`bonsai check --schema bonsai.error`, and `bonsai --help --json` (error_words)",
		"The word of a refusal's `error.code` in every word's `--json`; who is who usually takes the next step. A word is added, never renamed or taken out.")
	p.words(format.ErrorWords, "Who")
	p.section("Check words", "open", "format.CheckWords (internal/format/check.go)", "`bonsai check --schema bonsai.check`",
		"The code of a finding (makes `check` exit 1, and is a problem in `status --json`) or a warning (printed, never the exit code); who is who usually takes the next step.")
	p.words(format.CheckWords, "Kind", "Who")
	p.section("Check --pack words", "open", "format.PackCheckWords (internal/format/check.go)", "`bonsai check --schema bonsai.check`",
		"The code of a finding from `bonsai check --pack <folder>`; every one is a finding, and none appears in a project's check.")
	p.words(format.PackCheckWords, "Kind", "Who")
	p.section("Check words later steps build", "closed", "engine.checkLater (internal/engine/check.go)", "none",
		"Spec section 6's findings and warnings that are not built yet; each joins the check words above when its step lands.")
	p.later()
	p.section("Ask types", "open", "format.AskTypes (internal/format/ask.go), held to the ask schema's description of `type`",
		"`bonsai check --schema bonsai.ask`", "A pack may define its own type (contract section 9.1).")
	p.words(format.AskTypes)
	p.section("Active task: reasons there is none", "closed", "workspace.ActiveReasons (internal/workspace/active.go), held to formats/README.md's table \"Why there is none\"",
		"none; `bonsai status --active --json` prints the sentence of the one that applies (`why`)", "The codes of contract section 13's function and its fixtures' answers.")
	p.reasons()
	p.section("Generated kinds", "closed", "format.GeneratedKinds (internal/format/generated.go), held to bonsai.workspace/1's `generated` properties",
		"none; `bonsai init` writes each kind's default rule into bonsai.yaml's `generated:` section", "Spec section 6, \"Generated files\".")
	p.generated()
	p.section("Bonsai's document kinds", "closed", "workspace.BonsaiKindNames (internal/workspace/documents.go)", "none",
		"The kinds of Bonsai's own documents, in contract section 7.3's order; no pack may declare one of these names, and a pack's own kinds are the open part.")
	p.values(workspace.BonsaiKindNames())
	p.section("Pack file kinds", "closed", "workspace.PackFileKinds (internal/workspace/pack.go), held to the lock schema's list",
		"`bonsai check --schema bonsai.pack`", "The kinds a pack gives its files in bonsai/pack.yaml (`files[].kind`).")
	p.values(workspace.PackFileKinds)
	p.section("Needs kinds", "open", "status.NeedKinds (internal/status/needs.go)", "`bonsai status --json` (needs[].kind)",
		"What the workspace needs from a machine; a reader shows a kind it does not know as other.")
	p.words(status.NeedKinds)
	p.section("Claude Code states", "closed", "engine.ClaudeStates (internal/engine/claude.go)", "`bonsai status --full --json` (checks.claude_code.state)",
		"Claude Code's version against the floor (spec section 7).")
	p.words(engine.ClaudeStates)
	p.section("Plugin results", "open", "engine.PluginResults (internal/engine/lists.go)", "`bonsai init --json`, `bonsai update --json`, `bonsai unlink --json` (plugins[].result)",
		"What the plugin step did for one pack's plugin on this machine; a reader shows a result it does not know as other.")
	p.words(engine.PluginResults)
	p.orders()
	p.section("Log events", "open", "not built yet (step 5.2.0: the log's events in one Go table beside the log's Go type)", "none yet (step 5.2.0: `bonsai check --schema bonsai.log`)",
		"Contract section 8.2's event names; the page lists them from the table once it exists.")
	p.section("Log categories", "open", "not built yet (step 5.2.0: the log's categories in one Go table beside the log's Go type)", "none yet (step 5.2.0: `bonsai check --schema bonsai.log`)",
		"Contract section 8.2's tool categories; a pack's own are `<namespace>.<Name>`.")
	p.schemaLists()
	// The contents go between the header and the sections.
	body := p.b.String()
	cut := strings.Index(body, "\n## ")
	var toc strings.Builder
	toc.WriteString("\n## Contents\n\n")
	for _, n := range p.names {
		toc.WriteString("- " + n + "\n")
	}
	return []byte(body[:cut] + toc.String() + body[cut:])
}

type page struct {
	b     strings.Builder
	names []string // the sections' names, for the contents
}

func (p *page) line(format string, a ...any) { fmt.Fprintf(&p.b, format+"\n", a...) }

func (p *page) header() {
	p.line("# Reference: the lists")
	p.line("")
	p.line("Every list Bonsai owns: its values when Bonsai owns them, whether it is closed or open, where it is defined and the")
	p.line("command that prints it in a project. A closed list never grows inside a format's major; an open list may, and a reader")
	p.line("shows a value it does not know as `other`. Every list has one home (spec section 6): this page is read from the code.")
	p.line("")
	p.line("**This page is generated. Never edit it by hand.** It is written by `go generate ./...` from the tables in Bonsai's")
	p.line("code and the schemas in `formats/`; a test rebuilds it and fails on any difference, naming that command. Change a list")
	p.line("in its home, run `go generate ./...` from the repository's root, and commit the page with the change.")
	p.line("For a program, `bonsai --help --json` is the same knowledge for the command words, their flags and exit codes, and the")
	p.line("error words.")
}

// section opens a list's entry: its name, closed or open, its home and the command that prints it, then a note.
func (p *page) section(name, kind, home, command, note string) {
	p.names = append(p.names, name)
	p.line("")
	p.line("## %s", name)
	p.line("")
	p.line("- %s list.", strings.ToUpper(kind[:1])+kind[1:])
	p.line("- Defined in: %s.", home)
	p.line("- Printed in a project by: %s.", command)
	if note != "" {
		p.line("- %s", note)
	}
	p.line("")
}

func cell(s string) string {
	s = strings.ReplaceAll(s, "\n", " ")
	return strings.ReplaceAll(s, "|", "\\|")
}

func (p *page) exitCodes() {
	p.line("| Code | In short | Means | For |")
	p.line("|---|---|---|---|")
	for _, e := range format.ExitCodes {
		p.line("| %d | %s | %s | %s |", e.Code, cell(e.Short), cell(e.Means), cell(e.Applies))
	}
}

// words prints a table of words; extra names the Kind and Who columns to add.
func (p *page) words(ws []format.Word, extra ...string) {
	head, rule := "| Word |", "|---|"
	for _, x := range extra {
		head += " " + x + " |"
		rule += "---|"
	}
	p.line("%s Means |", head)
	p.line("%s---|", rule)
	for _, w := range ws {
		row := "| `" + w.Word + "` |"
		for _, x := range extra {
			v := w.Who
			if x == "Kind" {
				v = w.Kind
			}
			if v == "" {
				v = "-"
			}
			row += " " + v + " |"
		}
		p.line("%s %s |", row, cell(w.Means))
	}
}

func (p *page) values(vs []string) {
	var q []string
	for _, v := range vs {
		q = append(q, "`"+v+"`")
	}
	p.line("Values, in order: %s.", strings.Join(q, ", "))
}

func (p *page) later() {
	p.line("| Word | Built in | Means |")
	p.line("|---|---|---|")
	for _, l := range engine.CheckLaterWords() {
		p.line("| `%s` | %s | %s |", l.Word, cell(l.Step), cell(l.Means))
	}
}

func (p *page) reasons() {
	p.line("| Code | Means |")
	p.line("|---|---|")
	for _, r := range workspace.ActiveReasons {
		p.line("| `%s` | %s |", r.Code, cell(r.Means))
	}
}

func (p *page) generated() {
	p.line("| Kind | Where | What | Default | Never cleaned | Takes a rule |")
	p.line("|---|---|---|---|---|---|")
	for _, k := range format.GeneratedKinds {
		p.line("| `%s` | %s | %s | %s | %s | %s |", k.Kind, cell(k.Where), cell(k.What), cell(k.Default), cell(k.Never), map[bool]string{true: "yes", false: "no"}[k.Rule])
	}
}

func (p *page) orders() {
	for _, o := range engine.ClaudeOrders() {
		p.section("Claude Code's key order: "+o.Where, "closed", "engine."+o.Name+" (internal/engine/settings.go)", "none",
			"The order Claude Code writes them in, measured on Claude Code 2.1.294 and 2.1.295 (step 5.1.7); a key Bonsai adds goes where Claude Code would put it. The order is the value.")
		p.values(o.Keys)
	}
}

// ---------------------------------------------------------------- the lists the schemas hold

// listNames names the schema enums the plan names, by "<format>:<path>"; any other enum is named by its path.
var listNames = map[string]string{
	"task:status":                "Task statuses",
	"run:outcome":                "Run outcomes",
	"labels:labels[].kind":       "Label value kinds",
	"labels:labels[].set_by":     "Label set_by",
	"ask:op":                     "Ask ops",
	"ask:source":                 "Ask sources",
	"lock:files.*.kind":          "Lock file kinds",
	"status:active_task.how":     "Active task: how it was found",
	"error:next.who":             "Who takes a next step",
	"log:decision":               "Guard decisions",
	"changes:command":            "Changes: commands",
	"changes:settings[].change":  "Changes: a settings line's change",
	"changes:runs_code[].change": "Changes: a code item's change",
	"status:packs[].state":       "Status: a pack's state",
	"status:status_writes":       "Status: status_writes",
	"ladder:mode":                "Ladder modes",
	"ladder:rungs[].kind":        "Ladder rung kinds",
	"memory:kind":                "Memory kinds",
	"sessions:sessions[].kind":   "Sessions: a row's kind",
	"sessions:hours[].kind":      "Sessions: an hours row's kind",
	"labels:labels[].items":      "Label items kinds",
	"ask:verdict":                "Ask verdicts",
}

// skipped are the enums the page does not list again: a copy of a list shown at its home (the formats test holds each
// copy equal to its home), or a list shown in a section above.
var skipped = map[string]string{
	"help:error_words[].who":     "the error words' who: error:next.who's list",
	"status:error.next.who":      "a copy of error:next.who",
	"check:findings[].next.who":  "a copy of error:next.who",
	"check:warnings[].next.who":  "a copy of error:next.who",
	"check:error.next.who":       "a copy of error:next.who",
	"changes:error.next.who":     "a copy of error:next.who",
	"changes:plugins[].next.who": "a copy of error:next.who",
	"workspace:ladder[].kind":    "a copy of ladder:rungs[].kind",
	"status:lanes[].close":       "a copy of lanes:lanes[].close, shown under Lane rules",
	"lanes:lanes[].close":        "shown under Lane rules",
	"pack:files[].kind":          "shown as Pack file kinds",
	"ask:answer.verdict":         "a copy of ask:verdict",
}

type enum struct {
	format, path string
	values       []any
}

// enums walks a schema for every enum, in the schema's order.
func enums(name string, s schema.Object) []enum {
	var out []enum
	var walk func(n schema.Object, path string)
	walk = func(n schema.Object, path string) {
		if v, ok := n.Get("enum"); ok {
			if list, ok := v.([]any); ok {
				out = append(out, enum{name, path, list})
			}
		}
		if p, ok := n.Get("properties"); ok {
			if props, ok := p.(schema.Object); ok {
				for _, m := range props {
					if sub, ok := m.Value.(schema.Object); ok {
						join := m.Key
						if path != "" {
							join = path + "." + m.Key
						}
						walk(sub, join)
					}
				}
			}
		}
		if it, ok := n.Get("items"); ok {
			if sub, ok := it.(schema.Object); ok {
				walk(sub, path+"[]")
			}
		}
		if ap, ok := n.Get("additionalProperties"); ok {
			if sub, ok := ap.(schema.Object); ok {
				walk(sub, path+".*")
			}
		}
	}
	walk(s, "")
	return out
}

func valueText(v any) string {
	switch x := v.(type) {
	case nil:
		return "`null`"
	case string:
		return "`" + x + "`"
	case bool:
		return "`" + strconv.FormatBool(x) + "`"
	}
	return "`" + schema.Show(v) + "`"
}

// schemaLists prints Lane rules, then every other enum of every schema.
func (p *page) schemaLists() {
	lanes := format.MustLookup("lanes").Schema()
	var closeValues []string
	for _, e := range enums("lanes", lanes) {
		if e.path == "lanes[].close" {
			for _, v := range e.values {
				closeValues = append(closeValues, valueText(v))
			}
		}
	}
	p.section("Lane rules", "closed", "formats/schemas/lanes.schema.json (`lanes[].approve_first`, `lanes[].close`)", "`bonsai check --schema bonsai.lanes`",
		"The two rules Bonsai understands (contract section 6); everything else a lane says is for agents to read.")
	p.line("- `approve_first`: `true`, `false` (a move to running, verify or done needs the task to have read approved since it last read todo or plan).")
	p.line("- `close`: %s (who may move the task to done).", strings.Join(closeValues, ", "))
	for _, f := range format.All {
		for _, e := range enums(f.Name, f.Schema()) {
			key := e.format + ":" + e.path
			if _, skip := skipped[key]; skip {
				continue
			}
			name := listNames[key]
			if name == "" {
				name = f.Name + ": " + e.path
			}
			var vs []string
			for _, v := range e.values {
				vs = append(vs, valueText(v))
			}
			p.section(name, "closed", "formats/schemas/"+f.Name+".schema.json (`"+e.path+"`)", "`bonsai check --schema "+f.ID()+"`", "")
			p.line("Values: %s.", strings.Join(vs, ", "))
		}
	}
}
