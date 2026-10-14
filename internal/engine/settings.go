package engine

// Bonsai's lines in the project's .claude/settings.json (file kind keys, contract §14: "Bonsai's own entries in a
// project-owned JSON file"; spec §5, §6, §7). The rest of the file is the project's and is kept, value for value.
//
// A line is one entry Bonsai writes, as the preview names it (spec §6, "The preview names every settings line"):
//
//	hook         a hook command under hooks.<event>, in a group with its matcher (spec §7; a pack's hooks)
//	deny         a rule in permissions.deny (Bonsai's own, ownDeny; each never_edit path in bonsai.yaml as
//	             Edit(/<path>), neverEditRule; a pack's deny, as the pack gives it)
//	key          autoMemoryEnabled: false (spec §10) and disableAllHooks: false (spec §7)
//	marketplace  the inline marketplace under extraKnownMarketplaces, bonsai-<workspace name>-<8 hex>, each pack's
//	             plugin pinned to its locked commit by sha, with no version (spec §5)
//	plugin       <pack id>@<marketplace>: true under enabledPlugins (spec §5)
//
// What Bonsai last wrote. The lock records one fingerprint for the file, the SHA-256 of Bonsai's lines as written
// (linesHash). The lines themselves are not stored: the engine finds them in the file (claim) by what it can tell
// is its own, the lines it would write from the lock's packs at their locked commits and bonsai.yaml, and lines no
// person writes: autoMemoryEnabled and disableAllHooks, a marketplace named bonsai-..., a plugin enabled from one,
// and a hook line that runs `bonsai hook`. Only then is the fingerprint compared, so an edit of one of Bonsai's
// lines reads as an edit, and a line a person added beside them is never Bonsai's. When a path left never_edit in
// bonsai.yaml since the last update, its old Edit rule is found by trying the file's other Edit rules against the
// lock's fingerprint, each set of them in turn: only the set whose fingerprint matches exactly is Bonsai's, so a rule
// a person wrote is never taken for one of Bonsai's.
//
// The file's key order (step 5.1.7). Claude Code writes this file too (a project-scope plugin install, an uninstall),
// with JSON.stringify: two-space indent, a final newline, and the keys it knows in its own order, the ones it does not
// know after them in their order. As measured on Claude Code 2.1.294 and 2.1.295 (records/runs, 5.1.7), that order
// is stable: a second install writes the same bytes. So Bonsai changes only its own entries and keeps every other
// key, value and character where the file has it:
//   - an entry of Bonsai's that stays keeps its place, and its bytes when its value is unchanged (a marketplace in
//     Claude Code's key order is left as it is);
//   - a marketplace or plugin entry whose name changed (a new commit) takes the old entry's place;
//   - a key Bonsai adds goes where Claude Code would put it (claudeKeyOrder): after the last key of the file that
//     Claude Code writes before it; a deny list goes into permissions the same way (claudePermissionsOrder); a
//     marketplace is written in Claude Code's order (claudeMarketplaceOrder);
//   - strings are written as JSON.stringify writes them (schema.EncodeUTF8), so a project's non-ASCII text is kept.
// A new file is therefore written in the order Claude Code writes, and Claude Code's first install leaves it as it is.
// In a file that held settings of the project's own, Claude Code's first install may still move the project's keys
// into its own order once (keys Claude Code knows that claudeKeyOrder does not list, a file not in Claude Code's
// order); from then on the order stays, and Bonsai's writes keep it.

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/recorder"
	"github.com/LastStep/Bonsai/internal/schema"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// SettingsFile is the project's shared Claude Code settings file, the one settings file the engine writes.
const SettingsFile = ".claude/settings.json"

// Bonsai's own hook lines (spec §7's table, with design/plan-5.md's 5.2.4 note 1, 5.3.4 note 6 and 5.3.6 notes 2 and
// 5): the guard's, the stop gate's, the start line and the recorder's. Each line's command is in Bonsai's form
// (HookCommand), naming the installed bonsai's places; the constants below are each line's event, its timeout and, for
// the guard, its matcher.
const (
	GuardEvent   = "PreToolUse"
	GuardTimeout = "10"
	// GuardOnFailure is the guard's line's onFailure (Claude Code's hooks reference: "onFailure": "block", from
	// 2.1.295): a hook that cannot start, ends with an exit other than 0 or 2, times out or gives output Claude Code
	// cannot read then blocks the call as exit 2 does, so a line run in PowerShell (a Windows machine without Git Bash,
	// where no PowerShell reads the line) or a bonsai that cannot start refuses the call. Measured 11 Oct (step 5.3.6b,
	// real -p sessions): Claude Code 2.1.294 on Windows reads a settings file whose hook carries the field and ignores
	// the field (a hook exiting 2 still blocked, the file's other hooks ran, and a hook that could not start let the
	// call through, as before the field); 2.1.296 on WSL blocks such a call. It has no effect on Stop hooks or async
	// ones (the hooks reference), so only the guard's line carries it.
	GuardOnFailure = "block"

	// The stop gate's line (step 5.3.4 note 6) has no matcher, since Stop takes none; its || exit 2 turns any other
	// ending of bonsai into exit 2, as the guard's does. Measured on Claude Code 2.1.296 (WSL, a -p session): with no
	// bonsai on the PATH (/bin/sh: 1: bonsai: not found), the line kept the session going until Claude Code's cap
	// ended the turn. Claude Code reads a Stop hook's exit 2 whose stderr says "No such file or directory" as a hook
	// that failed to start and lets the session end (its hooks reference; measured the same day), so the lines test
	// for bonsai before they run it, and their own words for a bonsai that is not there use none of those words
	// (HookCommand).
	StopEvent   = "Stop"
	StopTimeout = "10"

	StartEvent   = "SessionStart"
	StartTimeout = "10"

	RecordTimeout = "10"
	// The SessionEnd line is synchronous (an async hook is killed as the session exits, so its line would be lost),
	// with 5 s: its timeout also raises Claude Code's 1.5 s budget for SessionEnd hooks to 5 s.
	RecordEndEvent   = "SessionEnd"
	RecordEndTimeout = "5"
)

// The installed bonsai's two places (spec §3; Rohan's answer (a) to step 5.3's first question, 9 Oct; plan-5 5.3.6
// note 2), as the shell that runs a hook line names them, in the order the lines try them: Windows' place first,
// C:\Program Files\Bonsai\bonsai.exe as Git Bash names it, which only an administrator can place; then WSL's and
// Linux's, which only root can place. Windows' comes first because on Linux a /c folder at the root needs root to make,
// while under Git Bash /usr/local/bin is a folder inside Git's own install, which a Git for Windows installed for one
// user lets that user write (check's git-user finding). They are the same on every machine and name no user, the
// one named exception to check 2's rule that no committed file holds a full path. Their one home: the hook lines
// (HookCommand), check, and from step 5.6 the installers read them. The PATH is never read.
const (
	PlaceWindows = "/c/Program Files/Bonsai/bonsai.exe"
	PlaceLinux   = "/usr/local/bin/bonsai"
)

// GuardTool is one tool the guard's line matches.
type GuardTool struct {
	Name  string // the tool's name, as a hook matcher names it
	Why   string // what it does that the guard judges
	Roles bool   // matched only where the lock declares roles (plan-5 5.5.11 note 2): a tool that starts or resumes an agent
}

// GuardTools are the tools the guard's line matches (plan-5 5.3.6 note 5), their one home: every tool that writes a
// file or runs a command, read from Claude Code's tools reference on 2.1.296 (11 Oct), which says a hook matcher names
// a tool by its exact name, and its hooks reference, which reads a matcher of names and | as a list of exact names.
// The other tools there read, search, fetch, ask, plan, schedule or message (Agent's and Workflow's subagents make
// their own tool calls, which this line meets). MultiEdit is not on the 2.1.296 page; it stays for a Claude Code that
// still has it, since the hook adapter reads it and a name no tool has matches nothing.
//
// Where the lock declares roles (step 5.5, "Role rules and stall stops"; Rohan's "A. Two lines", 14 Oct), the line
// also matches the tools that start or resume an agent (Roles): Agent, its older name Task, SendMessage and Workflow,
// so the guard judges who may call whom; its form, timeout and onFailure are the same. Everywhere else the line is as
// before, and such a project pays nothing new.
var GuardTools = []GuardTool{
	{Name: "Edit", Why: "edits a file"},
	{Name: "Write", Why: "creates or overwrites a file"},
	{Name: "MultiEdit", Why: "edits a file several times in one call (older Claude Code)"},
	{Name: "NotebookEdit", Why: "edits a Jupyter notebook's cells"},
	{Name: "Bash", Why: "runs a shell command"},
	{Name: "PowerShell", Why: "runs a PowerShell command"},
	{Name: "Monitor", Why: "runs a command in the background, under Bash's permission rules"},
	{Name: "Agent", Why: "starts a subagent of a type (only where the lock declares roles: who may call whom)", Roles: true},
	{Name: "Task", Why: "Agent's older name (only where the lock declares roles)", Roles: true},
	{Name: "SendMessage", Why: "resumes a subagent, or writes to the main session or another one (only where the lock declares roles)", Roles: true},
	{Name: "Workflow", Why: "runs a script that starts agents of its own (only where the lock declares roles)", Roles: true},
}

// GuardMatcher is the guard's line's matcher where the lock declares no roles: GuardTools' names joined by |, the
// roles' tools left out. GuardRolesMatcher is the line's matcher where it does: every name.
var (
	GuardMatcher      = guardMatcher(false)
	GuardRolesMatcher = guardMatcher(true)
)

func guardMatcher(roles bool) string {
	var names []string
	for _, t := range GuardTools {
		if roles || !t.Roles {
			names = append(names, t.Name)
		}
	}
	return strings.Join(names, "|")
}

// The guard's second line (plan-5 5.5.11 note 2; step 5.5, "The guard's two lines"), written only where the lock
// declares roles: bonsai hook guard --second on every tool the first line does not match (reads, searches, skills, the
// web, MCP tools, the question tool), in the form of the lines that never block when bonsai is missing (no || exit 2,
// no onFailure), so a missing or broken bonsai lets the call through. When bonsai runs, it refuses only past an agent's
// budget, and lets a call through on a fault of its own (internal/guard). Its matcher names every other tool by one
// negative lookahead a name of GuardTools, ^(?!Edit$)(?!Write$)...(?!Workflow$).*: Claude Code 2.1.300 (WSL, 15 Oct)
// took ^(?!Bash$).* as a pattern that met every tool but Bash; no | is used, which Claude Code may read as a list of
// exact names. The guard also passes at once, uncounted, a tool the first line judges, should both meet one.
const (
	SecondWord    = "guard --second"
	SecondTimeout = "10"
)

// SecondMatcher is the second line's matcher (above).
var SecondMatcher = secondMatcher()

func secondMatcher() string {
	var b strings.Builder
	b.WriteString("^")
	for _, t := range GuardTools {
		b.WriteString("(?!" + t.Name + "$)")
	}
	b.WriteString(".*")
	return b.String()
}

// The command of each of Bonsai's own hook lines, in this build's form: GuardCommand and StopCommand block (exit 2
// when bonsai is not at its place), SecondCommand, StartCommand and RecordCommand never do.
var (
	GuardCommand  = HookCommand("guard", true, hookPlaces()...)
	SecondCommand = HookCommand(SecondWord, false, hookPlaces()...)
	StopCommand   = HookCommand("stop", true, hookPlaces()...)
	StartCommand  = HookCommand("start", false, hookPlaces()...)
	RecordCommand = HookCommand("record", false, hookPlaces()...)
)

// ownHooks is this build's table of Bonsai's own hook lines, in spec §7's order, where the lock declares no roles;
// ownRolesHooks where it does (step 5.5.11: the guard's line matching the roles' tools too, and the second line after
// it). Spec §7 lists four kinds, and this build writes each: the guard (PreToolUse), the stop gate (Stop, step 5.3.4),
// start (SessionStart) and the recorder (step 5.2.4): SessionStart's record is written by hook start itself, so the
// recorder's lines are ten, nine async and SessionEnd's synchronous. A project linked before a line was added gets it
// at its next update only with --allow-exec (a hook line added runs code, step 5.1.1); bonsai check warns until then
// (own-hooks). A project's first pack with roles changes the guard's line and adds the second: a change to Bonsai's
// own lines, listed under "Runs code".
var (
	ownHooks      = buildOwnHooks(false)
	ownRolesHooks = buildOwnHooks(true)
)

// ownHooksFor gives the own hook lines for a lock that declares roles, or not.
func ownHooksFor(roles bool) []Line {
	if roles {
		return ownRolesHooks
	}
	return ownHooks
}

func buildOwnHooks(roles bool) []Line {
	guard := Line{Kind: "hook", Origin: "bonsai", Own: true, Event: GuardEvent, Matcher: GuardMatcher, Command: GuardCommand, Timeout: GuardTimeout,
		OnFailure: GuardOnFailure, Why: "Checks every file edit and shell command against the task's rights; refuses each one while " +
			"bonsai is not installed at its place, or cannot start."}
	ls := []Line{guard}
	if roles {
		ls[0].Matcher = GuardRolesMatcher
		ls[0].Why = "Checks every file edit, shell command and agent start against the task's rights and the roles' rules " +
			"(who may call whom, what each role may change, each agent's budget); refuses each one while bonsai is not installed " +
			"at its place, or cannot start."
		ls = append(ls, Line{Kind: "hook", Origin: "bonsai", Own: true, Event: GuardEvent, Matcher: SecondMatcher, Command: SecondCommand,
			Timeout: SecondTimeout, Why: "Counts every other tool call (reads, searches, skills, the web, MCP tools, questions) against " +
				"its agent's budget and refuses one past it; lets each through while bonsai is not installed at its place, or cannot start."})
	}
	ls = append(ls,
		Line{Kind: "hook", Origin: "bonsai", Own: true, Event: StopEvent, Command: StopCommand, Timeout: StopTimeout,
			Why: "Keeps a session whose BONSAI_TASK names a task from ending until a green local ladder result proves the " +
				"checkout's HEAD (bonsai ladder --task <id>); does nothing in a session that names no task. While bonsai is not " +
				"installed at its place, it refuses each session's end until Claude Code ends the turn after its cap of blocks."},
		Line{Kind: "hook", Origin: "bonsai", Own: true, Event: StartEvent, Command: StartCommand, Timeout: StartTimeout,
			Why: "Gives each session (new, resumed, cleared or compacted) the active task, its last ladder result and this " +
				"machine's labels, and records the session's start, with bonsai's own path and hash, in .bonsai/local/log/; never blocks."})
	// The recorder's lines, one per event it records (recorder.Events, their one home), in its order: async, but
	// SessionEnd's; SessionStart's record is hook start's.
	for _, e := range recorder.Events {
		switch e.Agent {
		case StartEvent:
			continue
		case RecordEndEvent:
			ls = append(ls, Line{Kind: "hook", Origin: "bonsai", Own: true, Event: e.Agent, Command: RecordCommand, Timeout: RecordEndTimeout,
				Why: "Records " + e.What + " in the project's log, .bonsai/local/log/; waits at most 5 s and never blocks."})
		default:
			ls = append(ls, Line{Kind: "hook", Origin: "bonsai", Own: true, Event: e.Agent, Command: RecordCommand, Timeout: RecordTimeout,
				Async: true, Why: "Records " + e.What + " in the project's log, .bonsai/local/log/ (never committed), in the background; never blocks."})
		}
	}
	return ls
}

// ownDeny is this build's table of Bonsai's own deny rules (design/plan-5.md, step 5.3.5; spec §6 and §7, "Deny
// rules the engine writes"; contract §3, §10.6): one over the files only Bonsai writes in each project, its log, asks
// and ladder results in .bonsai/local/; one for each table check --write rebuilds (tableFiles, from
// format.GeneratedKinds, their one home); and one over Bonsai's home, ~/.bonsai/, whose machine records the guard
// trusts to find a worktree's main checkout (5.3.1 note 2). Bonsai writes all of them as a program, never through the
// file tools, so the rules do not stop it (bonsai check --write, run through an agent's shell tool, still writes the
// tables); they stop an agent's Edit, Write and NotebookEdit and the shell routes Claude Code checks against Edit
// rules (a redirect, and tee and sed as the permissions reference says), and from step 5.3.2 the guard refuses the
// file tools there too. A home moved by BONSAI_HOME is not under ~/.bonsai: the guard alone covers it. A deny rule
// runs no code, so --yes alone writes these lines (step 5.1.1, rule 1), at a first link and when a project linked by
// an older build takes them at its next update; unlink takes them out; taking a pack out keeps them, since no pack
// gives them.
//
// The form (step 5.3.5 note 3; the permissions reference: a rule's path anchors at the filesystem root with //, at the
// user's home with ~/, at the session's starting folder with /, at its current folder with none), as measured in real
// -p sessions on Claude Code 2.1.296 (WSL) and 2.1.294 (Windows): //**/ matches the folder at any depth, on Windows on
// every drive and in any letter case, so a session started in the project's top folder meets it after a cd too, and a
// worktree session, whose .bonsai/local/ is the main checkout's, outside its own folder, meets it there; the tables
// take the same form, so a worktree session cannot edit its own tables or main's. ~/ held for a Write and a redirect
// into a folder of the home (tried on a scratch folder, never on a real ~/.bonsai).
//
// Why these are the engine's lines and not the base pack's (5.3.5 note 2; spec §7 lists the .bonsai/local/ rule among
// base's walls): from step 5.4 the stop gate trusts the ladder results in .bonsai/local/, and Bonsai links its own repo
// at 5.4, before base exists (5.5); and a project may link without base. When step 5.5 writes base's walls, these
// lines stay here, their one home, and base leaves them out; the walls round secret files stay base's, and base keeps
// Read(~/.bonsai/salt), a secret's wall.
var ownDeny = buildOwnDeny()

func buildOwnDeny() []Line {
	ls := []Line{{Kind: "deny", Origin: "bonsai", Own: true, Rule: "Edit(//**/" + LocalDir + "/**)",
		Why: "Agents can never write in " + LocalDir + "/, in this checkout or the main one (the log, asks and ladder " +
			"results): only Bonsai writes there, as a program."}}
	for _, t := range tableFiles() {
		ls = append(ls, Line{Kind: "deny", Origin: "bonsai", Own: true, Rule: "Edit(//**/" + t + ")",
			Why: "Agents can never edit " + t + " by hand: Bonsai rebuilds that table, run: bonsai check --write."})
	}
	return append(ls, Line{Kind: "deny", Origin: "bonsai", Own: true, Rule: "Edit(~/" + homeFolder + "/**)",
		Why: "Agents can never write in Bonsai's home, ~/" + homeFolder + "/ (this machine's records and settings for each " +
			"workspace, which the guard trusts): only Bonsai's own commands write there."})
}

// neverEditRule is the deny rule for a never_edit path of bonsai.yaml (spec §7, "Deny rules the engine writes"):
// Edit(/<path>), anchored at the session's starting folder, the project's top folder (step 5.3.5 note 3). Measured in
// real sessions (Claude Code 2.1.296 on WSL, 2.1.294 on Windows): the form part 3 wrote, Edit(<path>), anchors at the
// session's current folder, so after a shell cd into a subfolder neither it nor a bare file name stopped a Write or a
// redirect to the path, while Edit(/<path>) still did. A project linked before takes the new form at its next
// update --yes, shown as each rule changed.
func neverEditRule(path string) string { return "Edit(/" + path + ")" }

// neverEditSlot starts a never_edit rule's slot: the path, whatever the rule's form.
const neverEditSlot = "never_edit:"

// homeFolder is the Bonsai home's folder in the user's home when BONSAI_HOME is not set (workspace.Home).
const homeFolder = ".bonsai"

// tableFiles lists the tables Bonsai rebuilds, from format.GeneratedKinds (their one home): the generated kinds that
// are one file in .bonsai/, outside .bonsai/local/, in the order the kinds are listed.
func tableFiles() []string {
	var out []string
	for _, k := range format.GeneratedKinds {
		if strings.HasPrefix(k.Where, ".bonsai/") && !strings.HasSuffix(k.Where, "/") &&
			!strings.HasPrefix(k.Where, LocalDir+"/") {
			out = append(out, k.Where)
		}
	}
	return out
}

// OwnHookLines lists this build's own hook lines as the preview shows them, in the order the engine writes them, for a
// lock that declares no roles; OwnRolesHookLines for one that does.
func OwnHookLines() []string { return hookTexts(ownHooks) }

// OwnRolesHookLines lists this build's own hook lines for a lock that declares roles (ownRolesHooks).
func OwnRolesHookLines() []string { return hookTexts(ownRolesHooks) }

func hookTexts(ls []Line) []string {
	var out []string
	for _, l := range ls {
		out = append(out, l.Text())
	}
	return out
}

// Line is one of Bonsai's entries in .claude/settings.json.
type Line struct {
	Kind      string // hook, deny, key, marketplace or plugin
	Origin    string // where it comes from: bonsai (this build's own), bonsai.yaml (never_edit), old (an old Bonsai hook line), or a pack id
	Own       bool   // a line of Bonsai's own (ownHooks and ownDeny, or a `bonsai hook` line claim finds): set only here, never from a pack
	Slot      string // two lines with one slot are one line changed, not one removed and one added
	Event     string // hook: the Claude Code hook event
	Matcher   string // hook: the group's matcher, "" for none
	Command   string // hook: the shell line
	Timeout   string // hook: the timeout in seconds, as JSON number text; "" for none
	Async     bool   // hook: runs without waiting
	OnFailure string // hook: its onFailure, "block" on the guard's line (GuardOnFailure); "" for none
	Rule      string // deny: the permission rule
	Name      string // key, marketplace, plugin: its key in its object
	Value     any    // key, marketplace, plugin: its JSON value
	Why       string // the sentence the preview prints
}

// canon is the line's identity: two lines are the same line when their canon is.
func (l Line) canon() string {
	switch l.Kind {
	case "hook":
		c := strings.Join([]string{"hook", l.Event, l.Matcher, l.Command, l.Timeout, strconv.FormatBool(l.Async)}, "\t")
		if l.OnFailure != "" {
			// Only a line that has it says so: a line without it keeps the canon every build before step 5.3.6 took the
			// lock's fingerprint over.
			c += "\tonFailure=" + l.OnFailure
		}
		return c
	case "deny":
		return "deny\t" + l.Rule
	}
	return l.Kind + "\t" + l.Name + "\t" + schema.Show(l.Value)
}

// match is the line's identity when claim looks for it in the file: its canon, with every JSON object's keys in
// sorted order. Claude Code rewrites the settings file in an order of its own when it writes it (a project-scope
// plugin install moves the marketplace's owner after its plugins), and that is no edit of Bonsai's lines.
func (l Line) match() string {
	switch l.Kind {
	case "hook", "deny":
		return l.canon()
	}
	return l.Kind + "\t" + l.Name + "\t" + schema.Show(sortedKeys(l.Value))
}

// sortedKeys gives a JSON value with every object's keys in sorted order.
func sortedKeys(v any) any {
	switch x := v.(type) {
	case schema.Object:
		out := make(schema.Object, len(x))
		for i, m := range x {
			out[i] = schema.Member{Key: m.Key, Value: sortedKeys(m.Value)}
		}
		sort.SliceStable(out, func(i, j int) bool { return out[i].Key < out[j].Key })
		return out
	case []any:
		out := make([]any, len(x))
		for i, e := range x {
			out[i] = sortedKeys(e)
		}
		return out
	}
	return v
}

// Text is the line as the preview shows it.
func (l Line) Text() string {
	switch l.Kind {
	case "hook":
		s := l.Event
		if l.Matcher != "" {
			s += " (" + l.Matcher + ")"
		}
		s += ": " + l.Command
		if l.Async {
			s += " [async]"
		}
		if l.OnFailure != "" {
			s += " [onFailure: " + l.OnFailure + "]"
		}
		return s
	case "deny":
		return l.Rule
	case "marketplace":
		return l.Name + ": " + marketplaceText(l.Value)
	case "plugin":
		if l.Value == true {
			return l.Name + " enabled"
		}
		return l.Name + ": " + schema.Show(l.Value)
	}
	return l.Name + ": " + schema.Show(l.Value)
}

// marketplaceText names each plugin of an inline marketplace and its pinned commit.
func marketplaceText(v any) string {
	o, _ := v.(schema.Object)
	src, _ := o.Get("source")
	so, _ := src.(schema.Object)
	pl, _ := so.Get("plugins")
	list, _ := pl.([]any)
	var parts []string
	for _, p := range list {
		po, _ := p.(schema.Object)
		ps, _ := po.Get("source")
		pso, _ := ps.(schema.Object)
		parts = append(parts, po.String("name")+" at "+short(pso.String("sha")))
	}
	if len(parts) == 0 {
		return schema.Show(v)
	}
	return strings.Join(parts, ", ")
}

// entry is the line's JSON value as written: a hook's command object.
func (l Line) hookEntry() schema.Object {
	o := schema.Object{{Key: "type", Value: "command"}, {Key: "command", Value: l.Command}}
	if l.Timeout != "" {
		o = append(o, schema.Member{Key: "timeout", Value: json.Number(l.Timeout)})
	}
	if l.OnFailure != "" {
		o = append(o, schema.Member{Key: "onFailure", Value: l.OnFailure})
	}
	if l.Async {
		o = append(o, schema.Member{Key: "async", Value: true})
	}
	return o
}

func objectOf(kv ...any) schema.Object {
	o := schema.Object{}
	for i := 0; i+1 < len(kv); i += 2 {
		o = append(o, schema.Member{Key: kv[i].(string), Value: kv[i+1]})
	}
	return o
}

// packLines is what a pack gives the settings file: its plugin wiring, and its hook lines and deny rules when the
// pack was read at its commit (known).
type packLines struct {
	id, source, folder, commit string
	known                      bool
	hooks                      []workspace.HookEntry
	deny                       []workspace.DenyEntry
	roles                      bool // the pack declares roles: Bonsai's own lines are then ownRolesHooks (step 5.5.11)
}

// declaresRoles reports a pack's declares naming roles, for packLines.roles.
func declaresRoles(d *format.Declares) bool { return d != nil && len(d.Roles) > 0 }

// manifestRoles reports a pack.yaml declaring roles, for packLines.roles.
func manifestRoles(m *workspace.Pack) bool { return m != nil && m.Full != nil && len(m.Full.Roles) > 0 }

// MarketplaceName is the inline marketplace's name (spec §5): bonsai-<workspace name>-<8 hex>, the hex the first
// 8 characters of the SHA-256 of the packs' pins (Pin), each followed by a line feed, in bonsai.yaml's order. Two
// checkouts at the same commits and folders share it; a checkout whose update moved a pack to another commit or
// folder gets its own, so a plugin installed from the old folder is never taken for the new one's (step 5.1.1's
// verifier: "already installed" compared the commit, not the folder).
func MarketplaceName(name string, pins []string) string {
	var b strings.Builder
	for _, c := range pins {
		b.WriteString(c + "\n")
	}
	sum := sha256.Sum256([]byte(b.String()))
	return "bonsai-" + name + "-" + hex.EncodeToString(sum[:])[:8]
}

// Pin is what a pack adds to its marketplace's name: its locked commit, and, for a pack in a folder of its
// repository, a space and the folder. A pack at its repository's root pins its commit alone, so its marketplace keeps
// the name spec §5 gives it (the locked commits).
func Pin(commit, folder string) string {
	if folder == "" {
		return commit
	}
	return commit + " " + folder
}

// buildLines gives every line Bonsai writes for a workspace and its packs, in the order it writes them: the two
// keys, the deny rules (Bonsai's own, then never_edit, then each pack's), the hook lines (Bonsai's own, its roles'
// form when a pack declares roles, then each pack's), the marketplace, then each pack's plugin. A line that two sources give is written once, as the first
// gives it: a pack's rule equal to one of Bonsai's own stays Bonsai's.
func buildLines(cfg *workspace.Config, packs []packLines) []Line {
	ls := []Line{
		{Kind: "key", Origin: "bonsai", Name: "autoMemoryEnabled", Value: false,
			Why: "Turns Claude Code's own auto memory off in this project; Bonsai keeps the project's memory itself."},
		{Kind: "key", Origin: "bonsai", Name: "disableAllHooks", Value: false,
			Why: "Keeps hooks on in this project, so Bonsai's guard runs even when a user setting turns hooks off."},
	}
	ls = append(ls, ownDeny...)
	for _, p := range cfg.NeverEdit {
		ls = append(ls, Line{Kind: "deny", Origin: "bonsai.yaml", Rule: neverEditRule(p), Slot: neverEditSlot + p,
			Why: "Agents can never edit " + p + ": it is in never_edit in bonsai.yaml."})
	}
	for _, p := range packs {
		for _, d := range p.deny {
			ls = append(ls, Line{Kind: "deny", Origin: p.id, Rule: d.Rule, Why: p.id + ": " + d.Why})
		}
	}
	roles := false
	for _, p := range packs {
		roles = roles || p.roles
	}
	ls = append(ls, ownHooksFor(roles)...)
	for _, p := range packs {
		for _, h := range p.hooks {
			ls = append(ls, Line{Kind: "hook", Origin: p.id, Event: h.Event, Matcher: h.Matcher, Command: h.Command,
				Why: p.id + ": " + h.Why})
		}
	}
	if len(packs) > 0 {
		var pins []string
		var plugins []any
		for _, p := range packs {
			pins = append(pins, Pin(p.commit, p.folder))
			plugins = append(plugins, objectOf("name", p.id, "source", pluginSource(p.source, p.folder, p.commit)))
		}
		market := MarketplaceName(cfg.Name, pins)
		ls = append(ls, Line{Kind: "marketplace", Origin: "bonsai", Name: market,
			Value: objectOf("source", objectOf("source", "settings", "name", market, "owner", objectOf("name", "Bonsai"),
				"plugins", plugins)),
			Why: "Tells Claude Code where to fetch this workspace's packs as plugins, each pinned to its locked commit."})
		for _, p := range packs {
			ls = append(ls, Line{Kind: "plugin", Origin: p.id, Name: p.id + "@" + market, Value: true,
				Why: "Turns on the " + p.id + " plugin from that marketplace, so its roles and skills load as " + p.id + ":<name>."})
		}
	}
	return slotted(dedupe(ls))
}

func dedupe(ls []Line) []Line {
	seen := map[string]bool{}
	var out []Line
	for _, l := range ls {
		if c := l.canon(); !seen[c] {
			seen[c] = true
			out = append(out, l)
		}
	}
	return out
}

// slotted gives each line its slot: a key, the marketplace and a plugin by name; a hook by its origin, event,
// matcher and place among those; a deny rule by itself.
func slotted(ls []Line) []Line {
	n := map[string]int{}
	for i, l := range ls {
		if l.Slot != "" {
			continue
		}
		switch l.Kind {
		case "key":
			ls[i].Slot = "key:" + l.Name
		case "marketplace":
			ls[i].Slot = "marketplace"
		case "plugin":
			ls[i].Slot = "plugin:" + strings.SplitN(l.Name, "@", 2)[0]
		case "deny":
			ls[i].Slot = "deny:" + l.Rule
		case "hook":
			k := "hook:" + l.Origin + ":" + l.Event + ":" + l.Matcher
			if l.Own {
				k = "hook:@own:" + l.Event + ":" + l.Matcher // no pack id holds an @
			}
			ls[i].Slot = k + ":" + strconv.Itoa(n[k])
			n[k]++
		}
	}
	return ls
}

// linesHash is the fingerprint the lock records for the settings file: the SHA-256 of the lines' canon forms,
// sorted and each once, each followed by a line feed. The order lines sit in the file does not count.
func linesHash(ls []Line) string {
	set := map[string]bool{}
	for _, l := range ls {
		set[l.canon()] = true
	}
	keys := make([]string, 0, len(set))
	for k := range set {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	sum := sha256.Sum256([]byte(strings.Join(keys, "\n") + "\n"))
	return hex.EncodeToString(sum[:])
}

// settingsDoc is the settings file as read.
type settingsDoc struct {
	root   schema.Object // the document; empty when the file is missing
	exists bool
	raw    []byte
}

func readSettings(root string) (*settingsDoc, error) {
	p := filepath.Join(root, filepath.FromSlash(SettingsFile))
	raw, err := os.ReadFile(p)
	if errors.Is(err, fs.ErrNotExist) {
		return &settingsDoc{root: schema.Object{}}, nil
	}
	next := "fix .claude/settings.json (Claude Code reads it too), then run the command again"
	if err != nil {
		return nil, errorf("read-failed", ExitRuntime, "check the file's permissions, then run the command again",
			"%s cannot be read: %v", SettingsFile, err)
	}
	v, err := schema.Decode(bytes.TrimPrefix(raw, []byte("\xef\xbb\xbf")))
	if err != nil {
		return nil, errorf("bad-file", ExitInput, next, "%s is not JSON Bonsai reads: %v", SettingsFile, err)
	}
	o, ok := v.(schema.Object)
	if !ok {
		return nil, errorf("bad-file", ExitInput, next, "%s is not a JSON object", SettingsFile)
	}
	for _, k := range []string{"permissions", "hooks", "extraKnownMarketplaces", "enabledPlugins"} {
		if x, ok := o.Get(k); ok {
			if _, isObj := x.(schema.Object); !isObj {
				return nil, errorf("bad-file", ExitInput, next, "%s: %s is %s, not an object", SettingsFile, k, schema.Show(x))
			}
		}
	}
	if perms, ok := o.Get("permissions"); ok {
		if d, ok := perms.(schema.Object).Get("deny"); ok {
			if _, isList := d.([]any); !isList {
				return nil, errorf("bad-file", ExitInput, next, "%s: permissions.deny is not a list", SettingsFile)
			}
		}
	}
	if hooks, ok := o.Get("hooks"); ok {
		for _, m := range hooks.(schema.Object) {
			if _, isList := m.Value.([]any); !isList {
				return nil, errorf("bad-file", ExitInput, next, "%s: hooks.%s is not a list", SettingsFile, m.Key)
			}
		}
	}
	return &settingsDoc{root: o, exists: true, raw: raw}, nil
}

// ownedKeys are the top-level keys that are Bonsai's whatever their value.
var ownedKeys = map[string]bool{"autoMemoryEnabled": true, "disableAllHooks": true}

// claudeKeyOrder is the order Claude Code writes the top-level keys of a settings file in, for the keys measured on
// Claude Code 2.1.294 and 2.1.295 (step 5.1.7: a project-scope install rewrote a file holding each of them, given in
// another order): Bonsai's six keys and the keys around them. A key not listed is never an anchor (placeKey).
var claudeKeyOrder = []string{"$schema", "respectGitignore", "cleanupPeriodDays", "env", "includeCoAuthoredBy",
	"includeGitInstructions", "permissions", "model", "enableAllProjectMcpServers", "enabledMcpjsonServers",
	"disabledMcpjsonServers", "hooks", "disableAllHooks", "enabledPlugins", "extraKnownMarketplaces", "outputStyle",
	"spinnerTipsEnabled", "alwaysThinkingEnabled", "autoMemoryEnabled"}

// claudePermissionsOrder is the order Claude Code writes permissions' keys in (measured as claudeKeyOrder).
var claudePermissionsOrder = []string{"allow", "deny", "ask", "defaultMode", "additionalDirectories"}

// claudeMarketplaceOrder is the order Claude Code writes an inline marketplace's source object in (measured as
// claudeKeyOrder): its plugins before its owner. Bonsai's Line.Value keeps its own order, which the lock's fingerprint
// was taken over; the file gets this one (claudeOrdered).
var claudeMarketplaceOrder = []string{"source", "name", "plugins", "owner"}

// placeKey gives key its value in o: in place when o holds it, else inserted where Claude Code writes it (order):
// after the last key of o that order puts before key, or first when there is none. A key order does not list is
// appended.
func placeKey(o schema.Object, key string, v any, order []string) schema.Object {
	if i := o.Index(key); i >= 0 {
		o[i].Value = v
		return o
	}
	rank := map[string]int{}
	for i, k := range order {
		rank[k] = i
	}
	r, known := rank[key]
	if !known {
		return append(o, schema.Member{Key: key, Value: v})
	}
	at := 0
	for i, m := range o {
		if mr, ok := rank[m.Key]; ok && mr < r {
			at = i + 1
		}
	}
	out := make(schema.Object, 0, len(o)+1)
	out = append(out, o[:at]...)
	out = append(out, schema.Member{Key: key, Value: v})
	return append(out, o[at:]...)
}

// claudeOrdered gives a line's value as the file holds it: a marketplace's source object in Claude Code's order
// (claudeMarketplaceOrder), every other value as it is. It never changes the line's own value.
func claudeOrdered(l Line) any {
	o, ok := l.Value.(schema.Object)
	if l.Kind != "marketplace" || !ok {
		return l.Value
	}
	src, ok := o.Get("source")
	so, isObj := src.(schema.Object)
	if !ok || !isObj {
		return l.Value
	}
	var inner schema.Object
	for _, k := range claudeMarketplaceOrder {
		if v, ok := so.Get(k); ok {
			inner = append(inner, schema.Member{Key: k, Value: v})
		}
	}
	for _, m := range so {
		if inner.Index(m.Key) < 0 {
			inner = append(inner, m)
		}
	}
	out := cloneObject(o)
	out[out.Index("source")].Value = inner
	return out
}

// bonsaiMarketplaceOrder is the order buildLines writes an inline marketplace's source object in, which the lock's
// fingerprint of Bonsai's lines is taken over.
var bonsaiMarketplaceOrder = []string{"source", "name", "owner", "plugins"}

// bonsaiOrdered gives a marketplace's value read from the file with its source object in Bonsai's own order
// (bonsaiMarketplaceOrder), other keys after them as the file has them; any other value as it is.
func bonsaiOrdered(v any) any {
	o, ok := v.(schema.Object)
	if !ok {
		return v
	}
	i := o.Index("source")
	if i < 0 {
		return v
	}
	so, ok := o[i].Value.(schema.Object)
	if !ok {
		return v
	}
	var inner schema.Object
	for _, k := range bonsaiMarketplaceOrder {
		if x, ok := so.Get(k); ok {
			inner = append(inner, schema.Member{Key: k, Value: x})
		}
	}
	for _, m := range so {
		if inner.Index(m.Key) < 0 {
			inner = append(inner, m)
		}
	}
	out := cloneObject(o)
	out[i].Value = inner
	return out
}

// sameValue reports whether two JSON values are equal but for the order of their objects' keys.
func sameValue(a, b any) bool { return schema.Show(sortedKeys(a)) == schema.Show(sortedKeys(b)) }

// diskLines lists every line the file holds that Bonsai could own: the two keys, each deny rule, each command hook,
// each marketplace and each enabled plugin.
func diskLines(root schema.Object) []Line {
	var ls []Line
	for _, m := range root {
		if ownedKeys[m.Key] {
			ls = append(ls, Line{Kind: "key", Name: m.Key, Value: m.Value})
		}
	}
	if perms, ok := root.Get("permissions"); ok {
		d, _ := perms.(schema.Object).Get("deny")
		list, _ := d.([]any)
		for _, r := range list {
			if s, ok := r.(string); ok {
				ls = append(ls, Line{Kind: "deny", Rule: s})
			}
		}
	}
	if hooks, ok := root.Get("hooks"); ok {
		for _, ev := range hooks.(schema.Object) {
			groups, _ := ev.Value.([]any)
			for _, g := range groups {
				go_, ok := g.(schema.Object)
				if !ok {
					continue
				}
				matcher := go_.String("matcher")
				hs, _ := go_.Get("hooks")
				list, _ := hs.([]any)
				for _, h := range list {
					if l, ok := hookLine(ev.Key, matcher, h); ok {
						ls = append(ls, l)
					}
				}
			}
		}
	}
	for _, key := range []string{"extraKnownMarketplaces", "enabledPlugins"} {
		if x, ok := root.Get(key); ok {
			kind := "marketplace"
			if key == "enabledPlugins" {
				kind = "plugin"
			}
			for _, m := range x.(schema.Object) {
				ls = append(ls, Line{Kind: kind, Name: m.Key, Value: m.Value})
			}
		}
	}
	return ls
}

// hookLine reads one hook entry as a line; an entry that is not a command hook is the project's and no line.
func hookLine(event, matcher string, h any) (Line, bool) {
	o, ok := h.(schema.Object)
	if !ok || o.String("type") != "command" {
		return Line{}, false
	}
	cmd, ok := o.Get("command")
	c, isText := cmd.(string)
	if !ok || !isText {
		return Line{}, false
	}
	l := Line{Kind: "hook", Event: event, Matcher: matcher, Command: c}
	if t, ok := o.Get("timeout"); ok {
		if n, isNum := t.(json.Number); isNum {
			l.Timeout = string(n)
		}
	}
	if a, ok := o.Get("async"); ok && a == true {
		l.Async = true
	}
	if f, ok := o.Get("onFailure"); ok {
		if v, isText := f.(string); isText {
			l.OnFailure = v
		}
	}
	return l, true
}

// isBonsaiHook reports a hook line of Bonsai's own: one in exactly the form HookCommand writes (any word, its places
// the installed ones or a test build's one), or one that runs bonsai by name, `bonsai hook <word>...`, the form every
// build before step 5.3.6 wrote, so that a project linked by such a build has its lines taken and replaced. A line
// that only resembles the form (another message, another quoting, a command added) is not Bonsai's and is left as the
// project's own.
func isBonsaiHook(command string) bool {
	c := strings.TrimSpace(command)
	if c == "bonsai hook" || strings.HasPrefix(c, "bonsai hook ") {
		return true
	}
	places, word, ok := parseHookForm(c)
	return ok && (HookCommand(word, true, places...) == c || HookCommand(word, false, places...) == c)
}

// isOldBonsaiHook reports an old Bonsai hook line, one that calls Bonsai by an absolute path: a bonsai binary by
// its full path (/.../bonsai, C:\...\bonsai.exe, ~/.../bonsai), or Bonsai 0.4.3's sensor scripts
// (bash "<absolute>/agent/Sensors/<name>.sh" "<absolute>"). init takes such a line out when it links a project
// (spec §14, check 1: "the old absolute line is gone"). A line in Bonsai's own form (isBonsaiHook), which names the
// installed places, is never an old one.
func isOldBonsaiHook(command string) bool {
	if isBonsaiHook(command) {
		return false
	}
	words := shellWords(command)
	if len(words) == 0 {
		return false
	}
	if absolutePath(words[0]) {
		seg := words[0]
		if i := strings.LastIndexAny(seg, `/\`); i >= 0 {
			seg = seg[i+1:]
		}
		if s := strings.ToLower(seg); s == "bonsai" || s == "bonsai.exe" {
			return true
		}
	}
	if (words[0] == "bash" || words[0] == "sh") && len(words) > 1 && absolutePath(words[1]) &&
		strings.Contains(filepath.ToSlash(words[1]), "/agent/Sensors/") && strings.HasSuffix(words[1], ".sh") {
		return true
	}
	return false
}

func absolutePath(p string) bool {
	if strings.HasPrefix(p, "/") || strings.HasPrefix(p, "~/") || strings.HasPrefix(p, "$HOME/") ||
		strings.HasPrefix(strings.ToUpper(p), "%USERPROFILE%") {
		return true
	}
	return len(p) >= 3 && ((p[0] >= 'A' && p[0] <= 'Z') || (p[0] >= 'a' && p[0] <= 'z')) && p[1] == ':' &&
		(p[2] == '\\' || p[2] == '/')
}

// shellWords splits a command line into words the way a shell would for plain words and quoted ones; enough to
// find a command's program and first argument.
func shellWords(s string) []string {
	var words []string
	var cur strings.Builder
	in, quote := false, byte(0)
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case quote != 0:
			if c == quote {
				quote = 0
			} else {
				cur.WriteByte(c)
			}
		case c == '"' || c == '\'':
			quote, in = c, true
		case c == ' ' || c == '\t':
			if in {
				words = append(words, cur.String())
				cur.Reset()
				in = false
			}
		default:
			cur.WriteByte(c)
			in = true
		}
	}
	if in {
		words = append(words, cur.String())
	}
	return words
}

// claim finds Bonsai's lines in the file: those the reference lines (what Bonsai would have written from the
// lock's commits) hold, the lines no person writes, and, at a first link, old Bonsai hook lines. It says whether
// they are exactly what the lock says Bonsai last wrote (lockHash, "" for none), trying the file's other Edit
// rules, in each combination, as never_edit rules a person has since taken out of bonsai.yaml.
func claim(disk, ref []Line, firstLink bool, lockHash string) ([]Line, bool) {
	refs := map[string]Line{}
	for _, r := range ref {
		refs[r.match()] = r
	}
	var claimed, spare []Line
	seen := map[string]bool{}
	add := func(l Line) {
		if c := l.canon(); !seen[c] {
			seen[c] = true
			claimed = append(claimed, l)
		}
	}
	// ownAt counts Bonsai's own hook lines on the file's each event and matcher, in the file's order, so an own line
	// this build does not write takes the slot of its place there (the stop gate's line before the recorder's on Stop)
	// and pairs, in the preview, with this build's line in that place as one line changed.
	ownAt := map[string]int{}
	for _, d := range disk {
		ownKey, ownPos := "", 0
		if d.Kind == "hook" {
			ownKey = "hook:@own:" + d.Event + ":" + d.Matcher
			ownPos = ownAt[ownKey]
		}
		if r, ok := refs[d.match()]; ok {
			if r.Kind == "hook" && r.Own {
				ownAt[ownKey]++
			}
			add(r)
			continue
		}
		switch {
		case d.Kind == "key":
			d.Origin, d.Slot, d.Why = "bonsai", "key:"+d.Name, "Bonsai's own setting."
			add(d)
		case d.Kind == "marketplace" && strings.HasPrefix(d.Name, "bonsai-"):
			// Its value as Bonsai writes it (bonsaiOrdered): Claude Code moves the owner after the plugins, which is no
			// edit, and the fingerprint is taken over Bonsai's order.
			d.Origin, d.Slot, d.Value = "bonsai", "marketplace", bonsaiOrdered(d.Value)
			d.Why = "An old plugin marketplace of Bonsai's for this workspace: the one above replaces it."
			add(d)
		case d.Kind == "plugin" && strings.Contains(d.Name, "@bonsai-"):
			d.Origin, d.Slot = "bonsai", "plugin:"+strings.SplitN(d.Name, "@", 2)[0]
			d.Why = "Enables a plugin from an old marketplace of Bonsai's."
			add(d)
		case d.Kind == "hook" && isBonsaiHook(d.Command):
			ownAt[ownKey]++
			d.Origin, d.Own, d.Slot = "bonsai", true, ownKey+":"+strconv.Itoa(ownPos)
			d.Why = "A Bonsai hook line this build does not write."
			add(d)
		case d.Kind == "hook" && firstLink && isOldBonsaiHook(d.Command):
			d.Origin, d.Slot = "old", "old:"+d.canon()
			d.Why = "An old Bonsai hook line that calls bonsai by an absolute path; Bonsai's lines call it at its installed place."
			add(d)
		case d.Kind == "deny" && strings.HasPrefix(d.Rule, "Edit(") && strings.HasSuffix(d.Rule, ")"):
			// Its slot is its path's, in either form a build wrote (Edit(<path>) before step 5.3.5, Edit(/<path>)
			// since), so a rule that only changes form shows as one line changed, not as a path taken out.
			d.Origin, d.Slot = "bonsai.yaml", neverEditSlot+strings.TrimPrefix(strings.TrimSuffix(strings.TrimPrefix(d.Rule, "Edit("), ")"), "/")
			d.Why = "This file is no longer in never_edit in bonsai.yaml, so agents may edit it again."
			spare = append(spare, d)
		}
	}
	if lockHash == "" {
		return claimed, false
	}
	if linesHash(claimed) == lockHash {
		return claimed, true
	}
	var rest []Line
	for _, s := range spare {
		if !seen[s.canon()] {
			seen[s.canon()] = true
			rest = append(rest, s)
		}
	}
	// Each set of the other Edit rules, smallest first, while there are few enough to try them all (2^12 hashes);
	// past that, only all of them. A match is exact: the fingerprint is a SHA-256.
	if len(rest) > maxSpare {
		if wider := append(append([]Line{}, claimed...), rest...); linesHash(wider) == lockHash {
			return wider, true
		}
		return claimed, false
	}
	masks := make([]int, 0, 1<<len(rest))
	for m := 1; m < 1<<len(rest); m++ {
		masks = append(masks, m)
	}
	sort.SliceStable(masks, func(i, j int) bool { return bits(masks[i]) < bits(masks[j]) })
	for _, m := range masks {
		wider := append([]Line{}, claimed...)
		for i, s := range rest {
			if m&(1<<i) != 0 {
				wider = append(wider, s)
			}
		}
		if linesHash(wider) == lockHash {
			return wider, true
		}
	}
	return claimed, false
}

// maxSpare is how many of the file's other Edit rules claim tries in every combination.
const maxSpare = 12

func bits(m int) int {
	n := 0
	for ; m > 0; m &= m - 1 {
		n++
	}
	return n
}

// SettingsChange is one line the preview names: added, changed or removed, with its sentence (spec §6).
type SettingsChange struct {
	Change   string // add, change or remove
	Kind     string // hook, deny, key, marketplace or plugin
	Line     string // the line as written after the change (or, removed, as it was)
	Was      string // change: the line before
	Why      string
	RunsCode bool   // a hook line added or changed, and not one the lock last consented to: it runs code (spec §6)
	Origin   string // where the line comes from (Line.Origin): bonsai for Bonsai's own, or a pack id
	Own      bool   // a line of Bonsai's own (Line.Own), which no pack can set
	event    string // a hook line's event (Line.Event)
}

// OwnGuardChanges are the plan's changes that change or take out Bonsai's own guard line or stop gate's line (a line
// of Bonsai's own on PreToolUse or Stop that runs bonsai hook guard or stop): in an agent's session they are a
// person's command (step 5.3.6 note 8, cmd/bonsai), though a removed line runs nothing (consent.go). Adding one is not
// among them: it only puts the guard back.
func (p *Plan) OwnGuardChanges() []SettingsChange {
	var out []SettingsChange
	for _, c := range p.Settings {
		if c.Kind == "hook" && c.Own && (c.Change == "change" || c.Change == "remove") && (c.event == GuardEvent || c.event == StopEvent) &&
			guardOrStop(c) {
			out = append(out, c)
		}
	}
	return out
}

// guardOrStop reports a change of Bonsai's own line on PreToolUse or Stop that runs the guard or the stop gate, before
// or after (the recorder's own line on PreToolUse and on Stop is neither).
func guardOrStop(c SettingsChange) bool {
	for _, t := range []string{c.Line, c.Was} {
		if strings.Contains(t, "hook guard") || strings.Contains(t, "hook stop") {
			return true
		}
	}
	return false
}

// lineChanges lists the changes from Bonsai's lines in the file (claimed) to the new lines: a new line the file
// already holds (a person wrote the same line) is no change. reference holds the lines the lock last consented to,
// so a hook line that only comes back is not new code.
func lineChanges(claimed, lnew []Line, disk []Line, consented []Line) []SettingsChange {
	inNew, inClaimed, onDisk, ok := map[string]bool{}, map[string]bool{}, map[string]bool{}, map[string]bool{}
	for _, l := range lnew {
		inNew[l.canon()] = true
	}
	for _, l := range claimed {
		inClaimed[l.canon()] = true
	}
	for _, l := range disk {
		onDisk[l.canon()] = true
	}
	for _, l := range consented {
		ok[l.canon()] = true
	}
	var removed []Line
	for _, l := range claimed {
		if !inNew[l.canon()] {
			removed = append(removed, l)
		}
	}
	var out []SettingsChange
	for _, l := range lnew {
		c := l.canon()
		if inClaimed[c] || onDisk[c] {
			continue
		}
		sc := SettingsChange{Change: "add", Kind: l.Kind, Line: l.Text(), Why: l.Why,
			RunsCode: l.Kind == "hook" && !ok[c], Origin: l.Origin, Own: l.Own, event: l.Event}
		for i, r := range removed {
			if r.Slot == l.Slot {
				sc.Change, sc.Was = "change", r.Text()
				removed = append(removed[:i], removed[i+1:]...)
				break
			}
		}
		out = append(out, sc)
	}
	for _, r := range removed {
		out = append(out, SettingsChange{Change: "remove", Kind: r.Kind, Line: r.Text(), Why: r.Why, Origin: r.Origin, Own: r.Own, event: r.Event})
	}
	return out
}

// applyLines writes the new lines into the document in place of the claimed ones: a claimed line the new lines
// still hold stays where it is; the others go; a new line is added, the keys, marketplaces and plugins in place
// where their key was (else at the end), the deny rules and hook groups at the end of their lists. Every other
// entry of the file stays as it was, in its place. A list or object that held only Bonsai's lines goes with them.
func applyLines(root schema.Object, claimed, lnew []Line) schema.Object {
	doc := cloneObject(root)
	stays := map[string]bool{}
	for _, l := range lnew {
		stays[l.canon()] = true
	}
	gone := map[string]bool{}
	for _, l := range claimed {
		if !stays[l.canon()] {
			gone[l.canon()] = true
		}
	}
	newKeyed := map[string]map[string]any{"key": {}, "marketplace": {}, "plugin": {}}
	for _, l := range lnew {
		if m, ok := newKeyed[l.Kind]; ok {
			m[l.Name] = l.Value
		}
	}

	// Keys: set in place, or placed where Claude Code writes them; a claimed key no new line holds goes.
	for _, l := range claimed {
		if l.Kind == "key" {
			if _, keep := newKeyed["key"][l.Name]; !keep {
				doc = without(doc, l.Name)
			}
		}
	}
	for _, l := range lnew {
		if l.Kind == "key" {
			doc = placeKey(doc, l.Name, l.Value, claudeKeyOrder)
		}
	}

	// Deny rules.
	perms, _ := getObject(doc, "permissions")
	denyList, _ := perms.Get("deny")
	list, _ := denyList.([]any)
	var kept []any
	removedAny := false
	for _, r := range list {
		if s, ok := r.(string); ok && gone[(Line{Kind: "deny", Rule: s}).canon()] {
			removedAny = true
			continue
		}
		kept = append(kept, r)
	}
	present := map[string]bool{}
	for _, r := range kept {
		if s, ok := r.(string); ok {
			present[s] = true
		}
	}
	for _, l := range lnew {
		if l.Kind == "deny" && !present[l.Rule] {
			kept = append(kept, l.Rule)
			present[l.Rule] = true
		}
	}
	if len(kept) > 0 {
		perms = placeKey(perms, "deny", kept, claudePermissionsOrder)
	} else if removedAny {
		perms = without(perms, "deny")
	}
	doc = putObject(doc, "permissions", perms, removedAny)

	// Hook lines: claimed entries leave their groups; the new lines come as one group per event and matcher.
	hooks, _ := getObject(doc, "hooks")
	hooksTouched := false
	for i, ev := range hooks {
		groups, _ := ev.Value.([]any)
		var keptGroups []any
		for _, g := range groups {
			gObj, ok := g.(schema.Object)
			if !ok {
				keptGroups = append(keptGroups, g)
				continue
			}
			hs, _ := gObj.Get("hooks")
			hl, isList := hs.([]any)
			if !isList {
				keptGroups = append(keptGroups, g)
				continue
			}
			var keptHooks []any
			for _, h := range hl {
				if l, ok := hookLine(ev.Key, gObj.String("matcher"), h); ok && gone[l.canon()] {
					continue
				}
				keptHooks = append(keptHooks, h)
			}
			switch {
			case len(keptHooks) == len(hl):
				keptGroups = append(keptGroups, g)
			case len(keptHooks) > 0:
				keptGroups = append(keptGroups, set(gObj, "hooks", keptHooks))
				hooksTouched = true
			default:
				hooksTouched = true
			}
		}
		if keptGroups == nil {
			keptGroups = []any{}
		}
		hooks[i].Value = keptGroups
	}
	// Events left empty by the removal go; an event that was empty before stays.
	var keptEvents schema.Object
	for _, ev := range hooks {
		if g, _ := ev.Value.([]any); len(g) == 0 {
			if before, ok := getList(root, "hooks", ev.Key); ok && len(before) > 0 {
				continue
			}
		}
		keptEvents = append(keptEvents, ev)
	}
	hooks = keptEvents
	presentHook := map[string]bool{}
	for _, l := range diskLines(schema.Object{{Key: "hooks", Value: hooks}}) {
		presentHook[l.canon()] = true
	}
	type groupKey struct{ event, matcher string }
	var order []groupKey
	groups := map[groupKey][]any{}
	for _, l := range lnew {
		if l.Kind != "hook" || presentHook[l.canon()] {
			continue
		}
		k := groupKey{l.Event, l.Matcher}
		if _, ok := groups[k]; !ok {
			order = append(order, k)
		}
		groups[k] = append(groups[k], l.hookEntry())
	}
	for _, k := range order {
		g := schema.Object{}
		if k.matcher != "" {
			g = append(g, schema.Member{Key: "matcher", Value: k.matcher})
		}
		g = append(g, schema.Member{Key: "hooks", Value: groups[k]})
		existing, _ := hooks.Get(k.event)
		el, _ := existing.([]any)
		hooks = set(hooks, k.event, append(append([]any{}, el...), g))
	}
	doc = putObject(doc, "hooks", hooks, hooksTouched)

	// Marketplaces and plugins: a claimed entry no new line holds goes, and a new line of the same slot (the
	// marketplace, or a pack's plugin, under a new name) takes its place; an entry that stays keeps its place, and its
	// bytes when its value is the same but for key order; any other new line is appended.
	for _, kind := range []string{"marketplace", "plugin"} {
		key := map[string]string{"marketplace": "extraKnownMarketplaces", "plugin": "enabledPlugins"}[kind]
		obj, _ := getObject(doc, key)
		var fresh []Line
		for _, l := range lnew {
			if l.Kind == kind {
				fresh = append(fresh, l)
			}
		}
		placed := map[string]bool{}
		goneSlot := map[string]string{} // a claimed entry's name that goes -> its slot
		for _, l := range claimed {
			if _, keep := newKeyed[kind][l.Name]; l.Kind == kind && !keep {
				goneSlot[l.Name] = l.Slot
			}
		}
		touched := false
		var out schema.Object
		for _, m := range obj {
			if slot, gone := goneSlot[m.Key]; gone {
				touched = true
				for _, l := range fresh {
					if l.Slot == slot && !placed[l.Name] && obj.Index(l.Name) < 0 {
						out = append(out, schema.Member{Key: l.Name, Value: claudeOrdered(l)})
						placed[l.Name] = true
						break
					}
				}
				continue
			}
			for _, l := range fresh {
				if l.Name == m.Key {
					placed[l.Name] = true
					if !sameValue(m.Value, l.Value) {
						m.Value = claudeOrdered(l)
					}
				}
			}
			out = append(out, m)
		}
		for _, l := range fresh {
			if !placed[l.Name] {
				out = set(out, l.Name, claudeOrdered(l))
				placed[l.Name] = true
			}
		}
		doc = putObject(doc, key, out, touched)
	}
	return doc
}

func cloneObject(o schema.Object) schema.Object {
	out := make(schema.Object, len(o))
	for i, m := range o {
		out[i] = schema.Member{Key: m.Key, Value: cloneValue(m.Value)}
	}
	return out
}

func cloneValue(v any) any {
	switch x := v.(type) {
	case schema.Object:
		return cloneObject(x)
	case []any:
		out := make([]any, len(x))
		for i, e := range x {
			out[i] = cloneValue(e)
		}
		return out
	}
	return v
}

func getObject(o schema.Object, key string) (schema.Object, bool) {
	v, ok := o.Get(key)
	obj, _ := v.(schema.Object)
	return obj, ok
}

func getList(o schema.Object, key, inner string) ([]any, bool) {
	obj, ok := getObject(o, key)
	if !ok {
		return nil, false
	}
	v, ok := obj.Get(inner)
	l, _ := v.([]any)
	return l, ok
}

// set gives a key its value in place, or appends it.
func set(o schema.Object, key string, v any) schema.Object {
	if i := o.Index(key); i >= 0 {
		o[i].Value = v
		return o
	}
	return append(o, schema.Member{Key: key, Value: v})
}

func without(o schema.Object, key string) schema.Object {
	if i := o.Index(key); i >= 0 {
		return append(o[:i:i], o[i+1:]...)
	}
	return o
}

// putObject stores an inner object, in place or where Claude Code writes it (placeKey): an empty one is left out
// when the removal emptied it or it was not there.
func putObject(doc schema.Object, key string, inner schema.Object, removed bool) schema.Object {
	_, had := doc.Get(key)
	if len(inner) == 0 && (removed || !had) {
		return without(doc, key)
	}
	if len(inner) == 0 {
		return placeKey(doc, key, schema.Object{}, claudeKeyOrder)
	}
	return placeKey(doc, key, inner, claudeKeyOrder)
}
