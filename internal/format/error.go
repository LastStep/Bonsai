package format

// The error object (bonsai.error, spec §3 and §16 row 29; format review 4.5): what a command's --json carries when
// the command refuses or fails. It has no format line and no major: it is part of status, check and changes, and
// changes with them, only by additions.

import "github.com/LastStep/Bonsai/internal/schema"

// ErrorObject is the error object.
type ErrorObject struct {
	Code    string        `json:"code"`    // a fixed word: an open list whose known words are ErrorWords
	Message string        `json:"message"` // what went wrong, one sentence
	Next    Next          `json:"next"`
	Extra   schema.Object `json:"-"`
}

// Next is a next step: what to do, and whose step it is. The error object, a finding, a warning and the plugin step
// all take this shape (formats/README.md).
type Next struct {
	Do    string        `json:"do"`
	Who   string        `json:"who"` // the closed list: agent, person
	Extra schema.Object `json:"-"`
}

// ErrorWords are the known words of the error object's code: an open list (contract §2.2) whose one home is this
// table (formats/README.md); bonsai check --schema bonsai.error prints it, and step 5.1.10's reference page lists
// it. One word per cause, not per message: every refusal and failure of every command names one of them (step
// 5.1.4b), and cmd/bonsai's tests fail on a word that is not here. Each entry says what the word means and who
// usually takes the next step (Who: agent or person, the error object's next.who); a refusal may name the other
// when its own next step is the other's. A later piece adds a word by adding an entry here (step 5.2.0 added
// 5.2's four: ask-not-open, answer-own-session, label-not-defined and session-not-found; step 5.4.0 added the
// ladder's four: ladder-busy, no-task, rung-not-defined and main-not-verified; set 8, step 5.5.9, added bonsai
// message's six: role-not-declared, budget-over-max, report-no-status, report-not-in-thread, pointer-missing and
// message-not-found; set 10, step 5.6.0, added bonsai labels' two: namespace-taken and label-redefined); a word is added, never renamed or taken out (format review 4.5: adding one is an addition).
var ErrorWords = []Word{
	// The command line (exit 2).
	{Word: "unknown-command", Who: "agent", Means: "no command word was given, or one this Bonsai does not have"},
	{Word: "bad-flag", Who: "agent", Means: "the command line has a flag the word does not take, a flag without its value, a value given to a switch, a flag given twice that is taken once, or a word left over"},
	{Word: "missing-value", Who: "agent", Means: "a value the command needs was not given (init's --name, --source and --ref; hook's name)"},
	{Word: "bad-value", Who: "agent", Means: "a value was given in a form the command refuses (one of init's values; a --keep or --adopt path that names no conflict or edited file, or names one twice)"},
	{Word: "values-differ", Who: "person", Means: "init was given values that say otherwise than the bonsai.yaml already in the checkout"},
	{Word: "unknown-format", Who: "agent", Means: "check --schema names a format Bonsai does not know"},
	{Word: "not-built", Who: "agent", Means: "a word, flag or step this build of Bonsai does not have yet (a later version adds it)"},
	// Where the command runs (exit 4, or 3 for status).
	{Word: "not-a-checkout", Who: "agent", Means: "the folder is not inside a git checkout"},
	{Word: "not-linked", Who: "person", Means: "the checkout has no bonsai.yaml, so it is not linked to Bonsai"},
	{Word: "old-workspace", Who: "person", Means: "the checkout is a Bonsai 0.4.3 workspace, which this Bonsai neither reads nor changes"},
	{Word: "not-main-checkout", Who: "agent", Means: "the step runs only in the project's main checkout, and this is a worktree"},
	{Word: "no-lock", Who: "person", Means: "bonsai.yaml is in the checkout but .bonsai/lock.json is not, so update cannot tell what was consented to, nor unlink what Bonsai wrote"},
	{Word: "not-a-pack", Who: "agent", Means: "check --pack was given a folder that is not a pack's: it is not there, is not a folder, or holds no bonsai/pack.yaml"},
	// A file Bonsai reads (exit 2, or 4 for the lock).
	{Word: "bad-config", Who: "person", Means: "bonsai.yaml is not one Bonsai reads (a line the reader refuses, a field of the wrong kind, another format)"},
	{Word: "bad-lock", Who: "person", Means: ".bonsai/lock.json is not one Bonsai reads"},
	{Word: "bad-file", Who: "person", Means: "a project file Bonsai writes into is not in a form it reads (.claude/settings.json not a JSON object Bonsai reads, CLAUDE.md's Bonsai block markers broken)"},
	// The packs (exit 2, or 3 for a fetch).
	{Word: "bad-pack", Who: "person", Means: "a pack is not one Bonsai can link (no bonsai/pack.yaml, a pack.yaml refused, a file it names missing, an id other than bonsai.yaml's, a hook line's runs not matching what it runs)"},
	{Word: "packs-overlap", Who: "person", Means: "two packs write one path"},
	{Word: "ref-not-found", Who: "person", Means: "a pack's tag or commit is not in its source, or is not a commit"},
	{Word: "tag-moved", Who: "person", Means: "a pack's tag now resolves to another commit than the one the lock holds (spec section 5: a moved tag is refused): nothing was written"},
	{Word: "fetch-failed", Who: "agent", Means: "fetching a pack failed (the network, the server, or git)"},
	// This machine (exit 3).
	{Word: "git-missing", Who: "person", Means: "git is not on the PATH"},
	{Word: "bad-home", Who: "person", Means: "the Bonsai home cannot be found or used (BONSAI_HOME, or the pack cache in it)"},
	{Word: "read-failed", Who: "agent", Means: "a file or folder cannot be read"},
	{Word: "write-failed", Who: "agent", Means: "a file cannot be written: nothing in the project was written"},
	// Consent and conflicts (exits 4 and 5).
	{Word: "needs-yes", Who: "person", Means: "the command writes, and was given no --yes with no terminal to ask at: it printed the preview and wrote nothing"},
	{Word: "needs-allow-exec", Who: "person", Means: "the plan writes code that runs on this machine, which needs --allow-exec as well as --yes: nothing was written"},
	{Word: "conflicts", Who: "person", Means: "files edited here were changed by the pack too: nothing is written until each is settled with --keep or --adopt"},
	{Word: "plan-changed", Who: "person", Means: "update or bonsai pack was given --expect with a fingerprint the plan built now does not have: something the preview showed would now come out differently, so nothing was written (exit 4); the next step is to read the preview again and run it with the fingerprint it prints"},
	// Who runs a command, and where (exit 4; step 5.3.6).
	{Word: "person-command", Who: "person", Means: "an agent's session (one of the agent markers set: " + agentMarkerNames() + ") ran a command only a person runs: init or update with --allow-exec, unlink, an update that changes or takes out Bonsai's guard or stop line, or an init or update that would record another base branch, or settings for the checks looser than the ones this machine accepted; an agent may run one only in a project in command mode, under a running task whose bonsai.allows a person gave bonsai.yaml and .bonsai/lock.json. Two no grant opens: settings set and labels attach or detach, and a --yes (bonsai pack remove, pack set or add, update or init) whose plan takes protections away, a pack taken out or a deny rule or protected path dropped. Nothing was written"},
	{Word: "test-build", Who: "agent", Means: "a test build of Bonsai (stamped with its hook lines' place and a scratch folder) was asked to link, update, unlink or run the ladder outside its scratch folder, or was stamped with one of the two only: it writes nothing there"},
	// Asks, answers, the log's append and logs (steps 5.2.3 and 5.2.5).
	{Word: "ask-not-open", Who: "agent", Means: "bonsai answer or ask --resolve names a key that has no ask, or whose ask is already answered or resolved, or ask --status names a key that has no ask: the message says which, and nothing was written (exit 4)"},
	{Word: "answer-own-session", Who: "person", Means: "the session that filed the ask tried to answer it: only a person, or a session that did not ask, answers (contract section 9.3), and nothing was written (exit 4)"},
	{Word: "label-not-defined", Who: "agent", Means: "bonsai log append names a label that no definition in force has, or gives it a value of the wrong kind: nothing was written (exit 2)"},
	{Word: "session-not-found", Who: "agent", Means: "bonsai logs --session matches no session in the log, or matches several: the message names the matches (exit 4)"},
	// The ladder (step 5.4.0): a refusal of bonsai ladder. A stamped test build's ladder outside its scratch folder is
	// test-build; a result that cannot be written is write-failed; a flag or value it refuses is the command line's.
	{Word: "ladder-busy", Who: "agent", Means: "another ladder run holds this home's lock for the checkout and did not let go within the wait: nothing was run or written (exit 4); the message names the holder's process and when its budget ends, and the same command again after that is the next step"},
	{Word: "no-task", Who: "agent", Means: "a local ladder run names no task, or the task named is none: the message carries the active-task function's reason code (no task, two running, the file unreadable, no such task), and nothing was run (exit 2)"},
	{Word: "rung-not-defined", Who: "agent", Means: "a rung the command asked for is not in bonsai.yaml's ladder, or bonsai.yaml has no ladder: the message names the rung, and nothing was run (exit 2)"},
	{Word: "main-not-verified", Who: "person", Means: "the project's main checkout is not verified on this machine, or it is not on the workspace's base branch this machine recorded for it, so no result can be written where the stop gate reads it, and nothing was run (exit 4); the person's step is bonsai update in the main checkout, or switching it back to its base branch (contract section 13; the guard's rule of the same name)"},
	{Word: "config-uncommitted", Who: "person", Means: "the main checkout's bonsai.yaml differs from the one committed at its HEAD, or its HEAD holds none: the ladder reads its rungs, floor, ratchets and ci_marked_tests from the committed one only, so nothing was run (exit 4); the person's step is to commit the change in the main checkout or revert it"},
	{Word: "config-looser", Who: "person", Means: "the main checkout's bonsai.yaml, as committed at its HEAD, sets the checks looser than the settings this machine last accepted for the workspace (a rung out of the floor or the ladder, a rung's check changed, a ratchet's floor lowered, a CI mark added, a path out of protected, person_only, never_edit or never_read, a document moved; the message names each), or this machine's record holds none yet: nothing was run (exit 4); the person's step is bonsai update in the main checkout, which records them as accepted, or a commit that puts the settings back"},
	// Messages between agents (set 8; the command is step 5.5.10's): a refusal of bonsai message. An address of the wrong
	// form is bad-value; a flag its kind does not take (a brief's --status) is bad-flag. Who may write a brief to whom,
	// and whether --from is the writer's own address, are the guard's rules (step 5.5.11), not these words.
	{Word: "role-not-declared", Who: "agent", Means: "bonsai message's --from or --to names a role the lock does not declare, in a project whose linked packs declare roles: the message names the declared roles, and nothing was written (exit 2)"},
	{Word: "budget-over-max", Who: "agent", Means: "a brief's --budget gives a limit above that limit's max in force for its recipient's role (bonsai.yaml's, else the pack's): the message names the max, and nothing was written (exit 2); a person may raise the max in bonsai.yaml"},
	{Word: "report-no-status", Who: "agent", Means: "bonsai message --kind report was given no --status (done, blocked or needs-input): nothing was written (exit 2)"},
	{Word: "report-not-in-thread", Who: "agent", Means: "a report names a thread that holds no brief to its --from, or a --to that is not that brief's from (only a brief's own writer may close its dropped thread, blocked, to the brief's to): nothing was written (exit 4)"},
	{Word: "pointer-missing", Who: "agent", Means: "a --pointer names a file that does not exist, or one outside the project and the home: nothing was written (exit 2)"},
	{Word: "message-not-found", Who: "agent", Means: "bonsai message --show, --thread or --reply-to names an id that no message under .bonsai/local/messages/ has: nothing was written or shown (exit 4)"},
	// Labels attached on this machine (set 10; the commands are step 5.6.2's): a refusal of bonsai labels attach. A file
	// that does not read as bonsai.labels/1 is bad-value; an agent's session is person-command.
	{Word: "namespace-taken", Who: "person", Means: "bonsai labels attach was given definitions whose namespace is bonsai or a locked pack's, which only that pack defines (contract section 5.1): nothing was written (exit 4)"},
	{Word: "label-redefined", Who: "person", Means: "bonsai labels attach would change a definition already attached: a definition changed or removed, or other content at the same or an older version (contract section 5.1: a definition may be added, never redefined). Nothing was written (exit 4); the person's step is bonsai labels detach <namespace> and then the attach again"},
	// Part-way, and the unexpected (exit 3).
	{Word: "partly-written", Who: "agent", Means: "the command stopped part-way through writing, the lock not yet written (unlink: not yet removed): the same command again finishes the rest"},
	{Word: "unexpected", Who: "agent", Means: "something failed that Bonsai does not expect (no random number from the system, its own document not fitting its schema): run it again, and report it if it repeats"},
}

// ErrorWord finds a word in ErrorWords.
func ErrorWord(code string) (Word, bool) {
	for _, w := range ErrorWords {
		if w.Word == code {
			return w, true
		}
	}
	return Word{}, false
}

// ReadErrorObject reads an error object on its own (its JSON).
func ReadErrorObject(raw []byte) (*ErrorObject, error) {
	e := &ErrorObject{}
	if err := MustLookup("error").readJSONInto(raw, e); err != nil {
		return nil, err
	}
	return e, nil
}

// Encode writes the error object on its own, held to bonsai.error.
func (e *ErrorObject) Encode() ([]byte, error) { return MustLookup("error").Encode(e) }
