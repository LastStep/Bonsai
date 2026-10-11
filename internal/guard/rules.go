package guard

// The guard's rules table (design/plan-5.md, 5.3.2 note 9): every rule a guard decision names, the log record's rule
// field (contract §8.1), with whether it allows or refuses and one line on it. This table is the list's one home: the
// names below are its rows' names, docs/reference/lists.md prints it (go generate ./...), and TestEveryRule walks it,
// so a rule with no test fails. Names are kebab-case and never renamed (a log reader keys on them); a rule is added
// with a row here, its test and the page.
//
// Step 5.3.3's delete check and step 5.3.4's stop gate add their rows when they are wired in.

const (
	// What a path is judged by (rule.go, may.go).
	RuleProtected       = "protected"
	RulePersonOnly      = "person-only"
	RuleNeverGranted    = "never-granted"
	RuleBonsaiFiles     = "bonsai-files"
	RuleStandIn         = "bonsai-stand-in"
	RuleMainNotVerified = "main-not-verified"
	RuleOtherProject    = "other-project"
	RuleGranted         = "granted"
	RuleNotProtected    = "not-protected"
	RuleOutside         = "outside-project"

	// command mode's tripwires (tripwire.go).
	RuleTaskOutsideMain = "task-outside-main"
	RuleStatusLine      = "status-line"
	RuleLaneLine        = "lane-line"
	RuleGrantPersonOnly = "grant-person-only"
	RuleOutsideLabel    = "outside-label"
	RuleBadFormat       = "bad-format"
	RuleBadSettings     = "bad-settings"

	// Bonsai's own commands in a shell call (commands.go).
	RulePersonCommand = "person-command"
	RuleHomeCommand   = "home-command"
	RuleBonsaiCommand = "bonsai-command"
	RuleShell         = "shell-not-judged"
	RuleOtherTool     = "tool-not-judged"

	// When the guard cannot read or decide.
	RuleBadInput  = "bad-input"
	RuleBadGlob   = "bad-glob"
	RuleBadPath   = "bad-path"
	RuleBadLock   = "bad-lock"
	RuleBadConfig = "bad-config"
	RuleNoProject = "no-project"
	RuleLogFailed = "log-failed"
	RuleOverTime  = "over-time"
	RuleInternal  = "internal-error"
	RuleNotLinked = "not-linked"
)

// Rule is one row of the rules table.
type Rule struct {
	Name  string // the rule's name, kebab-case: the record's rule
	Allow bool   // true: the call goes on; false: it is blocked
	Line  string // one line on it, ASCII
}

// Rules is the guard's rules table, its one home, in the order the reference page lists it.
var Rules = []Rule{
	{RuleProtected, false, "a path on a protected list (bonsai.yaml's protected, a pack's declared protected paths) that no running task grants"},
	{RulePersonOnly, false, "a path on bonsai.yaml's person_only list, or one the floor counts as person-only (bonsai.yaml, " +
		".bonsai/lock.json, anything else in .bonsai/, and a bonsai.yaml or .bonsai/ at any depth), that no running task grants"},
	{RuleNeverGranted, false, "a path of the floor no grant opens: .claude/settings.json, .claude/settings.local.json and .git at a " +
		"checkout's top, and a .git at any depth (for a delete, the project's own top .git only)"},
	{RuleBonsaiFiles, false, "Bonsai's own files, in every mode and under any grant: .bonsai/local/, .bonsai/.gitignore, the two " +
		"tables (.bonsai/tasks.md, .bonsai/sessions.md), and anything in Bonsai's home"},
	{RuleStandIn, false, "a file named as Bonsai's program (bonsai, bonsai.exe, bonsai.cmd, bonsai.bat, bonsai.ps1, bonsai.com) " +
		"outside every linked project, which a folder on the PATH could hold"},
	{RuleMainNotVerified, false, "a guarded path where no project holding it is a verified checkout, so no task's grant can count"},
	{RuleOtherProject, false, "a delete or move of another linked project's top folder or its .git, .bonsai or .claude folder (that " +
		"project on this machine's record)"},
	{RuleGranted, true, "a guarded path the running task's bonsai.allows holds (the task read from the verified main checkout)"},
	{RuleNotProtected, true, "a path inside a linked project that nothing guards"},
	{RuleOutside, true, "a path in no linked project, in every form it takes"},
	{RuleTaskOutsideMain, false, "command mode: an edit of a task file outside the main checkout"},
	{RuleStatusLine, false, "command mode: an edit of a declared document's status line (a new document may start at its kind's " +
		"first status only)"},
	{RuleLaneLine, false, "command mode: an edit of a task's lane line"},
	{RuleGrantPersonOnly, false, "command mode: an edit that puts a person-only path into a task's bonsai.allows"},
	{RuleOutsideLabel, false, "command mode: an edit that sets a label whose definition reads set_by: outside"},
	{RuleBadFormat, false, "command mode: a write of a task, run report or STATE that fails its format"},
	{RuleBadSettings, false, "this machine's settings for the workspace cannot be read when a judgment needs its mode"},
	{RulePersonCommand, false, "bonsai with --allow-exec, bonsai unlink, or a line that unsets or changes CLAUDE_CODE_CHILD_SESSION: " +
		"a person's commands, allowed to an agent only in a project in command mode under a running task's grant of " +
		"bonsai.yaml and .bonsai/lock.json"},
	{RuleHomeCommand, false, "bonsai settings set, bonsai labels attach or detach: never in an agent session"},
	{RuleBonsaiCommand, true, "a bonsai command an agent may run: init and update judged as an edit of bonsai.yaml and the lock, " +
		"every other word free"},
	{RuleShell, true, "a Bash, PowerShell or Monitor call whose line changes nothing the guard judges"},
	{RuleOtherTool, true, "a tool other than the file tools and the shells"},
	{RuleBadInput, false, "the hook's payload cannot be read"},
	{RuleBadGlob, false, "a glob on a list is not a glob the guard reads"},
	{RuleBadPath, false, "Windows cannot name where the path leads, or it runs through a junction or link that does not resolve"},
	{RuleBadLock, false, "the lock's declares cannot be read (or the lock is missing while bonsai.yaml names packs)"},
	{RuleBadConfig, false, "a bonsai.yaml is there but cannot be read (the session's own: blocked with no record)"},
	{RuleNoProject, false, "CLAUDE_PROJECT_DIR is not set, or its folder cannot be read (blocked with no record)"},
	{RuleLogFailed, false, "the guard would allow the call but cannot write its record"},
	{RuleOverTime, false, "the guard ran past its own time limit"},
	{RuleInternal, false, "the guard failed inside (a panic)"},
	{RuleNotLinked, true, "neither the session's folder nor the path is in a linked project: allowed with no record (spec section 7)"},
}
