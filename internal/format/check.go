package format

// bonsai check --json (bonsai.check/1, spec §6): what check found in a project. Printed, never stored. Written by
// bonsai check (step 5.1.4b moves it onto Check.Encode); read by the studio, agents and CI.

import (
	"strings"

	"github.com/LastStep/Bonsai/internal/schema"
)

// Check is check --json's document.
type Check struct {
	Findings []Finding     `json:"findings"` // each one makes the exit code 1
	Warnings []Finding     `json:"warnings"` // never the exit code
	Error    *ErrorObject  `json:"error"`    // null unless check refused or failed
	Notes    []string      `json:"notes"`    // what check says to a person that is neither a finding nor a warning (set 6)
	Extra    schema.Object `json:"-"`
}

// Finding is one finding or warning.
type Finding struct {
	Code    string        `json:"code"`    // a fixed word: an open list whose known words are CheckWords
	File    *string       `json:"file"`    // the project-relative file it is about, or null
	Message string        `json:"message"` // one sentence
	Next    Next          `json:"next"`
	Extra   schema.Object `json:"-"`
}

// CheckWords are the known words of a finding's and a warning's code: an open list whose one home is this table
// (formats/README.md); bonsai check --schema bonsai.check prints it, and step 5.1.10's reference page lists it. One
// word per thing check looks for (spec section 6, "bonsai check findings" and "Warnings"), each either a finding (it
// makes check exit 1, and status --json lists it under problems) or a warning (printed, never the exit code, never a
// problem). Who is who usually takes the next step; a finding names the other when its own step is the other's.
// cmd/bonsai's TestCheckTable walks this table: a word with no test case fails, and so does a next step naming a
// command Bonsai does not have or a flag its word does not take. A later piece adds a word by adding an entry here
// with its case (as step 5.6.1 added stranded, which engine's checkLater named till then); a word is added, never renamed or
// taken out.
//
// Machine marks a word that is about this machine rather than the commit (set 7, plan-5 5.4.3 note 4): the home, the
// PATH, the installed Claude Code and its plugins, the cache and which Bonsai wrote the hook lines. A proof is of a
// commit, so rung 0 lists these as "not judged here". The reference page and bonsai check --schema bonsai.check show
// the mark.
var CheckWords = []Word{
	// The project's own files against the lock and bonsai.yaml (offline: the lock alone).
	{Word: "config", Kind: "finding", Who: "person", Means: "bonsai.yaml is not one Bonsai reads"},
	{Word: "lock", Kind: "finding", Who: "person", Means: ".bonsai/lock.json is missing or not one Bonsai reads"},
	{Word: "packs", Kind: "finding", Who: "person", Means: "bonsai.yaml and the lock name other packs, sources, refs or folders"},
	{Word: "changed", Kind: "finding", Who: "person", Means: "a file the lock lists was edited: a pack file, Bonsai's block in CLAUDE.md, or Bonsai's lines in .claude/settings.json"},
	{Word: "missing", Kind: "finding", Who: "person", Means: "a file the lock lists is gone"},
	{Word: "gitignore", Kind: "finding", Who: "person", Means: ".bonsai/.gitignore is missing or changed, so .bonsai/local/ could be committed"},
	{Word: "local", Kind: "finding", Who: "agent", Means: "git tracks or has staged a file from .bonsai/local/, which is never committed"},
	// Bonsai's document kinds (contract sections 2.3, 5, 6, 7).
	{Word: "format0", Kind: "finding", Who: "agent", Means: "a format-0 file the lock's format0 list fixes changed or is gone (contract section 2.3)"},
	{Word: "format0-new", Kind: "finding", Who: "agent", Means: "a task, run report or STATE file with no format: line that is not on the lock's format0 list: it takes a format: line first (contract section 2.3)"},
	{Word: "document", Kind: "finding", Who: "agent", Means: "a file of one of Bonsai's kinds (task, run, state, memory, the two tables) does not read under its format, 0 or 1"},
	{Word: "label", Kind: "finding", Who: "agent", Means: "a label's value is not of the kind its definition says, or it is on a kind of document its definition does not name (contract section 5.2)"},
	{Word: "label-twice", Kind: "finding", Who: "person", Means: "two sources (packs, or a file attached on this machine) define one label name (contract section 5.1)"},
	{Word: "approve-first", Kind: "finding", Who: "person", Means: "a task in a lane with approve_first reached running, verify or done without reading approved since it last read todo or plan, as git history shows (contract section 6)"},
	{Word: "absolute-path", Kind: "finding", Who: "agent", Means: "a committed Bonsai file (bonsai.yaml, a document of one of Bonsai's kinds) holds an absolute path in a field, a pack's source among them unless it is a remote URL (contract section 2.6; spec section 14, check 2)"},
	{Word: "block-size", Kind: "finding", Who: "person", Means: "Bonsai's block in CLAUDE.md is over its fixed 40 lines (spec section 6)"},
	{Word: "memory-index-size", Kind: "finding", Who: "agent", Means: "the memory index is over its fixed 120 lines or 12 KB (spec section 10)"},
	{Word: "memory-note-size", Kind: "finding", Who: "agent", Means: "a memory note is over 4 KB (spec section 10)"},
	{Word: "missing-path", Kind: "finding", Who: "agent", Means: "a project path named in CLAUDE.md, STATE or a memory note does not exist, or exists here only because git ignores it, so a clone does not have it (Bonsai's own block left out: the lock checks it)"},
	{Word: "secret", Kind: "finding", Who: "agent", Means: "a memory note or the memory index in the working tree holds a secret-shaped string, as Bonsai's redactor finds one: the finding names the note, the line and the kind of secret (the secret kinds' list), never the value; a person decides whether to rotate it (spec section 6)"},
	// Claude Code's settings in the project (spec sections 5 and 7).
	{Word: "settings-rule", Kind: "finding", Who: "person", Means: "a permission rule in .claude/settings.json or .claude/settings.local.json is not valid on its own"},
	{Word: "hooks-off", Kind: "finding", Who: "person", Means: "disableAllHooks is true in .claude/settings.json or .claude/settings.local.json, which turns off every hook, Bonsai's guard among them"},
	{Word: "settings-env", Kind: "finding", Who: "person", Means: "the env of .claude/settings.json or .claude/settings.local.json sets a name that wraps, redirects or steers a hook line, Bonsai's guard among them: " + hookEnvNames() + "; the finding names the file and the name, never the value (step 5.3.6)"},
	{Word: "plugin-version", Kind: "finding", Who: "person", Means: "a pack plugin's entry in a Bonsai marketplace carries a version, which a plugin pinned by commit never does (spec section 5)"},
	{Word: "plugin", Kind: "finding", Who: "person", Machine: true, Means: "plugin drift: Claude Code turns on another commit of a locked pack's plugin for this checkout (spec section 5)"},
	// This machine (the home and the PATH).
	{Word: "id-changed", Kind: "finding", Who: "person", Machine: true, Means: "bonsai.yaml's workspace id is not the one this machine last recorded for the checkout (contract section 3)"},
	{Word: "bonsai-path", Kind: "finding", Who: "person", Machine: true, Means: "the bonsai on the PATH is not the installed one that install.json records (spec section 3)"},
	{Word: "git-bash", Kind: "finding", Who: "person", Machine: true, Means: "on Windows, Git Bash is not where Claude Code looks for it, so Claude Code runs every hook line in PowerShell, where Bonsai's lines do not parse: the guard's line, with onFailure block, then refuses every call on Claude Code 2.1.295 or later and lets every call through on an older one, and the stop gate's and the recorder's lines never run (step 5.3.6)"},
	{Word: "git-user", Kind: "finding", Who: "person", Machine: true, Means: "on Windows, the Git Bash Claude Code runs hook lines in is a Git for Windows installed for one user (outside the Program Files folders), so its /usr/local/bin, where Bonsai's lines look for bonsai when Windows' installed place holds none, is a folder that user can write (step 5.3.6)"},
	{Word: "machine-record", Kind: "finding", Who: "person", Machine: true, Means: "a record of a checkout in Bonsai's home (workspaces/r-<hex>/workspace.json) is not one Bonsai reads: until a person mends it, the guard verifies no main checkout on this machine, so no task's grant counts in any project here (step 5.3.6)"},
	// Warnings.
	{Word: "claude-code-old", Kind: "warning", Who: "person", Machine: true, Means: "Claude Code is older than the floor: Bonsai's own or a pack's needs.claude_code, the higher (spec section 7)"},
	{Word: "claude-code-unknown", Kind: "warning", Who: "person", Machine: true, Means: "Claude Code's version could not be read (not on the PATH, or claude --version gave an answer Bonsai does not read)"},
	{Word: "same-id", Kind: "warning", Who: "person", Machine: true, Means: "another checkout on this machine holds the same workspace id (contract section 3)"},
	{Word: "stranded", Kind: "warning", Who: "person", Machine: true, Means: "a machine folder of this workspace in Bonsai's home is stranded under an old checkout path that holds no bonsai.yaml now (moved or gone), with settings or attached labels this checkout does not get; the next step gives a person's lines that bring them over and remove the old folder (contract section 3)"},
	{Word: "bonsai-newer", Kind: "warning", Who: "person", Machine: true, Means: "the Bonsai home's record of Bonsai's releases (cache/release.json, which bonsai status --full reads at most daily) holds a release newer than the installed copy's version (install.json's, else the running bonsai's); its next step points at the install lines, which a person runs on each side (step 5.6.5)"},
	// The personal memory layer in the Bonsai home (spec section 10; step 5.6.3): warnings, read from the home whatever
	// project check runs in, since the layer is the machine's, never a project's.
	{Word: "personal-format", Kind: "warning", Who: "person", Machine: true, Means: "the personal memory index (INDEX.md in the Bonsai home's personal/) or a note in its notes/ does not read as bonsai.memory/1, or the index's kind is not index"},
	{Word: "personal-budget", Kind: "warning", Who: "person", Machine: true, Means: "the personal memory index or a personal note is over its cap (the memory caps on the reference page of lists)"},
	{Word: "personal-index", Kind: "warning", Who: "person", Machine: true, Means: "a line of the personal memory index names a note that does not exist, or a note in personal/notes/ has no line in the index"},
	{Word: "personal-secret", Kind: "warning", Who: "person", Machine: true, Means: "the personal memory index or a personal note holds a secret-shaped string, as Bonsai's redactor finds one, in the layer that loads into every session on the machine: the warning names the file, the line and the kind of secret, never the value"},
	{Word: "personal-copy", Kind: "warning", Who: "person", Machine: true, Means: "on Windows, the personal memory copy is not known to match WSL's (the home setting personal_from): never refreshed since it was set, its last refresh failed, or the copy changed since that refresh; read from the last refresh's record, never from WSL"},
	{Word: "run-reports", Kind: "warning", Who: "person", Means: "run reports are past bonsai.yaml's generated.run rule, listed for a person to delete (spec section 6)"},
	{Word: "approve-first-unchecked", Kind: "warning", Who: "agent", Means: "approve_first was not checked: git history is not available (a shallow clone) or could not be read; never passed"},
	{Word: "plugin-missing", Kind: "warning", Who: "person", Machine: true, Means: "Claude Code reports a locked pack's plugin not installed for this checkout (it has registered the workspace's marketplace, or could not say whether it has)"},
	{Word: "plugin-trust", Kind: "warning", Who: "person", Machine: true, Means: "Claude Code has not registered the workspace's marketplace, which a Claude Code session in the checkout does once a person has trusted the folder (first-time trust, spec section 5), so the lock's plugin cannot be installed yet"},
	{Word: "plugin-unchecked", Kind: "warning", Who: "person", Machine: true, Means: "this machine's plugins were not compared with the lock (Claude Code not on the PATH, its answer unread, a local settings file unread)"},
	{Word: "cache", Kind: "warning", Who: "person", Machine: true, Means: "a lock written before formats set 4 names a pack this machine's cache lacks, so Bonsai's lines in .claude/settings.json were not checked"},
	{Word: "tables", Kind: "warning", Who: "agent", Means: "a generated table in the main checkout is stale: the tasks table (.bonsai/tasks.md) differs from a rebuild of the task files, or the sessions table (.bonsai/sessions.md) lacks a row for an ended session or subagent run in the log, or either is missing; the message names which; the tables lag between moves by design and grant nothing (spec section 6; contract section 7.5)"},
	{Word: "local-unchecked", Kind: "warning", Who: "agent", Means: "git ls-files failed, so files from .bonsai/local/ in git's index were not looked for"},
	{Word: "own-hooks", Kind: "warning", Who: "person", Machine: true, Means: "Bonsai's own hook lines in .claude/settings.json are not this build's (the project was linked or last updated by another Bonsai), while Bonsai's lines are otherwise as the lock says; a new hook line runs code, so update writes them only with a person's --allow-exec (spec section 7; step 5.1.1)"},
}

// HookEnvKeys are the names that, set in the env of a project's Claude Code settings (.claude/settings.json or
// .claude/settings.local.json), wrap, redirect or steer a hook line, Bonsai's guard and stop gate among them, or what
// Bonsai's own commands read: their one home (plan-5 5.3.6 note 3), which check's settings-env finding reads and its
// word's meaning lists (so the reference page and bonsai check --schema bonsai.check show them). Claude Code copies a
// settings file's env into the environment of every hook and every tool command, and applies a change to it in the
// running session (its env-vars and settings references). A name matches whatever its case, as Windows reads Path as
// PATH, and every name starting with HookEnvPrefix counts too. The user's own settings file is not read: it is the
// person's. Means says why each counts; a name is added here with its why, and the reference page is regenerated in
// the same commit. What step 5.3.6's sessions saw of each, set in a project's settings before a session and written
// into its local settings during one (Claude Code 2.1.296 in WSL, where hook lines run in sh; 2.1.294 on Windows,
// where they run in Git Bash) is in the comment beside it.
var HookEnvKeys = []Word{
	// Reached hook lines on both sides, set before and changed during: a decoy bonsai first in it answered the line.
	{Word: "PATH", Means: "where a hook line's shell finds a program it calls by name, bonsai among them"},
	// Reached hook lines on Windows, before and during; in WSL only the Bash tool's bash read it, not sh.
	{Word: "BASH_ENV", Means: "a script bash (Git Bash on Windows) runs before every hook line it runs"},
	// Reached no hook line on either side; listed as the plan has it, for a shell that reads it.
	{Word: "ENV", Means: "a script sh runs first when it starts interactive or as POSIX bash"},
	// Reached hook lines on both sides, before and during: the wrapper ran with the line as its argument.
	{Word: "CLAUDE_CODE_SHELL_PREFIX", Means: "a program Claude Code runs in place of every shell hook line, handing it the line"},
	// A decoy bash.exe named in a project's env was never started, before or during (Claude Code 2.1.294); the
	// env-vars reference names the variable, and the user's settings set it.
	{Word: "CLAUDE_CODE_GIT_BASH_PATH", Means: "on Windows, the bash.exe that runs every hook line (any file of that name)"},
	// Hooks were told the real project whatever env said; tool commands, and so Bonsai's commands an agent runs, were
	// told env's value, on both sides, before and during.
	{Word: "CLAUDE_PROJECT_DIR", Means: "the project a session's commands are told they run in, which Bonsai reads"},
	// Not set in the sessions; its reference: 0 turns the cap off. Claude Code 2.1.296 names it when a Stop hook has
	// blocked nine times in a row ("Set CLAUDE_CODE_STOP_HOOK_BLOCK_CAP to raise this limit"), so a low value in a
	// project's env would let a session end past Bonsai's stop gate sooner.
	{Word: "CLAUDE_CODE_STOP_HOOK_BLOCK_CAP", Means: "caps how many times in a row a Stop hook, Bonsai's stop gate among them, may block a session's end before Claude Code lets it end (0: no cap)"},
	// A blank value changed nothing: Claude Code set 1 in every hook and tool command, on both sides, before and during.
	{Word: "CLAUDE_CODE_CHILD_SESSION", Means: "how Bonsai's commands know they run inside an agent's session (step 5.3.6)"},
}

// HookEnvPrefix starts the names of Bonsai's own variables (its home, the running task, a test build's fault): any
// of them in a project's settings env is a settings-env finding too.
const HookEnvPrefix = "BONSAI_"

// HookEnvKey says whether a settings env name is one of HookEnvKeys or starts with HookEnvPrefix, whatever its case,
// and why it counts.
func HookEnvKey(name string) (why string, ok bool) {
	for _, k := range HookEnvKeys {
		if strings.EqualFold(name, k.Word) {
			return k.Means, true
		}
	}
	if len(name) >= len(HookEnvPrefix) && strings.EqualFold(name[:len(HookEnvPrefix)], HookEnvPrefix) {
		return "one of Bonsai's own variables (its home, the running task, a test build's fault)", true
	}
	return "", false
}

// hookEnvNames lists HookEnvKeys for the settings-env word's meaning.
func hookEnvNames() string {
	var names []string
	for _, k := range HookEnvKeys {
		names = append(names, k.Word)
	}
	return strings.Join(names, ", ") + ", or any name starting " + HookEnvPrefix
}

// LadderCheckWords are the known words of a finding's and a warning's code that are about the ladder's shape in
// bonsai.yaml (set 7, plan-5 5.4.0; the checks themselves are step 5.4.2's): their one home, beside CheckWords in the
// same bonsai.check/1 document and the same open list (bonsai check --schema bonsai.check prints both tables; no word
// is in both). They are a table of their own so that this file is final before 5.4.2 writes the checks: cmd/bonsai's
// TestCheckTable and engine's TestCheckWordsInTheCode walk CheckWords, and 5.4.2 walks this table the same way with
// its checks. None is about this machine. Who is person: bonsai.yaml is a person's file (an agent edits it only under
// a task a person gave it). CheckWord finds these too.
var LadderCheckWords = []Word{
	{Word: "rung-twice", Kind: "finding", Who: "person", Means: "two rungs of bonsai.yaml's ladder carry one number, so a task's bonsai.ladder and the ladder_floor cannot name one of them"},
	{Word: "floor-no-rung", Kind: "finding", Who: "person", Means: "ladder_floor names a rung number that no rung of the ladder has"},
	{Word: "ratchet-no-rung", Kind: "finding", Who: "person", Means: "a floor in bonsai.yaml's ratchets that no rung's ratchet names, so it holds back nothing"},
	{Word: "capture-pattern", Kind: "finding", Who: "person", Means: "a rung's capture holds a pattern that does not compile as a Go (RE2) regular expression: the finding names the rung, the capture and the error"},
	{Word: "ratchet-no-count", Kind: "finding", Who: "person", Means: "a rung's ratchet has no count: no capture of the same name, and the rung has no tests form to count passed tests from"},
	{Word: "rung-bash", Kind: "finding", Who: "person", Means: "a rung's command calls bash by name (spec section 3), so it would run a different shell on Windows than elsewhere"},
	{Word: "tests-unknown", Kind: "warning", Who: "person", Means: "a rung's tests is a word Bonsai does not know, so its tests are not read: the warning names the known ones (TestForms, bonsai check --schema bonsai.workspace)"},
}

// RoleCheckWords are the known words of a finding's code about the roles' limits a project sets in bonsai.yaml
// (packs[].roles; set 8, plan-5 5.5.9; the checks themselves are step 5.5.0's, note 8): their one home, beside
// CheckWords in the same bonsai.check/1 document and the same open list (bonsai check --schema bonsai.check prints
// every table; no word is in two). A table of its own, as LadderCheckWords is, so that this file is final before
// 5.5.0 writes the checks: cmd/bonsai's TestCheckTable and engine's TestCheckWordsInTheCode walk CheckWords, and
// 5.5.0 walks this table the same way with its checks. Each is a finding whose next step is an agent's (note 8); none
// is about this machine. CheckWord finds these too.
var RoleCheckWords = []Word{
	{Word: "role-undeclared", Kind: "finding", Who: "agent", Means: "bonsai.yaml sets a limit under a pack's roles for a role that pack does not declare: the finding names the pack's roles"},
	{Word: "role-default-over-max", Kind: "finding", Who: "agent", Means: "a role's limit in force has its default above its max (bonsai.yaml's or the pack's): the max then holds, and the finding names both"},
	{Word: "role-actions-over-turns", Kind: "finding", Who: "agent", Means: "a role's max of actions in force is above the role's max_turns, so Claude Code would stop the agent before the guard's budget does: lower the max, or move to a pack whose role allows more"},
}

// MessageCheckWords are the known words of a warning's code about the messages between agents in
// .bonsai/local/messages/ (set 8; the check is step 5.5.10's, note 5): their one home, beside CheckWords in the same
// open list, in a table of their own for the reason RoleCheckWords is. A warning never changes the exit code: one bad
// file never blocks the rest. Machine: the messages live in .bonsai/local/, never committed. CheckWord finds these too.
var MessageCheckWords = []Word{
	{Word: "message-malformed", Kind: "warning", Who: "agent", Machine: true, Means: "a message file under .bonsai/local/messages/ does not read as bonsai.message/1 (its frontmatter, a field, a report with no status, a newer major): the warning names the file and why, and bonsai messages lists it as malformed"},
}

// CheckWord finds a word in CheckWords, or in LadderCheckWords, RoleCheckWords or MessageCheckWords.
func CheckWord(code string) (Word, bool) {
	for _, table := range [][]Word{CheckWords, LadderCheckWords, RoleCheckWords, MessageCheckWords} {
		for _, w := range table {
			if w.Word == code {
				return w, true
			}
		}
	}
	return Word{}, false
}

// PackCheckWords are the known words of a finding's code from bonsai check --pack <folder> (step 5.1.9; spec section
// 5, "Every template and pack file documents itself"): one per rule a pack's folder is held to, their one home, beside
// CheckWords in the same bonsai.check/1 document and the same open list (bonsai check --schema bonsai.check prints both
// tables; no word is in both). Every one is a finding (it makes check --pack exit 1); a pack's folder has no warnings,
// and status --json never lists these (they are about a pack, not a project). Who is agent for each: a pack's maker,
// an agent in the pack's repository, edits the pack's own files. What counts for each rule is engine's checkpack.go.
// cmd/bonsai's TestCheckTable walks this table as it walks CheckWords: a word with no failing fixture fails, a
// fixture that fails on any other word fails, and so does a next step naming a command Bonsai does not have.
var PackCheckWords = []Word{
	{Word: "pack-schema", Kind: "finding", Who: "agent", Means: "bonsai/pack.yaml, bonsai/labels.yaml or bonsai/lanes.yaml does not fit its format as its writer writes it (format 1's YAML, the format line, every field present, each of a type and value the schema allows, in the schema's order), or the engine would refuse it at a link (a rule a schema cannot say, or a file pack.yaml names missing from the pack)"},
	{Word: "pack-comment", Kind: "finding", Who: "agent", Means: "a key in bonsai/pack.yaml, labels.yaml or lanes.yaml has no # comment at the end of its line or on the line just above it (spec section 5: every key documents itself)"},
	{Word: "pack-fields", Kind: "finding", Who: "agent", Means: "a template skill's fields table and its template differ: a field of the template with no row, a row naming no field the template holds, no template after the table, or, for a Bonsai format, fields other than the format's schema"},
	{Word: "pack-values", Kind: "finding", Who: "agent", Means: "an allowed-values cell of a template skill's fields table lists values other than the closed list its field has (a Bonsai format's schema, or the pack's own lanes, a declared kind's statuses or a choice label's values)"},
	{Word: "pack-why", Kind: "finding", Who: "agent", Means: "a deny rule in bonsai/pack.yaml has no why, the sentence update's preview prints for it: missing, null or blank"},
	{Word: "pack-plugin-version", Kind: "finding", Who: "agent", Means: ".claude-plugin/plugin.json carries a version, which a pack pinned by commit never does (spec section 5), or is not a JSON object Bonsai reads"},
	{Word: "pack-block", Kind: "finding", Who: "agent", Means: "the instruction block a project linked to this pack alone would get (its markers, imports, label definitions and bonsai/block.md) is over its fixed 40 lines (spec section 6)"},
	{Word: "pack-documents", Kind: "finding", Who: "agent", Means: "a document kind bonsai/pack.yaml declares is not well formed (contract section 7.3): a field the schema does not allow, a name of Bonsai's own or given twice, not exactly one of path and file, an id pattern Go does not read, a status twice, or a move or stamp naming a status the kind does not have"},
	{Word: "pack-protected", Kind: "finding", Who: "agent", Means: "a protected path bonsai/pack.yaml declares is not a well-formed glob: project-relative with forward slashes, no empty, . or .. segment, each segment one path.Match reads (** whole segments only)"},
	{Word: "pack-bash", Kind: "finding", Who: "agent", Means: "a hook command calls bash by name (spec section 3): a hook line in bonsai/pack.yaml, or a hook of the plugin itself"},
	{Word: "pack-runs", Kind: "finding", Who: "agent", Means: "a hook command in bonsai/pack.yaml names a file the pack writes (by its path in the project or its file name) that its runs does not list, so a change to that file would run unseen (step 5.1.1)"},
}

// PackRoleCheckWords are the known words of a finding's code from bonsai check --pack <folder> about what step 5.5
// adds to a pack (set 8; the rules are step 5.5.0's, note 9): its roles against its agents/ folder, and the form of its
// deny rules on project paths. Their one home, beside PackCheckWords in the same open list, in a table of their own so
// that this file is final before 5.5.0 writes the rules (cmd/bonsai's TestCheckPackTable and engine's
// TestCheckWordsInTheCode walk PackCheckWords; 5.5.0 walks this table the same way). Every one is a finding whose next
// step is an agent's. PackCheckWord finds these too.
var PackRoleCheckWords = []Word{
	{Word: "pack-role-file", Kind: "finding", Who: "agent", Means: "a role bonsai/pack.yaml declares has no agents/<name>.md in the plugin"},
	{Word: "pack-role-undeclared", Kind: "finding", Who: "agent", Means: "a file in the plugin's agents/ folder is not a role bonsai/pack.yaml declares"},
	{Word: "pack-role-turns", Kind: "finding", Who: "agent", Means: "a role file's maxTurns is not the max_turns bonsai/pack.yaml declares for it (missing on one side, or another number)"},
	{Word: "pack-deny-relative", Kind: "finding", Who: "agent", Means: "a deny rule in bonsai/pack.yaml on a project path is written without its leading / or //, so it would not hold after a shell cd (step 5.3.5): the finding names the form that holds"},
}

// PackCheckWord finds a word in PackCheckWords, or in PackRoleCheckWords.
func PackCheckWord(code string) (Word, bool) {
	for _, table := range [][]Word{PackCheckWords, PackRoleCheckWords} {
		for _, w := range table {
			if w.Word == code {
				return w, true
			}
		}
	}
	return Word{}, false
}

// ReadCheck reads check --json's output.
func ReadCheck(raw []byte) (*Check, error) {
	c := &Check{}
	if err := MustLookup("check").readJSONInto(raw, c); err != nil {
		return nil, err
	}
	return c, nil
}

// Encode writes the document, held to bonsai.check/1.
func (c *Check) Encode() ([]byte, error) { return MustLookup("check").Encode(c) }
