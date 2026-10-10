package format

// Bonsai's exit codes (spec §3), in one table: the words' own tables (cmd/bonsai's Word.Exits) say which of these each
// word returns and what it means for that word. bonsai --help's last lines, bonsai --help --json (bonsai.help/1) and
// the reference page of lists (docs/reference/lists.md) all read this table, their one home. A code is never renamed
// or taken out; a new one is an addition.

// ExitCode is one exit code and what it means.
type ExitCode struct {
	Code    int
	Short   string // the code in a few words, as bonsai --help lists it: "bad input"
	Means   string // one line: what it means
	Applies string // ExitEveryWord or ExitHook: the words it is for
}

// Who an exit code is for.
const (
	ExitEveryWord = "every word but hook"
	ExitHook      = "hook"
)

// ExitCodes are the exit codes: the six every word but hook uses, then hook's two (it speaks Claude Code's hook
// format, where 2 blocks the call).
var ExitCodes = []ExitCode{
	{0, "ok", "the command did what it was asked (check found nothing, status read the workspace, a preview was printed and nothing was due)", ExitEveryWord},
	{1, "check findings", "check found at least one finding (a warning never makes it 1)", ExitEveryWord},
	{2, "bad input", "the command line, or a file the command was given, is not one Bonsai reads: nothing was done", ExitEveryWord},
	{3, "runtime", "something failed that the command line did not cause: a file or folder could not be read or written, git is missing, a fetch failed", ExitEveryWord},
	{4, "wrong state or no --yes", "the project is in a state the command does not run in, or the command writes and was given no --yes with no terminal to ask at: it printed the preview and wrote nothing", ExitEveryWord},
	{5, "conflicts", "files edited here were changed by the pack too: nothing is written until each is settled", ExitEveryWord},
	{0, "allow", "the hook lets the call go on", ExitHook},
	{2, "block", "the hook stops the call; the reason and the next step are on stderr, which Claude Code shows to the agent", ExitHook},
}
