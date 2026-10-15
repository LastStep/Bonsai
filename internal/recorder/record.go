package recorder

// One payload as one bonsai.log/1 record (contract §8.1-§8.3; design/plan-5.md, 5.2.4 notes 4-7). The events are
// Events, their one home.
//
// Beside the fields every record has (record.New: format, id, at, workspace, session, agent as claude-code, checkout
// as the folder's name, branch, task and role from BONSAI_TASK and BONSAI_ROLE, labels as {}), each record names the
// agent's event (agent_event), the subagent's id and type when the payload names them, and remote when
// CLAUDE_CODE_BRIDGE_SESSION_ID matches its pattern. Then, by event:
//
//	Claude Code event              event                      also filled
//	SessionStart (hook start)      session_start              source, model, target (the active task found),
//	                                                          bonsai_path, bonsai_sha256
//	UserPromptSubmit               prompt                     kind: user or task-notification; never the words
//	PreToolUse, PermissionRequest  tool_start, permission     tool, category, target, tool_use_id, input_hash;
//	                                                          for AskUserQuestion, text as its questions
//	PostToolUse                    tool_end                   the same, and ok true
//	PostToolUseFailure             tool_fail                  the same, ok false, kind interrupt or error
//	Notification                   notice                     kind as its type, text as its message
//	SubagentStart, SubagentStop    subagent_start, _stop      on the start, target as the active task found
//	Stop                           stop                       nothing more
//	SessionEnd                     session_end                reason
//
// After a SubagentStop's or a SessionEnd's record, the agent's usage records: its token counts read from its own
// transcript, one record for each model, or one unknown with its reason (usage.go).
//
// The input hash (contract §8.1) is hookin.InputHash, keyed by the home's salt, over the tool input the hook adapter
// decoded; a call's events carry the same input, so a call's records pair. With no salt (workspace.Salt failed) it is
// null and the record is still written; a tool input the strict decoder refuses (a duplicate key, bad UTF-8) has no
// hash and no target.

import (
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/hookin"
	"github.com/LastStep/Bonsai/internal/record"
	"github.com/LastStep/Bonsai/internal/redact"
	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// Event is one Claude Code event the recorder writes a record for (contract §8.3).
type Event struct {
	Agent string // the Claude Code hook event: the record's agent_event
	Log   string // the log's event (contract §8.2)
	What  string // what its record holds, as the preview's sentence for its hook line says it
}

// Events are the Claude Code events Bonsai records, their one home: SessionStart (written by bonsai hook start) and
// the ten bonsai hook record writes, in the order the engine writes their hook lines (engine.ownHooks reads this
// table). A word of the log's events (format.LogEvents) for each.
var Events = []Event{
	{"SessionStart", "session_start", "the session's start, with the active task and bonsai's own path and hash"},
	{"UserPromptSubmit", "prompt", "each prompt submitted (its kind only, never its words)"},
	{"PreToolUse", "tool_start", "each tool call before it runs (the tool, its target reduced and redacted, a hash of its input)"},
	{"PermissionRequest", "permission", "each permission the session asks for"},
	{"PostToolUse", "tool_end", "each tool call that finished"},
	{"PostToolUseFailure", "tool_fail", "each tool call that failed or was interrupted"},
	{"Notification", "notice", "each notice Claude Code gives a person (redacted)"},
	{"SubagentStart", "subagent_start", "each subagent run's start and the active task it found"},
	{"SubagentStop", "subagent_stop", "each subagent run's end, and its token counts read from its transcript (never its text)"},
	{"Stop", "stop", "the end of each turn"},
	{"SessionEnd", "session_end", "the session's end, and its token counts read from its transcript (never its text)"},
}

// logEvent gives the log's event for a Claude Code event, "" for one Bonsai does not record.
func logEvent(agent string) string {
	for _, e := range Events {
		if e.Agent == agent {
			return e.Log
		}
	}
	return ""
}

// recorded reports an event `bonsai hook record` writes: every event of Events but SessionStart, whose record
// `bonsai hook start` writes (one process makes the session's file first).
func recorded(event string) bool {
	return logEvent(event) != "" && event != "SessionStart"
}

// RemoteEnv names the variable Claude Code sets for a hook while Remote Control is connected: the session's Remote
// Control id, kept in remote only when it matches remotePattern (the log schema's), and never redacted.
const RemoteEnv = "CLAUDE_CODE_BRIDGE_SESSION_ID"

var remotePattern = regexp.MustCompile(`^session_[A-Za-z0-9]{10,60}$`)

// writeLog appends a record to the main checkout's log (a variable for the tests).
var writeLog = record.WriteLog

// salt gives the input hash's key (a variable for the tests).
var salt = workspace.Salt

// build lays out the record of one payload, or nil for an event the recorder does not write. proj is the session's
// checkout (the project's folder), local where it writes, act the active task, read once when a record names it.
func build(o Options, p *hookin.Payload, proj string, cfg *workspace.Config, local workspace.Local, act *active) *format.Log {
	ev := logEvent(p.Event)
	if ev == "" {
		return nil
	}
	l := base(o, p, cfg, local, ev, o.Now())
	switch {
	case p.Event == "UserPromptSubmit":
		kind := "user"
		if strings.HasPrefix(strings.TrimLeftFunc(p.Prompt, unicode.IsSpace), "<task-notification>") {
			kind = "task-notification"
		}
		l.Kind = ptr(kind)
	case hookin.ToolEvent(p.Event):
		toolFields(l, p, proj)
	case p.Event == "Notification":
		l.Kind = text(p.NotificationType)
		l.Text = ptr(redact.Capped(p.Message, redact.TextCap))
	case p.Event == "SubagentStart":
		l.Target = act.target()
	case p.Event == "SessionStart":
		l.Source = text(p.Source)
		l.Model = text(p.Model)
		l.Target = act.target()
	case p.Event == "SessionEnd":
		l.Reason = text(p.Reason)
	}
	return l
}

// base lays out a record of the log's event ev for a payload, at the time given, with the fields every record of a
// hook's payload has: the common fields (record.New), agent_event, the subagent's id and type when the payload names
// them, checkout and branch redacted, and remote. build and the usage records (usage.go) start from it.
func base(o Options, p *hookin.Payload, cfg *workspace.Config, local workspace.Local, ev string, at time.Time) *format.Log {
	l := record.New(ev, record.Common{Workspace: cfg.ID, Local: local, Session: p.Session, Agent: "claude-code",
		Getenv: o.Getenv, Redact: redact.Text, At: at})
	l.AgentEvent = ptr(p.Event)
	l.SubagentID = ptr(hookin.ID(p.AgentID))
	l.SubagentType = text(p.AgentType)
	l.Checkout = text(str(l.Checkout))
	l.Branch = text(str(l.Branch))
	if r := o.Getenv(RemoteEnv); remotePattern.MatchString(r) {
		l.Remote = ptr(r)
	}
	return l
}

// toolFields fills a tool event's fields: tool, category, target, tool_use_id, input_hash, ok, and a failure's kind.
func toolFields(l *format.Log, p *hookin.Payload, proj string) {
	l.Tool = text(p.Tool)
	l.Category = ptr(Category(p.Tool, nil))
	l.ToolUseID = ptr(hookin.ID(p.ToolUseID))
	switch p.Event {
	case "PostToolUse":
		l.OK = boolPtr(true)
	case "PostToolUseFailure":
		l.OK = boolPtr(false)
		l.Kind = ptr("error")
		if p.IsInterrupt {
			l.Kind = ptr("interrupt")
		}
	}
	if p.Input == nil {
		return
	}
	in := p.InputObject()
	l.Category = ptr(Category(p.Tool, in))
	if p.Tool == "AskUserQuestion" {
		l.Text = ptr(redact.Question(in))
	} else {
		l.Target = ptr(redact.Target(p.Tool, in, proj, p.Cwd))
	}
	if key, err := salt(); err == nil {
		l.InputHash = ptr(hookin.InputHash(key, p.Input))
	}
}

// Category gives a tool's log category (contract §8.2): Read; Grep and Glob as Search; Edit, MultiEdit and
// NotebookEdit as Edit; Write; Bash and PowerShell as Shell, or Ladder when the command's head is `bonsai ladder`;
// Agent and Task as Agent; WebFetch and WebSearch as Web; an mcp__ tool as MCP; anything else Other. input is the
// tool's input, read for a shell command's head; nil reads no command.
func Category(tool string, input schema.Object) string {
	switch tool {
	case "Read":
		return "Read"
	case "Grep", "Glob":
		return "Search"
	case "Edit", "MultiEdit", "NotebookEdit":
		return "Edit"
	case "Write":
		return "Write"
	case "Bash", "PowerShell":
		if input != nil && redact.CommandHead(input.String("command"), tool == "PowerShell") == "bonsai ladder" {
			return "Ladder"
		}
		return "Shell"
	case "Agent", "Task":
		return "Agent"
	case "WebFetch", "WebSearch":
		return "Web"
	}
	if strings.HasPrefix(tool, "mcp__") {
		return "MCP"
	}
	return "Other"
}

// active is the active task of one hook run (contract §13, BONSAI_TASK counting: the hook's project is the session's
// own), read from the main checkout's task folder once, when a record or the opening context first names it: hook
// start's record and its context share one read.
type active struct {
	o     Options
	cfg   *workspace.Config
	local workspace.Local
	read  bool
	task  workspace.ActiveTask
	err   error
}

// get reads the active task the first time it is asked for, and gives that reading every time.
func (a *active) get() (workspace.ActiveTask, error) {
	if !a.read {
		a.read = true
		a.task, a.err = findActive(a.o, a.cfg, a.local)
	}
	return a.task, a.err
}

// target is the active task's id as a record's target, nil for none.
func (a *active) target() *string {
	t, err := a.get()
	if err != nil || t.ID == "" {
		return nil
	}
	return ptr(redact.Capped(t.ID, redact.TargetCap))
}

// findActive finds the active task: tasks read from the main checkout's task folder, its bonsai.yaml's
// documents.task (the session's own checkout's when main's cannot be read). A variable for the tests, which count
// its reads.
var findActive = func(o Options, cfg *workspace.Config, local workspace.Local) (workspace.ActiveTask, error) {
	dir := taskDir(cfg)
	if local.Main != local.Root {
		if main, err := workspace.LoadConfig(local.Main); err == nil && taskDir(main) != "" {
			dir = taskDir(main)
		}
	}
	return workspace.Active(workspace.ActiveInput{Main: local.Main, TaskDir: dir, Env: o.Getenv(workspace.TaskEnv), EnvCounts: true})
}

// taskDir reads bonsai.yaml's documents.task from the file as read (the lean read keeps it in Doc).
func taskDir(cfg *workspace.Config) string {
	if cfg == nil || cfg.Doc == nil {
		return ""
	}
	d, _ := cfg.Doc.Get("documents")
	type getter interface{ Get(string) (any, bool) }
	g, ok := d.(getter)
	if !ok {
		return ""
	}
	t, _ := g.Get("task")
	s, _ := t.(string)
	return s
}

// text gives a free-text field as written: redacted, or nil for "".
func text(s string) *string {
	if s == "" {
		return nil
	}
	return ptr(redact.Text(s))
}

// ptr gives "" as nil (written null), any other text as a pointer to it.
func ptr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func str(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func boolPtr(b bool) *bool { return &b }
