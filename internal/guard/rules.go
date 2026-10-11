package guard

// The guard's rules table (design/plan-5.md, 5.3.2 note 9): every rule a guard decision names, the log record's rule
// field (contract §8.1), with whether it allows or refuses and one line on it. This table is the list's one home: the
// names below are its rows' names, docs/reference/lists.md prints it (go generate ./...), and TestEveryRule walks it,
// so a rule with no test fails. Names are kebab-case and never renamed (a log reader keys on them); a rule is added
// with a row here, its test and the page.
//
// The stop gate (step 5.3.4, internal/stopgate: bonsai hook stop) names its decisions from this table too, its records
// guard records with agent_event Stop: the shared rows below (bad-input to not-linked) for the same cases, and its own
// stop- rows. Its package keeps the names as its own constants (it cannot import this package, whose tests run it),
// and a test there holds them to these rows, name and answer; stop_test.go here is each stop- row's test.

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

	// The delete check in a shell call (shell.go).
	RuleDeleteUnnamed   = "delete-unnamed"
	RuleDeleteProtected = "delete-protected"
	RuleDeleteUnsplit   = "delete-unsplit"
	RuleDeleteNamed     = "delete-named"

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

	// The stop gate's own (internal/stopgate), in the order it checks.
	RuleStopNotNamed       = "stop-not-named"
	RuleStopTaskMissing    = "stop-task-missing"
	RuleStopTaskTwice      = "stop-task-twice"
	RuleStopTaskBroken     = "stop-task-broken"
	RuleStopTaskUnreadable = "stop-task-unreadable"
	RuleStopBadRungs       = "stop-bad-rungs"
	RuleStopNoRunReport    = "stop-no-run-report"
	RuleStopNoResult       = "stop-no-result"
	RuleStopBadResult      = "stop-bad-result"
	RuleStopOtherTask      = "stop-other-task"
	RuleStopNotLocal       = "stop-not-local"
	RuleStopRed            = "stop-red"
	RuleStopHeadUnreadable = "stop-head-unreadable"
	RuleStopOtherCommit    = "stop-other-commit"
	RuleStopDirty          = "stop-dirty"
	RuleStopRungMissing    = "stop-rung-missing"
	RuleStopRungNotGreen   = "stop-rung-not-green"
	RuleStopProven         = "stop-proven"
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
		"tables (.bonsai/tasks.md, .bonsai/sessions.md), and anything in Bonsai's home (from WSL, a Windows user's too)"},
	{RuleStandIn, false, "a file named as Bonsai's program (bonsai, bonsai.exe, bonsai.cmd, bonsai.bat, bonsai.ps1, bonsai.com) " +
		"outside every linked project, which a folder on the PATH could hold"},
	{RuleMainNotVerified, false, "a guarded path where no project holding it is a verified checkout, so no task's grant can count; " +
		"for the stop gate, the session's checkout's main checkout is not verified, so its task and ladder result are not trusted"},
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
	{RuleDeleteUnnamed, false, "a shell call's recursive or bulk delete that does not name what it deletes (an operand " +
		"holding a glob, a variable or a substitution, or one that is ., .., /, ~ or a drive's root; xargs, a pipeline or a " +
		"loop feeding it; git clean with no path; a delete of several files by a glob; a git pathspec with magic other than " +
		":/, :(top) or :(literal)), and always git clean -x or -X, git stash --all and rm --no-preserve-root"},
	{RuleDeleteProtected, false, "a shell call's delete or move of a guarded path, or of a folder holding one, judged as an " +
		"edit of that path (the reason says what guards it)"},
	{RuleDeleteUnsplit, false, "a shell call's line the delete check cannot split (a quote or bracket left open, nesting " +
		"past three levels) that holds a delete word, or whose runner words run past the walk's bound"},
	{RuleDeleteNamed, true, "a shell call whose deletes and moves all name what they take, none of it guarded"},
	{RulePersonCommand, false, "bonsai with --allow-exec, bonsai unlink, or a line that unsets or changes CLAUDE_CODE_CHILD_SESSION: " +
		"a person's commands, allowed to an agent only in a project in command mode under a running task's grant of " +
		"bonsai.yaml and .bonsai/lock.json"},
	{RuleHomeCommand, false, "bonsai settings set, bonsai labels attach or detach: never in an agent session"},
	{RuleBonsaiCommand, true, "a bonsai command an agent may run: init and update judged as an edit of bonsai.yaml and the lock, " +
		"every other word free"},
	{RuleShell, true, "a Bash, PowerShell or Monitor call whose line changes nothing the guard judges: no delete or move, " +
		"no bonsai command, CLAUDE_CODE_CHILD_SESSION left as it is"},
	{RuleOtherTool, true, "a tool other than the file tools and the shells"},
	{RuleBadInput, false, "the hook's payload cannot be read (for the stop gate, or it is not a Stop payload)"},
	{RuleBadGlob, false, "a glob on a list is not a glob the guard reads"},
	{RuleBadPath, false, "Windows cannot name where the path leads, or it runs through a junction or link that does not resolve"},
	{RuleBadLock, false, "the lock's declares cannot be read (or the lock is missing while bonsai.yaml names packs)"},
	{RuleBadConfig, false, "a bonsai.yaml is there but cannot be read (the session's own: blocked with no record)"},
	{RuleNoProject, false, "CLAUDE_PROJECT_DIR is not set, or its folder cannot be read (blocked with no record)"},
	{RuleLogFailed, false, "the guard would allow the call, or the stop gate the session's end, but cannot write its record"},
	{RuleOverTime, false, "the guard or the stop gate ran past its own time limit"},
	{RuleInternal, false, "the guard or the stop gate failed inside (a panic)"},
	{RuleNotLinked, true, "neither the session's folder nor the path is in a linked project: allowed with no record (spec section 7); " +
		"for the stop gate, the session's folder is in no linked project"},
	{RuleStopNotNamed, true, "the stop gate: BONSAI_TASK names no task, so it does not engage: the session ends, with no record"},
	{RuleStopTaskMissing, false, "the stop gate: no task file in the main checkout has the id BONSAI_TASK names"},
	{RuleStopTaskTwice, false, "the stop gate: two or more task files have the named id"},
	{RuleStopTaskBroken, false, "the stop gate: the named task's file does not parse under its format"},
	{RuleStopTaskUnreadable, false, "the stop gate: the main checkout's task folder, or a file in it, cannot be read"},
	{RuleStopBadRungs, false, "the stop gate: the task's bonsai.ladder (format 0's ladder) or the workspace's ladder_floor is " +
		"not a list of rung numbers"},
	{RuleStopNoRunReport, false, "the stop gate: the session's checkout holds no run report for the task in its run folder " +
		"(a file named R-YYYY-MM-DD-T-nnnn.md for the task)"},
	{RuleStopNoResult, false, "the stop gate: the main checkout holds no ladder result for the task (.bonsai/local/ladder/T-nnnn.json for the task)"},
	{RuleStopBadResult, false, "the stop gate: the ladder result cannot be read, or does not read as bonsai.ladder/1"},
	{RuleStopOtherTask, false, "the stop gate: the ladder result at the task's path was made for another task: its task is " +
		"another id or null, or its workspace is another project's"},
	{RuleStopNotLocal, false, "the stop gate: the ladder result's mode is not local (a CI run is never a task's proof)"},
	{RuleStopRed, false, "the stop gate: the ladder result is not green"},
	{RuleStopHeadUnreadable, false, "the stop gate: the commit at the checkout's HEAD cannot be read from its git files"},
	{RuleStopOtherCommit, false, "the stop gate: the ladder result proves another commit than the checkout's HEAD"},
	{RuleStopDirty, false, "the stop gate: the ladder result was made on a tree with uncommitted changes"},
	{RuleStopRungMissing, false, "the stop gate: the result did not ask for a rung of the workspace's ladder_floor or of the " +
		"task's bonsai.ladder"},
	{RuleStopRungNotGreen, false, "the stop gate: a rung of the floor or of the task's bonsai.ladder is not green in the result"},
	{RuleStopProven, true, "the stop gate: a green local ladder result proves the checkout's HEAD, with the floor and the " +
		"task's rungs green: the session ends"},
}
