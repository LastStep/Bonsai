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
	RuleTimeNotNow      = "time-not-now"

	// The delete check in a shell call (shell.go).
	RuleDeleteUnnamed   = "delete-unnamed"
	RuleDeleteProtected = "delete-protected"
	RuleDeleteUnsplit   = "delete-unsplit"
	RuleDeleteNamed     = "delete-named"
	RuleMoveUnnamed     = "move-unnamed"

	// Bonsai's own commands in a shell call (commands.go).
	RulePersonCommand = "person-command"
	RuleHomeCommand   = "home-command"
	RuleBonsaiCommand = "bonsai-command"
	RuleShell         = "shell-not-judged"
	RuleOtherTool     = "tool-not-judged"

	// The roles' rules and the stall stops, where the lock declares roles (step 5.5.11; roles.go, budget.go,
	// rolecmds.go).
	RuleCallAllowed       = "call-allowed"
	RuleCallRefused       = "call-refused"
	RuleRoleWritesNothing = "role-writes-nothing"
	RuleBudgetSpent       = "budget-spent"
	RuleBudgetExempt      = "budget-exempt"
	RuleBriefNotTied      = "brief-not-tied"
	RuleStateUnreadable   = "state-unreadable"
	RuleRoleUnbacked      = "role-unbacked"
	RuleRolesFromFiles    = "roles-from-files"
	RuleSecondLineFault   = "second-line-fault"

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
	RuleStopNotNamed          = "stop-not-named"
	RuleStopTaskUnknown       = "stop-task-unknown"
	RuleStopTaskTwice         = "stop-task-twice"
	RuleStopTaskBroken        = "stop-task-broken"
	RuleStopTaskUnreadable    = "stop-task-unreadable"
	RuleStopBadRungs          = "stop-bad-rungs"
	RuleStopNoRunReport       = "stop-no-run-report"
	RuleStopBadRunReport      = "stop-bad-run-report"
	RuleStopNoResult          = "stop-no-result"
	RuleStopBadResult         = "stop-bad-result"
	RuleStopOtherTask         = "stop-other-task"
	RuleStopNotLocal          = "stop-not-local"
	RuleStopRed               = "stop-red"
	RuleStopHeadUnreadable    = "stop-head-unreadable"
	RuleStopOtherCommit       = "stop-other-commit"
	RuleStopDirty             = "stop-dirty"
	RuleStopConfigUncommitted = "stop-config-uncommitted"
	RuleStopConfigLooser      = "stop-config-looser"
	RuleStopRungNotAsked      = "stop-rung-not-asked"
	RuleStopRungNotGreen      = "stop-rung-not-green"
	RuleStopProven            = "stop-proven"
	RuleStopBudget            = "stop-budget"
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
		"tables (.bonsai/tasks.md, .bonsai/sessions.md), and anything in Bonsai's home (from WSL, a Windows user's too); where " +
		"the lock declares roles, an agent's shell write onto one, or the lock, and a change of the permissions, owner, ACL or " +
		"attributes of one, of a folder holding one, or by a glob that may name one (chmod, chown, chattr, setfacl, attrib, " +
		"icacls, Set-ItemProperty, Set-Acl and their kind)"},
	{RuleStandIn, false, "a file named as Bonsai's program (bonsai, bonsai.exe, bonsai.cmd, bonsai.bat, bonsai.ps1, bonsai.com) " +
		"outside every linked project, which a folder on the PATH could hold"},
	{RuleMainNotVerified, false, "a guarded path where no project holding it is a verified checkout, so no task's grant can count; " +
		"for the stop gate, the session's checkout's main checkout is not verified, or not on the base branch this machine " +
		"recorded for it, so its task and ladder result are not trusted"},
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
	{RuleTimeNotNow, false, "every mode: an Edit or Write of a fresh stamp not of Bonsai's clock (a task's created, started " +
		"or finished, a run report's started or finished, STATE's or a memory note's updated, a pack kind's stamp field): " +
		"a time more than 10 minutes from the UTC now, a day after the UTC today or more than 2 days before it, or one " +
		"that does not read in its form (contract section 23)"},
	{RuleDeleteUnnamed, false, "a shell call's recursive or bulk delete that does not name what it deletes (an operand " +
		"holding a glob, a variable or a substitution, or one that is ., .., /, ~ or a drive's root; xargs, a pipeline or a " +
		"loop feeding it, or a plain mv, but not a git mv: move-unnamed; git clean with no path; a delete of several files by a glob; a git pathspec with magic other than " +
		":/, :(top) or :(literal)), and always git clean -x or -X, git stash --all and rm --no-preserve-root"},
	{RuleDeleteProtected, false, "a shell call's delete or move of a guarded path, or of a folder holding one, judged as an " +
		"edit of that path (the reason says what guards it)"},
	{RuleDeleteUnsplit, false, "a shell call's line the delete check cannot split (a quote or bracket left open, nesting " +
		"past three levels) that holds a delete word, or whose runner words run past the walk's bound"},
	{RuleDeleteNamed, true, "a shell call whose deletes and moves all name what they take, none of it guarded"},
	{RuleMoveUnnamed, true, "a shell call's git mv whose operands come from a loop's variable, xargs or a pipeline: git " +
		"moves only what it tracks and keeps both sides in the next commit, so rung 0 judges what moved after the fact; " +
		"what it names is still judged, a named destination folder with everything below it and each literal word of the " +
		"loop's list (bash's for ... in, PowerShell's foreach and a literal list piped into ForEach-Object) " +
		"(delete-protected), and plain mv and every delete fed the same way stay delete-unnamed; text printf or echo " +
		"writes into xargs is not read as names"},
	{RulePersonCommand, false, "bonsai with --allow-exec, bonsai unlink, or a line that unsets or changes an agent marker (CLAUDE_CODE_CHILD_SESSION, " +
		"AI_AGENT or another of the list the error code person-command names): " +
		"a person's commands, allowed to an agent only in a project in command mode under a running task's grant of " +
		"bonsai.yaml and .bonsai/lock.json"},
	{RuleHomeCommand, false, "bonsai settings set, bonsai labels attach or detach: never in an agent session"},
	{RuleBonsaiCommand, true, "a bonsai command an agent may run: init and update judged as an edit of bonsai.yaml and the lock, " +
		"every other word free"},
	{RuleShell, true, "a Bash, PowerShell or Monitor call whose line changes nothing the guard judges: no delete or move, " +
		"no bonsai command, every agent marker left as it is"},
	{RuleOtherTool, true, "a tool other than the file tools and the shells"},
	{RuleCallAllowed, true, "where the lock declares roles: an Agent, Task, SendMessage or Workflow call the caller's role may make, " +
		"or a Skill whose frontmatter forks into a type it may call (the person's session any role but its own; a lead a declared " +
		"helper or a type no pack declares; SendMessage from the person's session to main or another session); the record's " +
		"target is the type asked for, or the message's to"},
	{RuleCallRefused, false, "where the lock declares roles: an Agent, Task, SendMessage or Workflow call the caller's role may not " +
		"make, or a Skill whose frontmatter forks into such a type (a helper or an undeclared type calls no one; a lead no lead, " +
		"no Workflow, no main or other session, and resumes only an agent it started; no one the session's role); a SendMessage " +
		"from an agent whose to names no agent the guard recorded; on an agent's shell line, a command starting Claude Code, a " +
		"bonsai message --kind brief to a role it may not call, and a bonsai message from a --from other than its own address " +
		"or from an agent with none"},
	{RuleRoleWritesNothing, false, "where the lock declares roles: a role that changes nothing (writes: false, every undeclared type " +
		"among them) calling Edit, Write, MultiEdit or NotebookEdit, or a shell write the guard can read into a linked project " +
		"(a redirect, tee, cp, ln or sed -i onto a path there; rm or mv of one; a change of its permissions, owner, ACL or " +
		"attributes; git's words that change the checkout or its history; bonsai init, update, unlink, check --write); a relative " +
		"path after a cd the guard cannot follow is read in the folder the command started in, and its refusal says so"},
	{RuleBudgetSpent, false, "where the lock declares roles: an agent past a limit in force (its minutes, actions or red runs), or " +
		"whose window is held spent (its state moved aside a second time), or whose call the guard cannot count (its line or " +
		"state cannot be written or read, even with its access set back: fail closed), every call but the three budget-exempt " +
		"names; the first refusal says what to report, its next step the exact command by the installed bonsai's place"},
	{RuleBudgetExempt, true, "past an agent's budget, one of the three calls it may still make: its report, one shell call holding one " +
		"bonsai message --kind report from its own address and nothing else; reading its tied brief or a newer one to its address " +
		"(bonsai message --show <id>, or a Read of the brief's file); Claude Code's hand-back tool; each shell call by the " +
		"installed bonsai at the place the hook lines run it from, never a bare bonsai"},
	{RuleBriefNotTied, true, "an agent's brief named at its start, by a SendMessage, or read by it, ties nothing (a record beside the " +
		"call's own): no message file holds it, it is not a brief to the agent's address, its from may not call the agent's role, " +
		"it is not newer than the tied one in its thread, or another agent tied it"},
	{RuleStateUnreadable, true, "an agent's state under .bonsai/local/agents/ cannot be read or written (a record beside the call's " +
		"own): where its own access was taken away, set back and counted on; else one that cannot be read is moved aside, kept and " +
		"named, and the agent counted afresh from this call, once (a second time holds its window spent); a call that still " +
		"cannot be counted is refused (budget-spent)"},
	{RuleRoleUnbacked, true, "an agent's type names a declared role, but no Agent, Task or Skill call the guard allowed for that " +
		"role started it, so it is held as a type no pack declares: it changes nothing, calls no one, and runs on a helper's " +
		"budget (a record beside its first call's own)"},
	{RuleRolesFromFiles, true, "the main checkout's HEAD holds neither bonsai.yaml nor the lock, so an agent's role facts and limits " +
		"were read from the files as they stand (a record beside the call's own)"},
	{RuleSecondLineFault, true, "the guard's second line could not read or decide (its input, bonsai.yaml, the lock, the roles as " +
		"committed, an agent's state, its own time limit) and lets the call through, as its line does while bonsai is missing"},
	{RuleBadInput, false, "the hook's payload cannot be read (for the stop gate, or it is not a Stop payload, or bonsai hook " +
		"stop's command line holds what it does not take: blocked with no record)"},
	{RuleBadGlob, false, "a glob on a list is not a glob the guard reads"},
	{RuleBadPath, false, "the guard cannot tell where the path leads: its symbolic links loop or run past 64, it runs through " +
		"a junction or link that does not resolve or into /proc (Linux), or Windows cannot name it"},
	{RuleBadLock, false, "the lock's declares cannot be read (or the lock is missing while bonsai.yaml names packs)"},
	{RuleBadConfig, false, "a bonsai.yaml is there but cannot be read (the session's own: blocked with no record)"},
	{RuleNoProject, false, "CLAUDE_PROJECT_DIR is not set, or its folder cannot be read (blocked with no record)"},
	{RuleLogFailed, false, "the guard would allow the call, or the stop gate the session's end, but cannot write its record"},
	{RuleOverTime, false, "the guard or the stop gate ran past its own time limit"},
	{RuleInternal, false, "the guard or the stop gate failed inside (a panic)"},
	{RuleNotLinked, true, "neither the session's folder nor the path is in a linked project: allowed with no record (spec section 7); " +
		"for the stop gate, the session's folder is in no linked project"},
	{RuleStopNotNamed, true, "the stop gate: BONSAI_TASK names no task, so it does not engage: the session ends, with no record"},
	{RuleStopTaskUnknown, false, "the stop gate: no task file in the main checkout has the id BONSAI_TASK names"},
	{RuleStopTaskTwice, false, "the stop gate: two or more task files have the named id"},
	{RuleStopTaskBroken, false, "the stop gate: the named task's file does not parse under its format"},
	{RuleStopTaskUnreadable, false, "the stop gate: the main checkout's task folder, or a file in it, cannot be read"},
	{RuleStopBadRungs, false, "the stop gate: the task's bonsai.ladder (format 0's ladder) or the workspace's ladder_floor is " +
		"not a list of rung numbers"},
	{RuleStopNoRunReport, false, "the stop gate: the session's checkout holds no run report for the task in its run folder " +
		"(a file named R-YYYY-MM-DD-T-nnnn.md for the task)"},
	{RuleStopBadRunReport, false, "the stop gate: a run report of the task in the session's checkout does not pass bonsai " +
		"check: check makes a finding on the file itself (format0-new, format0, document, label or absolute-path)"},
	{RuleStopNoResult, false, "the stop gate: the main checkout holds no ladder result for the task (.bonsai/local/ladder/T-nnnn.json for the task)"},
	{RuleStopBadResult, false, "the stop gate: the ladder result cannot be read, or does not read as bonsai.ladder/1"},
	{RuleStopOtherTask, false, "the stop gate: the ladder result at the task's path was made for another task: its task is " +
		"another id or null, or its workspace is another project's"},
	{RuleStopNotLocal, false, "the stop gate: the ladder result's mode is not local (a CI run is never a task's proof)"},
	{RuleStopRed, false, "the stop gate: the ladder result is not green"},
	{RuleStopHeadUnreadable, false, "the stop gate: the commit at the checkout's HEAD cannot be read from its git files, or git " +
		"cannot read bonsai.yaml at the main checkout's HEAD (an object HEAD needs is not there, or is damaged: git fsck names it)"},
	{RuleStopOtherCommit, false, "the stop gate: the ladder result proves another commit than the checkout's HEAD"},
	{RuleStopDirty, false, "the stop gate: the ladder result was made on a tree with uncommitted changes"},
	{RuleStopConfigUncommitted, false, "the stop gate: the main checkout's bonsai.yaml differs from the one committed at its HEAD, " +
		"or its HEAD holds none, so the ladder's settings there are not trusted"},
	{RuleStopConfigLooser, false, "the stop gate: the main checkout's bonsai.yaml, as committed at its HEAD, sets the checks looser " +
		"than the settings this machine last accepted for the workspace (a rung out of the floor or the ladder, a rung added " +
		"to the ladder, whose command runs inside the climb, a rung's check changed, a ratchet's floor lowered, a CI mark " +
		"added, a path out of protected, person_only, never_edit or never_read, a document moved), or this machine's " +
		"record holds none: a person records them with bonsai update"},
	{RuleStopRungNotAsked, false, "the stop gate: the result did not ask for a rung of the workspace's ladder_floor or of the " +
		"task's bonsai.ladder"},
	{RuleStopRungNotGreen, false, "the stop gate: a rung of the floor or of the task's bonsai.ladder is not green in the result"},
	{RuleStopProven, true, "the stop gate: a green local ladder result proves the checkout's HEAD, with the floor and the " +
		"task's rungs green: the session ends"},
	{RuleStopBudget, true, "the stop gate: a named session whose budget is spent has written its report: it may stop, its ladder " +
		"red or not (the stall stops)"},
}
