package engine

// The lists of this package that other code and the reference page of lists (docs/reference/lists.md, step 5.1.10)
// read, each in one table with a test that holds the table to the code that uses its words. Standard library only.

import "github.com/LastStep/Bonsai/internal/format"

// PluginResults are the results of the plugin step for one pack's plugin (PluginResult.Result; changes --json's
// plugins[].result): an open list, its known words here, their one home. The tests fail on a result the code gives
// that is not here.
var PluginResults = []format.Word{
	{Word: "installed", Means: "Claude Code has the plugin on, at the locked commit, for this checkout"},
	{Word: "waiting", Means: "not installed yet: the plugin carries code and --allow-exec was not given, or Claude Code has not registered the workspace's marketplace (a person's first session in the trusted folder does)"},
	{Word: "failed", Means: "Claude Code was asked and could not install (or list) the plugin; the next step is the command to run"},
	{Word: "skipped", Means: "Claude Code is not on the PATH, so nothing was installed (or, on unlink, nothing removed) on this machine"},
	{Word: "uninstalled", Means: "unlink only: the plugin was removed for this checkout, or was not installed (nothing to remove)"},
}

// ClaudeStates are the states of Claude Code's version against the floor (ClaudeState.State; status --full's
// checks.claude_code.state): a closed list, spec section 7.
var ClaudeStates = []format.Word{
	{Word: "ok", Means: "Claude Code's version was read and is at or above the floor"},
	{Word: "old", Means: "Claude Code's version was read and is below the floor (check warns: claude-code-old)"},
	{Word: "unknown", Means: "Claude Code's version could not be read: not on the PATH, failed, or an answer with no version (check warns: claude-code-unknown)"},
}

// LaterWord is a check word a later step builds.
type LaterWord struct {
	Word, Step, Means string
}

// CheckLaterWords are spec section 6's findings and warnings that later steps build, each with its step: their words
// join format.CheckWords when they are built.
func CheckLaterWords() []LaterWord {
	out := make([]LaterWord, len(checkLater))
	for i, l := range checkLater {
		out[i] = LaterWord{Word: l.Code, Step: l.Step, Means: l.What}
	}
	return out
}

// KeyOrder is one order Claude Code writes in, as Bonsai measured it.
type KeyOrder struct {
	Name  string   // the Go identifier
	Where string   // what it orders
	Keys  []string // in Claude Code's order
}

// ClaudeOrders are the orders Claude Code writes .claude/settings.json in, measured on Claude Code 2.1.294 and
// 2.1.295 (step 5.1.7); a key Bonsai adds goes where Claude Code would put it, so a later Claude Code write moves
// nothing.
func ClaudeOrders() []KeyOrder {
	return []KeyOrder{
		{"claudeKeyOrder", "the top-level keys of a settings file", append([]string{}, claudeKeyOrder...)},
		{"claudePermissionsOrder", "the keys of permissions", append([]string{}, claudePermissionsOrder...)},
		{"claudeMarketplaceOrder", "the keys of an inline marketplace's source object", append([]string{}, claudeMarketplaceOrder...)},
	}
}
