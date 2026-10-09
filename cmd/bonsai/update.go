package main

// bonsai update (spec §4, §6): bring this project's packs to the refs in bonsai.yaml. Its table of flags and exit
// codes is updateWord; it runs with init's run (engine.go), and with --json prints the changes output
// (bonsai.changes/1).

func init() { register(updateWord) }

var updateWord = &Word{
	Name:    "update",
	Order:   2,
	Title:   "bring this project's packs to the refs in bonsai.yaml (spec section 6).",
	Summary: "bring the packs to the refs in bonsai.yaml",
	Args:    "[flags]",
	About: `It fetches each pack, then decides per file from three fingerprints (what the lock says Bonsai wrote, what is on
disk, what the pack now gives): unchanged, updated, created, adopted, changed (your edit, left alone) or conflict.
The preview names every file and every line of .claude/settings.json it adds, changes or removes, with a sentence.
All or nothing: every file is staged, then renamed, the lock last.
Then, written or with nothing to change, it brings this machine's plugins to the lock: for each pack it runs
claude plugin install <pack>@<marketplace> --scope project in this checkout (a no-op once installed; the first time,
Claude Code may write .claude/settings.json again in its own key order). A pack's plugin that carries code parts
(hooks, MCP and LSP servers, monitors, mods) is installed on this machine only with --allow-exec, on each machine:
without it the plugin is "waiting" and the next step is bonsai update --allow-exec; one already installed at the
locked commit is left as it is. Bonsai installs only the project's own packs' plugins, and removes none. Claude Code
knows a new marketplace (a new commit, or a new checkout) only after a Claude Code session in the checkout, in a
trusted folder, has registered it: until then the plugin is "waiting", and the output names the next step. This step
never changes the exit code.
`,
	Flags: []Flag{
		{Name: "--yes", Help: "write without asking"},
		{Name: "--allow-exec", Help: "consent to what the preview lists under \"Runs code\" (needed as well as --yes; see below)"},
		{Name: "--diff", Help: "show each file's changes as a diff in the preview"},
		{Name: "--keep", Value: "P", Many: true, Help: "settle a conflict on P by keeping your edit (the file becomes kind kept; a later pack change to it\nis a conflict again)"},
		{Name: "--adopt", Value: "P", Many: true, Help: "settle a conflict on P by taking the pack's copy; yours is saved in the Bonsai home's cache"},
		{Name: "--json", Help: "print the changes document (bonsai.changes/1) instead of text; never asks"},
	},
	Notes: `Runs code: a hook line added or changed (Bonsai's own or a pack's), a pack file a pack's hook line runs (new or
changed), and a plugin's own code parts (hooks, MCP and LSP servers, monitors, mods) changed since the locked
commit. The preview lists each under "Runs code"; all or nothing, so without --allow-exec as well as --yes nothing
is written, at a terminal too. A removed hook line runs nothing.
`,
	Exits: []Exit{
		{Code: 0, Means: "updated (or nothing to change)"},
		{Code: 2, Means: "bad input: a flag or value wrong, a pack Bonsai cannot link, or a bonsai.yaml or .claude/settings.json\nBonsai does not read (nothing written)"},
		{Code: 3, Means: "runtime: a fetch failed, git or the Bonsai home missing, a file that cannot be read or written (nothing\nwritten); or the write stopped part-way (--json: result failed): run the same command again"},
		{Code: 4, Means: "wrong state (not linked, no lock, a lock Bonsai does not read, a pack taken out of bonsai.yaml, a\npack's tag moved to another commit), no --yes, or code without --allow-exec (the preview was printed;\nnothing written)"},
		{Code: 5, Means: "conflicts (nothing written)"},
	},
	Examples: []string{"bonsai update --yes"},
	Refused:  changesRefused,
	Run:      runEngine,
}
